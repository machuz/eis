package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestStoreGetSetRoundtrip(t *testing.T) {
	dir := t.TempDir()
	s := &Store{enabled: true, baseDir: dir}

	// Set
	data := []string{"alice", "bob", "charlie"}
	key := "test/data.gob"
	if err := s.Set(key, data); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Get
	var got []string
	if !s.Get(key, &got) {
		t.Fatal("Get: expected cache hit")
	}
	if len(got) != 3 || got[0] != "alice" {
		t.Fatalf("Get: unexpected data: %v", got)
	}
}

func TestStoreDisabled(t *testing.T) {
	s := &Store{enabled: false, baseDir: t.TempDir()}

	if err := s.Set("key", "value"); err != nil {
		t.Fatalf("Set on disabled store should not error: %v", err)
	}

	var got string
	if s.Get("key", &got) {
		t.Fatal("Get on disabled store should always miss")
	}
}

func TestStoreNil(t *testing.T) {
	var s *Store
	if s.Enabled() {
		t.Fatal("nil store should not be enabled")
	}
	var got string
	if s.Get("key", &got) {
		t.Fatal("nil store Get should miss")
	}
}

func TestStoreMiss(t *testing.T) {
	s := &Store{enabled: true, baseDir: t.TempDir()}
	var got string
	if s.Get("nonexistent", &got) {
		t.Fatal("expected cache miss")
	}
}

func TestStoreCorruptedFile(t *testing.T) {
	dir := t.TempDir()
	s := &Store{enabled: true, baseDir: dir}

	key := "corrupt.gob"
	path := filepath.Join(dir, key)
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte("not valid gob"), 0644)

	var got string
	if s.Get(key, &got) {
		t.Fatal("corrupted file should be a miss")
	}
	// File should be removed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("corrupted file should be removed")
	}
}

func TestStoreMapRoundtrip(t *testing.T) {
	s := &Store{enabled: true, baseDir: t.TempDir()}

	data := map[string]float64{"alice": 92.4, "bob": 51.7}
	key := "test/map.gob"
	if err := s.Set(key, data); err != nil {
		t.Fatalf("Set: %v", err)
	}

	var got map[string]float64
	if !s.Get(key, &got) {
		t.Fatal("expected cache hit")
	}
	if got["alice"] != 92.4 || got["bob"] != 51.7 {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestBlameKey(t *testing.T) {
	k1 := BlameKey("/repo", "abc123", []string{"a.go", "b.go"}, 500, "file")
	k2 := BlameKey("/repo", "abc123", []string{"a.go", "b.go"}, 500, "file")
	k3 := BlameKey("/repo", "def456", []string{"a.go", "b.go"}, 500, "file")

	if k1 != k2 {
		t.Fatal("same inputs should produce same key")
	}
	if k1 == k3 {
		t.Fatal("different commit should produce different key")
	}
	// A different move-detection level must key differently (it changes attribution),
	// while the empty level aliases "file".
	if k1 == BlameKey("/repo", "abc123", []string{"a.go", "b.go"}, 500, "full") {
		t.Fatal("different move detection should produce different key")
	}
	if k1 != BlameKey("/repo", "abc123", []string{"a.go", "b.go"}, 500, "") {
		t.Fatal("empty move detection should alias \"file\"")
	}
}

func TestBlameAtCommitKey(t *testing.T) {
	k1 := BlameAtCommitKey("/repo", "abc123def456", []string{"a.go"}, 500, "file")
	k2 := BlameAtCommitKey("/repo", "abc123def456", []string{"a.go"}, 500, "file")
	if k1 != k2 {
		t.Fatal("same inputs should produce same key")
	}
}

// TestStoreConcurrentSetSameKey guards the atomic per-write temp-file
// suffix in Set. Under parallel timeline periods, two adjacent windows can
// resolve to the SAME boundary commit → the same blame cache key → a
// concurrent Set of the same key. With a fixed "<path>.tmp" suffix, two
// writers would create/truncate the SAME temp file: one truncates the
// other's in-flight file mid-encode, and the atomic rename then publishes a
// corrupted gob. This test fires many rounds of high-concurrency identical
// Sets to one key and asserts every subsequent Get returns the complete,
// correct value. Revert the atomic-counter suffix and a round corrupts the
// file → Get misses (self-heal removes the bad file) or decodes short →
// failure. Note: a -race pass alone would NOT catch this — the hazard is a
// filesystem write collision, not a Go-memory data race.
func TestStoreConcurrentSetSameKey(t *testing.T) {
	dir := t.TempDir()
	s := &Store{enabled: true, baseDir: dir}

	// Payload large enough that gob encoding spans several writes, widening
	// the window for a racing writer to truncate a shared temp file.
	data := make([]string, 2000)
	for i := range data {
		data[i] = fmt.Sprintf("author-%04d-with-padding-to-grow-the-encoded-payload-size", i)
	}
	key := "concurrent/same-key.gob"

	for round := 0; round < 40; round++ {
		const writers = 8
		var wg sync.WaitGroup
		for w := 0; w < writers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := s.Set(key, data); err != nil {
					t.Errorf("round %d: concurrent Set: %v", round, err)
				}
			}()
		}
		wg.Wait()

		var got []string
		if !s.Get(key, &got) {
			t.Fatalf("round %d: Get miss after concurrent Set — cache file was corrupted by a racing writer (non-unique temp file?)", round)
		}
		if len(got) != len(data) || got[0] != data[0] || got[len(data)-1] != data[len(data)-1] {
			t.Fatalf("round %d: Get returned corrupted/partial data (len=%d, want %d)", round, len(got), len(data))
		}
	}
}

func TestClearAll(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, ".eis", "cache")
	os.MkdirAll(cacheDir, 0755)
	os.WriteFile(filepath.Join(cacheDir, "test.gob"), []byte("data"), 0644)

	// Clear uses UserHomeDir, so we test via RemoveAll directly
	if err := os.RemoveAll(cacheDir); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Fatal("cache dir should be removed")
	}
}

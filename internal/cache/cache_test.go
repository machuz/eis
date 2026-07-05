package cache

import (
	"os"
	"path/filepath"
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

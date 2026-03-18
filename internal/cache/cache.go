package cache

import (
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/machuz/engineering-impact-score/internal/git"
)

func init() {
	// Register types for gob encoding
	gob.Register([]git.BlameLine{})
	gob.Register([]git.Commit{})
	gob.Register([]string{})
	gob.Register(time.Time{})
}

// Store manages disk-based caching of expensive git operations.
// Cache is stored in ~/.eis/cache/ keyed by repo path hash.
type Store struct {
	enabled bool
	baseDir string
}

// New creates a cache store. If enabled is false, all Get calls return miss.
func New(enabled bool) *Store {
	home, err := os.UserHomeDir()
	if err != nil {
		return &Store{enabled: false}
	}
	return &Store{
		enabled: enabled,
		baseDir: filepath.Join(home, ".eis", "cache"),
	}
}

// NewWithDir creates a cache store with a custom base directory.
// If baseDir is empty, the default ~/.eis/cache is used.
func NewWithDir(enabled bool, baseDir string) *Store {
	if baseDir == "" {
		return New(enabled)
	}
	return &Store{enabled: enabled, baseDir: baseDir}
}

// Enabled returns whether the cache is active.
func (s *Store) Enabled() bool {
	return s != nil && s.enabled
}

// Get loads a cached value. Returns true on hit.
func (s *Store) Get(key string, dest interface{}) bool {
	if !s.Enabled() {
		return false
	}
	path := filepath.Join(s.baseDir, key)
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	if err := gob.NewDecoder(f).Decode(dest); err != nil {
		// Corrupted cache — remove it
		os.Remove(path)
		return false
	}
	return true
}

// Set stores a value to cache with atomic write.
func (s *Store) Set(key string, data interface{}) error {
	if !s.Enabled() {
		return nil
	}
	path := filepath.Join(s.baseDir, key)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	if err := gob.NewEncoder(f).Encode(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()

	return os.Rename(tmp, path)
}

// Clear removes cache for a specific repo, or all caches if repoPath is empty.
func Clear(repoPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	baseDir := filepath.Join(home, ".eis", "cache")

	if repoPath == "" {
		return os.RemoveAll(baseDir)
	}

	repoKey := hashString(repoPath)
	return os.RemoveAll(filepath.Join(baseDir, repoKey))
}

// CacheSize returns the total size of cache in bytes.
func CacheSize() (int64, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}
	baseDir := filepath.Join(home, ".eis", "cache")

	var total int64
	filepath.Walk(baseDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, nil
}

// --- Cache key builders ---

// repoHash returns a short hash for the repo path to namespace cache entries.
func repoHash(repoPath string) string {
	abs, err := filepath.Abs(repoPath)
	if err != nil {
		abs = repoPath
	}
	return hashString(abs)[:12]
}

// BlameKey returns the cache key for ConcurrentBlameFiles results.
// Uses HEAD commit hash — invalidated when HEAD moves.
func BlameKey(repoPath, headHash string, files []string, sampleSize int) string {
	filesHash := hashFileList(files, sampleSize)
	return filepath.Join(repoHash(repoPath), "blame", shortHash(headHash), filesHash+".gob")
}

// BlameAtCommitKey returns the cache key for blame at a specific commit.
// Immutable: blame at a fixed commit never changes.
func BlameAtCommitKey(repoPath, commitHash string, files []string, sampleSize int) string {
	filesHash := hashFileList(files, sampleSize)
	return filepath.Join(repoHash(repoPath), "blame-commit", shortHash(commitHash), filesHash+".gob")
}

// LogKey returns the cache key for ParseLog results.
func LogKey(repoPath, headHash string) string {
	return filepath.Join(repoHash(repoPath), "log", shortHash(headHash)+".gob")
}

// MergeLogKey returns the cache key for ParseMergeCommits results.
func MergeLogKey(repoPath, headHash string) string {
	return filepath.Join(repoHash(repoPath), "merge-log", shortHash(headHash)+".gob")
}

// DebtKey returns the cache key for CalcDebt results.
// Keyed by the fix commit hashes used.
func DebtKey(repoPath string, fixCommitHashes []string) string {
	h := sha256.New()
	for _, hash := range fixCommitHashes {
		h.Write([]byte(hash))
	}
	key := fmt.Sprintf("%x", h.Sum(nil))[:16]
	return filepath.Join(repoHash(repoPath), "debt", key+".gob")
}

// --- helpers ---

func shortHash(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:])
}

func hashFileList(files []string, sampleSize int) string {
	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("sample=%d\n", sampleSize)))
	h.Write([]byte(strings.Join(sorted, "\n")))
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

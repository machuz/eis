// Package cache re-exports the internal cache types for external consumers.
package cache

import "github.com/machuz/eis/v2/internal/cache"

// Store is a disk-based cache for expensive git operations.
type Store = cache.Store

var (
	// New creates a cache store with the default directory (~/.eis/cache).
	New = cache.New

	// NewWithDir creates a cache store with a custom base directory.
	NewWithDir = cache.NewWithDir

	// BlameKey returns the cache key for ConcurrentBlameFiles results.
	BlameKey = cache.BlameKey

	// LogKey returns the cache key for ParseLog results.
	LogKey = cache.LogKey

	// MergeLogKey returns the cache key for ParseMergeCommits results.
	MergeLogKey = cache.MergeLogKey

	// DebtKey returns the cache key for CalcDebt results.
	DebtKey = cache.DebtKey

	// Clear removes cache for a specific repo, or all caches if repoPath is empty.
	Clear = cache.Clear

	// CacheSize returns the total size of cache in bytes.
	CacheSize = cache.CacheSize
)

package cli

import (
	"fmt"

	"github.com/machuz/eis/v2/internal/cache"
)

func runCache(args []string) error {
	if len(args) == 0 {
		fmt.Println(`Usage:
  eis cache clear    Clear all cached data
  eis cache status   Show cache size`)
		return nil
	}

	switch args[0] {
	case "clear":
		if err := cache.Clear(""); err != nil {
			return fmt.Errorf("clear cache: %w", err)
		}
		fmt.Println("Cache cleared.")
		return nil

	case "status":
		size, err := cache.CacheSize()
		if err != nil {
			return fmt.Errorf("cache status: %w", err)
		}
		if size == 0 {
			fmt.Println("Cache is empty.")
		} else if size < 1024*1024 {
			fmt.Printf("Cache size: %.1f KB\n", float64(size)/1024)
		} else {
			fmt.Printf("Cache size: %.1f MB\n", float64(size)/(1024*1024))
		}
		return nil

	default:
		return fmt.Errorf("unknown cache command: %s (use clear or status)", args[0])
	}
}

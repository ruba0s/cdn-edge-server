package cache

import (
	"cdn-edge-server/internal/config"
	"fmt"
	"os"
	"path/filepath"
)

const (
	MaxCacheFiles = 5
)

// var CacheDir = "/Users/Ruba/Dev/cdn-edge-server/internal/cache/"

var queue []string                  // FIFO queue
var present = make(map[string]bool) // filename → bool (is present?)

// Init loads existing files into FIFO based on alphabetical order.
func Init() {
	files, _ := os.ReadDir(config.CacheDir)
	for _, f := range files {
		name := f.Name()

		if name == ".gitkeep" {
			continue // ignore git file (not part of edge server cache)
		}

		queue = append(queue, name)
		present[name] = true
	}
	fmt.Println("DEBUG QUEUE:", queue)
}

// Has checks if the file with the given name is present in the cache.
func Has(name string) bool {
	return present[name]
}

// Get reads and returns the given filename from the cache (only called in case of cache hit).
func Get(name string) ([]byte, error) {
	return os.ReadFile(config.CacheDir + name)
}

// Add adds the file with the given name to the cache.
func Add(name string, data []byte) error {
	// Cannot write git files to cache or server storage
	if name == ".gitkeep" {
		return fmt.Errorf(".gitkeep cannot be added to server storage")
	}

	// Write file
	err := os.WriteFile(filepath.Join(config.CacheDir, name), data, 0644)
	if err != nil {
		return err
	}

	// Register in metadata
	queue = append(queue, name)
	present[name] = true

	// Eviction check
	fmt.Println("DEBUG QUEUE:", queue)
	if len(queue) > MaxCacheFiles {
		evict()
	}

	return nil
}

// evict removes the last file in the cache's internal FIFO structure from the cache.
func evict() {
	oldest := queue[0]
	queue = queue[1:]       // pop front of queue
	delete(present, oldest) // mark popped file as unpresent in queue
	os.Remove(config.CacheDir + oldest)
}

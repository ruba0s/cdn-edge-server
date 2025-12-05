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
	return os.ReadFile(filepath.Join(config.CacheDir, name))
}

// Add adds the file with the given name to the cache.
func Add(name string, data []byte) error {
	// Cannot write git files to cache or server storage
	if name == ".gitkeep" {
		return fmt.Errorf(".gitkeep cannot be added to server storage")
	}

	// If file is already in cache, overwrite
	if present[name] {
		return os.WriteFile(filepath.Join(config.CacheDir, name), data, 0644)
	}

	// Eviction check (to ensure queue size remains within the max cache size)
	fmt.Println("DEBUG QUEUE:", queue)
	if len(queue) >= MaxCacheFiles {
		evict()
	}

	// Write file
	err := os.WriteFile(filepath.Join(config.CacheDir, name), data, 0644)
	if err != nil {
		return err
	}

	// Register in metadata
	queue = append(queue, name)
	present[name] = true

	return nil
}

// evict removes the last file in the cache's internal FIFO structure from the cache.
func evict() {
	oldest := queue[0]
	queue = queue[1:]       // pop front of queue
	delete(present, oldest) // mark popped file as unpresent in queue
	os.Remove(filepath.Join(config.CacheDir, oldest))
}

// Remove removes the file with the given name from the cache, if present.
func Remove(filename string) {
	if !present[filename] {
		return
	}

	// Remove from queue
	for i, f := range queue {
		if f == filename {
			queue = append(queue[:i], queue[i+1:]...)
			break
		}
	}

	// Remove from present map
	delete(present, filename)

	// Delete file from disk
	os.Remove(filepath.Join(config.CacheDir, filename))
}

// CacheContent returns a copy of the cache queue (list of cached filenames in order)
func CacheContent() []string {
	result := make([]string, len(queue))
	copy(result, queue)
	return result
}

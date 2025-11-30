package config

import (
	"os"
	"path/filepath"
)

var (
	ProjectRoot string
	CacheDir    string
	StorageDir  string
	OriginHost  = "127.0.0.1"
	OriginPort  = "4396"
)

func init() {
	// Find absolute project root at runtime
	wd, _ := os.Getwd()
	ProjectRoot = findProjectRoot(wd)

	CacheDir = filepath.Join(ProjectRoot, "internal/cache/files")
	StorageDir = filepath.Join(ProjectRoot, "internal/storage/files")
}

// Walk up directories until we find go.mod
func findProjectRoot(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// fallback: original directory
			return start
		}
		dir = parent
	}
}

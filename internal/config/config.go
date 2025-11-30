package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var (
	ProjectRoot string
	CacheDir    string
	StorageDir  string

	EdgeHost   string
	EdgePort   string
	OriginHost string
	OriginPort string
)

func init() {
	// Find project root
	wd, _ := os.Getwd()
	ProjectRoot = findProjectRoot(wd)

	// Load .env
	envErr := godotenv.Load(filepath.Join(ProjectRoot, ".env"))
	if envErr != nil {
		panic("missing .env file, please create one from .env.template.")
	}

	// Get required env variables
	EdgeHost = getReqEnvVar("EDGE_HOST")
	EdgePort = getReqEnvVar("EDGE_PORT")
	OriginHost = getReqEnvVar("ORIGIN_HOST")
	OriginPort = getReqEnvVar("ORIGIN_PORT")

	// Optionally override optional env variables, or use defaults
	CacheDir = getOptEnvVar("CACHE_DIR",
		filepath.Join(ProjectRoot, "internal/cache/files"),
	)
	StorageDir = getOptEnvVar("STORAGE_DIR",
		filepath.Join(ProjectRoot, "internal/storage/files"),
	)
}

func findProjectRoot(start string) string {
	// Walk up directories until we find go.mod
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

func getReqEnvVar(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Missing required environment variable: %s", key))
	}
	return value
}

func getOptEnvVar(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

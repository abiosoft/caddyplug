package internal

import (
	"os"
	"path/filepath"
)

// PluginsDir is the directory for built plugins.
func PluginsDir() string {
	return filepath.Join(LibDir(), "plugins")
}

// LibDir is the directory for caddy plugin loader resources.
func LibDir() string {
	return filepath.Join(os.Getenv("HOME"), "lib", "caddy")
}

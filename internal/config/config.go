package config

import (
	"os"
	"path/filepath"
)

// Root returns the canonical root directory for ansimotd.
// Priority: ANSIMOTD_DIR > XDG_CONFIG_HOME/ansimotd > ~/.config/ansimotd
func Root() string {
	if dir := os.Getenv("ANSIMOTD_DIR"); dir != "" {
		return dir
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "ansimotd")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ansimotd")
}

// ArtDir returns the directory where art packs are stored.
func ArtDir() string {
	return filepath.Join(Root(), "art")
}

// LastFile returns the path to the state file that records the last displayed file.
func LastFile() string {
	return filepath.Join(Root(), "last")
}

// EnsureDirs creates the root and art directories if they don't exist.
func EnsureDirs() error {
	return os.MkdirAll(ArtDir(), 0o755)
}

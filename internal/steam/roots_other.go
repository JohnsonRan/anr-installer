//go:build !windows

package steam

import (
	"os"
	"path/filepath"
)

// steamRoots returns the folders where Steam may be installed: native, Flatpak and the Steam Deck's.
func steamRoots() []string {
	home, _ := os.UserHomeDir()
	if home == "" {
		return nil
	}
	return []string{
		filepath.Join(home, ".steam", "steam"),
		filepath.Join(home, ".steam", "root"),
		filepath.Join(home, ".local", "share", "Steam"),
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".local", "share", "Steam"),
		filepath.Join(home, "Library", "Application Support", "Steam"),
	}
}

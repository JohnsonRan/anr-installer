// Package paths says where the installer keeps its files: in the user's folders, or for the portable version
// in its own folder, which a file named "portable" next to the executable marks.
package paths

import (
	"os"
	"path/filepath"
	"sync"
)

const name = "anr-installer"

var portable = sync.OnceValue(func() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return portableDir(filepath.Dir(exe))
})

// portableDir returns dir when it holds the "portable" marker, else "".
func portableDir(dir string) string {
	if fi, err := os.Stat(filepath.Join(dir, "portable")); err != nil || fi.IsDir() {
		return ""
	}
	return dir
}

// Portable returns the portable version's folder, or "" for the installed version.
func Portable() string { return portable() }

// Config is the folder of the user's settings: %APPDATA%\anr-installer, or the portable version's folder.
func Config() string {
	if dir := Portable(); dir != "" {
		return dir
	}
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, name)
}

// Cache is the folder of downloads and what can be fetched again: %LOCALAPPDATA%\anr-installer, or cache in
// the portable version's folder.
func Cache() string {
	if dir := Portable(); dir != "" {
		return filepath.Join(dir, "cache")
	}
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, name)
}

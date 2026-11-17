package steam

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// steamRoots returns the folders where Steam may be installed.
func steamRoots() []string {
	var roots []string
	for _, k := range []struct {
		root registry.Key
		path string
		name string
	}{
		{registry.CURRENT_USER, `Software\Valve\Steam`, "SteamPath"},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, "InstallPath"},
		{registry.LOCAL_MACHINE, `SOFTWARE\Valve\Steam`, "InstallPath"},
	} {
		key, err := registry.OpenKey(k.root, k.path, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		if v, _, err := key.GetStringValue(k.name); err == nil && v != "" {
			roots = append(roots, filepath.FromSlash(v))
		}
		key.Close()
	}
	for _, env := range []string{"ProgramFiles(x86)", "ProgramFiles"} {
		if p := os.Getenv(env); p != "" {
			roots = append(roots, filepath.Join(p, "Steam"))
		}
	}
	return roots
}

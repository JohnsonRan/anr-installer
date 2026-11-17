package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPortableMarker(t *testing.T) {
	dir := t.TempDir()
	if portableDir(dir) != "" {
		t.Error("a folder without the marker is portable")
	}
	os.Mkdir(filepath.Join(dir, "portable"), 0o755)
	if portableDir(dir) != "" {
		t.Error("a folder named portable is taken for the marker")
	}
	other := t.TempDir()
	os.WriteFile(filepath.Join(other, "portable"), nil, 0o644)
	if portableDir(other) != other {
		t.Error("the marker is not seen")
	}
}

func TestInstalledFolders(t *testing.T) {
	// The test binary runs from a temporary folder without the marker: the installed version's folders.
	t.Setenv("APPDATA", `C:\U\Roaming`)
	t.Setenv("LOCALAPPDATA", `C:\U\Local`)
	t.Setenv("XDG_CONFIG_HOME", "/u/config")
	t.Setenv("XDG_CACHE_HOME", "/u/cache")
	if Portable() != "" {
		t.Fatal("the test binary is portable")
	}
	if filepath.Base(Config()) != name || filepath.Base(Cache()) != name {
		t.Errorf("config %s, cache %s", Config(), Cache())
	}
}

package pkg

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// Deleting the downloads keeps versions.json and the comparisons, and a package is unpacked again afterwards.
func TestRemoveDownloads(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir()) // keep the cache out of the user's
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	zipPath := filepath.Join(DownloadsDir(), "anr-fg-V.B19-full.zip")
	writeZip(t, zipPath, map[string]string{"install.ps1": "# install"})
	if _, err := Extract(zipPath); err != nil {
		t.Fatal(err)
	}
	keep := []string{filepath.Join(CacheDir(), "versions.json"), filepath.Join(CacheDir(), "compare", "rec", "anr.mp4")}
	for _, p := range keep {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("kept"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	size := DownloadsSize()
	zi, _ := os.Stat(zipPath)
	if size <= zi.Size() {
		t.Fatalf("size %d should count the zip (%d) and the package unpacked from it", size, zi.Size())
	}

	freed, err := RemoveDownloads()
	if err != nil || freed != size {
		t.Fatalf("freed %d of %d: %v", freed, size, err)
	}
	for _, dir := range []string{DownloadsDir(), packagesDir()} {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("%s is left: %v", dir, err)
		}
	}
	for _, p := range keep {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s should stay: %v", p, err)
		}
	}
	if DownloadsSize() != 0 {
		t.Errorf("size after deleting: %d", DownloadsSize())
	}

	writeZip(t, zipPath, map[string]string{"install.ps1": "# install"})
	dir, err := Extract(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "install.ps1")); err != nil {
		t.Errorf("the package was not unpacked again: %v", err)
	}
}

func writeZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

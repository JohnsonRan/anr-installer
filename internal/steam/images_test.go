package steam

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPick(t *testing.T) {
	files := []string{
		`c/1/aa/library_hero_blur_schinese.jpg`,
		`c/1/aa/library_hero_schinese.jpg`,
		`c/1/bb/library_hero_english.jpg`,
		`c/1/cc/logo.png`,
		`c/1/dd/library_header_french.jpg`,
	}
	if got := pick(files, "library_hero", ".jpg", "schinese"); filepath.Base(got) != "library_hero_schinese.jpg" {
		t.Errorf("hero in Chinese: %s", got)
	}
	if got := pick(files, "library_hero", ".jpg", "german"); got == "" || filepath.Base(got) == "library_hero_blur_schinese.jpg" {
		t.Errorf("hero in another language: %s", got)
	}
	if got := pick(files, "logo", ".png", "schinese"); filepath.Base(got) != "logo.png" {
		t.Errorf("logo without a language: %s", got)
	}
	if got := pick(files, "library_header", ".jpg", "english"); filepath.Base(got) != "library_header_french.jpg" {
		t.Errorf("header in any language: %s", got)
	}
}

func TestListFilesNested(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{"3059e94fb89816fe1929a238ac2b55e8be348408.jpg", "5daf/library_hero_schinese.jpg", "14f4/logo_schinese.png"} {
		full := filepath.Join(dir, filepath.FromSlash(p))
		os.MkdirAll(filepath.Dir(full), 0o755)
		os.WriteFile(full, []byte("x"), 0o644)
	}
	files := listFiles(dir, 2)
	if len(files) != 3 {
		t.Fatalf("files %v", files)
	}
	if got := pick(files, "library_hero", ".jpg", "schinese"); got == "" {
		t.Error("no hero in a hash folder")
	}
}

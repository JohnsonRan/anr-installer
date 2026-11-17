package steam

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLibraryFolders(t *testing.T) {
	doc, err := ParseVDF(`"libraryfolders"
{
	"0"
	{
		"path"		"C:\\Program Files (x86)\\Steam"
		"apps" { "228980" "123" }
	}
	"1"
	{
		"path"		"D:\\SteamLibrary"
	}
}`)
	if err != nil {
		t.Fatal(err)
	}
	lf := doc.Get("LibraryFolders")
	if lf == nil || len(lf.Order) != 2 {
		t.Fatalf("libraryfolders: %+v", lf)
	}
	if got := lf.Get("1").Str("path"); got != `D:\SteamLibrary` {
		t.Errorf("path %q", got)
	}
	if got := lf.Get("0").Get("apps").Str("228980"); got != "123" {
		t.Errorf("apps %q", got)
	}
}

func TestGames(t *testing.T) {
	lib := t.TempDir()
	apps := filepath.Join(lib, "steamapps")
	must(t, os.MkdirAll(filepath.Join(apps, "common", "Cyber Game"), 0o755))
	must(t, os.MkdirAll(filepath.Join(apps, "common", "Proton 9.0"), 0o755))
	must(t, os.WriteFile(filepath.Join(apps, "appmanifest_1.acf"), []byte(`"AppState" { "appid" "1" "name" "Cyber Game" "installdir" "Cyber Game" }`), 0o644))
	must(t, os.WriteFile(filepath.Join(apps, "appmanifest_2.acf"), []byte(`"AppState" { "appid" "2" "name" "Proton 9.0" "installdir" "Proton 9.0" }`), 0o644))
	must(t, os.WriteFile(filepath.Join(apps, "appmanifest_3.acf"), []byte(`"AppState" { "appid" "3" "name" "Gone" "installdir" "Gone" }`), 0o644))
	games := Games([]string{lib})
	if len(games) != 1 || games[0].Name != "Cyber Game" || games[0].AppID != "1" {
		t.Fatalf("games %+v", games)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

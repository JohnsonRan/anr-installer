package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if d := Load(); len(d.Games) != 0 || d.ExeDirs == nil {
		t.Fatalf("empty: %+v", d)
	}
	d := &Data{Games: []Game{{Name: "Epic Game", Dir: `D:\Games\Epic`, Named: true}}, ExeDirs: map[string]string{Key(`D:\S\Game`): `D:\S\Game\bin`}, CloseToTray: true}
	if err := d.Save(); err != nil {
		t.Fatal(err)
	}
	got := Load()
	if len(got.Games) != 1 || got.Games[0] != d.Games[0] || got.ExeDirs[Key(`d:\s\game`)] != `D:\S\Game\bin` || !got.CloseToTray {
		t.Fatalf("loaded %+v", got)
	}
}

// A file from before the setting existed loads with closing quitting, as it did.
func TestLoadWithoutCloseToTray(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := os.MkdirAll(filepath.Dir(Path()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(), []byte(`{"games":[{"name":"Epic Game","dir":"D:\\Epic"}],"seen_release":"FG-V.B19"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if d := Load(); d.CloseToTray || len(d.Games) != 1 || d.SeenRelease != "FG-V.B19" {
		t.Fatalf("loaded %+v", d)
	}
}

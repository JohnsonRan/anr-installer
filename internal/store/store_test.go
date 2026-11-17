package store

import "testing"

func TestRoundTrip(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if d := Load(); len(d.Games) != 0 || d.ExeDirs == nil {
		t.Fatalf("empty: %+v", d)
	}
	d := &Data{Games: []Game{{Name: "Epic Game", Dir: `D:\Games\Epic`, Named: true}}, ExeDirs: map[string]string{Key(`D:\S\Game`): `D:\S\Game\bin`}}
	if err := d.Save(); err != nil {
		t.Fatal(err)
	}
	got := Load()
	if len(got.Games) != 1 || got.Games[0] != d.Games[0] || got.ExeDirs[Key(`d:\s\game`)] != `D:\S\Game\bin` {
		t.Fatalf("loaded %+v", got)
	}
}

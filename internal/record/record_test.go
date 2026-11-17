package record

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"anr-installer/internal/detect"
	"anr-installer/internal/pkg"
)

func release(tag string, assets ...string) *pkg.Release {
	r := &pkg.Release{Tag: tag, Assets: map[string]pkg.Asset{}}
	for _, a := range assets {
		r.Assets[a] = pkg.Asset{Name: a}
	}
	return r
}

func TestFindTakesTheNewestPackageWithSums(t *testing.T) {
	rels := []*pkg.Release{
		release("FG-V.B19", "anr-fg-V.B19-full.zip", "nr-record-1.3-windows.zip"), // no sums: cannot be checked
		release("FG-V.B18", "nr-record-1.2-windows.zip", "nr-record-1.2-SHA256SUMS.txt", "nr-record-1.2-linux.tar.gz",
			"nr-record-1.10-windows.zip", "nr-record-1.10-SHA256SUMS.txt"),
		release("FG-V.B17", "nr-record-1.1-windows.zip", "nr-record-1.1-SHA256SUMS.txt"),
	}
	p := Find(rels)
	if p == nil || p.Rel.Tag != "FG-V.B18" || p.Version != "1.10" || p.Asset() != "nr-record-1.10-windows.zip" {
		t.Fatalf("found %+v", p)
	}
	if p := Find(rels[:1]); p != nil {
		t.Errorf("a package without sums was found: %+v", p)
	}
}

func TestRecordingsNewestFirst(t *testing.T) {
	dir := t.TempDir()
	write := func(name, json string) {
		os.MkdirAll(filepath.Join(dir, name), 0o755)
		if json != "" {
			os.WriteFile(filepath.Join(dir, name, "record.json"), []byte(json), 0o644)
		}
	}
	write("20261111-165107-game", `{"format":"nr-record-4","feature":"fg","game":"C:/Games/Game.exe","time":"2026-11-11T16:51:07","width":3840,"height":2160,"frames":600,"multiplier":2,"status":"complete"}`)
	write("20261112-100546-game-sr", `{"format":"nr-record-4","feature":"sr","game":"game.exe"}`)
	write("20261110-090000-old", `{"format":"nr-record-3","game":"other.exe"}`)
	write("not-a-recording", "")
	write("20261113-000000-broken", `{`)

	recs := Recordings(dir)
	var names []string
	for _, r := range recs {
		names = append(names, r.Name)
	}
	if got := strings.Join(names, " "); got != "20261112-100546-game-sr 20261111-165107-game 20261110-090000-old" {
		t.Fatalf("recordings %q", got)
	}
	sr, fg, old := recs[0], recs[1], recs[2]
	if sr.FrameGeneration() || !fg.FrameGeneration() || !old.FrameGeneration() {
		t.Error("only fg recordings, and those before format 4, hold frame generation")
	}
	if !fg.Of(`D:\SteamLibrary\Game\bin\game.exe`) || !sr.Of("GAME.EXE") || old.Of("game.exe") || fg.Of("") {
		t.Error("a recording is of the game whose executable has its name")
	}
	if fg.Width != 3840 || fg.Frames != 600 || fg.Time.Hour() != 16 {
		t.Errorf("read %+v", fg)
	}
}

func TestLookFindsAShimOffThePath(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "nr-test-tool.cmd"), []byte("@echo off\r\n"), 0o755)
	if got := look("nr-test-tool", []string{t.TempDir(), dir}); got != filepath.Join(dir, "nr-test-tool.cmd") {
		t.Errorf("found %q", got)
	}
	if got := look("nr-test-missing", []string{dir}); got != "" {
		t.Errorf("found %q", got)
	}
}

// TestRealInstall installs the real nr-record package into a fake game folder and removes it. Set ANR_PKG_DIR
// to a folder that holds nr-record-*-windows.zip.
func TestRealInstall(t *testing.T) {
	zips, _ := filepath.Glob(filepath.Join(os.Getenv("ANR_PKG_DIR"), assetPrefix+"*"+assetSuffix))
	if os.Getenv("ANR_PKG_DIR") == "" || len(zips) == 0 {
		t.Skip("ANR_PKG_DIR not set, or it holds no nr-record package")
	}
	t.Setenv("LOCALAPPDATA", t.TempDir()) // keep the cache out of the user's
	downloads := filepath.Join(pkg.CacheDir(), "downloads")
	os.MkdirAll(downloads, 0o755)
	data, err := os.ReadFile(zips[0])
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(downloads, filepath.Base(zips[0])), data, 0o644)

	tools, err := Unpack(context.Background(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"tools/nr-replay.exe", "tools/nr-video.exe", "viewer/renderer.exe"} {
		if _, err := os.Stat(filepath.Join(tools, f)); err != nil {
			t.Error(err)
		}
	}

	game := t.TempDir()
	os.MkdirAll(filepath.Join(game, "OptiScaler"), 0o755)
	ini := filepath.Join(game, "OptiScaler.ini")
	os.WriteFile(ini, []byte("[FrameGen]\r\nFGNvngxReplacement=Apex Neural Rendering FG\r\n"), 0o644)
	say := func(line string) { t.Log(line) }
	if err := pkg.Run(context.Background(), tools, pkg.Install, game, say); err != nil {
		t.Fatal(err)
	}
	if !Installed(game) || detect.IniValue(ini, "FrameGen", "FGNvngxReplacement") != Dir {
		t.Fatal("nr-record is not installed")
	}
	if err := pkg.Run(context.Background(), tools, pkg.Uninstall, game, say); err != nil {
		t.Fatal(err)
	}
	if Installed(game) || detect.IniValue(ini, "FrameGen", "FGNvngxReplacement") != "Apex Neural Rendering FG" {
		t.Fatal("nr-record is not removed, or OptiScaler.ini is not restored")
	}
}

func TestMadeStillsPairsTheFramesBothSidesHave(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())
	c := &Comparison{Recording: &Recording{Name: "20261117-210405-delta"}, A: ANR("a.dll"), B: DLSSG("b.dll")}
	if c.MadeStills(5) != nil {
		t.Fatal("stills before any were made")
	}
	replay := `{"fps":120,"multiplier":2,"frames":[
		{"kind":"real","real":4,"file":"real_f4-f4_t0.000.nrraw"},
		{"kind":"generated","real":5,"file":"generated_f4-f5_t0.500.nrraw"},
		{"kind":"real","real":5,"file":"real_f5-f5_t0.000.nrraw"}]}`
	for side, pngs := range map[string][]string{
		"anr":   {"real_f4-f4_t0.000", "generated_f4-f5_t0.500", "real_f5-f5_t0.000"},
		"dlssg": {"generated_f4-f5_t0.500", "real_f5-f5_t0.000"},
	} {
		dir := filepath.Join(c.Out(), "stills", "5", side)
		os.MkdirAll(dir, 0o755)
		os.WriteFile(filepath.Join(dir, "replay.json"), []byte(replay), 0o644)
		for _, p := range pngs {
			os.WriteFile(filepath.Join(dir, p+".png"), []byte("png"), 0o644)
		}
	}
	stills := c.MadeStills(5)
	if len(stills) != 2 || stills[0].Name != "generated_f4-f5_t0.500" || !stills[0].Frame.Is(Frame{Kind: "generated"}) || stills[0].Frame.Is(stills[1].Frame) || stills[1].Name != "real_f5-f5_t0.000" ||
		filepath.Base(filepath.Dir(stills[0].B)) != "dlssg" {
		t.Fatalf("stills %+v", stills)
	}
}

// A comparison made again must not keep the stills of the videos it replaces.
func TestRemoveStills(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())
	c := &Comparison{Recording: &Recording{Name: "rec"}, A: ANR("a.dll"), B: DLSSG("b.dll")}
	dir := filepath.Join(c.Out(), "stills", "5", "anr")
	os.MkdirAll(dir, 0o755)
	c.RemoveStills(5)
	if _, err := os.Stat(filepath.Join(c.Out(), "stills", "5")); err == nil {
		t.Error("the stills are still there")
	}
}

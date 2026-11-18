package gui

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
	"anr-installer/internal/detect"
	"anr-installer/internal/pkg"
)

const (
	baseINI   = "[FrameGen]\nFGInput=nvngxfg\nFGNvngxReplacement=Apex Neural Rendering FG\n[Menu]\nScale=auto\nShortcutKey=auto\n"
	theirsINI = "; the new file\r\n[FrameGen]\r\nFGInput=nvngxfg\r\nFGNvngxReplacement=Apex Neural Rendering FG\r\n[Menu]\r\nScale=auto\r\nShortcutKey=0x24\r\n"
	mineINI   = "[FrameGen]\nFGInput=nvngxfg\nFGNvngxReplacement=nr-record\n[Menu]\nScale=1.5\nShortcutKey=0x2D\n"
)

// mergeGame is a game folder whose OptiScaler.ini an install has just replaced with theirsINI, and the release of
// V.B18, whose downloaded full package ships baseINI.
func mergeGame(t *testing.T) (a *app, r *reporter, dir string, m mainThread, base *pkg.Release) {
	a = testApp(t)
	m = onMainThread(a)
	dir = t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "OptiScaler.ini"), []byte(theirsINI), 0o644); err != nil {
		t.Fatal(err)
	}
	name := "anr-fg-V.B18-full.zip"
	if err := os.MkdirAll(pkg.DownloadsDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(pkg.DownloadsDir(), name))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create("game/OptiScaler.ini")
	w.Write([]byte(baseINI))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()
	base = &pkg.Release{Tag: "FG-V.B18", Version: "V.B18", Assets: map[string]pkg.Asset{name: {Name: name}}}
	return a, &reporter{a: a, tk: a.log.Begin("Delta · Upgrade to V.B19", time.Now())}, dir, m, base
}

func readINI(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "OptiScaler.ini"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// keep runs keepSettings in the background, doing the main thread's work until it returns.
func keep(t *testing.T, a *app, m mainThread, run func() error, while func()) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- run() }()
	var err error
	m.run(t, func() bool {
		if while != nil {
			while()
		}
		select {
		case err = <-done:
			return true
		default:
			return false
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	m.run(t, func() bool { return len(m) == 0 })
}

// In a batch, the user's values stay without a question, the conflicting one with a warning, but not those ANR FG
// needs; the log says which.
func TestKeepSettingsInABatch(t *testing.T) {
	a, r, dir, m, base := mergeGame(t)
	keep(t, a, m, func() error {
		return a.keepSettings(context.Background(), r, dir, []byte(mineINI), base, "V.B19", true)
	}, nil)
	want := "; the new file\r\n[FrameGen]\r\nFGInput=nvngxfg\r\nFGNvngxReplacement=Apex Neural Rendering FG\r\n[Menu]\r\nScale=1.5\r\nShortcutKey=0x2D\r\n"
	if got := readINI(t, dir); got != want {
		t.Errorf("OptiScaler.ini:\n%q\nwant\n%q", got, want)
	}
	if a.merge != nil {
		t.Error("a batch should not ask")
	}
	log := a.log.Text()
	for _, s := range []string{
		i18n.T("merge.log.kept", "Scale, ShortcutKey (new: 0x24)"),
		i18n.T("merge.log.pinned", "FGNvngxReplacement"),
	} {
		if !strings.Contains(log, s) {
			t.Errorf("no %q in the log:\n%s", s, log)
		}
	}
}

// Without the file the user's came from, a batch cannot tell what the user changed: it keeps the new file.
func TestKeepSettingsInABatchWithoutBase(t *testing.T) {
	a, r, dir, m, _ := mergeGame(t)
	keep(t, a, m, func() error { return a.keepSettings(context.Background(), r, dir, []byte(mineINI), nil, "V.B19", true) }, nil)
	if got := readINI(t, dir); got != theirsINI {
		t.Errorf("the new file should stay:\n%q", got)
	}
	log := a.log.Text()
	if !strings.Contains(log, i18n.T("merge.log.unknown")) || !strings.Contains(log, i18n.T("merge.log.batch.nobase")) {
		t.Errorf("the log should say why the new file stays:\n%s", log)
	}
}

// One install asks about the values both changed, and writes what the user chose.
func TestKeepSettingsAsks(t *testing.T) {
	a, r, dir, m, base := mergeGame(t)
	var tt *ui.Tester
	keep(t, a, m, func() error {
		return a.keepSettings(context.Background(), r, dir, []byte(mineINI), base, "V.B19", false)
	}, func() {
		if a.merge == nil || tt != nil {
			return
		}
		if len(a.merge.result.Conflicts) != 1 || len(a.merge.result.Kept) != 1 {
			t.Fatalf("conflicts %+v, kept %+v", a.merge.result.Conflicts, a.merge.result.Kept)
		}
		tt = ui.NewTester(a.view, 1040, 900)
		if !tt.HasText(i18n.T("merge.detail", "V.B19")) || !tt.HasText(i18n.T("merge.mine", "0x2D")) ||
			!tt.HasText(i18n.T("merge.pinned", "FGNvngxReplacement", "Apex Neural Rendering FG", "nr-record")) {
			t.Fatalf("the dialog should show the conflict and the pinned value: %q", tt.Texts())
		}
		if err := tt.Click(i18n.T("merge.new", "0x24")); err != nil {
			t.Fatal(err)
		}
		if err := tt.Click(i18n.T("merge.apply")); err != nil {
			t.Fatal(err)
		}
	})
	if got := readINI(t, dir); !strings.Contains(got, "Scale=1.5\r\n") || !strings.Contains(got, "ShortcutKey=0x24\r\n") {
		t.Errorf("the file should hold the kept value and the choice:\n%s", got)
	}
	if tt.Frame(); a.merge != nil || tt.HasText(i18n.T("merge.title")) {
		t.Error("the dialog should close")
	}
}

// Without the base, the dialog says it cannot tell what the user changed. Escape saves the choices as they are:
// the user's values.
func TestKeepSettingsWithoutBaseAsksAndEscapeKeeps(t *testing.T) {
	a, r, dir, m, _ := mergeGame(t)
	asked := false
	keep(t, a, m, func() error {
		return a.keepSettings(context.Background(), r, dir, []byte(mineINI), nil, "V.B19", false)
	}, func() {
		if a.merge == nil || asked {
			return
		}
		asked = true
		tt := ui.NewTester(a.view, 1040, 900)
		if !tt.HasText(i18n.T("merge.detail.nobase", "V.B19")) {
			t.Fatalf("the dialog should say why it asks about every difference: %q", tt.Texts())
		}
		tt.Key(0, ui.KeyEscape)
	})
	if got := readINI(t, dir); !strings.Contains(got, "Scale=1.5\r\n") || !strings.Contains(got, "ShortcutKey=0x2D\r\n") {
		t.Errorf("Escape should keep the user's values:\n%s", got)
	}
}

// Nothing the user changed: the file stays the new one, and the log says nothing of it.
func TestKeepSettingsWithNothingChanged(t *testing.T) {
	a, r, dir, _, base := mergeGame(t)
	if err := a.keepSettings(context.Background(), r, dir, []byte(baseINI), base, "V.B19", false); err != nil {
		t.Fatal(err)
	}
	if got := readINI(t, dir); got != theirsINI || a.merge != nil || strings.Contains(a.log.Text(), "OptiScaler.ini") {
		t.Errorf("a file the user did not change should stay as it is: %q, log %q", got, a.log.Text())
	}
}

// backups hands over the user's OptiScaler.ini when a full install will replace it, and only then.
func TestBackupsHandOverOptiScalerINI(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "OptiScaler.ini"), []byte(mineINI), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &detect.Result{ExeDir: dir, Installed: detect.KindFull, OptiINI: true}
	say := func(string, ...any) {}
	opti, err := backups(r, []step{{action: pkg.Install, kind: detect.KindFull}}, say)
	if err != nil || string(opti) != mineINI {
		t.Errorf("a full install should hand over the user's file: %q, %v", opti, err)
	}
	if opti, _ := backups(r, []step{{action: pkg.Uninstall, kind: detect.KindFull}}, say); opti != nil {
		t.Error("an uninstall installs no new file to merge into")
	}
}

// TestRealKeepSettings upgrades a scratch game from the older package to the newest with the real packages
// (release tier: ANR_PKG_DIR and ANR_OLD_PKG), with a setting the user changed in between, and checks that the
// upgrade keeps it.
func TestRealKeepSettings(t *testing.T) {
	dir, oldZip := os.Getenv("ANR_PKG_DIR"), os.Getenv("ANR_OLD_PKG")
	if dir == "" || oldZip == "" || runtime.GOOS != "windows" {
		t.Skip("ANR_PKG_DIR or ANR_OLD_PKG not set")
	}
	full, _ := filepath.Glob(filepath.Join(dir, "anr-fg-*-full.zip"))
	if len(full) != 1 {
		t.Fatalf("packages in %s: %v", dir, full)
	}
	a := testApp(t)
	m := onMainThread(a)
	oldInfo, err := pkg.ReadInfo(oldZip)
	if err != nil {
		t.Fatal(err)
	}
	// The older release, whose package the download cache holds: the base of the merge.
	name := filepath.Base(oldZip)
	data, err := os.ReadFile(oldZip)
	if err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(pkg.DownloadsDir(), 0o755)
	if err := os.WriteFile(filepath.Join(pkg.DownloadsDir(), name), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if rel := a.releaseOf(oldInfo.Version); rel != nil {
		rel.Assets = map[string]pkg.Asset{name: {Name: name}}
	} else {
		a.releases = append(a.releases, &pkg.Release{Tag: "FG-" + oldInfo.Version, Version: oldInfo.Version, Assets: map[string]pkg.Asset{name: {Name: name}}})
	}
	a.providers[oldInfo.ProviderSHA256] = oldInfo.Version

	folder := t.TempDir()
	os.WriteFile(filepath.Join(folder, "game.exe"), []byte("MZ dxgi.dll"), 0o644)
	os.WriteFile(filepath.Join(folder, "nvngx_dlssg.dll"), []byte("x"), 0o644)
	oldDir, err := pkg.Extract(oldZip)
	if err != nil {
		t.Fatal(err)
	}
	if err := pkg.Run(context.Background(), oldDir, pkg.Install, folder, func(string) {}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(folder, "OptiScaler.ini")
	ini, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(ini), "\nShortcutKey=auto") {
		t.Fatalf("the older package's OptiScaler.ini has no ShortcutKey=auto: %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(ini), "\nShortcutKey=auto", "\nShortcutKey=0x2D", 1)), 0o644); err != nil {
		t.Fatal(err)
	}

	g := &game{Name: "Scratch", Dir: folder, Scan: detect.Scan(folder, "")}
	a.games = []*game{g}
	a.run(g, "upgrade", []step{{pkg.Install, detect.KindFull, nil, full[0]}}, false)
	m.runFor(t, time.Minute, func() bool {
		if a.merge != nil {
			t.Fatalf("the upgrade asks about %+v:\n%s", a.merge.result, a.log.Text())
		}
		return !a.busy
	})
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "\nShortcutKey=0x2D") || !strings.Contains(string(got), "\nFGNvngxReplacement=Apex Neural Rendering FG") {
		t.Errorf("the upgrade should keep ShortcutKey and load ANR FG:\n%s", a.log.Text())
	}
	if !strings.Contains(a.log.Text(), i18n.T("merge.log.kept", "ShortcutKey")) || a.merge != nil {
		t.Errorf("the log should say what was kept, without asking:\n%s", a.log.Text())
	}
}

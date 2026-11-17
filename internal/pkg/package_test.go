package pkg

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"anr-installer/internal/detect"
)

// TestRealPackages installs, upgrades and removes the real packages in a fake game folder. Set ANR_PKG_DIR to
// a folder holding anr-fg-*-full.zip and anr-fg-*-addon.zip to run it.
func TestRealPackages(t *testing.T) {
	dir := os.Getenv("ANR_PKG_DIR")
	if dir == "" || runtime.GOOS != "windows" {
		t.Skip("ANR_PKG_DIR not set")
	}
	t.Setenv("LOCALAPPDATA", t.TempDir()) // keep the cache out of the user's
	full, _ := filepath.Glob(filepath.Join(dir, "anr-fg-*-full.zip"))
	addon, _ := filepath.Glob(filepath.Join(dir, "anr-fg-*-addon.zip"))
	if len(full) != 1 || len(addon) != 1 {
		t.Fatalf("packages in %s: %v %v", dir, full, addon)
	}
	info, err := ReadInfo(full[0])
	if err != nil || info.Kind != detect.KindFull {
		t.Fatalf("%+v %v", info, err)
	}
	pdir, err := Extract(full[0])
	if err != nil {
		t.Fatal(err)
	}

	game := t.TempDir()
	os.WriteFile(filepath.Join(game, "game.exe"), []byte("MZ dxgi.dll"), 0o644)
	os.WriteFile(filepath.Join(game, "nvngx_dlssg.dll"), []byte("x"), 0o644)
	var out []string
	log := func(s string) { out = append(out, s) }
	run := func(a Action) {
		t.Helper()
		if err := Run(context.Background(), pdir, a, game, log); err != nil {
			t.Fatalf("%s: %v\n%s", a, err, strings.Join(out, "\n"))
		}
	}
	run(Install)
	r := detect.Scan(game, "")
	if r.Installed != detect.KindFull || r.ANRHash != info.ProviderSHA256 {
		t.Fatalf("after install: %+v\n%s", r, strings.Join(out, "\n"))
	}
	run(Install) // the upgrade path: run again over our own install
	run(Uninstall)
	if r := detect.Scan(game, ""); r.Installed != detect.KindNone || r.OptiINI {
		t.Fatalf("after uninstall: %+v\n%s", r, strings.Join(out, "\n"))
	}
	t.Logf("installer output:\n%s", strings.Join(out, "\n"))
}

// TestRealDowngrade goes from a newer full package back to an older one the way the installer does: uninstall with
// the newer package, then install the older. Set ANR_PKG_DIR (the newer packages) and ANR_OLD_PKG (an older
// anr-fg-*-full.zip) to run it.
func TestRealDowngrade(t *testing.T) {
	dir, old := os.Getenv("ANR_PKG_DIR"), os.Getenv("ANR_OLD_PKG")
	if dir == "" || old == "" || runtime.GOOS != "windows" {
		t.Skip("ANR_PKG_DIR or ANR_OLD_PKG not set")
	}
	t.Setenv("LOCALAPPDATA", t.TempDir())
	full, _ := filepath.Glob(filepath.Join(dir, "anr-fg-*-full.zip"))
	if len(full) != 1 {
		t.Fatalf("packages in %s: %v", dir, full)
	}
	newDir, err := Extract(full[0])
	if err != nil {
		t.Fatal(err)
	}
	oldDir, err := Extract(old)
	if err != nil {
		t.Fatal(err)
	}
	oldInfo, err := ReadInfo(old)
	if err != nil {
		t.Fatal(err)
	}
	game := t.TempDir()
	os.WriteFile(filepath.Join(game, "game.exe"), []byte("MZ dxgi.dll"), 0o644)
	os.WriteFile(filepath.Join(game, "dxgi.dll"), []byte("the game's own"), 0o644)
	var out []string
	run := func(pdir string, a Action) {
		t.Helper()
		if err := Run(context.Background(), pdir, a, game, func(s string) { out = append(out, s) }); err != nil {
			t.Fatalf("%s: %v\n%s", a, err, strings.Join(out, "\n"))
		}
	}
	run(newDir, Install)
	run(newDir, Uninstall)
	if data, _ := os.ReadFile(filepath.Join(game, "dxgi.dll")); string(data) != "the game's own" {
		t.Fatal("uninstall did not restore the game's dxgi.dll")
	}
	run(oldDir, Install)
	r := detect.Scan(game, "")
	if r.Installed != detect.KindFull || r.ANRHash != oldInfo.ProviderSHA256 {
		t.Fatalf("after the downgrade: %+v (want %s)", r, oldInfo.Version)
	}
	// Nothing of the newer version is left: its programs folder holds the older package's programs alone.
	progs, _ := os.ReadDir(filepath.Join(game, "OptiScaler", "custom", detect.ANRDir, "dlssg", "programs"))
	want := map[string]bool{}
	zr, _ := zip.OpenReader(old)
	for _, f := range zr.File {
		if i := strings.Index(f.Name, "dlssg/programs/"); i >= 0 {
			if name, _, ok := strings.Cut(f.Name[i+len("dlssg/programs/"):], "/"); ok || name != "" {
				want[name] = true
			}
		}
	}
	zr.Close()
	for _, p := range progs {
		if !want[p.Name()] {
			t.Errorf("left over from the newer version: programs/%s", p.Name())
		}
	}
	run(oldDir, Uninstall)
	t.Logf("%s, then %s:\n%s", filepath.Base(full[0]), oldInfo.Version, strings.Join(out, "\n"))
}

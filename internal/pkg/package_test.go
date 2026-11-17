package pkg

import (
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

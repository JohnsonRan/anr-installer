package detect

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func write(t *testing.T, p, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestUnrealGame(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "Game.exe"), "x")
	bin := filepath.Join(root, "Game", "Binaries", "Win64")
	write(t, filepath.Join(bin, "Game-Win64-Shipping.exe"), "xxxxxxxx")
	// Streamline 8 folders down, as Phantom Blade Zero ships it.
	write(t, filepath.Join(root, "Engine", "Plugins", "Runtime", "Nvidia", "StreamlineCore", "Binaries", "ThirdParty", "Win64", "sl.dlss_g.dll"), "x")
	write(t, filepath.Join(root, "Engine", "Binaries", "CrashReportClient.exe"), "xxxxxxxxxxxxxxxx")
	r := Scan(root, "")
	if r.ExeDir != bin || !r.Supported() || !r.Streamline {
		t.Fatalf("%+v", r)
	}
	if r.Recommended() != KindFull {
		t.Errorf("recommended %q", r.Recommended())
	}
}

func TestFullInstalled(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "game.exe"), "x")
	write(t, filepath.Join(root, "nvngx_dlssg.dll"), "x")
	write(t, filepath.Join(root, "OptiScaler.ini"), "[FrameGen]\nFGNvngxReplacement=nr-record\n")
	write(t, filepath.Join(root, "OptiScaler", "nr_fg_install.json"), `{"proxy":"winmm.dll"}`)
	write(t, filepath.Join(root, "OptiScaler", "custom", ANRDir, "dlssg", "nvngx_dlssg.dll"), "anr")
	r := Scan(root, "")
	if r.Installed != KindFull || r.Proxy != "winmm.dll" || r.UserOpti || !r.NRRecord || r.ANRHash == "" {
		t.Fatalf("%+v", r)
	}
	if ok, _ := r.CanInstall(KindAddon); ok {
		t.Error("addon over a full install")
	}
}

func TestUserOptiScaler(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "game.exe"), "x")
	write(t, filepath.Join(root, "OptiScaler.ini"), "[Plugins]\nOptiDllPath=auto\n[FrameGen]\nFGInput=auto\n")
	write(t, filepath.Join(root, "OptiScaler", "OptiScaler.dll"), "x")
	r := Scan(root, "")
	if r.Installed != KindNone || !r.UserOpti || r.Recommended() != KindAddon {
		t.Fatalf("%+v", r)
	}
	if ok, _ := r.CanInstall(KindFull); ok {
		t.Error("full over the user's OptiScaler")
	}
	write(t, filepath.Join(root, "OptiScaler", "custom", ANRDir, "nr_fg_install.json"), "{}")
	if r := Scan(root, ""); r.Installed != KindAddon {
		t.Fatalf("%+v", r)
	}
}

func TestOldLayout(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "game.exe"), "x")
	write(t, filepath.Join(root, "OptiScaler.ini"), "")
	write(t, filepath.Join(root, "OptiScaler", "nr_fg_install.json"), `{}`)
	write(t, filepath.Join(root, "OptiScaler", "custom", OldANRDir, "dlssg", "nvngx_dlssg.dll"), "b01")
	if r := Scan(root, ""); r.Installed != KindFull || r.ANRHash == "" {
		t.Fatalf("full in custom/nrfg: %+v", r)
	}

	root = t.TempDir()
	write(t, filepath.Join(root, "game.exe"), "x")
	write(t, filepath.Join(root, "OptiScaler.ini"), "")
	write(t, filepath.Join(root, "OptiScaler", "OptiScaler.dll"), "x")
	write(t, filepath.Join(root, "OptiScaler", "custom", OldANRDir, "nr_fg_install.json"), "{}")
	write(t, filepath.Join(root, "OptiScaler", "custom", OldANRDir, "dlssg", "nvngx_dlssg.dll"), "b02")
	if r := Scan(root, ""); r.Installed != KindAddon || r.ANRHash == "" {
		t.Fatalf("add-on in custom/nrfg: %+v", r)
	}
}

func TestAntiCheat(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "bf6.exe"), "x")
	write(t, filepath.Join(root, "EAAntiCheat", "EAAntiCheat.Installer.exe"), "x")
	write(t, filepath.Join(root, "EAAntiCheat.GameServiceLauncher.exe"), "x")
	write(t, filepath.Join(root, "SP", "EasyAntiCheat", "Settings.json"), "x")
	r := Scan(root, "")
	if len(r.AntiCheat) != 2 || r.AntiCheat[0] == r.AntiCheat[1] {
		t.Fatalf("anti-cheat %q", r.AntiCheat)
	}
	for _, want := range []string{"EA Javelin Anticheat", "Easy Anti-Cheat"} {
		if !slices.Contains(r.AntiCheat, want) {
			t.Errorf("no %s in %q", want, r.AntiCheat)
		}
	}
	if r := Scan(t.TempDir(), ""); len(r.AntiCheat) != 0 {
		t.Errorf("an empty folder has %q", r.AntiCheat)
	}
	// A folder that only shares a prefix is not one.
	if antiCheatOf("pbr", true) != "" || antiCheatOf("eaanticheat.cfg", false) != "EA Javelin Anticheat" {
		t.Error("antiCheatOf")
	}
}

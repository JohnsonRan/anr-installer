package main

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"anr-installer/internal/detect"
	"anr-installer/internal/pkg"
)

func testApp(t *testing.T) *app {
	t.Helper()
	t.Setenv("LOCALAPPDATA", t.TempDir()) // an empty version cache
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	lang = "en"
	a := newApp()
	day := time.Date(2026, 11, 17, 12, 0, 0, 0, time.Local)
	var rels []*pkg.Release
	for _, v := range []string{"V.B19", "V.B18", "V.B04", "V.B03"} {
		rels = append(rels, &pkg.Release{Tag: "FG-" + v, Version: v, Published: day, URL: "https://example.invalid/" + v})
	}
	a.setReleases(rels)
	a.providers["aa"] = "V.B18"
	a.games = []*game{
		{Name: "Alpha", Dir: `C:\g\alpha`, Scan: &detect.Result{ExeDir: `C:\g\alpha`, DLSSG: []string{`C:\g\alpha\nvngx_dlssg.dll`}}},
		{Name: "Beta", Dir: `C:\g\beta`, Scan: &detect.Result{ExeDir: `C:\g\beta`}},
		{Name: "Gamma", Dir: `C:\g\gamma`, Scan: &detect.Result{ExeDir: `C:\g\gamma`, DLSSG: []string{"x"}, OptiINI: true, UserOpti: true}},
		{Name: "Delta", Dir: `C:\g\delta`, Scan: &detect.Result{ExeDir: `C:\g\delta`, DLSSG: []string{"x"}, OptiINI: true, Installed: detect.KindFull, Proxy: "dxgi.dll", NRRecord: true, ANRHash: "aa"}},
	}
	return a
}

func TestViewFiltersAndSelects(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, 1040, 760)
	if tt.HasText("Beta") {
		t.Error("Beta has no DLSS-G and should be hidden")
	}
	if !tt.HasText("Alpha") || !tt.HasText("Gamma") || !tt.HasText("Delta") {
		t.Fatalf("texts %q", tt.Texts())
	}
	if !tt.HasText("ANR full B18") {
		t.Errorf("the list should show the installed version: %q", tt.Texts())
	}
	if err := tt.Click("Only games with DLSS frame generation"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Beta") {
		t.Error("Beta should show with the filter off")
	}
	if err := tt.Click("Gamma"); err != nil {
		t.Fatal(err)
	}
	if g := a.selected(); g == nil || g.Name != "Gamma" {
		t.Fatalf("selected %+v", g)
	}
	if !tt.HasText("Auto (addon)") || !tt.HasText("Install V.B19") {
		t.Errorf("texts %q", tt.Texts())
	}
	a.kind = kindFull
	tt.Frame()
	if !tt.HasText(T("why.user-optiscaler")) {
		t.Errorf("no reason shown: %q", tt.Texts())
	}
}

func TestVersionsOffered(t *testing.T) {
	a := testApp(t)
	if len(a.versionLabels) != 3 || a.versionLabels[0] != "V.B19 · 2026-11-17 (latest)" || a.versionLabels[2] != "V.B04 · 2026-11-17" {
		t.Fatalf("labels %q: V.B03 and older install into custom/nrfg and are not offered", a.versionLabels)
	}
	if a.target().Version != "V.B19" {
		t.Errorf("default target %s", a.target().Version)
	}
}

func TestPlans(t *testing.T) {
	a := testApp(t)
	delta := a.games[3].Scan // V.B18, full
	pick := func(v string) {
		for i, r := range a.versionRel {
			if r.Version == v {
				a.versionSel = a.versionLabels[i]
			}
		}
	}
	for _, c := range []struct {
		target string
		hash   string
		want   planKind
		label  string
	}{
		{"V.B19", "aa", planUpgrade, "Upgrade to V.B19"},
		{"V.B18", "aa", planReinstall, "Reinstall"},
		{"V.B04", "aa", planDowngrade, "Downgrade to V.B04"},
		{"V.B19", "unknown", planUpdate, "Install V.B19"},
		{"V.B04", "unknown", planSwitch, "Switch to V.B04"},
	} {
		pick(c.target)
		delta.ANRHash = c.hash
		if got := a.plan(delta); got != c.want || a.planLabel(got) != c.label {
			t.Errorf("%s over %s: plan %d %q, want %d %q", c.target, c.hash, got, a.planLabel(got), c.want, c.label)
		}
	}
	if a.plan(a.games[0].Scan) != planInstall {
		t.Error("nothing installed")
	}
	if !planDowngrade.uninstallsFirst() || !planSwitch.uninstallsFirst() || planUpgrade.uninstallsFirst() || planUpdate.uninstallsFirst() {
		t.Error("uninstallsFirst")
	}
	// A downgrade uninstalls with the newest release's package, then installs the one chosen.
	pick("V.B04")
	delta.ANRHash = "aa"
	steps := []step{a.uninstallStep(delta.Installed), a.installStep(detect.KindFull)}
	if steps[0].rel.Version != "V.B19" || steps[0].action != pkg.Uninstall || steps[1].rel.Version != "V.B04" {
		t.Errorf("steps %+v", steps)
	}
}

func TestViewDowngrade(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, 1040, 760)
	if err := tt.Click("Delta"); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"Upgrade to V.B19", "Uninstall", "Installed version: V.B18", T("note.ini.reset"), T("note.nrrecord")} {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
	a.versionSel = a.versionLabels[2]
	tt.Frame()
	if !tt.HasText("Downgrade to V.B04") || !tt.HasText(T("note.downgrade", "V.B19", "V.B04")) {
		t.Errorf("downgrade not offered: %q", tt.Texts())
	}
}

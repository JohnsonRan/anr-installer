package main

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"anr-installer/internal/detect"
	"anr-installer/internal/pkg"
)

func testApp() *app {
	lang = "en"
	a := newApp()
	a.rel = &pkg.Release{Tag: "FG-V.B19", Version: "V.B19"}
	a.games = []*game{
		{Name: "Alpha", Dir: `C:\g\alpha`, Scan: &detect.Result{ExeDir: `C:\g\alpha`, DLSSG: []string{`C:\g\alpha\nvngx_dlssg.dll`}}},
		{Name: "Beta", Dir: `C:\g\beta`, Scan: &detect.Result{ExeDir: `C:\g\beta`}},
		{Name: "Gamma", Dir: `C:\g\gamma`, Scan: &detect.Result{ExeDir: `C:\g\gamma`, DLSSG: []string{"x"}, OptiINI: true, UserOpti: true}},
		{Name: "Delta", Dir: `C:\g\delta`, Scan: &detect.Result{ExeDir: `C:\g\delta`, DLSSG: []string{"x"}, OptiINI: true, Installed: detect.KindFull, Proxy: "dxgi.dll", NRRecord: true, ANRHash: "aa"}},
	}
	return a
}

func TestViewFiltersAndSelects(t *testing.T) {
	a := testApp()
	tt := ui.NewTester(a.view, 1040, 680)
	if tt.HasText("Beta") {
		t.Error("Beta has no DLSS-G and should be hidden")
	}
	if !tt.HasText("Alpha") || !tt.HasText("Gamma") || !tt.HasText("Delta") {
		t.Fatalf("texts %q", tt.Texts())
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
	if !tt.HasText("Auto (addon)") || !tt.HasText("Install") {
		t.Errorf("texts %q", tt.Texts())
	}
	// The full package refuses a folder with the user's OptiScaler.
	a.kind = kindFull
	tt.Frame()
	if !tt.HasText(T("why.user-optiscaler")) {
		t.Errorf("no reason shown: %q", tt.Texts())
	}
}

func TestViewInstalledFull(t *testing.T) {
	a := testApp()
	tt := ui.NewTester(a.view, 1040, 680)
	if err := tt.Click("Delta"); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"Upgrade to V.B19", "Uninstall", T("note.ini.reset"), T("note.nrrecord"), T("nrrecord.on")} {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
	a.providers["aa"] = "V.B19"
	tt.Frame()
	if !tt.HasText("Reinstall") {
		t.Errorf("same version should offer Reinstall: %q", tt.Texts())
	}
}

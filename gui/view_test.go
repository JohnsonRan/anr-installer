package gui

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
	"anr-installer/internal/detect"
	"anr-installer/internal/pkg"
	"anr-installer/internal/store"
)

func testApp(t *testing.T) *app {
	t.Helper()
	// An empty version cache and list of games, kept apart from the user's.
	t.Setenv("LOCALAPPDATA", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	i18n.SetLang("en")
	a := newApp()
	day := time.Date(2026, 11, 17, 12, 0, 0, 0, time.Local)
	var rels []*pkg.Release
	for _, v := range []string{"V.B19", "V.B18", "V.B04", "V.B03"} {
		rels = append(rels, &pkg.Release{Tag: "FG-" + v, Version: v, Published: day, URL: "https://example.invalid/" + v})
	}
	a.setReleases(rels)
	a.providers["aa"] = "V.B18"
	a.saved.SeenRelease = "FG-V.B19" // the notes of B19 have been seen: no dialog opens by itself
	a.games = []*game{
		{Name: "Alpha", Dir: `C:\g\alpha`, Scan: &detect.Result{Root: `C:\g\alpha`, ExeDir: `C:\g\alpha`, DLSSG: []string{`C:\g\alpha\nvngx_dlssg.dll`}}},
		{Name: "Beta", Dir: `C:\g\beta`, Scan: &detect.Result{Root: `C:\g\beta`, ExeDir: `C:\g\beta`}},
		{Name: "Gamma", Dir: `C:\g\gamma`, Scan: &detect.Result{Root: `C:\g\gamma`, ExeDir: `C:\g\gamma`, DLSSG: []string{`C:\g\gamma\x\nvngx_dlssg.dll`}, OptiINI: true, UserOpti: true}},
		{Name: "Delta", Dir: `C:\g\delta`, Scan: &detect.Result{Root: `C:\g\delta`, ExeDir: `C:\g\delta`, DLSSG: []string{`C:\g\delta\nvngx_dlssg.dll`}, OptiINI: true, Installed: detect.KindFull, Proxy: "dxgi.dll", NRRecord: true, ANRHash: "aa"}},
	}
	return a
}

func TestViewFiltersAndSelects(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, 1040, 760)
	if tt.HasText("Beta") {
		t.Error("Beta has no DLSS-G and should be hidden under Supported")
	}
	if !tt.HasText("Alpha") || !tt.HasText("Gamma") || !tt.HasText("ANR B18 → B19 · upgrade available") {
		t.Fatalf("texts %q", tt.Texts())
	}
	if g := a.selected(); g == nil || g.Name != "Alpha" {
		t.Fatalf("the first game should be chosen at start: %+v", g)
	}
	if err := tt.Click("All"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Beta") {
		t.Error("Beta should show under All")
	}
	if err := tt.Click("Installed"); err != nil {
		t.Fatal(err)
	}
	if tt.HasText("Alpha") || !tt.HasText("Delta") {
		t.Errorf("Installed shows %q", tt.Texts())
	}
	if err := tt.Click("Supported"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Gamma"); err != nil {
		t.Fatal(err)
	}
	if g := a.selected(); g == nil || g.Name != "Gamma" {
		t.Fatalf("selected %+v", g)
	}
	if !tt.HasText("Install V.B19") || !tt.HasText("Installed by you") || !tt.HasText("x") {
		t.Errorf("texts %q", tt.Texts())
	}
	if err := tt.Click("Advanced options"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Automatic (addon)") {
		t.Errorf("Advanced shows %q", tt.Texts())
	}
	a.kind = kindFull
	tt.Frame()
	if !tt.HasText(i18n.T("why.user-optiscaler")) {
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

// A language change names the versions again in it, and the version chosen, the newest, whose name changes,
// stays chosen.
func TestVersionLabelsFollowLanguage(t *testing.T) {
	a := testApp(t)
	t.Cleanup(func() { i18n.SetLang("en") })
	i18n.SetLang("ja")
	a.relabelVersions()
	if a.versionLabels[0] != "V.B19 · 2026-11-17 (最新)" || a.target() == nil || a.target().Version != "V.B19" {
		t.Errorf("labels %q, chosen %q", a.versionLabels, a.versionSel)
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
	for _, s := range []string{"Upgrade to V.B19", "Uninstall", "V.B18", "Full package · OptiScaler loads as dxgi.dll", i18n.T("note.ini.reset"), i18n.T("note.nrrecord")} {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
	a.versionSel = a.versionLabels[2]
	tt.Frame()
	if !tt.HasText("Downgrade to V.B04") || !tt.HasText(i18n.T("note.downgrade", "V.B19", "V.B04")) {
		t.Errorf("downgrade not offered: %q", tt.Texts())
	}
}

func TestManualGames(t *testing.T) {
	a := testApp(t)
	epic := &game{Name: "Epic Game", Dir: `D:\Epic\Game`, Manual: true,
		Scan: &detect.Result{Root: `D:\Epic\Game`, ExeDir: `D:\Epic\Game\bin`, Exe: `D:\Epic\Game\bin\game.exe`, DLSSG: []string{"x"}}}
	a.games = append(a.games, epic)
	a.games[0].ExeDir = `C:\g\alpha\bin` // a Steam game's .exe folder, chosen by hand
	a.persist()

	saved := store.Load()
	if len(saved.Games) != 1 || saved.Games[0].Dir != epic.Dir || saved.ExeDirs[store.Key(`C:\g\alpha`)] != `C:\g\alpha\bin` {
		t.Fatalf("saved %+v", saved)
	}
	// A new launch brings the game back.
	again := newApp()
	if len(again.games) != 1 || again.games[0].Name != "Epic Game" || !again.games[0].Manual {
		t.Fatalf("games at launch %+v", again.games)
	}

	// A game added by hand takes its .exe's name until the user renames it.
	a.scanned(epic, epic.Scan, "Epic Game: Remastered")
	if epic.Name != "Epic Game: Remastered" {
		t.Errorf("name %q", epic.Name)
	}
	epic.Name = "Mine"
	a.renamed(epic)
	a.scanned(epic, epic.Scan, "Epic Game: Remastered")
	if epic.Name != "Mine" || !store.Load().Games[0].Named {
		t.Errorf("a renamed game took its .exe's name: %q", epic.Name)
	}

	tt := ui.NewTester(a.view, 1040, 760)
	if err := tt.Click("Mine"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Added manually") || !tt.HasText("Supports DLSS frame generation · added manually") {
		t.Errorf("texts %q", tt.Texts())
	}
	if err := tt.Click("Remove from list"); err != nil {
		t.Fatal(err)
	}
	for _, g := range a.games {
		if g == epic {
			t.Error("not removed from the list")
		}
	}
	if len(store.Load().Games) != 0 {
		t.Error("still saved")
	}
	if g := a.selected(); g == nil || g.Name != "Alpha" {
		t.Errorf("after removing, the first game should be chosen: %+v", g)
	}
	// Steam games cannot be removed: they come back with the next scan.
	a.removeGame(a.games[0])
	if a.games[0].Name != "Alpha" {
		t.Error("a Steam game was removed")
	}
}

func TestAntiCheatWarning(t *testing.T) {
	a := testApp(t)
	a.games[0].Scan.AntiCheat = []string{"EA Javelin Anticheat"}
	tt := ui.NewTester(a.view, 1040, 760)
	if !tt.HasText(i18n.T("anticheat.found", "EA Javelin Anticheat")) {
		t.Errorf("no anti-cheat warning: %q", tt.Texts())
	}
	status := "Supports DLSS frame generation · anti-cheat"
	if !tt.HasText(status) {
		t.Errorf("the list does not mark the game: %q", tt.Texts())
	}
	// Too long for the list's width, the status wraps under the name rather than cutting "anti-cheat" off.
	name, _ := tt.Find("Alpha")
	if box, ok := tt.Find(status); !ok || box.H <= name.H {
		t.Errorf("the status is cut off: %v high, its name %v", box.H, name.H)
	}
	if err := tt.Click("Gamma"); err != nil {
		t.Fatal(err)
	}
	if tt.HasText(i18n.T("anticheat.found", "EA Javelin Anticheat")) {
		t.Error("a game without anti-cheat shows the warning")
	}
}

func TestAboutDialog(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, 1040, 760)
	if tt.HasText(i18n.T("about.updates")) {
		t.Fatal("the About dialog shows before it is opened")
	}
	a.openAbout()
	tt.Frame()
	for _, s := range []string{i18n.T("about.updates"), i18n.T("about.cantUpdate"), i18n.T("about.checkNow"), i18n.T("about.changelog"), i18n.T("about.license"), i18n.T("about.notice")} {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
	if tt.HasText(i18n.T("settings.updates")) {
		t.Error("the automatic checks belong to the settings page")
	}
	if err := tt.Click(i18n.T("about.close")); err != nil {
		t.Fatal(err)
	}
	if a.aboutOpen || tt.HasText(i18n.T("about.updates")) {
		t.Error("Close did not close the dialog")
	}
}

func TestUpdateReminders(t *testing.T) {
	a := testApp(t) // Delta has V.B18; the newest release is V.B19
	tt := ui.NewTester(a.view, 1040, 760)
	if !tt.HasText(i18n.T("updates.count", 1, "V.B19")) {
		t.Fatalf("no updates bar: %q", tt.Texts())
	}
	if !tt.HasText("ANR B18 → B19 · upgrade available") {
		t.Errorf("the row does not mark the upgrade: %q", tt.Texts())
	}
	if err := tt.Click(i18n.T("updates.count", 1, "V.B19")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if tt.HasText("Alpha") || !tt.HasText("Delta") {
		t.Errorf("the bar should show only Delta: %q", tt.Texts())
	}
	if !tt.HasText(i18n.T("update.available", "V.B19", "V.B18")) {
		t.Errorf("Delta's page does not say the new version is out: %q", tt.Texts())
	}
	// An unknown installed version is not taken for an old one.
	a.games[3].Scan.ANRHash = "unknown"
	a.onlyOutdated = false
	tt.Frame()
	if tt.HasText(i18n.T("updates.count", 1, "V.B19")) {
		t.Error("an unknown version counted as outdated")
	}
}

func TestReleaseNotesDialog(t *testing.T) {
	a := testApp(t)
	a.releases[0].Notes = "## Changes\n\n- Faster frame generation"
	tt := ui.NewTester(a.view, 1040, 760)
	if err := tt.Click("Delta"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click(i18n.T("notes.whatsnew")); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText(i18n.T("notes.title", "V.B19")) || !tt.HasText("Faster frame generation") {
		t.Errorf("notes not shown: %q", tt.Texts())
	}
	if err := tt.Click(i18n.T("about.close")); err != nil {
		t.Fatal(err)
	}
	if a.notesRel != nil {
		t.Error("Close did not close the notes")
	}
}

func TestNewReleaseOpensNotes(t *testing.T) {
	a := testApp(t)
	a.saved.SeenRelease = "FG-V.B18"
	a.games[2].Scan.Installed, a.games[2].Scan.ANRHash = detect.KindAddon, "aa" // Gamma: B18 too
	a.games[3].Scan.AntiCheat = []string{"EasyAntiCheat"}                       // Delta
	tt := ui.NewTester(a.view, 1040, 760)
	if !tt.HasText(i18n.T("notes.title", "V.B19")) {
		t.Fatalf("the notes of a new release did not open: %q", tt.Texts())
	}
	if !tt.HasText(i18n.T("notes.updateAll", 1)) || !tt.HasText(i18n.T("notes.anticheat", "Delta")) {
		t.Errorf("Update all should take Gamma alone and name Delta: %q", tt.Texts())
	}
	if err := tt.Click(i18n.T("about.close")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if tt.HasText(i18n.T("notes.title", "V.B19")) || a.saved.SeenRelease != "FG-V.B19" {
		t.Error("the notes should open once a release")
	}
	if games, _ := a.upgradable(); len(games) != 1 || games[0].Name != "Gamma" {
		t.Errorf("upgradable %v", games)
	}
}

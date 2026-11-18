package gui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
	"anr-installer/internal/paths"
	"anr-installer/internal/pkg"
	"anr-installer/internal/store"
)

func TestSettingsPage(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, 1040, 760)
	if err := tt.Click(i18n.T("settings.open")); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{i18n.T("settings.language"), i18n.T("settings.language.system"), i18n.T("settings.close"), i18n.T("settings.close.quit"),
		i18n.T("settings.updates"), i18n.T("settings.updates.detail"), i18n.T("settings.login"), i18n.T("settings.login.detail")} {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
	if tt.HasText(i18n.T("status")) {
		t.Error("the game's page shows under the settings")
	}
	if err := tt.Click(i18n.T("settings.close.quit")); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click(i18n.T("settings.close.tray")); err != nil {
		t.Fatal(err)
	}
	if !a.saved.CloseToTray || !store.Load().CloseToTray {
		t.Error("Minimize to the tray was not kept")
	}
	if !tt.HasText(i18n.T("settings.close.tray")) || tt.HasText(i18n.T("settings.close.quit")) {
		t.Errorf("the select does not show the choice: %q", tt.Texts())
	}
	if err := tt.Click(i18n.T("settings.language.system")); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click(i18n.T("settings.language.ja")); err != nil {
		t.Fatal(err)
	}
	if a.settings.locale != "ja" {
		t.Errorf("choosing Japanese chose %q", a.settings.locale)
	}
	if err := tt.Click(i18n.T("back")); err != nil {
		t.Fatal(err)
	}
	if a.settings != nil || !tt.HasText(i18n.T("status")) {
		t.Errorf("Back should return to the game's page: %q", tt.Texts())
	}
}

func TestSettingsStorage(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, 1040, 760)
	if err := tt.Click(i18n.T("settings.open")); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{i18n.T("settings.folder"), paths.Config(), i18n.T("open.folder"), i18n.T("settings.cache"), i18n.T("settings.cache.measuring"), i18n.T("settings.cache.delete")} {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
	a.settings.cacheSize = 2_345_000_000
	tt.Frame()
	if !tt.HasText(i18n.T("settings.cache.detail", "2.3 GB")) {
		t.Errorf("the size of the downloads is not shown: %q", tt.Texts())
	}
	for n, want := range map[int64]string{0: "0.0 MB", 512_300_000: "512.3 MB", 1_000_000_000: "1.0 GB"} {
		if got := sizeText(n); got != want {
			t.Errorf("sizeText(%d) = %q, want %q", n, got, want)
		}
	}
}

// Delete downloads removes downloads\ and packages\, keeps versions.json and the comparisons, counts the
// downloads again, and waits for a task that runs.
func TestDeleteDownloads(t *testing.T) {
	a := testApp(t)
	m := onMainThread(a)
	a.toggleSettings()
	cache := pkg.CacheDir()
	files := map[string]bool{ // kept
		filepath.Join(cache, "downloads", "anr-fg-V.B19-full.zip"):          false,
		filepath.Join(cache, "packages", "0123456789abcdef", "install.ps1"): false,
		filepath.Join(cache, "versions.json"):                               true,
		filepath.Join(cache, "compare", "20261117-210405-delta", "anr.mp4"): true,
	}
	for p := range files {
		writeFile(t, p, "x")
	}
	a.busy = true
	a.deleteDownloads()
	if a.log.Text() != "" {
		t.Fatal("Delete downloads started while another task ran")
	}
	a.busy, a.settings.cacheSize = false, 2
	a.deleteDownloads()
	if !a.busy {
		t.Error("Delete downloads is not the task in progress")
	}
	m.run(t, func() bool { return !a.busy })
	for _, dir := range []string{"downloads", "packages"} {
		files[filepath.Join(cache, dir)] = false
	}
	for p, want := range files {
		if _, err := os.Stat(p); (err == nil) != want {
			t.Errorf("%s: kept %v, want %v", p, err == nil, want)
		}
	}
	if log := a.log.Text(); !strings.Contains(log, i18n.T("log.ok")) || !strings.Contains(log, "\n  "+i18n.T("settings.cache.deleted", sizeText(2))+"\n") {
		t.Errorf("the task does not tell what was freed:\n%s", log)
	}
	if a.settings.cacheSize != -1 {
		t.Error("the downloads are not counted again")
	}
}

// Delete downloads that fails, as on a package in use, ends its task failed, which opens the log.
func TestDeleteDownloadsFails(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows keeps a file in use from being deleted")
	}
	a := testApp(t)
	m := onMainThread(a)
	a.toggleSettings()
	zip := filepath.Join(pkg.DownloadsDir(), "anr-fg-V.B19-full.zip")
	writeFile(t, zip, "x")
	f, err := os.Open(zip)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	a.deleteDownloads()
	m.run(t, func() bool { return !a.busy })
	if log := a.log.Text(); !strings.Contains(log, i18n.T("log.failed")) || !a.logOpen {
		t.Errorf("the failure is not shown (log open %v):\n%s", a.logOpen, log)
	}
}

func TestSettingsOverComparison(t *testing.T) {
	a, s := comparePageApp(t)
	tt := ui.NewTester(a.view, 1040, 900)
	if err := tt.Click(i18n.T("settings.open")); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText(i18n.T("settings.language")) || tt.HasText(i18n.T("cmp.title", "ANR FG", "DLSS-G")) {
		t.Fatalf("the settings should show in place of the comparison: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	if a.settings != nil || a.cmp != s || !tt.HasText(i18n.T("cmp.title", "ANR FG", "DLSS-G")) {
		t.Fatalf("Escape should return to the comparison: %q", tt.Texts())
	}
	if err := tt.Click(i18n.T("settings.open")); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Alpha"); err != nil {
		t.Fatal(err)
	}
	if a.settings != nil {
		t.Error("choosing another game should leave the settings")
	}
}

// The settings opened before any game is chosen stay open when the first game found is chosen by itself.
func TestSettingsStayAsFirstGameIsChosen(t *testing.T) {
	a := testApp(t)
	a.toggleSettings() // before the first frame, which chooses the first game
	tt := ui.NewTester(a.view, 1040, 760)
	if a.settings == nil || !tt.HasText(i18n.T("settings.language")) {
		t.Fatalf("the settings closed as the first game was chosen: %q", tt.Texts())
	}
}

func TestLangOf(t *testing.T) {
	for tag, want := range map[string]string{
		"zh-Hans": "zh-Hans", "zh-CN": "zh-Hans", "zh-SG": "zh-Hans", "ZH": "zh-Hans",
		"zh-Hant": "zh-Hant", "zh-Hant-TW": "zh-Hant", "zh-TW": "zh-Hant", "zh-HK": "zh-Hant",
		"ja": "ja", "ja-JP": "ja", "en": "en", "en-GB": "en", "de": "en", "": "en",
	} {
		if got := i18n.LangOf(tag); got != want {
			t.Errorf("langOf(%q) = %q, want %q", tag, got, want)
		}
	}
	for tag, want := range map[string]int{"": 0, "en": 1, "zh-Hans": 2, "zh-CN": 2, "zh-Hant": 3, "zh-TW": 3, "ja": 4, "ja-JP": 4, "de": 1} {
		if got := localeIndex(tag); got != want {
			t.Errorf("localeIndex(%q) = %d, want %d", tag, got, want)
		}
	}
}

func TestQuitAsks(t *testing.T) {
	for _, c := range []struct {
		busy, forced, final, want bool
	}{
		{false, false, false, false}, // no task: quits at once
		{true, false, false, true},
		{true, true, false, false},  // the answer "Quit anyway"
		{true, false, true, false},  // the session ends, or a quit let through closes the window
		{false, true, false, false}, // the flag left over changes nothing
	} {
		if got := quitAsks(c.busy, c.forced, c.final); got != c.want {
			t.Errorf("quitAsks(%v, %v, %v) = %v", c.busy, c.forced, c.final, got)
		}
	}
}

func TestQuitDialog(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, 1040, 760)
	if tt.HasText(i18n.T("quit.title")) {
		t.Fatal("the question shows before it is asked")
	}
	a.quitAsk = true
	tt.Frame()
	for _, s := range []string{i18n.T("quit.title"), i18n.T("quit.detail"), i18n.T("quit.keep"), i18n.T("quit.anyway")} {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
	if err := tt.Click(i18n.T("quit.keep")); err != nil {
		t.Fatal(err)
	}
	if a.quitAsk || a.quitForced || tt.HasText(i18n.T("quit.title")) {
		t.Error("Keep running should close the question and let no quit through")
	}
}

// Escape answers a dialog over a comparison or the settings, and leaves the page beneath open.
func TestEscapeAnswersTheDialogFirst(t *testing.T) {
	a, s := comparePageApp(t)
	tt := ui.NewTester(a.view, 1040, 900)
	a.aboutOpen = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if a.aboutOpen || a.cmp != s {
		t.Errorf("Escape over the comparison: About open %v, comparison open %v", a.aboutOpen, a.cmp == s)
	}
	if err := tt.Click(i18n.T("settings.open")); err != nil {
		t.Fatal(err)
	}
	a.quitAsk = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if a.quitAsk || a.settings == nil {
		t.Errorf("Escape over the settings: question open %v, settings open %v", a.quitAsk, a.settings != nil)
	}
}

func TestCloseHides(t *testing.T) {
	for _, c := range []struct {
		toTray, hasTray, quitting, want bool
	}{
		{false, false, false, false},
		{false, true, false, false}, // the setting off: closing quits
		{true, false, false, false}, // no tray icon to come back by
		{true, true, false, true},
		{true, true, true, false}, // a quit closes the window
	} {
		if got := closeHides(c.toTray, c.hasTray, c.quitting); got != c.want {
			t.Errorf("closeHides(%v, %v, %v) = %v", c.toTray, c.hasTray, c.quitting, got)
		}
	}
}

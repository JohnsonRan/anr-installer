package main

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"anr-installer/internal/store"
)

func TestSettingsPage(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, 1040, 760)
	if err := tt.Click(T("settings.open")); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{T("settings.language"), T("settings.language.system"), T("settings.close"), T("settings.close.quit")} {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
	if tt.HasText(T("status")) {
		t.Error("the game's page shows under the settings")
	}
	if err := tt.Click(T("settings.close.quit")); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click(T("settings.close.tray")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.saved.CloseToTray || !store.Load().CloseToTray {
		t.Error("Minimize to the tray was not kept")
	}
	if !tt.HasText(T("settings.close.tray")) || tt.HasText(T("settings.close.quit")) {
		t.Errorf("the select does not show the choice: %q", tt.Texts())
	}
	if err := tt.Click(T("back")); err != nil {
		t.Fatal(err)
	}
	if a.settings != nil || !tt.HasText(T("status")) {
		t.Errorf("Back should return to the game's page: %q", tt.Texts())
	}
}

func TestSettingsOverComparison(t *testing.T) {
	a, s := comparePageApp(t)
	s.mode = cmpStills // the test has no videos to show
	tt := ui.NewTester(a.view, 1040, 900)
	if err := tt.Click(T("settings.open")); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText(T("settings.language")) || tt.HasText(T("cmp.title", "ANR FG", "DLSS-G")) {
		t.Fatalf("the settings should show in place of the comparison: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	if a.settings != nil || a.cmp != s || !tt.HasText(T("cmp.title", "ANR FG", "DLSS-G")) {
		t.Fatalf("Escape should return to the comparison: %q", tt.Texts())
	}
	if err := tt.Click(T("settings.open")); err != nil {
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
	if a.settings == nil || !tt.HasText(T("settings.language")) {
		t.Fatalf("the settings closed as the first game was chosen: %q", tt.Texts())
	}
}

func TestLangOf(t *testing.T) {
	for tag, want := range map[string]string{"zh-Hans": "zh", "zh-Hant-TW": "zh", "ZH": "zh", "en": "en", "en-GB": "en", "de": "en", "": "en"} {
		if got := langOf(tag); got != want {
			t.Errorf("langOf(%q) = %q, want %q", tag, got, want)
		}
	}
	for tag, want := range map[string]int{"": 0, "en": 1, "zh-Hans": 2, "zh-CN": 2, "de": 1} {
		if got := localeIndex(tag); got != want {
			t.Errorf("localeIndex(%q) = %d, want %d", tag, got, want)
		}
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

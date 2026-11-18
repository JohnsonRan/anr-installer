package main

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"anr-installer/internal/store"
)

func TestShouldNotify(t *testing.T) {
	for _, c := range []struct {
		on               bool
		newest, notified string
		shown, want      bool
	}{
		{true, "FG-V.B20", "FG-V.B19", false, true},
		{true, "FG-V.B20", "", false, true},           // the first check after an upgrade from a version without it
		{false, "FG-V.B20", "FG-V.B19", false, false}, // the setting off
		{true, "FG-V.B20", "FG-V.B20", false, false},  // notified before
		{true, "FG-V.B20", "FG-V.B19", true, false},   // the window shows it: autoNotes does
		{true, "", "", false, false},                  // no release found
	} {
		if got := shouldNotify(c.on, c.newest, c.notified, c.shown); got != c.want {
			t.Errorf("shouldNotify(%v, %q, %q, %v) = %v", c.on, c.newest, c.notified, c.shown, got)
		}
	}
}

// A release found while the window shows is recorded without a notification, once the games are scanned.
func TestNotifyReleaseRecordsOnce(t *testing.T) {
	a := testApp(t) // no window: it counts as shown
	a.games[0].Scanning = true
	a.notifyRelease()
	if a.saved.NotifiedRelease != "" {
		t.Fatal("the release was recorded before the games were scanned, which the notification counts")
	}
	a.games[0].Scanning = false
	a.notifyRelease()
	if a.saved.NotifiedRelease != "FG-V.B19" || store.Load().NotifiedRelease != "FG-V.B19" {
		t.Errorf("the release was not recorded: %q", a.saved.NotifiedRelease)
	}
}

func TestSettingsReleaseNotify(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, 1040, 760)
	if err := tt.Click(T("settings.open")); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText(T("settings.notify")) || !tt.HasText(T("settings.notify.detail")) || !a.settings.notify {
		t.Fatalf("the notification setting should show, on: %q", tt.Texts())
	}
	clickSwitch(tt, T("settings.notify"), T("settings.notify.detail"))
	if !a.saved.ReleaseNotifyOff || !store.Load().ReleaseNotifyOff {
		t.Error("turning the notification off was not kept")
	}
}

// clickSwitch clicks the switch at the end of the settings row of title and detail, where the card's padding
// leaves it at the page's end. Its label is the title's text, which Click finds first.
func clickSwitch(tt *ui.Tester, title, detail string) {
	r, _ := tt.Find(title)
	d, _ := tt.Find(detail)
	end := float32(tt.Image().Bounds().Dx()) - padPageX - padCardX
	tt.ClickAt(end-space4, (r.Y+d.Y+d.H)/2)
	tt.Frame()
}

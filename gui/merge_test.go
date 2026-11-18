package gui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
)

const (
	theirsINI = "; the new file\r\n[FrameGen]\r\nFGInput=nvngxfg\r\nFGNvngxReplacement=Apex Neural Rendering FG\r\n[Menu]\r\nScale=auto\r\nShortcutKey=0x24\r\n"
	mineINI   = "[FrameGen]\nFGInput=nvngxfg\nFGNvngxReplacement=nr-record\n[Menu]\nScale=1.5\nShortcutKey=0x2D\n"
)

// mergeGame is a game folder whose OptiScaler.ini an install has just replaced with theirsINI.
func mergeGame(t *testing.T) (a *app, r *reporter, dir string, m mainThread) {
	a = testApp(t)
	m = onMainThread(a)
	dir = t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "OptiScaler.ini"), []byte(theirsINI), 0o644); err != nil {
		t.Fatal(err)
	}
	return a, &reporter{a: a, tk: a.log.Begin("Delta · Upgrade to V.B19", time.Now())}, dir, m
}

func readINI(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "OptiScaler.ini"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// In a batch, the user's values stay without a question, but not those ANR FG needs; the log says which.
func TestKeepSettingsInABatch(t *testing.T) {
	a, r, dir, m := mergeGame(t)
	done := make(chan error, 1)
	go func() { done <- a.keepSettings(context.Background(), r, dir, []byte(mineINI), nil, "V.B19", true) }()
	var err error
	m.run(t, func() bool {
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
	want := "; the new file\r\n[FrameGen]\r\nFGInput=nvngxfg\r\nFGNvngxReplacement=Apex Neural Rendering FG\r\n[Menu]\r\nScale=1.5\r\nShortcutKey=0x2D\r\n"
	if got := readINI(t, dir); got != want {
		t.Errorf("OptiScaler.ini:\n%q\nwant\n%q", got, want)
	}
	if a.merge != nil {
		t.Error("a batch should not ask")
	}
	log := a.log.Text()
	for _, s := range []string{
		i18n.T("merge.log.kept", "Scale (new: auto), ShortcutKey (new: 0x24)"),
		i18n.T("merge.log.pinned", "FGNvngxReplacement"),
	} {
		if !strings.Contains(log, s) {
			t.Errorf("no %q in the log:\n%s", s, log)
		}
	}
}

// One install asks about the values both changed, and writes what the user chose.
func TestKeepSettingsAsks(t *testing.T) {
	a, r, dir, m := mergeGame(t)
	done := make(chan error, 1)
	go func() { done <- a.keepSettings(context.Background(), r, dir, []byte(mineINI), nil, "V.B19", false) }()
	m.run(t, func() bool { return a.merge != nil })
	if len(a.merge.result.Conflicts) != 2 {
		t.Fatalf("conflicts %+v", a.merge.result.Conflicts)
	}
	tt := ui.NewTester(a.view, 1040, 900)
	if !tt.HasText(i18n.T("merge.title")) || !tt.HasText(i18n.T("merge.mine", "0x2D")) || !tt.HasText(i18n.T("merge.pinned", "FGNvngxReplacement", "Apex Neural Rendering FG", "nr-record")) {
		t.Fatalf("the dialog should show the conflicts and the pinned value: %q", tt.Texts())
	}
	if err := tt.Click(i18n.T("merge.new", "0x24")); err != nil { // ShortcutKey takes the new value; Scale stays the user's
		t.Fatal(err)
	}
	if err := tt.Click(i18n.T("merge.apply")); err != nil {
		t.Fatal(err)
	}
	var err error
	m.run(t, func() bool {
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
	if got := readINI(t, dir); !strings.Contains(got, "Scale=1.5\r\n") || !strings.Contains(got, "ShortcutKey=0x24\r\n") {
		t.Errorf("the file should hold the choices:\n%s", got)
	}
	if tt.Frame(); a.merge != nil || tt.HasText(i18n.T("merge.title")) {
		t.Error("the dialog should close")
	}
}

// Nothing the user changed: the file stays the new one, and the log says nothing.
func TestKeepSettingsWithNothingChanged(t *testing.T) {
	a, r, dir, _ := mergeGame(t)
	if err := a.keepSettings(context.Background(), r, dir, []byte(theirsINI), nil, "V.B19", false); err != nil {
		t.Fatal(err)
	}
	if got := readINI(t, dir); got != theirsINI || a.merge != nil || strings.Contains(a.log.Text(), "OptiScaler.ini") {
		t.Errorf("a file the user did not change should stay as it is: %q, log %q", got, a.log.Text())
	}
}

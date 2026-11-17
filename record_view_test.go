package main

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"anr-installer/internal/record"
)

// recordApp is testApp with Delta, which has ANR FG's full package, set up to record: nr-record installed, its
// own DLLs known, and two recordings of it among three.
func recordApp(t *testing.T) *app {
	a := testApp(t)
	d := a.games[3]
	d.Recorder = true
	d.Scan.Exe = `C:\g\delta\Delta.exe`
	d.Scan.ANRDll = `C:\g\delta\OptiScaler\custom\Apex Neural Rendering FG\dlssg\nvngx_dlssg.dll`
	day := time.Date(2026, 11, 17, 21, 4, 5, 0, time.Local)
	a.rec.recordings = []*record.Recording{
		{Name: "20261117-210405-delta", Time: day, Game: "Delta.exe", Feature: "fg", Width: 3840, Height: 2160, Frames: 600, Dropped: 4, Bytes: 15_600_000_000},
		{Name: "20261117-200000-delta-sr", Time: day.Add(-time.Hour), Game: "Delta.exe", Feature: "sr"},
		{Name: "20261116-120000-other", Time: day.Add(-30 * time.Hour), Game: "Other.exe", Feature: "fg"},
	}
	a.rec.compared = map[string]bool{}
	a.rec.tools = record.Tools{FFmpeg: `C:\s\ffmpeg.exe`, FFprobe: `C:\s\ffprobe.exe`, Scoop: `C:\s\scoop.cmd`}
	return a
}

func open(t *testing.T, a *app, game string) *ui.Tester {
	t.Helper()
	tt := ui.NewTester(a.view, 1040, 1400)
	if err := tt.Click(game); err != nil {
		t.Fatal(err)
	}
	return tt
}

func TestRecordSectionOfAGameThatRecords(t *testing.T) {
	a := recordApp(t)
	tt := open(t, a, "Delta")
	for _, want := range []string{
		T("rec.title"), T("rec.nrrecord.on"), T("rec.uninstall"), T("rec.found"),
		T("rec.recordings.count", 2, 3), "2026-11-17 21:04:05", "3840 × 2160 · 600 frames · 15.6 GB · 4 dropped",
		T("rec.compare"), T("rec.item.upscaler"), T("rec.item.cli"),
	} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText(T("rec.install")) || tt.HasText(T("rec.scoop.install")) || tt.HasText(T("rec.why.anr")) {
		t.Errorf("offers what is already there: %q", tt.Texts())
	}

	a.rec.compared["20261117-210405-delta"] = true
	tt.Frame()
	if !tt.HasText(T("rec.compare.open")) || !tt.HasText(T("rec.compare.again")) || tt.HasText(T("rec.compare")) {
		t.Errorf("a compared recording should open its comparison: %q", tt.Texts())
	}

	// OptiScaler.ini routed back to ANR FG, as after ANR FG is installed again.
	a.games[3].Scan.NRRecord = false
	tt.Frame()
	if !tt.HasText(T("rec.nrrecord.off")) || !tt.HasText(T("rec.reinstall")) {
		t.Errorf("nr-record that no longer gets frame generation should offer to install again: %q", tt.Texts())
	}
}

func TestRecordSectionAsksForMissingTools(t *testing.T) {
	a := recordApp(t)
	a.rec.tools = record.Tools{Scoop: `C:\s\scoop.cmd`}
	tt := open(t, a, "Delta")
	if !tt.HasText(T("rec.scoop.install")) || tt.HasText(T("rec.scoop.get")) {
		t.Errorf("with Scoop, FFmpeg should be offered through it: %q", tt.Texts())
	}

	a.rec.tools = record.Tools{}
	tt.Frame()
	if !tt.HasText(T("rec.ffmpeg.noscoop")) || !tt.HasText(T("rec.scoop.get")) || !tt.HasText(T("rec.check")) || tt.HasText(T("rec.scoop.install")) {
		t.Errorf("without Scoop, the user should be sent to install it: %q", tt.Texts())
	}
}

func TestRecordSectionNeedsTheFullPackage(t *testing.T) {
	a := recordApp(t)
	tt := open(t, a, "Alpha") // no ANR FG
	if !tt.HasText(T("rec.nrrecord.needs")) || tt.HasText(T("rec.install")) || tt.HasText(T("rec.ffmpeg")) {
		t.Errorf("a game without the full package cannot record: %q", tt.Texts())
	}
	a.games[3].Recorder = false
	a.rec.recordings = nil
	if err := tt.Click("Delta"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText(T("rec.install")) || tt.HasText(T("rec.ffmpeg")) {
		t.Errorf("a game with the full package should offer nr-record alone: %q", tt.Texts())
	}
}

package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"anr-installer/internal/pkg"
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
		{Name: "20261117-210405-delta", Time: day, Game: "Delta.exe", Feature: "fg", Width: 3840, Height: 2160, Frames: 600, Dropped: 4, Bytes: 15_600_000_000, Seconds: 10},
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
		T("rec.recordings.count", 2, 3), "2026-11-17 21:04:05", "10 s · 3840 × 2160 · 600 frames · 15.6 GB · 4 dropped", T("rec.slow"),
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
	if !tt.HasText(T("rec.compare.open")) || tt.HasText(T("rec.compare")) {
		t.Errorf("a compared recording should open its comparison: %q", tt.Texts())
	}

	// OptiScaler.ini routed back to ANR FG, as after ANR FG is installed again.
	a.games[3].Scan.NRRecord = false
	tt.Frame()
	if !tt.HasText(T("rec.nrrecord.off")) || !tt.HasText(T("rec.reinstall")) {
		t.Errorf("nr-record that no longer gets frame generation should offer to install again: %q", tt.Texts())
	}
}

// The menu beside a recording's button replays a longer part of it, and deletes it.
func TestRecordingMenuOffersSpansAndDelete(t *testing.T) {
	a := recordApp(t)
	tt := open(t, a, "Delta")
	if err := tt.Click(T("rec.more")); err != nil {
		t.Fatal(err)
	}
	menu := tt.Menu()
	want := []string{T("rec.span", 5), T("rec.span.all", "10"), T("rec.delete")}
	for _, w := range want {
		if !slices.Contains(menu, w) {
			t.Errorf("the menu %q lacks %q", menu, w)
		}
	}
	if slices.Contains(menu, T("rec.span", 15)) {
		t.Errorf("a 10 s recording has no first 15 seconds: %q", menu)
	}
	tt.CloseMenu()

	// A recording whose length cannot be read may be longer than any span: every span is offered, and the
	// whole of it without a length.
	a.rec.recordings[0].Seconds = 0
	tt.Frame()
	if err := tt.Click(T("rec.more")); err != nil {
		t.Fatal(err)
	}
	menu = tt.Menu()
	for _, w := range []string{T("rec.span", 5), T("rec.span", 15), T("rec.span.whole")} {
		if !slices.Contains(menu, w) {
			t.Errorf("the menu %q of a recording of unknown length lacks %q", menu, w)
		}
	}
	if got := seconds(0.117); got != "0.1" {
		t.Errorf("8 frames last %q s", got)
	}
}

// mainThread stands in for the window's main thread: a.update queues work on it later, as win.Update does, and
// run does that work on the test's goroutine until done holds.
type mainThread chan func()

func onMainThread(a *app) mainThread {
	m := make(mainThread, 8)
	a.update = func(fn func()) { m <- fn }
	return m
}

func (m mainThread) run(t *testing.T, done func() bool) {
	t.Helper()
	timeout := time.After(2 * time.Second)
	for !done() {
		select {
		case fn := <-m:
			fn()
		case <-timeout:
			t.Fatal("the work in the background did not finish")
		}
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Deleting a recording waits for a task that runs, then is the task in progress: it removes the recording's folder
// and its comparison's, leaves the other recording, and ends its task.
func TestDeleteRecording(t *testing.T) {
	a := recordApp(t)
	m := onMainThread(a)
	for _, name := range []string{"20261117-210405-delta", "20261116-120000-other"} {
		writeFile(t, filepath.Join(record.RecordingsDir(), name, "record.json"), `{"feature": "fg"}`)
		writeFile(t, filepath.Join(pkg.CacheDir(), "compare", name, "cmp.compare.json"), "{}")
	}
	a.rec.recordings = record.Recordings(record.RecordingsDir())
	gone, kept := a.rec.recordings[0], a.rec.recordings[1]
	a.busy = true
	a.removeRecording(a.games[3], gone)
	m.run(t, func() bool { return len(m) == 0 })
	if a.log.last() != nil {
		t.Fatal("a recording was deleted while another task ran")
	}
	a.busy = false
	a.removeRecording(a.games[3], gone)
	m.run(t, func() bool { return a.busy })
	m.run(t, func() bool { return !a.busy && len(a.rec.recordings) == 1 })
	for p, want := range map[string]bool{
		gone.Dir: false, (&record.Comparison{Recording: gone}).Out(): false,
		kept.Dir: true, (&record.Comparison{Recording: kept}).Out(): true,
	} {
		if _, err := os.Stat(p); (err == nil) != want {
			t.Errorf("%s: kept %v, want %v", p, err == nil, want)
		}
	}
	if e := a.log.last(); e == nil || e.task == nil || e.task.running() || e.task.failed {
		t.Errorf("the deletion's task did not end well: %+v", e)
	}
}

func TestPlayButton(t *testing.T) {
	a := recordApp(t)
	tt := open(t, a, "Delta") // its .exe is known
	if !tt.HasText(T("play")) {
		t.Errorf("no Play button: %q", tt.Texts())
	}
	if err := tt.Click("Alpha"); err != nil { // neither a Steam id nor an .exe found
		t.Fatal(err)
	}
	if tt.HasText(T("play")) {
		t.Error("Play is offered for a game that cannot be started")
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

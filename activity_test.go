package main

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestTaskLifecycle(t *testing.T) {
	lang = "en"
	start := time.Date(2026, 11, 18, 15, 4, 0, 0, time.Local)
	tk := &task{title: "Delta · Upgrade to V.B19", start: start}
	tk.step(levelInfo, "Installing V.B19 (full package)…")
	for i := range maxOutput + 5 {
		tk.out(fmt.Sprint(i))
	}
	if !tk.running() || len(tk.output) != maxOutput || tk.dropped != 5 || tk.output[0] != "5" {
		t.Fatalf("running %v, %d lines kept, %d dropped, first %q", tk.running(), len(tk.output), tk.dropped, tk.output[0])
	}
	tk.finish(start.Add(65*time.Second), failure(nil))
	if tk.running() || tk.failed || tk.open || len(tk.steps) != 1 {
		t.Errorf("a success: failed %v, open %v, steps %v", tk.failed, tk.open, tk.steps)
	}
	if got := durationText(tk.took(time.Now())); got != "1 min 5 s" {
		t.Errorf("duration %q", got)
	}

	bad := &task{start: start}
	bad.out("Access denied")
	bad.finish(start.Add(12*time.Second), failure(errors.New("exit status 1")))
	if !bad.failed || !bad.open || bad.steps[0] != (line{levelError, "Couldn't complete: exit status 1"}) {
		t.Errorf("a failure should end with its error step and show its output: %+v", bad)
	}
	for d, want := range map[time.Duration]string{
		0:                        "0 s",
		12400 * time.Millisecond: "12 s",
		2*time.Hour + 5*time.Minute + 9*time.Second: "2 h 5 min",
		-time.Second: "0 s", // the clock set back
	} {
		if got := durationText(d); got != want {
			t.Errorf("durationText(%v) = %q, want %q", d, got, want)
		}
	}
}

// testLog is a log with a task that succeeded, one that failed, and a note.
func testLog() activityLog {
	at := time.Date(2026, 11, 18, 15, 4, 5, 0, time.Local)
	ok := &task{title: "Alpha · Install V.B19", start: at, end: at.Add(12 * time.Second), output: []string{"Copying files", "Done"}}
	ok.step(levelInfo, "ANR FG V.B19 is installed in C:\\g\\alpha.")
	bad := &task{title: "Delta · Upgrade to V.B19", start: at.Add(time.Minute), output: []string{"Access to nr_fg.dll is denied"}, dropped: 3}
	bad.step(levelWarning, "Recording is now off for this game.")
	bad.finish(at.Add(time.Minute+65*time.Second), failure(errors.New("exit status 1")))
	var l activityLog
	l.add(ok)
	l.add(bad)
	l.entries = append(l.entries, entry{at: at.Add(3 * time.Minute), note: line{levelInfo, "Removed Epic (D:\\Epic) from the list."}})
	return l
}

func TestLogText(t *testing.T) {
	lang = "en"
	l := testLog()
	want := `2026-11-18 15:04:05  Alpha · Install V.B19 · Succeeded · 12 s
  ANR FG V.B19 is installed in C:\g\alpha.
  > Copying files
  > Done
2026-11-18 15:05:05  Delta · Upgrade to V.B19 · Failed · 1 min 5 s
  Recording is now off for this game.
  Couldn't complete: exit status 1
  Output: last 1 of 4 lines
  > Access to nr_fg.dll is denied
2026-11-18 15:07:05  Removed Epic (D:\Epic) from the list.
`
	if got := l.text(); got != want {
		t.Errorf("text:\n%s\nwant:\n%s", got, want)
	}
}

func TestActivityPanel(t *testing.T) {
	a := testApp(t)
	a.log, a.logOpen = testLog(), true
	tt := ui.NewTester(a.view, 1040, 760)
	for _, s := range []string{
		"Delta · Upgrade to V.B19", "Couldn't complete: exit status 1", "Access to nr_fg.dll is denied",
		"Hide output (last 1 of 4 lines)", "Removed Epic (D:\\Epic) from the list.",
	} {
		if !tt.HasText(s) {
			t.Errorf("a failed task shows its header, its error step and its output: no %q in %q", s, tt.Texts())
		}
	}
	if tt.HasText("Copying files") {
		t.Error("a task that succeeded shows its output before it is unfolded")
	}
	if err := tt.Click("Show output (2 lines)"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Copying files") || !tt.HasText("Hide output") {
		t.Errorf("the output did not unfold: %q", tt.Texts())
	}

	if err := tt.Click(T("log.copy")); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != a.log.text() || !tt.HasText(T("log.copied")) {
		t.Errorf("Copy log: clipboard %q, texts %q", tt.Clipboard(), tt.Texts())
	}
	tt.Advance(copiedFor)
	if !tt.HasText(T("log.copy")) {
		t.Error("the label should go back to Copy log")
	}
}

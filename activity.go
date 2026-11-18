package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// maxOutput is how many lines a program printed a task keeps, the last ones: a long replay prints far more.
const maxOutput = 2000

// level is what a line of the activity log tells: news, something to heed, or a failure.
type level uint8

const (
	levelInfo level = iota
	levelWarning
	levelError
)

// line is a line the app itself wrote to the log.
type line struct {
	level level
	text  string
}

// task is one run of work in the background, as an install, a replay or a deletion: its steps are the app's own
// lines, and its output what the programs it started printed. It changes on the main thread only.
type task struct {
	title      string    // the game and the action, as "Delta · Upgrade to V.B19"
	start, end time.Time // end is zero while it runs
	failed     bool
	steps      []line
	output     []string // the last maxOutput lines
	dropped    int      // the lines before them, no longer kept
	open       bool     // the output shows, unfolded
}

func (tk *task) running() bool { return tk.end.IsZero() }

func (tk *task) step(lv level, text string) { tk.steps = append(tk.steps, line{lv, text}) }

// out keeps a line a program printed, dropping the oldest beyond maxOutput.
func (tk *task) out(text string) {
	tk.output = append(tk.output, text)
	if len(tk.output) > maxOutput {
		tk.output = tk.output[1:]
		tk.dropped++
	}
}

// finish ends the task at at, failed with failure as its last step unless failure is "". A failed task unfolds
// its output, where the cause usually is.
func (tk *task) finish(at time.Time, failure string) {
	tk.end = at
	if failure != "" {
		tk.step(levelError, failure)
		tk.failed, tk.open = true, true
	}
}

// took is how long the task ran, until now while it runs.
func (tk *task) took(now time.Time) time.Duration {
	if tk.running() {
		return now.Sub(tk.start)
	}
	return tk.end.Sub(tk.start)
}

// failure is the last step of a task that ended with err, or "" when it succeeded.
func failure(err error) string {
	if err == nil {
		return ""
	}
	return T("failed", err.Error())
}

// durationText says how long a task took: "12 s", "1 min 5 s", or "2 h 5 min".
func durationText(d time.Duration) string {
	s := max(int(d.Round(time.Second)/time.Second), 0)
	switch {
	case s < 60:
		return T("dur.s", s)
	case s < 3600:
		return T("dur.min", s/60, s%60)
	}
	return T("dur.h", s/3600, s%3600/60)
}

// entry is an item of the log: a task, or else a note, a line of no task, as a game removed from the list.
type entry struct {
	task *task
	at   time.Time // the note's
	note line
}

// activityLog is what the activity panel shows, in the order it happened, and the panel's state.
type activityLog struct {
	entries    []entry
	rows       []logRow // the panel's rows, made from the entries each frame
	list       ui.ListState
	copiedAt   time.Time // when Copy log last wrote to the clipboard, which its label tells for a moment
	copyFailed bool
}

func (l *activityLog) add(tk *task) { l.entries = append(l.entries, entry{task: tk}) }

// last returns the newest entry, or nil.
func (l *activityLog) last() *entry {
	if len(l.entries) == 0 {
		return nil
	}
	return &l.entries[len(l.entries)-1]
}

// running returns the task that runs, or nil. Tasks run one at a time, so only the newest can.
func (l *activityLog) running() *task {
	for i := len(l.entries) - 1; i >= 0; i-- {
		if tk := l.entries[i].task; tk != nil {
			if tk.running() {
				return tk
			}
			return nil
		}
	}
	return nil
}

// stamp writes a time of the log in full, for the copied text and the time column's tooltip.
func stamp(t time.Time) string { return t.Format("2006-01-02 15:04:05") }

// text is the log as plain text, for Copy log: a task's header with its start, title and result, its steps, then
// its output after "> "; a note is one line after its time.
func (l *activityLog) text() string {
	var b strings.Builder
	for _, e := range l.entries {
		tk := e.task
		if tk == nil {
			fmt.Fprintf(&b, "%s  %s\n", stamp(e.at), e.note.text)
			continue
		}
		result := T("log.running")
		switch {
		case tk.failed:
			result = T("log.failed") + " · " + durationText(tk.took(tk.end))
		case !tk.running():
			result = T("log.ok") + " · " + durationText(tk.took(tk.end))
		}
		fmt.Fprintf(&b, "%s  %s · %s\n", stamp(tk.start), tk.title, result)
		for _, s := range tk.steps {
			fmt.Fprintf(&b, "  %s\n", s.text)
		}
		if tk.dropped > 0 {
			fmt.Fprintf(&b, "  %s\n", T("log.output.kept", len(tk.output), len(tk.output)+tk.dropped))
		}
		for _, o := range tk.output {
			fmt.Fprintf(&b, "  > %s\n", o)
		}
	}
	return b.String()
}

// note adds a line of no task, written at at, to the log.
func (l *activityLog) note(at time.Time, lv level, text string) {
	l.entries = append(l.entries, entry{at: at, note: line{lv, text}})
}

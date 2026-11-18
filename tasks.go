package main

import (
	"fmt"
	"time"
)

// gameTask is the title of a task in g: the game, then the action.
func gameTask(g *game, action string) string { return g.Name + " · " + action }

// begin adds a task named title, which runs from now, to the log.
func (a *app) begin(title string) *task {
	tk := &task{title: title, start: time.Now()}
	a.log.add(tk)
	return tk
}

// end ends tk as finish does. A failure opens the log, which shows it.
func (a *app) end(tk *task, failure string) {
	tk.finish(time.Now(), failure)
	a.status = ""
	if failure != "" {
		a.logOpen = true
	}
}

// note adds a line of no task to the log.
func (a *app) note(lv level, text string) { a.log.note(time.Now(), lv, text) }

// reporter writes the lines of work in the background to its task, through the main thread. Its first step
// opens the log, unless the log opened as the task began.
type reporter struct {
	a      *app
	tk     *task
	opened bool // main thread only
}

// say adds a step, which the status shows too.
func (r *reporter) say(format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	r.a.win.Update(func() {
		if !r.opened {
			r.opened, r.a.logOpen = true, true
		}
		r.a.status = text
		r.tk.step(levelInfo, text)
	})
}

// warn adds a step to heed.
func (r *reporter) warn(text string) { r.a.win.Update(func() { r.tk.step(levelWarning, text) }) }

// out adds a line a program printed.
func (r *reporter) out(text string) { r.a.win.Update(func() { r.tk.out(text) }) }

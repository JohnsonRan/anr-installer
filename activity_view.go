package main

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// copiedFor is how long Copy log's label tells whether the log got to the clipboard.
const copiedFor = 2 * time.Second

// What a row of the activity panel shows.
type rowKind uint8

const (
	rowTask   rowKind = iota // a task's header: its mark, title, start and duration
	rowStep                  // a step of a task
	rowFold                  // the row that unfolds a task's output
	rowOutput                // a line of a task's output
	rowNote
)

// logRow is a row of the activity panel: of entry, its step or output line i. It is the row's key too, with i
// counted from the task's first output line, kept or not, so that a row keeps its state as older lines drop.
type logRow struct {
	kind  rowKind
	entry int
	i     int
}

// layout makes the panel's rows from the entries.
func (l *activityLog) layout() []logRow {
	rows := l.rows[:0]
	for e, en := range l.entries {
		tk := en.task
		if tk == nil {
			rows = append(rows, logRow{rowNote, e, 0})
			continue
		}
		rows = append(rows, logRow{rowTask, e, 0})
		for i := range tk.steps {
			rows = append(rows, logRow{rowStep, e, i})
		}
		if len(tk.output) == 0 {
			continue
		}
		rows = append(rows, logRow{rowFold, e, 0})
		if tk.open {
			for i := range tk.output {
				rows = append(rows, logRow{rowOutput, e, i})
			}
		}
	}
	l.rows = rows
	return rows
}

// activity is the log, folded to a bar until opened: the bar tells what runs, or how the newest entry ended. It
// opens as a run starts.
func (a *app) activity(c *ui.Context, t *ui.Theme) {
	ui.Divider(c)
	ui.Column(c).Background(t.Text.Alpha(0.025)).Children(func() {
		head := ui.Row(c.Key("activity")).Padding(padRowY, padPageX).Gap(gapControls).AlignItems(ui.Center).Cursor(ui.CursorPointer).Label(T("log"))
		if head.Clicked() {
			a.logOpen = !a.logOpen
		}
		head.Children(func() {
			arrow := "▸"
			if a.logOpen {
				arrow = "▾"
			}
			// In the column of the tasks' marks, so that the bar's label starts where their titles do.
			ui.Text(c, arrow).TextColor(t.TextMuted).Width(logMark).TextAlign(ui.Center)
			roleBodyStrong.on(t, ui.Text(c, T("log")))
			a.activitySummary(c, t)
			if a.logOpen && len(a.log.entries) > 0 {
				a.copyLogButton(c)
				ui.Button(c, T("clear")).Disabled(a.busy).OnClick(func() { a.log = activityLog{} })
			}
		})
		if a.logOpen {
			a.activityPanel(c, t)
		}
	})
}

// activitySummary is the bar's line: the status and progress of the task that runs, else the newest entry, a
// task's result or a note.
func (a *app) activitySummary(c *ui.Context, t *ui.Theme) {
	if a.busy {
		status := a.status
		if tk := a.log.running(); status == "" && tk != nil {
			status = tk.title
		}
		roleSmall.on(t, ui.Text(c, status)).TextColor(t.TextMuted).MaxLines(1).Ellipsis("…").Grow(1).Shrink(1)
		if a.progress >= 0 {
			ui.Progress(c, a.progress).Width(progressWidth).Label(a.status)
		} else {
			ui.Spinner(c).Label(T("busy"))
		}
		return
	}
	e := a.log.last()
	text, color := "", t.TextMuted
	switch {
	case e == nil:
	case e.task == nil:
		text, color = e.note.text, levelColor(t, e.note.level, t.TextMuted)
	case e.task.failed:
		text, color = e.task.title+" · "+e.task.steps[len(e.task.steps)-1].text, t.Danger
	default:
		text = e.task.title
	}
	ui.Row(c).Gap(gapControls).AlignItems(ui.Center).Grow(1).Shrink(1).Children(func() {
		if e != nil && e.task != nil {
			taskMark(c, t, e.task)
		}
		roleSmall.on(t, ui.Text(c, text)).TextColor(color).MaxLines(1).Ellipsis("…").Tooltip(text).Grow(1).Shrink(1)
	})
}

// taskMark is a task's state: a spinner while it runs, then a check, or a cross when it failed.
func taskMark(c *ui.Context, t *ui.Theme, tk *task) ui.Element {
	switch {
	case tk.running():
		return ui.Spinner(c).Size(logMark, logMark).Label(T("log.running"))
	case tk.failed:
		return ui.Icon(c, iconCross).Size(logMark, logMark).TextColor(t.Danger).Label(T("log.failed"))
	}
	return ui.Icon(c, iconCheck).Size(logMark, logMark).TextColor(t.Success).Label(T("log.ok"))
}

// copyLogButton copies the log as text, and its label tells for a moment whether that worked, as GitHub's copy
// buttons do.
func (a *app) copyLogButton(c *ui.Context) {
	l := &a.log
	label := T("log.copy")
	if since := c.Now().Sub(l.copiedAt); !l.copiedAt.IsZero() && since < copiedFor {
		label = T("log.copied")
		if l.copyFailed {
			label = T("log.copy.failed")
		}
		c.After(copiedFor - since)
	}
	if ui.Button(c, label).Clicked() {
		now := c.Now()
		c.WriteClipboardThen(l.text(), func(err error) { l.copiedAt, l.copyFailed = now, err != nil })
	}
}

// activityPanel lists the log, the newest at the bottom, which it follows as lines come.
func (a *app) activityPanel(c *ui.Context, t *ui.Theme) {
	l := &a.log
	rows := l.layout()
	l.list.FollowEnd = true
	l.list.Key = func(i int) any {
		r := l.rows[i]
		if r.kind == rowOutput {
			r.i += l.entries[r.entry].task.dropped
		}
		return r
	}
	if l.running() != nil {
		c.After(time.Second) // the running task's duration counts up
	}
	list := ui.List(c, &l.list, len(rows), func(i int) { a.logRow(c, t, rows[i]) })
	list.Height(activityHeight).PaddingY(space1).Selectable().Children(func() {
		if len(rows) == 0 {
			roleSmall.on(t, ui.Text(c, T("log.empty"))).TextColor(t.TextMuted).Padding(padRowY, padPageX, padRowY, padLogText)
		}
	})
}

func (a *app) logRow(c *ui.Context, t *ui.Theme, r logRow) {
	e := &a.log.entries[r.entry]
	tk := e.task
	switch r.kind {
	case rowTask:
		ui.Row(c).Gap(gapControls).AlignItems(ui.Baseline).Padding(gapLogEntries, padPageX, 0, padPageX).Children(func() {
			taskMark(c, t, tk).AlignSelf(ui.Center)
			roleBodyStrong.on(t, ui.Text(c, tk.title)).MaxLines(1).Ellipsis("…").Tooltip(tk.title).Grow(1).Shrink(1)
			logTimes(c, t, tk.start, durationText(tk.took(c.Now())))
		})
	case rowStep:
		s := tk.steps[r.i]
		roleSmall.on(t, ui.Text(c, s.text)).TextColor(levelColor(t, s.level, t.Text)).Padding(0, padPageX, 0, padLogText)
	case rowFold:
		a.outputFold(c, t, tk)
	case rowOutput:
		roleSmall.on(t, ui.Text(c, tk.output[r.i])).Font("monospace").TextColor(t.TextMuted).Padding(0, padPageX, 0, padLogText)
	case rowNote:
		ui.Row(c).Gap(gapControls).AlignItems(ui.Baseline).Padding(gapLogEntries, padPageX, 0, padLogText).Children(func() {
			roleSmall.on(t, ui.Text(c, e.note.text)).TextColor(levelColor(t, e.note.level, t.Text)).Grow(1).Shrink(1)
			logTimes(c, t, e.at, "")
		})
	}
}

// logTimes ends a row with when it began and how long it took, in columns that line up from row to row.
func logTimes(c *ui.Context, t *ui.Theme, at time.Time, took string) {
	column := func(s string, width float32) ui.Element {
		return roleSmall.on(t, ui.Text(c, s)).TextColor(t.TextMuted).FontFeatures("tnum").NoWrap().TextAlign(ui.End).MinWidth(width)
	}
	column(at.Format("15:04"), logTimeW).Tooltip(stamp(at))
	column(took, logTookW)
}

// outputFold is the row that shows a task's output, or hides it, and says how many lines it holds.
func (a *app) outputFold(c *ui.Context, t *ui.Theme, tk *task) {
	n, total := len(tk.output), len(tk.output)+tk.dropped
	arrow, label := "▸", T("log.output.show", n)
	switch {
	case tk.open && tk.dropped > 0:
		arrow, label = "▾", T("log.output.hide.last", n, total)
	case tk.open:
		arrow, label = "▾", T("log.output.hide")
	case tk.dropped > 0:
		label = T("log.output.show.last", n, total)
	case n == 1:
		label = T("log.output.show.one")
	}
	ui.Row(c).Padding(0, padPageX, 0, padLogText).Children(func() {
		b := ui.ButtonBase(c).Gap(gapLines).Label(label)
		color := t.TextMuted
		if b.Hovered() {
			color = t.Text
		}
		// The rows below come or go: the row clicked stays where it is rather than the list following its end.
		b.OnClick(func() { tk.open = !tk.open; a.log.list.HoldPlace() }).Children(func() {
			ui.Text(c, arrow).TextColor(color)
			roleSmall.on(t, ui.Text(c, label)).TextColor(color)
		})
	})
}

// levelColor is the color of a line of level lv: info takes the color given.
func levelColor(t *ui.Theme, lv level, info ui.Color) ui.Color {
	switch lv {
	case levelWarning:
		return t.Warning
	case levelError:
		return t.Danger
	}
	return info
}

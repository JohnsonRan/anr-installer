package activity

import (
	"time"

	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
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
func (l *Log) layout() []logRow {
	rows := l.rows[:0]
	for e, en := range l.entries {
		tk := en.Task
		if tk == nil {
			rows = append(rows, logRow{rowNote, e, 0})
			continue
		}
		rows = append(rows, logRow{rowTask, e, 0})
		for i := range tk.Steps {
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

// Work is the task in progress as the activity bar tells it: whether one runs, what it does now, and how far it
// is, -1 unknown or -2 none.
type Work struct {
	Busy     bool
	Status   string
	Progress float64
}

// View is the log, folded to a bar until open: the bar tells what runs, or how the newest entry ended.
func (l *Log) View(c *ui.Context, t *ui.Theme, open *bool, w Work) {
	ui.Divider(c)
	ui.Column(c).Background(t.Text.Alpha(0.025)).Children(func() {
		head := ui.Row(c.Key("activity")).Padding(style.PadRowY, style.PadPageX).Gap(style.GapControls).AlignItems(ui.Center).Cursor(ui.CursorPointer).Label(i18n.T("log"))
		if head.Clicked() {
			*open = !*open
		}
		head.Children(func() {
			arrow := "▸"
			if *open {
				arrow = "▾"
			}
			// In the column of the tasks' marks, so that the bar's label starts where their titles do.
			ui.Text(c, arrow).TextColor(t.TextMuted).Width(style.LogMark).TextAlign(ui.Center)
			style.RoleBodyStrong.On(t, ui.Text(c, i18n.T("log")))
			l.summary(c, t, w)
			if *open && len(l.entries) > 0 {
				l.copyButton(c)
				ui.Button(c, i18n.T("clear")).Disabled(w.Busy).OnClick(func() { *l = Log{} })
			}
		})
		if *open {
			l.panel(c, t)
		}
	})
}

// summary is the bar's line: the status and progress of the task that runs, else the newest entry, a task's
// result or a note.
func (l *Log) summary(c *ui.Context, t *ui.Theme, w Work) {
	if w.Busy {
		status := w.Status
		if tk := l.running(); status == "" && tk != nil {
			status = tk.title
		}
		style.RoleSmall.On(t, ui.Text(c, status)).TextColor(t.TextMuted).MaxLines(1).Ellipsis("…").Grow(1).Shrink(1)
		if w.Progress >= 0 {
			ui.Progress(c, w.Progress).Width(style.ProgressWidth).Label(w.Status)
		} else {
			ui.Spinner(c).Label(i18n.T("busy"))
		}
		return
	}
	e := l.Last()
	text, color := "", t.TextMuted
	switch {
	case e == nil:
	case e.Task == nil:
		text, color = e.note.Text, levelColor(t, e.note.level, t.TextMuted)
	case e.Task.Failed:
		text, color = e.Task.title+" · "+e.Task.Steps[len(e.Task.Steps)-1].Text, t.Danger
	default:
		text = e.Task.title
	}
	ui.Row(c).Gap(style.GapControls).AlignItems(ui.Center).Grow(1).Shrink(1).Children(func() {
		if e != nil && e.Task != nil {
			taskMark(c, t, e.Task)
		}
		style.RoleSmall.On(t, ui.Text(c, text)).TextColor(color).MaxLines(1).Ellipsis("…").Tooltip(text).Grow(1).Shrink(1)
	})
}

// taskMark is a task's state: a spinner while it runs, then a check, or a cross when it failed.
func taskMark(c *ui.Context, t *ui.Theme, tk *Task) ui.Element {
	switch {
	case tk.Running():
		return ui.Spinner(c).Size(style.LogMark, style.LogMark).Label(i18n.T("log.running"))
	case tk.Failed:
		return ui.Icon(c, style.IconCross).Size(style.LogMark, style.LogMark).TextColor(t.Danger).Label(i18n.T("log.failed"))
	}
	return ui.Icon(c, style.IconCheck).Size(style.LogMark, style.LogMark).TextColor(t.Success).Label(i18n.T("log.ok"))
}

// copyButton copies the log as text, and its label tells for a moment whether that worked, as GitHub's copy
// buttons do.
func (l *Log) copyButton(c *ui.Context) {
	label := i18n.T("log.copy")
	if since := c.Now().Sub(l.copiedAt); !l.copiedAt.IsZero() && since < copiedFor {
		label = i18n.T("log.copied")
		if l.copyFailed {
			label = i18n.T("log.copy.failed")
		}
		c.After(copiedFor - since)
	}
	if ui.Button(c, label).Clicked() {
		now := c.Now()
		c.WriteClipboardThen(l.text(), func(err error) { l.copiedAt, l.copyFailed = now, err != nil })
	}
}

// panel lists the log, the newest at the bottom, which it follows as lines come.
func (l *Log) panel(c *ui.Context, t *ui.Theme) {
	rows := l.layout()
	l.list.FollowEnd = true
	l.list.Key = func(i int) any {
		r := l.rows[i]
		if r.kind == rowOutput {
			r.i += l.entries[r.entry].Task.dropped
		}
		return r
	}
	if l.running() != nil {
		c.After(time.Second) // the running task's duration counts up
	}
	list := ui.List(c, &l.list, len(rows), func(i int) { l.row(c, t, rows[i]) })
	list.Height(style.ActivityHeight).PaddingY(style.Space1).Selectable().Children(func() {
		if len(rows) == 0 {
			style.RoleSmall.On(t, ui.Text(c, i18n.T("log.empty"))).TextColor(t.TextMuted).Padding(style.PadRowY, style.PadPageX, style.PadRowY, style.PadLogText)
		}
	})
}

func (l *Log) row(c *ui.Context, t *ui.Theme, r logRow) {
	e := &l.entries[r.entry]
	tk := e.Task
	switch r.kind {
	case rowTask:
		ui.Row(c).Gap(style.GapControls).AlignItems(ui.Baseline).Padding(style.GapLogEntries, style.PadPageX, 0, style.PadPageX).Children(func() {
			taskMark(c, t, tk).AlignSelf(ui.Center)
			style.RoleBodyStrong.On(t, ui.Text(c, tk.title)).MaxLines(1).Ellipsis("…").Tooltip(tk.title).Grow(1).Shrink(1)
			logTimes(c, t, tk.start, durationText(tk.took(c.Now())))
		})
	case rowStep:
		s := tk.Steps[r.i]
		style.RoleSmall.On(t, ui.Text(c, s.Text)).TextColor(levelColor(t, s.level, t.Text)).Padding(0, style.PadPageX, 0, style.PadLogText)
	case rowFold:
		l.outputFold(c, t, tk)
	case rowOutput:
		style.RoleSmall.On(t, ui.Text(c, tk.output[r.i])).Font("monospace").TextColor(t.TextMuted).Padding(0, style.PadPageX, 0, style.PadLogText)
	case rowNote:
		ui.Row(c).Gap(style.GapControls).AlignItems(ui.Baseline).Padding(style.GapLogEntries, style.PadPageX, 0, style.PadLogText).Children(func() {
			style.RoleSmall.On(t, ui.Text(c, e.note.Text)).TextColor(levelColor(t, e.note.level, t.Text)).Grow(1).Shrink(1)
			logTimes(c, t, e.at, "")
		})
	}
}

// logTimes ends a row with when it began and how long it took, in columns that line up from row to row.
func logTimes(c *ui.Context, t *ui.Theme, at time.Time, took string) {
	column := func(s string, width float32) ui.Element {
		return style.RoleSmall.On(t, ui.Text(c, s)).TextColor(t.TextMuted).FontFeatures("tnum").NoWrap().TextAlign(ui.End).MinWidth(width)
	}
	column(at.Format("15:04"), style.LogTimeW).Tooltip(stamp(at))
	column(took, style.LogTookW)
}

// outputFold is the row that shows a task's output, or hides it, and says how many lines it holds.
func (l *Log) outputFold(c *ui.Context, t *ui.Theme, tk *Task) {
	n, total := len(tk.output), len(tk.output)+tk.dropped
	arrow, label := "▸", i18n.T("log.output.show", n)
	switch {
	case tk.open && tk.dropped > 0:
		arrow, label = "▾", i18n.T("log.output.hide.last", n, total)
	case tk.open:
		arrow, label = "▾", i18n.T("log.output.hide")
	case tk.dropped > 0:
		label = i18n.T("log.output.show.last", n, total)
	case n == 1:
		label = i18n.T("log.output.show.one")
	}
	ui.Row(c).Padding(0, style.PadPageX, 0, style.PadLogText).Children(func() {
		b := ui.ButtonBase(c).Gap(style.GapLines).Label(label)
		color := t.TextMuted
		if b.Hovered() {
			color = t.Text
		}
		// The rows below come or go: the row clicked stays where it is rather than the list following its end.
		b.OnClick(func() { tk.open = !tk.open; l.list.HoldPlace() }).Children(func() {
			ui.Text(c, arrow).TextColor(color)
			style.RoleSmall.On(t, ui.Text(c, label)).TextColor(color)
		})
	})
}

// levelColor is the color of a line of level lv: info takes the color given.
func levelColor(t *ui.Theme, lv Level, info ui.Color) ui.Color {
	switch lv {
	case LevelWarning:
		return t.Warning
	case LevelError:
		return t.Danger
	}
	return info
}

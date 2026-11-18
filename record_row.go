package main

import (
	"fmt"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/internal/record"
)

// recordingRow is a recording: when it was made and what it holds, with its comparison.
func (a *app) recordingRow(c *ui.Context, t *ui.Theme, g *game, rec *record.Recording, blocked bool) {
	title := rec.Time.Format("2006-01-02 15:04:05")
	if !rec.FrameGeneration() {
		settingsRow(c, t, iconFrames, title, T("rec.item.upscaler"), func() { valueText(c, T("rec.item.cli"), t.TextMuted) })
		return
	}
	detail := T("rec.item", rec.Width, rec.Height, rec.Frames, fmt.Sprintf("%.1f GB", float64(rec.Bytes)/1e9))
	if rec.Seconds > 0 {
		detail = T("rec.item.seconds", seconds(rec.Seconds)) + " · " + detail
	}
	if rec.Dropped > 0 {
		detail += T("rec.item.dropped", rec.Dropped)
	}
	settingsRow(c, t, iconFrames, title, detail, func() {
		// The button compares the start of the recording, or opens the comparison made; its menu replays a
		// longer part, and deletes the recording.
		label := T("rec.compare")
		if a.rec.compared[rec.Name] {
			label = T("rec.compare.open")
		}
		split := ui.SplitButton(c, label)
		if a.rec.compared[rec.Name] {
			split.Button.Disabled(a.busy).OnClick(func() { a.openCompare(g, rec) })
		} else {
			split.Button.Disabled(blocked).OnClick(func() { a.compare(g, rec, compareSpans[0]) })
		}
		split.Arrow.Label(T("rec.more")).Tooltip(T("rec.more"))
		split.Arrow.PopoverMenuButton(func(m *ui.Menu) {
			for _, secs := range compareSpans {
				// A recording of unknown length may be longer than any span.
				if (rec.Seconds <= 0 || rec.Seconds > float64(secs)) && m.Item(T("rec.span", secs)).Disabled(blocked).Chosen() {
					a.compare(g, rec, secs)
				}
			}
			all := T("rec.span.whole")
			if rec.Seconds > 0 {
				all = T("rec.span.all", seconds(rec.Seconds))
			}
			if m.Item(all).Disabled(blocked).Chosen() {
				a.compare(g, rec, 0)
			}
			m.Separator()
			if m.Item(T("rec.delete")).Disabled(a.busy).Chosen() {
				a.deleteRecording(g, rec)
			}
		}, ui.PopoverMenuButtonOptions{At: ui.AnchorBottomEnd, Self: ui.AnchorTopEnd})
	})
}

// compareSpans are the lengths of a recording's start, in seconds, that a comparison replays: a replay runs
// far slower than the game did, so the first is what the button takes.
var compareSpans = []int{5, 15}

// assumedFPS is the rate of recorded frames taken for a recording whose length cannot be read.
const assumedFPS = 60

// seconds writes a length in seconds: with a decimal below 10, as 8 frames last a tenth of a second.
func seconds(s float64) string {
	if s < 10 {
		return fmt.Sprintf("%.1f", s)
	}
	return fmt.Sprintf("%.0f", s)
}

// deleteRecording deletes rec, made in g, and its comparison, after the user confirms: a recording cannot be made
// again.
func (a *app) deleteRecording(g *game, rec *record.Recording) {
	made := rec.Time.Format("2006-01-02 15:04:05")
	go func() {
		res, err := mygo.Dialog.Message(mygo.MessageOptions{
			Parent:             a.win,
			Type:               mygo.MessageWarning,
			Title:              T("rec.delete.title"),
			Message:            T("rec.delete.title"),
			Detail:             T("rec.delete.detail", made, fmt.Sprintf("%.1f GB", float64(rec.Bytes)/1e9)),
			Buttons:            []string{T("anticheat.cancel"), T("rec.delete.confirm")},
			DefaultButton:      0,
			CancelButton:       0,
			DestructiveButtons: []int{1},
		})
		if err != nil || res.Button != 1 {
			return
		}
		// Deleting tens of GB takes a while: it is the task in progress meanwhile, as an install is. Whether
		// another task runs is known on the main thread only, where win.Update runs later, so the deletion starts
		// from there.
		a.win.Update(func() {
			if a.busy {
				return
			}
			a.busy, a.progress, a.status = true, -1, T("rec.deleting")
			tk := a.begin(gameTask(g, T("rec.delete.task", made)))
			go func() {
				err := rec.Remove()
				a.win.Update(func() {
					a.busy, a.progress = false, -2
					a.end(tk, failure(err))
					a.refreshRecord()
				})
			}()
		})
	}()
}

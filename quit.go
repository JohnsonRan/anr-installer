package main

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// quitAsks reports whether closing the window or quitting asks first: only while a task runs, and never once the
// user answered "Quit anyway" (forced), nor for a quit that is final: the end of the session, which cannot wait,
// or a quit already let through, which closes the window.
func quitAsks(busy, forced, final bool) bool {
	return busy && !forced && !final
}

// askQuit brings the window forward with the question whether to quit while a task runs.
func (a *app) askQuit() {
	a.quitAsk = true
	a.showWindow()
	a.win.Invalidate()
}

// quitAnyway lets the next quit through, then quits.
func (a *app) quitAnyway() {
	a.quitAsk, a.quitForced = false, true
	mygo.App.Quit()
}

// quitDialog asks whether to quit while a task runs. Keeping it running is the default, as Escape and a click
// outside answer.
func (a *app) quitDialog(c *ui.Context, t *ui.Theme) {
	ui.Modal(c, &a.quitAsk, func() {
		ui.Column(c).Width(confirmWidth).Gap(gapSection).Children(func() {
			ui.Row(c).Gap(gapGroup).AlignItems(ui.Start).Children(func() {
				ui.Icon(c, iconWarning).Size(rowIcon, rowIcon).TextColor(t.Warning)
				ui.Column(c).Grow(1).Shrink(1).Gap(gapInCard).Children(func() {
					roleHeading.on(t, ui.Text(c, T("quit.title")))
					ui.Text(c, T("quit.detail")).TextColor(t.TextMuted)
				})
			})
			ui.Row(c).Gap(gapControls).Justify(ui.End).Children(func() {
				ui.Button(c, T("quit.anyway")).OnClick(a.quitAnyway)
				ui.PrimaryButton(c, T("quit.keep")).OnClick(func() { a.quitAsk = false })
			})
		})
	})
}

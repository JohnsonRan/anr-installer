package gui

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
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
		ui.Column(c).Width(style.ConfirmWidth).Gap(style.GapSection).Children(func() {
			ui.Row(c).Gap(style.GapGroup).AlignItems(ui.Start).Children(func() {
				ui.Icon(c, style.IconWarning).Size(style.RowIcon, style.RowIcon).TextColor(t.Warning)
				ui.Column(c).Grow(1).Shrink(1).Gap(style.GapInCard).Children(func() {
					style.RoleHeading.On(t, ui.Text(c, i18n.T("quit.title")))
					ui.Text(c, i18n.T("quit.detail")).TextColor(t.TextMuted)
				})
			})
			ui.Row(c).Gap(style.GapControls).Justify(ui.End).Children(func() {
				ui.Button(c, i18n.T("quit.anyway")).OnClick(a.quitAnyway)
				ui.PrimaryButton(c, i18n.T("quit.keep")).OnClick(func() { a.quitAsk = false })
			})
		})
	})
}

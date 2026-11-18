package gui

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
)

// The window buttons' glyphs.
var (
	iconMin     = style.MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 10"><rect x="0" y="4.5" width="10" height="1" fill="#000"/></svg>`))
	iconMax     = style.MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 10"><rect x="0.5" y="0.5" width="9" height="9" fill="none" stroke="#000"/></svg>`))
	iconRestore = style.MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 10"><path d="M2.5 2.5V0.5h7v7h-2" fill="none" stroke="#000"/><rect x="0.5" y="2.5" width="7" height="7" fill="none" stroke="#000"/></svg>`))
	iconClose   = style.MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 10"><path d="M0.5 0.5l9 9M9.5 0.5l-9 9" stroke="#000" stroke-width="1.1"/></svg>`))
)

const titleBarHeight = 52

// titleBar draws the window's title bar in place of the system's: the logo, the title, the app's version, which
// opens the About dialog, the settings button and the window buttons. Its empty parts drag the window, and a
// double click maximizes it. The ANR FG releases and their errors are on the game page, by the install button.
func (a *app) titleBar(c *ui.Context, t *ui.Theme) {
	bar := c.TitleBar() // zero while the view draws the buttons, as here
	ui.Row(c).Height(titleBarHeight).Padding(0, bar.Right, 0, bar.Left+style.PadPage).Gap(style.GapGroup).DragWindow().Children(func() {
		ui.Image(c, style.Logo).Size(style.TitleLogo, style.TitleLogo)
		style.RoleBodyStrong.On(t, ui.Text(c, i18n.T("title"))).NoWrap().Grow(1).Shrink(1)
		if a.relLoading {
			ui.Spinner(c).Label(i18n.T("latest.loading"))
		}
		if v := mygo.App.Version(); v != "" {
			a.versionButton(c, t, "v"+v)
		}
		ui.Row(c).AlignSelf(ui.Stretch).Margin(0, 0, 0, style.Space1).Children(func() {
			a.windowButton(c, t, style.IconGear, style.TitleIcon, "settings.open", false, a.toggleSettings)
			if a.win == nil { // tests: no window to drive
				return
			}
			a.windowButton(c, t, iconMin, style.CaptionIcon, "win.minimize", false, a.win.Minimize)
			if a.win.IsMaximized() {
				a.windowButton(c, t, iconRestore, style.CaptionIcon, "win.restore", false, a.win.Unmaximize).MaximizeButton()
			} else {
				a.windowButton(c, t, iconMax, style.CaptionIcon, "win.maximize", false, a.win.Maximize).MaximizeButton()
			}
			a.windowButton(c, t, iconClose, style.CaptionIcon, "win.close", true, a.win.Close)
		})
	})
}

// windowButton is a caption button as Windows 11 draws them: flat, 46 DIPs wide, the close button red on hover.
// key is the text of its label and tooltip.
func (a *app) windowButton(c *ui.Context, t *ui.Theme, icon *ui.SVG, size float32, key string, danger bool, action func()) ui.Element {
	b := ui.ButtonBase(c.Key(key)).Width(46).AlignSelf(ui.Stretch).Center().Label(i18n.T(key)).Tooltip(i18n.T(key))
	fg := t.Text
	switch {
	case danger && b.Pressed():
		b.Background(ui.Hex("#C42B1C").Alpha(0.85))
		fg = ui.RGB(255, 255, 255)
	case danger && b.Hovered():
		b.Background(ui.Hex("#C42B1C"))
		fg = ui.RGB(255, 255, 255)
	case b.Pressed():
		b.Background(t.Text.Alpha(0.12))
	case b.Hovered():
		b.Background(t.Text.Alpha(0.07))
	}
	b.OnClick(action).Children(func() {
		ui.Icon(c, icon).Size(size, size).TextColor(fg)
	})
	return b
}

// versionButton is the app's version as quiet text in the title bar, which opens the About dialog.
func (a *app) versionButton(c *ui.Context, t *ui.Theme, label string) {
	b := ui.ButtonBase(c.Key("app-version")).Padding(style.Space1, style.Space2).Radius(style.NoticeRadius).Label(i18n.T("about.open")).Tooltip(i18n.T("about.open"))
	if b.Hovered() {
		b.Background(t.Text.Alpha(0.07))
	}
	b.OnClick(a.openAbout).Children(func() {
		style.RoleSmall.On(t, ui.Text(c, label)).TextColor(t.TextMuted)
	})
}

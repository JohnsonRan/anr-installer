package main

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

//go:embed resources/logo.svg
var logoSVG []byte

var (
	logo        = mustSVG(logoSVG)
	iconMin     = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 10"><rect x="0" y="4.5" width="10" height="1" fill="#000"/></svg>`))
	iconMax     = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 10"><rect x="0.5" y="0.5" width="9" height="9" fill="none" stroke="#000"/></svg>`))
	iconRestore = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 10"><path d="M2.5 2.5V0.5h7v7h-2" fill="none" stroke="#000"/><rect x="0.5" y="2.5" width="7" height="7" fill="none" stroke="#000"/></svg>`))
	iconClose   = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 10"><path d="M0.5 0.5l9 9M9.5 0.5l-9 9" stroke="#000" stroke-width="1.1"/></svg>`))
)

func mustSVG(data []byte) *ui.SVG {
	s, err := ui.ParseSVG(data)
	if err != nil {
		panic(err)
	}
	return s
}

const titleBarHeight = 52

// titleBar draws the window's title bar in place of the system's: the logo, the title, the latest release, and
// the window buttons. Its empty parts drag the window, and a double click maximizes it.
func (a *app) titleBar(c *ui.Context, t *ui.Theme) {
	bar := c.TitleBar() // zero while the view draws the buttons, as here
	ui.Row(c).Height(titleBarHeight).Padding(0, bar.Right, 0, bar.Left+16).Gap(12).DragWindow().Children(func() {
		ui.Image(c, logo).Size(28, 28)
		ui.Row(c).Grow(1).Shrink(1).Gap(10).AlignItems(ui.Baseline).Children(func() {
			roleBodyStrong.on(t, ui.Text(c, T("title"))).NoWrap()
			roleSmall.on(t, ui.Text(c, T("subtitle"))).TextColor(t.TextMuted).MaxLines(1).Ellipsis("…").Shrink(1)
		})
		switch {
		case a.relLoading:
			ui.Spinner(c).Label(T("latest.loading"))
		case a.relErr != "":
			roleSmall.on(t, ui.Text(c, T("latest.error", a.relErr))).TextColor(t.Danger).MaxLines(1).Ellipsis("…").MaxWidth(320).Tooltip(a.relErr)
			ui.Button(c, T("refresh")).OnClick(a.fetchRelease)
		case a.rel != nil:
			ui.Badge(c, T("latest", a.rel.Version)).Background(t.Accent).TextColor(t.AccentText)
		}
		if a.win == nil { // tests: no window to drive
			return
		}
		ui.Row(c).AlignSelf(ui.Stretch).Margin(0, 0, 0, 4).Children(func() {
			a.windowButton(c, t, iconMin, "minimize", false, a.win.Minimize)
			if a.win.IsMaximized() {
				a.windowButton(c, t, iconRestore, "restore", false, a.win.Unmaximize).MaximizeButton()
			} else {
				a.windowButton(c, t, iconMax, "maximize", false, a.win.Maximize).MaximizeButton()
			}
			a.windowButton(c, t, iconClose, "close", true, a.win.Close)
		})
	})
}

// windowButton is a caption button as Windows 11 draws them: flat, 46 DIPs wide, the close button red on hover.
func (a *app) windowButton(c *ui.Context, t *ui.Theme, icon *ui.SVG, key string, danger bool, action func()) ui.Element {
	b := ui.ButtonBase(c.Key(key)).Width(46).AlignSelf(ui.Stretch).Center().Label(T("win." + key)).Tooltip(T("win." + key))
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
		ui.Icon(c, icon).Size(10, 10).TextColor(fg)
	})
	return b
}

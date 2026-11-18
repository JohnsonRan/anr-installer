package main

import (
	_ "embed"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

//go:embed assets/logo.svg
var logoSVG []byte

// iconPNG is the app icon, built into the executable for the update window: nothing ships beside the executable.
//
//go:embed assets/icon.png
var iconPNG []byte

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

// titleBar draws the window's title bar in place of the system's: the logo, the title, the app's version, which
// opens the About dialog, the settings button and the window buttons. Its empty parts drag the window, and a
// double click maximizes it. The ANR FG releases and their errors are on the game page, by the install button.
func (a *app) titleBar(c *ui.Context, t *ui.Theme) {
	bar := c.TitleBar() // zero while the view draws the buttons, as here
	ui.Row(c).Height(titleBarHeight).Padding(0, bar.Right, 0, bar.Left+padPage).Gap(gapGroup).DragWindow().Children(func() {
		ui.Image(c, logo).Size(titleLogo, titleLogo)
		roleBodyStrong.on(t, ui.Text(c, T("title"))).NoWrap().Grow(1).Shrink(1)
		if a.relLoading {
			ui.Spinner(c).Label(T("latest.loading"))
		}
		if v := mygo.App.Version(); v != "" {
			a.versionButton(c, t, "v"+v)
		}
		ui.Row(c).AlignSelf(ui.Stretch).Margin(0, 0, 0, space1).Children(func() {
			a.windowButton(c, t, iconGear, titleIcon, "settings.open", false, a.toggleSettings)
			if a.win == nil { // tests: no window to drive
				return
			}
			a.windowButton(c, t, iconMin, captionIcon, "win.minimize", false, a.win.Minimize)
			if a.win.IsMaximized() {
				a.windowButton(c, t, iconRestore, captionIcon, "win.restore", false, a.win.Unmaximize).MaximizeButton()
			} else {
				a.windowButton(c, t, iconMax, captionIcon, "win.maximize", false, a.win.Maximize).MaximizeButton()
			}
			a.windowButton(c, t, iconClose, captionIcon, "win.close", true, a.win.Close)
		})
	})
}

// windowButton is a caption button as Windows 11 draws them: flat, 46 DIPs wide, the close button red on hover.
// key is the text of its label and tooltip.
func (a *app) windowButton(c *ui.Context, t *ui.Theme, icon *ui.SVG, size float32, key string, danger bool, action func()) ui.Element {
	b := ui.ButtonBase(c.Key(key)).Width(46).AlignSelf(ui.Stretch).Center().Label(T(key)).Tooltip(T(key))
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
	b := ui.ButtonBase(c.Key("app-version")).Padding(space1, space2).Radius(noticeRadius).Label(T("about.open")).Tooltip(T("about.open"))
	if b.Hovered() {
		b.Background(t.Text.Alpha(0.07))
	}
	b.OnClick(a.openAbout).Children(func() {
		roleSmall.on(t, ui.Text(c, label)).TextColor(t.TextMuted)
	})
}

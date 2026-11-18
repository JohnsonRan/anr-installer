// Package gui is the installer's window: the list of games, a game's page with its install and recording
// controls, the settings, the dialogs and the notification area icon. Run opens it.
package gui

import (
	"github.com/egoist/mygo"
	// MyGo's own controls and dialogs in the languages the window speaks besides English.
	_ "github.com/egoist/mygo/locales/ja"
	_ "github.com/egoist/mygo/locales/zh-Hans"
	_ "github.com/egoist/mygo/locales/zh-Hant"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/plugins/updater/native"
	"github.com/egoist/mygo/ui"

	"anr-installer/assets"
	"anr-installer/gui/i18n"
	"anr-installer/internal/paths"
)

// The window's sizes in DIPs: the minimum, and the size its pages are laid out for, which it opens at the first
// time when the display has room; later launches restore the size the user left it at.
const (
	minWidth, minHeight       = 760, 520
	preferWidth, preferHeight = 1280, 900
	workAreaShare             = 0.9 // of the display's work area the first window may take
)

// defaultSize is the window's size at its first launch on a display whose work area is work: the size the pages
// are laid out for, made smaller to leave room around it on the display, never below the minimum. MyGo centres a
// new window without fitting it to the display, which put the title bar out of reach on a small one.
func defaultSize(work mygo.Rectangle) (width, height int) {
	fit := func(prefer, room, least int) int {
		if room <= 0 { // no display reported
			return prefer
		}
		return max(min(prefer, int(float64(room)*workAreaShare)), least)
	}
	return fit(preferWidth, work.Width, minWidth), fit(preferHeight, work.Height, minHeight)
}

// Run opens the window and runs the app until it quits. A second launch hands over to the running one and
// returns at once.
func Run() error {
	// MyGo keeps the language chosen, the window's place and the updater's state in the folder of the app's own
	// settings: one folder for everything, and the portable version's beside its executable.
	mygo.App.SetPath(mygo.PathUserData, paths.Config())
	// A second launch brings the running window forward (OnSecondInstance) and exits.
	if !mygo.App.RequestSingleInstanceLock() {
		return nil
	}
	a := newApp()
	// Updates are offered in a window MyGo draws itself, which needs no WebView2: checked 10 seconds after launch,
	// then daily, and from the version button of the title bar. It speaks the language the window does, which
	// MyGo resolves among these and keeps when the settings choose one.
	mygo.Use(native.New(updater.Options{Icon: assets.IconPNG}))
	mygo.App.SetLanguages(i18n.Languages...)
	mygo.App.WhenReady(func() {
		i18n.SetLang(i18n.LangOf(mygo.App.Locale()))
		// Started as the user signs in, the window stays out of sight: hidden behind the tray icon when there is
		// one (below), else minimized.
		atLogin := mygo.App.WasOpenedAtLogin()
		hidden := atLogin && a.saved.CloseToTray
		width, height := defaultSize(mygo.Screen.PrimaryDisplay().WorkArea)
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:             i18n.T("title"),
			Width:             width,
			Height:            height,
			MinWidth:          minWidth,
			MinHeight:         minHeight,
			StateKey:          "main",
			TitleBarStyle:     mygo.TitleBarHidden,
			HideWindowButtons: true,
			Hidden:            hidden,
			Content:           ui.View(a.view),
		})
		a.update = a.win.Update
		mygo.App.OnLocaleChanged(func(locale string) {
			i18n.SetLang(i18n.LangOf(locale))
			a.win.SetTitle(i18n.T("title"))
			a.relabelTray()
			a.relabelVersions()
			a.win.Invalidate()
		})
		mygo.App.OnSecondInstance(func([]string, string) { a.showWindow() })
		// A quit from the tray or an update's restart waits for the user's answer while a task runs; the answer
		// "Quit anyway" lets the one quit after it through. That quit does not relaunch, as MyGo drops a relaunch
		// once OnBeforeQuit prevents it: the update is installed already and applies at the next launch.
		mygo.App.OnBeforeQuit(func(e *mygo.QuitEvent) {
			forced := a.quitForced
			a.quitForced = false
			if quitAsks(a.busy, forced, e.SessionEnding) {
				e.PreventDefault()
				a.askQuit()
				return
			}
			a.quitting = true
		})
		mygo.App.OnQuitCanceled(func() { a.quitting = false })
		mygo.App.OnWillQuit(func(*mygo.QuitEvent) { a.removeTray() })
		a.win.OnClose(func(e *mygo.CloseEvent) {
			switch {
			case closeHides(a.saved.CloseToTray, a.tray != nil, a.quitting):
				e.PreventDefault()
				a.win.Hide()
			case quitAsks(a.busy, false, a.quitting): // closing the only window quits
				e.PreventDefault()
				a.askQuit()
			}
		})
		a.syncTray()
		if atLogin && a.tray == nil {
			// No tray icon, by the setting or a failure, would bring a hidden window back: it shows minimized. A
			// window made hidden for the icon that failed shows first: MyGo takes a window out of its cloak as it
			// shows, and Minimize is a bare SW_MINIMIZE.
			if hidden {
				a.win.Show()
			}
			a.win.Minimize()
		}
		a.start()
		// A recording is made in the game, while this window is behind it: look for new ones on the way back.
		a.win.OnFocus(a.refreshRecord)
		// The About dialog and the settings page show the updater's state: draw them again when a check or a
		// preference changes it.
		updater.OnChange(func() {
			a.autoCheck = updater.AutomaticChecks()
			a.win.Invalidate()
		})
	})
	return mygo.App.Run()
}

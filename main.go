// anr-installer installs, upgrades and removes ANR FG (Apex Neural Rendering frame generation) in Steam games:
// it finds the games, tells which ship DLSS frame generation, picks the full or add-on package, downloads and
// checks it, and runs the package's own installer.
package main

import (
	"log"

	"github.com/egoist/mygo"
	_ "github.com/egoist/mygo/locales/zh-Hans" // MyGo's own controls and dialogs in Simplified Chinese
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/plugins/updater/native"
	"github.com/egoist/mygo/ui"
)

func main() {
	// A second launch brings the running window forward (OnSecondInstance) and exits.
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}
	a := newApp()
	// Updates are offered in a window MyGo draws itself, which needs no WebView2: checked 10 seconds after launch,
	// then daily, and from the version button of the title bar. It speaks the language the window does, which
	// MyGo resolves among these and keeps when the settings choose one.
	mygo.Use(native.New(updater.Options{Icon: iconPNG}))
	mygo.App.SetLanguages("en", "zh-Hans")
	mygo.App.WhenReady(func() {
		lang = langOf(mygo.App.Locale())
		// Started as the user signs in, the window stays out of sight: hidden behind the tray icon when there is
		// one (below), else minimized.
		atLogin := mygo.App.WasOpenedAtLogin()
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:             T("title"),
			Width:             1040,
			Height:            760,
			MinWidth:          760,
			MinHeight:         520,
			StateKey:          "main",
			TitleBarStyle:     mygo.TitleBarHidden,
			HideWindowButtons: true,
			Hidden:            atLogin,
			Content:           ui.View(a.view),
		})
		mygo.App.OnLocaleChanged(func(locale string) {
			lang = langOf(locale)
			a.win.SetTitle(T("title"))
			a.relabelTray()
			a.win.Invalidate()
		})
		mygo.App.OnSecondInstance(func([]string, string) { a.showWindow() })
		// A quit from the tray or an update's restart waits for the user's answer while a task runs; the answer
		// "Quit anyway" lets the one quit after it through.
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
			// No tray icon, by the setting or a failure, would bring a hidden window back: it shows minimized.
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
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

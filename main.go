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
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:             T("title"),
			Width:             1040,
			Height:            760,
			MinWidth:          760,
			MinHeight:         520,
			StateKey:          "main",
			TitleBarStyle:     mygo.TitleBarHidden,
			HideWindowButtons: true,
			Content:           ui.View(a.view),
		})
		mygo.App.OnLocaleChanged(func(locale string) {
			lang = langOf(locale)
			a.win.SetTitle(T("title"))
			a.relabelTray()
			a.win.Invalidate()
		})
		mygo.App.OnSecondInstance(func([]string, string) { a.showWindow() })
		mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { a.quitting = true })
		mygo.App.OnQuitCanceled(func() { a.quitting = false })
		mygo.App.OnWillQuit(func(*mygo.QuitEvent) { a.removeTray() })
		a.win.OnClose(func(e *mygo.CloseEvent) {
			if closeHides(a.saved.CloseToTray, a.tray != nil, a.quitting) {
				e.PreventDefault()
				a.win.Hide()
			}
		})
		a.syncTray()
		a.start()
		// A recording is made in the game, while this window is behind it: look for new ones on the way back.
		a.win.OnFocus(a.refreshRecord)
		// The About dialog shows the updater's state: draw it again when a check or a preference changes it.
		updater.OnChange(func() {
			a.autoCheck = updater.AutomaticChecks()
			a.win.Invalidate()
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

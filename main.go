// anr-installer installs, upgrades and removes ANR FG (Apex Neural Rendering frame generation) in Steam games:
// it finds the games, tells which ship DLSS frame generation, picks the full or add-on package, downloads and
// checks it, and runs the package's own installer.
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/plugins/updater/native"
	"github.com/egoist/mygo/ui"
)

func main() {
	a := newApp()
	// Updates are offered in a window MyGo draws itself, which needs no WebView2: checked 10 seconds after launch,
	// then daily, and from the version button of the title bar. The window speaks the system's language.
	mygo.Use(native.Plugin)
	mygo.App.WhenReady(func() {
		setLanguage(mygo.App.PreferredLanguages())
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
		a.start()
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

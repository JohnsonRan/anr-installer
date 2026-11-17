// anr-installer installs, upgrades and removes ANR FG (Apex Neural Rendering frame generation) in Steam games:
// it finds the games, tells which ship DLSS frame generation, picks the full or add-on package, downloads and
// checks it, and runs the package's own installer.
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func main() {
	a := newApp()
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
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

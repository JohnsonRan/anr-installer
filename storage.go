package main

import (
	"fmt"
	"os"

	"github.com/egoist/mygo"

	"anr-installer/internal/paths"
	"anr-installer/internal/pkg"
)

// sizeText says a size in bytes in MB, or in GB from 1 GB. A no-break space joins the number to its unit, so
// that a line of text that wraps never parts them.
func sizeText(n int64) string {
	if n >= 1e9 {
		return fmt.Sprintf("%.1f GB", float64(n)/1e9)
	}
	return fmt.Sprintf("%.1f MB", float64(n)/1e6)
}

// openDataFolder opens the folder of games.json, which holds nothing until a setting is first saved.
func (a *app) openDataFolder() {
	dir := paths.Config()
	err := os.MkdirAll(dir, 0o755)
	if err == nil {
		err = mygo.Shell.OpenPath(dir)
	}
	if err != nil {
		a.note(levelError, T("failed", err.Error()))
	}
}

// measureCache counts the downloaded packages off the main thread for the settings page s.
func (a *app) measureCache(s *settingsState) {
	s.cacheSize = -1
	if a.win == nil { // tests: no MyGo app runs
		return
	}
	go func() {
		n := pkg.DownloadsSize()
		a.win.Update(func() { s.cacheSize = n })
	}()
}

// deleteDownloads deletes the downloaded packages off the main thread, then counts them again. It is the task
// in progress meanwhile, so that no install downloads into the folders it deletes. A package downloads again
// when it is needed, so nothing asks first.
func (a *app) deleteDownloads() {
	if a.busy {
		return
	}
	a.busy, a.progress = true, -1
	s := a.settings
	tk := a.begin(T("settings.cache.delete"))
	go func() {
		freed, err := pkg.RemoveDownloads()
		a.update(func() {
			a.busy, a.progress = false, -2
			fail := ""
			if err != nil {
				fail = T("settings.cache.failed", err.Error())
			} else {
				tk.step(levelInfo, T("settings.cache.deleted", sizeText(freed)))
			}
			a.end(tk, fail)
			a.measureCache(s)
		})
	}()
}

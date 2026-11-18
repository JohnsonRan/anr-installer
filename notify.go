package main

import (
	"context"
	"time"

	"github.com/egoist/mygo"

	"anr-installer/internal/pkg"
)

// releaseCheckEvery is how often the app looks again for a new ANR FG release while it runs.
const releaseCheckEvery = 6 * time.Hour

// watchReleases looks for a new ANR FG release every releaseCheckEvery, quietly: a check that fails keeps the
// releases loaded, sets no error and shows no spinner, and adds a warning to the activity log alone.
func (a *app) watchReleases() {
	go func() {
		for range time.Tick(releaseCheckEvery) {
			rels, err := pkg.Releases(context.Background())
			a.win.Update(func() {
				if err != nil {
					a.note(levelWarning, T("notify.check.failed", err.Error()))
					return
				}
				a.releasesFound(rels)
			})
		}
	}()
}

// releasesFound keeps the releases a check found, reads the versions of their packages, and tells of a new one.
// It clears the error of a check that failed before.
func (a *app) releasesFound(rels []*pkg.Release) {
	a.relErr = ""
	a.setReleases(rels)
	a.probe(rels)
	a.notifyRelease()
}

// shouldNotify reports whether a check that found newest as the newest release shows a notification: with the
// setting on, for a release not notified before, and only while the window is out of sight, as in the window
// autoNotes shows the release.
func shouldNotify(on bool, newest, notified string, windowShown bool) bool {
	return on && newest != "" && newest != notified && !windowShown
}

// notifyRelease tells of a new ANR FG release with a system notification while the window is hidden or
// minimized. The release is recorded as it is first found, notified of or not, so that it notifies once. It
// waits for the games to be scanned, which the notification counts: scanning calls it again.
func (a *app) notifyRelease() {
	rel := a.rel
	if rel == nil || rel.Tag == a.saved.NotifiedRelease || a.scanning() {
		return
	}
	shown := a.win == nil || a.win.IsVisible() && !a.win.IsMinimized() // tests: no window
	notify := shouldNotify(!a.saved.ReleaseNotifyOff, rel.Tag, a.saved.NotifiedRelease, shown)
	a.saved.NotifiedRelease = rel.Tag
	a.saveSettings()
	if notify {
		a.showReleaseNotice(rel)
	}
}

// scanning reports whether the games are still being found or read.
func (a *app) scanning() bool {
	if a.loadingGames {
		return true
	}
	for _, g := range a.games {
		if g.Scanning {
			return true
		}
	}
	return false
}

// showReleaseNotice shows the notification of rel. A click brings the window forward, its list showing only the
// games to upgrade when there are any.
func (a *app) showReleaseNotice(rel *pkg.Release) {
	body := T("notify.open")
	if n := a.outdatedCount(); n > 0 {
		body = T("notify.upgrade", n)
	}
	note := mygo.NewNotification(mygo.NotificationOptions{ID: "anr-release", Title: T("notify.title", rel.Version), Body: body})
	note.OnClick(func() {
		if a.outdatedCount() > 0 {
			a.onlyOutdated, a.row = true, -1
		}
		a.showWindow()
		a.win.Invalidate()
	})
	// Show waits for the system, which on some platforms asks the user first: off the main thread.
	go func() {
		if err := note.Show(); err != nil {
			a.win.Update(func() { a.note(levelWarning, T("notify.failed", err.Error())) })
		}
	}()
}

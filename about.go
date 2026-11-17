package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/ui"
)

const repoURL = "https://github.com/JohnsonRan/anr-installer"

// aboutLinks are the pages the About dialog opens, in the order it shows them.
var aboutLinks = []struct{ key, url string }{
	{"about.github", repoURL},
	{"about.changelog", repoURL + "/blob/main/CHANGELOG.md"},
	{"about.issues", repoURL + "/issues"},
	{"about.license", repoURL + "/blob/main/LICENSE"},
	{"about.anrfg", "https://github.com/apex-compute/ANR"},
}

// openAbout shows the About dialog, with the update preference as the updater keeps it.
func (a *app) openAbout() {
	a.autoCheck = updater.AutomaticChecks()
	a.aboutOpen = true
}

// portable reports whether this copy runs from a folder of its own rather than the per-user install, which
// the installer puts in %LOCALAPPDATA%\Programs.
func portable() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	programs := filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs")
	return os.Getenv("LOCALAPPDATA") == "" || !strings.HasPrefix(strings.ToLower(exe), strings.ToLower(programs)+string(filepath.Separator))
}

// updateStatus is the line under "Software update": why this copy cannot update, the version found, or when
// the updater last checked.
func updateStatus() string {
	switch {
	case !mygo.Updater.Enabled():
		return T("about.cantUpdate")
	case updater.AvailableVersion() != "":
		return T("about.available", updater.AvailableVersion())
	case !updater.LastCheck().IsZero():
		return T("about.lastCheck", updater.LastCheck().Local().Format("2006-01-02 15:04"))
	}
	return T("about.neverChecked")
}

// about is the About dialog: what the app is, its version, its updates and the project's links.
func (a *app) about(c *ui.Context, t *ui.Theme) {
	ui.Modal(c, &a.aboutOpen, func() {
		ui.Column(c).Width(460).Gap(gapSection).Children(func() {
			ui.Column(c).AlignItems(ui.Center).Gap(gapLines).Children(func() {
				ui.Image(c, logo).Size(64, 64)
				roleHeading.on(t, ui.Text(c, T("title")))
				ui.Row(c).Gap(gapControls).Children(func() {
					version := mygo.App.Version()
					if version == "" {
						version = "dev"
					}
					roleSmall.on(t, ui.Text(c, T("about.version", version))).TextColor(t.TextMuted)
					edition := T("about.installed")
					if portable() {
						edition = T("about.portable")
					}
					ui.Badge(c, edition)
				})
				roleSmall.on(t, ui.Text(c, T("subtitle"))).TextColor(t.TextMuted).TextAlign(ui.Center)
			})

			card(c, t, T("about.updates"), func() {
				ui.Row(c).Gap(gapGroup).Children(func() {
					roleSmall.on(t, ui.Text(c, updateStatus())).TextColor(t.TextMuted).Grow(1).Shrink(1)
					ui.Button(c, T("about.checkNow")).Disabled(!mygo.Updater.Enabled()).OnClick(updater.CheckForUpdates)
				})
				ui.Row(c).Gap(gapGroup).Children(func() {
					ui.Column(c).Grow(1).Shrink(1).Gap(gapLines).Children(func() {
						ui.Text(c, T("about.auto"))
						roleSmall.on(t, ui.Text(c, T("about.autoHint"))).TextColor(t.TextMuted)
					})
					if ui.Switch(c, &a.autoCheck).Label(T("about.auto")).Disabled(!mygo.Updater.Enabled()).Changed() {
						updater.SetAutomaticChecks(a.autoCheck)
					}
				})
			})

			ui.Row(c).Gap(gapControls).Wrap().Justify(ui.Center).Children(func() {
				for _, l := range aboutLinks {
					ui.Button(c.Key(l.key), T(l.key)).Tooltip(l.url).OnClick(func() { mygo.Shell.OpenExternal(l.url) })
				}
			})

			ui.Column(c).Gap(gapLines).Children(func() {
				roleSmall.on(t, ui.Text(c, T("about.notice"))).TextColor(t.TextMuted).TextAlign(ui.Center)
				roleSmall.on(t, ui.Text(c, T("about.credits"))).TextColor(t.TextMuted).TextAlign(ui.Center)
			})

			ui.Row(c).Justify(ui.End).Children(func() {
				ui.Button(c, T("about.close")).OnClick(func() { a.aboutOpen = false })
			})
		})
	})
}

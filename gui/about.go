package gui

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
	"anr-installer/internal/paths"
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

func (a *app) openAbout() { a.aboutOpen = true }

// portable reports whether this copy is the portable version, which keeps its files in its own folder.
func portable() bool { return paths.Portable() != "" }

// updateStatus is the line under "Software update": why this copy cannot update, the version found, or when
// the updater last checked.
func updateStatus() string {
	switch {
	case !mygo.Updater.Enabled():
		return i18n.T("about.cantUpdate")
	case updater.AvailableVersion() != "":
		return i18n.T("about.available", updater.AvailableVersion())
	case !updater.LastCheck().IsZero():
		return i18n.T("about.lastCheck", updater.LastCheck().Local().Format("2006-01-02 15:04"))
	}
	return i18n.T("about.neverChecked")
}

// about is the About dialog: what the app is, its version, its updates and the project's links.
func (a *app) about(c *ui.Context, t *ui.Theme) {
	ui.Modal(c, &a.aboutOpen, func() {
		ui.Column(c).Width(460).Gap(style.GapSection).Children(func() {
			ui.Column(c).AlignItems(ui.Center).Gap(style.GapLines).Children(func() {
				ui.Image(c, style.Logo).Size(64, 64)
				style.RoleHeading.On(t, ui.Text(c, i18n.T("title")))
				ui.Row(c).Gap(style.GapControls).Children(func() {
					version := mygo.App.Version()
					if version == "" {
						version = "dev"
					}
					style.RoleSmall.On(t, ui.Text(c, i18n.T("about.version", version))).TextColor(t.TextMuted)
					edition := i18n.T("about.installed")
					if portable() {
						edition = i18n.T("about.portable")
					}
					ui.Badge(c, edition)
				})
				style.RoleSmall.On(t, ui.Text(c, i18n.T("subtitle"))).TextColor(t.TextMuted).TextAlign(ui.Center)
			})

			card(c, t, i18n.T("about.updates"), func() {
				ui.Row(c).Gap(style.GapGroup).Children(func() {
					style.RoleSmall.On(t, ui.Text(c, updateStatus())).TextColor(t.TextMuted).Grow(1).Shrink(1)
					ui.Button(c, i18n.T("about.checkNow")).Disabled(!mygo.Updater.Enabled()).OnClick(updater.CheckForUpdates)
				})
			})

			ui.Row(c).Gap(style.GapControls).Wrap().Justify(ui.Center).Children(func() {
				for _, l := range aboutLinks {
					ui.Button(c.Key(l.key), i18n.T(l.key)).Tooltip(l.url).OnClick(func() { mygo.Shell.OpenExternal(l.url) })
				}
			})

			ui.Column(c).Gap(style.GapLines).Children(func() {
				style.RoleSmall.On(t, ui.Text(c, i18n.T("about.notice"))).TextColor(t.TextMuted).TextAlign(ui.Center)
				style.RoleSmall.On(t, ui.Text(c, i18n.T("about.credits"))).TextColor(t.TextMuted).TextAlign(ui.Center)
			})

			ui.Row(c).Justify(ui.End).Children(func() {
				ui.Button(c, i18n.T("about.close")).OnClick(func() { a.aboutOpen = false })
			})
		})
	})
}

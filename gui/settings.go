package gui

import (
	"slices"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/ui"

	"anr-installer/gui/activity"
	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
	"anr-installer/internal/paths"
)

// localeTags are the languages the language setting offers, as MyGo's locale tags: "" follows the system.
var localeTags = append([]string{""}, i18n.Languages...)

// settingsState is the settings page, shown in place of a game's page or a comparison, which stays beneath it.
type settingsState struct {
	game   *game  // the game chosen when the page opened: choosing another closes the page
	locale string // the language chosen, as mygo.App.ChosenLocale reported it when the page opened
	// The switches' states, which they write between frames: whether the app starts at sign-in, as
	// mygo.App.OpenAtLogin reported it when the page opened, and whether a new release notifies.
	atLogin, notify bool
	cacheSize       int64 // bytes of the downloaded packages, -1 while they are counted
}

// toggleSettings opens the settings page, or closes it.
func (a *app) toggleSettings() {
	if a.settings != nil {
		a.closeSettings()
		return
	}
	s := &settingsState{game: a.selected(), notify: !a.saved.ReleaseNotifyOff}
	if a.win != nil { // tests: no MyGo app runs
		s.locale = mygo.App.ChosenLocale()
		s.atLogin = mygo.App.OpenAtLogin()
	}
	a.autoCheck = updater.AutomaticChecks()
	a.measureCache(s)
	a.settings = s
}

func (a *app) closeSettings() { a.settings = nil }

// localeIndex returns the place in localeTags of the language chosen: a tag of another form, as MYGO_LOCALE may
// set, by the window's language it gives.
func localeIndex(tag string) int {
	if i := slices.Index(localeTags, tag); i >= 0 {
		return i
	}
	return slices.Index(localeTags, i18n.LangOf(tag))
}

// chooseLocale shows the language of tag and keeps it for the next launches; OnLocaleChanged redraws the window.
func (a *app) chooseLocale(tag string) {
	a.settings.locale = tag
	if a.win != nil {
		mygo.App.SetLocale(tag)
	}
}

// saveSettings keeps the settings in a.saved. Unlike persist, it leaves the games kept as they are: before the
// Steam libraries are read, persist would forget the .exe folders chosen for Steam games.
func (a *app) saveSettings() {
	if err := a.saved.Save(); err != nil {
		a.note(activity.LevelError, i18n.T("save.failed", err.Error()))
	}
}

// setOpenAtLogin makes the app start when the user signs in, or stops it from doing so, as the switch of s says.
// On a failure the switch goes back to what the system has.
func (a *app) setOpenAtLogin(s *settingsState) {
	if a.win == nil { // tests: the user's sign-in entries stay as they are
		return
	}
	if err := mygo.App.SetOpenAtLogin(s.atLogin); err != nil {
		a.note(activity.LevelError, i18n.T("settings.login.failed", err.Error()))
		s.atLogin = mygo.App.OpenAtLogin()
	}
}

// setCloseToTray keeps what the close button does, and adds or removes the tray icon to match.
func (a *app) setCloseToTray(on bool) {
	a.saved.CloseToTray = on
	a.saveSettings()
	a.syncTray()
}

func (a *app) settingsPage(c *ui.Context, t *ui.Theme) {
	s := a.settings
	c.OnWindowShortcut(0, ui.KeyEscape, a.closeSettings)
	ui.Column(c).Padding(style.PadPage, style.PadPageX).Gap(style.GapSection).Children(func() {
		style.PageHeader(c, t, a.closeSettings, i18n.T("settings.open"), "", nil)
		style.Section(c, t, i18n.T("settings.group.general"), nil, func() {
			style.SettingsRow(c, t, style.IconGlobe, i18n.T("settings.language"), i18n.T("settings.language.detail"), func() {
				labels := []string{i18n.T("settings.language.system")}
				for _, tag := range i18n.Languages {
					labels = append(labels, i18n.T("settings.language."+tag))
				}
				sel := labels[localeIndex(s.locale)]
				if ui.Select(c.Key("language"), &sel, labels).Label(i18n.T("settings.language")).Changed() {
					a.chooseLocale(localeTags[slices.Index(labels, sel)])
				}
			})
			style.SettingsRow(c, t, style.IconWindow, i18n.T("settings.close"), i18n.T("settings.close.detail"), func() {
				quit, tray := i18n.T("settings.close.quit"), i18n.T("settings.close.tray")
				sel := quit
				if a.saved.CloseToTray {
					sel = tray
				}
				if ui.Select(c.Key("close"), &sel, []string{quit, tray}).Label(i18n.T("settings.close")).Changed() {
					a.setCloseToTray(sel == tray)
				}
			})
			style.SettingsRow(c, t, style.IconPower, i18n.T("settings.login"), i18n.T("settings.login.detail"), func() {
				if ui.Switch(c, &s.atLogin).Label(i18n.T("settings.login")).Changed() {
					a.setOpenAtLogin(s)
				}
			})
		})
		style.Section(c, t, i18n.T("settings.group.updates"), nil, func() {
			style.SettingsRow(c, t, style.IconUpdate, i18n.T("settings.updates"), i18n.T("settings.updates.detail"), func() {
				if ui.Switch(c, &a.autoCheck).Label(i18n.T("settings.updates")).Disabled(!mygo.Updater.Enabled()).Changed() {
					updater.SetAutomaticChecks(a.autoCheck)
				}
			})
			style.SettingsRow(c, t, style.IconBell, i18n.T("settings.notify"), i18n.T("settings.notify.detail"), func() {
				if ui.Switch(c, &s.notify).Label(i18n.T("settings.notify")).Changed() {
					a.saved.ReleaseNotifyOff = !s.notify
					a.saveSettings()
				}
			})
		})
		style.Section(c, t, i18n.T("settings.group.storage"), nil, func() {
			style.ValueRow(c, t, style.IconFolder, i18n.T("settings.folder"), paths.Config(), func() {
				ui.Button(c, i18n.T("open.folder")).OnClick(a.openDataFolder)
			})
			size := i18n.T("settings.cache.measuring")
			if s.cacheSize >= 0 {
				size = i18n.T("settings.cache.detail", sizeText(s.cacheSize))
			}
			style.SettingsRow(c, t, style.IconDownload, i18n.T("settings.cache"), size, func() {
				ui.Button(c, i18n.T("settings.cache.delete")).Disabled(a.busy || s.cacheSize <= 0).OnClick(a.deleteDownloads)
			})
		})
	})
}

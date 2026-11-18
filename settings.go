package main

import (
	"slices"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/ui"

	"anr-installer/internal/paths"
)

// localeTags are the languages the language setting offers, as MyGo's locale tags: "" follows the system.
var localeTags = append([]string{""}, appLanguages...)

// settingsState is the settings page, shown in place of a game's page or a comparison, which stays beneath it.
type settingsState struct {
	game   *game  // the game chosen when the page opened: choosing another closes the page
	locale string // the language chosen, as mygo.App.ChosenLocale reported it when the page opened
	// The selects' choices, as their labels, kept across frames: a select writes its choice after the frame.
	langSel, closeSel string
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
	return slices.Index(localeTags, langOf(tag))
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
		a.note(levelError, T("save.failed", err.Error()))
	}
}

// setOpenAtLogin makes the app start when the user signs in, or stops it from doing so, as the switch of s says.
// On a failure the switch goes back to what the system has.
func (a *app) setOpenAtLogin(s *settingsState) {
	if a.win == nil { // tests: the user's sign-in entries stay as they are
		return
	}
	if err := mygo.App.SetOpenAtLogin(s.atLogin); err != nil {
		a.note(levelError, T("settings.login.failed", err.Error()))
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
	ui.Column(c).Padding(padPage, padPageX).Gap(gapSection).Children(func() {
		pageHeader(c, t, a.closeSettings, T("settings.open"), "", nil)
		section(c, t, T("settings.group.general"), nil, func() {
			settingsRow(c, t, iconGlobe, T("settings.language"), T("settings.language.detail"), func() {
				labels := []string{T("settings.language.system")}
				for _, tag := range appLanguages {
					labels = append(labels, T("settings.language."+tag))
				}
				choose(c.Key("language"), &s.langSel, labels, localeIndex(s.locale), T("settings.language"), func(i int) {
					a.chooseLocale(localeTags[i])
				})
			})
			settingsRow(c, t, iconWindow, T("settings.close"), T("settings.close.detail"), func() {
				labels := []string{T("settings.close.quit"), T("settings.close.tray")}
				cur := 0
				if a.saved.CloseToTray {
					cur = 1
				}
				choose(c.Key("close"), &s.closeSel, labels, cur, T("settings.close"), func(i int) { a.setCloseToTray(i == 1) })
			})
			settingsRow(c, t, iconPower, T("settings.login"), T("settings.login.detail"), func() {
				if ui.Switch(c, &s.atLogin).Label(T("settings.login")).Changed() {
					a.setOpenAtLogin(s)
				}
			})
		})
		section(c, t, T("settings.group.updates"), nil, func() {
			settingsRow(c, t, iconUpdate, T("settings.updates"), T("settings.updates.detail"), func() {
				if ui.Switch(c, &a.autoCheck).Label(T("settings.updates")).Disabled(!mygo.Updater.Enabled()).Changed() {
					updater.SetAutomaticChecks(a.autoCheck)
				}
			})
			settingsRow(c, t, iconBell, T("settings.notify"), T("settings.notify.detail"), func() {
				if ui.Switch(c, &s.notify).Label(T("settings.notify")).Changed() {
					a.saved.ReleaseNotifyOff = !s.notify
					a.saveSettings()
				}
			})
		})
		section(c, t, T("settings.group.storage"), nil, func() {
			valueRow(c, t, iconFolder, T("settings.folder"), paths.Config(), func() {
				ui.Button(c, T("open.folder")).OnClick(a.openDataFolder)
			})
			size := T("settings.cache.measuring")
			if s.cacheSize >= 0 {
				size = T("settings.cache.detail", sizeText(s.cacheSize))
			}
			settingsRow(c, t, iconDownload, T("settings.cache"), size, func() {
				ui.Button(c, T("settings.cache.delete")).Disabled(a.busy || s.cacheSize <= 0).OnClick(a.deleteDownloads)
			})
		})
	})
}

// choose draws a select of labels showing labels[cur], and calls pick with the index of another label chosen.
// sel holds the choice across frames: the select writes it after the frame that drew it, so pick is called as
// the next frame starts, before the select shows the new choice.
func choose(c *ui.Context, sel *string, labels []string, cur int, label string, pick func(int)) {
	if i := slices.Index(labels, *sel); i >= 0 && i != cur {
		pick(i)
		cur = i
	}
	*sel = labels[cur]
	ui.Select(c, sel, labels).Label(label)
}

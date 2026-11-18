package main

import (
	"slices"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/ui"
)

// localeTags are the languages the language setting offers, as MyGo's locale tags: "" follows the system.
var localeTags = []string{"", "en", "zh-Hans"}

// settingsState is the settings page, shown in place of a game's page or a comparison, which stays beneath it.
type settingsState struct {
	game   *game  // the game chosen when the page opened: choosing another closes the page
	locale string // the language chosen, as mygo.App.ChosenLocale reported it when the page opened
	// The selects' choices, as their labels, kept across frames: a select writes its choice after the frame.
	langSel, closeSel string
}

// toggleSettings opens the settings page, or closes it.
func (a *app) toggleSettings() {
	if a.settings != nil {
		a.closeSettings()
		return
	}
	s := &settingsState{game: a.selected()}
	if a.win != nil { // tests: no MyGo app runs
		s.locale = mygo.App.ChosenLocale()
	}
	a.autoCheck = updater.AutomaticChecks()
	a.settings = s
}

func (a *app) closeSettings() { a.settings = nil }

// localeIndex returns the place in localeTags of the language chosen: a tag of another form, as MYGO_LOCALE may
// set, by the window's language it gives.
func localeIndex(tag string) int {
	if i := slices.Index(localeTags, tag); i >= 0 {
		return i
	}
	if langOf(tag) == "zh" {
		return 2
	}
	return 1
}

// chooseLocale shows the language of tag and keeps it for the next launches; OnLocaleChanged redraws the window.
func (a *app) chooseLocale(tag string) {
	a.settings.locale = tag
	if a.win != nil {
		mygo.App.SetLocale(tag)
	}
}

// setCloseToTray keeps what the close button does, and adds or removes the tray icon to match.
func (a *app) setCloseToTray(on bool) {
	a.saved.CloseToTray = on
	a.persist()
	a.syncTray()
}

func (a *app) settingsPage(c *ui.Context, t *ui.Theme) {
	s := a.settings
	c.OnWindowShortcut(0, ui.KeyEscape, a.closeSettings)
	ui.Column(c).Padding(padPage, padPageX).Gap(gapSection).Children(func() {
		ui.Row(c).Gap(gapGroup).Children(func() {
			ui.Button(c, T("back")).OnClick(a.closeSettings)
			roleHeading.on(t, ui.Text(c, T("settings.open"))).MaxLines(1).Grow(1).Shrink(1)
		})
		ui.Column(c).Gap(gapCardRows).Children(func() {
			settingsRow(c, t, iconGlobe, T("settings.language"), T("settings.language.detail"), func() {
				labels := []string{T("settings.language.system"), T("settings.language.en"), T("settings.language.zh")}
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
			settingsRow(c, t, iconUpdate, T("settings.updates"), T("settings.updates.detail"), func() {
				if ui.Switch(c, &a.autoCheck).Label(T("settings.updates")).Disabled(!mygo.Updater.Enabled()).Changed() {
					updater.SetAutomaticChecks(a.autoCheck)
				}
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

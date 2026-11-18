package gui

import (
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/gui/activity"
	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
	"anr-installer/internal/detect"
)

func (a *app) view(c *ui.Context) {
	// apexcompute.com's look, always dark, as the site is.
	t := style.ApexTheme()
	c.SetTheme(t)
	a.autoNotes()
	a.filter()
	if a.row < 0 && len(a.visible) > 0 {
		a.row = 0 // a game is always shown, the first one at start
	}
	ui.Column(c).Fill().Children(func() {
		a.titleBar(c, t)
		ui.Divider(c)
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Width(300).Padding(style.PadPage).Gap(style.GapControls).Children(func() { a.sidebar(c, t) })
			ui.Divider(c).Vertical()
			ui.Column(c).Grow(1).Shrink(1).Children(func() {
				g := a.selected()
				if a.cmp != nil && a.cmp.game != g {
					a.closeCompare() // another game was chosen
				}
				if s := a.settings; s != nil && s.game != g {
					if s.game == nil {
						s.game = g // the first game found, chosen by itself: not the user's choice
					} else {
						a.closeSettings() // another game was chosen
					}
				}
				if a.settings != nil {
					ui.Scroll(c.Key(a.settings)).Grow(1).Children(func() { a.settingsPage(c, t) })
				} else if a.cmp != nil {
					ui.Scroll(c.Key(a.cmp)).Grow(1).Children(func() { a.comparePage(c, t) })
				} else if g == nil {
					a.empty(c, t)
				} else {
					ui.Scroll(c.Key(g)).Grow(1).Children(func() { a.detail(c, t, g) })
				}
				a.log.View(c, t, &a.logOpen, activity.Work{Busy: a.busy, Status: a.status, Progress: a.progress})
				a.about(c, t)
				a.releaseNotes(c, t)
				a.quitDialog(c, t)
			})
		})
	})
}

// --- the list of games ---

func (a *app) sidebar(c *ui.Context, t *ui.Theme) {
	ui.SearchField(c, &a.query).Label(i18n.T("search")).Placeholder(i18n.T("search"))
	// As wide as its labels: shared equally, the list's width cuts the longest, "Supported" or "インストール済み".
	ui.Segmented(c, &a.show, i18n.T("show.all"), i18n.T("show.supported"), i18n.T("show.installed")).AlignSelf(ui.Start).Label(i18n.T("show"))
	a.updatesBar(c, t)
	a.filter()
	// The chosen game shows as a lighter row with a bar beside it, as the navigation of Windows 11's Settings.
	ui.Themed(c, style.ApexNavTheme()).Grow(1).Children(func() {
		ui.List(c, &a.list, len(a.visible)).Grow(1).Rows(func(r ui.ListRow) {
			g := a.games[a.visible[r.Index]]
			a.gameRow(c, t, g, r.Selected())
		}).Children(func() {
			if len(a.visible) == 0 {
				msg := i18n.T("no.games")
				switch {
				case a.loadingGames:
					msg = i18n.T("scanning.steam")
				case a.noSteam && len(a.games) == 0:
					msg = i18n.T("no.steam")
				}
				style.RoleSmall.On(t, ui.Text(c, msg)).TextColor(t.TextMuted).Padding(style.PadPage).TextAlign(ui.Center)
			}
		})
	})
	supported, installed := 0, 0
	for _, g := range a.games {
		if g.Scan != nil && g.Scan.Supported() {
			supported++
		}
		if g.Scan != nil && g.Scan.Installed != detect.KindNone {
			installed++
		}
	}
	style.RoleSmall.On(t, ui.Text(c, i18n.T("games.count", len(a.games), supported, installed))).TextColor(t.TextMuted)
	style.RoleSmall.On(t, ui.Text(c, i18n.T("unofficial"))).TextColor(t.TextMuted).Tooltip(i18n.T("unofficial.tip"))
	ui.Row(c).Gap(style.GapControls).Children(func() {
		ui.Button(c, i18n.T("add.game")).Grow(1).Tooltip(i18n.T("add.game.tip")).OnClick(a.addFolder)
		ui.Button(c, i18n.T("refresh")).Disabled(a.loadingGames || a.busy).Tooltip(i18n.T("refresh.tip")).OnClick(a.scanSteam)
	})
}

func (a *app) gameRow(c *ui.Context, t *ui.Theme, g *game, selected bool) {
	text := t.Text
	if selected {
		text = style.ApexInk // on the navigation theme's lighter row
	}
	row := ui.Row(c).Padding(style.PadRowY, style.PadRowX).Gap(style.GapGroup)
	if g.Manual {
		row.ContextMenu(func(m *ui.Menu) {
			if m.Item(i18n.T("open.folder")).Chosen() {
				mygo.Shell.OpenPath(g.Dir)
			}
			m.Separator()
			if m.Item(i18n.T("remove")).Disabled(a.busy).Chosen() {
				a.removeGame(g)
			}
		})
	}
	row.Children(func() {
		// The bar sits at the row's edge, out of the flow, so that the icons line up with the update bar's.
		if selected {
			ui.Column(c).Absolute().Left(0).TopPercent(50).TranslatePercent(0, -50).Size(style.NavPillW, style.NavPillH).Radius(style.NavPillW).Background(style.ApexRoyal)
		}
		gameIcon(c, t, g, 36)
		ui.Column(c).Grow(1).Shrink(1).Gap(style.GapLines).Children(func() {
			style.RoleBodyStrong.On(t, ui.Text(c, g.Name)).TextColor(text).MaxLines(1).Ellipsis("…").Tooltip(g.Name)
			status, color := a.rowStatus(t, g)
			if selected {
				color = style.ApexMuted
			}
			// The status wraps, so that what ends it, as "anti-cheat", shows. Three lines hold the longest: a game
			// added by hand without frame generation, with anti-cheat, in Japanese.
			style.RoleSmall.On(t, ui.Text(c, status)).TextColor(color).MaxLines(3).Ellipsis("…").Tooltip(status)
		})
	})
}

// rowSep joins the parts of a game's status. Its first space does not break, so that a wrapped status ends a line
// with the dot rather than starts one with it.
const rowSep = " · "

// rowStatus is the line under a game's name: what is installed, else whether ANR FG can work in it. The versions
// are joined by spaces that do not break, so that a wrapped status keeps "ANR B18 → B19" on one line.
func (a *app) rowStatus(t *ui.Theme, g *game) (string, ui.Color) {
	r := g.Scan
	suffix := ""
	if g.Manual {
		suffix = rowSep + i18n.T("manual")
	}
	if r != nil && len(r.AntiCheat) > 0 {
		suffix += rowSep + i18n.T("anticheat.short")
	}
	switch {
	case g.Scanning || r == nil:
		return i18n.T("scanning"), t.TextMuted
	case r.Installed != detect.KindNone:
		v := "ANR FG"
		if iv := a.versionOf(r); iv != "" {
			v = "ANR " + strings.TrimPrefix(iv, "V.")
		}
		if latest, ok := a.outdated(r); ok {
			return v + " → " + strings.TrimPrefix(latest, "V.") + rowSep + i18n.T("row.upgrade") + suffix, t.Warning
		}
		return v + rowSep + string(r.Installed) + suffix, style.ApexPeriwinkle
	case r.Supported():
		return i18n.T("row.supported") + suffix, t.Success
	}
	return i18n.T("row.unsupported") + suffix, t.TextMuted
}

// gameIcon shows the game's icon, or its initial on a tile when it has none.
func gameIcon(c *ui.Context, t *ui.Theme, g *game, size float32) {
	if g.Icon != nil {
		ui.Image(c, g.Icon).Size(size, size).Fit(ui.Cover).Radius(size / 5)
		return
	}
	initial := "?"
	for _, r := range g.Name {
		initial = strings.ToUpper(string(r))
		break
	}
	ui.Column(c).Size(size, size).Radius(size / 5).Center().Background(t.Text.Alpha(0.08)).Children(func() {
		style.Glyph(ui.Text(c, initial), size).TextColor(t.TextMuted)
	})
}

// --- nothing to show ---

func (a *app) empty(c *ui.Context, t *ui.Theme) {
	ui.Column(c).Grow(1).Center().Gap(style.GapControls).Padding(style.PadEmpty).Children(func() {
		ui.Image(c, style.Logo).Size(72, 72)
		if a.loadingGames {
			style.RoleHeading.On(t, ui.Text(c, i18n.T("scanning.steam")))
			return
		}
		style.RoleHeading.On(t, ui.Text(c, i18n.T("empty.title")))
		ui.Text(c, i18n.T("empty.body")).TextColor(t.TextMuted).TextAlign(ui.Center).MaxWidth(420)
		ui.PrimaryButton(c, i18n.T("add.game")).OnClick(a.addFolder)
	})
}

// updatesBar tells how many games run an older ANR FG than the newest release, and shows them alone, or all
// games again, when clicked.
func (a *app) updatesBar(c *ui.Context, t *ui.Theme) {
	n := a.outdatedCount()
	if n == 0 && !a.onlyOutdated {
		return
	}
	latest := ""
	if a.rel != nil {
		latest = a.rel.Version
	}
	bar := ui.ButtonBase(c.Key("updates")).Padding(style.PadRowY, style.PadRowX).Radius(style.CardRadius).
		Background(t.Warning.Alpha(0.10)).Border(1, t.Warning.Alpha(0.30)).Label(i18n.T("updates.count", n, latest))
	if bar.Hovered() {
		bar.Background(t.Warning.Alpha(0.16))
	}
	bar.OnClick(func() { a.onlyOutdated = !a.onlyOutdated; a.row = -1 })
	style.Callout(c, bar, func() {
		ui.Icon(c, style.IconUpdate).Size(style.RowIcon, style.RowIcon).TextColor(t.Warning)
		style.RoleSmall.On(t, ui.Text(c, i18n.T("updates.count", n, latest)))
	}, func() {
		action := i18n.T("updates.show")
		if a.onlyOutdated {
			action = i18n.T("updates.all")
		}
		style.RoleSmall.On(t, ui.Text(c, action)).TextColor(t.Warning)
	})
}

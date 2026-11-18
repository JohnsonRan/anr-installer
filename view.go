package main

import (
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/internal/detect"
)

func (a *app) view(c *ui.Context) {
	// apexcompute.com's look, always dark, as the site is.
	t := apexTheme()
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
			ui.Column(c).Width(300).Padding(padPage).Gap(gapControls).Children(func() { a.sidebar(c, t) })
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
				a.activity(c, t)
				a.about(c, t)
				a.releaseNotes(c, t)
				a.quitDialog(c, t)
			})
		})
	})
}

// --- the list of games ---

func (a *app) sidebar(c *ui.Context, t *ui.Theme) {
	ui.SearchField(c, &a.query).Label(T("search")).Placeholder(T("search"))
	// As wide as its labels: shared equally, the list's width cuts the longest, "Supported" or "インストール済み".
	ui.Segmented(c, &a.show, T("show.all"), T("show.supported"), T("show.installed")).AlignSelf(ui.Start).Label(T("show"))
	a.updatesBar(c, t)
	a.filter()
	// The chosen game shows as a lighter row with a bar beside it, as the navigation of Windows 11's Settings.
	ui.Themed(c, apexNavTheme()).Grow(1).Children(func() {
		ui.List(c, &a.list, len(a.visible)).Grow(1).Rows(func(r ui.ListRow) {
			g := a.games[a.visible[r.Index]]
			a.gameRow(c, t, g, r.Selected())
		}).Children(func() {
			if len(a.visible) == 0 {
				msg := T("no.games")
				switch {
				case a.loadingGames:
					msg = T("scanning.steam")
				case a.noSteam && len(a.games) == 0:
					msg = T("no.steam")
				}
				roleSmall.on(t, ui.Text(c, msg)).TextColor(t.TextMuted).Padding(padPage).TextAlign(ui.Center)
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
	roleSmall.on(t, ui.Text(c, T("games.count", len(a.games), supported, installed))).TextColor(t.TextMuted)
	roleSmall.on(t, ui.Text(c, T("unofficial"))).TextColor(t.TextMuted).Tooltip(T("unofficial.tip"))
	ui.Row(c).Gap(gapControls).Children(func() {
		ui.Button(c, T("add.game")).Grow(1).Tooltip(T("add.game.tip")).OnClick(a.addFolder)
		ui.Button(c, T("refresh")).Disabled(a.loadingGames || a.busy).Tooltip(T("refresh.tip")).OnClick(a.scanSteam)
	})
}

func (a *app) gameRow(c *ui.Context, t *ui.Theme, g *game, selected bool) {
	text := t.Text
	if selected {
		text = apexInk // on the navigation theme's lighter row
	}
	row := ui.Row(c).Padding(padRowY, padRowX).Gap(gapGroup)
	if g.Manual {
		row.ContextMenu(func(m *ui.Menu) {
			if m.Item(T("open.folder")).Chosen() {
				mygo.Shell.OpenPath(g.Dir)
			}
			m.Separator()
			if m.Item(T("remove")).Disabled(a.busy).Chosen() {
				a.removeGame(g)
			}
		})
	}
	row.Children(func() {
		// The bar sits at the row's edge, out of the flow, so that the icons line up with the update bar's.
		if selected {
			ui.Column(c).Absolute().Left(0).TopPercent(50).TranslatePercent(0, -50).Size(navPillW, navPillH).Radius(navPillW).Background(apexRoyal)
		}
		gameIcon(c, t, g, 36)
		ui.Column(c).Grow(1).Shrink(1).Gap(gapLines).Children(func() {
			roleBodyStrong.on(t, ui.Text(c, g.Name)).TextColor(text).MaxLines(1).Ellipsis("…").Tooltip(g.Name)
			status, color := a.rowStatus(t, g)
			if selected {
				color = apexMuted
			}
			// The status wraps, so that what ends it, as "anti-cheat", shows. Three lines hold the longest: a game
			// added by hand without frame generation, with anti-cheat, in Japanese.
			roleSmall.on(t, ui.Text(c, status)).TextColor(color).MaxLines(3).Ellipsis("…").Tooltip(status)
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
		suffix = rowSep + T("manual")
	}
	if r != nil && len(r.AntiCheat) > 0 {
		suffix += rowSep + T("anticheat.short")
	}
	switch {
	case g.Scanning || r == nil:
		return T("scanning"), t.TextMuted
	case r.Installed != detect.KindNone:
		v := "ANR FG"
		if iv := a.versionOf(r); iv != "" {
			v = "ANR " + strings.TrimPrefix(iv, "V.")
		}
		if latest, ok := a.outdated(r); ok {
			return v + " → " + strings.TrimPrefix(latest, "V.") + rowSep + T("row.upgrade") + suffix, t.Warning
		}
		return v + rowSep + string(r.Installed) + suffix, apexPeriwinkle
	case r.Supported():
		return T("row.supported") + suffix, t.Success
	}
	return T("row.unsupported") + suffix, t.TextMuted
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
		glyph(ui.Text(c, initial), size).TextColor(t.TextMuted)
	})
}

// --- nothing to show ---

func (a *app) empty(c *ui.Context, t *ui.Theme) {
	ui.Column(c).Grow(1).Center().Gap(gapControls).Padding(padEmpty).Children(func() {
		ui.Image(c, logo).Size(72, 72)
		if a.loadingGames {
			roleHeading.on(t, ui.Text(c, T("scanning.steam")))
			return
		}
		roleHeading.on(t, ui.Text(c, T("empty.title")))
		ui.Text(c, T("empty.body")).TextColor(t.TextMuted).TextAlign(ui.Center).MaxWidth(420)
		ui.PrimaryButton(c, T("add.game")).OnClick(a.addFolder)
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
	bar := ui.ButtonBase(c.Key("updates")).Padding(padRowY, padRowX).Radius(cardRadius).
		Background(t.Warning.Alpha(0.10)).Border(1, t.Warning.Alpha(0.30)).Label(T("updates.count", n, latest))
	if bar.Hovered() {
		bar.Background(t.Warning.Alpha(0.16))
	}
	bar.OnClick(func() { a.onlyOutdated = !a.onlyOutdated; a.row = -1 })
	callout(c, bar, func() {
		ui.Icon(c, iconUpdate).Size(rowIcon, rowIcon).TextColor(t.Warning)
		roleSmall.on(t, ui.Text(c, T("updates.count", n, latest)))
	}, func() {
		action := T("updates.show")
		if a.onlyOutdated {
			action = T("updates.all")
		}
		roleSmall.on(t, ui.Text(c, action)).TextColor(t.Warning)
	})
}

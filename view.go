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
			ui.Column(c).Width(300).Padding(padPanel).Gap(gapControls).Children(func() { a.sidebar(c, t) })
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
	ui.Segmented(c, &a.show, T("show.all"), T("show.supported"), T("show.installed")).FillEqually().Label(T("show"))
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
				roleSmall.on(t, ui.Text(c, msg)).TextColor(t.TextMuted).Padding(padPanel).TextAlign(ui.Center)
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
	roleSmall.on(t, ui.Text(c, T("games.count", len(a.games), supported, installed))).TextColor(t.TextMuted).PaddingX(space1)
	roleSmall.on(t, ui.Text(c, T("unofficial"))).TextColor(t.TextMuted).PaddingX(space1).Tooltip(T("unofficial.tip"))
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
		pill := ui.Column(c).Size(navPillW, navPillH).Radius(navPillW)
		if selected {
			pill.Background(apexRoyal)
		}
		gameIcon(c, t, g, 36)
		ui.Column(c).Grow(1).Shrink(1).Gap(gapLines).Children(func() {
			roleBodyStrong.on(t, ui.Text(c, g.Name)).TextColor(text).MaxLines(1).Ellipsis("…")
			status, color := a.rowStatus(t, g)
			if selected {
				color = apexMuted
			}
			roleSmall.on(t, ui.Text(c, status)).TextColor(color).MaxLines(1).Ellipsis("…")
		})
	})
}

// rowStatus is the line under a game's name: what is installed, else whether ANR FG can work in it.
func (a *app) rowStatus(t *ui.Theme, g *game) (string, ui.Color) {
	r := g.Scan
	suffix := ""
	if g.Manual {
		suffix = " · " + T("manual")
	}
	if r != nil && len(r.AntiCheat) > 0 {
		suffix += " · " + T("anticheat.short")
	}
	switch {
	case g.Scanning || r == nil:
		return T("scanning"), t.TextMuted
	case r.Installed != detect.KindNone:
		v := "ANR FG"
		if iv := a.versionOf(r); iv != "" {
			v = "ANR " + strings.TrimPrefix(iv, "V.")
		}
		if latest, ok := a.outdated(r); ok {
			return v + " → " + strings.TrimPrefix(latest, "V.") + " · " + T("row.upgrade") + suffix, t.Warning
		}
		return v + " · " + string(r.Installed) + suffix, apexPeriwinkle
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

// --- activity ---

// activity is the log, folded to its last line until opened; it opens as a run starts.
func (a *app) activity(c *ui.Context, t *ui.Theme) {
	ui.Divider(c)
	ui.Column(c).Background(t.Text.Alpha(0.025)).Children(func() {
		head := ui.Row(c.Key("activity")).Padding(padRowY, padPage).Gap(gapControls).Cursor(ui.CursorPointer).Label(T("log"))
		if head.Clicked() {
			a.logOpen = !a.logOpen
		}
		head.Children(func() {
			arrow := "▸"
			if a.logOpen {
				arrow = "▾"
			}
			ui.Text(c, arrow).TextColor(t.TextMuted)
			roleBodyStrong.on(t, ui.Text(c, T("log")))
			last := a.status
			if last == "" && len(a.logLines) > 0 {
				last = a.logLines[len(a.logLines)-1]
			}
			roleSmall.on(t, ui.Text(c, strings.TrimSpace(last))).TextColor(t.TextMuted).MaxLines(1).Ellipsis("…").Grow(1).Shrink(1)
			switch {
			case a.busy && a.progress >= 0:
				ui.Progress(c, a.progress).Width(160).Label(a.status)
			case a.busy:
				ui.Spinner(c).Label(T("busy"))
			}
			if a.logOpen && len(a.logLines) > 0 && !a.busy {
				ui.Button(c, T("clear")).OnClick(func() { a.logLines, a.status = nil, "" })
			}
		})
		if a.logOpen {
			ui.List(c, &a.logList, len(a.logLines), func(i int) {
				roleSmall.on(t, ui.Text(c, a.logLines[i])).Font("monospace").Selectable().Padding(0, padPage)
			}).Height(180).PaddingY(space1).Children(func() {
				if len(a.logLines) == 0 {
					roleSmall.on(t, ui.Text(c, T("log.empty"))).TextColor(t.TextMuted).Padding(padRowY, padPage)
				}
			})
		}
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
	bar := ui.ButtonBase(c.Key("updates")).Padding(padRowY, padRowX).Gap(gapControls).Radius(cardRadius).
		Background(t.Warning.Alpha(0.10)).Border(1, t.Warning.Alpha(0.30)).Label(T("updates.count", n, latest))
	if bar.Hovered() {
		bar.Background(t.Warning.Alpha(0.16))
	}
	bar.OnClick(func() { a.onlyOutdated = !a.onlyOutdated; a.row = -1 }).Children(func() {
		ui.Icon(c, iconUpdate).Size(rowIcon, rowIcon).TextColor(t.Warning)
		roleSmall.on(t, ui.Text(c, T("updates.count", n, latest))).Grow(1).Shrink(1)
		action := T("updates.show")
		if a.onlyOutdated {
			action = T("updates.all")
		}
		roleSmall.on(t, ui.Text(c, action)).TextColor(t.Warning)
	})
}

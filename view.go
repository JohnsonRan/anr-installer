package main

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/internal/detect"
)

func (a *app) view(c *ui.Context) {
	// The system's own look: its light or dark appearance and its accent color.
	t := c.Theme()
	a.filter()
	if a.row < 0 && len(a.visible) > 0 {
		a.row = 0 // a game is always shown, the first one at start
	}
	ui.Column(c).Fill().Children(func() {
		a.titleBar(c, t)
		ui.Divider(c)
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Width(300).Padding(padPanel).Gap(gapControls).Background(t.Text.Alpha(0.025)).Children(func() { a.sidebar(c, t) })
			ui.Divider(c).Vertical()
			ui.Column(c).Grow(1).Shrink(1).Children(func() {
				g := a.selected()
				if g == nil {
					a.empty(c, t)
				} else {
					ui.Scroll(c.Key(g)).Grow(1).Children(func() {
						ui.Column(c).Padding(padPage).Gap(gapSection).Children(func() { a.detail(c, t, g) })
					})
				}
				a.activity(c, t)
			})
		})
	})
}

// --- the list of games ---

func (a *app) sidebar(c *ui.Context, t *ui.Theme) {
	ui.SearchField(c, &a.query).Label(T("search")).Placeholder(T("search"))
	ui.Segmented(c, &a.show, T("show.all"), T("show.supported"), T("show.installed")).FillEqually().Label(T("show"))
	a.filter()
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
	ui.Row(c).Gap(gapControls).Children(func() {
		ui.Button(c, T("add.game")).Grow(1).Tooltip(T("add.game.tip")).OnClick(a.addFolder)
		ui.Button(c, T("refresh")).Disabled(a.loadingGames || a.busy).Tooltip(T("refresh.tip")).OnClick(a.scanSteam)
	})
}

func (a *app) gameRow(c *ui.Context, t *ui.Theme, g *game, selected bool) {
	text := t.Text
	if selected {
		text = t.AccentText
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
		gameIcon(c, t, g, 36)
		ui.Column(c).Grow(1).Shrink(1).Gap(gapLines).Children(func() {
			roleBodyStrong.on(t, ui.Text(c, g.Name)).TextColor(text).MaxLines(1).Ellipsis("…")
			status, color := a.rowStatus(t, g)
			if selected {
				color = t.AccentText.Alpha(0.8)
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
	switch {
	case g.Scanning || r == nil:
		return T("scanning"), t.TextMuted
	case r.Installed != detect.KindNone:
		v := "ANR FG"
		if iv := a.versionOf(r); iv != "" {
			v = "ANR " + strings.TrimPrefix(iv, "V.")
		}
		return v + " · " + string(r.Installed) + suffix, t.Accent
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

// --- a game ---

func (a *app) detail(c *ui.Context, t *ui.Theme, g *game) {
	a.hero(c, t, g)
	r := g.Scan
	if r == nil {
		ui.Row(c).Gap(gapControls).Children(func() {
			ui.Spinner(c)
			ui.Text(c, T("scanning")).TextColor(t.TextMuted)
		})
		return
	}
	a.callouts(c, t, g)
	ui.Row(c).Gap(gapCards).AlignItems(ui.Stretch).Wrap().Children(func() {
		a.anrCard(c, t, r)
		a.dlssCard(c, t, r)
		a.optiCard(c, t, r)
	})
	a.installCard(c, t, g)
	a.locationCard(c, t, g)
}

// hero is the game's banner, as Steam's library heads a game's page: its hero picture with its logo over it,
// and what to do. A game without pictures gets a plain header of its icon and name.
func (a *app) hero(c *ui.Context, t *ui.Theme, g *game) {
	if g.Hero == nil {
		ui.Row(c).FillWidth().Padding(padCardY, padCardX).Gap(gapGroup).Radius(10).Background(t.Text.Alpha(0.035)).Border(1, t.Border.Alpha(0.6)).Children(func() {
			gameIcon(c, t, g, 52)
			a.title(c, t, g, t.Text, t.TextMuted)
			a.actions(c, g, false)
		})
		return
	}
	white := ui.RGB(255, 255, 255)
	ui.Column(c).FillWidth().Height(250).Radius(10).Clip().Justify(ui.End).Children(func() {
		ui.Image(c, g.Hero).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(ui.Cover).ObjectPosition(0.5, 0.3)
		ui.Row(c).FillWidth().ZIndex(1).Padding(heroFade, padCardX, padCardY, padCardX).Gap(gapGroup).AlignItems(ui.End).
			Gradient(ui.RGBA(0, 0, 0, 0), ui.RGBA(0, 0, 0, 0.72), 180).Children(func() {
			if g.Logo != nil && !g.Manual {
				ui.Column(c).Grow(1).Shrink(1).Gap(gapInCard).Children(func() {
					ui.Image(c, g.Logo).Size(320, 96).Fit(ui.Contain).ObjectPosition(0, 1).Label(g.Name)
					roleSmall.on(t, ui.Text(c, a.meta(g))).TextColor(white.Alpha(0.75))
				})
			} else {
				a.title(c, t, g, white, white.Alpha(0.75))
			}
			a.actions(c, g, true)
		})
	})
}

// title is the game's name, which the user can rename for a game added by hand, and where it comes from.
func (a *app) title(c *ui.Context, t *ui.Theme, g *game, color, muted ui.Color) {
	ui.Column(c).Grow(1).Shrink(1).Gap(gapLines).Children(func() {
		if g.Manual {
			if roleTitle.on(t, ui.EditableText(c.Key("name"), &g.Name)).TextColor(color).Tooltip(T("rename.tip")).Changed() {
				a.renamed(g)
			}
		} else {
			roleTitle.on(t, ui.Text(c, g.Name)).TextColor(color).MaxLines(1).Ellipsis("…")
		}
		roleSmall.on(t, ui.Text(c, a.meta(g))).TextColor(muted)
	})
}

func (a *app) meta(g *game) string {
	if g.Manual {
		return T("manual.long")
	}
	if g.AppID != "" {
		return "Steam · " + g.AppID
	}
	return "Steam"
}

// actions are the buttons of the banner, or the progress while something runs. Over a picture the secondary
// buttons are translucent.
func (a *app) actions(c *ui.Context, g *game, overArt bool) {
	white := ui.RGB(255, 255, 255)
	r := g.Scan
	ui.Row(c).Gap(gapControls).Children(func() {
		if a.busy {
			ui.Spinner(c).Label(T("busy"))
			if a.progress >= 0 {
				ui.Progress(c, a.progress).Width(140).Label(a.status)
			}
			return
		}
		ghost := func(e ui.Element) ui.Element {
			if !overArt {
				return e
			}
			return e.Background(white.Alpha(0.14)).TextColor(white).Border(1, white.Alpha(0.2))
		}
		if g.Manual {
			ghost(ui.Button(c, T("remove"))).Tooltip(T("remove.tip")).OnClick(func() { a.removeGame(g) })
		}
		if r == nil {
			return
		}
		if r.Installed != detect.KindNone {
			ghost(ui.Button(c, T("uninstall"))).Disabled(a.uninstallBlocked(r.Installed) != "" || g.Scanning).OnClick(func() { a.uninstall(g) })
		}
		k := a.chosenKind(g)
		why := a.why(g, k)
		b := ui.PrimaryButton(c, a.planLabel(a.plan(r))).Disabled(why != "" || g.Scanning).OnClick(func() { a.install(g, k) })
		if why != "" {
			b.Tooltip(why)
		}
	})
}

// why returns why installing kind k into g cannot run now, or "".
func (a *app) why(g *game, k detect.Kind) string {
	why := a.blocked(g, k)
	if why == "" && g.Scan != nil && a.plan(g.Scan).uninstallsFirst() {
		why = a.uninstallBlocked(g.Scan.Installed)
	}
	return why
}

// callouts say what needs the user's attention before installing.
func (a *app) callouts(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	k := a.chosenKind(g)
	if why := a.why(g, k); why != "" && !g.Scanning && !a.busy {
		callout(c, t, t.Danger, why)
	}
	if !r.Supported() && r.Installed == detect.KindNone {
		callout(c, t, t.Warning, T("dlssg.no"))
	}
	switch plan := a.plan(r); {
	case plan.uninstallsFirst():
		latest := ""
		if a.rel != nil {
			latest = a.rel.Version
		}
		callout(c, t, t.Warning, T("note.downgrade", latest, a.targetVersion()))
	case r.Installed == detect.KindFull && k == detect.KindFull && r.OptiINI:
		callout(c, t, t.Accent, T("note.ini.reset"))
	}
	if r.NRRecord {
		callout(c, t, t.Warning, T("note.nrrecord"))
	}
	if runtime.GOOS != "windows" {
		callout(c, t, t.Accent, T("note.proton"))
	}
}

func callout(c *ui.Context, t *ui.Theme, color ui.Color, s string) {
	ui.Row(c).Gap(gapControls).Padding(padCardY, padCardX).Radius(8).Background(color.Alpha(0.10)).Border(1, color.Alpha(0.35)).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Size(4, 16).Radius(2).Background(color)
		roleSmall.on(t, ui.Text(c, s)).Grow(1).Shrink(1)
	})
}

// card is a panel with a small title.
func card(c *ui.Context, t *ui.Theme, title string, body func()) ui.Element {
	return ui.Column(c).Padding(padCardY, padCardX).Gap(gapInCard).Radius(10).Background(t.Text.Alpha(0.035)).Border(1, t.Border.Alpha(0.6)).Children(func() {
		roleCaption.on(t, ui.Text(c, title)).TextColor(t.TextMuted)
		body()
	})
}

// statCard is a card of one value and a line about it.
func statCard(c *ui.Context, t *ui.Theme, title, value string, color ui.Color, detail string) {
	card(c, t, title, func() {
		ui.Row(c).Gap(gapControls).Children(func() {
			ui.Column(c).Size(8, 8).Radius(4).Background(color)
			roleHeading.on(t, ui.Text(c, value)).MaxLines(1).Ellipsis("…").Shrink(1)
		})
		if detail != "" {
			roleSmall.on(t, ui.Text(c, detail)).TextColor(t.TextMuted).MaxLines(1).TruncateMiddle().Tooltip(detail)
		}
	}).Grow(1).Basis(0).MinWidth(180)
}

func (a *app) anrCard(c *ui.Context, t *ui.Theme, r *detect.Result) {
	switch r.Installed {
	case detect.KindFull:
		statCard(c, t, "ANR FG", a.installedVersion(r), t.Accent, T("anr.full", r.Proxy))
	case detect.KindAddon:
		statCard(c, t, "ANR FG", a.installedVersion(r), t.Accent, T("anr.addon"))
	default:
		statCard(c, t, "ANR FG", T("anr.none"), t.TextMuted, "")
	}
}

func (a *app) dlssCard(c *ui.Context, t *ui.Theme, r *detect.Result) {
	if !r.Supported() {
		statCard(c, t, T("dlss.title"), T("dlss.none"), t.Warning, T("dlss.none.detail"))
		return
	}
	where, err := filepath.Rel(r.Root, filepath.Dir(r.DLSSG[0]))
	if err != nil || where == "." {
		where = T("dlss.root")
	}
	if r.Streamline {
		where = "Streamline · " + where
	}
	statCard(c, t, T("dlss.title"), T("dlss.found"), t.Success, where)
}

func (a *app) optiCard(c *ui.Context, t *ui.Theme, r *detect.Result) {
	detail := ""
	switch {
	case r.Replacement == detect.ANRDir:
		detail = T("opti.fg", "ANR FG")
	case r.Replacement != "":
		detail = T("opti.fg", r.Replacement)
	}
	switch {
	case r.Installed == detect.KindFull:
		statCard(c, t, "OptiScaler", T("opti.ours"), t.Accent, detail)
	case r.UserOpti:
		statCard(c, t, "OptiScaler", T("opti.yours"), t.Success, detail)
	default:
		statCard(c, t, "OptiScaler", T("opti.none"), t.TextMuted, T("opti.none.detail"))
	}
}

// installCard chooses what to install: the version up front, the package and its source under Advanced.
func (a *app) installCard(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	card(c, t, T("install.title"), func() {
		ui.Row(c).Gap(gapLabelField).Wrap().Children(func() {
			ui.Text(c, T("version")).Width(labelWidth).TextColor(t.TextMuted)
			switch {
			case a.source == sourceLocal:
				if a.local != nil {
					ui.Text(c, T("zip.info", filepath.Base(a.localZip), a.local.Version, a.local.Kind)).MaxLines(1).TruncateMiddle().Shrink(1)
				} else {
					ui.Text(c, T("zip.none")).TextColor(t.TextMuted)
				}
			case len(a.versionLabels) > 0:
				ui.Select(c, &a.versionSel, a.versionLabels).Label(T("version")).Disabled(a.busy).MinWidth(240)
				if rel := a.target(); rel != nil {
					ui.Button(c, T("release.notes")).OnClick(func() { mygo.Shell.OpenExternal(rel.URL) })
				}
			case a.relLoading:
				ui.Spinner(c)
				ui.Text(c, T("latest.loading")).TextColor(t.TextMuted)
			default:
				ui.Text(c, T("latest.error", a.relErr)).TextColor(t.Danger).MaxLines(2).Shrink(1)
				ui.Button(c, T("retry")).OnClick(a.fetchRelease)
			}
		})
		ui.Collapsible(c, T("advanced"), &a.advanced, func() {
			ui.Column(c).Gap(gapInCard).PaddingY(space1).Children(func() {
				ui.Row(c).Gap(gapLabelField).Children(func() {
					ui.Text(c, T("package")).Width(labelWidth).TextColor(t.TextMuted)
					ui.Segmented(c, &a.kind, T("kind.auto", r.Recommended()), "full", "addon").Label(T("package")).Disabled(a.busy)
				})
				roleSmall.on(t, ui.Text(c, T("package.help"))).TextColor(t.TextMuted).PaddingStart(labelWidth + gapLabelField)
				ui.Row(c).Gap(gapLabelField).Children(func() {
					ui.Text(c, T("source")).Width(labelWidth).TextColor(t.TextMuted)
					ui.Segmented(c, &a.source, T("source.github0"), T("source.local")).Label(T("source")).Disabled(a.busy)
					if a.source == sourceLocal {
						ui.Button(c, T("choose.zip")).Disabled(a.busy).OnClick(a.chooseZip)
					}
				})
				if a.localErr != "" && a.source == sourceLocal {
					roleSmall.on(t, ui.Text(c, a.localErr)).TextColor(t.Danger).PaddingStart(labelWidth + gapLabelField)
				}
			})
		})
	})
}

// locationCard shows where the installers go: the folder of the game's .exe.
func (a *app) locationCard(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	card(c, t, T("exe.folder"), func() {
		roleSmall.on(t, ui.Text(c, r.ExeDir)).Font("monospace").Selectable().MaxLines(2).TruncateMiddle()
		if r.Exe != "" {
			roleSmall.on(t, ui.Text(c, filepath.Base(r.Exe))).TextColor(t.TextMuted)
		}
		ui.Row(c).Gap(gapControls).Children(func() {
			ui.Button(c, T("change")).Disabled(a.busy || g.Scanning).OnClick(func() { a.changeExeDir(g) })
			if g.ExeDir != "" {
				ui.Button(c, T("auto.exe")).Disabled(a.busy || g.Scanning).Tooltip(T("auto.exe.tip")).OnClick(func() { a.resetExeDir(g) })
			}
			ui.Button(c, T("open.folder")).OnClick(func() {
				if r.Exe != "" {
					mygo.Shell.ShowItemInFolder(r.Exe)
				} else {
					mygo.Shell.OpenPath(r.ExeDir)
				}
			})
		})
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

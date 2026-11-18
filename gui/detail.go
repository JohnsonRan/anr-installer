package gui

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
	"anr-installer/internal/detect"
)

// The page of a game, laid out as a page of Windows 11's Settings: a header with the game and what to do, info
// bars for what needs the user's attention, then cards, one a row, for what the installer found.

func (a *app) detail(c *ui.Context, t *ui.Theme, g *game) {
	ui.Column(c).Padding(style.PadPage, style.PadPageX).Gap(style.GapSection).Children(func() {
		a.header(c, t, g)
		r := g.Scan
		if r == nil {
			ui.Row(c).Gap(style.GapControls).Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("scanning")).TextColor(t.TextMuted)
			})
			return
		}
		a.notices(c, t, g)
		style.Section(c, t, i18n.T("status"), nil, func() { a.statusCards(c, t, g) })
		a.recordSection(c, t, g)
		a.advancedCard(c, t, g)
	})
}

// header names the game and holds the actions: install the chosen version, uninstall, or take a game added by
// hand off the list. While a step runs, it shows its progress. A game with Steam library art gets it as the
// header card's background, its logo over it; others get their icon and name.
func (a *app) header(c *ui.Context, t *ui.Theme, g *game) {
	if g.Hero != nil {
		a.heroHeader(c, t, g)
		return
	}
	// The actions go under the game where the window is too narrow for them beside it.
	ui.Row(c).Wrap().Justify(ui.SpaceBetween).AlignItems(ui.Center).GapX(style.GapGroup).GapY(style.GapInCard).Children(func() {
		ui.Row(c).Gap(style.GapGroup).Shrink(1).Children(func() {
			gameIcon(c, t, g, style.HeaderIcon)
			ui.Column(c).Shrink(1).Gap(style.GapLines).Children(func() {
				if g.Manual {
					if style.RoleTitle.On(t, ui.EditableText(c.Key("name"), &g.Name)).Tooltip(i18n.T("rename.tip")).Changed() {
						a.renamed(g)
					}
				} else {
					style.RoleTitle.On(t, ui.Text(c, g.Name)).MaxLines(1).Ellipsis("…").Tooltip(g.Name)
				}
				line := a.headerLine(g)
				style.RoleSmall.On(t, ui.Text(c, line)).TextColor(t.TextMuted).MaxLines(1).Ellipsis("…").Tooltip(line)
			})
		})
		a.actions(c, t, g)
	})
}

// heroHeader is the header as a card of the game's library art, darkened toward its bottom, where the logo
// (else the name), the state line and the actions sit.
func (a *app) heroHeader(c *ui.Context, t *ui.Theme, g *game) {
	white := ui.RGB(255, 255, 255)
	// The card grows past its height where the actions go under the game, in a narrow window.
	ui.Column(c).FillWidth().MinHeight(style.HeroHeight).Radius(style.CardRadius).Clip().Border(1, style.ApexLine).Justify(ui.End).Children(func() {
		ui.Image(c, g.Hero).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(ui.Cover).ObjectPosition(0.5, 0.3)
		ui.Row(c).FillWidth().ZIndex(1).Padding(style.HeroFade, style.PadCardX, style.PadCardY, style.PadCardX).Wrap().Justify(ui.SpaceBetween).AlignItems(ui.End).
			GapX(style.GapGroup).GapY(style.GapInCard).Gradient(style.ApexBG.Alpha(0), style.ApexBG.Alpha(0.92), 180).Children(func() {
			ui.Column(c).Shrink(1).Gap(style.GapLines).Children(func() {
				if g.Logo != nil {
					ui.Image(c, g.Logo).Size(style.HeroLogoW, style.HeroLogoH).Fit(ui.Contain).ObjectPosition(0, 1).Label(g.Name)
				} else {
					style.RoleTitle.On(t, ui.Text(c, g.Name)).TextColor(white).MaxLines(1).Ellipsis("…").Tooltip(g.Name)
				}
				line := a.headerLine(g)
				style.RoleSmall.On(t, ui.Text(c, line)).TextColor(white.Alpha(0.8)).MaxLines(1).Ellipsis("…").Tooltip(line)
			})
			a.actions(c, t, g)
		})
	})
}

// headerLine says where the game comes from and what is installed in it.
func (a *app) headerLine(g *game) string {
	from := "Steam"
	if g.Manual {
		from = i18n.T("manual.long")
	}
	r := g.Scan
	switch {
	case r == nil:
		return from
	case r.Installed != detect.KindNone:
		return from + " · " + i18n.T("state.installed", a.installedVersion(r), r.Installed)
	}
	return from + " · " + i18n.T("anr.none")
}

// actions are the header's buttons, or the progress of the step that runs.
func (a *app) actions(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	// Under the game, in a narrow window, the buttons wrap onto more lines where one is too narrow for them.
	ui.Row(c).Gap(style.GapControls).Wrap().Shrink(1).Children(func() {
		if a.busy {
			ui.Spinner(c).Label(i18n.T("busy"))
			if a.progress >= 0 {
				ui.Progress(c, a.progress).Width(style.ProgressWidth).Label(a.status)
			}
			return
		}
		if g.Manual {
			ui.Button(c, i18n.T("remove")).Tooltip(i18n.T("remove.tip")).OnClick(func() { a.removeGame(g) })
		}
		if r == nil {
			return
		}
		if g.AppID != "" || r.Exe != "" {
			ui.Button(c, i18n.T("play")).Tooltip(i18n.T("play.tip")).OnClick(func() { a.play(g) })
		}
		if r.Installed != detect.KindNone {
			ui.Button(c, i18n.T("uninstall")).Disabled(a.uninstallBlocked(r.Installed) != "" || g.Scanning).OnClick(func() { a.uninstall(g) })
		}
		a.installButton(c, t, g)
	})
}

// installButton installs the version chosen in its menu, which lists the releases, the newest first, and the
// chosen one's release notes. With a local zip as the source, it is a plain button.
func (a *app) installButton(c *ui.Context, t *ui.Theme, g *game) {
	k := a.chosenKind(g)
	why := a.why(g, k)
	label := a.planLabel(a.plan(g.Scan))
	if a.source == sourceLocal || len(a.versionLabels) == 0 {
		b := ui.PrimaryButton(c, label).Disabled(why != "" || g.Scanning).OnClick(func() { a.confirmInstall(g, k) })
		if why != "" {
			b.Tooltip(why)
		}
		return
	}
	split := ui.PrimarySplitButton(c, label)
	split.Button.Disabled(why != "" || g.Scanning).OnClick(func() { a.confirmInstall(g, k) })
	if why != "" {
		split.Button.Tooltip(why)
	}
	split.Arrow.Label(i18n.T("version")).Tooltip(i18n.T("version.choose"))
	split.Arrow.PopoverMenuButton(func(m *ui.Menu) {
		for i, label := range a.versionLabels {
			if m.Item(label).Checked(label == a.versionSel).Chosen() {
				a.versionSel = a.versionLabels[i]
			}
		}
		if rel := a.target(); rel != nil {
			m.Separator()
			if m.Item(i18n.T("release.notes.of", rel.Version)).Chosen() {
				a.showNotes(rel)
			}
		}
	}, ui.PopoverMenuButtonOptions{At: ui.AnchorBottomEnd, Self: ui.AnchorTopEnd})
}

// why returns why installing kind k into g cannot run now, or "".
func (a *app) why(g *game, k detect.Kind) string {
	why := a.blocked(g, k)
	if why == "" && g.Scan != nil && a.plan(g.Scan).uninstallsFirst() {
		why = a.uninstallBlocked(g.Scan.Installed)
	}
	if why == "" && a.source == sourceGitHub && a.relErr != "" {
		why = i18n.T("latest.error", a.relErr)
	}
	return why
}

// notices are info bars, as Windows 11 shows them: what stops the install, the anti-cheat risk, then what an
// install will do to the user's files.
func (a *app) notices(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	k := a.chosenKind(g)
	// The bars are gathered first: with none, the page has no group for them, whose gaps would double the room
	// between the header and the cards.
	var bars []func()
	bar := func(color ui.Color, icon *ui.SVG, s string, controls func()) {
		bars = append(bars, func() { style.InfoBar(c, t, color, icon, s, controls) })
	}
	if why := a.why(g, k); why != "" && !g.Scanning && !a.busy {
		var retry func()
		if a.relErr != "" {
			retry = func() { ui.Button(c, i18n.T("retry")).OnClick(a.fetchRelease) }
		}
		bar(t.Danger, style.IconWarning, why, retry)
	}
	if latest, ok := a.outdated(r); ok {
		bar(t.Warning, style.IconUpdate, i18n.T("update.available", latest, a.versionOf(r)), func() {
			ui.Button(c, i18n.T("notes.whatsnew")).OnClick(func() { a.showNotes(a.rel) })
		})
	}
	if len(r.AntiCheat) > 0 {
		bar(t.Warning, style.IconWarning, i18n.T("anticheat.found", strings.Join(r.AntiCheat, ", ")), nil)
	}
	if !r.Supported() && r.Installed == detect.KindNone {
		bar(t.Warning, style.IconWarning, i18n.T("dlssg.no"), nil)
	}
	switch plan := a.plan(r); {
	case plan.uninstallsFirst():
		latest := ""
		if a.rel != nil {
			latest = a.rel.Version
		}
		bar(style.ApexRoyal, style.IconInfo, i18n.T("note.downgrade", latest, a.targetVersion()), nil)
	case r.Installed == detect.KindFull && k == detect.KindFull && r.OptiINI:
		bar(style.ApexRoyal, style.IconInfo, i18n.T("note.ini.reset"), nil)
	}
	if r.NRRecord {
		bar(style.ApexRoyal, style.IconInfo, i18n.T("note.nrrecord"), nil)
	}
	if runtime.GOOS != "windows" {
		bar(style.ApexRoyal, style.IconInfo, i18n.T("note.proton"), nil)
	}
	if len(bars) == 0 {
		return
	}
	ui.Column(c).Gap(style.GapCardRows).Children(func() {
		for _, b := range bars {
			b()
		}
	})
}

// statusCards are what the installer found in the game's folder, a card each, and where the installers go.
func (a *app) statusCards(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	switch r.Installed {
	case detect.KindFull:
		style.SettingsRow(c, t, style.IconChip, "ANR FG", i18n.T("anr.full", r.Proxy), func() { valueText(c, a.installedVersion(r)+" · full", t.Text) })
	case detect.KindAddon:
		style.SettingsRow(c, t, style.IconChip, "ANR FG", i18n.T("anr.addon"), func() { valueText(c, a.installedVersion(r)+" · addon", t.Text) })
	default:
		style.SettingsRow(c, t, style.IconChip, "ANR FG", i18n.T("anr.none.detail"), func() { valueText(c, i18n.T("anr.none"), t.TextMuted) })
	}

	if r.Supported() {
		where, err := filepath.Rel(r.Root, filepath.Dir(r.DLSSG[0]))
		if err != nil || where == "." {
			where = i18n.T("dlss.root")
		}
		if r.Streamline {
			where = "Streamline · " + where
		}
		style.ValueRow(c, t, style.IconFrames, i18n.T("dlss.title"), where, func() { statusValue(c, i18n.T("dlss.found"), t.Success) })
	} else {
		style.SettingsRow(c, t, style.IconFrames, i18n.T("dlss.title"), i18n.T("dlss.none.detail"), func() { statusValue(c, i18n.T("dlss.none"), t.Warning) })
	}

	detail := ""
	switch {
	case r.Installed == detect.KindNone && !r.UserOpti:
		detail = i18n.T("opti.none.detail")
	case r.Replacement == detect.ANRDir:
		detail = i18n.T("opti.fg", "ANR FG")
	case r.Replacement != "":
		detail = i18n.T("opti.fg", r.Replacement)
	}
	switch {
	case r.Installed == detect.KindFull:
		style.SettingsRow(c, t, style.IconLayers, "OptiScaler", detail, func() { valueText(c, i18n.T("opti.ours"), t.Text) })
	case r.UserOpti:
		style.SettingsRow(c, t, style.IconLayers, "OptiScaler", detail, func() { valueText(c, i18n.T("opti.yours"), t.Text) })
	default:
		style.SettingsRow(c, t, style.IconLayers, "OptiScaler", detail, func() { valueText(c, i18n.T("opti.none"), t.TextMuted) })
	}

	style.ValueRow(c, t, style.IconFolder, i18n.T("exe.folder"), r.ExeDir, func() {
		ui.Button(c, i18n.T("change")).Disabled(a.busy || g.Scanning).OnClick(func() { a.changeExeDir(g) })
		if g.ExeDir != "" {
			ui.Button(c, i18n.T("auto.exe")).Disabled(a.busy || g.Scanning).Tooltip(i18n.T("auto.exe.tip")).OnClick(func() { a.resetExeDir(g) })
		}
		ui.Button(c, i18n.T("open.folder")).OnClick(func() {
			if r.Exe != "" {
				mygo.Shell.ShowItemInFolder(r.Exe)
			} else {
				mygo.Shell.OpenPath(r.ExeDir)
			}
		})
	})
}

func valueText(c *ui.Context, s string, color ui.Color) {
	ui.Text(c, s).TextColor(color).MaxLines(1)
}

// statusValue is a value with a dot of its color, for what was found or not.
func statusValue(c *ui.Context, s string, color ui.Color) {
	ui.Row(c).Gap(style.GapControls).Children(func() {
		ui.Column(c).Size(style.StatusDot, style.StatusDot).Radius(style.StatusDot / 2).Background(color)
		ui.Text(c, s).MaxLines(1)
	})
}

// advancedCard is an expander, as Windows 11's Settings has: the package kind and its source, which Auto and
// GitHub get right for most games, open below its header.
func (a *app) advancedCard(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	ui.Column(c).Radius(style.CardRadius).Background(style.ApexSurface).Border(1, style.ApexLine).Clip().Children(func() {
		head := ui.ButtonBase(c.Key("advanced")).Padding(style.PadCardY, style.PadCardX).MinHeight(style.SettingsRowHeight).Gap(style.GapGroup).Label(i18n.T("advanced"))
		if head.Hovered() {
			head.Background(style.ApexSurface2)
		}
		head.OnClick(func() { a.advanced = !a.advanced }).Children(func() {
			ui.Icon(c, style.IconSliders).Size(style.RowIcon, style.RowIcon).TextColor(t.TextMuted)
			ui.Column(c).Grow(1).Shrink(1).Gap(style.GapLines).Children(func() {
				ui.Text(c, i18n.T("advanced"))
				style.RoleSmall.On(t, ui.Text(c, i18n.T("advanced.desc"))).TextColor(t.TextMuted)
			})
			chevron := style.IconChevronD
			if a.advanced {
				chevron = style.IconChevronU
			}
			ui.Icon(c, chevron).Size(style.RowIcon, style.RowIcon).TextColor(t.TextMuted)
		})
		if !a.advanced {
			return
		}
		ui.Divider(c)
		ui.Column(c).Padding(style.PadCardY, style.PadCardX).PaddingStart(style.PadCardX + style.RowIcon + style.GapGroup).Gap(style.GapInCard).Children(func() {
			ui.Row(c).Gap(style.GapLabelField).Children(func() {
				ui.Text(c, i18n.T("package")).Width(style.LabelWidth).TextColor(t.TextMuted)
				ui.Row(c).Disabled(a.busy).Children(func() {
					style.Segmented(c, t, &a.kind, i18n.T("kind.auto", r.Recommended()), "full", "addon").Label(i18n.T("package"))
				})
			})
			style.RoleSmall.On(t, ui.Text(c, i18n.T("package.help"))).TextColor(t.TextMuted).PaddingStart(style.LabelWidth + style.GapLabelField)
			ui.Row(c).Gap(style.GapLabelField).Children(func() {
				ui.Text(c, i18n.T("source")).Width(style.LabelWidth).TextColor(t.TextMuted)
				ui.Row(c).Disabled(a.busy).Children(func() {
					style.Segmented(c, t, &a.source, i18n.T("source.github0"), i18n.T("source.local")).Label(i18n.T("source"))
				})
				if a.source == sourceLocal {
					ui.Button(c, i18n.T("choose.zip")).Disabled(a.busy).OnClick(a.chooseZip)
				}
			})
			if a.source == sourceLocal {
				msg, color := i18n.T("zip.none"), t.TextMuted
				switch {
				case a.localErr != "":
					msg, color = a.localErr, t.Danger
				case a.local != nil:
					msg, color = i18n.T("zip.info", filepath.Base(a.localZip), a.local.Version, a.local.Kind), t.Text
				}
				style.RoleSmall.On(t, ui.Text(c, msg)).TextColor(color).PaddingStart(style.LabelWidth + style.GapLabelField).MaxLines(2)
			}
		})
	})
}

// card is a group on a surface, with a small title, for the About dialog.
func card(c *ui.Context, t *ui.Theme, title string, body func()) ui.Element {
	return ui.Column(c).Padding(style.PadCardY, style.PadCardX).Gap(style.GapInCard).Radius(style.CardRadius).Background(style.ApexSurface).Border(1, style.ApexLine).Children(func() {
		style.RoleCaption.On(t, ui.Text(c, title)).TextColor(t.TextMuted)
		body()
	})
}

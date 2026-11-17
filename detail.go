package main

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/internal/detect"
)

// The page of a game, laid out as a page of Windows 11's Settings: a header with the game and what to do, info
// bars for what needs the user's attention, then cards, one a row, for what the installer found.

func (a *app) detail(c *ui.Context, t *ui.Theme, g *game) {
	ui.Column(c).Padding(padPage, padPageX).Gap(gapSection).Children(func() {
		a.header(c, t, g)
		r := g.Scan
		if r == nil {
			ui.Row(c).Gap(gapControls).Children(func() {
				ui.Spinner(c)
				ui.Text(c, T("scanning")).TextColor(t.TextMuted)
			})
			return
		}
		a.notices(c, t, g)
		ui.Column(c).Gap(gapCardRows).Children(func() {
			roleBodyStrong.on(t, ui.Text(c, T("status"))).PaddingY(space1)
			a.statusCards(c, t, g)
		})
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
	ui.Row(c).Gap(gapGroup).Children(func() {
		gameIcon(c, t, g, headerIcon)
		ui.Column(c).Grow(1).Shrink(1).Gap(gapLines).Children(func() {
			if g.Manual {
				if roleTitle.on(t, ui.EditableText(c.Key("name"), &g.Name)).Tooltip(T("rename.tip")).Changed() {
					a.renamed(g)
				}
			} else {
				roleTitle.on(t, ui.Text(c, g.Name)).MaxLines(1).Ellipsis("…")
			}
			roleSmall.on(t, ui.Text(c, a.headerLine(g))).TextColor(t.TextMuted).MaxLines(1).Ellipsis("…")
		})
		a.actions(c, t, g)
	})
}

// heroHeader is the header as a card of the game's library art, darkened toward its bottom, where the logo
// (else the name), the state line and the actions sit.
func (a *app) heroHeader(c *ui.Context, t *ui.Theme, g *game) {
	white := ui.RGB(255, 255, 255)
	ui.Column(c).FillWidth().Height(heroHeight).Radius(cardRadius).Clip().Border(1, apexLine).Justify(ui.End).Children(func() {
		ui.Image(c, g.Hero).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(ui.Cover).ObjectPosition(0.5, 0.3)
		ui.Row(c).FillWidth().ZIndex(1).Padding(heroFade, padCardX, padCardY, padCardX).Gap(gapGroup).AlignItems(ui.End).
			Gradient(apexBG.Alpha(0), apexBG.Alpha(0.92), 180).Children(func() {
			ui.Column(c).Grow(1).Shrink(1).Gap(gapLines).Children(func() {
				if g.Logo != nil {
					ui.Image(c, g.Logo).Size(heroLogoW, heroLogoH).Fit(ui.Contain).ObjectPosition(0, 1).Label(g.Name)
				} else {
					roleTitle.on(t, ui.Text(c, g.Name)).TextColor(white).MaxLines(1).Ellipsis("…")
				}
				roleSmall.on(t, ui.Text(c, a.headerLine(g))).TextColor(white.Alpha(0.8)).MaxLines(1).Ellipsis("…")
			})
			a.actions(c, t, g)
		})
	})
}

// headerLine says where the game comes from and what is installed in it.
func (a *app) headerLine(g *game) string {
	from := "Steam"
	if g.Manual {
		from = T("manual.long")
	}
	r := g.Scan
	switch {
	case r == nil:
		return from
	case r.Installed != detect.KindNone:
		return from + " · " + T("state.installed", a.installedVersion(r), r.Installed)
	}
	return from + " · " + T("anr.none")
}

// actions are the header's buttons, or the progress of the step that runs.
func (a *app) actions(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	// The buttons wrap onto a second line where the window is too narrow for them beside the game's name.
	ui.Row(c).Gap(gapControls).Wrap().Justify(ui.End).Shrink(1).Children(func() {
		if a.busy {
			ui.Spinner(c).Label(T("busy"))
			if a.progress >= 0 {
				ui.Progress(c, a.progress).Width(progressWidth).Label(a.status)
			}
			return
		}
		if g.Manual {
			ui.Button(c, T("remove")).Tooltip(T("remove.tip")).OnClick(func() { a.removeGame(g) })
		}
		if r == nil {
			return
		}
		if g.AppID != "" || r.Exe != "" {
			ui.Button(c, T("play")).Tooltip(T("play.tip")).OnClick(func() { a.play(g) })
		}
		if r.Installed != detect.KindNone {
			ui.Button(c, T("uninstall")).Disabled(a.uninstallBlocked(r.Installed) != "" || g.Scanning).OnClick(func() { a.uninstall(g) })
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
	split.Arrow.Label(T("version")).Tooltip(T("version.choose"))
	split.Arrow.PopoverMenuButton(func(m *ui.Menu) {
		for i, label := range a.versionLabels {
			if m.Item(label).Checked(label == a.versionSel).Chosen() {
				a.versionSel = a.versionLabels[i]
			}
		}
		if rel := a.target(); rel != nil {
			m.Separator()
			if m.Item(T("release.notes.of", rel.Version)).Chosen() {
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
		why = T("latest.error", a.relErr)
	}
	return why
}

// notices are info bars, as Windows 11 shows them: what stops the install, the anti-cheat risk, then what an
// install will do to the user's files.
func (a *app) notices(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	k := a.chosenKind(g)
	ui.Column(c).Gap(gapCardRows).Children(func() {
		if why := a.why(g, k); why != "" && !g.Scanning && !a.busy {
			infoBar(c, t, t.Danger, iconWarning, why, func() {
				if a.relErr != "" {
					ui.Button(c, T("retry")).OnClick(a.fetchRelease)
				}
			})
		}
		if latest, ok := a.outdated(r); ok {
			infoBar(c, t, t.Warning, iconUpdate, T("update.available", latest, a.versionOf(r)), func() {
				ui.Button(c, T("notes.whatsnew")).OnClick(func() { a.showNotes(a.rel) })
			})
		}
		if len(r.AntiCheat) > 0 {
			infoBar(c, t, t.Warning, iconWarning, T("anticheat.found", strings.Join(r.AntiCheat, ", ")), nil)
		}
		if !r.Supported() && r.Installed == detect.KindNone {
			infoBar(c, t, t.Warning, iconWarning, T("dlssg.no"), nil)
		}
		switch plan := a.plan(r); {
		case plan.uninstallsFirst():
			latest := ""
			if a.rel != nil {
				latest = a.rel.Version
			}
			infoBar(c, t, apexRoyal, iconInfo, T("note.downgrade", latest, a.targetVersion()), nil)
		case r.Installed == detect.KindFull && k == detect.KindFull && r.OptiINI:
			infoBar(c, t, apexRoyal, iconInfo, T("note.ini.reset"), nil)
		}
		if r.NRRecord {
			infoBar(c, t, apexRoyal, iconInfo, T("note.nrrecord"), nil)
		}
		if runtime.GOOS != "windows" {
			infoBar(c, t, apexRoyal, iconInfo, T("note.proton"), nil)
		}
	})
}

// infoBar is a tinted bar with an icon in its color, a message, and controls at its end.
func infoBar(c *ui.Context, t *ui.Theme, color ui.Color, icon *ui.SVG, s string, controls func()) {
	ui.Row(c).Gap(gapGroup).Padding(padCardY, padCardX).Radius(cardRadius).Background(color.Alpha(0.10)).Border(1, color.Alpha(0.30)).AlignItems(ui.Start).Children(func() {
		ui.Icon(c, icon).Size(rowIcon, rowIcon).TextColor(color)
		ui.Text(c, s).Grow(1).Shrink(1)
		if controls != nil {
			ui.Row(c).Gap(gapControls).Children(controls)
		}
	})
}

// statusCards are what the installer found in the game's folder, a card each, and where the installers go.
func (a *app) statusCards(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	switch r.Installed {
	case detect.KindFull:
		settingsRow(c, t, iconChip, "ANR FG", T("anr.full", r.Proxy), func() { valueText(c, a.installedVersion(r)+" · full", t.Text) })
	case detect.KindAddon:
		settingsRow(c, t, iconChip, "ANR FG", T("anr.addon"), func() { valueText(c, a.installedVersion(r)+" · addon", t.Text) })
	default:
		settingsRow(c, t, iconChip, "ANR FG", T("anr.none.detail"), func() { valueText(c, T("anr.none"), t.TextMuted) })
	}

	if r.Supported() {
		where, err := filepath.Rel(r.Root, filepath.Dir(r.DLSSG[0]))
		if err != nil || where == "." {
			where = T("dlss.root")
		}
		if r.Streamline {
			where = "Streamline · " + where
		}
		settingsRow(c, t, iconFrames, T("dlss.title"), where, func() { statusValue(c, T("dlss.found"), t.Success) })
	} else {
		settingsRow(c, t, iconFrames, T("dlss.title"), T("dlss.none.detail"), func() { statusValue(c, T("dlss.none"), t.Warning) })
	}

	detail := ""
	switch {
	case r.Installed == detect.KindNone && !r.UserOpti:
		detail = T("opti.none.detail")
	case r.Replacement == detect.ANRDir:
		detail = T("opti.fg", "ANR FG")
	case r.Replacement != "":
		detail = T("opti.fg", r.Replacement)
	}
	switch {
	case r.Installed == detect.KindFull:
		settingsRow(c, t, iconLayers, "OptiScaler", detail, func() { valueText(c, T("opti.ours"), t.Text) })
	case r.UserOpti:
		settingsRow(c, t, iconLayers, "OptiScaler", detail, func() { valueText(c, T("opti.yours"), t.Text) })
	default:
		settingsRow(c, t, iconLayers, "OptiScaler", detail, func() { valueText(c, T("opti.none"), t.TextMuted) })
	}

	settingsRow(c, t, iconFolder, T("exe.folder"), r.ExeDir, func() {
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
}

// settingsRow is a card of Windows 11's Settings: an icon, a title over its description, and what goes at its
// end, a value or controls.
func settingsRow(c *ui.Context, t *ui.Theme, icon *ui.SVG, title, description string, end func()) {
	ui.Row(c).Gap(gapGroup).Padding(padCardY, padCardX).MinHeight(settingsRowHeight).Radius(cardRadius).
		Background(apexSurface).Border(1, apexLine).Children(func() {
		ui.Icon(c, icon).Size(rowIcon, rowIcon).TextColor(t.TextMuted)
		ui.Column(c).Grow(1).Shrink(1).Gap(gapLines).Children(func() {
			ui.Text(c, title).MaxLines(1)
			if description != "" {
				roleSmall.on(t, ui.Text(c, description)).TextColor(t.TextMuted).MaxLines(1).TruncateMiddle().Tooltip(description)
			}
		})
		if end != nil {
			ui.Row(c).Gap(gapControls).Children(end)
		}
	})
}

func valueText(c *ui.Context, s string, color ui.Color) {
	ui.Text(c, s).TextColor(color).MaxLines(1)
}

// statusValue is a value with a dot of its color, for what was found or not.
func statusValue(c *ui.Context, s string, color ui.Color) {
	ui.Row(c).Gap(gapControls).Children(func() {
		ui.Column(c).Size(statusDot, statusDot).Radius(statusDot / 2).Background(color)
		ui.Text(c, s).MaxLines(1)
	})
}

// advancedCard is an expander, as Windows 11's Settings has: the package kind and its source, which Auto and
// GitHub get right for most games, open below its header.
func (a *app) advancedCard(c *ui.Context, t *ui.Theme, g *game) {
	r := g.Scan
	ui.Column(c).Radius(cardRadius).Background(apexSurface).Border(1, apexLine).Clip().Children(func() {
		head := ui.ButtonBase(c.Key("advanced")).Padding(padCardY, padCardX).MinHeight(settingsRowHeight).Gap(gapGroup).Label(T("advanced"))
		if head.Hovered() {
			head.Background(apexSurface2)
		}
		head.OnClick(func() { a.advanced = !a.advanced }).Children(func() {
			ui.Icon(c, iconSliders).Size(rowIcon, rowIcon).TextColor(t.TextMuted)
			ui.Column(c).Grow(1).Shrink(1).Gap(gapLines).Children(func() {
				ui.Text(c, T("advanced"))
				roleSmall.on(t, ui.Text(c, T("advanced.desc"))).TextColor(t.TextMuted)
			})
			chevron := iconChevronD
			if a.advanced {
				chevron = iconChevronU
			}
			ui.Icon(c, chevron).Size(rowIcon, rowIcon).TextColor(t.TextMuted)
		})
		if !a.advanced {
			return
		}
		ui.Divider(c)
		ui.Column(c).Padding(padCardY, padCardX).PaddingStart(padCardX + rowIcon + gapGroup).Gap(gapInCard).Children(func() {
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
			if a.source == sourceLocal {
				msg, color := T("zip.none"), t.TextMuted
				switch {
				case a.localErr != "":
					msg, color = a.localErr, t.Danger
				case a.local != nil:
					msg, color = T("zip.info", filepath.Base(a.localZip), a.local.Version, a.local.Kind), t.Text
				}
				roleSmall.on(t, ui.Text(c, msg)).TextColor(color).PaddingStart(labelWidth + gapLabelField).MaxLines(2)
			}
		})
	})
}

// card is a group on a surface, with a small title, for the About dialog.
func card(c *ui.Context, t *ui.Theme, title string, body func()) ui.Element {
	return ui.Column(c).Padding(padCardY, padCardX).Gap(gapInCard).Radius(cardRadius).Background(apexSurface).Border(1, apexLine).Children(func() {
		roleCaption.on(t, ui.Text(c, title)).TextColor(t.TextMuted)
		body()
	})
}

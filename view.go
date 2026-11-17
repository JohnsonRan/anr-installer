package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/internal/detect"
)

func (a *app) view(c *ui.Context) {
	a.filter()
	t := c.Theme()
	ui.Column(c).Fill().Children(func() {
		a.titleBar(c, t)
		ui.Divider(c)
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Width(340).Padding(16).Gap(10).Children(func() { a.sidebar(c, t) })
			ui.Divider(c).Vertical()
			ui.Column(c).Grow(1).Shrink(1).Children(func() {
				ui.Scroll(c).Grow(1).Children(func() {
					ui.Column(c).Padding(16, 20).Gap(14).Children(func() { a.detail(c, t) })
				})
				ui.Column(c).Height(210).Padding(0, 20, 16, 20).Children(func() { a.logPanel(c, t) })
			})
		})
	})
}

func (a *app) sidebar(c *ui.Context, t *ui.Theme) {
	ui.SearchField(c, &a.query).Label(T("search")).Placeholder(T("search"))
	ui.Checkbox(c, &a.onlySupported, T("only.supported"))
	a.filter()
	ui.List(c, &a.list, len(a.visible)).Grow(1).Border(1, t.Border).Radius(8).Rows(func(r ui.ListRow) {
		g := a.games[a.visible[r.Index]]
		muted, text := t.TextMuted, t.Text
		if r.Selected() {
			muted, text = t.AccentText, t.AccentText
		}
		ui.Row(c).Padding(8, 10).Gap(10).Children(func() {
			gameIcon(c, t, g, 34)
			ui.Column(c).Grow(1).Shrink(1).Gap(4).Children(func() {
				ui.Text(c, g.Name).TextColor(text).MaxLines(1).Ellipsis("…")
				ui.Row(c).Gap(6).Children(func() {
					switch {
					case g.Scanning || g.Scan == nil:
						ui.Text(c, T("scanning")).FontSize(11).TextColor(muted)
					default:
						a.badges(c, t, g, r.Selected())
					}
				})
			})
		})
	}).Children(func() {
		if len(a.visible) == 0 {
			msg := T("no.games")
			switch {
			case a.loadingGames:
				msg = T("scanning.steam")
			case a.noSteam && len(a.games) == 0:
				msg = T("no.steam")
			}
			ui.Text(c, msg).TextColor(t.TextMuted).Padding(12)
		}
	})
	supported := 0
	for _, g := range a.games {
		if g.Scan != nil && g.Scan.Supported() {
			supported++
		}
	}
	ui.Text(c, T("games.count", len(a.games), supported)).FontSize(12).TextColor(t.TextMuted)
	ui.Row(c).Gap(8).Children(func() {
		ui.Button(c, T("add.folder")).Grow(1).OnClick(a.addFolder)
		ui.Button(c, T("refresh")).Disabled(a.loadingGames || a.busy).OnClick(a.scanSteam)
	})
}

func (a *app) badges(c *ui.Context, t *ui.Theme, g *game, selected bool) {
	r := g.Scan
	badge := func(s string, bg ui.Color) {
		b := ui.Badge(c, s).FontSize(11)
		if selected {
			b.Background(t.AccentText.Alpha(0.25)).TextColor(t.AccentText)
		} else if bg != (ui.Color{}) {
			b.Background(bg).TextColor(t.AccentText)
		}
	}
	if r.Supported() {
		badge(T("badge.supported"), t.Success)
	}
	switch r.Installed {
	case detect.KindFull, detect.KindAddon:
		s := T("badge." + string(r.Installed))
		if v := a.versionOf(r); v != "" {
			s += " " + strings.TrimPrefix(v, "V.")
		}
		badge(s, t.Accent)
	default:
		if r.UserOpti {
			badge(T("badge.opti"), ui.Color{})
		}
	}
	if g.Manual {
		badge(T("badge.manual"), ui.Color{})
	}
}

func (a *app) detail(c *ui.Context, t *ui.Theme) {
	g := a.selected()
	if g == nil {
		ui.Column(c).Height(200).Center().Children(func() {
			ui.Text(c, T("pick.game")).TextColor(t.TextMuted)
		})
		return
	}
	if g.Header != nil {
		// The store header as a banner, the name over its darkened bottom.
		ui.Column(c).FillWidth().Height(150).Radius(10).Clip().Justify(ui.End).Children(func() {
			ui.Image(c, g.Header).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(ui.Cover)
			ui.Row(c).Padding(28, 16, 12, 16).Gap(10).FillWidth().ZIndex(1).Gradient(ui.RGBA(0, 0, 0, 0), ui.RGBA(0, 0, 0, 0.75), 180).Children(func() {
				gameIcon(c, t, g, 32)
				ui.Text(c, g.Name).FontSize(20).Bold().TextColor(ui.RGB(255, 255, 255)).MaxLines(1).Ellipsis("…").Shrink(1)
			})
		})
	} else {
		ui.Row(c).Gap(12).Children(func() {
			gameIcon(c, t, g, 40)
			ui.Text(c, g.Name).FontSize(20).Bold().MaxLines(1).Ellipsis("…").Shrink(1)
		})
	}
	if g.Scan == nil {
		ui.Row(c).Gap(8).Children(func() {
			ui.Spinner(c)
			ui.Text(c, T("scanning")).TextColor(t.TextMuted)
		})
		return
	}
	r := g.Scan

	// Where the installer goes.
	section(c, t, T("exe.folder"), func() {
		ui.Row(c).Gap(8).Children(func() {
			ui.Text(c, r.ExeDir).Grow(1).MaxLines(2).Selectable().TruncateMiddle()
			ui.Button(c, T("change")).Disabled(a.busy || g.Scanning).OnClick(func() { a.changeExeDir(g) })
			ui.Button(c, T("open.folder")).OnClick(func() {
				if r.Exe != "" {
					mygo.Shell.ShowItemInFolder(r.Exe)
				} else {
					mygo.Shell.OpenPath(r.ExeDir)
				}
			})
		})
	})

	// What is there.
	section(c, t, T("status"), func() {
		if r.Supported() {
			names := map[string]bool{}
			for _, p := range r.DLSSG {
				names[filepath.Base(p)] = true
			}
			var list []string
			for n := range names {
				list = append(list, n)
			}
			line(c, t.Success, T("dlssg.yes", strings.Join(list, ", ")))
		} else {
			line(c, t.Warning, T("dlssg.no"))
		}
		switch r.Installed {
		case detect.KindFull:
			line(c, t.Accent, T("inst.full", r.Proxy))
			line(c, t.TextMuted, T("inst.version", a.installedVersion(r)))
		case detect.KindAddon:
			line(c, t.Accent, T("inst.addon"))
			line(c, t.TextMuted, T("inst.version", a.installedVersion(r)))
		default:
			line(c, t.TextMuted, T("inst.none"))
			if r.UserOpti {
				line(c, t.TextMuted, T("opti.user"))
			}
		}
		if r.NRRecord {
			line(c, t.Warning, T("nrrecord.on"))
		}
	})

	// What to put in.
	rec := r.Recommended()
	section(c, t, T("package"), func() {
		ui.Row(c).Children(func() {
			ui.Segmented(c, &a.kind, T("kind.auto", rec), "full", "addon").Label(T("package")).Disabled(a.busy)
		})
	})
	section(c, t, T("source"), func() {
		ui.Row(c).Gap(10).Children(func() {
			ui.Segmented(c, &a.source, T("source.github0"), T("source.local")).Label(T("source")).Disabled(a.busy)
			if a.source == sourceGitHub && len(a.versionLabels) > 0 {
				ui.Select(c, &a.versionSel, a.versionLabels).Label(T("version")).Disabled(a.busy).MinWidth(220)
				if rel := a.target(); rel != nil {
					ui.Button(c, T("release.notes")).OnClick(func() { mygo.Shell.OpenExternal(rel.URL) })
				}
			}
			if a.source == sourceLocal {
				ui.Button(c, T("choose.zip")).Disabled(a.busy).OnClick(a.chooseZip)
				switch {
				case a.localErr != "":
					ui.Text(c, a.localErr).TextColor(t.Danger).MaxLines(2).Grow(1)
				case a.local != nil:
					ui.Text(c, T("zip.info", filepath.Base(a.localZip), a.local.Kind)).TextColor(t.TextMuted).MaxLines(1).TruncateMiddle().Grow(1)
				default:
					ui.Text(c, T("zip.none")).TextColor(t.TextMuted)
				}
			}
		})
	})

	k := a.chosenKind(g)
	why := a.blocked(g, k)
	plan := a.plan(r)
	if why == "" && plan.uninstallsFirst() {
		why = a.uninstallBlocked(r.Installed)
	}
	label := a.planLabel(plan)
	ui.Row(c).Gap(10).Children(func() {
		ui.PrimaryButton(c, label).Disabled(a.busy || why != "" || g.Scanning).OnClick(func() { a.install(g, k) })
		if r.Installed != detect.KindNone {
			ui.Button(c, T("uninstall")).Disabled(a.busy || a.uninstallBlocked(r.Installed) != "" || g.Scanning).OnClick(func() { a.uninstall(g) })
		}
		if a.busy {
			ui.Spinner(c).Label(T("busy"))
		}
	})
	if why != "" && !g.Scanning {
		ui.Text(c, why).TextColor(t.Warning)
	}
	var notes []string
	switch {
	case plan.uninstallsFirst():
		latest := ""
		if a.rel != nil {
			latest = a.rel.Version
		}
		notes = append(notes, T("note.downgrade", latest, a.targetVersion()))
	case r.Installed == detect.KindFull && k == detect.KindFull && r.OptiINI:
		notes = append(notes, T("note.ini.reset"))
	}
	if r.NRRecord {
		notes = append(notes, T("note.nrrecord"))
	}
	notes = append(notes, T("note.close"))
	if runtime.GOOS != "windows" {
		notes = append(notes, T("note.proton"))
	}
	for _, n := range notes {
		ui.Text(c, "• "+n).FontSize(12).TextColor(t.TextMuted)
	}
}

func (a *app) logPanel(c *ui.Context, t *ui.Theme) {
	ui.Column(c).Grow(1).Gap(6).Children(func() {
		ui.Row(c).Gap(8).Children(func() {
			ui.Text(c, T("log")).Bold().Grow(1)
			if a.progress >= 0 {
				ui.Progress(c, a.progress).Width(180).Label(a.status)
			}
			ui.Button(c, T("clear")).Disabled(a.busy || len(a.logLines) == 0).OnClick(func() { a.logLines = nil })
		})
		if a.status != "" {
			ui.Text(c, a.status).FontSize(12).TextColor(t.TextMuted).MaxLines(2)
		}
		ui.List(c, &a.logList, len(a.logLines), func(i int) {
			ui.Text(c, a.logLines[i]).Font("monospace").FontSize(12).Selectable().Padding(1, 10)
		}).Grow(1).Padding(6, 0).Border(1, t.Border).Radius(8).Background(t.Surface)
	})
}

func section(c *ui.Context, t *ui.Theme, title string, body func()) {
	ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, title).FontSize(12).Bold().TextColor(t.TextMuted)
		body()
	})
}

func line(c *ui.Context, dot ui.Color, s string) {
	ui.Row(c).Gap(8).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Size(8, 8).Radius(4).Background(dot).Margin(6, 0, 0, 0)
		ui.Text(c, s).Grow(1).Shrink(1)
	})
}

var _ = fmt.Sprintf

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
		ui.Text(c, initial).FontSize(size * 0.45).Bold().TextColor(t.TextMuted)
	})
}

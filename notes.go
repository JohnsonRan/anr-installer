package main

import (
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"

	"anr-installer/internal/pkg"
)

// showNotes opens the release notes of rel in the window.
func (a *app) showNotes(rel *pkg.Release) { a.notesRel = rel }

// releaseNotes is the dialog of a release's notes, as GitHub has them, in Markdown, with a link to the
// release's page.
func (a *app) releaseNotes(c *ui.Context, t *ui.Theme) {
	rel := a.notesRel
	if rel == nil {
		return
	}
	open := true
	ui.Modal(c, &open, func() {
		ui.Column(c).Width(notesWidth).Gap(gapSection).Children(func() {
			ui.Column(c).Gap(gapLines).Children(func() {
				roleHeading.on(t, ui.Text(c, T("notes.title", rel.Version)))
				if !rel.Published.IsZero() {
					roleSmall.on(t, ui.Text(c, T("notes.date", rel.Published.Local().Format("2006-01-02")))).TextColor(t.TextMuted)
				}
			})
			ui.Scroll(c.Key(rel.Tag)).MaxHeight(notesHeight).Children(func() {
				if strings.TrimSpace(rel.Notes) == "" {
					ui.Text(c, T("notes.none")).TextColor(t.TextMuted)
					return
				}
				markdown.View(c, rel.Notes, &a.md)
			})
			latest := a.rel != nil && rel == a.rel
			games, anticheat := a.upgradable()
			if latest && len(anticheat) > 0 {
				roleSmall.on(t, ui.Text(c, T("notes.anticheat", strings.Join(anticheat, ", ")))).TextColor(t.TextMuted)
			}
			ui.Row(c).Gap(gapControls).Children(func() {
				ui.Button(c, T("notes.github")).OnClick(func() { mygo.Shell.OpenExternal(rel.URL) })
				ui.Spacer(c)
				if latest && len(games) > 0 {
					ui.PrimaryButton(c, T("notes.updateAll", len(games))).Disabled(a.busy).OnClick(func() { a.notesRel = nil; a.upgradeAll() })
					ui.Button(c, T("about.close")).OnClick(func() { a.notesRel = nil })
				} else {
					ui.PrimaryButton(c, T("about.close")).OnClick(func() { a.notesRel = nil })
				}
			})
		})
	})
	if !open {
		a.notesRel = nil
	}
}

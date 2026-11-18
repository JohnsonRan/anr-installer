package gui

import (
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"

	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
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
		ui.Column(c).Width(style.NotesWidth).Gap(style.GapSection).Children(func() {
			ui.Column(c).Gap(style.GapLines).Children(func() {
				style.RoleHeading.On(t, ui.Text(c, i18n.T("notes.title", rel.Version)))
				if !rel.Published.IsZero() {
					style.RoleSmall.On(t, ui.Text(c, i18n.T("notes.date", rel.Published.Local().Format("2006-01-02")))).TextColor(t.TextMuted)
				}
			})
			ui.Scroll(c.Key(rel.Tag)).MaxHeight(style.NotesHeight).Children(func() {
				if strings.TrimSpace(rel.Notes) == "" {
					ui.Text(c, i18n.T("notes.none")).TextColor(t.TextMuted)
					return
				}
				markdown.View(c, rel.Notes, &a.md)
			})
			latest := a.rel != nil && rel == a.rel
			games, anticheat := a.upgradable()
			if latest && len(anticheat) > 0 {
				style.RoleSmall.On(t, ui.Text(c, i18n.T("notes.anticheat", strings.Join(anticheat, ", ")))).TextColor(t.TextMuted)
			}
			ui.Row(c).Gap(style.GapControls).Children(func() {
				ui.Button(c, i18n.T("notes.github")).OnClick(func() { mygo.Shell.OpenExternal(rel.URL) })
				ui.Spacer(c)
				if latest && len(games) > 0 {
					ui.PrimaryButton(c, i18n.T("notes.updateAll", len(games))).Disabled(a.busy).OnClick(func() { a.notesRel = nil; a.upgradeAll() })
					ui.Button(c, i18n.T("about.close")).OnClick(func() { a.notesRel = nil })
				} else {
					ui.PrimaryButton(c, i18n.T("about.close")).OnClick(func() { a.notesRel = nil })
				}
			})
		})
	})
	if !open {
		a.notesRel = nil
	}
}

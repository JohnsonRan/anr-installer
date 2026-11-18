package gui

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
	"anr-installer/internal/detect"
	"anr-installer/internal/ini"
	"anr-installer/internal/pkg"
)

// pinned are the OptiScaler.ini values ANR FG's full package sets so that OptiScaler loads ANR FG as its frame
// generation: kept from the user's file, ANR FG would not run (nr-record sets FGNvngxReplacement to itself, which
// its own installer undoes).
var pinned = []ini.Key{{Section: "FrameGen", Name: "FGInput"}, {Section: "FrameGen", Name: "FGNvngxReplacement"}}

func isPinned(k ini.Key) bool {
	for _, p := range pinned {
		if strings.EqualFold(p.Section, k.Section) && strings.EqualFold(p.Name, k.Name) {
			return true
		}
	}
	return false
}

// keepSettings writes the user's OptiScaler.ini values (mine, the file before the install) into the one the
// install put in exeDir, merged three ways against the file of the release the user's came with (base; nil when
// the version is unknown, which makes every value both changed). Values both changed are asked about in the
// dialog, or kept in a batch. It runs in the background, after the install of version.
func (a *app) keepSettings(ctx context.Context, r *reporter, exeDir string, mine []byte, base *pkg.Release, version string, batch bool) error {
	path := filepath.Join(exeDir, "OptiScaler.ini")
	theirs, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var baseFile *ini.File
	if base != nil {
		if data, err := pkg.PackageFile(ctx, base, detect.KindFull, "game/OptiScaler.ini"); err == nil {
			baseFile = ini.Parse(data)
		}
	}
	res := ini.Merge(baseFile, ini.Parse(mine), ini.Parse(theirs), isPinned)
	if len(res.Kept)+len(res.Conflicts)+len(res.Pinned) == 0 {
		return nil
	}
	keepMine := make([]bool, len(res.Conflicts))
	if len(res.Conflicts) > 0 {
		if batch {
			setAll(keepMine, true)
		} else {
			answer := make(chan []bool, 1)
			a.update(func() { a.merge = newMerge(version, res, func(choices []bool) { answer <- choices }) })
			select {
			case keepMine = <-answer:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	values := map[ini.Key]string{}
	var kept, taken []string
	conflictKept := false
	for _, c := range res.Kept {
		values[c.Key] = c.Mine
		kept = append(kept, c.Key.Name)
	}
	for i, c := range res.Conflicts {
		if !keepMine[i] {
			taken = append(taken, c.Key.Name)
			continue
		}
		values[c.Key] = c.Mine
		kept = append(kept, i18n.T("merge.log.kept.conflict", c.Key.Name, c.Theirs))
		conflictKept = true
	}
	if len(values) > 0 {
		if err := os.WriteFile(path, ini.Apply(theirs, values), 0o644); err != nil {
			return err
		}
	}
	if len(kept) > 0 {
		text := i18n.T("merge.log.kept", strings.Join(kept, ", "))
		if conflictKept && batch {
			r.warn(text) // kept without asking: the user may want the new values
		} else {
			r.say("%s", text)
		}
	}
	if len(taken) > 0 {
		r.say("%s", i18n.T("merge.log.new", strings.Join(taken, ", ")))
	}
	if len(res.Pinned) > 0 {
		r.say("%s", i18n.T("merge.log.pinned", keyNames(res.Pinned)))
	}
	return nil
}

// mergeState is the dialog that asks which OptiScaler.ini values to keep after an install replaced the file: the
// settings the user and the new version both changed, one choice each.
type mergeState struct {
	version string     // the version installed, which brought the new file
	result  ini.Result // what the merge found
	mine    []bool     // per conflict: keep the user's value
	open    bool
	answer  func([]bool) // gets the choices once, as the dialog closes
}

// newMerge opens the dialog for result, with the user's values chosen. answer gets the choices as it closes.
func newMerge(version string, result ini.Result, answer func([]bool)) *mergeState {
	m := &mergeState{version: version, result: result, mine: make([]bool, len(result.Conflicts)), open: true, answer: answer}
	for i := range m.mine {
		m.mine[i] = true
	}
	return m
}

// close answers with the choices made, once.
func (m *mergeState) close() {
	if m.answer != nil {
		m.answer(m.mine)
		m.answer = nil
	}
	m.open = false
}

// keyNames lists the names of the keys of changes, for a line of text.
func keyNames(changes []ini.Change) string {
	names := make([]string, len(changes))
	for i, c := range changes {
		names[i] = c.Key.Name
	}
	return strings.Join(names, ", ")
}

// mergeDialog asks which value each conflicting setting keeps. Closing it otherwise, as Escape does, saves the
// choices as they are: the user's values unless changed.
func (a *app) mergeDialog(c *ui.Context, t *ui.Theme) {
	m := a.merge
	if m == nil {
		return
	}
	ui.Modal(c, &m.open, func() {
		ui.Column(c).Width(style.NotesWidth).Gap(style.GapSection).Children(func() {
			ui.Column(c).Gap(style.GapInCard).Children(func() {
				style.RoleHeading.On(t, ui.Text(c, i18n.T("merge.title")))
				ui.Text(c, i18n.T("merge.detail", m.version)).TextColor(t.TextMuted)
			})
			ui.Scroll(c).MaxHeight(style.NotesHeight).Children(func() {
				ui.Column(c).Gap(style.GapCardRows).Children(func() {
					for i, cf := range m.result.Conflicts {
						choice := 1
						if m.mine[i] {
							choice = 0
						}
						mine, theirs := i18n.T("merge.mine", cf.Mine), i18n.T("merge.new", cf.Theirs)
						ui.Row(c.Key(cf.Key)).Padding(style.PadCardY, style.PadCardX).Gap(style.GapGroup).Wrap().AlignItems(ui.Center).
							Radius(style.CardRadius).Background(style.ApexSurface).Border(1, style.ApexLine).Children(func() {
							ui.Column(c).Grow(1).Shrink(1).Gap(style.GapLines).Children(func() {
								style.RoleBodyStrong.On(t, ui.Text(c, cf.Key.Name)).MaxLines(1).Ellipsis("…").Tooltip(cf.Key.Name)
								style.RoleSmall.On(t, ui.Text(c, "["+cf.Key.Section+"]")).TextColor(t.TextMuted)
							})
							if style.Segmented(c, t, &choice, mine, theirs).Label(cf.Key.Name).Changed() {
								m.mine[i] = choice == 0
							}
						})
					}
					if len(m.result.Pinned) > 0 {
						style.Section(c, t, i18n.T("merge.pinned.title"), nil, func() {
							for _, p := range m.result.Pinned {
								style.InfoBar(c, t, style.ApexRoyal, style.IconInfo, i18n.T("merge.pinned", p.Key.Name, p.Theirs, p.Mine), nil)
							}
						})
					}
				})
			})
			if len(m.result.Kept) > 0 {
				line := i18n.T("merge.kept", keyNames(m.result.Kept))
				style.RoleSmall.On(t, ui.Text(c, line)).TextColor(t.TextMuted).MaxLines(2).Ellipsis("…").Tooltip(line)
			}
			ui.Row(c).Gap(style.GapControls).Wrap().Children(func() {
				ui.Button(c, i18n.T("merge.all.mine")).OnClick(func() { setAll(m.mine, true) })
				ui.Button(c, i18n.T("merge.all.new")).OnClick(func() { setAll(m.mine, false) })
				ui.Column(c).Grow(1)
				ui.PrimaryButton(c, i18n.T("merge.apply")).OnClick(m.close)
			})
		})
	})
	if !m.open {
		m.close()
		a.merge = nil
	}
}

func setAll(v []bool, to bool) {
	for i := range v {
		v[i] = to
	}
}

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// compareStills is the lossless images of the frames around the moment the videos are at.
func (a *app) compareStills(c *ui.Context, t *ui.Theme, s *compareState) {
	if s.stillErr != "" {
		infoBar(c, t, t.Danger, iconWarning, T("cmp.still.error", s.stillErr), nil)
	}
	if s.stills == nil || s.real != s.realFrame() {
		// The stills come from the DLLs in the game now, with FFmpeg: what stops a comparison stops them too.
		why := a.compareBlocked(s.game)
		infoBar(c, t, apexRoyal, iconInfo, T("cmp.still.none"), func() {
			ui.Button(c, T("cmp.still.make", s.realFrame())).Disabled(why != "" || a.busy || s.loading).OnClick(func() { a.exportStills(s) })
		})
		if why != "" {
			infoBar(c, t, t.Warning, iconWarning, why, nil)
		}
		if s.stills == nil {
			return
		}
	}
	if s.stillA == nil || s.stillB == nil {
		ui.Row(c).Gap(gapControls).Children(func() {
			ui.Spinner(c).Label(T("cmp.still.loading"))
			ui.Text(c, T("cmp.still.loading")).TextColor(t.TextMuted)
		})
		return
	}
	w, h := s.stillA.Size()
	s.pictures(c, t, float32(w)/float32(h), func(side int) ui.Element {
		if side == 0 {
			return ui.Image(c, s.stillA).Fit(ui.Contain)
		}
		return ui.Image(c, s.stillB).Fit(ui.Contain)
	})
	ui.Row(c).Gap(gapControls).Children(func() {
		ui.Button(c, T("cmp.still.prev")).Disabled(s.still == 0 || s.loading).OnClick(func() { a.showStill(s, s.still-1) })
		ui.Button(c, T("cmp.still.next")).Disabled(s.still >= len(s.stills)-1 || s.loading).OnClick(func() { a.showStill(s, s.still+1) })
		name := s.stills[s.still].Name
		label := T("cmp.still.name", s.still+1, len(s.stills), name)
		roleSmall.on(t, ui.Text(c, label)).TextColor(t.TextMuted).MaxLines(1).TruncateMiddle().Tooltip(label).Grow(1).Shrink(1)
		ui.Button(c, T("open.folder")).OnClick(func() { mygo.Shell.ShowItemInFolder(s.stills[s.still].A) })
	})
}

// exportStills replays the frames around the moment shown through both sides and shows them as stills.
func (a *app) exportStills(s *compareState) {
	if s.playing {
		s.toggle()
	}
	shown, cmp := s.frame(), *s.cmp
	real := shown.Real
	s.stillErr = ""
	a.recordWork(s.game, gameTask(s.game, T("cmp.still.make", real)), func(ctx context.Context, r *reporter, tools func() (string, error)) error {
		dir, err := tools()
		if err != nil {
			return err
		}
		cmp.Tools = dir
		stills, err := cmp.Stills(ctx, real, func(label string) { r.say("%s", T("cmp.still.running", label)) }, r.out)
		if err != nil {
			return err
		}
		r.say("%s", T("cmp.still.done", len(stills), filepath.Dir(stills[0].A)))
		a.win.Update(func() {
			if a.cmp != s {
				return // the page was left meanwhile
			}
			s.real, s.stills = real, stills
			// The stills lead up to the recorded frame, the last one: start at the frame the video is paused at.
			at := len(stills) - 1
			for i, still := range stills {
				if still.Frame.Is(shown) {
					at = i
				}
			}
			a.showStill(s, at)
		})
		return nil
	})
}

// showStill shows still i, reading its two images off the main thread.
func (a *app) showStill(s *compareState, i int) {
	s.still, s.loading, s.stillA, s.stillB = i, true, nil, nil
	still := s.stills[i]
	go func() {
		dataA, errA := os.ReadFile(still.A)
		dataB, errB := os.ReadFile(still.B)
		a.win.Update(func() {
			if a.cmp != s || s.still != i {
				return
			}
			s.loading = false
			if errA == nil && errB == nil {
				s.stillA, errA = ui.DecodeBitmap(dataA)
				s.stillB, errB = ui.DecodeBitmap(dataB)
			}
			if err := errors.Join(errA, errB); err != nil {
				// Stills that cannot be shown are removed, so that exporting the frame makes them again.
				s.stills, s.stillA, s.stillB, s.stillErr = nil, nil, nil, err.Error()
				s.cmp.RemoveStills(s.real)
			}
		})
	}()
}

package compare

import (
	"errors"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
	"anr-installer/internal/record"
)

// stillsView is the lossless images of the frames around the moment the videos are at.
func (s *State) stillsView(c *ui.Context, t *ui.Theme, act Actions) {
	if s.stillErr != "" {
		style.InfoBar(c, t, t.Danger, style.IconWarning, i18n.T("cmp.still.error", s.stillErr), nil)
	}
	if s.stills == nil || s.real != s.realFrame() {
		// The stills come from the DLLs in the game now, with FFmpeg: what stops a comparison stops them too.
		why := act.Blocked()
		style.InfoBar(c, t, style.ApexRoyal, style.IconInfo, i18n.T("cmp.still.none"), func() {
			ui.Button(c, i18n.T("cmp.still.make", s.realFrame())).Disabled(why != "" || act.Busy || s.loading).OnClick(act.Export)
		})
		if why != "" {
			style.InfoBar(c, t, t.Warning, style.IconWarning, why, nil)
		}
		if s.stills == nil {
			return
		}
	}
	if s.stillA == nil || s.stillB == nil {
		ui.Row(c).Gap(style.GapControls).Children(func() {
			ui.Spinner(c).Label(i18n.T("cmp.still.loading"))
			ui.Text(c, i18n.T("cmp.still.loading")).TextColor(t.TextMuted)
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
	ui.Row(c).Gap(style.GapControls).Children(func() {
		ui.Button(c, i18n.T("cmp.still.prev")).Disabled(s.still == 0 || s.loading).OnClick(func() { s.show(s.still - 1) })
		ui.Button(c, i18n.T("cmp.still.next")).Disabled(s.still >= len(s.stills)-1 || s.loading).OnClick(func() { s.show(s.still + 1) })
		name := s.stills[s.still].Name
		label := i18n.T("cmp.still.name", s.still+1, len(s.stills), name)
		style.RoleSmall.On(t, ui.Text(c, label)).TextColor(t.TextMuted).MaxLines(1).TruncateMiddle().Tooltip(label).Grow(1).Shrink(1)
		ui.Button(c, i18n.T("open.folder")).OnClick(func() { mygo.Shell.ShowItemInFolder(s.stills[s.still].A) })
	})
}

// SetStills shows the stills made of real frame real, starting at the one of the frame shown when they were
// asked for. It drops them when the page was left meanwhile.
func (s *State) SetStills(real int, stills []record.Still, shown record.Frame) {
	if s.closed {
		return
	}
	s.real, s.stills = real, stills
	// The stills lead up to the recorded frame, the last one: start at the frame the video is paused at.
	at := len(stills) - 1
	for i, still := range stills {
		if still.Frame.Is(shown) {
			at = i
		}
	}
	s.show(at)
}

// show shows still i, reading its two images off the main thread.
func (s *State) show(i int) {
	s.still, s.loading, s.stillA, s.stillB = i, true, nil, nil
	still := s.stills[i]
	go func() {
		dataA, errA := os.ReadFile(still.A)
		dataB, errB := os.ReadFile(still.B)
		s.update(func() {
			if s.closed || s.still != i {
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

// StillsJob pauses the videos and returns what making the stills of the frame shown takes: the frame, and a copy
// of the comparison for the job to change.
func (s *State) StillsJob() (shown record.Frame, cmp record.Comparison) {
	if s.playing {
		s.toggle()
	}
	s.stillErr = ""
	return s.frame(), *s.cmp
}

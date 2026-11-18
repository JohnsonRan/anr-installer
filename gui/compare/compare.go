// Package compare is the comparison page: a recording replayed through two frame generation DLLs, shown as two
// videos played on one clock and, for a frame chosen in them, as lossless stills.
package compare

import (
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/video"

	"anr-installer/gui/i18n"
	"anr-installer/gui/style"
	"anr-installer/internal/record"
)

// How the comparison page shows its two pictures.
const (
	cmpDivider = iota // A left of a divider, B right of it, as one picture
	cmpSide           // A beside B
	cmpOnlyA
	cmpOnlyB
)

// What the comparison page shows.
const (
	cmpVideo  = iota // the two videos, played together
	cmpStills        // lossless images of the frames around one moment
)

// cmpRates are the speeds the videos play at.
var cmpRates = []float64{0.25, 0.5, 1}

// State is the comparison page: a recording replayed through two frame generation DLLs, as two videos
// played on one clock and, for a frame chosen in them, as lossless stills.
type State struct {
	cmp    *record.Comparison
	update func(func()) // runs a function on the main thread
	closed bool         // the page was left: work that ends after drops its result

	a, b   *ui.VideoPlayer // nil when the videos could not be opened
	err    string
	replay *record.Replay // what the videos' frames are; nil when it cannot be read
	fps    float64

	mode, layout, rate int
	divider            float64 // where the divider is, in percent of the picture's width
	pos                float64 // the time shown, in seconds
	playing            bool

	real     int            // the recorded real frame the stills are of
	stills   []record.Still // nil until exported
	still    int
	stillA   *ui.Bitmap // the still shown
	stillB   *ui.Bitmap
	loading  bool
	stillErr string // why the stills could not be shown, or ""
}

// Actions is what the comparison page does through the app.
type Actions struct {
	Subject string               // the game's name, which the line under the title starts with
	Busy    bool                 // a task runs, so none can start
	Blocked func() string        // why the stills cannot be made now, or ""
	Back    func()               // leaves the page
	Viewer  func(args ...string) // opens nr-record's viewer with args
	Export  func()               // makes the stills of the frame shown
}

// New is the comparison of cmp without its videos, which Open adds: it shows the stills only.
// update runs a function on the main thread.
func New(cmp *record.Comparison, update func(func())) *State {
	s := &State{cmp: cmp, update: update, mode: cmpStills, divider: 50, rate: len(cmpRates) - 1, fps: 60}
	if r, err := cmp.Replay(); err == nil {
		s.replay = r
		if r.FPS > 0 {
			s.fps = r.FPS
		}
	}
	return s
}

// Open is the comparison of cmp with its two videos, opened paused at their start.
func Open(cmp *record.Comparison, update func(func())) *State {
	s := New(cmp, update)
	s.mode = cmpVideo
	pathA, pathB := s.cmp.Videos()
	var err error
	if s.a, err = video.Open(pathA); err == nil {
		s.b, err = video.Open(pathB)
	}
	if err != nil {
		s.closeVideos()
		s.a, s.b, s.err = nil, nil, err.Error()
	} else {
		s.a.Pause()
		s.b.Pause()
	}
	return s
}

// Close stops the videos as the page is left.
func (s *State) Close() {
	s.closed = true
	s.closeVideos()
}

func (s *State) closeVideos() {
	for _, p := range []*ui.VideoPlayer{s.a, s.b} {
		if p != nil {
			p.Close()
		}
	}
}

// duration is the length of the shorter video, in seconds.
func (s *State) duration() float64 {
	if s.a == nil {
		return 0
	}
	return min(s.a.Duration(), s.b.Duration()).Seconds()
}

// seek shows both videos at sec.
func (s *State) seek(sec float64) {
	if s.a == nil {
		return
	}
	s.pos = min(max(sec, 0), s.duration())
	d := time.Duration(s.pos * float64(time.Second))
	s.a.Seek(d)
	s.b.Seek(d)
}

// toggle plays or pauses both videos. Paused, B goes to the frame A stopped at.
func (s *State) toggle() {
	if s.a == nil {
		return
	}
	if s.playing {
		s.playing = false
		s.a.Pause()
		s.b.Pause()
		s.seek(s.a.Position().Seconds())
		return
	}
	if s.pos >= s.duration() {
		s.seek(0)
	}
	s.playing = true
	s.a.Play()
	s.b.Play()
}

// step pauses and moves both videos by n frames.
func (s *State) step(n int) {
	if s.playing {
		s.toggle()
	}
	s.seek((math.Round(s.pos*s.fps) + float64(n)) / s.fps)
}

// follow takes the time from the videos while they play, and stops at their end.
func (s *State) follow() {
	if s.a == nil || !s.playing {
		return
	}
	s.pos = s.a.Position().Seconds()
	if s.a.Ended() || s.b.Ended() {
		s.playing, s.pos = false, s.duration() // at the end, so that playing again starts over
		s.a.Pause()
		s.b.Pause()
	}
}

// frame is the output frame shown, as the videos' replay knows it.
func (s *State) frame() record.Frame {
	n := int(math.Round(s.pos * s.fps))
	if s.replay == nil || len(s.replay.Frames) == 0 {
		return record.Frame{Kind: "real", Real: n, A: n, B: n}
	}
	return s.replay.Frames[min(max(n, 0), len(s.replay.Frames)-1)]
}

// realFrame is the recorded real frame behind the output frame shown.
func (s *State) realFrame() int { return s.frame().Real }

func clock(sec float64) string {
	d := time.Duration(sec * float64(time.Second))
	return fmt.Sprintf("%02d:%02d.%03d", int(d.Minutes()), int(d.Seconds())%60, d.Milliseconds()%1000)
}

// Page is the page of the comparison, in place of a game's page: a title row with the way back, what to show
// and how, the picture, and the controls of the videos or the stills under it.
func (s *State) Page(c *ui.Context, t *ui.Theme, act Actions) {
	if s.a != nil && s.err == "" {
		for _, p := range []*ui.VideoPlayer{s.a, s.b} {
			if err := p.Err(); err != nil {
				s.err = err.Error() // a video the system cannot decode fails after it opened
			}
		}
	}
	s.follow()
	// The keys work wherever the focus is, and with none: the page opens from a button that is gone by then. A
	// focused control that takes the key itself, as a button takes Space, comes first.
	c.OnWindowShortcut(0, ui.KeySpace, s.toggle)
	c.OnWindowShortcut(0, ui.KeyLeft, func() { s.step(-1) })
	c.OnWindowShortcut(0, ui.KeyRight, func() { s.step(1) })
	c.OnWindowShortcut(0, ui.KeyEscape, act.Back)
	ui.Column(c).Padding(style.PadPage, style.PadPageX).Gap(style.GapSection).Children(func() {
		line := act.Subject + " · " + s.cmp.Recording.Time.Format("2006-01-02 15:04:05")
		style.PageHeader(c, t, act.Back, i18n.T("cmp.title", s.cmp.A.Label, s.cmp.B.Label), line, func() {
			ui.Button(c, i18n.T("cmp.viewer")).Disabled(act.Busy).OnClick(func() { act.Viewer("--compare", "--project", s.cmp.Project()) })
		})
		if s.err != "" {
			style.InfoBar(c, t, t.Danger, style.IconWarning, i18n.T("cmp.error", s.err), nil)
			return
		}
		ui.Row(c).Gap(style.GapGroup).Wrap().Children(func() {
			if s.a != nil { // a comparison without videos shows the stills only
				ui.Segmented(c, &s.mode, i18n.T("cmp.video"), i18n.T("cmp.stills")).Label(i18n.T("cmp.mode"))
			}
			ui.Column(c).Grow(1)
			ui.Segmented(c, &s.layout, i18n.T("cmp.divider"), i18n.T("cmp.side"), s.cmp.A.Label, s.cmp.B.Label).Label(i18n.T("cmp.layout"))
		})
		if s.mode == cmpVideo {
			s.videos(c, t)
		} else {
			s.stillsView(c, t, act)
		}
	})
}

// videos is the two videos and their transport.
func (s *State) videos(c *ui.Context, t *ui.Theme) {
	w, h := s.a.Size()
	if !s.a.Ready() || !s.b.Ready() || h == 0 {
		w, h = 16, 9 // until the first pictures come: the room of a picture, with nothing moving after
	}
	s.pictures(c, t, float32(w)/float32(h), func(side int) ui.Element {
		if side == 0 {
			return ui.Video(c, s.a).Fit(ui.Contain)
		}
		return ui.Video(c, s.b).Fit(ui.Contain)
	})
	pos := s.pos
	if ui.Slider(c, &pos, 0, max(s.duration(), 1/s.fps)).Label(i18n.T("cmp.time")).Changed() {
		s.seek(pos)
	}
	// The speed goes to the row's end, and onto a line of its own where the window is too narrow for it.
	ui.Row(c).Wrap().AlignItems(ui.Center).Gap(style.GapControls).Children(func() {
		label := i18n.T("cmp.play")
		if s.playing {
			label = i18n.T("cmp.pause")
		}
		ui.PrimaryButton(c, label).OnClick(s.toggle)
		ui.Button(c, i18n.T("cmp.prev")).OnClick(func() { s.step(-1) })
		ui.Button(c, i18n.T("cmp.next")).OnClick(func() { s.step(1) })
		style.RoleSmall.On(t, ui.Text(c, clock(s.pos)+" / "+clock(s.duration()))).TextColor(t.TextMuted).NoWrap()
		labels := make([]string, len(cmpRates))
		for i, r := range cmpRates {
			labels[i] = fmt.Sprintf("%g×", r)
		}
		if ui.Segmented(c, &s.rate, labels...).MarginStart(ui.Auto).Label(i18n.T("cmp.rate")).Changed() {
			s.a.SetRate(cmpRates[s.rate])
			s.b.SetRate(cmpRates[s.rate])
		}
	})
	style.InfoBar(c, t, style.ApexRoyal, style.IconInfo, i18n.T("cmp.note"), nil)
}

// pictures lays out the two pictures, each of the ratio given, by the layout chosen. With the divider, A lies
// over B, clipped at the divider, so that the two read as one picture.
func (s *State) pictures(c *ui.Context, t *ui.Theme, ratio float32, picture func(side int) ui.Element) {
	frame := func(fill func()) {
		ui.Column(c).Grow(1).Basis(0).AspectRatio(ratio).Radius(style.CardRadius).Clip().Background(style.ApexBG).Border(1, style.ApexLine).Children(fill)
	}
	whole := func(e ui.Element) ui.Element { return e.Absolute().Top(0).Left(0).Right(0).Bottom(0) }
	switch s.layout {
	case cmpSide:
		ui.Row(c).Gap(style.GapCards).Children(func() {
			frame(func() { whole(picture(0)) })
			frame(func() { whole(picture(1)) })
		})
	case cmpOnlyA:
		ui.Row(c).Children(func() { frame(func() { whole(picture(0)) }) })
	case cmpOnlyB:
		ui.Row(c).Children(func() { frame(func() { whole(picture(1)) }) })
	default:
		side := func(label string) ui.Element {
			return ui.Row(c).Absolute().Top(style.Space2).Padding(style.Space1, style.Space2).Radius(style.NoticeRadius).Background(style.ApexBG.Alpha(0.7)).Children(func() {
				style.RoleCaption.On(t, ui.Text(c, label))
			})
		}
		ui.Row(c).Children(func() {
			frame(func() {
				whole(picture(1))
				whole(picture(0)).ClipInsetPercent(0, float32(100-s.divider), 0, 0, 0)
				side(s.cmp.A.Label).Left(style.Space2)
				side(s.cmp.B.Label).Right(style.Space2)
				at := float32(s.divider)
				ui.Column(c).Absolute().Top(0).Bottom(0).LeftPercent(at).Width(style.DividerLine).Background(style.ApexInk)
				ui.Column(c).Absolute().TopPercent(50).LeftPercent(at).Size(style.DividerHandle, style.DividerHandle).TranslatePercent(-50, -50).
					Radius(style.DividerHandle/2).Background(style.ApexInk).Border(1, style.ApexLine)
				// The whole picture is the divider's slider: a drag or a click anywhere on it moves the divider
				// there, and the arrow keys move it while it has the focus.
				whole(ui.SliderBase(c, &s.divider, 0, 100)).ZIndex(1).Cursor(ui.CursorResizeEW).Label(i18n.T("cmp.divider.at"))
			})
		})
	}
}

package main

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/egoist/mygo/ui"

	"anr-installer/internal/record"
)

func testBitmap(t *testing.T) *ui.Bitmap {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 32, 18))); err != nil {
		t.Fatal(err)
	}
	bm, err := ui.DecodeBitmap(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return bm
}

// comparePageApp is recordApp with the comparison page of Delta's recording open, without videos: the tests
// run with no media framework to play them.
func comparePageApp(t *testing.T) (*app, *compareState) {
	a := recordApp(t)
	a.row = 2 // Delta, the third of the supported games
	g := a.games[3]
	s := &compareState{game: g, cmp: a.comparison(g, a.rec.recordings[0]), divider: 50, fps: 120}
	s.replay = &record.Replay{FPS: 120, Multiplier: 2, Frames: []record.Frame{{Kind: "real", Real: 0}, {Kind: "generated", Real: 1}, {Kind: "real", Real: 1}}}
	a.cmp = s
	return a, s
}

func TestComparePageShowsStillsAndGoesBack(t *testing.T) {
	a, s := comparePageApp(t)
	s.mode, s.pos = cmpStills, 2.0/120 // the third output frame: real frame 1
	tt := ui.NewTester(a.view, 1040, 900)
	if !tt.HasText(T("cmp.title", "ANR FG", "DLSS-G")) || !tt.HasText(T("cmp.still.make", 1)) {
		t.Fatalf("the page should offer the stills of the frame shown: %q", tt.Texts())
	}
	if tt.HasText(T("rec.title")) {
		t.Error("the game's page shows under the comparison")
	}

	s.real = 1
	s.stills = []record.Still{{Name: "generated_f0-f1_t0.500"}, {Name: "real_f1-f1_t0.000"}}
	s.still, s.stillA, s.stillB = 1, testBitmap(t), testBitmap(t)
	tt.Frame()
	if !tt.HasText(T("cmp.still.name", 2, 2, "real_f1-f1_t0.000")) || tt.HasText(T("cmp.still.make", 1)) {
		t.Errorf("the stills of the frame shown should be up: %q", tt.Texts())
	}
	// Dragging on the picture moves the divider to where the pointer is.
	pic, ok := tt.Find(T("cmp.divider.at"))
	if !ok {
		t.Fatal("the divider is not on the picture")
	}
	y := pic.Y + pic.H/2
	tt.Press(pic.X+pic.W/2, y)
	tt.Move(pic.X+pic.W/4, y)
	tt.Release(pic.X+pic.W/4, y)
	if s.divider < 20 || s.divider > 30 {
		t.Errorf("dragged to a quarter of the picture, the divider is at %.0f%%", s.divider)
	}
	for layout := range cmpOnlyB + 1 {
		s.layout = layout
		tt.Frame()
	}

	if err := tt.Click(T("back")); err != nil {
		t.Fatal(err)
	}
	if a.cmp != nil || !tt.HasText(T("rec.title")) {
		t.Errorf("Back should return to the game's page: %q", tt.Texts())
	}
}

func TestComparePageClosesWhenAnotherGameIsChosen(t *testing.T) {
	a, s := comparePageApp(t)
	s.err = "no decoder"
	tt := ui.NewTester(a.view, 1040, 900)
	if !tt.HasText(T("cmp.error", "no decoder")) {
		t.Fatalf("a video that cannot play should say so: %q", tt.Texts())
	}
	if err := tt.Click("Alpha"); err != nil {
		t.Fatal(err)
	}
	if a.cmp != nil {
		t.Error("choosing another game should leave the comparison")
	}
}

func TestClockAndRealFrame(t *testing.T) {
	if got := clock(62.017); got != "01:02.017" {
		t.Errorf("clock %q", got)
	}
	_, s := comparePageApp(t)
	s.pos = 1.0 / 120
	if s.realFrame() != 1 {
		t.Errorf("the generated frame before real frame 1 belongs to it: %d", s.realFrame())
	}
	s.pos = 99
	if s.realFrame() != 1 {
		t.Errorf("past the end, the last frame: %d", s.realFrame())
	}
}

// The page opens from a button that is gone once it shows, so its keys must work with nothing focused.
func TestComparePageKeysWorkWithoutFocus(t *testing.T) {
	a, s := comparePageApp(t)
	s.mode = cmpStills // the test has no videos to show
	tt := ui.NewTester(a.view, 1040, 900)
	tt.Key(0, ui.KeyEscape)
	if a.cmp != nil {
		t.Error("Escape should leave the comparison with nothing focused")
	}
}

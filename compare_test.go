package main

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"time"

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

// testCompare is a comparison of a recording of Delta, without videos: the tests run with no media framework to
// play them. Its page is drawn with acts, whose back records that it was called.
func testCompare(t *testing.T) (s *compareState, acts *compareActions, back *bool) {
	setLang("en")
	rec := &record.Recording{Time: time.Date(2026, 11, 18, 15, 4, 5, 0, time.Local)}
	s = newCompare(&record.Comparison{Recording: rec, A: record.ANR("anr.dll"), B: record.DLSSG("dlssg.dll")}, func(fn func()) { fn() })
	s.fps = 120
	s.replay = &record.Replay{FPS: 120, Multiplier: 2, Frames: []record.Frame{{Kind: "real", Real: 0}, {Kind: "generated", Real: 1}, {Kind: "real", Real: 1}}}
	back = new(bool)
	acts = &compareActions{
		subject: "Delta",
		blocked: func() string { return "" },
		back:    func() { *back = true },
		viewer:  func(...string) {},
		export:  func() {},
	}
	return s, acts, back
}

func comparePageTester(s *compareState, acts *compareActions) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		th := apexTheme()
		c.SetTheme(th)
		ui.Scroll(c).Fill().Children(func() { s.page(c, th, *acts) })
	}, 1040, 900)
}

func TestComparePageShowsStills(t *testing.T) {
	s, acts, back := testCompare(t)
	s.pos = 2.0 / 120 // the third output frame: real frame 1
	tt := comparePageTester(s, acts)
	if !tt.HasText(T("cmp.title", "ANR FG", "DLSS-G")) || !tt.HasText(T("cmp.still.make", 1)) {
		t.Fatalf("the page should offer the stills of the frame shown: %q", tt.Texts())
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
	if !*back {
		t.Error("Back should leave the page")
	}
}

func TestComparePageShowsAVideoError(t *testing.T) {
	s, acts, _ := testCompare(t)
	s.err = "no decoder"
	tt := comparePageTester(s, acts)
	if !tt.HasText(T("cmp.error", "no decoder")) {
		t.Fatalf("a video that cannot play should say so: %q", tt.Texts())
	}
}

func TestClockAndRealFrame(t *testing.T) {
	if got := clock(62.017); got != "01:02.017" {
		t.Errorf("clock %q", got)
	}
	s, _, _ := testCompare(t)
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
	s, acts, back := testCompare(t)
	tt := comparePageTester(s, acts)
	tt.Key(0, ui.KeyEscape)
	if !*back {
		t.Error("Escape should leave the comparison with nothing focused")
	}
}

// Stills that come after the page was left are dropped.
func TestStillsAfterThePageWasLeft(t *testing.T) {
	s, _, _ := testCompare(t)
	s.close()
	s.setStills(1, []record.Still{{Name: "real_f1-f1_t0.000"}}, record.Frame{Kind: "real", Real: 1})
	if s.stills != nil {
		t.Error("the stills of a page that was left should be dropped")
	}
}

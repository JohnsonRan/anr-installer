package gui

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"anr-installer/gui/compare"
	"anr-installer/gui/i18n"
)

// comparePageApp is recordApp with the comparison page of Delta's recording open, without videos: the tests
// run with no media framework to play them.
func comparePageApp(t *testing.T) (*app, *compareView) {
	a := recordApp(t)
	a.row = 2 // Delta, the third of the supported games
	g := a.games[3]
	a.cmp = &compareView{game: g, State: compare.New(a.comparison(g, a.rec.recordings[0]), func(fn func()) { fn() })}
	return a, a.cmp
}

func TestComparePageGoesBack(t *testing.T) {
	a, _ := comparePageApp(t)
	tt := ui.NewTester(a.view, 1040, 900)
	if !tt.HasText(i18n.T("cmp.title", "ANR FG", "DLSS-G")) || tt.HasText(i18n.T("rec.title")) {
		t.Fatalf("the comparison should show in place of the game's page: %q", tt.Texts())
	}
	if err := tt.Click(i18n.T("back")); err != nil {
		t.Fatal(err)
	}
	if a.cmp != nil || !tt.HasText(i18n.T("rec.title")) {
		t.Errorf("Back should return to the game's page: %q", tt.Texts())
	}
}

func TestComparePageClosesWhenAnotherGameIsChosen(t *testing.T) {
	a, _ := comparePageApp(t)
	tt := ui.NewTester(a.view, 1040, 900)
	if err := tt.Click("Alpha"); err != nil {
		t.Fatal(err)
	}
	if a.cmp != nil {
		t.Error("choosing another game should leave the comparison")
	}
}

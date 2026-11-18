package style

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// textRole is a role of the type scale. Every text in the window takes one, here and nowhere else, so that a
// role has one size and weight everywhere (UI.md, "Layout, type and loading"). Sizes are multiples of the
// theme's FontSize, 14 DIP on Windows, so that they follow the system's text size setting.
type textRole struct {
	scale  float32 // of the theme's FontSize
	weight int
	line   float32 // line height as a multiple of the size; 0 keeps the theme's
}

var (
	RoleTitle      = textRole{1.6, 700, 1.2}   // the game's name on its banner: 22 DIP
	RoleHeading    = textRole{1.15, 600, 1.25} // a card's value, the empty state's heading: 16 DIP
	roleBody       = textRole{1, 400, 0}       // the theme's text, which needs no call: 14 DIP
	RoleBodyStrong = textRole{1, 600, 0}       // the app's name, a game's name in the list, a bar's label
	RoleSmall      = textRole{0.86, 400, 1.35} // meta lines, notes, statuses, paths, the log: 12 DIP
	RoleCaption    = textRole{0.79, 600, 1.3}  // a card's title: 11 DIP
)

// On sets the role on a text element, against the theme t.
func (r textRole) On(t *ui.Theme, e ui.Element) ui.Element {
	e = e.FontSize(roundDIP(t.FontSize * r.scale)).FontWeight(r.weight)
	if r.line > 0 {
		e = e.LineHeight(r.line)
	}
	return e
}

// Glyph sets the initial drawn on a game's tile in place of its icon: part of the icon, it scales with it.
func Glyph(e ui.Element, iconSize float32) ui.Element {
	return e.FontSize(roundDIP(iconSize * 0.45)).FontWeight(700)
}

func roundDIP(v float32) float32 { return float32(math.Round(float64(v))) }

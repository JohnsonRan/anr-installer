// Package style is the window's design system: apexcompute.com's colors, the icons, the type and spacing
// scales, and the components built from them. Every text size, gap and padding comes from here.
package style

import (
	"github.com/egoist/mygo/ui"

	"anr-installer/assets"
)

// Logo is the app's logo.
var Logo = MustSVG(assets.LogoSVG)

// MustSVG parses an SVG built into the app, which cannot fail but by a programming error.
func MustSVG(data []byte) *ui.SVG {
	s, err := ui.ParseSVG(data)
	if err != nil {
		panic(err)
	}
	return s
}

// The colors of apexcompute.com, whose pages are dark navy: four steps of surface, lines, ink, a silver primary
// button, teal for focus and what went well, and royal blue for what is chosen.
var (
	ApexBG         = ui.Hex("#0a0f18")
	ApexSurface    = ui.Hex("#101823")
	ApexSurface2   = ui.Hex("#162131")
	apexSurface3   = ui.Hex("#202d3f")
	ApexLine       = ui.Hex("#243143")
	apexLineHi     = ui.Hex("#364960")
	ApexInk        = ui.Hex("#f0f4f8")
	ApexMuted      = ui.Hex("#a7b4c7")
	apexFaint      = ui.Hex("#8292a9")
	apexTeal       = ui.Hex("#29decc")
	ApexRoyal      = ui.Hex("#5b7cff")
	ApexPeriwinkle = ui.Hex("#9db1ff")
	apexSilver     = ui.Hex("#e4e9f0")
	apexSilverHi   = ui.Hex("#f2f5f9")
	apexSilverLow  = ui.Hex("#ccd4df")
	apexOnSilver   = ui.Hex("#131b28")
)

var apexThemeCache, apexNavCache *ui.Theme

// ApexTheme is the window's theme: always dark, as apexcompute.com is.
func ApexTheme() *ui.Theme {
	if apexThemeCache == nil {
		t := *ui.DarkTheme()
		t.Background = ApexBG
		t.Surface = ApexSurface2
		t.SurfaceHover = apexSurface3
		t.SurfacePressed = ApexSurface
		t.Border = apexLineHi
		t.Text = ApexInk
		t.TextMuted = ApexMuted
		// The primary button is silver with dark text, as the site's; switches and progress take it too.
		t.Accent = apexSilver
		t.AccentHover = apexSilverHi
		t.AccentPressed = apexSilverLow
		t.AccentText = apexOnSilver
		t.InactiveAccent = apexSilver
		t.Success = apexTeal
		t.Warning = ui.Hex("#f5b544")
		t.Danger = ui.Hex("#ff6b6b")
		t.Focus = apexTeal
		t.Selection = ApexRoyal.Alpha(0.35)
		t.Radius = CardRadius
		apexThemeCache = &t
	}
	return apexThemeCache
}

// ApexNavTheme is the theme of the list of games, whose chosen row is a lighter surface, as the navigation of
// Windows 11's Settings shows it, rather than the accent.
func ApexNavTheme() *ui.Theme {
	if apexNavCache == nil {
		t := *ApexTheme()
		t.Accent = apexSurface3
		t.AccentHover = apexSurface3
		t.AccentPressed = apexSurface3
		t.AccentText = ApexInk
		t.InactiveAccent = apexSurface3
		apexNavCache = &t
	}
	return apexNavCache
}

// Icons of the settings rows and the title bar, drawn in the color of their text.
var (
	IconChip     = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="5" width="10" height="10" rx="2"/><path d="M8 2.5v2.5M12 2.5v2.5M8 15v2.5M12 15v2.5M2.5 8h2.5M2.5 12h2.5M15 8h2.5M15 12h2.5"/></svg>`))
	IconFrames   = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linejoin="round"><rect x="2.5" y="5.5" width="9" height="7" rx="1.5"/><rect x="8.5" y="7.5" width="9" height="7" rx="1.5"/></svg>`))
	IconLayers   = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linejoin="round"><path d="M10 3 2.5 7 10 11l7.5-4z"/><path d="M2.5 10.5 10 14.5l7.5-4M2.5 13.5 10 17.5l7.5-4"/></svg>`))
	IconFolder   = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linejoin="round"><path d="M2.5 5.5a1.5 1.5 0 0 1 1.5-1.5h3.5l2 2H16a1.5 1.5 0 0 1 1.5 1.5v7A1.5 1.5 0 0 1 16 16H4a1.5 1.5 0 0 1-1.5-1.5z"/></svg>`))
	IconSliders  = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linecap="round"><path d="M3 6h8M15 6h2M3 14h2M9 14h8"/><circle cx="13" cy="6" r="2"/><circle cx="7" cy="14" r="2"/></svg>`))
	IconWarning  = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 3 18 17H2z"/><path d="M10 8.5v4M10 14.8v.2"/></svg>`))
	IconInfo     = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round"><circle cx="10" cy="10" r="7.5"/><path d="M10 9v5M10 6.3v.2"/></svg>`))
	IconPower    = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round"><path d="M10 2.5v7"/><path d="M6.2 5.2a6.5 6.5 0 1 0 7.6 0"/></svg>`))
	IconBell     = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><path d="M5 13.5V9a5 5 0 0 1 10 0v4.5l1.5 1.5h-13z"/><path d="M8.5 17.5a1.5 1.5 0 0 0 3 0"/></svg>`))
	IconDownload = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 3.5v10M6 9.5l4 4 4-4"/><path d="M4 16.5h12"/></svg>`))
	IconUpdate   = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 15V5M6 9l4-4 4 4"/><path d="M4 16.5h12"/></svg>`))
	IconChevronD = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="m6 8 4 4 4-4"/></svg>`))
	IconChevronU = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="m6 12 4-4 4 4"/></svg>`))
	IconRecord   = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4"><circle cx="10" cy="10" r="7.5"/><circle cx="10" cy="10" r="3"/></svg>`))
	IconGear     = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linejoin="round"><path d="M8.47 4.3 8.91 2.28 11.09 2.28 11.53 4.3 12.95 4.89 14.69 3.77 16.23 5.31 15.11 7.05 15.7 8.47 17.72 8.91 17.72 11.09 15.7 11.53 15.11 12.95 16.23 14.69 14.69 16.23 12.95 15.11 11.53 15.7 11.09 17.72 8.91 17.72 8.47 15.7 7.05 15.11 5.31 16.23 3.77 14.69 4.89 12.95 4.3 11.53 2.28 11.09 2.28 8.91 4.3 8.47 4.89 7.05 3.77 5.31 5.31 3.77 7.05 4.89z"/><circle cx="10" cy="10" r="2.5"/></svg>`))
	IconGlobe    = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linejoin="round"><circle cx="10" cy="10" r="7.5"/><path d="M2.5 10h15M10 2.5c-2 2-3.2 4.6-3.2 7.5s1.2 5.5 3.2 7.5c2-2 3.2-4.6 3.2-7.5S12 4.5 10 2.5z"/></svg>`))
	IconWindow   = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><rect x="2.5" y="3.5" width="15" height="13" rx="1.5"/><path d="M2.5 7h15M12.5 5.25h.01M15 5.25h.01"/></svg>`))
	IconCheck    = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><circle cx="10" cy="10" r="7.5"/><path d="m6.6 10.2 2.3 2.3 4.5-4.9"/></svg>`))
	IconCross    = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.6" stroke-linecap="round"><circle cx="10" cy="10" r="7.5"/><path d="m7.4 7.4 5.2 5.2m0-5.2-5.2 5.2"/></svg>`))
	IconFilm     = MustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linejoin="round"><rect x="2.5" y="4" width="15" height="12" rx="1.5"/><path d="M6 4v12M14 4v12M2.5 8H6M2.5 12H6M14 8h3.5M14 12h3.5"/></svg>`))
)

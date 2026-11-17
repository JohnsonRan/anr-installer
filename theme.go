package main

import "github.com/egoist/mygo/ui"

// The colors of apexcompute.com, whose pages are dark navy: four steps of surface, lines, ink, a silver primary
// button, teal for focus and what went well, and royal blue for what is chosen.
var (
	apexBG         = ui.Hex("#0a0f18")
	apexSurface    = ui.Hex("#101823")
	apexSurface2   = ui.Hex("#162131")
	apexSurface3   = ui.Hex("#202d3f")
	apexLine       = ui.Hex("#243143")
	apexLineHi     = ui.Hex("#364960")
	apexInk        = ui.Hex("#f0f4f8")
	apexMuted      = ui.Hex("#a7b4c7")
	apexFaint      = ui.Hex("#8292a9")
	apexTeal       = ui.Hex("#29decc")
	apexRoyal      = ui.Hex("#5b7cff")
	apexPeriwinkle = ui.Hex("#9db1ff")
	apexSilver     = ui.Hex("#e4e9f0")
	apexSilverHi   = ui.Hex("#f2f5f9")
	apexSilverLow  = ui.Hex("#ccd4df")
	apexOnSilver   = ui.Hex("#131b28")
)

var apexThemeCache, apexNavCache *ui.Theme

// apexTheme is the window's theme: always dark, as apexcompute.com is.
func apexTheme() *ui.Theme {
	if apexThemeCache == nil {
		t := *ui.DarkTheme()
		t.Background = apexBG
		t.Surface = apexSurface2
		t.SurfaceHover = apexSurface3
		t.SurfacePressed = apexSurface
		t.Border = apexLineHi
		t.Text = apexInk
		t.TextMuted = apexMuted
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
		t.Selection = apexRoyal.Alpha(0.35)
		t.Radius = cardRadius
		apexThemeCache = &t
	}
	return apexThemeCache
}

// apexNavTheme is the theme of the list of games, whose chosen row is a lighter surface, as the navigation of
// Windows 11's Settings shows it, rather than the accent.
func apexNavTheme() *ui.Theme {
	if apexNavCache == nil {
		t := *apexTheme()
		t.Accent = apexSurface3
		t.AccentHover = apexSurface3
		t.AccentPressed = apexSurface3
		t.AccentText = apexInk
		t.InactiveAccent = apexSurface3
		apexNavCache = &t
	}
	return apexNavCache
}

// Icons of the settings rows, drawn in the color of their text.
var (
	iconChip     = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="5" width="10" height="10" rx="2"/><path d="M8 2.5v2.5M12 2.5v2.5M8 15v2.5M12 15v2.5M2.5 8h2.5M2.5 12h2.5M15 8h2.5M15 12h2.5"/></svg>`))
	iconFrames   = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linejoin="round"><rect x="2.5" y="5.5" width="9" height="7" rx="1.5"/><rect x="8.5" y="7.5" width="9" height="7" rx="1.5"/></svg>`))
	iconLayers   = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linejoin="round"><path d="M10 3 2.5 7 10 11l7.5-4z"/><path d="M2.5 10.5 10 14.5l7.5-4M2.5 13.5 10 17.5l7.5-4"/></svg>`))
	iconFolder   = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linejoin="round"><path d="M2.5 5.5a1.5 1.5 0 0 1 1.5-1.5h3.5l2 2H16a1.5 1.5 0 0 1 1.5 1.5v7A1.5 1.5 0 0 1 16 16H4a1.5 1.5 0 0 1-1.5-1.5z"/></svg>`))
	iconSliders  = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linecap="round"><path d="M3 6h8M15 6h2M3 14h2M9 14h8"/><circle cx="13" cy="6" r="2"/><circle cx="7" cy="14" r="2"/></svg>`))
	iconWarning  = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 3 18 17H2z"/><path d="M10 8.5v4M10 14.8v.2"/></svg>`))
	iconInfo     = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round"><circle cx="10" cy="10" r="7.5"/><path d="M10 9v5M10 6.3v.2"/></svg>`))
	iconUpdate   = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 15V5M6 9l4-4 4 4"/><path d="M4 16.5h12"/></svg>`))
	iconChevronD = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="m6 8 4 4 4-4"/></svg>`))
	iconChevronU = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="m6 12 4-4 4 4"/></svg>`))
	iconRecord   = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4"><circle cx="10" cy="10" r="7.5"/><circle cx="10" cy="10" r="3"/></svg>`))
	iconFilm     = mustSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="#000" stroke-width="1.4" stroke-linejoin="round"><rect x="2.5" y="4" width="15" height="12" rx="1.5"/><path d="M6 4v12M14 4v12M2.5 8H6M2.5 12H6M14 8h3.5M14 12h3.5"/></svg>`))
)

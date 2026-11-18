//go:build screenshots

package gui

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"anr-installer/gui/activity"
	"anr-installer/gui/i18n"
	"anr-installer/internal/pkg"
	"anr-installer/internal/steam"
)

// The README screenshots, drawn in each language of the READMEs from the Steam games on this computer: their scans
// and Steam's pictures of them, read as the app reads them, with no network. The release list, the log and the
// recordings are made up. Run from the repository root:
//
//	go test -tags screenshots -run TestScreenshots ./gui
//
// They go to docs/images, as <view>.<language>.png.

// The window's default size, as its content draws it on a 1080p display, at 150%.
const shotWidth, shotHeight, shotScale = 1265, 892, 1.5

var shotLangs = map[string]string{"en": "en", "zh-Hans": "zh-CN", "ja": "ja"}

// shotApp is recordApp with this computer's Steam games, in language lang, the first supported one chosen.
// versions tells the versions of ANR FG installed in them.
func shotApp(t *testing.T, lang string, versions *pkg.VersionCache) *app {
	a := recordApp(t)
	i18n.SetLang(lang)                                     // after testApp, which sets English; Steam's pictures follow it
	t.Setenv("APPDATA", `C:\Users\Player\AppData\Roaming`) // the data folder the settings show
	for _, info := range versions.Infos {
		a.providers[info.ProviderSHA256] = info.Version
	}
	a.games = nil
	for _, s := range steam.Games(steam.Libraries()) {
		g := &game{Name: s.Name, AppID: s.AppID, Dir: s.Dir}
		g.Scan, _ = scanGame(g.Dir, "", false)
		im := steam.GameImages(g.AppID, steamLanguage())
		hero := im.Hero
		if hero == "" {
			hero = im.Header
		}
		g.Icon, g.Hero, g.Logo = shotBitmap(im.Icon), shotBitmap(hero), shotBitmap(im.Logo)
		a.games = append(a.games, g)
	}
	if len(a.games) == 0 {
		t.Skip("no Steam games on this computer")
	}
	a.row = 0
	a.filter()
	a.relabelVersions()
	return a
}

func shotBitmap(path string) *ui.Bitmap {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	bm, _ := ui.DecodeBitmap(data)
	return bm
}

// shotLog is a log with a reinstall of ANR FG in the chosen game, as the app writes it.
func shotLog(a *app) {
	at := time.Date(2026, 11, 18, 21, 4, 0, 0, time.Local)
	g := a.selected()
	tk := a.log.Begin(gameTask(g, i18n.T("reinstall")), at)
	tk.Step(activity.LevelInfo, i18n.T("backup.file", "OptiScaler.ini", "OptiScaler.ini.bak-20261118-210400"))
	tk.Step(activity.LevelInfo, "anr-fg-V.B19-full.zip ✓ SHA-256")
	tk.Step(activity.LevelInfo, i18n.T("running.install", "V.B19", "full"))
	tk.Step(activity.LevelInfo, i18n.T("done.install", "V.B19", g.Scan.ExeDir))
	tk.Out(`ANR FG (Apex Neural Rendering) installed in ` + g.Scan.ExeDir + `, OptiScaler as dxgi.dll (kept nr_fg.ini: True)`)
	tk.Finish(at.Add(5*time.Second), "")
	a.logOpen = true
}

func saveShot(t *testing.T, img image.Image, name string) {
	t.Helper()
	f, err := os.Create(filepath.Join("..", "docs", "images", name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestScreenshots(t *testing.T) {
	// The user's, read before testApp moves the cache to a scratch folder for the rest of the test.
	versions := pkg.LoadVersionCache()
	for lang, suffix := range shotLangs {
		a := shotApp(t, lang, versions)
		saveShot(t, ui.Render(a.view, shotWidth, shotHeight, shotScale), "game."+suffix+".png")

		shotLog(a)
		saveShot(t, ui.Render(a.view, shotWidth, shotHeight, shotScale), "log."+suffix+".png")

		a.logOpen = false
		a.toggleSettings()
		a.settings.cacheSize = 46_700_000
		saveShot(t, ui.Render(a.view, shotWidth, shotHeight, shotScale), "settings."+suffix+".png")

		recordShot(t, lang, suffix, versions)
	}
}

// recordShot draws the "Record and compare" section of the chosen game with made-up recordings, which the real
// app shows only for a game that nr-record recorded: from its heading row to the advanced options' card under it.
func recordShot(t *testing.T, lang, suffix string, versions *pkg.VersionCache) {
	const height = 1600 // the whole page, unscrolled
	a := shotApp(t, lang, versions)
	// recordApp's recordings of Delta, as the chosen game's, with nr-record in it.
	g := a.selected()
	for _, r := range a.rec.recordings {
		if r.Of("Delta.exe") {
			r.Game = filepath.Base(g.Scan.Exe)
		}
	}
	g.Recorder, g.Scan.NRRecord = true, true
	tt := ui.NewTester(a.view, shotWidth, height)
	top, ok := tt.Find(i18n.T("rec.title"))
	next, ok2 := tt.Find(i18n.T("advanced"))
	if !ok || !ok2 {
		t.Fatalf("%s: the section is not on the page: %q", lang, tt.Texts())
	}
	img := ui.Render(a.view, shotWidth, height, shotScale)
	const margin = 12 // short of the gap between the cards around the section
	saveShot(t, img.SubImage(image.Rect(int((next.X-margin)*shotScale), int((top.Y-margin)*shotScale),
		int((next.X+next.W+margin)*shotScale), int((next.Y-margin/2)*shotScale))), "record."+suffix+".png")
}

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
	"anr-installer/internal/detect"
)

// The README screenshots, drawn from made-up games in each language of the READMEs. Run from the repository root:
//
//	go test -tags screenshots -run TestScreenshots ./gui
//
// They go to docs/images, as <view>.<language>.png.

// The window's default size, as its content draws it on a 1080p display, at 150%.
const shotWidth, shotHeight, shotScale = 1265, 892, 1.5

var shotLangs = map[string]string{"en": "en", "zh-Hans": "zh-CN", "ja": "ja"}

// shotApp is a made-up library: games with ANR FG, one to upgrade, one with anti-cheat and one added by hand.
func shotApp(t *testing.T) *app {
	a := recordApp(t)
	t.Setenv("APPDATA", `C:\Users\Player\AppData\Roaming`) // the data folder the settings show
	a.providers = map[string]string{"aa": "V.B18", "bb": "V.B19"}
	a.games = []*game{
		{Name: "Aurora Drift", Dir: `D:\SteamLibrary\steamapps\common\Aurora Drift`},
		{Name: "Cinder Peaks", Dir: `D:\SteamLibrary\steamapps\common\Cinder Peaks`},
		{Name: "Hollow Lantern", Dir: `D:\SteamLibrary\steamapps\common\Hollow Lantern`},
		{Name: "Ironvale Tactics", Dir: `E:\SteamLibrary\steamapps\common\Ironvale Tactics`},
		{Name: "Neon Harbor", Dir: `E:\SteamLibrary\steamapps\common\Neon Harbor`},
		{Name: "Quiet Meridian", Dir: `D:\Games\Quiet Meridian`, Manual: true},
		{Name: "Skyward Relay", Dir: `E:\SteamLibrary\steamapps\common\Skyward Relay`},
		{Name: "Tidebound", Dir: `D:\SteamLibrary\steamapps\common\Tidebound`},
	}
	results := []detect.Result{
		{Installed: detect.KindFull, Proxy: "dxgi.dll", ANRHash: "aa", OptiINI: true},
		{Installed: detect.KindFull, Proxy: "dxgi.dll", ANRHash: "bb"},
		{},
		{Installed: detect.KindAddon, Proxy: "winmm.dll", ANRHash: "bb", OptiINI: true, UserOpti: true},
		{Installed: detect.KindFull, Proxy: "dxgi.dll", ANRHash: "bb"},
		{},
		{AntiCheat: []string{"Easy Anti-Cheat"}},
		{},
	}
	for i, g := range a.games {
		r := results[i]
		r.Root, r.ExeDir, r.Exe = g.Dir, g.Dir, g.Dir+`\`+g.Name+".exe"
		if !g.Manual {
			r.DLSSG = []string{g.Dir + `\nvngx_dlssg.dll`}
		}
		g.Scan = &r
	}
	// recordApp's recordings are Aurora Drift's.
	aurora := a.games[0]
	aurora.Recorder, aurora.Scan.NRRecord = true, true
	aurora.Scan.ANRDll = aurora.Dir + `\OptiScaler\custom\Apex Neural Rendering FG\dlssg\nvngx_dlssg.dll`
	return a
}

// shotLog is a log with the install of ANR FG in Neon Harbor and the upgrade of Aurora Drift.
func shotLog(a *app) {
	at := time.Date(2026, 11, 18, 21, 4, 0, 0, time.Local)
	aurora, neon := a.games[0], a.games[4]
	tk := a.log.Begin(gameTask(neon, i18n.T("install.v", "V.B19")), at)
	tk.Step(activity.LevelInfo, "anr-fg-V.B19-full.zip ✓ SHA-256")
	tk.Step(activity.LevelInfo, i18n.T("running.install", "V.B19", "full"))
	tk.Step(activity.LevelInfo, i18n.T("done.install", "V.B19", neon.Dir))
	tk.Out(`ANR FG (Apex Neural Rendering) installed in ` + neon.Dir + `, OptiScaler as dxgi.dll`)
	tk.Finish(at.Add(6*time.Second), "")

	tk = a.log.Begin(gameTask(aurora, i18n.T("upgrade", "V.B19")), at.Add(2*time.Minute))
	tk.Step(activity.LevelInfo, i18n.T("backup.file", "OptiScaler.ini", "OptiScaler.ini.bak-20261118-210612"))
	tk.Step(activity.LevelInfo, "anr-fg-V.B19-full.zip ✓ SHA-256")
	tk.Step(activity.LevelInfo, i18n.T("running.install", "V.B19", "full"))
	tk.Step(activity.LevelInfo, i18n.T("done.install", "V.B19", aurora.Dir))
	tk.Out(`ANR FG (Apex Neural Rendering) installed in ` + aurora.Dir + `, OptiScaler as dxgi.dll (kept nr_fg.ini: True)`)
	tk.Finish(at.Add(2*time.Minute+5*time.Second), "")
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
	for lang, suffix := range shotLangs {
		a := shotApp(t)
		i18n.SetLang(lang) // after testApp, which sets English
		a.row = 0          // Aurora Drift
		a.relabelVersions()
		saveShot(t, ui.Render(a.view, shotWidth, shotHeight, shotScale), "game."+suffix+".png")

		shotLog(a)
		saveShot(t, ui.Render(a.view, shotWidth, shotHeight, shotScale), "log."+suffix+".png")

		a.logOpen = false
		a.toggleSettings()
		a.settings.cacheSize = 46_700_000
		saveShot(t, ui.Render(a.view, shotWidth, shotHeight, shotScale), "settings."+suffix+".png")

		recordShot(t, lang, suffix)
	}
}

// recordShot draws the "Record and compare" section of a game with recordings, which the real app shows only for
// a game that nr-record recorded: from its heading row to the advanced options' card under it.
func recordShot(t *testing.T, lang, suffix string) {
	const height = 1600 // the whole page, unscrolled
	a := shotApp(t)
	i18n.SetLang(lang)
	a.row = 0
	a.relabelVersions()
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

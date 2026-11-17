package main

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/egoist/mygo"

	"anr-installer/internal/detect"
	"anr-installer/internal/record"
	"anr-installer/internal/store"
)

// scanGame inspects a game folder off the main thread and, for a game added by hand, reads the product name of
// the .exe found.
func scanGame(dir, exeDir string, manual bool) (*detect.Result, string) {
	r := detect.Scan(dir, exeDir)
	name := ""
	if manual && r.Exe != "" {
		name = detect.ExeName(r.Exe)
	}
	return r, name
}

// scanned keeps a scan's result, on the main thread. A game added by hand that the user has not renamed takes
// its .exe's product name.
func (a *app) scanned(g *game, r *detect.Result, name string) {
	g.Scan, g.Scanning = r, false
	g.Recorder = record.Installed(r.ExeDir)
	if g.Manual && !g.Named && name != "" && name != g.Name {
		g.Name = name
		a.persist()
	}
}

// persist saves the games added by hand and the .exe folders chosen for Steam games.
func (a *app) persist() {
	var games []store.Game
	exeDirs := map[string]string{}
	for _, g := range a.games {
		if g.Manual {
			games = append(games, store.Game{Name: g.Name, Dir: g.Dir, ExeDir: g.ExeDir, Named: g.Named})
		} else if g.ExeDir != "" {
			exeDirs[store.Key(g.Dir)] = g.ExeDir
		}
	}
	a.saved.Games, a.saved.ExeDirs = games, exeDirs
	if err := a.saved.Save(); err != nil {
		a.logf("%s", T("save.failed", err.Error()))
	}
}

// removeGame takes a game added by hand out of the list. Its files stay as they are, ANR FG too.
func (a *app) removeGame(g *game) {
	if !g.Manual || a.busy {
		return
	}
	i := slices.Index(a.games, g)
	if i < 0 {
		return
	}
	a.games = slices.Delete(a.games, i, i+1)
	a.row = -1
	a.persist()
	a.logf("%s", T("removed", g.Name, g.Dir))
}

// renamed records a name the user gave a game added by hand.
func (a *app) renamed(g *game) {
	if g.Name == "" {
		g.Name = filepath.Base(g.Dir)
	}
	g.Named = true
	a.persist()
}

// resetExeDir goes back to the .exe folder found for g.
func (a *app) resetExeDir(g *game) {
	g.ExeDir = ""
	a.persist()
	a.rescan(g)
}

// confirmInstall installs kind k into g, after the user accepts the risk when the game ships an anti-cheat
// system: it may take OptiScaler and ANR FG for a cheat. The question waits off the main thread.
func (a *app) confirmInstall(g *game, k detect.Kind) {
	if g.Scan == nil || len(g.Scan.AntiCheat) == 0 {
		a.install(g, k)
		return
	}
	names := strings.Join(g.Scan.AntiCheat, ", ")
	go func() {
		res, err := mygo.Dialog.Message(mygo.MessageOptions{
			Parent:             a.win,
			Type:               mygo.MessageWarning,
			Title:              T("anticheat.title"),
			Message:            T("anticheat.title"),
			Detail:             T("anticheat.detail", names, g.Name),
			Buttons:            []string{T("anticheat.cancel"), T("anticheat.accept")},
			DefaultButton:      0,
			CancelButton:       0,
			DestructiveButtons: []int{1},
		})
		if err != nil || res.Button != 1 {
			return
		}
		a.win.Update(func() {
			a.logf("%s", T("anticheat.accepted", names))
			a.install(g, k)
		})
	}()
}

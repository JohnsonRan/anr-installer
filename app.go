package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/internal/detect"
	"anr-installer/internal/pkg"
	"anr-installer/internal/steam"
)

// game is a game in the list: found in a Steam library, or a folder the user added.
type game struct {
	Name     string
	AppID    string
	Dir      string
	Manual   bool
	ExeDir   string // the .exe folder the user chose, or "" for the one found
	Scan     *detect.Result
	Scanning bool
	Icon     *ui.Bitmap // Steam's icon of the game, or its .exe's; nil: none
	Header   *ui.Bitmap // Steam's header picture; nil: none
}

// pictures reads the game's pictures: Steam's cached icon and header, else the icon of its .exe. It runs off the
// main thread; the bitmaps are made on it.
func (a *app) pictures(g *game, r *detect.Result) {
	im := steam.GameImages(g.AppID)
	var icon, header []byte
	if im.Icon != "" {
		icon, _ = os.ReadFile(im.Icon)
	}
	if im.Header != "" {
		header, _ = os.ReadFile(im.Header)
	}
	if icon == nil && r != nil && r.Exe != "" {
		icon, _ = mygo.Shell.FileIcon(r.Exe, 64)
	}
	a.win.Update(func() {
		if icon != nil {
			g.Icon, _ = ui.DecodeBitmap(icon)
		}
		if header != nil {
			g.Header, _ = ui.DecodeBitmap(header)
		}
	})
}

const (
	kindAuto = iota
	kindFull
	kindAddon
)

const (
	sourceGitHub = iota
	sourceLocal
)

// app is the installer's state, read by the view on the main thread. Background work changes it through
// win.Update.
type app struct {
	win *mygo.Window

	games         []*game
	loadingGames  bool
	noSteam       bool
	query         string
	onlySupported bool
	list          ui.ListState
	row           int // the row chosen in the filtered list, -1 none
	visible       []int

	rel        *pkg.Release
	relErr     string
	relLoading bool

	kind     int
	source   int
	localZip string
	local    *pkg.Info
	localErr string

	// providers maps the SHA-256 of an ANR FG provider DLL to the version of the package that ships it.
	providers map[string]string

	busy     bool
	progress float64 // -1 unknown, 0..1, or -2 none
	status   string
	logLines []string
	logList  ui.ListState
}

func newApp() *app {
	a := &app{onlySupported: true, row: -1, progress: -2, providers: map[string]string{}}
	a.list.Key = func(i int) any {
		if i < len(a.visible) {
			return a.games[a.visible[i]].Dir
		}
		return i
	}
	a.list.Selected = &a.row
	a.logList.FollowEnd = true
	return a
}

// start finds the games and the latest release in the background.
func (a *app) start() {
	a.scanSteam()
	a.fetchRelease()
}

func (a *app) scanSteam() {
	a.loadingGames = true
	go func() {
		libs := steam.Libraries()
		found := steam.Games(libs)
		a.win.Update(func() {
			a.loadingGames = false
			a.noSteam = len(libs) == 0
			known := map[string]*game{}
			for _, g := range a.games {
				known[strings.ToLower(g.Dir)] = g
			}
			var games []*game
			for _, s := range found {
				if g, ok := known[strings.ToLower(s.Dir)]; ok && !g.Manual {
					games = append(games, g)
					continue
				}
				games = append(games, &game{Name: s.Name, AppID: s.AppID, Dir: s.Dir})
			}
			for _, g := range a.games {
				if g.Manual {
					games = append(games, g)
				}
			}
			a.games = games
			a.scanAll()
		})
	}()
}

// scanAll inspects every game folder, a few at a time.
func (a *app) scanAll() {
	var todo []*game
	for _, g := range a.games {
		g.Scanning = true
		todo = append(todo, g)
	}
	jobs := make(chan *game)
	var wg sync.WaitGroup
	for range min(4, runtime.NumCPU()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for g := range jobs {
				r := detect.Scan(g.Dir, g.ExeDir)
				a.win.Update(func() { g.Scan, g.Scanning = r, false })
				a.pictures(g, r)
			}
		}()
	}
	go func() {
		for _, g := range todo {
			jobs <- g
		}
		close(jobs)
		wg.Wait()
	}()
}

func (a *app) rescan(g *game) {
	g.Scanning = true
	dir, exeDir := g.Dir, g.ExeDir
	go func() {
		r := detect.Scan(dir, exeDir)
		a.win.Update(func() { g.Scan, g.Scanning = r, false })
		a.pictures(g, r)
	}()
}

func (a *app) fetchRelease() {
	a.relLoading, a.relErr = true, ""
	go func() {
		rel, err := pkg.Latest(context.Background())
		a.win.Update(func() {
			a.relLoading = false
			if err != nil {
				a.relErr = err.Error()
				return
			}
			a.rel = rel
		})
	}()
}

// selected returns the game chosen in the list, or nil.
func (a *app) selected() *game {
	if a.row < 0 || a.row >= len(a.visible) {
		return nil
	}
	return a.games[a.visible[a.row]]
}

// filter recomputes the rows the list shows.
func (a *app) filter() {
	q := strings.ToLower(strings.TrimSpace(a.query))
	a.visible = a.visible[:0]
	for i, g := range a.games {
		if q != "" && !strings.Contains(strings.ToLower(g.Name), q) {
			continue
		}
		if a.onlySupported && !g.Manual && g.Scan != nil && !g.Scan.Supported() && g.Scan.Installed == detect.KindNone {
			continue
		}
		a.visible = append(a.visible, i)
	}
}

func (a *app) addFolder() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: a.win, Title: T("pick.folder.title"), Directory: true})
		if err != nil || len(paths) == 0 {
			return
		}
		dir := paths[0]
		a.win.Update(func() {
			for i, g := range a.games {
				if strings.EqualFold(g.Dir, dir) {
					a.selectGame(i)
					return
				}
			}
			g := &game{Name: filepath.Base(dir), Dir: dir, Manual: true, ExeDir: dir}
			a.games = append(a.games, g)
			a.rescan(g)
			a.selectGame(len(a.games) - 1)
		})
	}()
}

func (a *app) selectGame(i int) {
	a.filter()
	for row, gi := range a.visible {
		if gi == i {
			a.row = row
			return
		}
	}
}

func (a *app) changeExeDir(g *game) {
	start := g.Dir
	if g.Scan != nil {
		start = g.Scan.ExeDir
	}
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: a.win, Title: T("pick.folder.title"), Directory: true, DefaultPath: start})
		if err != nil || len(paths) == 0 {
			return
		}
		a.win.Update(func() {
			g.ExeDir = paths[0]
			a.rescan(g)
		})
	}()
}

func (a *app) chooseZip() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent: a.win, Title: T("pick.zip.title"),
			Filters: []mygo.FileFilter{{Name: "ANR FG (*.zip)", Extensions: []string{"zip"}}},
		})
		if err != nil || len(paths) == 0 {
			return
		}
		info, ierr := pkg.ReadInfo(paths[0])
		a.win.Update(func() {
			a.localZip, a.local, a.localErr = paths[0], info, ""
			if ierr != nil {
				a.local, a.localErr = nil, ierr.Error()
				return
			}
			a.providers[info.ProviderSHA256] = info.Version
			a.source = sourceLocal
		})
	}()
}

// chosenKind returns the package kind the install uses for g.
func (a *app) chosenKind(g *game) detect.Kind {
	switch a.kind {
	case kindFull:
		return detect.KindFull
	case kindAddon:
		return detect.KindAddon
	}
	if g.Scan == nil {
		return detect.KindFull
	}
	return g.Scan.Recommended()
}

// blocked returns why an install of kind k into g cannot run, or "".
func (a *app) blocked(g *game, k detect.Kind) string {
	if g.Scan == nil {
		return T("scanning")
	}
	if ok, why := g.Scan.CanInstall(k); !ok {
		return T("why." + why)
	}
	return a.sourceBlocked(k)
}

func (a *app) sourceBlocked(k detect.Kind) string {
	if a.source == sourceLocal {
		if a.local == nil {
			return T("why.no-zip")
		}
		if a.local.Kind != k {
			return T("why.zip-kind", a.local.Kind, k)
		}
		return ""
	}
	if a.rel == nil {
		return T("why.no-release")
	}
	return ""
}

// targetVersion is the version an install would put in.
func (a *app) targetVersion() string {
	if a.source == sourceLocal && a.local != nil {
		return a.local.Version
	}
	if a.rel != nil {
		return a.rel.Version
	}
	return ""
}

func (a *app) logf(format string, args ...any) {
	a.logLines = append(a.logLines, fmt.Sprintf(format, args...))
}

// run installs or uninstalls the package of kind k in g, in the background.
func (a *app) run(g *game, action pkg.Action, k detect.Kind) {
	if a.busy || g.Scan == nil {
		return
	}
	a.busy, a.progress = true, -1
	scan := g.Scan
	gameDir := scan.ExeDir
	source, rel, localZip := a.source, a.rel, a.localZip
	a.logf("── %s · %s · %s", g.Name, action, k)
	a.logf("%s", gameDir)

	update := func(fn func()) { a.win.Update(fn) }
	say := func(format string, args ...any) {
		line := fmt.Sprintf(format, args...)
		update(func() { a.status = line; a.logLines = append(a.logLines, line) })
	}
	go func() {
		err := func() error {
			ctx := context.Background()
			zipPath := localZip
			if source == sourceGitHub {
				name := rel.AssetName(k)
				update(func() { a.status = T("downloading", name, "") })
				p, err := pkg.Download(ctx, rel, k, func(done, total int64) {
					update(func() {
						if total > 0 {
							a.progress = float64(done) / float64(total)
						}
						a.status = T("downloading", name, fmt.Sprintf("%.1f / %.1f MB", float64(done)/1e6, float64(total)/1e6))
					})
				})
				if err != nil {
					return err
				}
				zipPath = p
				say("%s ✓ SHA-256", name)
			}
			update(func() { a.progress = -1 })
			info, err := pkg.ReadInfo(zipPath)
			if err != nil {
				return err
			}
			if info.Kind != k {
				return fmt.Errorf("%s", T("why.zip-kind", info.Kind, k))
			}
			update(func() { a.providers[info.ProviderSHA256] = info.Version })
			say("%s", T("extracting"))
			dir, err := pkg.Extract(zipPath)
			if err != nil {
				return err
			}
			if action == pkg.Install && k == detect.KindFull && scan.Installed == detect.KindFull && scan.OptiINI {
				ini := filepath.Join(gameDir, "OptiScaler.ini")
				bak := ini + ".bak-" + time.Now().Format("20060102-150405")
				if data, err := os.ReadFile(ini); err == nil {
					if err := os.WriteFile(bak, data, 0o644); err != nil {
						return err
					}
					say("%s", T("backup.ini", filepath.Base(bak)))
				}
			}
			say("%s", T("running", string(action)))
			if err := pkg.Run(ctx, dir, action, gameDir, func(line string) {
				update(func() { a.logLines = append(a.logLines, "  "+line) })
			}); err != nil {
				return err
			}
			if action == pkg.Install {
				say("%s", T("done.install", info.Version, gameDir))
				if scan.NRRecord {
					say("%s", T("warn.nrrecord"))
				}
			} else {
				say("%s", T("done.uninstall", gameDir))
			}
			return nil
		}()
		update(func() {
			a.busy, a.progress = false, -2
			if err != nil {
				a.status = T("failed", err.Error())
				a.logLines = append(a.logLines, a.status)
			}
			a.rescan(g)
		})
	}()
}

// installedVersion describes the version of ANR FG installed in r.
func (a *app) installedVersion(r *detect.Result) string {
	if r.ANRHash == "" {
		return T("inst.version.unk")
	}
	if v, ok := a.providers[r.ANRHash]; ok {
		return v
	}
	if latest := a.targetVersion(); latest != "" && a.knowsVersion(latest) {
		return T("inst.version.old", latest)
	}
	return T("inst.version.unk")
}

func (a *app) knowsVersion(v string) bool {
	for _, known := range a.providers {
		if known == v {
			return true
		}
	}
	return false
}

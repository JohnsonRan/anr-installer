package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"

	"anr-installer/internal/detect"
	"anr-installer/internal/pkg"
	"anr-installer/internal/steam"
	"anr-installer/internal/store"
)

// game is a game in the list: found in a Steam library, or a folder the user added.
type game struct {
	Name     string
	AppID    string
	Dir      string
	Manual   bool   // added by hand: a game outside Steam
	Named    bool   // renamed by the user; else a manual game takes its .exe's product name
	ExeDir   string // the .exe folder the user chose, or "" for the one found
	Scan     *detect.Result
	Scanning bool
	Icon     *ui.Bitmap // Steam's icon of the game, or its .exe's; nil: none
	Hero     *ui.Bitmap // the picture Steam's library shows at the top of the game's page, else its store header; nil: none
	Logo     *ui.Bitmap // the game's logo, which Steam's library draws over the hero; nil: none
	Recorder bool       // nr-record's capture DLLs are in the game
}

// steamLanguage is the language Steam names its localized pictures by, for the window's language.
func steamLanguage() string {
	if lang == "zh" {
		return "schinese"
	}
	return "english"
}

// pictures reads the game's pictures: Steam's cached icon, hero (else header) and logo, else the icon of its
// .exe. It runs off the main thread; the bitmaps are made on it.
func (a *app) pictures(g *game, r *detect.Result) {
	im := steam.GameImages(g.AppID, steamLanguage())
	read := func(p string) []byte {
		if p == "" {
			return nil
		}
		b, _ := os.ReadFile(p)
		return b
	}
	icon, hero, logo := read(im.Icon), read(im.Hero), read(im.Logo)
	if hero == nil {
		hero = read(im.Header)
	}
	if icon == nil && r != nil && r.Exe != "" {
		icon, _ = mygo.Shell.FileIcon(r.Exe, 64)
	}
	decode := func(b []byte) *ui.Bitmap {
		if b == nil {
			return nil
		}
		bm, _ := ui.DecodeBitmap(b)
		return bm
	}
	a.win.Update(func() { g.Icon, g.Hero, g.Logo = decode(icon), decode(hero), decode(logo) })
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

	games        []*game
	loadingGames bool
	noSteam      bool
	query        string
	show         int              // which games the list shows: showAll, showSupported or showInstalled
	advanced     bool             // the advanced install options are open
	logOpen      bool             // the activity log is open
	onlyOutdated bool             // the list shows only the games whose ANR FG can be upgraded
	aboutOpen    bool             // the About dialog is open
	notesRel     *pkg.Release     // the release whose notes the dialog shows, or nil
	md           markdown.Options // how the notes are drawn
	queue        []*game          // games waiting to be upgraded by "Update all"
	autoCheck    bool             // the settings page's copy of the updater's automatic checks
	list         ui.ListState
	row          int // the row chosen in the filtered list, -1 none
	visible      []int

	// saved holds the games added by hand and the .exe folders chosen, kept across launches.
	saved *store.Data

	rel           *pkg.Release   // the newest release
	releases      []*pkg.Release // every release, newest first
	relErr        string
	relLoading    bool
	versionSel    string         // the label of the version to install, from versionLabels
	versionRel    []*pkg.Release // the releases offered, as versionLabels lists them
	versionLabels []string

	// versions caches what each published package says of itself; probing reads the packages it lacks.
	versions *pkg.VersionCache
	probing  bool

	kind     int
	source   int
	localZip string
	local    *pkg.Info
	localErr string

	// providers maps the SHA-256 of an ANR FG provider DLL to the version of the package that ships it.
	providers map[string]string

	busy     bool
	progress float64     // -1 unknown, 0..1, or -2 none
	status   string      // what the task in progress does now
	log      activityLog // the tasks and notes the activity panel shows

	rec recState      // nr-record: its package, the tools it needs and the recordings
	cmp *compareState // the comparison page, shown in place of a game's page; nil: none

	settings *settingsState // the settings page, shown in place of the others; nil: none
	tray     *mygo.Tray     // the notification area icon, while closing the window hides it; nil: none
	quitting bool           // a quit is under way, which closes the window whatever the setting
	quitAsk  bool           // the dialog asking whether to quit while a task runs is open
	// quitForced lets the next quit through without asking: the user answered "Quit anyway".
	quitForced bool
}

func newApp() *app {
	a := &app{show: showSupported, row: -1, progress: -2, providers: map[string]string{}}
	a.list.Key = func(i int) any {
		if i < len(a.visible) {
			return a.games[a.visible[i]].Dir
		}
		return i
	}
	a.list.Selected = &a.row
	a.saved = store.Load()
	for _, s := range a.saved.Games {
		a.games = append(a.games, &game{Name: s.Name, Dir: s.Dir, ExeDir: s.ExeDir, Manual: true, Named: s.Named})
	}
	a.versions = pkg.LoadVersionCache()
	for _, info := range a.versions.Infos {
		a.providers[info.ProviderSHA256] = info.Version
	}
	return a
}

// start finds the games and the latest release in the background.
func (a *app) start() {
	a.scanSteam()
	a.fetchRelease()
	a.watchReleases()
	a.refreshRecord()
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
				games = append(games, &game{Name: s.Name, AppID: s.AppID, Dir: s.Dir, ExeDir: a.saved.ExeDirs[store.Key(s.Dir)]})
			}
			for _, g := range a.games {
				if g.Manual {
					games = append(games, g)
				}
			}
			a.games = games
			a.scanAll()
			a.notifyRelease() // with no game to scan, nothing else calls it
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
				r, name := scanGame(g.Dir, g.ExeDir, g.Manual)
				a.win.Update(func() { a.scanned(g, r, name) })
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
	dir, exeDir, manual := g.Dir, g.ExeDir, g.Manual
	go func() {
		r, name := scanGame(dir, exeDir, manual)
		a.win.Update(func() { a.scanned(g, r, name) })
		a.pictures(g, r)
	}()
}

// selected returns the game chosen in the list, or nil.
func (a *app) selected() *game {
	if a.row < 0 || a.row >= len(a.visible) {
		return nil
	}
	return a.games[a.visible[a.row]]
}

// Which games the list shows.
const (
	showAll       = iota
	showSupported // with DLSS frame generation or ANR FG, and those added by hand
	showInstalled // with ANR FG
)

func (a *app) shows(g *game) bool {
	if a.onlyOutdated {
		_, ok := a.outdated(g.Scan)
		return ok
	}
	if g.Scan == nil {
		return a.show != showInstalled
	}
	switch a.show {
	case showSupported:
		return g.Manual || g.Scan.Supported() || g.Scan.Installed != detect.KindNone
	case showInstalled:
		return g.Scan.Installed != detect.KindNone
	}
	return true
}

// filter recomputes the rows the list shows.
func (a *app) filter() {
	q := strings.ToLower(strings.TrimSpace(a.query))
	a.visible = a.visible[:0]
	for i, g := range a.games {
		if q != "" && !strings.Contains(strings.ToLower(g.Name), q) {
			continue
		}
		if !a.shows(g) {
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
			g := &game{Name: filepath.Base(dir), Dir: dir, Manual: true}
			a.games = append(a.games, g)
			a.persist()
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
			a.persist()
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

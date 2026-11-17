package steam

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Game is a game installed in a Steam library.
type Game struct {
	AppID   string
	Name    string
	Dir     string // steamapps/common/<installdir>
	Library string
}

// Libraries returns the Steam library folders (each holding steamapps/), the Steam folder first, without duplicates.
func Libraries() []string {
	var libs []string
	seen := map[string]bool{}
	add := func(p string) {
		if p == "" {
			return
		}
		p = filepath.Clean(p)
		key := strings.ToLower(p)
		if seen[key] {
			return
		}
		if fi, err := os.Stat(filepath.Join(p, "steamapps")); err != nil || !fi.IsDir() {
			return
		}
		seen[key] = true
		libs = append(libs, p)
	}
	for _, root := range steamRoots() {
		add(root)
		for _, name := range []string{"libraryfolders.vdf", "libraryFolders.vdf"} {
			data, err := os.ReadFile(filepath.Join(root, "steamapps", name))
			if err != nil {
				continue
			}
			doc, err := ParseVDF(string(data))
			if err != nil {
				continue
			}
			lf := doc.Get("libraryfolders")
			if lf == nil {
				continue
			}
			for _, k := range lf.Order {
				entry := lf.Children[k]
				if entry.Children != nil {
					add(entry.Str("path"))
				} else if entry.Value != "" && isDigits(k) { // the old format: "1" "D:\\SteamLibrary"
					add(entry.Value)
				}
			}
		}
	}
	return libs
}

// Games returns the games installed in libs, sorted by name.
func Games(libs []string) []Game {
	var games []Game
	seen := map[string]bool{}
	for _, lib := range libs {
		apps := filepath.Join(lib, "steamapps")
		manifests, _ := filepath.Glob(filepath.Join(apps, "appmanifest_*.acf"))
		for _, m := range manifests {
			data, err := os.ReadFile(m)
			if err != nil {
				continue
			}
			doc, err := ParseVDF(string(data))
			if err != nil {
				continue
			}
			st := doc.Get("AppState")
			if st == nil {
				continue
			}
			dir := st.Str("installdir")
			if dir == "" {
				continue
			}
			full := filepath.Join(apps, "common", dir)
			if fi, err := os.Stat(full); err != nil || !fi.IsDir() {
				continue
			}
			name := st.Str("name")
			if name == "" {
				name = dir
			}
			if isTool(name) {
				continue
			}
			key := strings.ToLower(full)
			if seen[key] {
				continue
			}
			seen[key] = true
			games = append(games, Game{AppID: st.Str("appid"), Name: name, Dir: full, Library: lib})
		}
	}
	sort.Slice(games, func(i, j int) bool { return strings.ToLower(games[i].Name) < strings.ToLower(games[j].Name) })
	return games
}

// isTool reports Steam's own runtimes and redistributables, which are no games.
func isTool(name string) bool {
	n := strings.ToLower(name)
	for _, p := range []string{"proton", "steam linux runtime", "steamworks common redistributables", "steamvr"} {
		if strings.HasPrefix(n, p) {
			return true
		}
	}
	return false
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

package steam

import (
	"os"
	"path/filepath"
	"strings"
)

// Images are the pictures Steam's library caches for a game, as paths ("" when missing).
type Images struct {
	Icon   string // the small square icon
	Header string // the 460×215 store header
	Hero   string // the 1920×620 picture at the top of the game's page in the library
	Logo   string // the game's logo, which the library draws over the hero
}

// GameImages finds the cached pictures of appID in Steam's appcache/librarycache. Steam keeps them in
// librarycache/<appid>/, each kind in a folder of its own named by a hash (library_hero.jpg,
// library_header.jpg, logo.png), with a suffix for a language they are drawn in (library_hero_schinese.jpg);
// earlier versions kept them in <appid>/ itself (header.jpg) or as <appid>_icon.jpg and <appid>_header.jpg.
// lang is the Steam language preferred (schinese, english, …): a picture in it, else one without a language,
// else one in any language.
func GameImages(appID, lang string) Images {
	var im Images
	if appID == "" {
		return im
	}
	for _, root := range steamRoots() {
		cache := filepath.Join(root, "appcache", "librarycache")
		dir := filepath.Join(cache, appID)
		files := listFiles(dir, 2)
		im.Hero = pick(files, "library_hero", ".jpg", lang)
		im.Logo = pick(files, "logo", ".png", lang)
		im.Header = pick(files, "library_header", ".jpg", lang)
		if im.Header == "" {
			im.Header = pick(files, "header", ".jpg", lang)
		}
		for _, f := range files {
			name := strings.ToLower(filepath.Base(f))
			if filepath.Dir(f) == dir && strings.HasSuffix(name, ".jpg") && len(name) == 44 && isHex(name[:40]) {
				im.Icon = f
				break
			}
		}
		if im.Icon == "" && exists(filepath.Join(cache, appID+"_icon.jpg")) {
			im.Icon = filepath.Join(cache, appID+"_icon.jpg")
		}
		if im.Header == "" && exists(filepath.Join(cache, appID+"_header.jpg")) {
			im.Header = filepath.Join(cache, appID+"_header.jpg")
		}
		if im.Hero == "" && exists(filepath.Join(cache, appID+"_library_hero.jpg")) {
			im.Hero = filepath.Join(cache, appID+"_library_hero.jpg")
		}
		if im.Icon != "" || im.Header != "" || im.Hero != "" {
			return im
		}
	}
	return im
}

// pick returns the file named base+ext, or base_<language>+ext, preferring lang, then no language, then any.
func pick(files []string, base, ext, lang string) string {
	var plain, any string
	for _, f := range files {
		name := strings.ToLower(filepath.Base(f))
		if !strings.HasSuffix(name, ext) {
			continue
		}
		stem := strings.TrimSuffix(name, ext)
		switch {
		case stem == base:
			plain = f
		case strings.HasPrefix(stem, base+"_"):
			l := strings.TrimPrefix(stem, base+"_")
			if l == "blur" || strings.HasPrefix(l, "blur_") { // library_hero_blur: not the picture itself
				continue
			}
			if l == lang {
				return f
			}
			if any == "" {
				any = f
			}
		}
	}
	if plain != "" {
		return plain
	}
	return any
}

// listFiles returns the files in dir and in its folders, depth levels down.
func listFiles(dir string, depth int) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		if e.IsDir() {
			if depth > 1 {
				files = append(files, listFiles(p, depth-1)...)
			}
			continue
		}
		files = append(files, p)
	}
	return files
}

func isHex(s string) bool {
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

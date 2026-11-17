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
}

// GameImages finds the cached pictures of appID in Steam's appcache/librarycache: the current layout
// (librarycache/<appid>/<sha1>.jpg, header.jpg) and the older one (<appid>_icon.jpg, <appid>_header.jpg).
func GameImages(appID string) Images {
	var im Images
	if appID == "" {
		return im
	}
	for _, root := range steamRoots() {
		cache := filepath.Join(root, "appcache", "librarycache")
		dir := filepath.Join(cache, appID)
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				name := strings.ToLower(e.Name())
				switch {
				case e.IsDir():
					// Some games keep their pictures one folder down.
					if im.Header == "" && exists(filepath.Join(dir, e.Name(), "header.jpg")) {
						im.Header = filepath.Join(dir, e.Name(), "header.jpg")
					}
				case name == "header.jpg":
					im.Header = filepath.Join(dir, e.Name())
				case im.Icon == "" && strings.HasSuffix(name, ".jpg") && len(name) == 44 && isHex(name[:40]):
					im.Icon = filepath.Join(dir, e.Name())
				}
			}
		}
		if im.Icon == "" && exists(filepath.Join(cache, appID+"_icon.jpg")) {
			im.Icon = filepath.Join(cache, appID+"_icon.jpg")
		}
		if im.Header == "" && exists(filepath.Join(cache, appID+"_header.jpg")) {
			im.Header = filepath.Join(cache, appID+"_header.jpg")
		}
		if im.Icon != "" || im.Header != "" {
			return im
		}
	}
	return im
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

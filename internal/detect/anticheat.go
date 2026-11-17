package detect

import "strings"

// antiCheats are the anti-cheat systems recognized by the folders and files they put in a game's folder.
// A game protected by one may ban a player whose game loads OptiScaler and ANR FG, which hook into its
// rendering, so the installer warns before installing into it.
var antiCheats = []struct {
	name  string
	dirs  []string // folder names, lower case
	files []string // file names or prefixes ending in *, lower case
}{
	{"Easy Anti-Cheat", []string{"easyanticheat", "easyanticheat_eos"}, []string{"start_protected_game.exe", "easyanticheat*"}},
	{"EA Javelin Anticheat", []string{"eaanticheat"}, []string{"eaanticheat*"}},
	{"BattlEye", []string{"battleye"}, []string{"beservice*", "beclient*"}},
	{"nProtect GameGuard", []string{"gameguard"}, []string{"gamemon.des", "gamemon64.des"}},
	{"XIGNCODE3", []string{"xigncode"}, []string{"x3.xem", "xhunter1.sys"}},
	{"Anti-Cheat Expert", []string{"anticheatexpert"}, []string{"ace-base*"}},
	{"mhyprot", nil, []string{"mhyprot*"}},
	{"PunkBuster", []string{"pb"}, []string{"pbsvc*"}},
}

// antiCheatOf returns the anti-cheat system a folder or file name belongs to, or "".
func antiCheatOf(name string, dir bool) string {
	for _, ac := range antiCheats {
		if dir {
			for _, d := range ac.dirs {
				if name == d {
					return ac.name
				}
			}
			continue
		}
		for _, f := range ac.files {
			if p, ok := strings.CutSuffix(f, "*"); ok && strings.HasPrefix(name, p) || name == f {
				return ac.name
			}
		}
	}
	return ""
}

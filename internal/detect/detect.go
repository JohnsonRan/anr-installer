// Package detect inspects a game folder: whether the game ships DLSS frame generation, where its .exe is,
// and what of OptiScaler and ANR FG is installed there.
package detect

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ANRDir is the folder ANR FG installs into, under OptiScaler's custom/ folder.
const ANRDir = "Apex Neural Rendering FG"

// OldANRDir is the folder V.B01 to V.B03 installed into.
const OldANRDir = "nrfg"

// Kind is an ANR FG package kind.
type Kind string

const (
	KindNone  Kind = ""
	KindFull  Kind = "full"  // ANR FG with its own OptiScaler
	KindAddon Kind = "addon" // ANR FG added to an OptiScaler the user installed
)

// Result is what Scan found in a game folder.
type Result struct {
	Root   string // the folder scanned (the game's install folder)
	ExeDir string // the folder of the game's .exe, where the installers go
	Exe    string // the .exe chosen in ExeDir

	DLSSG      []string // DLSS frame generation libraries found (nvngx_dlssg.dll, sl.dlss_g.dll)
	Streamline bool     // the game uses NVIDIA Streamline

	Installed   Kind   // ANR FG installed by a package of this kind, or KindNone
	Proxy       string // the name OptiScaler loads as, for a full install
	OptiINI     bool   // OptiScaler.ini is in ExeDir
	UserOpti    bool   // an OptiScaler not installed by ANR FG's full package
	ANRDll      string // the installed ANR FG provider, if any
	ANRHash     string // its SHA-256
	NRRecord    bool   // OptiScaler.ini routes frame generation to nr-record
	Replacement string // OptiScaler.ini's FGNvngxReplacement
}

// Supported reports whether the game ships DLSS frame generation, which ANR FG takes the place of.
func (r *Result) Supported() bool { return len(r.DLSSG) > 0 }

// Recommended returns the package kind that fits the folder.
func (r *Result) Recommended() Kind {
	switch {
	case r.Installed != KindNone:
		return r.Installed
	case r.UserOpti:
		return KindAddon
	default:
		return KindFull
	}
}

// CanInstall reports whether a package of kind k can go into the folder, and why not.
func (r *Result) CanInstall(k Kind) (bool, string) {
	switch k {
	case KindFull:
		if r.UserOpti {
			return false, "user-optiscaler"
		}
	case KindAddon:
		if !r.OptiINI {
			return false, "no-optiscaler"
		}
		if r.Installed == KindFull {
			return false, "full-installed"
		}
	}
	return true, ""
}

// maxDepth bounds the walk: Unreal games keep Streamline 8 folders down, in
// Engine/Plugins/Runtime/Nvidia/StreamlineCore/Binaries/ThirdParty/Win64.
const maxDepth = 12

var skipDirs = map[string]bool{
	"_commonredist": true, "redist": true, "redistributables": true, "directx": true, "vcredist": true,
	"easyanticheat": true, "battleye": true, "__installer": true, "support": true, "movies": true,
	"shadercache": true, "logs": true, ".git": true,
}

var skipExe = []string{
	"crash", "report", "unins", "redist", "setup", "dxsetup", "easyanticheat", "eac_", "battleye", "be_service",
	"cefprocess", "helper", "launcher", "uploader", "webhelper", "updater", "install", "vc_", "dotnet", "ue4prereq",
	"ueprereq", "dxwebsetup", "quicksfv", "benchmark",
}

type exeInfo struct {
	path string
	size int64
}

// Scan inspects the game installed in root. ExeDir may be forced with exeDir (as the user chose it).
func Scan(root, exeDir string) *Result {
	r := &Result{Root: root}
	var exes []exeInfo
	optiDirs := map[string]bool{}
	dlssgDirs := map[string]bool{}
	base := strings.Count(filepath.Clean(root), string(filepath.Separator))
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		name := strings.ToLower(d.Name())
		if d.IsDir() {
			if p != root && (skipDirs[name] || strings.Count(filepath.Clean(p), string(filepath.Separator))-base > maxDepth) {
				return fs.SkipDir
			}
			if name == "optiscaler" { // our full install's folder: no game files inside
				return fs.SkipDir
			}
			return nil
		}
		dir := filepath.Dir(p)
		switch {
		case name == "nvngx_dlssg.dll" || name == "sl.dlss_g.dll":
			r.DLSSG = append(r.DLSSG, p)
			dlssgDirs[dir] = true
			if name == "sl.dlss_g.dll" {
				r.Streamline = true
			}
		case name == "sl.interposer.dll":
			r.Streamline = true
		case name == "optiscaler.ini":
			optiDirs[dir] = true
		case strings.HasSuffix(name, ".exe") && !skippedExe(name):
			if fi, err := d.Info(); err == nil {
				exes = append(exes, exeInfo{p, fi.Size()})
			}
		}
		return nil
	})

	if exeDir == "" {
		exeDir, r.Exe = pickExeDir(exes, optiDirs, dlssgDirs)
	} else {
		r.Exe = largestExe(exes, exeDir)
	}
	if exeDir == "" {
		exeDir = root
	}
	r.ExeDir = exeDir
	r.inspectInstall()
	return r
}

func skippedExe(name string) bool {
	for _, s := range skipExe {
		if strings.Contains(name, s) {
			return true
		}
	}
	return false
}

func pickExeDir(exes []exeInfo, optiDirs, dlssgDirs map[string]bool) (string, string) {
	if len(exes) == 0 {
		return "", ""
	}
	dirs := map[string]bool{}
	for _, e := range exes {
		dirs[filepath.Dir(e.path)] = true
	}
	// An earlier install of OptiScaler sits next to the .exe.
	for d := range optiDirs {
		if dirs[d] {
			return d, largestExe(exes, d)
		}
	}
	// Unreal Engine games: <Game>/Binaries/Win64/<Game>-Win64-Shipping.exe.
	var best exeInfo
	for _, e := range exes {
		n := strings.ToLower(filepath.Base(e.path))
		if (strings.HasSuffix(n, "-win64-shipping.exe") || strings.HasSuffix(n, "-wingdk-shipping.exe")) && e.size > best.size {
			best = e
		}
	}
	if best.path != "" {
		return filepath.Dir(best.path), best.path
	}
	// The DLSS libraries next to an .exe.
	for d := range dlssgDirs {
		if dirs[d] {
			return d, largestExe(exes, d)
		}
	}
	for _, e := range exes {
		if e.size > best.size {
			best = e
		}
	}
	return filepath.Dir(best.path), best.path
}

func largestExe(exes []exeInfo, dir string) string {
	var best exeInfo
	for _, e := range exes {
		if strings.EqualFold(filepath.Dir(e.path), dir) && e.size > best.size {
			best = e
		}
	}
	return best.path
}

// inspectInstall reads what OptiScaler and ANR FG left in ExeDir, as the packages' installers record it.
func (r *Result) inspectInstall() {
	dir := r.ExeDir
	ini := filepath.Join(dir, "OptiScaler.ini")
	r.OptiINI = exists(ini)
	fullRecord := filepath.Join(dir, "OptiScaler", "nr_fg_install.json")
	if exists(fullRecord) {
		r.Installed = KindFull
		r.Proxy = "dxgi.dll"
		if data, err := os.ReadFile(fullRecord); err == nil {
			var rec struct {
				Proxy string `json:"proxy"`
			}
			if json.Unmarshal(data, &rec) == nil && rec.Proxy != "" {
				r.Proxy = rec.Proxy
			}
		}
		r.ANRDll = providerIn(filepath.Join(dir, "OptiScaler", "custom"))
	} else if r.OptiINI {
		r.UserOpti = true
		main := OptiMainDir(dir, ini)
		for _, name := range []string{ANRDir, OldANRDir} {
			if exists(filepath.Join(main, "custom", name, "nr_fg_install.json")) {
				r.Installed = KindAddon
				r.ANRDll = filepath.Join(main, "custom", name, "dlssg", "nvngx_dlssg.dll")
				break
			}
		}
	}
	if r.ANRDll != "" {
		r.ANRHash, _ = HashFile(r.ANRDll)
	}
	if r.OptiINI {
		r.Replacement = IniValue(ini, "FrameGen", "FGNvngxReplacement")
		r.NRRecord = strings.EqualFold(r.Replacement, "nr-record")
	}
}

// providerIn returns the ANR FG provider under an OptiScaler custom/ folder, in the current layout or, for
// V.B01 to V.B03, in custom/nrfg.
func providerIn(custom string) string {
	for _, name := range []string{ANRDir, OldANRDir} {
		if p := filepath.Join(custom, name, "dlssg", "nvngx_dlssg.dll"); exists(p) {
			return p
		}
	}
	return filepath.Join(custom, ANRDir, "dlssg", "nvngx_dlssg.dll")
}

// OptiMainDir returns the folder of OptiScaler's files, as the add-on installer finds it from OptiDllPath.
func OptiMainDir(gameDir, ini string) string {
	path := strings.TrimSpace(IniValue(ini, "", "OptiDllPath"))
	if path == "" || strings.EqualFold(path, "auto") {
		if fi, err := os.Stat(filepath.Join(gameDir, "OptiScaler")); err == nil && fi.IsDir() {
			return filepath.Join(gameDir, "OptiScaler")
		}
		return gameDir
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(gameDir, path)
}

// IniValue returns key's value in section of an INI file ("" section: the first key of that name anywhere).
func IniValue(path, section, key string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	in := section == ""
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		l := strings.TrimSpace(strings.TrimPrefix(sc.Text(), "\ufeff"))
		if strings.HasPrefix(l, "[") && strings.HasSuffix(l, "]") {
			if section != "" {
				in = l[1:len(l)-1] == section
			}
			continue
		}
		if !in {
			continue
		}
		if k, v, ok := strings.Cut(l, "="); ok && strings.EqualFold(strings.TrimSpace(k), key) {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// HashFile returns the SHA-256 of a file, in lower-case hex.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

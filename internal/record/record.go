// Package record installs nr-record, ANR's recorder of the inputs a game gives its frame generation, and drives
// the tools of its package: it lists the recordings, replays one through two frame generation DLLs into
// videos, and opens the package's viewer on them.
package record

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"anr-installer/internal/pkg"
)

// The package of a release is nr-record-<version>-windows.zip, with its hash in
// nr-record-<version>-SHA256SUMS.txt.
const (
	assetPrefix = "nr-record-"
	assetSuffix = "-windows.zip"
	sumsSuffix  = "-SHA256SUMS.txt"
)

// Dir is the folder nr-record installs into, under OptiScaler's custom/ folder.
const Dir = "nr-record"

// Package is an nr-record package of a release.
type Package struct {
	Rel     *pkg.Release
	Version string // 1.2
}

// Find returns the nr-record package of the newest release that carries one, or nil.
func Find(rels []*pkg.Release) *Package {
	for _, r := range rels {
		best := ""
		for name := range r.Assets {
			v, ok := assetVersion(name)
			if _, sums := r.Assets[assetPrefix+v+sumsSuffix]; ok && sums && (best == "" || newer(v, best)) {
				best = v
			}
		}
		if best != "" {
			return &Package{r, best}
		}
	}
	return nil
}

// assetVersion returns the version in the name of an nr-record package for Windows.
func assetVersion(name string) (string, bool) {
	v, ok := strings.CutPrefix(name, assetPrefix)
	if !ok {
		return "", false
	}
	v, ok = strings.CutSuffix(v, assetSuffix)
	return v, ok && v != ""
}

// newer reports whether version a (1.10) is later than b (1.2), comparing their numbers in order.
func newer(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := range max(len(as), len(bs)) {
		var x, y int
		if i < len(as) {
			x, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			y, _ = strconv.Atoi(bs[i])
		}
		if x != y {
			return x > y
		}
	}
	return false
}

// Asset is the file name of the package.
func (p *Package) Asset() string { return assetPrefix + p.Version + assetSuffix }

// ErrNoPackage says that no nr-record package is known: the releases could not be read and none was
// downloaded before.
var ErrNoPackage = errors.New("no nr-record package: the releases aren't known and none is downloaded")

// Unpack returns the folder of the unpacked package p, which holds its installers, tools and viewer. It
// downloads and checks p when the cache lacks it. With p nil, as when GitHub cannot be reached, it takes the
// newest package downloaded before.
func Unpack(ctx context.Context, p *Package, progress pkg.Progress) (string, error) {
	downloads := pkg.DownloadsDir()
	if p == nil {
		names, _ := filepath.Glob(filepath.Join(downloads, assetPrefix+"*"+assetSuffix))
		best, bestV := "", ""
		for _, n := range names {
			if v, ok := assetVersion(filepath.Base(n)); ok && (best == "" || newer(v, bestV)) {
				best, bestV = n, v
			}
		}
		if best == "" {
			return "", ErrNoPackage
		}
		return pkg.Extract(best)
	}
	// A file under its final name was checked when it was downloaded, so the tools and the viewer open
	// without GitHub. One that no longer unpacks was damaged since, and is fetched and checked again.
	if dir, err := pkg.Extract(filepath.Join(downloads, p.Asset())); err == nil {
		return dir, nil
	}
	zip, err := pkg.DownloadAsset(ctx, p.Rel, p.Asset(), assetPrefix+p.Version+sumsSuffix, progress)
	if err != nil {
		return "", err
	}
	return pkg.Extract(zip)
}

// Installed reports whether nr-record's capture DLLs are in the game whose .exe is in exeDir.
func Installed(exeDir string) bool {
	_, err := os.Stat(filepath.Join(exeDir, "OptiScaler", "custom", Dir))
	return err == nil
}

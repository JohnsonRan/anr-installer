package pkg

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// DownloadsDir is the folder of the downloaded package zips, each checked against its release's sums.
func DownloadsDir() string { return filepath.Join(CacheDir(), "downloads") }

// packagesDir is the folder of the packages unpacked from the zips, a folder each, named by the zip's hash.
func packagesDir() string { return filepath.Join(CacheDir(), "packages") }

// extractedMark is the file Extract writes once a package is unpacked whole.
const extractedMark = ".extracted"

// DownloadsSize returns how many bytes the downloaded and unpacked packages take. What cannot be read counts
// as nothing.
func DownloadsSize() int64 {
	var n int64
	for _, dir := range []string{DownloadsDir(), packagesDir()} {
		filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				if fi, err := d.Info(); err == nil {
					n += fi.Size()
				}
			}
			return nil
		})
	}
	return n
}

// RemoveDownloads deletes the downloaded and unpacked packages, and returns how many bytes it freed. Install,
// uninstall and nr-record download and unpack a package again when it is missing. The rest of the cache stays:
// versions.json, which would take reading every package again, and compare/, which holds the user's
// comparisons. os.RemoveAll deletes all it can before it reports a failure, a package's mark among it, so that
// a package only partly deleted, as when its viewer still runs, is unpacked again rather than taken for whole.
func RemoveDownloads() (int64, error) {
	before := DownloadsSize()
	err := errors.Join(os.RemoveAll(packagesDir()), os.RemoveAll(DownloadsDir()))
	return before - DownloadsSize(), err
}

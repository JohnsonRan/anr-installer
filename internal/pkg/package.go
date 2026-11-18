package pkg

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"anr-installer/internal/detect"
)

// Info is a package's nr_fg_package.json.
type Info struct {
	Product        string      `json:"product"`
	Version        string      `json:"version"`
	Kind           detect.Kind `json:"kind"`
	ProviderSHA256 string      `json:"provider_sha256"`
}

// ReadInfo reads nr_fg_package.json from a package zip.
func ReadInfo(zipPath string) (*Info, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return infoFromZip(zr.File, filepath.Base(zipPath))
}

func infoFromZip(files []*zip.File, name string) (*Info, error) {
	for _, f := range files {
		if filepath.Base(f.Name) != "nr_fg_package.json" || strings.Count(strings.Trim(f.Name, "/"), "/") > 1 {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		var info Info
		if err := json.NewDecoder(rc).Decode(&info); err != nil {
			return nil, err
		}
		if info.Kind != detect.KindFull && info.Kind != detect.KindAddon {
			return nil, fmt.Errorf("unknown package kind %q", info.Kind)
		}
		return &info, nil
	}
	return nil, fmt.Errorf("%s is no ANR FG package (no nr_fg_package.json)", name)
}

// Extract unpacks a package zip into a folder of the cache named after its content hash, once, and returns the
// folder that holds install.ps1 / install.sh.
func Extract(zipPath string) (string, error) {
	sum, err := detect.HashFile(zipPath)
	if err != nil {
		return "", err
	}
	dest := filepath.Join(packagesDir(), sum[:16])
	done := filepath.Join(dest, extractedMark)
	if _, err := os.Stat(done); err != nil {
		os.RemoveAll(dest)
		if err := unzip(zipPath, dest); err != nil {
			os.RemoveAll(dest)
			return "", err
		}
		if err := os.WriteFile(done, nil, 0o644); err != nil {
			return "", err
		}
	}
	// The scripts are at the root, or in a single folder of it.
	if _, err := os.Stat(filepath.Join(dest, "install.ps1")); err == nil {
		return dest, nil
	}
	entries, _ := os.ReadDir(dest)
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(dest, e.Name(), "install.ps1")); err == nil {
				return filepath.Join(dest, e.Name()), nil
			}
		}
	}
	return "", fmt.Errorf("the package has no install.ps1")
}

func unzip(zipPath, dest string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()
	root, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	for _, f := range zr.File {
		p := filepath.Join(root, filepath.FromSlash(f.Name))
		if p != root && !strings.HasPrefix(p, root+string(filepath.Separator)) {
			return fmt.Errorf("unsafe path in zip: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(p, 0o755); err != nil {
				return err
			}
			continue
		}
		if f.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic link in zip: %s", f.Name)
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		if cerr := out.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return err
		}
	}
	return nil
}

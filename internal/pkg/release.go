// Package pkg finds, downloads, verifies and runs the ANR FG packages.
package pkg

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"anr-installer/internal/detect"
	"anr-installer/internal/paths"
)

// Repo is the GitHub repository whose releases carry the packages.
const Repo = "apex-compute/ANR"

const userAgent = "anr-installer"

// Release is an ANR FG release on GitHub.
type Release struct {
	Tag       string // FG-V.B19
	Version   string // V.B19
	URL       string
	Published time.Time
	Notes     string // the release notes, in Markdown
	Assets    map[string]Asset
}

// MinInstallable is the oldest version offered for install: from V.B04 the packages install into
// custom/Apex Neural Rendering FG, the layout the current installers and uninstallers know; V.B01 to V.B03 used
// custom/nrfg.
const MinInstallable = 4

// Number returns the build number of a version such as V.B19 (19), or 0 when it has none.
func Number(version string) int {
	n := 0
	for _, c := range strings.TrimPrefix(version, "V.B") {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// Asset is a file of a release.
type Asset struct {
	Name string
	URL  string
	Size int64
}

// AssetName returns the package file name of kind in this release.
func (r *Release) AssetName(k detect.Kind) string {
	return fmt.Sprintf("anr-fg-%s-%s.zip", r.Version, k)
}

var client = &http.Client{Timeout: 0}

// apiBase is GitHub's API (a test server in tests).
var apiBase = "https://api.github.com"

// Releases returns the published ANR FG releases (tags FG-V.*) that have both packages, newest first.
func Releases(ctx context.Context) ([]*Release, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiBase+"/repos/"+Repo+"/releases?per_page=100", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub: %s", resp.Status)
	}
	var list []struct {
		Tag        string    `json:"tag_name"`
		Draft      bool      `json:"draft"`
		Prerelease bool      `json:"prerelease"`
		HTMLURL    string    `json:"html_url"`
		Published  time.Time `json:"published_at"`
		Body       string    `json:"body"`
		Assets     []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
			Size int64  `json:"size"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	var rels []*Release
	for _, r := range list {
		if r.Draft || r.Prerelease || !strings.HasPrefix(r.Tag, "FG-V.") {
			continue
		}
		rel := &Release{Tag: r.Tag, Version: strings.TrimPrefix(r.Tag, "FG-"), URL: r.HTMLURL, Published: r.Published, Notes: r.Body, Assets: map[string]Asset{}}
		for _, a := range r.Assets {
			rel.Assets[a.Name] = Asset{a.Name, a.URL, a.Size}
		}
		_, full := rel.Assets[rel.AssetName(detect.KindFull)]
		_, addon := rel.Assets[rel.AssetName(detect.KindAddon)]
		if full && addon && Number(rel.Version) > 0 {
			rels = append(rels, rel)
		}
	}
	if len(rels) == 0 {
		return nil, fmt.Errorf("no ANR FG release found in %s", Repo)
	}
	sort.Slice(rels, func(i, j int) bool { return Number(rels[i].Version) > Number(rels[j].Version) })
	return rels, nil
}

// CacheDir is where downloaded and extracted packages are kept.
func CacheDir() string { return paths.Cache() }

// Progress reports bytes done of total (total 0: unknown).
type Progress func(done, total int64)

// Download fetches kind's package of rel into the cache, checks it against the release's SHA256SUMS.txt and
// returns its path. A package already downloaded and matching is reused.
func Download(ctx context.Context, rel *Release, k detect.Kind, progress Progress) (string, error) {
	name := rel.AssetName(k)
	asset, ok := rel.Assets[name]
	if !ok {
		return "", fmt.Errorf("%s has no %s", rel.Tag, name)
	}
	want, err := expectedHash(ctx, rel, name)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(CacheDir(), "downloads")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dest := filepath.Join(dir, name)
	if got, err := detect.HashFile(dest); err == nil && got == want {
		return dest, nil
	}
	if err := fetch(ctx, asset.URL, dest+".part", asset.Size, progress); err != nil {
		return "", err
	}
	got, err := detect.HashFile(dest + ".part")
	if err != nil {
		return "", err
	}
	if got != want {
		os.Remove(dest + ".part")
		return "", fmt.Errorf("%s: SHA-256 %s does not match SHA256SUMS.txt (%s)", name, got, want)
	}
	if err := os.Rename(dest+".part", dest); err != nil {
		return "", err
	}
	return dest, nil
}

func expectedHash(ctx context.Context, rel *Release, name string) (string, error) {
	sums, ok := rel.Assets["SHA256SUMS.txt"]
	if !ok {
		return "", fmt.Errorf("%s has no SHA256SUMS.txt", rel.Tag)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", sums.URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("SHA256SUMS.txt: %s", resp.Status)
	}
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == name {
			return strings.ToLower(f[0]), nil
		}
	}
	return "", fmt.Errorf("SHA256SUMS.txt lists no %s", name)
}

func fetch(ctx context.Context, url, dest string, size int64, progress Progress) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download: %s", resp.Status)
	}
	if size == 0 {
		size = resp.ContentLength
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	var done int64
	buf := make([]byte, 256*1024)
	last := time.Time{}
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				f.Close()
				return werr
			}
			done += int64(n)
			if progress != nil && time.Since(last) > 100*time.Millisecond {
				last = time.Now()
				progress(done, size)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			f.Close()
			return rerr
		}
	}
	if progress != nil {
		progress(done, size)
	}
	return f.Close()
}

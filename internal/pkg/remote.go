package pkg

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"anr-installer/internal/detect"
)

// rangeReader reads a remote file with HTTP range requests, in blocks it keeps, so that archive/zip reads a
// package's directory and one small entry instead of downloading the whole package.
type rangeReader struct {
	ctx    context.Context
	url    string
	size   int64
	mu     sync.Mutex
	blocks map[int64][]byte
}

const blockSize = 64 << 10

func (r *rangeReader) ReadAt(p []byte, off int64) (int, error) {
	if off >= r.size {
		return 0, io.EOF
	}
	n := 0
	for n < len(p) && off+int64(n) < r.size {
		at := off + int64(n)
		start := at / blockSize * blockSize
		b, err := r.block(start)
		if err != nil {
			return n, err
		}
		n += copy(p[n:], b[at-start:])
	}
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (r *rangeReader) block(start int64) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.blocks[start]; ok {
		return b, nil
	}
	end := min(start+blockSize, r.size) - 1
	req, err := http.NewRequestWithContext(r.ctx, "GET", r.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	// GitHub's release downloads take explicit ranges; a suffix range ("bytes=-N") is refused with 501.
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("range request: %s", resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, end-start+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) != end-start+1 {
		return nil, fmt.Errorf("range request: %d bytes for %d", len(b), end-start+1)
	}
	r.blocks[start] = b
	return b, nil
}

// RemoteInfo reads nr_fg_package.json from a release's package without downloading the package.
func RemoteInfo(ctx context.Context, a Asset) (*Info, error) {
	if a.Size <= 0 {
		return nil, fmt.Errorf("%s: unknown size", a.Name)
	}
	rr := &rangeReader{ctx: ctx, url: a.URL, size: a.Size, blocks: map[int64][]byte{}}
	zr, err := zip.NewReader(rr, a.Size)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", a.Name, err)
	}
	return infoFromZip(zr.File, a.Name)
}

// VersionCache maps package file names to what their nr_fg_package.json says. Published packages do not change,
// so it is kept on disk and read once per package.
type VersionCache struct {
	mu    sync.Mutex
	path  string
	Infos map[string]*Info `json:"packages"`
}

// LoadVersionCache reads the cache from the cache folder, or starts an empty one.
func LoadVersionCache() *VersionCache {
	c := &VersionCache{path: filepath.Join(CacheDir(), "versions.json"), Infos: map[string]*Info{}}
	if data, err := os.ReadFile(c.path); err == nil {
		json.Unmarshal(data, c)
		if c.Infos == nil {
			c.Infos = map[string]*Info{}
		}
	}
	return c
}

// Get returns the cached info of a package file.
func (c *VersionCache) Get(name string) *Info {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Infos[name]
}

// Put records a package's info and saves the cache.
func (c *VersionCache) Put(name string, info *Info) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Infos[name] = info
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return
	}
	if data, err := json.MarshalIndent(c, "", " "); err == nil {
		tmp := c.path + ".tmp"
		if os.WriteFile(tmp, data, 0o644) == nil {
			os.Rename(tmp, c.path)
		}
	}
}

// probeWorkers is how many packages Probe reads at once.
const probeWorkers = 6

// Probe reads the info of both packages of each release that the cache lacks, a few at a time, calling found for
// each info, cached or read, from any goroutine. It returns the first error, as with no network, after the reads
// under way end.
func (c *VersionCache) Probe(ctx context.Context, rels []*Release, found func(*Info)) error {
	var missing []Asset
	for _, rel := range rels {
		for _, k := range []detect.Kind{detect.KindFull, detect.KindAddon} {
			name := rel.AssetName(k)
			if info := c.Get(name); info != nil {
				found(info)
			} else if asset, ok := rel.Assets[name]; ok {
				missing = append(missing, asset)
			}
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan Asset)
	var (
		wg    sync.WaitGroup
		once  sync.Once
		first error
	)
	for range min(probeWorkers, len(missing)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for asset := range jobs {
				info, err := RemoteInfo(ctx, asset)
				if err != nil {
					once.Do(func() { first = err; cancel() })
					continue
				}
				c.Put(asset.Name, info)
				found(info)
			}
		}()
	}
feed:
	for _, asset := range missing {
		select {
		case jobs <- asset:
		case <-ctx.Done():
			break feed
		}
	}
	close(jobs)
	wg.Wait()
	return first
}

// PackageFile reads the file name, a path inside the package such as "game/OptiScaler.ini", from kind's package of
// rel: from the downloaded package when the cache holds it, else with range requests, without downloading the
// package. What it reads is not checked against the release's sums: callers only compare it, never install it.
func PackageFile(ctx context.Context, rel *Release, k detect.Kind, name string) ([]byte, error) {
	asset, ok := rel.Assets[rel.AssetName(k)]
	if !ok {
		return nil, fmt.Errorf("%s has no %s", rel.Tag, rel.AssetName(k))
	}
	if zr, err := zip.OpenReader(filepath.Join(DownloadsDir(), asset.Name)); err == nil {
		defer zr.Close()
		return zipFile(zr.File, asset.Name, name)
	}
	if asset.Size <= 0 {
		return nil, fmt.Errorf("%s: unknown size", asset.Name)
	}
	rr := &rangeReader{ctx: ctx, url: asset.URL, size: asset.Size, blocks: map[int64][]byte{}}
	zr, err := zip.NewReader(rr, asset.Size)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", asset.Name, err)
	}
	return zipFile(zr.File, asset.Name, name)
}

func zipFile(files []*zip.File, pkgName, name string) ([]byte, error) {
	for _, f := range files {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, fmt.Errorf("%s has no %s", pkgName, name)
}

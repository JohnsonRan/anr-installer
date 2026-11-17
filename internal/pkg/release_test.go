package pkg

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"anr-installer/internal/detect"
)

// fakePackage returns a package zip whose nr_fg_package.json says version and kind, padded with incompressible
// data so that reading its directory takes several range blocks.
func fakePackage(t *testing.T, version string, kind detect.Kind) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	pad := make([]byte, 300<<10)
	rand.Read(pad)
	w, _ := zw.CreateHeader(&zip.FileHeader{Name: "game/OptiScaler/custom/Apex Neural Rendering FG/dlssg/programs.bin", Method: zip.Store})
	w.Write(pad)
	w, _ = zw.Create("nr_fg_package.json")
	fmt.Fprintf(w, `{"product":"ANR FG Beta","version":%q,"kind":%q,"provider_sha256":"sha-%s-%s"}`, version, kind, version, kind)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// fakeGitHub serves a release list, the packages (with range requests, refusing suffix ranges as GitHub does)
// and SHA256SUMS.txt of each release.
type fakeGitHub struct {
	srv     *httptest.Server
	files   map[string][]byte
	sums    map[string]string
	full    atomic.Int64 // requests for a whole package
	ranges  atomic.Int64
	offline bool
}

func newFakeGitHub(t *testing.T, versions ...string) *fakeGitHub {
	f := &fakeGitHub{files: map[string][]byte{}, sums: map[string]string{}}
	f.srv = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.srv.Close)
	var rels []string
	for _, v := range versions {
		var assets []string
		var sums strings.Builder
		for _, k := range []detect.Kind{detect.KindFull, detect.KindAddon} {
			name := fmt.Sprintf("anr-fg-%s-%s.zip", v, k)
			data := fakePackage(t, v, k)
			f.files["/dl/"+name] = data
			sum := sha256.Sum256(data)
			fmt.Fprintf(&sums, "%s  %s\n", hex.EncodeToString(sum[:]), name)
			assets = append(assets, fmt.Sprintf(`{"name":%q,"browser_download_url":"%s/dl/%s","size":%d}`, name, f.srv.URL, name, len(data)))
		}
		f.sums["/sums/"+v] = sums.String()
		assets = append(assets, fmt.Sprintf(`{"name":"SHA256SUMS.txt","browser_download_url":"%s/sums/%s","size":1}`, f.srv.URL, v))
		rels = append(rels, fmt.Sprintf(`{"tag_name":"FG-%s","html_url":"u","published_at":"2026-11-17T00:20:54Z","assets":[%s]}`, v, strings.Join(assets, ",")))
	}
	rels = append(rels, `{"tag_name":"FG-V.B99","prerelease":true,"assets":[]}`, `{"tag_name":"nr-record-1.3","assets":[]}`)
	f.files["/repos/"+Repo+"/releases"] = []byte("[" + strings.Join(rels, ",") + "]")
	apiBase = f.srv.URL
	t.Cleanup(func() { apiBase = "https://api.github.com" })
	return f
}

func (f *fakeGitHub) serve(w http.ResponseWriter, r *http.Request) {
	if f.offline {
		http.Error(w, "offline", http.StatusServiceUnavailable)
		return
	}
	if s, ok := f.sums[r.URL.Path]; ok {
		fmt.Fprint(w, s)
		return
	}
	data, ok := f.files[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	if rg := r.Header.Get("Range"); rg != "" {
		if strings.HasPrefix(rg, "bytes=-") {
			http.Error(w, "suffix range", http.StatusNotImplemented)
			return
		}
		f.ranges.Add(1)
	} else if strings.HasPrefix(r.URL.Path, "/dl/") {
		f.full.Add(1)
	}
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
}

func useTempCache(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
}

func TestReleasesNewestFirst(t *testing.T) {
	useTempCache(t)
	newFakeGitHub(t, "V.B09", "V.B19", "V.B10")
	rels, err := Releases(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, r := range rels {
		got = append(got, r.Version)
	}
	if strings.Join(got, " ") != "V.B19 V.B10 V.B09" {
		t.Fatalf("releases %v", got)
	}
	if rels[0].Published.IsZero() {
		t.Error("no publish date")
	}
	if Number("V.B04") != 4 || Number("V.Bx") != 0 {
		t.Error("Number")
	}
}

func TestDownloadChecksSums(t *testing.T) {
	useTempCache(t)
	gh := newFakeGitHub(t, "V.B19")
	rels, err := Releases(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var last int64
	p, err := Download(context.Background(), rels[0], detect.KindFull, func(done, total int64) { last = done })
	if err != nil {
		t.Fatal(err)
	}
	want := gh.files["/dl/anr-fg-V.B19-full.zip"]
	if data, _ := os.ReadFile(p); !bytes.Equal(data, want) || last != int64(len(want)) {
		t.Fatalf("downloaded %d bytes, progress %d", len(data), last)
	}
	gh.sums["/sums/V.B19"] = "0000  anr-fg-V.B19-addon.zip\n"
	if _, err := Download(context.Background(), rels[0], detect.KindAddon, nil); err == nil {
		t.Fatal("a package not matching SHA256SUMS.txt was accepted")
	}
}

func TestProbeReadsVersionsWithoutDownloading(t *testing.T) {
	useTempCache(t)
	gh := newFakeGitHub(t, "V.B18", "V.B19")
	rels, err := Releases(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	cache := LoadVersionCache()
	found := map[string]string{}
	if err := cache.Probe(context.Background(), rels, func(i *Info) { found[i.ProviderSHA256] = i.Version }); err != nil {
		t.Fatal(err)
	}
	if len(found) != 4 || found["sha-V.B18-addon"] != "V.B18" || found["sha-V.B19-full"] != "V.B19" {
		t.Fatalf("found %v", found)
	}
	if gh.full.Load() != 0 || gh.ranges.Load() == 0 {
		t.Fatalf("%d whole downloads, %d range requests", gh.full.Load(), gh.ranges.Load())
	}
	// Each package is read in a few blocks, not whole: the padding alone is five blocks.
	if n := gh.ranges.Load(); n > 4*3 {
		t.Errorf("%d range requests for 4 packages", n)
	}

	// The cache answers later probes without the network.
	gh.offline = true
	again := LoadVersionCache()
	n := 0
	if err := again.Probe(context.Background(), rels, func(*Info) { n++ }); err != nil || n != 4 {
		t.Fatalf("cached probe: %d infos, %v", n, err)
	}
}

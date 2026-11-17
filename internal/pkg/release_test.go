package pkg

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"anr-installer/internal/detect"
)

func TestLatestAndDownload(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	zipData := []byte("PK fake package")
	sum := sha256.Sum256(zipData)
	good := hex.EncodeToString(sum[:])
	sums := good + "  anr-fg-V.B19-full.zip\n" + good + "  anr-fg-V.B19-addon.zip\n"
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/" + Repo + "/releases":
			fmt.Fprintf(w, `[
 {"tag_name":"FG-V.B20","prerelease":true,"assets":[]},
 {"tag_name":"nr-record-1.3","assets":[]},
 {"tag_name":"FG-V.B19","html_url":"u","assets":[
  {"name":"anr-fg-V.B19-full.zip","browser_download_url":"%[1]s/full","size":%[2]d},
  {"name":"anr-fg-V.B19-addon.zip","browser_download_url":"%[1]s/full","size":%[2]d},
  {"name":"SHA256SUMS.txt","browser_download_url":"%[1]s/sums","size":1}]}]`, srv.URL, len(zipData))
		case "/full":
			w.Write(zipData)
		case "/sums":
			fmt.Fprint(w, sums)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	apiBase = srv.URL
	defer func() { apiBase = "https://api.github.com" }()

	rel, err := Latest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rel.Version != "V.B19" {
		t.Fatalf("version %q", rel.Version)
	}
	var last int64
	p, err := Download(context.Background(), rel, detect.KindFull, func(done, total int64) { last = done })
	if err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(p); string(data) != string(zipData) || last != int64(len(zipData)) {
		t.Fatalf("downloaded %q, progress %d", data, last)
	}

	sums = "0000  anr-fg-V.B19-addon.zip\n"
	if _, err := Download(context.Background(), rel, detect.KindAddon, nil); err == nil {
		t.Fatal("a package not matching SHA256SUMS.txt was accepted")
	}
}

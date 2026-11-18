package main

import (
	"os"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/plugins/updater/native"
)

// TestMain uses the updater plugin as main does, so that views reading its state run in tests.
func TestMain(m *testing.M) {
	mygo.Use(native.New(updater.Options{Icon: iconPNG}))
	os.Exit(m.Run())
}

func TestDefaultSize(t *testing.T) {
	for _, c := range []struct {
		name         string
		work         mygo.Rectangle
		wantW, wantH int
	}{
		{"1080p at 100%", mygo.Rectangle{Width: 1920, Height: 1032}, 1280, 900},
		{"1080p at 150%", mygo.Rectangle{Width: 1280, Height: 688}, 1152, 619},
		{"1366x768 at 100%", mygo.Rectangle{Width: 1366, Height: 720}, 1229, 648},
		{"smaller than the minimum", mygo.Rectangle{Width: 800, Height: 560}, 760, 520},
		{"no display reported", mygo.Rectangle{}, 1280, 900},
	} {
		if w, h := defaultSize(c.work); w != c.wantW || h != c.wantH {
			t.Errorf("%s: defaultSize = %d × %d, want %d × %d", c.name, w, h, c.wantW, c.wantH)
		}
	}
}

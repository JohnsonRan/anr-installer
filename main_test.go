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

package main

import (
	"os"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater/native"
)

// TestMain uses the updater plugin as main does, so that views reading its state run in tests.
func TestMain(m *testing.M) {
	mygo.Use(native.Plugin)
	os.Exit(m.Run())
}

// Package store keeps what the user told the installer: the game folders added by hand, and the .exe folders
// chosen in place of the ones found.
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"anr-installer/internal/paths"
)

// Game is a game folder added by hand, as a game outside Steam.
type Game struct {
	Name   string `json:"name"`
	Dir    string `json:"dir"`
	ExeDir string `json:"exe_dir,omitempty"`
	// Named is set once the user renamed the game; until then its name follows its .exe.
	Named bool `json:"named,omitempty"`
}

// Data is what is kept.
type Data struct {
	Games []Game `json:"games"`
	// ExeDirs maps a Steam game's folder (lower-cased) to the .exe folder the user chose for it.
	ExeDirs map[string]string `json:"exe_dirs,omitempty"`
}

// Path is the file the data is kept in.
func Path() string { return filepath.Join(paths.Config(), "games.json") }

// Load reads the data, or returns empty data when there is none or it cannot be read.
func Load() *Data {
	d := &Data{}
	if b, err := os.ReadFile(Path()); err == nil {
		json.Unmarshal(b, d)
	}
	if d.ExeDirs == nil {
		d.ExeDirs = map[string]string{}
	}
	return d
}

// Save writes the data, replacing the file whole.
func (d *Data) Save() error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// Key is how folders are compared: Windows paths ignore case.
func Key(dir string) string { return strings.ToLower(filepath.Clean(dir)) }

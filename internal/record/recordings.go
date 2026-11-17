package record

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Recording is a folder nr-record wrote: what its record.json says of it.
type Recording struct {
	// What the folder says, not record.json: its "time" has no zone, and is not read.
	Dir  string    `json:"-"`
	Name string    `json:"-"` // the folder's name: <date>-<time>-<game>
	Time time.Time `json:"-"` // when it started

	Game       string `json:"game"`    // the game's executable
	Feature    string `json:"feature"` // fg, sr or rr; "" in recordings before format 4, which are fg
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Frames     int    `json:"frames"`  // real frames written
	Dropped    int    `json:"dropped"` // real frames the recorder could not keep up with
	Multiplier int    `json:"multiplier"`
	Bytes      int64  `json:"bytes"`
	Status     string `json:"status"` // recording, complete, or stopped: <why>
}

// FrameGeneration reports whether the recording holds the inputs of frame generation, which replays through
// ANR FG and DLSS-G; the others hold an upscaler's.
func (r *Recording) FrameGeneration() bool { return r.Feature == "" || r.Feature == "fg" }

// Of reports whether the recording was made in the game whose executable is exe.
func (r *Recording) Of(exe string) bool {
	return exe != "" && strings.EqualFold(stem(r.Game), stem(exe))
}

// stem is a program's name without its folder and .exe; record.json may hold either form.
func stem(exe string) string {
	exe = exe[strings.LastIndexAny(exe, `\/`)+1:]
	if len(exe) > 4 && strings.EqualFold(exe[len(exe)-4:], ".exe") {
		exe = exe[:len(exe)-4]
	}
	return exe
}

// RecordingsDir is where nr-record puts its recordings: %LOCALAPPDATA%\nr-record on Windows, and
// $XDG_DATA_HOME/nr-record, else ~/.local/share/nr-record, under Proton.
func RecordingsDir() string {
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		return filepath.Join(local, "nr-record")
	}
	if data := os.Getenv("XDG_DATA_HOME"); data != "" {
		return filepath.Join(data, "nr-record")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "nr-record")
}

// Recordings lists the recordings in dir, the newest first. A folder is a recording when it has a record.json.
func Recordings(dir string) []*Recording {
	entries, _ := os.ReadDir(dir)
	var recs []*Recording
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name(), "record.json"))
		if err != nil {
			continue
		}
		r := &Recording{Dir: filepath.Join(dir, e.Name()), Name: e.Name()}
		if json.Unmarshal(data, r) != nil {
			continue
		}
		// The folder's name starts with the local time the recording began.
		if len(r.Name) >= 15 {
			r.Time, _ = time.ParseInLocation("20060102-150405", r.Name[:15], time.Local)
		}
		if r.Time.IsZero() {
			if fi, err := e.Info(); err == nil {
				r.Time = fi.ModTime()
			}
		}
		recs = append(recs, r)
	}
	sort.SliceStable(recs, func(i, j int) bool { return recs[i].Time.After(recs[j].Time) })
	return recs
}

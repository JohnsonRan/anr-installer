package record

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"anr-installer/internal/pkg"
)

// Side is one frame generation DLL a recording is replayed through.
type Side struct {
	Label     string // its name in the videos and the viewer
	Generator string // nr-replay's --generator: provider (a DLSS-G provider such as ANR FG) or nvidia (DLSS-G)
	DLL       string
	name      string // of its files in the comparison's folder
}

// ANR is ANR FG's provider DLL as a side.
func ANR(dll string) Side { return Side{"ANR FG", "provider", dll, "anr"} }

// DLSSG is a game's nvngx_dlssg.dll as a side. Replaying through it needs an NVIDIA RTX 40 or 50 series GPU.
func DLSSG(dll string) Side { return Side{"DLSS-G", "nvidia", dll, "dlssg"} }

// Comparison is a recording replayed through two sides, as two videos and the viewer's project of them.
type Comparison struct {
	Tools     string // the unpacked nr-record package
	Recording *Recording
	A, B      Side
	FFmpeg    string
	Frames    int // how many recorded real frames to replay, from the first; 0: all
}

// Out is the folder the comparison's videos and project go to.
func (c *Comparison) Out() string {
	return filepath.Join(pkg.CacheDir(), "compare", c.Recording.Name)
}

// Project is the viewer's project of the two videos.
func (c *Comparison) Project() string { return filepath.Join(c.Out(), "cmp.compare.json") }

// Done reports whether the comparison was made before.
func (c *Comparison) Done() bool {
	_, err := os.Stat(c.Project())
	return err == nil
}

func (c *Comparison) video(s Side) string { return filepath.Join(c.Out(), s.name+".mp4") }

func (c *Comparison) tool(name string) string { return filepath.Join(c.Tools, "tools", name+".exe") }

// Make replays the recording through both sides into videos and writes their project. It calls step with a
// side's label as its replay starts, and sends each line the tools print to out. The frames stream from
// nr-replay into nr-video, so that only the videos reach the disk.
func (c *Comparison) Make(ctx context.Context, step func(label string), out func(string)) error {
	if err := os.MkdirAll(c.Out(), 0o755); err != nil {
		return err
	}
	// A project left by an earlier run must not outlive a replay that fails now, nor its stills new videos:
	// the DLLs may have changed since.
	os.Remove(c.Project())
	os.RemoveAll(filepath.Join(c.Out(), "stills"))
	for _, s := range []Side{c.A, c.B} {
		step(s.Label)
		replay := exec.CommandContext(ctx, c.tool("nr-replay"), "--record", c.Recording.Dir,
			"--out", filepath.Join(c.Out(), s.name), "--generator", s.Generator, "--dll", s.DLL, "--label", s.Label, "--stream", "-")
		if c.Frames > 0 && c.Frames < c.Recording.Frames {
			replay.Args = append(replay.Args, "--frames", fmt.Sprintf("0:%d", c.Frames))
		}
		video := exec.CommandContext(ctx, c.tool("nr-video"), "video", "-", "--out", c.video(s), "--ffmpeg", c.FFmpeg)
		frames, err := replay.StdoutPipe()
		if err != nil {
			return err
		}
		video.Stdin = frames
		if err := pkg.Stream(out, replay, video); err != nil {
			return err
		}
	}
	return pkg.Stream(out, exec.CommandContext(ctx, c.tool("nr-video"), "project", c.video(c.A), c.video(c.B),
		"--out", c.Project(), "--labels", c.A.Label, c.B.Label))
}

// View opens the package's viewer with args, such as --compare --project FILE or --captures, and returns
// once it has started. The viewer takes FFmpeg and FFprobe from NR_FFMPEG and NR_FFPROBE, and looks for them
// itself without those.
func View(tools string, t Tools, args ...string) error {
	cmd := exec.Command(filepath.Join(tools, "viewer", "renderer.exe"), args...)
	cmd.Dir = tools
	if t.Ready() {
		cmd.Env = append(os.Environ(), "NR_FFMPEG="+t.FFmpeg, "NR_FFPROBE="+t.FFprobe)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// Replay is what nr-replay says of a replay in its replay.json.
type Replay struct {
	FPS        float64 `json:"fps"`        // of the output frames, at the recorded pace; 0 when unknown
	Multiplier int     `json:"multiplier"` // generated frames per real frame, plus 1
	Frames     []Frame `json:"frames"`     // every output frame, in display order
}

// Frame is an output frame of a replay.
type Frame struct {
	Kind  string `json:"kind"` // real or generated
	Real  int    `json:"real"` // the recorded real frame whose calls made it
	A     int    `json:"a"`    // the real frames it lies between; a real frame has A = B
	B     int    `json:"b"`
	Index int    `json:"index"` // a generated frame's place between them, from 1; 0 for a real frame
	File  string `json:"file"`  // its raw frame's file, "" without raw frames
}

// readReplay reads the replay.json of the replay in dir.
func readReplay(dir string) (*Replay, error) {
	data, err := os.ReadFile(filepath.Join(dir, "replay.json"))
	if err != nil {
		return nil, err
	}
	r := &Replay{}
	return r, json.Unmarshal(data, r)
}

// Replay returns side A's replay, which made its video: the frames of both videos are the same, in the same
// order, as both come from one recording.
func (c *Comparison) Replay() (*Replay, error) { return readReplay(filepath.Join(c.Out(), c.A.name)) }

// Videos returns the two videos of the comparison.
func (c *Comparison) Videos() (a, b string) { return c.video(c.A), c.video(c.B) }

// Still is one output frame of both sides, as lossless PNG images of the frames themselves, where a video
// holds them compressed.
type Still struct {
	Name  string // generated_f4-f5_t0.333, real_f5-f5_t0.000
	A, B  string
	Frame Frame // which output frame it is
}

// Is reports whether f and g are the same output frame of a recording, whichever replay made them.
func (f Frame) Is(g Frame) bool {
	return f.Kind == g.Kind && f.A == g.A && f.B == g.B && f.Index == g.Index
}

// stillsDir is the folder of the stills around the recorded real frame real.
func (c *Comparison) stillsDir(real int) string {
	return filepath.Join(c.Out(), "stills", strconv.Itoa(real))
}

// Stills replays the recorded real frames around real through both sides as raw frames, and writes each
// output frame as a PNG. The replay starts one real frame earlier: a replay's first real frame gets no
// generated frames before it. The raw frames, 66 MB each at 3840 x 2160, are removed once the PNGs exist.
func (c *Comparison) Stills(ctx context.Context, real int, step func(label string), out func(string)) ([]Still, error) {
	if stills := c.MadeStills(real); len(stills) > 0 {
		return stills, nil
	}
	// Whatever an earlier run that failed left is no still.
	c.RemoveStills(real)
	for _, s := range []Side{c.A, c.B} {
		step(s.Label)
		dir := filepath.Join(c.stillsDir(real), s.name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		frames := fmt.Sprintf("%d:%d", max(real-1, 0), real+1)
		if err := pkg.Stream(out, exec.CommandContext(ctx, c.tool("nr-replay"), "--record", c.Recording.Dir, "--out", dir,
			"--generator", s.Generator, "--dll", s.DLL, "--label", s.Label, "--frames", frames, "--raw")); err != nil {
			return nil, err
		}
		r, err := readReplay(dir)
		if err != nil {
			return nil, err
		}
		var numbers []string
		for i := range r.Frames {
			numbers = append(numbers, strconv.Itoa(i))
		}
		if len(numbers) == 0 {
			return nil, fmt.Errorf("the replay through %s made no frames", s.Label)
		}
		if err := pkg.Stream(out, exec.CommandContext(ctx, c.tool("nr-video"), "png", dir, "--frames", strings.Join(numbers, ","),
			"--out", dir, "--ffmpeg", c.FFmpeg)); err != nil {
			return nil, err
		}
		raws, _ := filepath.Glob(filepath.Join(dir, "*.nrraw"))
		for _, raw := range raws {
			os.Remove(raw)
		}
	}
	stills := c.MadeStills(real)
	if len(stills) == 0 {
		return nil, errors.New("the replays made no frame that both sides have")
	}
	return stills, nil
}

// RemoveStills removes the stills around real, as when they cannot be shown.
func (c *Comparison) RemoveStills(real int) { os.RemoveAll(c.stillsDir(real)) }

// MadeStills returns the stills around real made before, in display order: the frames both sides have.
func (c *Comparison) MadeStills(real int) []Still {
	dirA, dirB := filepath.Join(c.stillsDir(real), c.A.name), filepath.Join(c.stillsDir(real), c.B.name)
	r, err := readReplay(dirA)
	if err != nil {
		return nil
	}
	var stills []Still
	for _, f := range r.Frames {
		name := strings.TrimSuffix(f.File, ".nrraw")
		a, b := filepath.Join(dirA, name+".png"), filepath.Join(dirB, name+".png")
		if name == "" || !isFile(a) || !isFile(b) {
			continue
		}
		stills = append(stills, Still{name, a, b, f})
	}
	return stills
}

func isFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

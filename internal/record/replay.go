package record

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

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
	// A project left by an earlier run must not outlive a replay that fails now.
	os.Remove(c.Project())
	for _, s := range []Side{c.A, c.B} {
		step(s.Label)
		replay := exec.CommandContext(ctx, c.tool("nr-replay"), "--record", c.Recording.Dir,
			"--out", filepath.Join(c.Out(), s.name), "--generator", s.Generator, "--dll", s.DLL, "--label", s.Label, "--stream", "-")
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

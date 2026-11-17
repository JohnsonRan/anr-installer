package record

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"anr-installer/internal/pkg"
)

// Tools are the programs nr-record's tools and viewer need, FFmpeg and FFprobe, and Scoop, which installs
// them. A field is the program's path, or "" when it is not installed.
type Tools struct {
	FFmpeg, FFprobe, Scoop string
}

// Ready reports whether FFmpeg and FFprobe are installed.
func (t Tools) Ready() bool { return t.FFmpeg != "" && t.FFprobe != "" }

// FindTools looks for the programs on the PATH, then in Scoop's shim folders: a program installed while the
// installer runs is not on the PATH the installer started with.
func FindTools() Tools {
	dirs := shimDirs()
	return Tools{FFmpeg: look("ffmpeg", dirs), FFprobe: look("ffprobe", dirs), Scoop: look("scoop", dirs)}
}

// shimDirs are the folders Scoop puts its programs' shims in, for the user and for all users.
func shimDirs() []string {
	var dirs []string
	add := func(root string) {
		if root != "" {
			dirs = append(dirs, filepath.Join(root, "shims"))
		}
	}
	add(os.Getenv("SCOOP"))
	if home, err := os.UserHomeDir(); err == nil {
		add(filepath.Join(home, "scoop"))
	}
	add(os.Getenv("SCOOP_GLOBAL"))
	if data := os.Getenv("ProgramData"); data != "" {
		add(filepath.Join(data, "scoop"))
	}
	return dirs
}

func look(name string, dirs []string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	for _, dir := range dirs {
		for _, ext := range []string{".exe", ".cmd"} {
			p := filepath.Join(dir, name+ext)
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				return p
			}
		}
	}
	return ""
}

// InstallFFmpeg installs Scoop's ffmpeg package, which holds FFmpeg and FFprobe, sending each line Scoop
// prints to out.
func InstallFFmpeg(ctx context.Context, scoop string, out func(string)) error {
	return pkg.Stream(out, exec.CommandContext(ctx, scoop, "install", "ffmpeg"))
}

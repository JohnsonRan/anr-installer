package pkg

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The test binary stands in for the programs Stream runs: with PKG_TEST_ROLE set it is one of them.
func TestMain(m *testing.M) {
	switch os.Getenv("PKG_TEST_ROLE") {
	case "emit": // frames on standard output, a redrawn progress line and a summary on standard error
		os.Stdout.WriteString("frame 1\nframe 2\n")
		os.Stderr.WriteString("10%\r90%\r100%\nreplayed 2 frames\n")
		os.Exit(0)
	case "sink": // standard input into the file named by its last argument
		data, _ := io.ReadAll(os.Stdin)
		os.WriteFile(os.Args[len(os.Args)-1], data, 0o644)
		os.Stdout.WriteString("wrote the video\n")
		os.Exit(0)
	case "big": // more than a pipe holds, to a reader that may be gone
		buf := make([]byte, 1<<20)
		for range 64 {
			if _, err := os.Stdout.Write(buf); err != nil {
				os.Exit(5)
			}
		}
		os.Exit(0)
	case "fail":
		os.Stderr.WriteString("no such DLL\n")
		os.Exit(3)
	}
	os.Exit(m.Run())
}

func role(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), "PKG_TEST_ROLE="+name)
	return cmd
}

func TestStreamPipesOneCommandIntoTheNext(t *testing.T) {
	file := filepath.Join(t.TempDir(), "out")
	emit, sink := role("emit"), role("sink", file)
	frames, err := emit.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	sink.Stdin = frames
	var lines []string
	if err := Stream(func(l string) { lines = append(lines, l) }, emit, sink); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(file); string(data) != "frame 1\nframe 2\n" {
		t.Errorf("the second command read %q", data)
	}
	got := strings.Join(lines, "|")
	for _, want := range []string{"100%", "replayed 2 frames", "wrote the video"} {
		if !strings.Contains(got, want) {
			t.Errorf("lines %q lack %q", got, want)
		}
	}
	if strings.Contains(got, "frame 1") || strings.Contains(got, "10%") {
		t.Errorf("lines %q hold piped output or an overwritten progress line", got)
	}
}

func TestStreamReturnsTheFirstFailure(t *testing.T) {
	file := filepath.Join(t.TempDir(), "out")
	fail, sink := role("fail"), role("sink", file)
	frames, _ := fail.StdoutPipe()
	sink.Stdin = frames
	var lines []string
	err := Stream(func(l string) { lines = append(lines, l) }, fail, sink)
	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != 3 {
		t.Fatalf("error %v", err)
	}
	if !strings.Contains(strings.Join(lines, "|"), "no such DLL") {
		t.Errorf("lines %q", lines)
	}
	if err := Stream(func(string) {}, exec.Command(filepath.Join(t.TempDir(), "missing.exe"))); err == nil {
		t.Error("a command that cannot start gave no error")
	}
}

// A reader that exits early must not leave the writer blocked on a full pipe: the writer has to see its
// reader gone.
func TestStreamEndsWhenTheReaderExitsEarly(t *testing.T) {
	big, fail := role("big"), role("fail")
	frames, err := big.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	fail.Stdin = frames
	done := make(chan error, 1)
	go func() { done <- Stream(func(string) {}, big, fail) }()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "exit status 3") {
			t.Errorf("the reader's failure is not reported: %v", err)
		}
	case <-time.After(15 * time.Second):
		big.Process.Kill()
		t.Fatal("Stream hung on a writer whose reader had exited")
	}
}

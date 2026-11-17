package pkg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// Action is what a package's script does.
type Action string

const (
	Install   Action = "install"
	Uninstall Action = "uninstall"
)

// Run runs the package's install or uninstall script in pkgDir for the game folder gameDir, sending each line
// it prints to out. On Windows it runs the PowerShell script the package's .cmd runs, without the .cmd's pause;
// elsewhere the shell script, for Proton.
func Run(ctx context.Context, pkgDir string, a Action, gameDir string, out func(string)) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
			"-File", filepath.Join(pkgDir, string(a)+".ps1"), "-Game", gameDir)
	} else {
		cmd = exec.CommandContext(ctx, "sh", filepath.Join(pkgDir, string(a)+".sh"), gameDir)
	}
	cmd.Dir = pkgDir
	return Stream(out, cmd)
}

// Stream runs the commands at once, without console windows, and sends each line they print to out. A command
// whose Stdout is already set keeps it, so that one command can feed the next. It returns every command's
// failure, each under its program's name: in a pipe, one command's failure makes the others fail too.
func Stream(out func(string), cmds ...*exec.Cmd) error {
	pr, pw := io.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		sc := bufio.NewScanner(pr)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		for sc.Scan() {
			line := strings.TrimRight(decode(sc.Bytes()), " \r")
			// A progress display redraws its line after a carriage return: keep what it showed last.
			if i := strings.LastIndexByte(line, '\r'); i >= 0 {
				line = line[i+1:]
			}
			if line != "" {
				out(line)
			}
		}
		io.Copy(io.Discard, pr)
	}()
	var errs []error
	var started []*exec.Cmd
	for _, cmd := range cmds {
		if cmd.Env == nil {
			cmd.Env = environ()
		}
		if cmd.Stdout == nil {
			cmd.Stdout = pw
		}
		cmd.Stderr = pw
		// A program the command started may outlive it and keep the output open: do not wait for that one.
		cmd.WaitDelay = waitDelay
		hideWindow(cmd)
		if err := cmd.Start(); err != nil {
			errs = append(errs, err)
			break
		}
		// The command reads the pipe through its own handle. While this process keeps one too, the writer
		// never learns that its reader has gone, and blocks for good once the pipe is full.
		if f, ok := cmd.Stdin.(*os.File); ok {
			f.Close()
		}
		started = append(started, cmd)
	}
	for _, cmd := range started {
		if len(started) < len(cmds) {
			cmd.Process.Kill() // one did not start: stop those that did
		}
		if err := cmd.Wait(); err != nil && !errors.Is(err, exec.ErrWaitDelay) {
			errs = append(errs, fmt.Errorf("%s: %w", filepath.Base(cmd.Path), err))
		}
	}
	pw.Close()
	<-done
	return errors.Join(errs...)
}

// waitDelay is how long Stream waits for a command's output to close after the command has exited.
const waitDelay = 5 * time.Second

// environ is the environment without PSModulePath: Windows PowerShell 5.1 started from PowerShell 7 inherits
// its PSModulePath and then misses its own modules (Get-FileHash is not recognized), so it builds its default.
func environ() []string {
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(kv), "PSMODULEPATH=") {
			env = append(env, kv)
		}
	}
	return env
}

// decode turns a line of console output into UTF-8: Windows PowerShell 5 writes in the console's ANSI code page,
// which is GBK on Chinese Windows.
func decode(b []byte) string {
	if utf8.Valid(b) || runtime.GOOS != "windows" {
		return string(b)
	}
	if s, err := simplifiedchinese.GBK.NewDecoder().Bytes(b); err == nil {
		return string(s)
	}
	return string(b)
}

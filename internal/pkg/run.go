package pkg

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	// Windows PowerShell 5.1 started from PowerShell 7 inherits its PSModulePath and then misses its own
	// modules (Get-FileHash is not recognized): let it build its default.
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(kv), "PSMODULEPATH=") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	hideWindow(cmd)
	pr, pw := io.Pipe()
	cmd.Stdout, cmd.Stderr = pw, pw
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		sc := bufio.NewScanner(pr)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		for sc.Scan() {
			if line := strings.TrimRight(decode(sc.Bytes()), " \r"); line != "" {
				out(line)
			}
		}
		io.Copy(io.Discard, pr)
	}()
	err := cmd.Wait()
	pw.Close()
	<-done
	return err
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

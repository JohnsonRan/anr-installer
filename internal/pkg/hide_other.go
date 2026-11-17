//go:build !windows

package pkg

import "os/exec"

func hideWindow(*exec.Cmd) {}

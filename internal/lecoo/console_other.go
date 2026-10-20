//go:build !windows

package lecoo

import "os/exec"

func hideConsoleWindow(cmd *exec.Cmd) {}

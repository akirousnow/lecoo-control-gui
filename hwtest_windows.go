package main

import (
	"os/exec"
	"syscall"
)

func runHwtest(path string) error {
	cmd := exec.Command(path, "hwtest")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000010}
	return cmd.Start()
}

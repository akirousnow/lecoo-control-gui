//go:build !windows

package main

import "errors"

func runHwtest(path string) error {
	return errors.New("硬件自检只在 Windows 上可用")
}

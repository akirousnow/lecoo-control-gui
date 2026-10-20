//go:build !windows

package lecoo

import "os/exec"

type killJob struct{}

func newKillJob() (*killJob, error) {
	return &killJob{}, nil
}

func (j *killJob) attach(cmd *exec.Cmd) error {
	return nil
}

func (j *killJob) close() {}

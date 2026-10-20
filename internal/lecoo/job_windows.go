package lecoo

import (
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

type killJob struct {
	handle windows.Handle
}

func newKillJob() (*killJob, error) {
	handle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(handle, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		windows.CloseHandle(handle)
		return nil, err
	}
	return &killJob{handle: handle}, nil
}

func (j *killJob) attach(cmd *exec.Cmd) error {
	if j == nil || cmd == nil || cmd.Process == nil {
		return nil
	}
	process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(process)
	return windows.AssignProcessToJobObject(j.handle, process)
}

func (j *killJob) close() {
	if j == nil || j.handle == 0 {
		return
	}
	windows.CloseHandle(j.handle)
	j.handle = 0
}

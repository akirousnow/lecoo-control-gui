package main

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	wsExToolWindow = 0x00000080
	wsExAppWindow  = 0x00040000

	swpNoSize       = 0x0001
	swpNoMove       = 0x0002
	swpNoZOrder     = 0x0004
	swpFrameChanged = 0x0020
)

var gwlExStyle = uintptr(0xFFFFFFEC)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procGetWindowLongPtr         = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtr         = user32.NewProc("SetWindowLongPtrW")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
)

func appWindowHandle() syscall.Handle {
	pid := uint32(os.Getpid())
	var found syscall.Handle
	callback := syscall.NewCallback(func(hwnd syscall.Handle, _ uintptr) uintptr {
		var owner uint32
		procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&owner)))
		if owner != pid {
			return 1
		}
		visible, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
		if visible == 0 {
			return 1
		}
		found = hwnd
		return 0
	})
	procEnumWindows.Call(callback, 0)
	return found
}

func setTaskbarVisible(visible bool) {
	hwnd := appWindowHandle()
	if hwnd == 0 {
		return
	}
	current, _, _ := procGetWindowLongPtr.Call(uintptr(hwnd), gwlExStyle)
	style := current
	if visible {
		style &^= wsExToolWindow
		style |= wsExAppWindow
	} else {
		style |= wsExToolWindow
		style &^= wsExAppWindow
	}
	if style == current {
		return
	}
	procSetWindowLongPtr.Call(uintptr(hwnd), gwlExStyle, style)
	procSetWindowPos.Call(uintptr(hwnd), 0, 0, 0, 0, 0, swpNoMove|swpNoSize|swpNoZOrder|swpFrameChanged)
}

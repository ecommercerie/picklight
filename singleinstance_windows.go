//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const mutexName = "PickLight_SingleInstance_Mutex"

var showWindowMsg uint32

func init() {
	msgName, _ := syscall.UTF16PtrFromString("PickLight_ShowWindow")
	ret, _, _ := user32.NewProc("RegisterWindowMessageW").Call(uintptr(unsafe.Pointer(msgName)))
	showWindowMsg = uint32(ret)
}

func acquireSingleInstance() bool {
	name, _ := syscall.UTF16PtrFromString(mutexName)
	_, _, err := kernel32.NewProc("CreateMutexW").Call(0, 0, uintptr(unsafe.Pointer(name)))

	const errorAlreadyExists = 183
	if errno, ok := err.(syscall.Errno); ok && errno == errorAlreadyExists {
		className, _ := syscall.UTF16PtrFromString("PickLightTray")
		hwnd, _, _ := user32.NewProc("FindWindowW").Call(uintptr(unsafe.Pointer(className)), 0)
		if hwnd != 0 {
			msgName, _ := syscall.UTF16PtrFromString("PickLight_ShowWindow")
			wmShow, _, _ := user32.NewProc("RegisterWindowMessageW").Call(uintptr(unsafe.Pointer(msgName)))
			if wmShow != 0 {
				user32.NewProc("PostMessageW").Call(hwnd, wmShow, 0, 0)
			}
		}
		fmt.Println("PickLight is already running. Bringing existing window to front.")
		return false
	}
	return true
}

func isShowWindowMessage(umsg uint32) bool {
	return showWindowMsg != 0 && umsg == showWindowMsg
}

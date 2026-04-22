//go:build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	advapi32       = syscall.NewLazyDLL("advapi32.dll")
	pRegOpenKeyEx  = advapi32.NewProc("RegOpenKeyExW")
	pRegCloseKey   = advapi32.NewProc("RegCloseKey")
	pRegSetValueEx = advapi32.NewProc("RegSetValueExW")
	pRegDeleteValue = advapi32.NewProc("RegDeleteValueW")
	pRegQueryValueEx = advapi32.NewProc("RegQueryValueExW")
)

const (
	hkeyCurrentUser = 0x80000001
	keyWrite        = 0x20006
	keyRead         = 0x20019
	regSZ           = 1
)

const regRunPath = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
const regValueName = "PickLight"

func isStartupEnabled() bool {
	keyPtr, _ := syscall.UTF16PtrFromString(regRunPath)
	var hKey syscall.Handle
	ret, _, _ := pRegOpenKeyEx.Call(hkeyCurrentUser, uintptr(unsafe.Pointer(keyPtr)), 0, keyRead, uintptr(unsafe.Pointer(&hKey)))
	if ret != 0 {
		return false
	}
	defer pRegCloseKey.Call(uintptr(hKey))

	valPtr, _ := syscall.UTF16PtrFromString(regValueName)
	ret, _, _ = pRegQueryValueEx.Call(uintptr(hKey), uintptr(unsafe.Pointer(valPtr)), 0, 0, 0, 0)
	return ret == 0
}

func setStartupEnabled(enabled bool) error {
	keyPtr, _ := syscall.UTF16PtrFromString(regRunPath)
	var hKey syscall.Handle
	ret, _, err := pRegOpenKeyEx.Call(hkeyCurrentUser, uintptr(unsafe.Pointer(keyPtr)), 0, keyWrite, uintptr(unsafe.Pointer(&hKey)))
	if ret != 0 {
		return fmt.Errorf("RegOpenKeyEx: %w", err)
	}
	defer pRegCloseKey.Call(uintptr(hKey))

	valPtr, _ := syscall.UTF16PtrFromString(regValueName)

	if enabled {
		exePath, _ := os.Executable()
		exePathW, _ := syscall.UTF16FromString(exePath)
		dataSize := uint32(len(exePathW) * 2)
		ret, _, err = pRegSetValueEx.Call(uintptr(hKey), uintptr(unsafe.Pointer(valPtr)), 0, regSZ, uintptr(unsafe.Pointer(&exePathW[0])), uintptr(dataSize))
		if ret != 0 {
			return fmt.Errorf("RegSetValueEx: %w", err)
		}
	} else {
		pRegDeleteValue.Call(uintptr(hKey), uintptr(unsafe.Pointer(valPtr)))
	}
	return nil
}

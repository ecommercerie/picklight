//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"picklight/internal/updater"
)

// applyUpdate downloads the new exe, writes a batch script that waits for
// this process to exit, replaces the exe, and restarts it.
// The script is launched elevated (UAC) since Program Files requires admin.
func applyUpdate(status updater.UpdateStatus) error {
	if status.DownloadURL == "" {
		return fmt.Errorf("no download URL available")
	}

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable: %w", err)
	}

	tmpExe := filepath.Join(os.TempDir(), "picklight-update.exe")
	if err := updater.DownloadTo(status.DownloadURL, tmpExe); err != nil {
		return err
	}

	batPath := filepath.Join(os.TempDir(), "picklight-update.bat")
	batContent := fmt.Sprintf(`@echo off
timeout /t 3 /nobreak >nul
copy /y "%s" "%s" >nul
del "%s" >nul
start "" "%s"
del "%%~f0"
`, tmpExe, currentExe, tmpExe, currentExe)

	if err := os.WriteFile(batPath, []byte(batContent), 0644); err != nil {
		return fmt.Errorf("write update script: %w", err)
	}

	shell32dll := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32dll.NewProc("ShellExecuteW")

	verb, _ := syscall.UTF16PtrFromString("runas")
	exe, _ := syscall.UTF16PtrFromString("cmd")
	args, _ := syscall.UTF16PtrFromString("/C \"" + batPath + "\"")

	ret, _, _ := shellExecute.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(exe)),
		uintptr(unsafe.Pointer(args)),
		0,
		0, // SW_HIDE
	)
	if ret <= 32 {
		return fmt.Errorf("failed to launch elevated update script (code %d)", ret)
	}

	return nil
}

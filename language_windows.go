//go:build windows

package main

import "syscall"

var kernel32Lang = syscall.NewLazyDLL("kernel32.dll")

// detectOSLanguage returns "fr" or "en" based on the Windows UI language.
func detectOSLanguage() string {
	proc := kernel32Lang.NewProc("GetUserDefaultUILanguage")
	langID, _, _ := proc.Call()
	// Primary language is the low 10 bits
	primary := langID & 0x3FF
	if primary == 0x0C { // LANG_FRENCH
		return "fr"
	}
	return "en"
}

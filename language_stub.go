//go:build !windows

package main

import (
	"os"
	"strings"
)

// detectOSLanguage returns "fr" or "en" based on the LANG environment variable.
func detectOSLanguage() string {
	lang := os.Getenv("LANG")
	if strings.HasPrefix(strings.ToLower(lang), "fr") {
		return "fr"
	}
	return "en"
}

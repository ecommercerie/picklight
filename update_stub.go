//go:build !windows

package main

import (
	"fmt"

	"picklight/internal/updater"
)

func applyUpdate(status updater.UpdateStatus) error {
	return fmt.Errorf("auto-update is only supported on Windows")
}

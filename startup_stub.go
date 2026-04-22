//go:build !windows

package main

func isStartupEnabled() bool            { return false }
func setStartupEnabled(enabled bool) error { return nil }

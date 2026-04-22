//go:build !windows

package main

func acquireSingleInstance() bool          { return true }
func isShowWindowMessage(umsg uint32) bool { return false }

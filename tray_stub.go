//go:build !windows

package main

func (a *App) initTray()       {}
func (a *App) cleanupTray()    {}
func updateTrayTip(tip string) {}

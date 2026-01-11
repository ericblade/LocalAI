//go:build windows
// +build windows

package main

import (
	"syscall"
)

func loadGosdLib() uintptr {
	lib, err := syscall.LoadLibrary("gosd.dll")
	if err != nil {
		panic(err)
	}
	return uintptr(lib)
}

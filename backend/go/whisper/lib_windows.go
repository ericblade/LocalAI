//go:build windows
// +build windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
)

func loadWhisperLib() uintptr {
	// Get library name from environment variable, default to fallback
	libName := os.Getenv("WHISPER_LIBRARY")
	if libName == "" {
		libName = "libgowhisper.dll"
	}

	lib, err := syscall.LoadLibrary(libName)
	if err != nil {
		// Try adding the DLL directory to the search path
		dllDir := filepath.Dir(libName)
		if dllDir != "" && dllDir != "." {
			os.Setenv("PATH", dllDir+";"+os.Getenv("PATH"))
			lib, err = syscall.LoadLibrary(filepath.Base(libName))
		}
		if err != nil {
			panic(err)
		}
	}
	return uintptr(lib)
}

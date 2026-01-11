//go:build !windows
// +build !windows

package main

import (
	"os"

	"github.com/ebitengine/purego"
)

func loadWhisperLib() uintptr {
	// Get library name from environment variable, default to fallback
	libName := os.Getenv("WHISPER_LIBRARY")
	if libName == "" {
		libName = "./libgowhisper-fallback.so"
	}

	gosd, err := purego.Dlopen(libName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}
	return gosd
}

//go:build !windows
// +build !windows

package main

import (
	"github.com/ebitengine/purego"
)

func loadGosdLib() uintptr {
	gosd, err := purego.Dlopen("./libgosd.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}
	return gosd
}

//go:build !cgo
// +build !cgo

package main

import (
	"flag"
	"fmt"
	"os"
)

// version, commit, date are injected at build time (see .goreleaser.yaml).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var (
		showVersion bool
		noUpdate    bool
	)

	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&noUpdate, "no-update", false, "disable update checks")
	flag.Parse()

	if showVersion {
		fmt.Printf("Survive It %s (%s) %s\n", version, commit, date)
		return
	}

	_ = noUpdate
	fmt.Fprintln(os.Stderr, "Survive It now requires the 3D client build (cgo/raylib enabled).")
	os.Exit(1)
}

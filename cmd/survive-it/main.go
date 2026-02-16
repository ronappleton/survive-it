package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/appengine-ltd/survive-it/internal/ui"
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

	app := ui.NewApp(ui.AppConfig{
		Version:  version,
		Commit:   commit,
		BuildDate: date,
		NoUpdate: noUpdate,
	})

	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/appengine-ltd/survive-it/internal/profilegen"
)

func main() {
	var bboxRaw string
	var outPath string
	var id string
	var name string
	var cellMeters int
	var source string

	flag.StringVar(&bboxRaw, "bbox", "", "bbox as minLon,minLat,maxLon,maxLat")
	flag.StringVar(&outPath, "out", "", "output path for profile JSON")
	flag.StringVar(&id, "id", "", "profile id (defaults from output filename)")
	flag.StringVar(&name, "name", "", "profile display name")
	flag.IntVar(&cellMeters, "cell", 100, "cell size in meters")
	flag.StringVar(&source, "source", "", "source note override")
	flag.Parse()

	if strings.TrimSpace(bboxRaw) == "" {
		die("--bbox is required")
	}
	if strings.TrimSpace(outPath) == "" {
		die("--out is required")
	}
	bbox, err := profilegen.ParseBBox(bboxRaw)
	if err != nil {
		die(err.Error())
	}
	if strings.TrimSpace(id) == "" {
		base := strings.TrimSuffix(filepath.Base(outPath), filepath.Ext(outPath))
		id = strings.TrimSpace(base)
	}
	if strings.TrimSpace(id) == "" {
		die("unable to derive id; set --id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	profile, err := profilegen.GenerateProfile(ctx, profilegen.Options{
		ID:         id,
		Name:       name,
		BBox:       bbox,
		CellMeters: cellMeters,
		CacheRoot:  filepath.Join(".cache", "genprofile"),
		Source:     source,
	})
	if err != nil {
		die(fmt.Sprintf("generate profile: %v", err))
	}
	if err := profilegen.WriteProfile(outPath, profile); err != nil {
		die(fmt.Sprintf("write profile: %v", err))
	}
	fmt.Printf("wrote %s\n", outPath)
	fmt.Printf("profile=%s elev(p10/p50/p90)=%.2f/%.2f/%.2f slope(p50/p90)=%.2f/%.2f river=%.3f lake=%.3f\n",
		profile.ID,
		profile.ElevP10,
		profile.ElevP50,
		profile.ElevP90,
		profile.SlopeP50,
		profile.SlopeP90,
		profile.RiverDensity,
		profile.LakeCoverage,
	)
}

func die(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

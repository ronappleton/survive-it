package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/appengine-ltd/survive-it/internal/game"
	"github.com/appengine-ltd/survive-it/internal/profilegen"
)

func main() {
	var force bool
	var only string
	var cellMeters int

	flag.BoolVar(&force, "force", false, "regenerate profiles even if JSON exists")
	flag.StringVar(&only, "only", "", "generate only a specific profile id")
	flag.IntVar(&cellMeters, "cell", 100, "cell size in meters")
	flag.Parse()

	scenarios := game.BuiltInScenarios()
	type job struct {
		ProfileID string
		Name      string
		BBox      [4]float64
	}
	jobsByID := map[string]job{}
	for _, s := range scenarios {
		if s.LocationMeta == nil {
			continue
		}
		id := strings.TrimSpace(s.LocationMeta.ProfileID)
		if id == "" {
			continue
		}
		if only != "" && only != id {
			continue
		}
		if _, exists := jobsByID[id]; exists {
			continue
		}
		name := strings.TrimSpace(s.LocationMeta.Name)
		if name == "" {
			name = strings.TrimSpace(s.Name)
		}
		jobsByID[id] = job{ProfileID: id, Name: name, BBox: s.LocationMeta.BBox}
	}

	if len(jobsByID) == 0 {
		fmt.Println("no scenario profiles to generate")
		return
	}

	ids := make([]string, 0, len(jobsByID))
	for id := range jobsByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	if err := os.MkdirAll(filepath.Join("assets", "profiles"), 0o755); err != nil {
		die(err)
	}
	cacheRoot := filepath.Join(".cache", "genprofile")
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		die(err)
	}

	wrote := 0
	skipped := 0
	failed := 0
	for _, id := range ids {
		j := jobsByID[id]
		outPath := filepath.Join("assets", "profiles", id+".json")
		if !force {
			if _, err := os.Stat(outPath); err == nil {
				fmt.Printf("skip %s (exists)\n", outPath)
				skipped++
				continue
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		profile, err := profilegen.GenerateProfile(ctx, profilegen.Options{
			ID:         id,
			Name:       j.Name,
			BBox:       j.BBox,
			CellMeters: cellMeters,
			CacheRoot:  cacheRoot,
		})
		cancel()
		if err != nil {
			fmt.Printf("fail %s: %v\n", id, err)
			failed++
			continue
		}
		if err := profilegen.WriteProfile(outPath, profile); err != nil {
			fmt.Printf("fail write %s: %v\n", outPath, err)
			failed++
			continue
		}
		fmt.Printf("wrote %s\n", outPath)
		wrote++
	}

	fmt.Printf("done wrote=%d skipped=%d failed=%d\n", wrote, skipped, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

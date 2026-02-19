package game

import (
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestLoadGenProfileMissingReturnsFalse(t *testing.T) {
	t.Setenv("SURVIVE_IT_PROFILE_DIR", t.TempDir())
	if profile, ok := LoadGenProfile("does_not_exist"); ok || profile != nil {
		t.Fatalf("expected missing profile to return nil,false")
	}
}

func TestLoadGenProfileFromFixture(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SURVIVE_IT_PROFILE_DIR", dir)
	blob := []byte(`{
  "id": "fixture",
  "name": "Fixture",
  "cell_meters": 100,
  "elev_p10": -24,
  "elev_p50": -4,
  "elev_p90": 26,
  "slope_p50": 3.2,
  "slope_p90": 8.1,
  "ruggedness": 0.62,
  "river_density": 0.064,
  "lake_coverage": 0.018
}
`)
	path := filepath.Join(dir, "fixture.json")
	if err := os.WriteFile(path, blob, 0o644); err != nil {
		t.Fatalf("write profile fixture: %v", err)
	}
	profile, ok := LoadGenProfile("fixture")
	if !ok {
		t.Fatalf("expected fixture profile to load")
	}
	if profile.ID != "fixture" || profile.Name != "Fixture" {
		t.Fatalf("unexpected profile payload: %+v", profile)
	}
	if profile.RiverDensity <= 0 {
		t.Fatalf("expected river density > 0")
	}
}

func TestGenerateWorldTopologyWithProfileDeterministic(t *testing.T) {
	profile := &GenProfile{
		ID:           "deterministic",
		Name:         "Deterministic",
		CellMeters:   100,
		ElevP10:      -20,
		ElevP50:      2,
		ElevP90:      34,
		SlopeP50:     4,
		SlopeP90:     10,
		Ruggedness:   0.9,
		RiverDensity: 0.07,
		LakeCoverage: 0.015,
	}
	a := GenerateWorldTopologyWithProfile(811, "temperate_rainforest", 64, 64, profile)
	b := GenerateWorldTopologyWithProfile(811, "temperate_rainforest", 64, 64, profile)
	if a.Width != b.Width || a.Height != b.Height || len(a.Cells) != len(b.Cells) {
		t.Fatalf("topology dimensions mismatch")
	}
	for i := range a.Cells {
		if a.Cells[i] != b.Cells[i] {
			t.Fatalf("cell mismatch at %d: %+v vs %+v", i, a.Cells[i], b.Cells[i])
		}
	}
}

func TestGenerateWorldTopologyWithProfileTargets(t *testing.T) {
	profile := &GenProfile{
		ID:           "targets",
		Name:         "Targets",
		CellMeters:   100,
		ElevP10:      -18,
		ElevP50:      5,
		ElevP90:      30,
		SlopeP50:     3.8,
		SlopeP90:     10.5,
		Ruggedness:   0.82,
		RiverDensity: 0.08,
		LakeCoverage: 0.02,
	}
	topo := GenerateWorldTopologyWithProfile(90210, "temperate_rainforest", 72, 72, profile)
	if len(topo.Cells) == 0 {
		t.Fatalf("expected topology cells")
	}
	elev := make([]float64, len(topo.Cells))
	riverCount := 0
	for i, cell := range topo.Cells {
		elev[i] = float64(cell.Elevation)
		if cell.Flags&TopoFlagRiver != 0 {
			riverCount++
		}
	}
	sort.Float64s(elev)
	p10 := percentileFromSorted(elev, 0.10)
	p50 := percentileFromSorted(elev, 0.50)
	p90 := percentileFromSorted(elev, 0.90)
	if math.Abs(p10-profile.ElevP10) > 8 {
		t.Fatalf("p10 out of tolerance: got %.2f want %.2f", p10, profile.ElevP10)
	}
	if math.Abs(p50-profile.ElevP50) > 6 {
		t.Fatalf("p50 out of tolerance: got %.2f want %.2f", p50, profile.ElevP50)
	}
	if math.Abs(p90-profile.ElevP90) > 8 {
		t.Fatalf("p90 out of tolerance: got %.2f want %.2f", p90, profile.ElevP90)
	}
	riverDensity := float64(riverCount) / float64(len(topo.Cells))
	if math.Abs(riverDensity-profile.RiverDensity) > 0.03 {
		t.Fatalf("river density out of tolerance: got %.4f want %.4f", riverDensity, profile.RiverDensity)
	}
}

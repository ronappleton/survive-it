package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// GenProfile stores compact terrain statistics distilled from a real-world area.
// Runtime generation uses only these local profiles; raw elevation downloads are build-time only.
type GenProfile struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	CellMeters   int     `json:"cell_meters"`
	ElevP10      float64 `json:"elev_p10"`
	ElevP50      float64 `json:"elev_p50"`
	ElevP90      float64 `json:"elev_p90"`
	SlopeP50     float64 `json:"slope_p50"`
	SlopeP90     float64 `json:"slope_p90"`
	Ruggedness   float64 `json:"ruggedness"`
	RiverDensity float64 `json:"river_density"`
	LakeCoverage float64 `json:"lake_coverage"`
	Notes        string  `json:"notes,omitempty"`
	Source       string  `json:"source,omitempty"`
}

func DefaultGenProfile() *GenProfile {
	return &GenProfile{
		ID:           "default_procedural",
		Name:         "Default Procedural Profile",
		CellMeters:   100,
		ElevP10:      -36,
		ElevP50:      -8,
		ElevP90:      26,
		SlopeP50:     3.2,
		SlopeP90:     9.0,
		Ruggedness:   0.58,
		RiverDensity: 0.055,
		LakeCoverage: 0.03,
		Notes:        "Fallback profile matching legacy procedural terrain behavior.",
		Source:       "internal default",
	}
}

func LoadGenProfile(profileID string) (*GenProfile, bool) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return nil, false
	}
	path := profilePath(profileID)
	blob, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var profile GenProfile
	if err := json.Unmarshal(blob, &profile); err != nil {
		return nil, false
	}
	if profile.ID == "" {
		profile.ID = profileID
	}
	if profile.CellMeters <= 0 {
		profile.CellMeters = 100
	}
	normalizeGenProfile(&profile)
	return &profile, true
}

func LoadScenarioGenProfile(s Scenario) (*GenProfile, bool) {
	if s.LocationMeta == nil {
		return nil, false
	}
	if strings.TrimSpace(s.LocationMeta.ProfileID) == "" {
		return nil, false
	}
	return LoadGenProfile(s.LocationMeta.ProfileID)
}

func profilePath(profileID string) string {
	return filepath.Join(profileDir(), profileID+".json")
}

func profileDir() string {
	override := strings.TrimSpace(os.Getenv("SURVIVE_IT_PROFILE_DIR"))
	if override != "" {
		return override
	}
	return filepath.Join("assets", "profiles")
}

func normalizeGenProfile(profile *GenProfile) {
	if profile == nil {
		return
	}
	if profile.ElevP10 > profile.ElevP50 {
		profile.ElevP10 = profile.ElevP50 - 1
	}
	if profile.ElevP90 < profile.ElevP50 {
		profile.ElevP90 = profile.ElevP50 + 1
	}
	if profile.ElevP90 < profile.ElevP10+1 {
		profile.ElevP90 = profile.ElevP10 + 1
	}
	profile.SlopeP50 = clampFloat(profile.SlopeP50, 0.1, 35)
	profile.SlopeP90 = clampFloat(profile.SlopeP90, profile.SlopeP50+0.1, 55)
	profile.Ruggedness = clampFloat(profile.Ruggedness, 0.05, 3.0)
	profile.RiverDensity = clampFloat(profile.RiverDensity, 0.001, 0.35)
	profile.LakeCoverage = clampFloat(profile.LakeCoverage, 0, 0.35)
}

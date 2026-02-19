package game

// Discovery summary:
// - Scenario is the central config carried into RunState and consumed by weather/topology/resource systems.
// - LocationMeta is already optional and runtime-only, so climate can be added the same way.
// - Keeping Climate optional preserves backwards compatibility for custom scenarios.
type Scenario struct {
	ID                 ScenarioID
	Name               string
	Location           string
	LocationMeta       *ScenarioLocation `json:"-"`
	Climate            *ClimateProfile   `json:"-"`
	Biome              string
	MapWidthCells      int
	MapHeightCells     int
	Wildlife           []string
	Description        string
	Daunting           string
	Motivation         string
	SupportedModes     []GameMode
	DefaultDays        int
	IssuedKit          IssuedKit
	SeasonSets         []SeasonSet
	DefaultSeasonSetID SeasonSetID
}

type ScenarioLocation struct {
	Name      string
	BBox      [4]float64
	ProfileID string
}

type IssuedKit struct {
	Pot         bool
	Firestarter bool
}

const (
	ScenarioVancouverIslandID ScenarioID = "vancouver_island"
	ScenarioJungleID          ScenarioID = "jungle"
	ScenarioArcticID          ScenarioID = "arctic"
	ScenarioRandomID          ScenarioID = "random"
)

type ScenarioID string

type SeasonID string

const (
	SeasonSetAloneDefaultID  = "alone_default"
	SeasonSetWetDefaultID    = "wet_default"
	SeasonSetWinterDefaultID = "winter_default"
	SeasonSetDryDefaultID    = "dry_default"
)

type SeasonSetID string

const (
	SeasonAutumn = "autumn"
	SeasonWinter = "winter"
	SeasonWet    = "wet"
	SeasonDry    = "dry"
)

type Season struct {
	ID SeasonID
}

type SeasonPhase struct {
	Season SeasonID
	Days   int // 0 = till end
}

type SeasonSet struct {
	ID     SeasonSetID
	Phases []SeasonPhase
}

var externalScenarios []Scenario

func SetExternalScenarios(scenarios []Scenario) {
	externalScenarios = append([]Scenario(nil), scenarios...)
}

func ExternalScenarios() []Scenario {
	return append([]Scenario(nil), externalScenarios...)
}

func AllScenarios() []Scenario {
	base := BuiltInScenarios()
	if len(externalScenarios) == 0 {
		return base
	}
	out := make([]Scenario, 0, len(base)+len(externalScenarios))
	out = append(out, base...)
	out = append(out, externalScenarios...)
	return out
}

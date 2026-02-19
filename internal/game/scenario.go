package game

type Scenario struct {
	ID                 ScenarioID
	Name               string
	Location           string
	Biome              string
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

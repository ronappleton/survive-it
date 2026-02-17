package game

type Scenario struct {
	ID                 ScenarioID
	Name               string
	Biome              string
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

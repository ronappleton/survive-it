# Weather, Physiology, and Daily Effects

## Weather Model

Primary files:

- `internal/game/environment.go`
- `internal/game/weather_state.go`
- `internal/game/weather_effects.go`

Weather is deterministic by:

- run seed
- biome
- season
- day

Weather types:

- sunny, clear, cloudy
- rain, heavy_rain, storm
- snow, blizzard
- windy, heatwave

## Temperature

- biome range from `TemperatureRangeForBiome`
- per-day base temp from deterministic hash
- season modifier and weather-type modifier applied

## Season Resolution

`internal/game/season_resolver.go`:

- scenario season sets are phase-based (`SeasonPhase`)
- `Days == 0` means "until end"
- current season derived from run day

## Daily Player Effects

`AdvanceDay` applies:

- base weather impact
- temperature stress impact
- weather streak impact
- biome special-case impact
- player stat adjustments (endurance/bushcraft/mental)
- crafted clothing/kit weather modifiers
- camp impacts from shelter/fire
- ailments and deficiency/dehydration penalties

## Physiology and Metabolism

Files:

- `internal/game/physiology.go`
- `internal/game/metabolism.go`
- `internal/game/metabolism_realtime.go`
- `internal/game/player_decay.go`

Model includes:

- nutrition reserves (calories/protein/fat/sugar)
- effect bars (hunger/thirst/fatigue)
- realtime fractional consumption through day
- daily deficiency streak tracking
- malnutrition/dehydration ailment triggers

## Ailments and Disease Risk

- animal disease risk metadata is defined in `internal/game/animals.go`
- disease application logic is in `internal/game/food_simulation.go`
- ailment penalties are applied daily in `advance_day.go`

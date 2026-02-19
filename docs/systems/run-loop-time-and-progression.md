# Run Loop, Time, and Progression

## Runtime Loop

- UI frame loop is in `internal/gui/app.go` (`Run`, `update`, `draw`).
- Active gameplay uses `screenRun` and related run sub-screens.

## Entering a Run

- `internal/game/state.go` `NewRunState`:
  - validates config
  - resolves scenario
  - creates players
  - initializes weather
  - initializes runtime player stats
  - initializes topology

## Clock and Day Progression

- In-run elapsed real time is tracked by `runPlayedFor` in `internal/gui/app.go`.
- Auto-day duration is based on options (`Game Hours Per Day`).
- `ApplyRealtimeMetabolism` consumes partial-day reserves continuously.
- When a day completes, `AdvanceDay` runs and weather/day messages are logged.

## `AdvanceDay` Pipeline

`internal/game/advance_day.go`:

1. Consume pending metabolism remainder.
2. Increment day and refresh weather.
3. Compute weather impact (`weatherImpactForDay`) and camp impact.
4. Per-player:
   - weather + crafted-clothing modifiers
   - camp effects
   - ailment penalties
   - deficiency/dehydration effects
   - clamp and refresh effect bars
5. Camp progression and food degradation.
6. Wildlife cell-state decay (`hunt pressure`, `disturbance`, `depletion`, `carcass token`).

## Progression and Skill Growth

- Skill growth helper: `applySkillEffort` in `internal/game/player_progression.go`.
- Skills are advanced during actions (crafting, hunting, fishing, foraging, gathering, trap operations).
- Traits are represented as signed modifiers (`TraitModifier`) and are applied to action quality/chance calculations.

## Run Outcome

`EvaluateRun` (`internal/game/advance_day.go`) returns:

- `ongoing`
- `completed` (for fixed day-length runs)
- `critical` (player at zero energy/hydration or max hunger/thirst/fatigue)

# World Topology, Map, Fog, and Encounters

## Topology Grid

Source: `internal/game/topology.go`.

Each 100m cell is `TopoCell`:

- elevation, moisture, temperature
- biome enum
- flags (`water`, `river`, `lake`, `coast`)
- roughness (movement cost influence)

## Real-World Terrain Profiles (Offline Runtime)

Source: `internal/game/gen_profile.go`, `cmd/genprofile`, `cmd/genprofiles`.

`GenProfile` files under `assets/profiles/*.json` are compact terrain summaries distilled from real-world elevation samples:

- `elev_p10/p50/p90`
- `slope_p50/p90`
- `ruggedness`
- `river_density`
- `lake_coverage`

Runtime is fully offline:

- the game only loads local JSON profiles via `LoadGenProfile`
- no map/elevation network calls are made while playing
- if a profile is missing, `DefaultGenProfile()` keeps generation backward compatible

## Build-Time Profile Generation

One-time developer tools:

- `go run ./cmd/genprofile --bbox "minLon,minLat,maxLon,maxLat" --out "assets/profiles/<id>.json" --cell 100`
- `go run ./cmd/genprofiles`

Notes:

- raw download/cache data goes to `.cache/genprofile/...`
- only distilled `assets/profiles/*.json` should be committed

## Adding a New Real-World Scenario

1. Set scenario metadata in `internal/game/scenarios_builtin.go`:
   - `LocationMeta.Name`
   - `LocationMeta.BBox`
   - `LocationMeta.ProfileID`
2. Run `go run ./cmd/genprofiles` (or `cmd/genprofile` for one profile).
3. Commit `assets/profiles/<profile_id>.json`.

## Map Size by Mode and Scenario

Sizing function: `topologySizeForScenario`.

- `Isolation Protocol`: default `36x36` (clamped `28..46` per axis)
- `Naked & Afraid`: default `100x100` (clamped `88..125`)
- `Naked & Afraid XL`: default `125x125` (clamped `100..150`)

Scenario-level fields:

- `Scenario.MapWidthCells`
- `Scenario.MapHeightCells`

## Generation Pipeline

`GenerateWorldTopologyWithProfile` builds deterministic terrain from seed and profile:

1. layered value noise for elevation
2. profile-aware percentile mapping (`p10/p50/p90`)
3. moisture and temperature maps with biome biases
4. biome assignment by thresholds
5. downhill flow accumulation and profile-tuned river threshold
6. lake coverage expansion toward profile target
7. coast/water flags and roughness assignment

## Fog of War

- Fog mask is stored in `RunState.FogMask`.
- `Isolation Protocol`: unrevealed at start, reveal persists permanently.
- `Paired Exposure` and `Expedition Survival`: fully revealed.
- Reveal call: `RevealFog(x,y,radius)`.

## Movement + Terrain Cost

Source: `internal/game/travel.go`.

Per-step movement cost includes:

- roughness multiplier
- slope delta penalty
- water crossing penalty (reduced by watercraft)

Movement updates:

- map position and travel totals
- clock advancement via action hours
- player energy/hydration/morale costs
- fog reveal
- per-cell state action effects
- encounter checks

## Watercraft Movement Interaction

Crafted boats modify travel speed and water traversal cost:

- `dugout_canoe`
- `reed_coracle`
- `brush_raft`

Boost selection is automatic by best available craft.

## Wildlife Encounter Engine

Source: `internal/game/wildlife.go`.

Deterministic roll key dimensions:

- seed, x, y, day, time block
- action type (`move`, `forage`, `hunt`, `fish`)
- roll index + salt

Channels:

- mammals
- birds
- fish
- insects

Selection is weighted by:

- biome
- time block (dawn/day/dusk/night)
- action
- near-water status
- per-cell state (`disturbance`, `hunt pressure`, `depletion`, `carcass token`)

## Persistent Ecological State

Per-cell `CellState` persists meaningful pressure only:

- `HuntPressure`
- `Disturbance`
- `Depletion`
- `CarcassToken`

Decay occurs daily in `decayCellStates`.

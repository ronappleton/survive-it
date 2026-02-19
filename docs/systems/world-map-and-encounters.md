# World Topology, Map, Fog, and Encounters

## Topology Grid

Source: `internal/game/topology.go`.

Each 100m cell is `TopoCell`:

- elevation, moisture, temperature
- biome enum
- flags (`water`, `river`, `lake`, `coast`)
- roughness (movement cost influence)

## Map Size by Mode and Scenario

Sizing function: `topologySizeForScenario`.

- `Alone`: default `36x36` (clamped `28..46` per axis)
- `Naked & Afraid`: default `100x100` (clamped `88..125`)
- `Naked & Afraid XL`: default `125x125` (clamped `100..150`)

Scenario-level fields:

- `Scenario.MapWidthCells`
- `Scenario.MapHeightCells`

## Generation Pipeline

`GenerateWorldTopology` builds deterministic terrain from seed:

1. layered value noise for elevation
2. moisture and temperature maps with biases from biome tags
3. biome assignment by thresholds
4. downhill flow and accumulation for river carving
5. lake/coast/water flags
6. roughness assignment

## Fog of War

- Fog mask is stored in `RunState.FogMask`.
- `Alone`: unrevealed at start, reveal persists permanently.
- `N&A` and `N&A XL`: fully revealed.
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

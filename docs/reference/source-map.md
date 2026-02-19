# Source File Map

This is a file-level index of the current codebase.

## Entry Points

- `cmd/survive-it/main.go`: desktop app entrypoint (Raylib UI).
- `cmd/survive-it/main_headless.go`: headless/runtime entry variant.
- `cmd/docsgen/main.go`: documentation catalog generator (`go run ./cmd/docsgen`).

## `internal/game` (simulation runtime)

### Core runtime and config

- `internal/game/config.go`: game mode + run config validation.
- `internal/game/state.go`: run state structure and initialization.
- `internal/game/scenario.go`: scenario model, season sets, external scenario plumbing.
- `internal/game/scenarios_builtin.go`: built-in scenario definitions.
- `internal/game/season_resolver.go`: season phase resolution by run day.
- `internal/game/advance_day.go`: day advancement, daily effects, run outcome checks.

### Player and progression

- `internal/game/player.go`: player state/config and player creation.
- `internal/game/player_progression.go`: skill progression + trait modifier math.
- `internal/game/physiology.go`: physiology profiles by body type.
- `internal/game/player_decay.go`: dehydration/malnutrition decay and ailment triggers.

### Metabolism and food simulation

- `internal/game/metabolism.go`: reserves, effect bars, daily needs, penalties.
- `internal/game/metabolism_realtime.go`: realtime fractional metabolism update.
- `internal/game/food_simulation.go`: consume-catch nutrition and disease events.
- `internal/game/food_inventory_actions.go`: gut/cook/preserve/eat inventory pipeline.
- `internal/game/run_food.go`: hunt/fish result plumbing into run command outputs.

### World and environment

- `internal/game/environment.go`: biome weather tables, temperature ranges, wildlife lists.
- `internal/game/weather_state.go`: deterministic weather state generation by day.
- `internal/game/weather_effects.go`: weather impact and player adjustment logic.
- `internal/game/gear_weather_effects.go`: clothing + kit weather modifiers.
- `internal/game/topology.go`: topology generation, fog, biome cells, cell-state decay.
- `internal/game/wildlife.go`: deterministic encounter engine and channel/species weighting.
- `internal/game/travel.go`: movement cost/time, map position, encounters during travel.

### Resource, crafting, and inventory systems

- `internal/game/kit.go`: kit item catalog.
- `internal/game/inventory_system.go`: camp/personal inventory and capacity/carry logic.
- `internal/game/environment_resources.go`: plants/resources/trees/shelters/fire/craftables.
- `internal/game/crafting_quality.go`: craft quality scoring.
- `internal/game/trapping.go`: trap specs and trap set/check simulation.

### Command execution

- `internal/game/run_commands.go`: run command routing and command handlers.

### Utilities

- `internal/game/random.go`: seeded RNG helpers.

### Tests

- `internal/game/animals_test.go`: animal catalog, catch, and carcass-flow tests.
- `internal/game/environment_resources_test.go`: resources/crafting/inventory/trap/food tests.
- `internal/game/metabolism_test.go`: metabolism and deficiency behavior tests.
- `internal/game/random_test.go`: deterministic RNG tests.
- `internal/game/run_commands_test.go`: run command behavior tests.
- `internal/game/scenarios_builtin_test.go`: scenario validation tests.
- `internal/game/topology_wildlife_test.go`: topology determinism/fog/encounter balance tests.
- `internal/game/weather_test.go`: weather and biome effect tests.

## `internal/gui` (Raylib application UI)

- `internal/gui/app.go`: main UI loop, screens, run HUD, input pipeline, parser integration.
- `internal/gui/extra_screens.go`: setup builders/editors (stats, players, scenario builder, inventory pages).
- `internal/gui/run_map.go`: run-screen minimap + full-screen topology map rendering.
- `internal/gui/intent_queue.go`: intent queue + command sink boundary.
- `internal/gui/scenario_store.go`: custom scenario load/save and normalization.

## `internal/parser` (intent parser)

- `internal/parser/types.go`: intent/context/command definition types.
- `internal/parser/normalize.go`: input normalization, quantity parsing, direction mapping.
- `internal/parser/registry.go`: command registration and fuzzy command matching.
- `internal/parser/parse.go`: parse orchestration, free-text inference, entity resolution, clarification.
- `internal/parser/README.md`: parser package usage notes.
- `internal/parser/parser_test.go`: parser normalization/fuzzy/inference/pronoun tests.

## `internal/update`

- `internal/update/update.go`: update check/apply logic.
- `internal/update/update_security_test.go`: update security validation tests.

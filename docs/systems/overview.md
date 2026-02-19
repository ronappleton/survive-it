# Systems Overview

Survive It is a simulation-heavy survival game built in Go.

## Core Runtime

- `internal/game/state.go`: run state creation, scenario resolution, topology init.
- `internal/game/config.go`: mode/config validation.
- `internal/game/run_commands.go`: strict command execution and command routing.
- `internal/game/run_food.go`: hunt/fish command execution helpers.
- `internal/game/travel.go`: movement, terrain cost, map position, travel outcomes.
- `internal/game/advance_day.go`: daily tick, weather and camp impacts, progression.

## Player and Survival Model

- `internal/game/player.go`: player model and initial stat setup.
- `internal/game/player_progression.go`: skill progression and trait modifier helpers.
- `internal/game/metabolism.go`: nutrition reserves and effect-bar mechanics.
- `internal/game/metabolism_realtime.go`: partial-day metabolism consumption.
- `internal/game/player_decay.go`: deficiency and dehydration penalties.
- `internal/game/physiology.go`: body-type baseline drains/carry modifiers.

## Environment and World

- `internal/game/environment.go`: biome weather distributions, temp ranges, wildlife/insects lists.
- `internal/game/weather_state.go`: deterministic weather state for each day.
- `internal/game/weather_effects.go`: weather/temperature impact math.
- `internal/game/topology.go`: topology grid generation, fog mask, cell state decay.
- `internal/game/wildlife.go`: deterministic encounter engine (mammal/bird/fish/insect).

## Crafting, Resources, Inventory, Food

- `internal/game/environment_resources.go`: plants, resources, trees, bark stripping, fire, shelters, craftables.
- `internal/game/trapping.go`: trap specs, set/check behavior, catch roll logic.
- `internal/game/inventory_system.go`: camp/personal inventory, carry limits, capacity logic.
- `internal/game/food_inventory_actions.go`: gut/cook/preserve/eat and spoilage interfaces.
- `internal/game/food_simulation.go`: disease and nutrition outcomes from catch consumption model.
- `internal/game/kit.go`: all personal/issued kit items.

## Scenario and Mode Layer

- `internal/game/scenario.go`: scenario model.
- `internal/game/scenarios_builtin.go`: built-in scenarios and defaults.
- `internal/game/season_resolver.go`: season phase resolution by run day.

## UI + Input

- `internal/gui/app.go`: screen management, run UI, input submission, parser integration.
- `internal/gui/run_map.go`: run-screen minimap and full-screen map rendering.
- `internal/gui/extra_screens.go`: scenario builder, kit picker, and additional screens.
- `internal/gui/intent_queue.go`: command sink/intent queue boundary.
- `internal/parser/*`: deterministic command + free-text parser.

## Data References

The full item/spec references are generated from source and live in:

- `docs/reference/catalogs/`

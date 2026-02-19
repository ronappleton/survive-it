# Inventory, Crafting, Trapping, and Food Processing

## Inventory Model

Source: `internal/game/inventory_system.go`.

Two inventories:

- Camp/shelter inventory (`RunState.CampInventory` + wood/resource stock)
- Per-player personal inventory (`PlayerState.PersonalItems`)

Capacity rules:

- Camp capacity from shelter type and crafted storage items.
- Personal carry limit from player stats/physiology.
- Items have quantity, unit, per-unit weight, quality, and age days.

Core inventory commands:

- `inventory camp`
- `inventory personal [p#]`
- `inventory stash <id> [qty] [p#]`
- `inventory take <id> [qty] [p#]`
- `inventory add <id> [qty] [p#]`
- `inventory drop <id> [qty] [p#]`

## Gathering and Craft Inputs

Primary catalogs are in `internal/game/environment_resources.go`:

- `PlantCatalog`
- `ResourceCatalog`
- `TreeCatalog`

Key gathering commands:

- `forage ...`
- `collect ...`
- `wood gather|dry|stock`
- `bark strip ...`

## Crafting

Craft model: `CraftableCatalog` + `CraftItem`.

Craft outcomes include:

- quality tier (`excellent/good/poor/...` via crafting quality scoring)
- action hours consumed
- storage destination (`personal` or `camp`)

Requirements can include:

- min bushcraft
- active fire
- active shelter
- prerequisite crafted items
- required resource quantities

## Trapping

Source: `internal/game/trapping.go`.

System includes:

- trap specs by biome
- setup quality/effectiveness
- water requirements
- crafted/resource/kit prerequisites
- pending catches and rearm/break cycles

Trap commands:

- `trap list`
- `trap set <id> [p#]`
- `trap status`
- `trap check`

## Food Pipeline (Realistic Carcass-First)

Source: `internal/game/food_inventory_actions.go`.

Flow:

1. acquire carcass (`hunt`/`fish`/traps)
2. `gut <carcass> [kg] [p#]`
   - includes intestine puncture risk
   - can produce inedible/spoiled output
3. `cook <raw_meat> [kg] [p#]`
4. `preserve <smoke|dry|salt> <meat> [kg] [p#]`
5. `eat <food_item> [grams|kg] [p#]`

### Preservation and Degradation

- food items have shelf-life and decay rates
- smoking/drying/salting produce different shelf lives and illness risk profiles
- day advancement degrades stored food

## Clothing and Weather Interaction

Crafted clothing and kit can directly modify weather impacts:

- logic in `internal/game/gear_weather_effects.go`
- examples: `hide_jacket`, `grass_cape`, `woven_tunic`, `hide_moccasins`, plus kit thermal/rain gear

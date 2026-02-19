# Commands Reference

Source of truth: `internal/game/run_commands.go` (`ExecuteRunCommand`).

## Core

- `help`
- `commands`
- `next`
- `save`
- `load`
- `menu`

## Hunting and Gathering

- `hunt <land|fish|air> [p#]`
- `fish [p#]`
- `forage [roots|berries|fruits|vegetables|any] [p#] [grams]`

## Resources and Materials

- `trees`
- `plants`
- `resources`
- `collect <resource|any> [qty] [p#]`
- `bark strip [tree|any] [qty] [p#]`
- `wood gather [kg] [p#]`
- `wood dry [kg] [p#]`
- `wood stock`

## Inventory

- `inventory`
- `inventory camp`
- `inventory personal [p#]`
- `inventory stash <item_id> [qty] [p#]`
- `inventory take <item_id> [qty] [p#]`
- `inventory add <item_id> [qty] [p#]`
- `inventory drop <item_id> [qty] [p#]`

## Trapping

- `trap list`
- `trap set <id> [p#]`
- `trap status`
- `trap check`

## Carcass and Food Pipeline

- `gut <small_game_carcass|bird_carcass|fish_carcass> [kg] [p#]`
- `cook <raw_small_game_meat|raw_bird_meat|raw_fish_meat> [kg] [p#]`
- `preserve <smoke|dry|salt> <raw_or_cooked_meat_id> [kg] [p#]`
- `smoke <meat_id> [kg] [p#]`
- `dry <meat_id> [kg] [p#]`
- `salt <meat_id> [kg] [p#]`
- `eat <food_item> [grams|kg] [p#]`

## Movement and Navigation

- `go <north|south|east|west|n|s|e|w> [km] [p#]`

## Fire, Shelter, Crafting

- `fire status`
- `fire methods`
- `fire prep tinder|kindling|feather [count] [p#]`
- `fire ember bow|hand [woodtype] [p#]`
- `fire ignite [woodtype] [kg] [p#]`
- `fire build [woodtype] [kg] [p#]`
- `fire tend [woodtype] [kg] [p#]`
- `fire out`
- `shelter list`
- `shelter build <id> [p#]`
- `shelter status`
- `craft list`
- `craft make <id> [p#]`
- `craft inventory`

## Equipment Actions

- `actions [p#]`
- `use <item> <action> [p#]`

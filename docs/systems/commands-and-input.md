# Commands and Input Parsing

## Input Flow

Main path: `internal/gui/app.go`.

On Enter in run input bar:

1. Build `ParseContext` from live state (`buildParseContext`).
2. Parse with `cmdParser.Parse`.
3. If clarify is returned:
   - store `pendingClarify` and `pendingOptions`
   - render clarify prompt above input
   - allow numeric selection (`1`-`9`) or typed option
4. Otherwise enqueue `Intent` to command sink (`intentQueue`).
5. UI drains queue in `processIntentQueue` and executes via `executeIntent`.

## Parser Design

Parser package: `internal/parser`.

- Normalization/tokenization: `normalize.go`
- Registry + aliases + fuzzy command matching: `registry.go`
- Intent inference/entity resolution/clarification: `parse.go`
- Types: `types.go`

### Parsing Priorities

1. Exact command
2. Alias
3. Prefix
4. Levenshtein fuzzy (threshold by token length)

### Free-Text Rules

Deterministic rule table maps phrases like:

- bag/inventory phrasing -> `inventory`
- need fire phrasing -> `fire build`
- look/location phrasing -> `look`
- movement phrasing -> `go <direction>`
- preserve/smoke phrasing -> `preserve`
- eat/drink/sleep phrasing -> direct verbs

### Pronouns

- `it/that/them` resolve from `ParseContext.LastEntity`.
- if unresolved, parser returns clarification request.

## Strict Run Commands

Strict command executor: `internal/game/run_commands.go`.

Major command families:

- `help` / `commands`
- hunting + foraging: `hunt`, `fish`, `forage`
- map/movement: `go`
- resources: `trees`, `plants`, `resources`, `collect`, `bark strip`
- inventory: `inventory camp|personal|stash|take|add|drop`
- traps: `trap list|set|status|check`
- food pipeline: `gut`, `cook`, `preserve`, `eat`
- camp systems: `wood`, `fire`, `shelter`, `craft`
- equipment actions: `actions`, `use`
- run controls: `next`, `save`, `load`, `menu`

## Run Screen Shortcuts

- `M`: full map toggle
- `Shift+P`: players view
- `Shift+H`: command library
- `Shift+I`: inventory view
- `S`: save
- `L`: load
- `Esc`: back/menu

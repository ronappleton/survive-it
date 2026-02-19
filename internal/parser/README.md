# Parser

Deterministic command+intent parser for the run input bar.

## Registering Commands

Use `RegisterCommand` with `CommandDef`:

```go
p := parser.New()
p.RegisterCommand(parser.CommandDef{
    Canonical: "inspect",
    Aliases:   []string{"examine", "check"},
    MinArgs:   1,
    MaxArgs:   4,
    HandlerKey:"inspect",
})
```

`CommandDef` fields:
- `Canonical`: canonical verb
- `Aliases`: accepted aliases/phrases
- `MinArgs`, `MaxArgs`: arg guardrails
- `HandlerKey`: engine-side routing key

## Building ParseContext

Build from current state each submit:
- `Inventory`: current player item names/ids
- `Nearby`: nearby/reachable entities
- `KnownDirections`: `north/south/east/west` (+ short forms if desired)
- `LastEntity`: last referenced/acted-on entity for pronoun resolution

Example:

```go
ctx := parser.ParseContext{
    Inventory: []string{"ferro rod", "knife"},
    Nearby: []string{"stick", "stone"},
    KnownDirections: []string{"north", "south", "east", "west", "n", "s", "e", "w"},
    LastEntity: "ferro rod",
}
intent := p.Parse(ctx, input)
```

## Thresholds + Scoring

Ranking priority:
1. exact canonical
2. exact alias
3. prefix
4. levenshtein match (distance limit by token length)

Entity resolution uses exact/prefix/fuzzy and applies in-scope boosts:
- nearby boost for `take`-style intents
- inventory boost for `drop/use/eat/drink`

If top candidates are too close, parser returns `ClarifyQuestion`.

## Free-Text Inference Rules

Deterministic rule-table covers:
- bag/inventory phrases -> `inventory`
- fire-need phrases -> `fire build`
- preservation phrases (`smoke/preserve/dry meat`) -> `preserve`
- location phrases -> `look`
- movement phrases -> `go <direction>`
- eat/drink/sleep phrases -> matching verbs

## Adding New Inference Rules

Edit `inferFreeTextIntent` in `parse.go`:
- add explicit phrase checks
- return a concrete `Intent` (no side effects)
- keep rules deterministic and test them in `parser_test.go`

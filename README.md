# Survive It

Survival simulation game implemented in Go with a Raylib-based desktop client.

## Build and Run

Requires a cgo-enabled build (Raylib client).

```bash
go run ./cmd/survive-it
```

Print version:

```bash
go run ./cmd/survive-it --version
```

Disable update checks:

```bash
go run ./cmd/survive-it --no-update
```

## Documentation

Full docs live in [`docs/`](./docs/README.md).

Quick links:

- [Systems Overview](./docs/systems/overview.md)
- [Commands and Input Parsing](./docs/systems/commands-and-input.md)
- [World Topology, Map, Fog, and Encounters](./docs/systems/world-map-and-encounters.md)
- [Inventory, Crafting, Trapping, and Food Processing](./docs/systems/inventory-crafting-trapping-and-food.md)
- [Weather, Physiology, and Effects](./docs/systems/weather-physiology-and-effects.md)
- [Commands Reference](./docs/reference/commands-reference.md)
- [Source File Map](./docs/reference/source-map.md)
- [Generated Data Catalogs](./docs/reference/catalogs/README.md)

## Catalog Regeneration

Data reference markdown files are generated from source catalogs:

```bash
go run ./cmd/docsgen
```

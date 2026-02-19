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
- [World Map/Encounter System Reference Alias](./docs/reference/systems/world-map-and-encounters.md)

## Catalog Regeneration

Data reference markdown files are generated from source catalogs:

```bash
go run ./cmd/docsgen
```

## Real-World Terrain Profiles

Runtime map generation is offline and uses local distilled profile JSON files:

- profile files: `assets/profiles/*.json`
- runtime loader: `internal/game/gen_profile.go`
- generation tools: `cmd/genprofile`, `cmd/genprofiles`

Generate/refresh profiles (developer-time only, uses network once to sample elevation data):

```bash
go run ./cmd/genprofiles
```

Generate one profile manually:

```bash
go run ./cmd/genprofile --bbox "minLon,minLat,maxLon,maxLat" --out "assets/profiles/example.json" --cell 100
```

Raw downloaded cache is stored in `.cache/genprofile/` and is git-ignored.

## macOS `.app` Packaging

After GoReleaser builds Darwin binaries into `dist/`, package `.app` bundles (including `assets/`) for both `amd64` and `arm64`:

```bash
./scripts/package-macos-app.sh v1.2.3
```

Outputs:

- `dist/survive-it_<version>_darwin_amd64.app.zip`
- `dist/survive-it_<version>_darwin_arm64.app.zip`
- `dist/macos-app-artifacts.txt` (artifact list for CI upload)

### Signing and Notarization (recommended)

Unsigned/ad-hoc-signed apps downloaded from the internet may be blocked by Gatekeeper and shown as "damaged".

The packaging script supports Developer ID signing + notarization via env vars:

- `MACOS_SIGN_IDENTITY` (Developer ID Application identity name)
- `MACOS_NOTARY_APPLE_ID`
- `MACOS_NOTARY_TEAM_ID`
- `MACOS_NOTARY_APP_PASSWORD`

CI workflows also support optional certificate import with:

- `MACOS_CERT_P12_BASE64`
- `MACOS_CERT_PASSWORD`

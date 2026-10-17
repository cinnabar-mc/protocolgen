# Active Bedrock game catalog

This module provides shared game facts for Bedrock servers and clients. Import
it as `github.com/cinnabar-mc/protocolgen/generated/data`. It uses only the Go
standard library and does not import the generators or source loader.

There is one active catalog. Older releases remain available through Git and
Go module pins, rather than separate catalog folders for each version.
`MinecraftVersion`, `ProtocolVersion` and
`SourceLockSHA256` identify its target and complete input lock.
[release.json](release.json) records the same target and every pinned input
revision and digest. Consumers using the separate `data/source` module can
compare the catalog hash against SHA-256 of `json.Marshal(sources.Lock)` to
reject mismatched module revisions even when the game version is unchanged.

```go
import "github.com/cinnabar-mc/protocolgen/generated/data/block"

state, ok := block.StateByHash(hash)
if ok {
    properties, _ := block.PropertiesAt(state.Properties)
    boxes, available := block.Shape(properties.CollisionShape)
    // available distinguishes a missing shape from a known empty shape.
    _, _ = boxes, available
}
```

| Package | Contents |
| --- | --- |
| `data` | Target release, independent source versions and generated counts |
| `block` | State hashes, physical properties, light, tint, shapes and provisional mining-tool facts |
| `biome` | Climate, terrain values, water color, rain and tags |
| `entity` | Base collision, health, movement, scale, projectile and family values |
| `item` | Runtime registry, stack limits, equipment flags and food values |
| `voxelshape` | Named shapes in block-local coordinates |

Named values include `block.Stone`, `biome.Plains`, `entity.Arrow`, `item.Apple`
and `voxelshape.Anvil`. Each package also exposes lookups and iteration.
Treat returned slices and definitions as immutable.

## Independent source versions

The runtime target is Minecraft 1.26.50, protocol 2193, with 2,076 runtime items.
The semantic snapshot uses CloudburstMC/Data commit
`659ce1e2eee3a67045693f4fd5515c6ccf571953` for blocks, biomes and named voxel
shapes, and BDS `1.26.32.2` for base entity and food inputs. The same Cloudburst
revision supplies runtime item identities; Allay remains an item-property and
supplemental-tag input only.
These inputs contain 22,091 block states, 89 biomes, 96 entities, 38 foods and
220 named voxel shapes. Biome metadata may lack a numeric ID; check `HasID`
before using `ID`. [semantic_sources.json](semantic_sources.json) and
`version_generated.go` record these independent versions and counts.
`BDSVersion` describes the entity and food inputs, not the runtime
protocol target.

`block.MiningByName` reports effective tools, drop-qualifying tools and a
minimum tier from pinned Prismarine Bedrock 1.26.30 data. These classifications
are provisional for the target release. A false lookup means the source lacks
usable classification; it does not establish that a block is hand-harvestable.
The evidence version, commit and SHA-256 are exposed beside the table.

Entity values describe base components. They do not execute goals, events,
filters or component-group transitions. Food definitions alone are not a
runtime item registry. Neighbor-dependent block shapes describe the extraction
context; live worlds may need their own models to resolve connections.
Unavailable shapes remain distinct from known empty shapes. Thirty reviewed
liquid-clip boxes in the Cloudburst extract have invalid bounds and are marked
unavailable. Their exact source values appear in `semantic_sources.json`, and
`GeneratedCounts.UnavailableLiquidClipShapes` reports their count. Valid geometry
is retained; no guessed or clamped replacement is supplied. The full registry
includes Education and internal entries; membership does not establish retail
availability.

## Generation and API ownership

[Authored generators and source locks](../../data/README.md) live in `data/`.
Run all generators there when changing the release. `*_generated.go` files
are owned by the generator named in their header. The small type, lookup and
test files beside those values are authored API code and may be edited.
`release.json` belongs to runtime generation; `semantic_sources.json` belongs
to semantic generation. Each preserves outputs owned by the other generators. Any source-lock
change requires all generators: catalog tests compare their complete lock
hashes and reject a partial update.

These values are semantic projections, not complete game packs. Raw BDS packs,
textures, structures, localization files and executables are not included.
See [NOTICE.md](NOTICE.md) for attribution.

The `registry` package contains normalized NBT payloads for block states, data-driven
blocks and complete item registry entries. Accessors return fresh byte slices.
Consumers decode them with their own NBT library; this module remains dependency-free.
`ProtocolSnapshot` and `ReleaseChannel` connect the active catalog to its codec
snapshot without confusing preview and retail labels.

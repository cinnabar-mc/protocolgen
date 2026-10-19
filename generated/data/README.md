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
| `block` | State hashes, physical properties, light, tint, shapes and mining-tool facts with explicit evidence status |
| `biome` | Climate, terrain values, water color, rain and tags |
| `entity` | Base collision, health, movement, scale, projectile and family values |
| `item` | Runtime registry, stack limits, native tags, equipment flags and food values |
| `voxelshape` | Named shapes in block-local coordinates |

Named values include `block.Stone`, `biome.Plains`, `entity.Arrow`, `item.Apple`
and `voxelshape.Anvil`. Each package also exposes lookups and iteration.
Treat returned slices and definitions as immutable.

## Independent source versions

The runtime target is Minecraft 1.26.50, protocol 2193, with 2,076 runtime items.
The native snapshot is release 1.26.51, protocol 2193. `NativeRef` pins its commit.
It supplies the block palette, scalar block properties, ordinary-tool and harvest
classifications, item fields and tags, and biome scalar values and IDs. Selected
CC0-1.0 input files are vendored so regeneration does not require upstream access.
The catalog retains measured bucket stack corrections because raw field values
can differ from effective runtime limits. Raw durability is not exposed.

CloudburstMC/Data supplies geometry, rendering fields, biome tags, item identities
and component payloads. Allay supplies only the two missing offhand/head capability
tags. BDS `1.26.32.2` supplies base entity and food inputs; `BDSVersion` describes
those inputs, independently of the target protocol. The catalog contains 22,091
block states, 89 biomes, 96 entities, 38 foods and 220 named voxel shapes.
[semantic_sources.json](semantic_sources.json) and `release.json` record the pins.

`block.MiningByName` provides native ordinary-tool effectiveness and harvest
requirements, including explicit no-tool results. Its `Evidence` is `MiningNative`.
Sword/shears speed targets still use the older Prismarine Bedrock 1.26.30 input;
`SpecialEvidence` is `MiningProvisional`. These tags do not define the complete
mining algorithm. A false lookup means an unmeasured identifier and must not be
interpreted as permission to harvest by hand. Native and special-tool source
identities are exposed beside the table.

`item.RuntimeTags(name)` returns native recipe and equipment tag memberships.
Unknown or untagged identifiers return nil; returned slices are immutable.

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

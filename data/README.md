# Shared game-data generation

This authored module owns the source lock, input verification and generators
for the active [catalog](../generated/data/README.md). Its module path is
`github.com/bedrock-mc/protocolgen/data`. Runtime consumers import
`github.com/cinnabar-mc/protocolgen/generated/data` instead. The runtime module uses only the standard library. The tooling module uses a
pinned NBT codec to normalize registry payloads; neither module imports the other.

## Layout and ownership

- `source/` holds the release lock, verified input loader and selected native measurements and reviewed
  registry/correction inputs. External packs remain outside the repository.
- `internal/generator` and `cmd/generate` produce semantic values.
- `internal/registrygen` and `cmd/registrygen` normalize block palettes, data-driven
  block components and complete item component payloads.
- `internal/runtimegen` and `cmd/runtimegen` produce runtime item values and
  release metadata.
- `internal/mininggen` and `cmd/mininggen` produce native block tool facts with provisional special-tool speed evidence.
- `../generated/data/` is the single active output catalog. It has its own
  `go.mod`; it is not a versioned packet-output directory.

The catalog's small API and lookup files are authored beside their data.
Files ending in `*_generated.go` are owned by their named generator and must
not be edited by hand. All generators default to `../generated/data` when
run from this directory. `-out` selects another catalog root.

## Regeneration

Regenerate runtime items and release metadata from this directory:

```sh
go run ./cmd/runtimegen
go run ./cmd/registrygen
go run ./cmd/mininggen
```

The runtime catalog targets Minecraft 1.26.50, protocol 2193, with 2,076 items.
Sources include the locked Cloudburst registry, native item properties and tags,
two supplemental Allay inventory-capability tags, and measured corrections. Generation rejects missing values and stale corrections.
It writes `item/runtime_generated.go`, `release_generated.go` and `release.json`
in one rollback-protected batch. The source-lock identity is SHA-256 over
the canonical input lock and its resolved release record. JSON indentation does
not affect it.

Regenerate semantic values using local input trees:

```sh
go run ./cmd/generate \
  -cloudburst /path/to/CloudburstMC-Data \
  -bds /path/to/bedrock-server-1.26.32.2
```

Versions and digests come from [source/lock.json](source/lock.json). The
semantic generator verifies all three Cloudburst semantic files and every
consumed BDS JSON file before writing. Added, missing or changed input files fail validation.
The independent BDS 1.26.32.2 semantic provenance is preserved; it is not a
claim that all inputs were captured from the target runtime release.
The vendored CC0-1.0 [didntpot snapshot](source/inputs/1.26.51/native/README.md)
supplies the block palette, scalar block properties, tool requirements and tags,
item stack fields and tags, and biome scalar values and numeric IDs. It covers
22,091 states, 1,477 block identifiers, 2,076 items and 89 biomes. State ordinals,
hashes and identifier domains must match the independent registries; absent
required fields, mismatched IDs and inconsistent tag memberships fail generation.
The snapshot is release 1.26.51, protocol 2193; the runtime target retains its
1.26.50 protocol label. The unknown-block hash keeps the catalog's reserved value.

Cloudburst `blocks.json` remains the source for collision, outline and rendering
fields. Cloudburst also supplies biome tag memberships, named voxel shapes,
item identities and component payloads, all from the revision in the lock.
Native biome IDs replace the previous Cinnabar registry snapshot. Native scalars
replace the corresponding Cloudburst fields, with exact-domain validation.
Named shapes are decoded from occupied grid cells into block-local boxes.

Thirty liquid-clip boxes in this extract have reversed bounds. The reviewed
`liquid_clip_omissions` input identifies each exact name, hash and source box.
Generation marks those fields unavailable and records them in
`semantic_sources.json`; it never clamps them or declares them empty. New,
changed and obsolete exceptions fail generation. Required collision and outline
shapes remain strictly validated. Zero-volume outlines remain known-empty.

All commands accept `-cache /path/to/cache` and `-lock /path/to/lock.json`.
Use a reviewed alternate lock to change a release; there are no independent
version flags. Any source-lock change requires all generators. Catalog tests
require the semantic, runtime and registry lock digests to match, so a partial
update fails even when the game version or source revision stays the same.
Semantic generation preserves runtime release metadata. Each generator removes
stale Go files only when they carry its exact ownership header, preserving the other generator's
outputs and authored files.

From the repository root:

```sh
make test-data
make verify-data CLOUDBURST_DIR=/path/to/CloudburstMC-Data BDS_DIR=/path/to/bds
```

`test-data` checks both modules independently. Generation tests use synthetic
fixtures. A repository-level AST check reads the sibling catalog's `Runtime`
struct and verifies every field is emitted without adding a module dependency
or maintaining a second schema list. Run the generator tests from a full
repository checkout. CI downloads and authenticates the locked Cloudburst
block/biome/shape files and compares the full generated projection with the
committed catalog. It also regenerates runtime items, mining facts, registry
payloads and metadata. To run the Cloudburst check locally, set
`PROTOCOLGEN_CLOUDBURST_DIR` when running `make test-data`. Full semantic
regeneration additionally requires the local BDS inputs.

## Shared input access

`source.Open("")` loads the embedded lock. `Sources.Read` authenticates each
input with SHA-256. `Sources.ResourcePack` downloads and extracts the locked
sample resource pack into the user's cache. Texture, glyph and block-metadata
capture tools share those pins while retaining their own processing logic.
Raw packs, textures and executables are not published here.

See [NOTICE.md](NOTICE.md) for source attribution.

## Remaining sources

The native item dump stores raw fields, not every effective runtime getter.
Twelve bucket stack limits therefore use measured corrections: generation
requires the raw value to remain unchanged, rejecting stale overrides. Effective
durability is not exposed from this dump: component-based items can have zero or
conflicting raw values. Consumers must retain their measured durability inputs.

Only `minecraft:allow_offhand` and `minecraft:head` remain from Allay's supplemental
tag table. The native dump does not provide these inventory-capability getters.
Body-slot flags and recipe tags use native memberships. Missing data never means
a default inventory capability.

Ordinary pickaxe, axe, shovel and hoe effectiveness, plus harvest tool and tier
requirements, use native block tags and `requiresCorrectToolForDrops`. Known
no-tool blocks are emitted explicitly, including beds. Prismarine's older input
is used only for sword/shears speed targets absent from the dump. Those speed
rules remain provisional; complete native mining parity needs additional evidence.
Cloudburst geometry, biome tags and item components, and official pack-derived
entity/food data also remain independent inputs because the selected dump does
not replace them completely.

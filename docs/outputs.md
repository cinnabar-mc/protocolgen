# Generated outputs

Each release in [data/source/releases.json](../data/source/releases.json) has a
snapshot directory `generated/<version>/`; `default_protocol` names the current
one.

## Protocol snapshots

| Path | Contents | Consumed by |
| --- | --- | --- |
| `generated/<version>/manifest.json` | Canonical packet and type layouts, with source pins, corrections and adjudications | Every emitter and `verify-gophertunnel`. Its SHA-256 is `MANIFEST_SHA256` in the contract crate and is pinned by Cinnabar's `assets/bedrock-target.json` |
| `generated/<version>/go/protocol` | Go codec: shared types grouped by domain, one file per packet under `packet/`, direction-aware packet pools | `differential/` harnesses (1.26.51); `internal/roundtrip` tests (1.26.40) |
| `generated/<version>/rust` | Rust crate `bedrock-protocol-<version>` with round-trip tests | CI only (check, clippy, test) |
| `generated/<default>/rust-contract` | The same crate under the [Rust codec contract](rust-codec-contract.md) | Vendored by Cinnabar as `crates/protocol/vendor/bedrock-protocol-1-26-51`; `differential/retail` runner and bench |
| `generated/1.26.50/changelog.md`, `gophertunnel.md` | Packet changes from 2168 to 2187, and the gophertunnel edits they need | Whoever updates gophertunnel |
| `generated/<default>/gophertunnel-layout.json`, `semantics.json` | Gophertunnel file layout and field names; actor-ID field markings | `emit-go -layout`; `emit-go` reads `semantics.json` by default |
| `docs/gophertunnel-gap-1.26.51.md` | Remaining differences from gophertunnel | Reference |

The other files in a snapshot directory are reviewed inputs: `source-lock.json`,
`corrections/`, `adjudications.json`, `adjudication-selections.json`,
`directions.json`, `nbt-encodings.json`, `naming.json`, `domains.json`,
`docs.json`, `hotfix.json`, `lens-claims.json` and `evidence/`.

## Game-data catalog

`generated/data` is the standard-library-only Go module
`github.com/cinnabar-mc/protocolgen/generated/data`; its
[README](../generated/data/README.md) documents the API. Inputs are pinned in
[data/source/lock.json](../data/source/lock.json). "Native" is the vendored
didntpot/bedrock-data 1.26.51 snapshot under `data/source/inputs/1.26.51/native`.
Generators live in `data/cmd`.

| Package or file | Contents | Sources | Generator | Cinnabar consumer |
| --- | --- | --- | --- | --- |
| `block` | Block states, light, tint, collision and rendering shapes | Native palette, properties and tags; Cloudburst `blocks.json`; reviewed liquid-clip omissions | `generate` | `tools/registrygen` |
| `block/mining_generated.go` | Tool effectiveness and harvest requirements | Native block tags and properties; Prismarine 1.26.30 for sword and shears speed | `mininggen` | `registrygen/cmd/blockdestroy` |
| `biome` | Climate, terrain, water colour, rain, tags and IDs | Native biome definitions and IDs; Cloudburst biome tags | `generate` | `tools/registrygen` |
| `entity` | Base collision, health, movement, scale and family values | BDS 1.26.32.2 `behavior_packs/vanilla/entities` | `generate` | None |
| `item` (foods) | Food definitions | BDS 1.26.32.2 `behavior_packs/vanilla/items` | `generate` | None |
| `item/runtime_generated.go` | Runtime item registry, stack limits, tags and equipment flags | Cloudburst item identities; native item properties and tags; Allay offhand/head tags; reviewed corrections | `runtimegen` | `registrygen/cmd/itemtags` |
| `voxelshape` | Named shapes in block-local coordinates | Cloudburst `voxel_shapes.json` | `generate` | None |
| `registry` | Normalized NBT for block states, data-driven blocks and item registry entries | Native canonical block states; Cloudburst data-driven blocks, item components and identities | `registrygen` | `tools/registrygen` |
| `release_generated.go`, `release.json` | Target release and the full input lock | `data/source` lock and release index | `runtimegen` | `tools/registrygen` |
| `version_generated.go`, `semantic_sources.json` | Semantic source pins and generated counts | Same lock | `generate` | `tools/registrygen` |

## Other directories

| Path | What it is | Still needed |
| --- | --- | --- |
| `candidates/` | Pinned sources and corrections for the 1.26.60 preview; ingested, not emitted | Yes, until 1.26.60 becomes a snapshot. CI ingests it |
| `differential/` | Codec comparison against gophertunnel, and the retail corpus run across Valentine, the contract crate, generated Go and gophertunnel | Yes. CI runs `make differential` and `make retail` |
| `tools/` | Gophertunnel oracle lock and accepted divergences, overlay seeders for new snapshots, the release-index check | Yes. `verify-gophertunnel`, `gophertunnel-layout` and CI use them |
| `migration/` | Axolotl protocol code lifted with its history | No build, test or CI uses it; kept as provenance |

## Not committed

| Path | Contents | Produced by |
| --- | --- | --- |
| `build/gophertunnel-layout` | Current release's Go in gophertunnel's layout, for `diff -r` | `make gophertunnel-layout GOPHERTUNNEL_DIR=…` |
| `build/retail` | Per-packet verdicts comparing Valentine, the contract crate, generated Go and gophertunnel | `make retail`; CI artifact `retail-verdicts` |
| `build/bench` | Codec benchmark comparison against gophertunnel | `make bench` |
| `ORACLE_REPORT` (default `/tmp/protocolgen-gophertunnel-report.json`) | Manifest-to-gophertunnel comparison | `make regen-1.26.51`; CI artifact `gophertunnel-oracle-2193` |

## Regenerating and checking

Every `verify` target regenerates and fails if the committed output changes.
CI runs all of them except `verify-data`.

| Output | Regenerate | Check |
| --- | --- | --- |
| 1.26.40 and the 1.26.44 hotfix | `make regen hotfix` | `make verify` (also runs `differential`) |
| 1.26.50 | `make regen-1.26.50` | `make verify-1.26.50` |
| 1.26.51, including `rust-contract` | `make regen-1.26.51` | `make verify-1.26.51` |
| All Rust crates, without doc checkouts | `make regen-rust` | `make verify-rust` |
| `rust-contract` only | `make rust-contract` | Same target; also runs clippy and tests |
| `generated/data` | `make generate-data CLOUDBURST_DIR=… BDS_DIR=…` | `make verify-data` |

The `regen` targets need `MOJANG_DIR` and `ENDSTONE_DIR` checked out at the
snapshot's `source-lock.json` revisions. Only `generate` needs local data
inputs: `CLOUDBURST_DIR` holds the three locked Cloudburst files and `BDS_DIR` a
BDS 1.26.32.2 tree for entities and foods. The other generators fetch or read
their locked inputs themselves, so CI checks `generated/data` by rerunning only
those.

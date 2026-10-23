# Generated outputs

Each release in [data/source/releases.json](../data/source/releases.json) has a
snapshot directory `generated/<version>/`; `default_protocol` names the current
one.

## Committed

| Path | Contents | Consumed by |
| --- | --- | --- |
| `generated/<version>/manifest.json` | Canonical packet and type layouts, with source pins, corrections and adjudications | Every emitter, `verify-gophertunnel`, the BDS capture. Its SHA-256 is `MANIFEST_SHA256` in the contract crate and is pinned by Cinnabar's `assets/bedrock-target.json` |
| `generated/<version>/go/protocol` | Go codec: shared types grouped by domain, one file per packet under `packet/`, direction-aware packet pools | `differential/` harnesses (1.26.51); `internal/roundtrip` tests (1.26.40) |
| `generated/<version>/rust` | Rust crate `bedrock-protocol-<version>` with round-trip tests | CI only (check, clippy, test) |
| `generated/<default>/rust-contract` | The same crate under the [Rust codec contract](rust-codec-contract.md) | Vendored by Cinnabar as `crates/protocol/vendor/bedrock-protocol-1-26-51`; `differential/retail` runner and bench |
| `generated/<version>/vanilla-data` | Raw packet payloads and NBT registries captured from a pinned BDS, indexed by `capture.json` (1.26.40 and 1.26.44 only) | No current consumer here or in Cinnabar |
| `generated/1.26.50/changelog.md`, `gophertunnel.md` | Packet changes from 2168 to 2187, and the gophertunnel edits they need | Whoever updates gophertunnel |
| `generated/<default>/gophertunnel-layout.json`, `semantics.json` | Gophertunnel file layout and field names; actor-ID field markings | `emit-go -layout`; `emit-go` reads `semantics.json` by default |
| `docs/gophertunnel-gap-1.26.51.md` | Remaining differences from gophertunnel | Reference |
| `generated/data` | Go module `github.com/cinnabar-mc/protocolgen/generated/data`: block, biome, entity, item and voxel-shape facts plus NBT registry payloads; standard library only. See its [README](../generated/data/README.md) | Cinnabar's `tools/registrygen` (block and biome registries, item tags, block-destroy table) |

The other files in a snapshot directory are reviewed inputs: `source-lock.json`,
`corrections/`, `adjudications.json`, `adjudication-selections.json`,
`directions.json`, `nbt-encodings.json`, `naming.json`, `domains.json`,
`docs.json`, `hotfix.json`, `vanilla-source.json`, `lens-claims.json` and
`evidence/`.

Each Go and Rust codec output includes `LICENSE-MIT` and `NOTICE`, emitted from
`internal/licensing/`. The MIT grant covers Cinnabar's authored implementation;
the notice retains the separate rights and attribution for imported comments
and inputs. It does not relicense the source documents or game-data catalog.

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
| `vanilla-data` | `make vanilla-data BDS_BINARY=…`, or the `vanilla-data` workflow | None; CI validates only the capture setup |

The `regen` targets need `MOJANG_DIR` and `ENDSTONE_DIR` checked out at the
snapshot's `source-lock.json` revisions. CI checks `generated/data` by rerunning
only the runtime, mining and registry generators, which need no external inputs.

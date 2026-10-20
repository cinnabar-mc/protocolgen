# Differential codec harness

This module compares protocol 2193 payloads between the generated 1.26.51 Go codec and the gophertunnel fork commit that `gophertunnel.json` pins. That lock is the one gophertunnel pin for every module here: `scripts/pin-gophertunnel.sh [VERSION]` rewrites their `go.mod` files from it, and `TestGophertunnelPinsMatchTheLock` fails when one disagrees.

Pool parity accepts five deprecated oracle-only IDs that are still registered by gophertunnel but absent from the manifest: 55 `AdventureSettings`, 117 `ScriptCustomEvent`, 163 `FilterText`, 173 `PhotoInfoRequest`, and 197 `ClientCheatAbility`.
Run `go test ./...` for the committed corpus. Run `go test -fuzz=FuzzDifferential -fuzztime=10m -promote` to fuzz and promote successful cases.

## Comparison rules

The contract is one-directional: every input the pinned oracle accepts must
decode here and re-encode byte-identically. Three structural allowances,
each verified against oracle source:

- **Bool canonicalization**: fields the manifest types as bool but the oracle
  types as passthrough uint8 re-encode as canonical 0/1 here. Only reachable
  with non-canonical bool bytes, which vanilla peers never send.
- **Oracle truncation leniency**: the oracle zero-pads truncated fixed-width
  reads (unchecked Read counts), has no varint byte cap, and skips NBT
  validation. The generated runtime rejects these malformed inputs.
- **Accepted divergences** (`accepted-divergences.json`): per-packet classes
  where the oracle rewrites decoded state on encode — derived enum names and
  ordinals, presence flags derived from values, zeroed reserved or
  inactive-union fields, empty-NBT compound insertion, lossy sound-position
  conversion. The generated codec preserves wire bytes in every adjudicated
  case.

## Retail 2193

`make retail` (CI job `retail`) runs the 53 batches pinned in
`retail/corpus.json` through four decoders and fails on any disagreement that
`retail/exceptions.json` does not accept, or on an exception that accepts
nothing. `scripts/fetch-cinnabar.sh` checks out the pinned Cinnabar commit,
sparsely, into `build/cinnabar`, and Valentine from the last Cinnabar commit
that vendored it into `build/valentine`.

- **valentine:** Cinnabar's vendored crate, framed as `codec.rs`'s
  `decode_batch` frames, without Cinnabar's UI and audio policy checks.
- **rust:** the checked-in `generated/1.26.51/rust-contract` crate, through
  `Frames`, `Packet::decode_frame` with `Limits::receiving_from(sender)` and
  `encode_frame`.
- **go:** the generated Go, with the reader's limits for client packets only.
  An id outside the manifest is forwarded as unknown; a manifest id from the
  wrong sender is a direction error.
- **gophertunnel:** the fork commit `gophertunnel.json` pins, which is also
  the one Cinnabar's fixturegen pins at the corpus commit (`prepare` checks
  they match), used as its connections use it:
  `packet.Decoder` framing, the sender's pool, `packet.Unknown` for absent
  ids, limits for client packets.

Each decoder writes `build/retail/verdicts/<decoder>.tsv`, one line per entry:
fixture, case, entry, header (`id:sender:target`), outcome, error class, then
digests (8 bytes of SHA-256) of the received entry, of the decoder's own
rendering of the decoded value and of the re-encoded entry, whether decoding
the re-encoding renders the same, and a detail. Field renderings are per
decoder, so they are compared only across that round trip.

| Case | Input | Expected |
| --- | --- | --- |
| `retail` | the corpus | decodes, re-encodes byte for byte, reads back the same |
| `crafting` | gophertunnel CraftingData for every recipe list and descriptor kind | as `retail` |
| `large` | generated-Go UpdateSoftEnum with 3000 values | as `retail` |
| `unknown-id` | retail entries under an id nobody defines | kept as unknown, forwarded byte for byte |
| `wrong-sender` | one-direction retail batches from the other peer | direction error |
| `truncated`, `trailing` | retail entries one byte short or long | decode error |
| `element-limit` | retail, Rust and Go with a zero element limit | the two agree; errors are limit errors |
| `alloc-failure` | `large`, Rust with allocations over 64 KiB refused | `AllocationFailed`, no abort |
| `zero-copy` | `unknown-id`, Rust | frame and unknown bodies share the batch; `compact` copies |

A panic is its own outcome and always fails the comparison, even on
malformed input. Startup cancellation is session behavior and is not exercised
here.

An exception names a decoder, case, kind and the fixtures (or `*`) that share
one cause, with a reason and evidence; each listed fixture must still disagree.
Exceptions about gophertunnel carry `oracle`, the module version in
`retail/go.mod`, and stop loading when that pin moves.

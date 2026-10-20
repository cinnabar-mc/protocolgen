# Rust codec contract

The API a Rust consumer such as Cinnabar can rely on from protocolgen's
generated Rust crate. This is version 2.

`protocolgen emit-rust -codec-contract 2` emits it. The default protocol's
crate is checked in under `generated/<version>/rust-contract`, and
`make rust-contract` regenerates it, fails on drift, then lints and tests it.
Without the flag, `emit-rust` produces the profile-neutral crate under
`generated/*/rust`.

Every packet layout comes from the manifest. The contract only covers how the
generated code reads, writes, bounds and reports those layouts.

## Versioning

- There is one crate per game version, named `bedrock-protocol-<version>`.
  `lib.rs` exports `GAME_VERSION`, `PROTOCOL_VERSION`, `CODEC_CONTRACT` and
  `MANIFEST_SHA256`, the SHA-256 of the stable manifest serialization.
- Any change a consumer must adapt to bumps the contract version.
  `emit-rust` rejects versions it does not implement. A consumer pins a
  protocolgen commit and can assert `CODEC_CONTRACT` at compile time.
- Each crate carries its own `wire` runtime, so two protocol versions share
  no types.
- Version 2 adds `DecodeError::RecursiveDepthExceeded` and a bounded recursive-value
  decoder. Consumers upgrading from version 1 must update exhaustive error matches
  and their contract assertion. Wire layouts and the canonical manifest are unchanged.
- Field-site ids (`fields`) follow emission order and can change whenever the
  manifest changes. Use the constants or `fields::find(path)`, never raw
  numbers.

## Decoding: limits and allocation

- Before anything is allocated, every declared length is checked against the
  remaining input, using the smallest wire size of one element from the
  manifest. If the input cannot hold the count, decoding fails with
  `LengthNotRepresentable`, before any schema bound or limit is checked, so
  a truncated packet is never reported as a bound violation.
- `wire::Limits` has a default element ceiling, a default byte ceiling for
  strings and buffers, and per-field overrides by `FieldId`. Every
  length-prefixed site has a constant in `fields`, and `fields::PATHS` names
  it:
  `Owner.field`, `Union::Variant.field`, `Union::Variant` for a payload, plus
  the suffixes `[]` (element), `{key}` and `{value}`. An optional value uses
  its field's site. Owners and fields use their Rust names.
- `Limits::receiving_from(Peer::Server)` sets no length ceiling beyond the input,
  matching gophertunnel's `Dial`, which reads with limits disabled.
  `Limits::receiving_from(Peer::Client)` caps collections at 4096 elements,
  matching the generated Go reader's `maxSliceLength`; gophertunnel's
  `Listener` also enables its limits.
- Every recursive manifest edge acquires capacity from the reader's shared
  `MAX_RECURSIVE_VALUE_DEPTH` budget before decoding its child. This client-owned
  safety ceiling also bounds the resulting tree's recursive clone/drop depth;
  collection and byte limits cannot disable it. Refusal returns
  `RecursiveDepthExceeded` before consuming child bytes. Capacity is restored on
  success or a decode error, while the cursor retains any bytes already read.
  This policy does not state a native nesting maximum. Both Rust profiles use
  the same runtime policy.
- Allocation can fail without aborting. `wire::alloc_vec` reserves at most
  `MAX_EAGER_RESERVE_BYTES` (1 MiB) up front, and `wire::push` grows the
  vector with `try_reserve`. A refused allocation returns `AllocationFailed`.
  When a reader does not share its source, its copies are bounded by the input
  size.
- A limit or schema violation fails only that packet. Its frame boundary is
  already fixed, so the rest of the batch still decodes (see Frames).
- Totals across several fields, such as the bytes of every autocomplete
  option, are left to the consumer. Strings are zero-copy, so those totals can
  be computed after decoding without extra allocation.
- Varints must use their shortest encoding and fit their width. A bool is true
  for any non-zero byte. Unknown enum values decode to `Unknown(raw)`, and an
  unknown union selector is `UnknownVariant`. An enum type shared by several
  fields names the values of every field.
- Bitsets drop bits at or beyond their declared width in both directions, as
  the client does. Only an encoding still unterminated after
  `ceil(width / 7)` bytes is `InvalidBitset`.

## Zero-copy strings and borrowed views

- Strings decode as `wire::Str`, the exact bytes received. Bedrock strings
  are byte strings: the generated Go reader and gophertunnel also keep their
  bytes unchecked. UTF-8 is checked on access through `to_str` or
  `to_string_lossy`, so a malformed string cannot fail a packet and always
  re-encodes byte for byte. Strings with a schema pattern are checked while
  decoding.
- With `Reader::from_shared`, `Decode::decode_shared` or `Packet::decode_frame`,
  strings, byte buffers and NBT are refcounted slices of the batch buffer.
  Decoding a packet allocates only for collections and boxes.
- `wire::RawFrame` is the lazy view. It parses only the header, so a
  consumer can route or forward on the id and decode later.
  `RawFrame::compact` copies one entry out so that a retained packet does
  not keep the whole batch alive.
- This version has no lifetime-parameterized view types (see Remaining work).

## Encoding

- `Encode::encode` returns `EncodeResult<()>`. `wire::Writer` appends to a
  caller-owned `Vec<u8>` and can carry an output limit. Every write reserves
  with `try_reserve`. If the output is too large or the allocator refuses,
  encoding returns an error instead of aborting.
- `wire::encode_into` and `Packet::encode_frame` restore the buffer to its
  previous length on error.
- A length above `u32::MAX` returns `LengthOverflow`, and a value outside its
  schema bounds returns `SchemaConstraint`. Encoding never panics on input
  data.
- `Packet::encode_frame(sender, sender_subclient, target_subclient, out)`
  appends one `[length][header][body]` entry. A batch is `wire::BATCH_HEADER`
  followed by entries.

## Frames and unknown packets

- `wire::Frames` iterates the entries of an uncompressed batch payload.
  Framing errors end the batch: a missing `0xfe` with `from_batch`, a declared
  length past the input, a header that overruns its entry, or more entries
  than the configured maximum.
- `Packet::decode_frame(frame, sender, limits)` decodes one entry:
  - an id missing from the manifest returns `Decoded::Unknown`, with the
    header and the exact body;
  - a known id from the wrong sender is rejected before the body is read;
  - the body must consume its entry exactly, or the result is `TrailingBytes`.
  Every error from `decode_frame` covers that entry only.
- Header bits above the target subclient are ignored, as gophertunnel's
  `Header.Read` does. `UnknownPacket` keeps the entry as received, so
  `UnknownPacket::encode_frame` forwards it byte for byte, those bits
  included.

## Direction checks

- Packet directions come from the manifest's reviewed `directions.json`.
  `PacketId::direction` and `wire::Direction::permits(Peer)` expose them.
- `decode_from` and `decode_frame` reject a packet the sender may not send.
  `encode_frame` refuses to write one the local peer may not send.
  `wire::Direction` and `wire::Peer` are protocol-independent.

## Cinnabar's patches

These are the edits listed in Cinnabar's `crates/protocol/vendor/UPSTREAM.md`
and how each one is now covered:

| Patch | Disposition |
| --- | --- |
| `DisconnectPacket` hide-screen conditional | No correction is needed. From 1.26.40 to 1.26.51 the manifest models `Messages` as a union whose `var_u32` selector is 0 for messages and 1 for none. For bytes 0 and 1 this matches gophertunnel's `HideDisconnectionScreen` bool followed by the two strings only when it is false. Valentine's generator lowered the union to its first variant's struct and dropped the selector; the hand patch restored it. protocolgen's emitter already generates `DisconnectMessages` with the selector. Two differences remain: a selector byte of 2 or more is `UnknownVariant` here, while gophertunnel treats any non-zero byte as hidden; and the empty variant is still named after its placeholder (`Empty1`, emitted as `Empty`). |
| Reserved Education Edition bindings (`tools/protocol-normalize`) | Not a manifest correction. It renames packets 109, 137, 150, 170, 178, 181, 183 and 304 and some enum values and fields, but leaves the wire unchanged. Corrections must carry wire evidence, and renaming the shared manifest would change the Go output too. If Cinnabar keeps the neutral names, protocolgen needs a Rust-only naming profile (see Remaining work). |
| Shared codec hardening | Now part of the contract runtime: fallible collection growth with an out-of-line grow path, length checks before allocation, bounded NBT nesting, UUIDs as two little-endian `u64` halves, lenient strings, exact entry boundaries, and canonical varints. |
| Reviewed corrections (binary buffers, little-endian union arms, actor-ID varints, inventory filtered names, PlayerList selectors) | Already protocolgen corrections under `generated/1.26.51/corrections`. |
| Jolyne login, pack and transport patches | Not codec behavior. They leave with Jolyne once session ownership moves to the Go core. |

## Retail fixtures

`differential/retail/corpus.json` pins Cinnabar's packet fixtures by commit and
SHA-256: 53 uncompressed batches, 34 sent by the server and 19 by the client.
Most were encoded by the gophertunnel fork through `tools/fixturegen`,
including `available_commands_live_356513`, which pads a fixture to an
observed live body length; the two `_live_zeqa` inventory fixtures are not
fixturegen output. So the corpus checks agreement with gophertunnel's layouts.
It does not prove vanilla behavior.

`make retail` compares Valentine, this contract's output, the generated Go
and that gophertunnel commit on every fixture and on derived malformed,
misdirected, unknown, large and crafting inputs. See
[the differential README](../differential/README.md#retail-2193) for the runners,
cases and reviewed exception list.

## Remaining work

- **Make the contract the default output:** flip the default once Cinnabar
  adopts it. Every checked-in version then needs a regeneration run.
- **Lifetime-borrowed views (`&'a str`, slice-backed collections):** not
  implemented. Valentine's `Borrowed*` views have no equivalent, beyond
  zero-copy `Str` and `Bytes` and the lazy `RawFrame`. Add them only if
  profiling shows collection allocation matters on hot packets such as
  LevelChunk and SubChunk.
- **Union-variant naming overlay:** needed for placeholder variants such as
  Disconnect's `Empty1`. Today the naming overlay only covers type ids.
- **Rust-only naming profile:** needed only if Cinnabar keeps the reserved
  Education Edition names.
- **Crafting ingredient descriptors:** the manifest models the ingredient as
  vanilla sends it, a string-to-string map, so it decodes molang (`tags`) and
  item-tag (`item_tag`) maps. Oracle divergence 52 comes from the molang
  writer at the commit `tools/gophertunnel-oracle/lock.json` pins, which sends
  a `molang` key and an int16 version instead; the fork commit the
  differential suites pin writes the `tags` map, and the retail `crafting`
  case decodes it. Typed descriptor variants would be an API addition, not a
  wire fix.
- **Exact encoded size:** `encode_frame` writes the body and then moves it to
  close up the length prefix. There is no `encoded_size` yet, which Valentine
  had.
- **Zero-width collection elements:** these count as one byte when a declared
  count is checked against the input.
- **Field limits:** Cinnabar's values, such as `MAX_UI_TEXT_BYTES` and
  `MAX_CHAT_PARAMETERS`, belong in Cinnabar's `Limits` setup, not in
  protocolgen.

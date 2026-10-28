# Automation readiness — September 16, 2026

CI verifies reproducible generation, tests, and an exact reviewed comparison baseline. Passing CI is not proof that every packet agrees with vanilla. This inventory separates decisions already resolved for a pinned release, new U6 schema gaps, and independent-verification limits.

## Checked-in snapshots

| Snapshot | Protocol | Packets | Adjudications | Correction operations in current files | Recorded manifest overrides |
|---|---:|---:|---:|---:|---:|
| 1.26.40 | 2168 | 229 | 168 | 81 | 81 |
| 1.26.44 | 2168 | 229 | 168 | 0 | 81 |
| 1.26.50 | 2187 | 231 | 173 | 83 | 83 |

The .44 snapshot inherits .40 and applies an evidenced RemoveScore.ObjectiveName optional-wrapper hotfix. These manifests have no unresolved/opaque nodes; their decisions are resolved for their pinned targets. The baseline .50 manifest retained eight obsolete overrides after commit `bbaa61a` deliberately removed their source corrections (TextData.LineGapHeight and seven sound-slot optional markers). Regeneration now removes those stale wrappers and records 83 overrides, matching 83 correction operations.

## U6 raw Mojang gaps

Pinned source `19e25de129227373e97dfa340af93ccf12c0feb9`, Minecraft 1.26.60-beta.23/protocol 2208: 235 packets, 745 top-level field claims, **79 reachable unresolved node occurrences**. This is raw ingestion, not a reconciled U6 release. All 235 packet directions are unknown in this raw source.

The previous 44 selector gaps are resolved by the zero-based VarUInt32 rule for wire variants. No new discriminator field is needed. Entering those alternatives reveals 50 additional unresolved nodes, so the total changes from 73 to 79 rather than falling to 29. See [the rule and PlayerList example](mojang-u6-metadata.md#variant-selectors).

### 50 ItemStackNetIdVariant occurrences (16 fields)

- `AddPlayerPacket.Carried Item`
- `AddItemActorPacket.Item`
- `InventoryTransactionPacket.Transaction` (13 occurrences)
- `MobEquipmentPacket.Item`
- `MobArmorEquipmentPacket.Head`
- `MobArmorEquipmentPacket.Torso`
- `MobArmorEquipmentPacket.Legs`
- `MobArmorEquipmentPacket.Feet`
- `MobArmorEquipmentPacket.Body`
- `InventoryContentPacket.Slots`
- `InventoryContentPacket.Storage Item`
- `InventorySlotPacket.Storage Item`
- `InventorySlotPacket.Item`
- `PlayerAuthInputPacket.Item Use Transaction` (3 occurrences)
- `PlayerAuthInputPacket.Item Stack Request` (11 occurrences)
- `ItemStackRequestPacket.Requests` (11 occurrences)

### 18 untyped payload occurrences

- `StartGamePacket.Block Properties`
- `StartGamePacket.Player Property Data`
- `AddPlayerPacket.Entity Data`
- `AddActorPacket.Actor Data`
- `AddItemActorPacket.Entity Data`
- `SetActorDataPacket.Actor Data`
- `BlockActorDataPacket.Actor Data Tags`
- `UpdateTradePacket.Data`
- `UpdateEquipPacket.Data`
- `AvailableActorIdentifiersPacket.Identifier List`
- `LevelEventGenericPacket.__[[CTD]]__`
- `StructureTemplateDataResponsePacket.Structure's NBT`
- `PositionTrackingDBServerBroadcastPacket.Position tracking data`
- `ItemRegistryPacket.Item Data`
- `SyncActorPropertyPacket.Property Data`
- `AddVolumeEntityPacket.Components`
- `JigsawStructureDataPacket.Jigsaw Structure Data Tag`
- `ClientboundDataStorePacket.Updates`

### 10 colour alternatives

All ten occurrences are inside `ClientboundAttributeLayerSyncPacket.Data`.
The colour schemas describe alternate JSON representations without a wire
selector marker. Their actual binary representation still needs evidence.

### 1 WebToken self-reference

- `ServerToClientHandshakePacket.Handshake WebToken`

The largest remaining metadata win is the precise wire representation of
[ItemStackNetIdVariant](https://github.com/Mojang/bedrock-protocol-docs/blob/19e25de129227373e97dfa340af93ccf12c0feb9/json/ItemStackNetIdVariant.json).
Next are concrete encodings for untyped payloads, including which NBT format
and framing each field uses, followed by colour and WebToken representations.
Byte examples for representative variants would help independently verify
these rules. Also reconcile optionals/defaults, bytes versus text, signedness,
length prefixes, and encode/decode asymmetry against the exact same release.
Zero unresolved nodes alone does not prove wire correctness.

## Existing manual evidence categories

- Source corrections: arbitrary byte buffers versus UTF-8 (login, chunks, item user data, features, hashes, script/debug/photo payloads); default/required/conditional fields; flattened packet structure; discriminator widths and values; fixed-u8 inventory/equipment fields; NBT editor payload; camera target mode; optional booleans; diagnostics arrays; Text category; PrimitiveShapes attached actor runtime ID.

- Adjudications: .40 has 168 complete-field selections; .50 has 173, of which 128 carry prior decisions after byte-equivalence checks and 45 are reviewed new selections. Commands `carry-adjudications` and `adjudicate-claims` already automate safe fingerprinting; humans review changes and evidence, not manually recompute hashes.

- Direction overlays:229/.40 and 231/.50 packet entries, pinned to gophertunnel pool evidence. NBT overlays:18 exact field-path encoding entries per snapshot. Add/update machine extraction with exhaustiveness guards for new versions.

- Consumer metadata overlays: naming 121/.40 and 120/.50 entries; domains 504/.40 and 510/.50; documentation 859/.40 and 834/.50. These are ergonomics/semantic mappings, not unresolved wire layouts. Seeding exists under tools/seed-gophertunnel-overlays; changed/new names and hand-written integration still need review.

- Same-protocol hotfix derivation currently supports only wrap_optional and needs pinned exact-version evidence/base hashes. Other hotfix shapes require implementing an explicit safe operation, not arbitrary replacement.

## Independent oracle verification (protocol 2225)

The `.60` manifest compared with gophertunnel `77149b94f8024826e34efda6f625c366596118d2` has **218 agreements, 21 reviewed divergences, and no unresolved comparisons**. Earlier locks and baselines remain in git history; the helper summaries and interface variant values are reviewed for the locked revision only. The reviewed divergences include extractor limitations and framing differences; they are not proved vanilla bugs. No disputed codec was changed merely to satisfy this oracle.

The comparison is exact: both sides are lowered to a finite wire language and compared as automata, so optional fields and unions never need a cartesian path product and no packet is skipped for size. A disagreement is reported as the shortest wire path on which the two languages differ. Bool-guarded Go fields compare as optionals, a Go conditional or switch on a discriminant read earlier compares as the union variant on that read (with an `else` covering every other discriminant), a constant discriminant written at the start of a type-switch case names its variant, `if ...; return` leaves the rest of the block as the else path, a little-endian colour int and `BEARGB` are the same bytes, and recursive types are unrolled twice on both sides before the remainder is compared as one recursion marker. Only extraction the oracle cannot lower at all (an opaque helper, a runtime loop bound) is `UNRESOLVED`; there are none at this pin.

CI fails on a new or changed divergence. Each reviewed entry pins the complete manifest packet, source operation tree, diverging path, and oracle revision, excluding diagnostic checkout paths. If analysis or input changes, its acceptance fingerprint expires. Entries that stop diverging must also be removed after review. Where CloudburstMC/Protocol breaks a two-way tie, the entry says which side it supports. At protocol 2193 the former 2168 conflicts in StartGame and ServerPresenceInfo presence info, HurtArmor, UpdateAbilities, PlayerUpdateEntityOverrides and ClientboundUpdateSoundData agree. At protocol 2225 the former 2193 conflicts in StructureTemplateDataResponse, ClientboundDebugRenderer and ServerboundDiagnostics agree, and so do StructureBlockUpdate, CameraInstruction, PlayerVideoCapture and CameraSpline after the fork corrected their layouts.

### All reviewed divergences

The authoritative per-packet evidence and settlement criteria are in [accepted-divergences.json](../tools/gophertunnel-oracle/accepted-divergences.json).

| ID | Packet | Remaining question or tooling limitation |
|---:|---|---|
| 5 | `DisconnectPacket` | Wire conflict: the canonical Messages union uses a var_u32 selector; the protocol 2225 Disconnect marshal uses a HideDisconnectionScreen bool to decide whether the message strings follow. |
| 9 | `TextPacket` | Control-layout conflict with identical bytes: both 1.26.60 sources model the category byte as the var_u32 selector of a three-way body union whose cases start with the u8 message type, while the protocol 2225 Text marshal writes the category and type as two u8 bytes and dispatches on the type. Category values 0 to 2 encode as the same single byte either way. |
| 12 | `AddPlayerPacket` | Extractor limitation: ItemInstance extra-data is decoded from one byte buffer and its canBePlacedOn/canBreak arrays are parsed inside it, so the extractor counts them as outer operations where the canonical manifest has one opaque user-data buffer. Entity metadata also differs in variant-marker placement, as in 13. |
| 13 | `AddActorPacket` | Control-layout conflict with identical bytes: both sides write a var_u32 metadata selector and a u8 legacy type byte; the canonical union places its variant marker after the selector and carries the type byte inside each case, while the protocol 2225 EntityMetadata writer dispatches on the value's Go type after both bytes. |
| 15 | `AddItemActorPacket` | Extractor limitation: ItemInstance extra-data is decoded from one byte buffer and its canBePlacedOn/canBreak arrays are parsed inside it, so the extractor counts them as outer operations where the canonical manifest has one opaque user-data buffer. |
| 30 | `InventoryTransactionPacket` | Control-layout and extractor limitation: the protocol 2225 marshal selects the transaction data through a reviewed interface table, so its variant marker follows the actions array rather than the selector. ItemInstance extra-data is decoded from one byte buffer and its canBePlacedOn/canBreak arrays are parsed inside it, so the extractor counts them as outer operations where the canonical manifest has one opaque user-data buffer. |
| 31 | `MobEquipmentPacket` | Extractor limitation: ItemInstance extra-data is decoded from one byte buffer and its canBePlacedOn/canBreak arrays are parsed inside it, so the extractor counts them as outer operations where the canonical manifest has one opaque user-data buffer. |
| 32 | `MobArmorEquipmentPacket` | Extractor limitation: ItemInstance extra-data is decoded from one byte buffer and its canBePlacedOn/canBreak arrays are parsed inside it, so the extractor counts them as outer operations where the canonical manifest has one opaque user-data buffer. |
| 39 | `SetActorDataPacket` | Extractor/control-layout conflict: EntityMetadata writes its cereal selector and legacy type byte before dispatch, while the canonical union annotates its variant earlier. This is not sufficient evidence of a wire width defect. |
| 49 | `InventoryContentPacket` | Extractor limitation: ItemInstance extra-data is decoded from one byte buffer and its canBePlacedOn/canBreak arrays are parsed inside it, so the extractor counts them as outer operations where the canonical manifest has one opaque user-data buffer. |
| 50 | `InventorySlotPacket` | Wire conflict: Mojang declares Container Id as uint8, and 8- and 16-bit fields are fixed-width regardless of Compression, while the protocol 2225 InventorySlot marshal writes WindowID as var_u32. The encodings agree below 128, so the fork is the suspect side. ItemInstance extra-data also shows the extractor limitation recorded for 12 and 15. |
| 52 | `CraftingDataPacket` | Modeling limitation: the canonical shaped-recipe ingredient (Endstone cerealizer<RecipeIngredient>::SerializedData) is a var_u32-counted string-to-string map followed by zigzag aux value and count, while the protocol 2225 ItemDescriptorCount and CloudburstMC readIngredient write a var_u32 descriptor kind, a descriptor-name string, a payload chosen by that name, then the zigzag count. The map is byte-identical for the default and invalid descriptors and diverges for molang and item-tag descriptors. A descriptor discriminated by a string is not expressible in the claim format, so no correction is possible without a manifest schema extension. |
| 63 | `PlayerListPacket` | Extractor limitation: PlayerListEntry uses an early return for removal entries, but the extractor also includes the subsequent addition payload. Selector placement and skin-colour endianness require separate verification after early-return handling. |
| 65 | `LegacyTelemetryEventPacket` | Wire/control conflict: the LegacyTelemetry event union and the protocol 2225 Event marshaler disagree on event discriminants and payload fields, including Contents Color. A packet-level exception does not establish any event as correct. |
| 78 | `CommandBlockUpdatePacket` | Wire conflict: the canonical Target union uses var_u32, while protocol 2225 CommandBlockUpdate uses a bool to distinguish block and actor targets. |
| 108 | `SetScorePacket` | Control-layout conflict with identical bytes: both sides now write a var_u32 selector; the canonical union places its variant marker right after the selector, while the protocol 2225 ScoreboardEntry writes the type-name string and entry ID before dispatching. |
| 144 | `PlayerAuthInputPacket` | Control-layout conflict, same as 147: the item stack request action selector, the per-action type byte and the switch annotation appear at different positions. The item-use framing difference recorded at protocol 2168 no longer occurs, because the protocol 2225 fork writes the actions array without the extra presence flag. |
| 145 | `CreativeContentPacket` | Extractor limitation: ItemInstance extra-data is decoded from one byte buffer and its canBePlacedOn/canBreak arrays are parsed inside it, so the extractor counts them as outer operations where the canonical manifest has one opaque user-data buffer. |
| 147 | `ItemStackRequestPacket` | Extractor/control-layout conflict: ItemStackRequest cereal action selector, legacy action byte and switch annotation appear at different positions; action-ID mappings also need exact evidence. Removing marker tokens would not prove the cases correspond. |
| 174 | `SubChunkPacket` | Modeling difference: 1.26.60 Endstone and the protocol 2225 fork both prefix each height map row with a var_u32 count, but the fork requires the count to be 16 and reads exactly 16 heights, while the canonical row is a general var_u32-counted array. Vanilla always writes 16. |
| 345 | `ClientboundAttributeLayerSyncPacket` | Wire conflict: the canonical attribute operation is an unconditional string, while the protocol 2225 ClientBoundAttributeLayerSync wraps BoolOperation in an optional. CloudburstMC readAttributeData (v944, inherited by v2168) reads the operation string unconditionally after the value, matching the manifest, so the protocol 2225 gophertunnel marshal is the suspect side. |

Five deprecated oracle-only IDs are intentionally excluded from the generated pools: 55 `AdventureSettings`, 117 `ScriptCustomEvent`, 163 `FilterText`, 173 `PhotoInfoRequest`, 197 `ClientCheatAbility`. This is a coverage boundary, not a request to restore obsolete packets.

### Differential byte-test allowances

The separate executable differential harness targets protocol 2225. Its committed corpus covers shared packet pools, but this does not prove every branch or value. `differential/accepted-divergences.json` has **43 allowances across 32 packet IDs**: 36 re-encoded-byte allowances and 7 decode-success allowances. All use empty `input_hex`, so they apply to a packet/direction/error class rather than one exact sample. These are broader than the new AST comparison fingerprints and should be narrowed to semantic predicates or exact reproductions before unattended promotion.

Affected IDs: 8, 9, 11, 12, 15, 30, 31, 32, 49, 50, 52, 56, 65, 80, 81, 86, 93, 108, 119, 124, 148, 153, 162, 165, 166, 184, 190, 313, 326, 347. The file records each rationale. Many concern oracle canonicalization, derived values, reserved bytes, or malformed-input tolerance; they are not automatically generated-code defects.

## What remains for an automated release pipeline

1. Automatically detect matching Mojang/Endstone releases, verify full version/protocol/tree locks, prepare a new version directory, and run raw-claim/reconciliation reports. Current Make targets and CI pins are version-specific. U6 still needs a separately pinned matching runtime graph and all evidence/overlays; the enum improvement alone is not a U6 release.

2. Group changed disagreements by reusable type and present exact evidence to review. Safely carry unchanged adjudications using the existing command; never blindly carry corrections or let one source silently win. Add official metadata to remove manual decisions where it actually supplies equivalent evidence.

3. Resolve the remaining raw primitive, untyped payload, colour and WebToken gaps above and verify source optionality and directional asymmetry. Some require better metadata/frontends, some a wire capture or binary serializer trace. Human adjudication remains appropriate whenever the upstream sources contradict.

4. Complete independent verification: comparator control-flow/recursive/encapsulated-buffer handling; per-version oracle locks and coverage; live decode/encode corpus of rare branches. Axolotl parity is currently a synthetic small check in make regen; the parity adapter rejects asymmetric fields, and the CI parity command exercises only a small synthetic fixture. Recursive and encapsulated markers already have adapter support.

5. Complete downstream integration: Go already emits concrete Reader/Writer implementations and packet pools. The remaining integration work is adopting generated packet APIs in consumers and validating their actual network behavior, not writing a byte runtime from scratch. The update guide is transcription assistance. Rust borrowed views are optional ergonomics, not a wire automation blocker.

6. Capture BDS data after corrections with exact binary/codec pins; the capture workflow requires explicit EULA acceptance and does not automatically commit captures. New versions need a matching generated-codec binding: the current build tags select `.40` or `.44`, not `.50`. Matching Endstone/BDS pins are required for registry export; .44 has no matching Endstone pin. PMMP creative/recipe outputs still lack canonical block-item/legacy mapping; runtime registry cannot invent it. New versions require generated decoder binding and gophertunnel pin updates.

7. CI now regenerates `.40`, its `.44` hotfix, `.50`, `.51` and `.60`, checks both languages, and uploads the protocol 2225 oracle report. The remaining release-orchestration work is a generic version matrix, per-version independent-oracle/differential coverage, and an evidence gate for automatic promotion.

A realistically fully automated flow can prepare, regenerate, compare and validate an update; it should stop for human evidence only on genuinely new ambiguities. Removing that stop without new evidence would make the generator less trustworthy.

# Data provenance

The semantic generator and initial data snapshot were moved from
[bedrock-mc/bedrock-data](https://github.com/bedrock-mc/bedrock-data) at commit
`a852f424`.

The semantic values are derived from:

- Native block palettes, scalar properties and tags, item fields and tags, and
  biome scalar values and numeric IDs from the CC0-1.0
  [didntpot/bedrock-data](https://github.com/didntpot/bedrock-data) snapshot.
  Selected original files and its licence are retained in
  [source/inputs/1.26.60/native](source/inputs/1.26.60/native/README.md).
  The active snapshot is `release/1.26.60` at commit
  `1e041bf547b7c12e94d2b998f68d5ed146975742`, with upstream file hashes
  verified against `versions.json`.
- Block geometry and rendering fields, named voxel shapes, biome tag memberships,
  data-driven block components and complete item registry identities/components
  from [CloudburstMC/Data](https://github.com/CloudburstMC/Data)
- Base entity components and food definitions distributed
  with Minecraft Bedrock Dedicated Server
- The offhand/head inventory-capability tags from
  [AllayMC/Allay](https://github.com/AllayMC/Allay), plus the documented corrections
  in the source inputs
- Resource inputs from [Mojang/bedrock-samples](https://github.com/Mojang/bedrock-samples)
- Provisional sword/shears speed classifications from
  [PrismarineJS/minecraft-data](https://github.com/PrismarineJS/minecraft-data),
  pinned to Bedrock 1.26.30 `blocks.json`. Its README declares MIT. Native
  tool requirements replace its harvest classifications; special-tool speed
  targets remain unverified for the catalog's target release.

The source revisions and versions are recorded in
[`semantic_sources.json`](../generated/data/semantic_sources.json), with input digests in
[`source/lock.json`](source/lock.json). Upstream data and Minecraft
materials remain subject to their respective terms. No full BDS source pack,
proprietary source catalog or executable is included in this module.

The pinned Cloudburst extract contains 30 invalid liquid-clip boxes. Their exact
source values are recorded as omissions; the catalog reports those shapes as
unavailable. The public [Endstone data exporter](https://github.com/EndstoneMC/endstone/blob/3491c609ddfde392cee2b062e3063e39aae87274/src/endstone/core/devtools/vanilla_data.cpp)
illustrates why an output box without the query's success flag cannot establish
valid geometry. This is not a claim about the exact tool revision that produced
the Cloudburst files.

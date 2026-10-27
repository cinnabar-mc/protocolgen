# Native Bedrock measurements

Selected CC0-1.0 data from [didntpot/bedrock-data](https://github.com/didntpot/bedrock-data),
commit `1e041bf547b7c12e94d2b998f68d5ed146975742`, directory `release/1.26.60`
(protocol 2225). The original files are retained byte for byte. Their SHA-256
digests match the upstream `versions.json` and are pinned in the parent source
lock. See LICENSE for the upstream licence.

The palette supplies the native state order and identifier domain. State
properties use palette ordinals; supplemental geometry joins by hash and name.
The tag tables are sparse maps. Item tags in this snapshot are supplied by
`item_tags.json`, rather than repeated in `item_properties_table.json`.

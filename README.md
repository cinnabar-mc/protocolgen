# protocolgen

Generates Bedrock packet codecs for Go and Rust, plus a shared Go game-data
catalog. Mojang's [protocol docs](https://github.com/Mojang/bedrock-protocol-docs)
and Endstone's [BDS dump](https://github.com/EndstoneMC/protocol-docs) are
reconciled into one version-pinned manifest per release; disagreements are
settled only by fingerprinted corrections, and the emitters read only that
manifest.

[docs/outputs.md](docs/outputs.md) lists everything the repo generates and who
uses it. [docs/protocolgen-v2.md](docs/protocolgen-v2.md) explains the manifest,
corrections and adjudications.

## Regenerate

Requires Go 1.26, jq, and Rust (edition 2024) with rustfmt.

```sh
# Rust crates from the committed manifests; no doc checkouts needed
make regen-rust

# Full pipeline for the current release
make regen-1.26.60 \
  MOJANG_DIR=/path/to/bedrock-protocol-docs/json \
  ENDSTONE_DIR=/path/to/protocol-docs
```

Check out both doc sources at the revisions in the release's
`source-lock.json`. Mojang's docs are EULA-restricted and never committed here.

`go run ./cmd/protocolgen <command> -h` documents each CLI step; with no
arguments it lists the commands.

## License

Cinnabar's authored generator, runtime and generated implementation code is
available under [MIT](LICENSE-MIT). This grant covers only our copyrightable
contributions; imported code, source documentation, schemas and game-data
materials retain their own terms. See [NOTICE](NOTICE) for the scope and
Gophertunnel attribution, and [data/NOTICE.md](data/NOTICE.md) for data sources.

Both codec emitters include the same license and notice in their output.
Their canonical copies live in `internal/licensing/`; the emitter tests check
that these root copies stay identical.

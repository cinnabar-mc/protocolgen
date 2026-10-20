SHELL := /bin/sh

GO_CACHE ?= /tmp/go-build-cache
GO := GOCACHE=$(GO_CACHE) go
PROTOCOLGEN := $(GO) run ./cmd/protocolgen

MANIFEST := generated/1.26.40/manifest.json
NAMING := generated/1.26.40/naming.json
DOMAINS := generated/1.26.40/domains.json
DOCS := generated/1.26.40/docs.json
SOURCE_LOCK := generated/1.26.40/source-lock.json
MOJANG_DIR ?=
ENDSTONE_DIR ?=
GOPHERTUNNEL_DIR ?=
LAYOUT_GOPHERTUNNEL_DIR ?= $(if $(GOPHERTUNNEL_DIR),$(GOPHERTUNNEL_DIR),../gophertunnel)
RELEASE_INDEX := data/source/releases.json
ORACLE_REPORT ?= /tmp/protocolgen-gophertunnel-report.json
GOPHER_ARGS = $(if $(GOPHERTUNNEL_DIR),-gophertunnel $(GOPHERTUNNEL_DIR))

VANILLA_VERSION ?= $(shell jq -er .default_capture $(RELEASE_INDEX))
VANILLA_MANIFEST := ../generated/$(VANILLA_VERSION)/manifest.json
VANILLA_SOURCE := ../generated/$(VANILLA_VERSION)/vanilla-source.json
VANILLA_OUT := ../generated/$(VANILLA_VERSION)/vanilla-data
VANILLA_GO_ARGS = $(shell jq -er --arg version "$(VANILLA_VERSION)" '.releases[$$version].capture.go_args | join(" ")' $(RELEASE_INDEX))
BDS_ADDRESS ?= 127.0.0.1:19132
BDS_BINARY ?=

HOTFIX_MANIFEST := generated/1.26.44/manifest.json
HOTFIX_SPEC := generated/1.26.44/hotfix.json
HOTFIX_NAMING := generated/1.26.44/naming.json
HOTFIX_DOMAINS := generated/1.26.44/domains.json
HOTFIX_DOCS := generated/1.26.44/docs.json

TARGET_12650 := generated/1.26.50
CLAIMS_12650_MOJANG ?= /tmp/protocolgen-1.26.50-mojang-claims.json
CLAIMS_12650_ENDSTONE ?= /tmp/protocolgen-1.26.50-endstone-claims.json

TARGET_12651 := generated/1.26.51
CLAIMS_12651_MOJANG ?= /tmp/protocolgen-1.26.51-mojang-claims.json
CLAIMS_12651_ENDSTONE ?= /tmp/protocolgen-1.26.51-endstone-claims.json

CANDIDATE_12660 := candidates/1.26.60-preview.25
CLAIMS_12660_MOJANG ?= /tmp/protocolgen-1.26.60-preview.25-mojang-claims.json
CLAIMS_12660_ENDSTONE ?= /tmp/protocolgen-1.26.60-preview.25-endstone-claims.json

.PHONY: regen regen-1.26.50 regen-1.26.51 ingest-1.26.60 hotfix vanilla-data differential verify verify-1.26.50 verify-1.26.51 rust-contract emit-rust regen-rust verify-rust

differential:
	$(GO) -C differential test ./...

ingest-1.26.60:
	@test -n "$(MOJANG_DIR)" || (echo "MOJANG_DIR is required" >&2; exit 2)
	@test -n "$(ENDSTONE_DIR)" || (echo "ENDSTONE_DIR is required" >&2; exit 2)
	$(PROTOCOLGEN) ingest \
		-lock $(CANDIDATE_12660)/source-lock.json -kind mojang -id mojang \
		-root $(MOJANG_DIR) -corrections $(CANDIDATE_12660)/corrections/mojang \
		-out $(CLAIMS_12660_MOJANG)
	$(PROTOCOLGEN) ingest \
		-lock $(CANDIDATE_12660)/source-lock.json -kind endstone -id endstone \
		-root $(ENDSTONE_DIR) -out $(CLAIMS_12660_ENDSTONE)

regen:
	@test -n "$(MOJANG_DIR)" || (echo "MOJANG_DIR is required" >&2; exit 2)
	@test -n "$(ENDSTONE_DIR)" || (echo "ENDSTONE_DIR is required" >&2; exit 2)
	$(PROTOCOLGEN) reconcile \
		-lock $(SOURCE_LOCK) \
		-directions generated/1.26.40/directions.json \
		-nbt-encodings generated/1.26.40/nbt-encodings.json \
		-mojang $(MOJANG_DIR) \
		-mojang-corrections generated/1.26.40/corrections/mojang \
		-endstone $(ENDSTONE_DIR) \
		-endstone-corrections generated/1.26.40/corrections/endstone \
		-adjudications generated/1.26.40/adjudications.json \
		-out $(MANIFEST)
	$(PROTOCOLGEN) validate -manifest $(MANIFEST)
	$(PROTOCOLGEN) emit-go \
		-manifest $(MANIFEST) \
		-naming $(NAMING) \
		-domains $(DOMAINS) \
		-docs $(DOCS) \
		-out generated/1.26.40/go \
		-protocol-import protocolgen/generated/1.26.40/go/protocol
	$(call emit-rust,$(MANIFEST),generated/1.26.40/rust,-naming $(NAMING) -domains $(DOMAINS) -docs $(DOCS))
	$(PROTOCOLGEN) parity \
		-manifest testdata/parity/v2-small.json \
		-axolotl testdata/parity/axolotl-v1-small.json
	$(GO) build ./...
	$(GO) test ./...
	$(GO) vet ./...

regen-1.26.50:
	@test -n "$(MOJANG_DIR)" || (echo "MOJANG_DIR is required" >&2; exit 2)
	@test -n "$(ENDSTONE_DIR)" || (echo "ENDSTONE_DIR is required" >&2; exit 2)
	$(PROTOCOLGEN) ingest \
		-lock $(TARGET_12650)/source-lock.json -kind mojang -id mojang \
		-root $(MOJANG_DIR) -corrections $(TARGET_12650)/corrections/mojang \
		-out $(CLAIMS_12650_MOJANG)
	$(PROTOCOLGEN) ingest \
		-lock $(TARGET_12650)/source-lock.json -kind endstone -id endstone \
		-root $(ENDSTONE_DIR) -corrections $(TARGET_12650)/corrections/endstone \
		-out $(CLAIMS_12650_ENDSTONE)
	$(PROTOCOLGEN) reconcile-claims \
		-lock $(TARGET_12650)/source-lock.json \
		-claims $(CLAIMS_12650_MOJANG) \
		-claims $(CLAIMS_12650_ENDSTONE) \
		-claims $(TARGET_12650)/lens-claims.json \
		-adjudications $(TARGET_12650)/adjudications.json \
		-directions $(TARGET_12650)/directions.json \
		-nbt-encodings $(TARGET_12650)/nbt-encodings.json \
		-out $(TARGET_12650)/manifest.json
	$(PROTOCOLGEN) validate -manifest $(TARGET_12650)/manifest.json
	$(PROTOCOLGEN) emit-go \
		-manifest $(TARGET_12650)/manifest.json \
		-naming $(TARGET_12650)/naming.json \
		-domains $(TARGET_12650)/domains.json \
		-docs $(TARGET_12650)/docs.json \
		-out $(TARGET_12650)/go \
		-protocol-import protocolgen/generated/1.26.50/go/protocol
	$(call emit-rust-snapshot,$(TARGET_12650),rust)
	$(GO) test ./generated/1.26.50/go/... -count=1
	cargo test --manifest-path $(TARGET_12650)/rust/Cargo.toml

regen-1.26.51:
	@test -n "$(MOJANG_DIR)" || (echo "MOJANG_DIR is required" >&2; exit 2)
	@test -n "$(ENDSTONE_DIR)" || (echo "ENDSTONE_DIR is required" >&2; exit 2)
	$(PROTOCOLGEN) ingest \
		-lock $(TARGET_12651)/source-lock.json -kind mojang -id mojang \
		-root $(MOJANG_DIR) -corrections $(TARGET_12651)/corrections/mojang \
		-out $(CLAIMS_12651_MOJANG)
	$(PROTOCOLGEN) ingest \
		-lock $(TARGET_12651)/source-lock.json -kind endstone -id endstone \
		-root $(ENDSTONE_DIR) -corrections $(TARGET_12651)/corrections/endstone \
		-out $(CLAIMS_12651_ENDSTONE)
	$(PROTOCOLGEN) reconcile-claims \
		-lock $(TARGET_12651)/source-lock.json \
		-claims $(CLAIMS_12651_MOJANG) \
		-claims $(CLAIMS_12651_ENDSTONE) \
		-adjudications $(TARGET_12651)/adjudications.json \
		-directions $(TARGET_12651)/directions.json \
		-nbt-encodings $(TARGET_12651)/nbt-encodings.json \
		-out $(TARGET_12651)/manifest.json
	$(PROTOCOLGEN) validate -manifest $(TARGET_12651)/manifest.json
	$(PROTOCOLGEN) emit-go \
		-manifest $(TARGET_12651)/manifest.json \
		-naming $(TARGET_12651)/naming.json \
		-domains $(TARGET_12651)/domains.json \
		-docs $(TARGET_12651)/docs.json \
		-out $(TARGET_12651)/go \
		-protocol-import protocolgen/generated/1.26.51/go/protocol
	$(call emit-rust-snapshot,$(TARGET_12651),rust)
	$(call emit-rust-contract,$(TARGET_12651))
	$(PROTOCOLGEN) verify-gophertunnel \
		-manifest $(TARGET_12651)/manifest.json \
		-report $(ORACLE_REPORT) $(GOPHER_ARGS)
	$(GO) test ./generated/1.26.51/go/... -count=1
	cargo test --manifest-path $(TARGET_12651)/rust/Cargo.toml

hotfix:
	$(PROTOCOLGEN) hotfix \
		-base $(MANIFEST) \
		-spec $(HOTFIX_SPEC) \
		-out $(HOTFIX_MANIFEST)
	$(PROTOCOLGEN) validate -manifest $(HOTFIX_MANIFEST)
	$(PROTOCOLGEN) emit-go \
		-manifest $(HOTFIX_MANIFEST) \
		-naming $(HOTFIX_NAMING) \
		-domains $(HOTFIX_DOMAINS) \
		-docs $(HOTFIX_DOCS) \
		-out generated/1.26.44/go \
		-protocol-import protocolgen/generated/1.26.44/go/protocol
	$(call emit-rust,$(HOTFIX_MANIFEST),generated/1.26.44/rust,-naming $(HOTFIX_NAMING) -domains $(HOTFIX_DOMAINS) -docs $(HOTFIX_DOCS))

vanilla-data:
	@jq -e --arg version "$(VANILLA_VERSION)" '.releases[$$version].capture != null' $(RELEASE_INDEX) >/dev/null || (echo "No capture adapter for $(VANILLA_VERSION)" >&2; exit 2)
	@test -n "$(BDS_BINARY)" || (echo "BDS_BINARY is required" >&2; exit 2)
	$(GO) -C vanilla-data run $(VANILLA_GO_ARGS) ./cmd/vanilla-data \
		-manifest $(VANILLA_MANIFEST) \
		-source $(VANILLA_SOURCE) \
		-out $(VANILLA_OUT) \
		-bds-binary $(BDS_BINARY) \
		-address $(BDS_ADDRESS)

verify: regen hotfix differential
	@test -z "$$(git status --porcelain -- generated)" || (echo "regeneration produced drift:" >&2; git status --short -- generated >&2; exit 1)

verify-1.26.50: regen-1.26.50
	@test -z "$$(git status --porcelain -- $(TARGET_12650))" || (echo "1.26.50 regeneration produced drift" >&2; exit 1)

verify-1.26.51: regen-1.26.51
	@test -z "$$(git status --porcelain -- $(TARGET_12651))" || (echo "1.26.51 regeneration produced drift" >&2; exit 1)

# Emit manifest $(1) to $(2), passing optional profile/overlay flags $(3), then
# normalize every Rust source and integration test to its checked-in representation.
define emit-rust
	$(PROTOCOLGEN) emit-rust -manifest $(1) -out $(2) $(3)
	cargo fmt --manifest-path $(2)/Cargo.toml
endef

# Emit one snapshot directory $(1) into subdirectory $(2), requiring its reviewed
# overlays and passing optional profile flags $(3) through the shared normalization.
define emit-rust-snapshot
	$(call emit-rust,$(1)/manifest.json,$(1)/$(2),-naming $(1)/naming.json -domains $(1)/domains.json -docs $(1)/docs.json $(3))
endef

# Emit the documented Rust codec contract into $(1)/rust-contract.
define emit-rust-contract
	$(call emit-rust-snapshot,$(1),rust-contract,-codec-contract 2)
endef

# Regenerate the default protocol's codec-contract crate, fail on drift from
# the checked-in copy, then lint and test it.
CONTRACT_TARGET = generated/$(shell jq -er .default_protocol $(RELEASE_INDEX))

# Regenerate ordinary Rust for SNAPSHOT_DIR using checked manifest inputs only.
emit-rust:
	@test -n "$(SNAPSHOT_DIR)" || (echo "SNAPSHOT_DIR is required" >&2; exit 2)
	$(call emit-rust-snapshot,$(SNAPSHOT_DIR),rust)

# Every release snapshot and the default contract profile share the production
# emission/format path without requiring external documentation source checkouts.
regen-rust:
	@set -eu; \
	snapshots=$$(jq -er '[.releases[].snapshot] | unique | .[]' $(RELEASE_INDEX)); \
	for snapshot in $$snapshots; do \
		$(MAKE) emit-rust SNAPSHOT_DIR="generated/$$snapshot"; \
	done
	$(call emit-rust-contract,$(CONTRACT_TARGET))

verify-rust: regen-rust
	@set -eu; \
	paths=$$(jq -er '[.releases[].snapshot | "generated/\(.)/rust"] + ["generated/\(.default_protocol)/rust-contract"] | unique | .[]' $(RELEASE_INDEX)); \
	drift=$$(git status --porcelain --untracked-files=all -- $$paths); \
	if test -n "$$drift"; then \
		echo "Rust regeneration produced drift:" >&2; \
		printf '%s\n' "$$drift" >&2; \
		exit 1; \
	fi

rust-contract:
	$(call emit-rust-contract,$(CONTRACT_TARGET))
	@test -z "$$(git status --porcelain -- $(CONTRACT_TARGET)/rust-contract)" || (echo "codec-contract regeneration produced drift:" >&2; git status --short -- $(CONTRACT_TARGET)/rust-contract >&2; exit 1)
	cargo clippy --manifest-path $(CONTRACT_TARGET)/rust-contract/Cargo.toml -- -D warnings
	cargo test --manifest-path $(CONTRACT_TARGET)/rust-contract/Cargo.toml

# Run Cinnabar's pinned retail fixtures and the committed synthetic corpus
# through Valentine, the codec-contract crate, the generated Go and
# gophertunnel, and fail on any disagreement that
# differential/retail/exceptions.json does not accept (see differential/README.md).
RETAIL := differential/retail
RETAIL_OUT := build/retail
RETAIL_CINNABAR := build/cinnabar
RETAIL_VALENTINE := build/valentine
RETAIL_RUNNER = cargo run --manifest-path $(RETAIL)/runner/Cargo.toml --target-dir build/retail-target --

.PHONY: retail

retail:
	scripts/fetch-cinnabar.sh $(RETAIL)/corpus.json $(RETAIL_CINNABAR) $(RETAIL_VALENTINE)
	rm -rf $(RETAIL_OUT)
	$(GO) -C $(RETAIL) vet ./...
	$(GO) -C $(RETAIL) test ./...
	$(GO) -C $(RETAIL) run ./cmd/retail prepare -cinnabar $(CURDIR)/$(RETAIL_CINNABAR) -out $(CURDIR)/$(RETAIL_OUT)
	$(GO) -C $(RETAIL) run ./cmd/retail decode -out $(CURDIR)/$(RETAIL_OUT)
	$(RETAIL_RUNNER) rust $(RETAIL_OUT)
	$(RETAIL_RUNNER) valentine $(RETAIL_OUT)
	$(GO) -C $(RETAIL) run ./cmd/retail compare -cinnabar $(CURDIR)/$(RETAIL_CINNABAR) -out $(CURDIR)/$(RETAIL_OUT)
	cargo test --manifest-path $(RETAIL)/bench/Cargo.toml --target-dir build/retail-target

# Run gophertunnel's and the contract crate's codec benches over testdata/bench
# and print bench-compare's table (see .github/workflows/bench.yml).
BENCH_OUT := build/bench
BENCH_GO_PROCS ?= 4
BENCH_CARGO = cargo bench --manifest-path $(RETAIL)/bench/Cargo.toml --target-dir build/bench-target

.PHONY: bench

bench:
	mkdir -p $(BENCH_OUT)
	$(GO) -C $(RETAIL) test -run '^$$' -bench . -benchmem -count 5 -cpu $(BENCH_GO_PROCS) . > $(BENCH_OUT)/go.txt
	rm -rf build/bench-target/criterion
	CRITERION_HOME=$(CURDIR)/build/bench-target/criterion $(BENCH_CARGO) --bench codec -- --warm-up-time 1 --measurement-time 3
	$(BENCH_CARGO) --bench allocs
	$(GO) -C $(RETAIL) run ./cmd/retail bench-check -go $(CURDIR)/$(BENCH_OUT)/go.txt -count 5 -criterion $(CURDIR)/build/bench-target/criterion
	$(GO) -C $(RETAIL) tool bench-compare --go-procs $(BENCH_GO_PROCS) --go $(CURDIR)/$(BENCH_OUT)/go.txt --criterion $(CURDIR)/build/bench-target/criterion

# Emit the 1.26.51 tree laid out like the gophertunnel checkout in
# LAYOUT_GOPHERTUNNEL_DIR (constants beside their packets under fork names, fork field
# names) into build/gophertunnel-layout, and refresh the seeded overlay and gap
# report. The overlay is never applied to the checked-in generated tree.
LAYOUT_TARGET = generated/$(shell jq -er .default_protocol $(RELEASE_INDEX))

gophertunnel-layout:
	$(GO) run ./tools/seed-gophertunnel-layout \
		-manifest $(LAYOUT_TARGET)/manifest.json \
		-naming $(LAYOUT_TARGET)/naming.json \
		-gophertunnel "$(LAYOUT_GOPHERTUNNEL_DIR)" \
		-out $(LAYOUT_TARGET)/gophertunnel-layout.json \
		-docs $(LAYOUT_TARGET)/docs.json \
		-semantics-out $(LAYOUT_TARGET)/semantics.json \
		-report docs/gophertunnel-gap-$(notdir $(LAYOUT_TARGET)).md
	rm -rf build/gophertunnel-layout
	$(PROTOCOLGEN) emit-go \
		-manifest $(LAYOUT_TARGET)/manifest.json \
		-naming $(LAYOUT_TARGET)/naming.json \
		-domains $(LAYOUT_TARGET)/domains.json \
		-docs $(LAYOUT_TARGET)/docs.json \
		-layout $(LAYOUT_TARGET)/gophertunnel-layout.json \
		-out build/gophertunnel-layout \
		-protocol-import protocolgen/build/gophertunnel-layout/protocol

CLOUDBURST_DIR ?=
BDS_DIR ?=

.PHONY: test-data generate-data verify-data

test-data:
	$(GO) -C data test ./...
	$(GO) -C data vet ./...
	$(GO) -C generated/data test ./...
	$(GO) -C generated/data vet ./...

generate-data:
	$(GO) -C data run ./cmd/generate -cloudburst "$(CLOUDBURST_DIR)" -bds "$(BDS_DIR)"
	$(GO) -C data run ./cmd/runtimegen
	$(GO) -C data run ./cmd/mininggen
	$(GO) -C data run ./cmd/registrygen

verify-data: generate-data
	git diff --exit-code -- generated/data
	@test -z "$$(git ls-files --others --exclude-standard -- generated/data)" || (echo "generation added untracked catalog files" >&2; exit 1)

#!/usr/bin/env bash
# Checks out what a retail corpus pins from Cinnabar, sparsely: the fixtures
# and fixturegen's go.mod at the corpus commit into CHECKOUT_DIR, and the
# vendored Valentine codec at its own commit into VALENTINE_DIR.
set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "usage: $0 CORPUS_JSON CHECKOUT_DIR VALENTINE_DIR" >&2
  exit 2
fi

corpus=$1
repository=$(jq -er .source.repository "$corpus")

# checkout DIR COMMIT PATH... sparsely checks out PATHs of COMMIT into DIR.
checkout() {
  local dir=$1 commit=$2
  shift 2
  if [[ "$(git -C "$dir" rev-parse --verify --quiet HEAD 2>/dev/null || true)" == "$commit" ]]; then
    return
  fi
  if [[ ! -d "$dir/.git" ]]; then
    git init --quiet "$dir"
    git -C "$dir" remote add origin "$repository"
    git -C "$dir" config remote.origin.promisor true
    git -C "$dir" config remote.origin.partialclonefilter blob:none
  fi
  git -C "$dir" sparse-checkout set --no-cone "$@"
  git -C "$dir" fetch --quiet --depth 1 --filter=blob:none origin "$commit"
  git -C "$dir" checkout --quiet --detach "$commit"
  test "$(git -C "$dir" rev-parse HEAD)" == "$commit"
}

checkout "$2" "$(jq -er .source.commit "$corpus")" "/$(jq -er .source.directory "$corpus")/" /tools/fixturegen/go.mod
checkout "$3" "$(jq -er .valentine.commit "$corpus")" "/$(jq -er .valentine.directory "$corpus")/"

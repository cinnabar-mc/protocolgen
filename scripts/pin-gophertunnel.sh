#!/usr/bin/env bash
# Points each differential module at the gophertunnel fork version in its gophertunnel.json; VERSION
# rewrites the harness lock first.
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)
lock=$root/differential/gophertunnel.json
if [[ $# -eq 1 ]]; then
  jq --arg version "$1" '.version = $version' "$lock" > "$lock.tmp"
  mv "$lock.tmp" "$lock"
fi
# Each module follows the lock beside it: the harness tracks the current protocol, while retail stays on
# the gophertunnel commit its pinned captures were recorded with.
for dir in "$root/differential" "$root/differential/retail"; do
  target="$(jq -er .module "$dir/gophertunnel.json")@$(jq -er .version "$dir/gophertunnel.json")"
  go -C "$dir" mod edit -replace "github.com/sandertv/gophertunnel=$target"
  go -C "$dir" mod tidy
done

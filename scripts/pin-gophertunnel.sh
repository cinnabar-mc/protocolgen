#!/usr/bin/env bash
# Points every differential module at the gophertunnel fork version in
# differential/gophertunnel.json, or at VERSION after writing it there.
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)
lock=$root/differential/gophertunnel.json
if [[ $# -eq 1 ]]; then
  jq --arg version "$1" '.version = $version' "$lock" > "$lock.tmp"
  mv "$lock.tmp" "$lock"
fi
target="$(jq -er .module "$lock")@$(jq -er .version "$lock")"
for dir in "$root/differential" "$root/differential/retail"; do
  go -C "$dir" mod edit -replace "github.com/sandertv/gophertunnel=$target"
  go -C "$dir" mod tidy
done

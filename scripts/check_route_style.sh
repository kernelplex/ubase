#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"

cd "$ROOT_DIR"

# Find single-line route literals which hurt readability, e.g.:
#   return contracts.Route{Path: "...", RequiresPermission: ..., Func: handler}

violations=$(rg -n "return\s+(contracts|ubwww)\\.Route\{[^\n}]*\}" lib/ubadminpanel || true)

if [[ -n "$violations" ]]; then
  echo "Route style check failed. Please split Route literals across lines:" >&2
  echo "$violations" >&2
  exit 1
fi

echo "Route style check passed."


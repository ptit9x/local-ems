#!/usr/bin/env bash
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
arch_dir="$repo_root/docs/architecture"
mmdc="$repo_root/ems-cloud/node_modules/.bin/mmdc"
if [[ ! -x "$mmdc" ]]; then
  echo "Mermaid CLI is missing. Run: cd ems-cloud && npm install" >&2
  exit 1
fi
"$mmdc" -c "$arch_dir/mermaid-config.json" -i "$arch_dir/mermaid.mmd" -o "$arch_dir/architecture.png" -w 3600 -H 2200 -b '#08111f'
echo "Rendered technical architecture.png"
echo "Preserved icon-rich architecture_simple.png (see architecture_simple.prompt.md)"


#!/bin/bash
set -euo pipefail

# Clone all target repositories for EIS analysis
# Usage: ./clone-repos.sh [data-dir]

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DATA_DIR="${1:-$SCRIPT_DIR/../data/repos}"

mkdir -p "$DATA_DIR"

# All 25 repos: org/repo
REPOS=(
  # Tier 1: Famous Projects
  "facebook/react"
  "kubernetes/kubernetes"
  "hashicorp/terraform"
  "redis/redis"
  "rust-lang/rust"
  # Tier 2: Infrastructure
  "prometheus/prometheus"
  "grafana/grafana"
  "grafana/loki"
  "argoproj/argo-cd"
  "envoyproxy/envoy"
  # Tier 2: Backend Frameworks
  "fastapi/fastapi"
  "nestjs/nest"
  "spring-projects/spring-boot"
  "expressjs/express"
  "phoenixframework/phoenix"
  # Tier 2: Developer Tools
  "evanw/esbuild"
  "swc-project/swc"
  "vitejs/vite"
  "prettier/prettier"
  "eslint/eslint"
  # Tier 2: Data / Systems
  "duckdb/duckdb"
  "ClickHouse/ClickHouse"
  "apache/arrow"
  "pola-rs/polars"
  "apache/superset"
)

total=${#REPOS[@]}
current=0

for repo in "${REPOS[@]}"; do
  current=$((current + 1))
  name=$(basename "$repo")
  target="$DATA_DIR/$name"

  if [ -d "$target/.git" ]; then
    echo "[$current/$total] SKIP $repo (already cloned)"
    continue
  fi

  echo "[$current/$total] CLONE $repo -> $target"
  # Full clone needed for git blame analysis
  git clone --quiet "https://github.com/$repo.git" "$target"
  echo "  -> done ($(du -sh "$target" | cut -f1))"
done

echo ""
echo "=== Clone Summary ==="
echo "Total repos: $total"
echo "Data directory: $DATA_DIR"
du -sh "$DATA_DIR" 2>/dev/null || true

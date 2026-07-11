#!/bin/bash
# Generate the observation substrate for the collapse backtest: for each OSS repo,
# a full-history monthly EIS timeline (with per-(module,author) survival) + a
# cumulative analyze (module churn / cochange / ownership). Writes data/*.json.
#
# Requires an `eis` that emits `module_survival_by_author` in `timeline --format
# json` (merged to main in #366). Point $EIS at it, or put `eis` on PATH.
#   go build -o eis ./cmd/eis   &&   EIS=$PWD/eis research/collapse-backtest/run.sh
#
# Clones one repo at a time and keeps it (curated configs live in
# ../oss-gravity-map/configs). ~20 min, ~1 GB. Re-run resumes (skips done repos).
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
EIS="${EIS:-eis}"
CFGDIR="$HERE/../oss-gravity-map/configs"
mkdir -p "$HERE/data"
cd "$HERE"

# shortname  full_name              curated_config
REPOS=(
  "express  expressjs/express  express.yaml"
  "redis    redis/redis        redis.yaml"
  "eslint   eslint/eslint      eslint.yaml"
  "prettier prettier/prettier  prettier.yaml"
  "fastapi  fastapi/fastapi    fastapi.yaml"
  "nest     nestjs/nest        nest.yaml"
  "vite     vitejs/vite        vite.yaml"
  "esbuild  evanw/esbuild      esbuild.yaml"
)

for entry in "${REPOS[@]}"; do
  read -r name full cfg <<< "$entry"
  echo "=== [$name] $(date +%H:%M:%S) ==="
  if [ -s "data/$name.timeline.json" ]; then echo "[$name] SKIP (done)"; continue; fi
  [ -d "$name" ] || git clone --quiet "https://github.com/$full.git" "$name" 2>&1 | tail -1
  C="$CFGDIR/$cfg"
  echo "[$name] analyze..."
  "$EIS" analyze  --format json --config "$C" "$name" > "data/$name.analyze.json"  2> "data/$name.analyze.err"
  echo "[$name] timeline (monthly, sequential)..."
  "$EIS" timeline --format json --span 1m --period-concurrency 1 --workers 8 \
         --config "$C" "$name" > "data/$name.timeline.json" 2> "data/$name.timeline.err"
  echo "[$name] DONE  timeline=$(ls -lh data/$name.timeline.json 2>/dev/null | awk '{print $5}')"
done
echo "=== ALL DONE $(date +%H:%M:%S) ==="

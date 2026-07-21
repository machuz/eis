#!/bin/bash
# Build the substrate for the abandonment backtest.
#
# Predictors come from `eis timeline` (module_survival_by_author, #366).
# The outcome comes from raw `git log` (activity.py). Two pipelines, on purpose:
# if the same machinery produced both, the correlation would be manufactured.
#
#   go build -o /tmp/eis ./cmd/eis
#   EIS=/tmp/eis research/abandonment-backtest/run.sh react express esbuild
#
# Reuses the clones and curated configs in ../oss-gravity-map (no re-cloning).
# Resumes: a repo with a non-empty timeline.json is skipped.
set -uo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
EIS="${EIS:-eis}"
GMAP="$HERE/../oss-gravity-map"
REPOS="${REPOS:-$GMAP/data/repos}"
CFGDIR="${CFGDIR:-$GMAP/configs}"
DATA="${EIS_BT_DATA:-$HERE/data}"
SPAN="${SPAN:-1m}"

mkdir -p "$DATA"

names=("$@")
if [ ${#names[@]} -eq 0 ]; then
  names=(react express esbuild eslint prettier vite nest redis fastapi prometheus)
fi

for name in "${names[@]}"; do
  repo="$REPOS/$name"
  cfg="$CFGDIR/$name.yaml"
  if [ ! -d "$repo" ]; then echo "[$name] SKIP (no clone at $repo)"; continue; fi
  if [ ! -f "$cfg" ]; then echo "[$name] SKIP (no config at $cfg)"; continue; fi
  if [ -s "$DATA/$name.timeline.json" ]; then echo "[$name] SKIP (done)"; continue; fi

  echo "=== [$name] $(date +%H:%M:%S) ==="
  "$EIS" analyze --format json --config "$cfg" "$repo" \
      > "$DATA/$name.analyze.json" 2> "$DATA/$name.analyze.err" || {
    echo "[$name] analyze FAILED — see $DATA/$name.analyze.err"; continue; }

  "$EIS" timeline --format json --span "$SPAN" --period-concurrency 4 --workers 8 \
      --config "$cfg" "$repo" \
      > "$DATA/$name.timeline.json" 2> "$DATA/$name.timeline.err" || {
    echo "[$name] timeline FAILED — see $DATA/$name.timeline.err"; continue; }

  echo "[$name] DONE timeline=$(ls -lh "$DATA/$name.timeline.json" | awk '{print $5}')"
done

echo "=== outcome substrate (raw git) ==="
EIS_BT_DATA="$DATA" REPOS="$REPOS" python3 "$HERE/activity.py"

echo
echo "now run:  EIS_BT_DATA=$DATA python3 $HERE/abandonment.py --pooled"

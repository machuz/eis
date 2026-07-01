#!/bin/bash
set -euo pipefail

# Observe the research target repos and POST one EIS research run to the
# OrbitLens research store, ONE repo at a time so disk stays bounded (the full
# 25-repo dataset is ~50GB — too big to hold at once on a CI runner).
#
# Per repo: clone (full history) -> eis analyze --format json -> ingest (--only)
# -> delete the clone. Every per-repo POST shares one --run-id and upserts
# idempotently, so the run accretes repo by repo.
#
# Env:
#   RESEARCH_INGEST_TOKEN  bearer for POST /oss/research/ingest (required unless DRY_RUN=1)
#   GITHUB_TOKEN           enables the Commits-API email->id fallback (recommended)
#   API_BASE               ingest base URL (default: https://api.stg.orbitlens.io)
#   RUN_ID                 run lineage id (default: utc date, ingest-YYYYMMDD)
#   ONLY                   comma-separated manifest dirs (default: all)
#   MAX_COMMIT_PAGES       Commits-API pages for the fallback (default: 10)
#   DRY_RUN                1 = assemble + print, never POST
#
# Requires: go (to build eis + the ingest tool), git, python3.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR/.."
REPO_ROOT="$(cd "$PROJECT_DIR/../.." && pwd)"

API_BASE="${API_BASE:-https://api.stg.orbitlens.io}"
RUN_ID="${RUN_ID:-ingest-$(date -u +%Y%m%d)}"
MAX_COMMIT_PAGES="${MAX_COMMIT_PAGES:-10}"
TIMELINE_SPAN="${TIMELINE_SPAN:-1y}"   # timeline window span (3m/6m/1y); 1y keeps row counts sane
ONLY="${ONLY:-}"
DRY_RUN="${DRY_RUN:-0}"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
DATA_REPOS="$WORK/repos"
DATA_RESULTS="$WORK/results"
mkdir -p "$DATA_REPOS" "$DATA_RESULTS"

echo "== building eis + ingest tool =="
EIS_BIN="$WORK/eis"
INGEST_BIN="$WORK/ingest"
( cd "$REPO_ROOT" && go build -o "$EIS_BIN" ./cmd/eis )
( cd "$REPO_ROOT" && go build -o "$INGEST_BIN" ./research/oss-gravity-map/cmd/ingest )

MANIFEST="$PROJECT_DIR/ingest-repos.yaml"
CONFIGS="$PROJECT_DIR/configs"

# Manifest dirs to process (respect ONLY). Read into an array without mapfile so
# this runs on macOS's bash 3.2 too, not just CI's bash 5.
ROWS_FILE="$WORK/rows.tsv"
python3 - "$MANIFEST" "$ONLY" > "$ROWS_FILE" <<'PY'
import sys, yaml
manifest, only = sys.argv[1], sys.argv[2]
want = {d.strip() for d in only.split(",") if d.strip()}
repos = yaml.safe_load(open(manifest))["repos"]
for r in repos:
    if not want or r["dir"] in want:
        print(f"{r['dir']}\t{r['full_name']}")
PY
DIRS=()
while IFS= read -r row; do
  [ -n "$row" ] && DIRS+=("$row")
done < "$ROWS_FILE"

echo "== run $RUN_ID -> $API_BASE  (${#DIRS[@]} repos, dry_run=$DRY_RUN) =="
ingested=0
for row in "${DIRS[@]}"; do
  dir="${row%%$'\t'*}"
  full="${row#*$'\t'}"
  echo ""
  echo "-- $full --"

  git clone --quiet "https://github.com/$full.git" "$DATA_REPOS/$dir"

  cfg_flag=""
  [ -f "$CONFIGS/$dir.yaml" ] && cfg_flag="--config $CONFIGS/$dir.yaml"
  # Cumulative (通期): the HEAD standing + lifetime_gravity.
  # shellcheck disable=SC2086
  "$EIS_BIN" analyze $cfg_flag --format json "$DATA_REPOS/$dir" \
    2>/dev/null | sed '/^Analyzing:/d; /^Loaded /d; /^SKIP:/d' > "$DATA_RESULTS/$dir.json"

  # Timeline (軌跡): per-period windows over the full history. Best-effort — a
  # failure just leaves the cumulative standing (the ingest tool treats a missing
  # timeline file as optional).
  # shellcheck disable=SC2086
  "$EIS_BIN" timeline $cfg_flag --format json --span "$TIMELINE_SPAN" --periods 0 "$DATA_REPOS/$dir" \
    2>/dev/null | sed '/^Analyzing:/d; /^Loaded /d; /^SKIP:/d' > "$DATA_RESULTS/$dir-timeline.json" || true

  ingest_flags=(--manifest "$MANIFEST" --results-dir "$DATA_RESULTS" \
    --timeline-dir "$DATA_RESULTS" \
    --repos-dir "$DATA_REPOS" --configs-dir "$CONFIGS" \
    --run-id "$RUN_ID" --api-base "$API_BASE" \
    --max-commit-pages "$MAX_COMMIT_PAGES" --only "$dir")
  [ "$DRY_RUN" = "1" ] && ingest_flags+=(--dry-run)

  if "$INGEST_BIN" "${ingest_flags[@]}"; then
    ingested=$((ingested + 1))
  else
    echo "WARN ingest failed for $full (continuing)"
  fi

  # Reclaim disk before the next (possibly huge) clone.
  rm -rf "$DATA_REPOS/$dir" "$DATA_RESULTS/$dir.json" "$DATA_RESULTS/$dir-timeline.json"
done

echo ""
echo "== done: $ingested/${#DIRS[@]} repos ingested into run $RUN_ID =="

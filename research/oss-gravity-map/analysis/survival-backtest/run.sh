#!/usr/bin/env bash
# Orchestrate the survival backtest for one repo: extract predictors @T, compute
# cohort-survival outcome @HEAD (module + author), leaving CSVs in OUT for
# backtest.py to pool. See README.md.
#
# Usage:
#   EIS=/path/to/eis CORPUS=/path/to/oss-gravity-map OUT=/tmp/bt \
#     run.sh <repo-name> [horizon-days] [anchor-date]
set -euo pipefail

REPO_NAME="${1:?repo name (e.g. react)}"
HORIZON="${2:-1095}"
ANCHOR="${3:-}"

EIS="${EIS:?set EIS to the eis binary}"
CORPUS="${CORPUS:?set CORPUS to the oss-gravity-map dir}"
OUT="${OUT:-/tmp/bt}"
HERE="$(cd "$(dirname "$0")" && pwd)"

REPO="$CORPUS/data/repos/$REPO_NAME"
CFG="$CORPUS/configs/$REPO_NAME.yaml"
[ -d "$REPO/.git" ] || { echo "no repo clone at $REPO" >&2; exit 1; }
CFG_ARG=(); [ -f "$CFG" ] && CFG_ARG=(--config "$CFG")
mkdir -p "$OUT"

echo "[$REPO_NAME] extracting predictors @T ..." >&2
ANCHOR_ARG=(); [ -n "$ANCHOR" ] && ANCHOR_ARG=(--anchor-date "$ANCHOR")
python3 "$HERE/extract_predictors.py" --eis "$EIS" --repo "$REPO" \
  ${CFG_ARG[@]+"${CFG_ARG[@]}"} \
  --horizon-days "$HORIZON" ${ANCHOR_ARG[@]+"${ANCHOR_ARG[@]}"} \
  --out-modules "$OUT/${REPO_NAME}_mods.csv" \
  --out-authors "$OUT/${REPO_NAME}_auths.csv" \
  --out-meta "$OUT/${REPO_NAME}_meta.json"

ANCHOR_END="$(python3 -c "import json;print(json.load(open('$OUT/${REPO_NAME}_meta.json'))['anchor_end'])")"
echo "[$REPO_NAME] anchor=$ANCHOR_END — cohort-survival outcome @HEAD ..." >&2
python3 "$HERE/outcome_cohort_survival.py" --repo "$REPO" --anchor-date "$ANCHOR_END" \
  --modules "$OUT/${REPO_NAME}_mods.csv" \
  --out-modules "$OUT/${REPO_NAME}_out_mods.csv" \
  --sample-files "${SAMPLE_FILES:-500}" \
  --with-authors --out-authors "$OUT/${REPO_NAME}_out_auths.csv"

echo "[$REPO_NAME] done." >&2

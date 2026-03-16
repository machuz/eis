#!/bin/bash
set -euo pipefail

# Run EIS analysis on all cloned repositories
# Uses per-repo config files from configs/ directory
# Usage: ./analyze-repos.sh [data-dir] [results-dir]
#
# Requires: eis CLI in PATH (brew install machuz/tap/eis)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR/.."
DATA_DIR="${1:-$PROJECT_DIR/data/repos}"
RESULTS_DIR="${2:-$PROJECT_DIR/data/results}"
CONFIG_DIR="$PROJECT_DIR/configs"

mkdir -p "$RESULTS_DIR"

# Check eis is available
if ! command -v eis &> /dev/null; then
  echo "ERROR: eis CLI not found. Install with: brew tap machuz/tap && brew install eis"
  exit 1
fi

EIS_VERSION=$(eis version 2>&1 | head -1 || echo "unknown")
echo "EIS version: $EIS_VERSION"
echo "Data dir: $DATA_DIR"
echo "Results dir: $RESULTS_DIR"
echo ""

# Find all cloned repos
repos=()
for dir in "$DATA_DIR"/*/; do
  if [ -d "$dir/.git" ]; then
    repos+=("$dir")
  fi
done

total=${#repos[@]}
current=0
succeeded=0
failed=0

echo "Found $total repositories to analyze"
echo "======================================"
echo ""

for repo_path in "${repos[@]}"; do
  current=$((current + 1))
  name=$(basename "$repo_path")
  result_file="$RESULTS_DIR/${name}.json"
  log_file="$RESULTS_DIR/${name}.log"

  if [ -f "$result_file" ] && [ -s "$result_file" ]; then
    echo "[$current/$total] SKIP $name (results exist)"
    succeeded=$((succeeded + 1))
    continue
  fi

  echo "[$current/$total] ANALYZE $name ..."
  start_time=$(date +%s)

  # Use per-repo config if available
  config_file="$CONFIG_DIR/${name}.yaml"
  config_flag=""
  if [ -f "$config_file" ]; then
    config_flag="--config $config_file"
    echo "  -> using config: ${name}.yaml"
  fi

  # Run EIS, strip progress lines from JSON output
  if eis analyze $config_flag --format json "$repo_path" 2>"$log_file" | sed '/^Analyzing:/d; /^Loaded /d; /^SKIP:/d' > "$result_file"; then
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))
    # Count members in result
    members=$(python3 -c "import json,sys; d=json.load(open('$result_file')); print(sum(len(r.get('members',[])) for r in d.get('domains',[])))" 2>/dev/null || echo "?")
    echo "  -> OK (${elapsed}s, ${members} members)"
    succeeded=$((succeeded + 1))
  else
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))
    echo "  -> FAILED (${elapsed}s, see $log_file)"
    rm -f "$result_file"  # Remove empty/partial result
    failed=$((failed + 1))
  fi
done

echo ""
echo "=== Analysis Summary ==="
echo "Total: $total"
echo "Succeeded: $succeeded"
echo "Failed: $failed"
echo "Results: $RESULTS_DIR"

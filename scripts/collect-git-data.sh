#!/bin/bash
#
# Engineering Impact Score — Git Data Collector
#
# Collects raw git data from repositories for analysis.
# Output goes to ./output/<timestamp>/ for Claude to analyze.
#
# Usage:
#   ./scripts/collect-git-data.sh /path/to/repo1 /path/to/repo2 ...
#
# Or with a file listing repos (one per line):
#   ./scripts/collect-git-data.sh $(cat repos.txt)
#

set -euo pipefail

if [ $# -eq 0 ]; then
    echo "Usage: $0 <repo_path> [repo_path ...]"
    echo ""
    echo "Example:"
    echo "  $0 /path/to/backend/api /path/to/frontend/web-app"
    echo ""
    echo "Or list repos in a file:"
    echo "  $0 \$(cat repos.txt)"
    exit 1
fi

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
OUTPUT_DIR="./output/${TIMESTAMP}"
mkdir -p "${OUTPUT_DIR}"

echo "=== Engineering Impact Score — Data Collection ==="
echo "Output: ${OUTPUT_DIR}"
echo "Repos: $#"
echo ""

for REPO_PATH in "$@"; do
    if [ ! -d "${REPO_PATH}/.git" ]; then
        echo "SKIP: ${REPO_PATH} (not a git repo)"
        continue
    fi

    REPO_NAME=$(basename "${REPO_PATH}")
    REPO_OUT="${OUTPUT_DIR}/${REPO_NAME}"
    mkdir -p "${REPO_OUT}"

    echo "--- Collecting: ${REPO_NAME} ---"

    cd "${REPO_PATH}"

    # Ensure we're on main/master
    DEFAULT_BRANCH=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@' || echo "main")
    git checkout "${DEFAULT_BRANCH}" --quiet 2>/dev/null || true
    git pull --quiet 2>/dev/null || true

    # 1. Git log with numstat (for Production + Quality)
    echo "  [1/5] git log..."
    git log --all --no-merges --format="COMMIT:%H|%an|%ae|%ai|%s" --numstat \
        > "${REPO_OUT}/git-log-numstat.txt" 2>/dev/null

    # 2. Author summary (quick overview)
    echo "  [2/5] author summary..."
    git log --all --no-merges --format="%an" | sort | uniq -c | sort -rn \
        > "${REPO_OUT}/author-summary.txt" 2>/dev/null

    # 3. Blame data for .go files
    echo "  [3/5] blame (*.go)..."
    git ls-files -- '*.go' 2>/dev/null | head -500 | while read f; do
        git blame --line-porcelain -w "$f" 2>/dev/null
    done | grep -E "^(author |committer-time |filename )" \
        > "${REPO_OUT}/blame-go.txt" 2>/dev/null || true

    # 4. Blame data for .ts/.tsx files
    echo "  [4/5] blame (*.ts/*.tsx)..."
    git ls-files -- '*.ts' '*.tsx' 2>/dev/null | head -500 | while read f; do
        git blame --line-porcelain -w "$f" 2>/dev/null
    done | grep -E "^(author |committer-time |filename )" \
        > "${REPO_OUT}/blame-ts.txt" 2>/dev/null || true

    # 5. Fix commits list (for Debt Cleanup analysis)
    echo "  [5/5] fix commits..."
    git log --all --no-merges --format="%H|%an|%s" \
        | grep -iE "^[^|]*\|[^|]*\|(fix|revert|hotfix)" \
        > "${REPO_OUT}/fix-commits.txt" 2>/dev/null || true
    git log --all --no-merges --format="%H|%an|%s" \
        | grep "修正" \
        >> "${REPO_OUT}/fix-commits.txt" 2>/dev/null || true

    echo "  Done. $(wc -l < "${REPO_OUT}/git-log-numstat.txt" | tr -d ' ') lines of log data"

    cd - > /dev/null
done

# Write metadata
cat > "${OUTPUT_DIR}/metadata.txt" << METADATA
Collection timestamp: ${TIMESTAMP}
Repos analyzed: $@
Date: $(date -u +%Y-%m-%dT%H:%M:%SZ)
METADATA

echo ""
echo "=== Collection complete ==="
echo "Output directory: ${OUTPUT_DIR}"
echo ""
echo "Next step: Feed this data to Claude for analysis."
echo "  claude 'Analyze the git data in ${OUTPUT_DIR} using the Engineering Impact Score methodology. See PROMPT.md for instructions.'"

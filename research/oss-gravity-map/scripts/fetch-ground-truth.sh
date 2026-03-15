#!/bin/bash
set -euo pipefail

# Fetch ground truth data from GitHub API for each repository
# Uses: gh CLI (authenticated)
#
# Ground truth sources:
#   1. Top contributors by commits
#   2. Members with write/admin access (if available)
#   3. Release authors
#   4. CODEOWNERS file
#   5. Repository metadata (stars, age, contributor count)
#
# Usage: ./fetch-ground-truth.sh [output-dir]

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="${1:-$SCRIPT_DIR/../data/ground-truth}"

mkdir -p "$OUTPUT_DIR"

# Check gh is available and authenticated
if ! command -v gh &> /dev/null; then
  echo "ERROR: gh CLI not found. Install with: brew install gh"
  exit 1
fi

if ! gh auth status &> /dev/null; then
  echo "ERROR: gh not authenticated. Run: gh auth login"
  exit 1
fi

REPOS=(
  "facebook/react"
  "kubernetes/kubernetes"
  "hashicorp/terraform"
  "redis/redis"
  "rust-lang/rust"
  "prometheus/prometheus"
  "grafana/grafana"
  "grafana/loki"
  "argoproj/argo-cd"
  "envoyproxy/envoy"
  "fastapi/fastapi"
  "nestjs/nest"
  "spring-projects/spring-boot"
  "expressjs/express"
  "phoenixframework/phoenix"
  "evanw/esbuild"
  "swc-project/swc"
  "vitejs/vite"
  "prettier/prettier"
  "eslint/eslint"
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
  output_file="$OUTPUT_DIR/${name}.json"

  if [ -f "$output_file" ]; then
    echo "[$current/$total] SKIP $name (already fetched)"
    continue
  fi

  echo "[$current/$total] FETCH $repo ..."

  # 1. Repository metadata
  echo "  -> repo metadata"
  repo_meta=$(gh api "repos/$repo" --jq '{
    full_name: .full_name,
    stars: .stargazers_count,
    forks: .forks_count,
    language: .language,
    created_at: .created_at,
    pushed_at: .pushed_at,
    default_branch: .default_branch,
    description: .description
  }' 2>/dev/null || echo '{}')

  # 2. Top 100 contributors (by commits)
  echo "  -> top contributors"
  contributors=$(gh api "repos/$repo/contributors?per_page=100" --jq '[.[] | {
    login: .login,
    contributions: .contributions,
    type: .type
  }]' 2>/dev/null || echo '[]')

  # 3. Recent releases (last 20) — who publishes releases
  echo "  -> releases"
  releases=$(gh api "repos/$repo/releases?per_page=20" --jq '[.[] | {
    tag: .tag_name,
    author: .author.login,
    published_at: .published_at
  }]' 2>/dev/null || echo '[]')

  # 4. CODEOWNERS (if exists)
  echo "  -> CODEOWNERS"
  default_branch=$(echo "$repo_meta" | python3 -c "import json,sys; print(json.load(sys.stdin).get('default_branch','main'))" 2>/dev/null || echo "main")
  codeowners=""
  for path in "CODEOWNERS" ".github/CODEOWNERS" "docs/CODEOWNERS"; do
    content=$(gh api "repos/$repo/contents/$path?ref=$default_branch" --jq '.content' 2>/dev/null || echo "")
    if [ -n "$content" ] && [ "$content" != "null" ]; then
      codeowners=$(echo "$content" | base64 -d 2>/dev/null || echo "")
      break
    fi
  done

  # 5. Assemble ground truth JSON
  python3 -c "
import json, sys

repo_meta = json.loads('''$repo_meta''') if '''$repo_meta'''.strip() else {}
contributors = json.loads('''$contributors''') if '''$contributors'''.strip() else []
releases = json.loads('''$releases''') if '''$releases'''.strip() else []

# Extract unique release authors
release_authors = list(set(r['author'] for r in releases if r.get('author')))

# Top 10 contributors as architect candidates
architect_candidates = [c['login'] for c in contributors[:10] if c.get('type') == 'User']

output = {
    'repo': '$repo',
    'name': '$name',
    'metadata': repo_meta,
    'top_contributors': contributors,
    'release_authors': release_authors,
    'architect_candidates': architect_candidates,
    'codeowners_present': bool('''$codeowners'''.strip()),
    'fetched_at': '$(date -u +%Y-%m-%dT%H:%M:%SZ)'
}

with open('$output_file', 'w') as f:
    json.dump(output, f, indent=2)

print(f'  -> saved ({len(contributors)} contributors, {len(release_authors)} release authors)')
" 2>/dev/null || echo "  -> FAILED (python error)"

  # Rate limit awareness
  sleep 1
done

echo ""
echo "=== Ground Truth Summary ==="
echo "Total repos: $total"
echo "Output: $OUTPUT_DIR"
ls -la "$OUTPUT_DIR"/*.json 2>/dev/null | wc -l | xargs -I{} echo "Files: {}"

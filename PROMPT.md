# Engineering Impact Score — Claude Analysis Prompt

Copy and paste this prompt into Claude Code (or any Claude interface) to analyze your team.

---

## Quick Start (Claude Code)

```bash
# 1. Clone and configure
git clone https://github.com/machuz/engineering-impact-score.git
cd engineering-impact-score
cp config.example.yaml config.yaml
# Edit config.yaml with your repo paths, aliases, etc.

# 2. Run Claude Code with the analysis prompt
claude "Follow the instructions in PROMPT.md to calculate Engineering Impact Scores for my team. Use config.yaml for configuration."
```

## Alternative: Manual Data Collection + Claude

```bash
# 1. Collect git data
./scripts/collect-git-data.sh /path/to/repo1 /path/to/repo2

# 2. Feed to Claude
claude "Analyze the git data in ./output/<timestamp>/ using the Engineering Impact Score methodology described in PROMPT.md"
```

---

## Full Analysis Prompt

Use the following prompt with Claude Code. It will read your config, run git commands, and output scores.

---

### PROMPT START

You are calculating Engineering Impact Scores for a software team. This is a 7-axis model that quantifies engineer impact using git history data.

**Read `config.yaml` first** to understand the repo paths, domains, aliases, and settings.

Then execute the following steps for each domain (BE, FE, Infra):

#### Step 1: Production (lines changed)

For each repo in the domain:
```bash
cd <repo_path>
git log --all --no-merges --format="COMMIT:%an||%s" --numstat
```

Parse the output. For each author, sum `insertions + deletions`, excluding files matching `exclude_file_patterns` from config.

#### Step 2: Quality (fix ratio)

From the same git log data, for each author:
- Count total commits
- Count fix commits: subject matches `^(fix|revert|hotfix)` (case-insensitive) OR contains `修正`
- `quality = 100 - (fix_commits / total_commits × 100)`

#### Step 3: Survival (recency-weighted blame)

For each repo in the domain:
```bash
git ls-files -- <blame_extensions from config> | while read f; do
  git blame --line-porcelain -w "$f" 2>/dev/null
done
```

Parse `author` and `committer-time` fields. For each blame line:
```
weight = exp(-days_since_commit / tau)   # tau from config, default 180
```
Sum weighted lines per author.

**Note:** For large repos, sample up to 500 files to keep runtime reasonable.

#### Step 4: Design (architecture commits)

For each repo in the domain:
```bash
git log --all --no-merges --format="%an" -- <architecture_patterns from config>
```
Count commits per author across all architecture patterns.

#### Step 5: Breadth (repo count)

For each repo in `all_repos` from config:
```bash
git log --all --no-merges --format="%an" | sort -u
```
Count how many repos each author has commits in.

#### Step 6: Debt Cleanup

For each repo in the domain, find fix commits and trace original authors:

1. List fix commits:
```bash
git log --all --no-merges --format="%H|%an|%s" | grep -iE "^[^|]*\|[^|]*\|(fix|revert|hotfix)"
```

2. For each fix commit (sample up to 50 per repo):
```bash
# Get changed files
git diff-tree --no-commit-id -r <hash> --name-only

# For each file, blame at parent commit
git blame <hash>^ -- <file> --line-porcelain 2>/dev/null | grep "^author "
```

3. Track:
   - `debt_generated[original_author] += 1` when someone else fixes their code
   - `debt_cleaned[fixer] += 1` when they fix someone else's code
   - `debt_ratio = debt_cleaned / max(debt_generated, 1)`

4. If `debt_generated + debt_cleaned < debt_threshold` (from config): use neutral score 50.

#### Step 7: Indispensability (Bus Factor)

For each module directory in each repo:
```bash
git ls-files '<module>/**/*.<ext>' | while read f; do
  git blame --line-porcelain -w "$f" 2>/dev/null
done | grep "^author " | sed 's/^author //' | sort | uniq -c | sort -rn
```

For each module:
- If top author owns >= 80% of lines: CRITICAL
- If top author owns >= 60% of lines: HIGH
- `indispensability = critical_count × 1.0 + high_count × 0.5`

#### Step 8: Normalize and Score

Apply aliases from config. Exclude authors in `exclude_authors`.

For each metric, normalize within the domain:
```
norm(value, max_in_domain) = min(value / max_in_domain × 100, 100)
```

Calculate total score using weights from config:
```
total = production × w_prod + quality × w_qual + survival × w_surv
      + design × w_design + breadth × w_breadth
      + debt_cleanup × w_debt + indispensability × w_indisp
```

#### Step 9: Output

Produce:

1. **Rankings table** per domain:
```
| # | Member | Prod | Qual | Surv | Design | Breadth | Debt | Indisp | Total | Archetype |
```

2. **Archetype classification** for each member:
   - **Architect**: Prod↑ Surv↑ Design↑ Debt↑
   - **Solid Cleaner**: Prod→ Qual↑ Surv↑ Debt↑
   - **Mass Producer**: Prod↑ Qual↓ Surv↓ Debt↓
   - **Political**: Breadth↑ Prod↓ Surv↓ Design↓
   - **Specialist**: narrow but deep
   - **Growing**: low volume, high quality

3. **Bus Factor risk map**: modules with CRITICAL/HIGH concentration

4. **Key insights**: notable patterns, risks, recommendations

### PROMPT END

---

## Tips

- **Large repos**: Blame analysis is the slowest part. The prompt instructs sampling up to 500 files per repo. For very large repos, you may want to reduce this.
- **Token usage**: Expect 500K–1.5M tokens for a 10-repo, 10-person team. Use a flat-rate plan (Claude Max) if available.
- **Quarterly tracking**: Run every 3 months and compare scores. Rising Survival = growing design skills. Rising Debt Cleanup = increasing team contribution.
- **Privacy**: Scores contain real names from git history. Handle results with appropriate confidentiality.

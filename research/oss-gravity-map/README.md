# OSS Gravity Map

**Mapping the Software Universe — Empirical Validation of EIS**

> *Who actually shaped the structure of open-source software?*

This research project applies [EIS (Engineering Impact Score)](https://github.com/machuz/engineering-impact-score) to 25 major open-source projects, validating the **Software Cosmology** model against real-world ground truth.

## What This Does

Conventional OSS rankings measure **activity** (stars, commits, contributors). This project measures **gravity** — the structural influence that shapes systems over time.

```
Activity Ranking          Gravity Ranking
─────────────────         ─────────────────
Who commits the most  →   Who shapes the architecture
Who has the most PRs  →   Whose code survives under pressure
Who appears in CODEOWNERS → Who generates structural pull
```

## Dataset

**25 projects, ~50,000 engineers, 8 languages/ecosystems**

| Tier | Projects | Purpose |
|------|----------|---------|
| Tier 1 (5) | React, Kubernetes, Terraform, Redis, Rust | Architect detection — EIS gravity ≈ known architects |
| Tier 2 (20) | Prometheus, Grafana, FastAPI, NestJS, esbuild, DuckDB, ... | Hidden architect discovery + entropy fighters |

## Hypotheses

1. **Architect Detection**: Top EIS gravity engineers overlap with known architects (target recall: 60-80%)
2. **Hidden Architects**: EIS reveals high-gravity engineers not listed as maintainers — software dark matter
3. **Entropy Fighters**: Engineers with high Debt Cleanup + high Survival are actively fighting system entropy
4. **Collapse Risk**: Gravity concentration predicts bus factor risk

## Metrics

| Metric | Definition |
|--------|-----------|
| Top-k Overlap | EIS top-10 ∩ GitHub contributor top-10 |
| Architect Recall | Known architects found in EIS top-20 |
| Spearman ρ | Rank correlation: EIS gravity vs contributor ranking |
| Gravity Concentration | Top-3 gravity / total gravity (bus factor proxy) |

## Quick Start

```bash
# Prerequisites
brew tap machuz/tap && brew install eis
brew install gh && gh auth login

# Full pipeline (~2-3 hours for clone + analysis)
make all

# Or step by step:
make clone          # Clone 25 repos (~50GB)
make ground-truth   # Fetch GitHub maintainer data
make analyze        # Run EIS on all repos
make report         # Generate RESULTS.md
```

## Architecture Configs (PR Welcome)

Each repository has a dedicated `configs/<repo>.yaml` defining **architecture patterns** — the files EIS treats as structurally significant.

These configs are **open for community contribution**. If you know a project well and think the architecture patterns are wrong or incomplete, submit a PR.

```yaml
# configs/react.yaml (example)
architecture_patterns:
  - "packages/react/src/*.js"
  - "packages/react-reconciler/src/*.js"
  - "packages/scheduler/src/*.js"
  ...
```

**Why this matters:** Architecture patterns directly affect the Design score, which feeds into gravity. Getting these right is essential for fair analysis. Open-sourcing the configs ensures transparency and accountability.

See [`configs/`](configs/) for all 25 repository configs.

## Project Structure

```
research/oss-gravity-map/
├── dataset.yaml              # 25 repos + known architects
├── Makefile                  # Pipeline automation
├── configs/                  # Per-repo eis.yaml configs (PR welcome!)
│   ├── react.yaml
│   ├── kubernetes.yaml
│   └── ...
├── scripts/
│   ├── clone-repos.sh        # Clone all target repos
│   ├── analyze-repos.sh      # Run EIS analysis (uses configs/)
│   ├── fetch-ground-truth.sh # GitHub API ground truth
│   ├── generate-configs.py   # Config generator (initial bootstrap)
│   └── analyze-results.py    # Statistical analysis
├── data/
│   ├── repos/                # Cloned repositories (gitignored)
│   ├── results/              # EIS JSON output per repo
│   └── ground-truth/         # GitHub API data per repo
└── analysis/
    ├── RESULTS.md            # Generated report
    └── results.json          # Structured output
```

## Expected Output

### Top OSS Gravity Engineers
```
Rank  Engineer              Project       Gravity  Role
────  ────────────────────  ──────────    ───────  ────────
1     ???                   Kubernetes    ???      Architect
2     ???                   React         ???      Architect
3     ???                   Rust          ???      Architect
...
```

### The Most Interesting Part

When an **unknown engineer** appears in the Top 10 gravity ranking — that is proof that EIS is truly observing the structure of the software universe.

```
Git remembers the past.
AI imagines the future.

Between them, engineers shape gravity.

This map reveals who bends that universe.
```

## Paper

Results from this validation will be incorporated into:

> **Observing the Software Universe: Detecting Architectural Influence from Git History**

---

Part of [Engineering Impact Score](https://github.com/machuz/engineering-impact-score) — the Git Telescope.

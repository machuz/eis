# Engineering Impact Score

**Quantify engineer combat power using nothing but git history.**

A 7-axis scoring model that measures an engineer's real impact on a codebase. No surveys, no subjective reviews — just `git log` and `git blame`.

> On my team (14 repos, 10+ engineers), this matched gut feeling with eerie accuracy.

## The 7 Axes

| Axis | Weight | What it measures |
|---|---|---|
| **Production** | 15% | Lines changed (excluding auto-generated files) |
| **First-pass Quality** | 10% | Low fix/revert commit ratio |
| **Code Survival** | **25%** | Recency-weighted blame survival (tau=180 exponential decay) |
| **Design** | 20% | Commits to architecture files |
| **Breadth** | 10% | Number of repositories contributed to |
| **Debt Cleanup** | 15% | Ratio of others' debt cleaned vs. own debt generated |
| **Indispensability** | 5% | Modules where you own 80%+ of blame lines (Bus Factor) |

**Code Survival is the core thesis** — exponential time decay ensures "are you *still* writing durable designs?" matters most.

## Score Guide

| Score | Assessment | Approx. Hourly Rate (JPY) |
|---|---|---|
| 80–100 | Irreplaceable core member | ¥12,000–20,000 |
| 60–79 | Near-core | ¥9,000–15,000 |
| **40–59** | **Senior equivalent (40+ is genuinely strong)** | **¥7,000–11,000** |
| 30–39 | Mid-level | ¥6,000–9,000 |
| 20–29 | Junior–Mid | ¥5,000–8,000 |
| –19 | Junior | ¥3,500–6,000 |

**40 = Senior.** This metric is deliberately harsh. Scoring 40+ across 7 relative axes requires serious, well-rounded ability.

## Engineer Archetypes

The 7-axis distribution reveals archetypes:

| Type | Prod | Qual | Surv | Design | Breadth | Debt | Indisp | Risk |
|---|---|---|---|---|---|---|---|---|
| **Architect** | ◎ | △–○ | ◎ | ◎ | ○ | ◎ | ◎ | — |
| **Mass Producer** | ◎ | ✕ | ✕ | △ | △ | ✕ | △ | **High** |
| **Solid Cleaner** | ○ | ◎ | ◎ | ○ | ○ | ◎ | △ | — |
| **Political** | ✕ | △ | ✕ | ✕ | ◎ | △ | ✕ | **High** |
| **Specialist** | ◎ | ◎ | ◎ | ○ | ✕ | ○ | ◎ | △ Silo |
| **Growing** | △ | ◎ | ○ | ✕ | △ | ○ | ✕ | — |

**Mass Producer and Political types look productive on individual metrics** but score low overall. Only multi-axis evaluation exposes them.

## Key Formulas

### Recency-Weighted Survival (Core)

```python
import math
from collections import defaultdict

tau = 180  # days — weight ≈ 0.37 at 6 months

weighted_survival = defaultdict(float)
for line in blame_lines:
    days_alive = (now - line.committer_time).days
    weight = math.exp(-days_alive / tau)
    weighted_survival[line.author] += weight
```

### Debt Cleanup Ratio

```python
for fix_commit in fix_commits:
    fixer = fix_commit.author
    for changed_line in fix_commit.changed_lines:
        original_author = git_blame(file, at=parent_commit)
        if original_author != fixer:
            debt_generated[original_author] += 1
            debt_cleaned[fixer] += 1

debt_ratio = debt_cleaned / max(debt_generated, 1)
# > 1 = Cleaner   |   < 1 = Debt creator
```

### Indispensability (Bus Factor)

```python
for module in all_modules:
    top_share = max(blame_distribution[module].values()) / total
    if top_share >= 0.8:
        critical_modules[top_author].append(module)   # CRITICAL
    elif top_share >= 0.6:
        high_risk_modules[top_author].append(module)   # HIGH

indispensability = critical_count * 1.0 + high_count * 0.5
```

### Normalization & Total Score

```python
def norm(value, max_value):
    return min(value / max_value * 100, 100)

# Normalize within domain (BE/FE/Infra separately)
total = (
    norm_production * 0.15
    + norm_quality * 0.10
    + norm_survival * 0.25
    + norm_design * 0.20
    + norm_breadth * 0.10
    + norm_debt_cleanup * 0.15
    + norm_indispensability * 0.05
)
```

## Design Principles

- **BE / FE / Infra are scored separately** — mixing them contaminates rankings
- **Relative scoring within domain** — the top person in each axis gets 100
- **Debt threshold** — members with fewer than 10 debt events get a neutral score (50) to avoid extreme ratios
- **Accuracy scales with codebase design quality** — well-structured codebases (Clean Architecture, DDD) yield more meaningful scores. If the score doesn't match gut feeling, it may signal poor codebase structure rather than a metric problem

## Quick Start

### Option 1: Claude Code (Recommended)

```bash
git clone https://github.com/machuz/engineering-impact-score.git
cd engineering-impact-score

# Configure your repos and team
cp config.example.yaml config.yaml
# Edit config.yaml with your repo paths, author aliases, etc.

# Run the analysis
claude "Follow the instructions in PROMPT.md to calculate Engineering Impact Scores for my team. Use config.yaml for configuration."
```

Claude will read your config, run git commands on your repos, and output full rankings with archetypes and Bus Factor analysis.

### Option 2: Collect Data First, Then Analyze

```bash
# Collect raw git data from your repos
./scripts/collect-git-data.sh /path/to/repo1 /path/to/repo2 /path/to/repo3

# Feed collected data to Claude
claude "Analyze the git data in ./output/<timestamp>/ using the Engineering Impact Score methodology described in PROMPT.md"
```

### What You Get

- **Rankings table** per domain (BE / FE / Infra) with all 7 axis scores
- **Archetype classification** (Architect, Solid Cleaner, Mass Producer, Political, etc.)
- **Bus Factor risk map** showing modules with dangerous ownership concentration
- **Actionable insights** on team structure, risks, and recommendations

### Requirements

- Git repos cloned locally
- [Claude Code](https://claude.ai/code) (recommended) or any Claude interface
- ~30–60 min for 10 repos / 10 engineers
- ~500K–1.5M tokens (use flat-rate Claude Max plan for best experience)

## Configuration

See [`config.example.yaml`](config.example.yaml) for all options:

- **Domains**: group repos into BE / FE / Infra
- **Architecture patterns**: define which files count as "design files" for each domain
- **Aliases**: merge variant git author names into canonical names
- **Weights**: customize axis weights (default: Survival 25%, Design 20%, Production 15%, Debt 15%, Quality 10%, Breadth 10%, Indispensability 5%)
- **Survival tau**: decay half-life in days (default: 180)

## Blog Posts

- [Japanese (はてなブログ)](docs/blog-ja.md) — Full article with real-world measurement results
- [English](docs/blog-en.md) — Full article (English version)
- [dev.to](docs/blog-devto.md) — dev.to optimized version with images

## Roadmap

- [x] Scoring methodology and formulas
- [x] Claude Code analysis prompt
- [x] Data collection script
- [x] Configuration template
- [ ] Standalone CLI tool (no AI dependency)
- [ ] GitHub Action for automated quarterly tracking
- [ ] Dashboard visualization
- [ ] Multi-language commit message support for Quality detection

## Support

If this metric helped you see your team differently, consider supporting the project:

- [GitHub Sponsors](https://github.com/sponsors/machuz)
- PayPay: `w_machu7`

## License

MIT

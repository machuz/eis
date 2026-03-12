# Engineering Impact Score

[![dev.to](https://img.shields.io/badge/dev.to-Read%20Article-0A0A0A?logo=devdotto&logoColor=white)](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)
[![はてなブログ](https://img.shields.io/badge/はてなブログ-記事を読む-00A4DE)](https://ma2k8.hateblo.jp/entry/2026/03/11/153212)
[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github)](https://github.com/sponsors/machuz)

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

| Score | Assessment | Approx. Hourly (JPY) | Approx. Total Comp (USD) |
|---|---|---|---|
| 80–100 | Irreplaceable core member | ¥12,000–20,000 | $250K–400K+ |
| 60–79 | Near-core | ¥9,000–15,000 | $180K–300K |
| **40–59** | **Senior equivalent (40+ is genuinely strong)** | **¥7,000–11,000** | **$140K–220K** |
| 30–39 | Mid-level | ¥6,000–9,000 | $100K–160K |
| 20–29 | Junior–Mid | ¥5,000–8,000 | $80K–120K |
| –19 | Junior | ¥3,500–6,000 | $60K–90K |

USD figures are rough estimates and vary significantly by market (SF vs. Midwest, US vs. Europe, etc.).

**40 = Senior.** This metric is deliberately harsh. Scoring 40+ across 7 relative axes requires serious, well-rounded ability.

## Engineer Archetypes

The 7-axis distribution reveals archetypes:

| Type | Prod | Qual | Surv | Design | Breadth | Debt | Indisp | Risk |
|---|---|---|---|---|---|---|---|---|
| **Architect** | ◎ | △–○ | ◎ | ◎ | ○ | ◎ | ◎ | — |
| **Former Architect** | △ | △ | ✕ | ◎ | ○ | △ | ◎ | **⚠️ Handoff** |
| **Mass Producer** | ◎ | ✕ | ✕ | △ | △ | ✕ | △ | **High** |
| **Solid Cleaner** | ○ | ◎ | ◎ | ○ | ○ | ◎ | △ | — |
| **Drifter** | ✕ | △ | ✕ | ✕ | ◎ | △ | ✕ | **High** |
| **Specialist** | ◎ | ◎ | ◎ | ○ | ✕ | ○ | ◎ | △ Silo |
| **Growing** | △ | ◎ | ○ | ✕ | △ | ○ | ✕ | — |

**Mass Producer and Drifter types look productive on individual metrics** but score low overall. Only multi-axis evaluation exposes them.

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

## How This Differs from Existing Metrics

| Framework | What it measures | Signal source | Individual? | Key limitation |
|---|---|---|---|---|
| **DORA** | Deployment speed & stability | CI/CD pipeline | No (team) | Doesn't measure code quality or individual impact |
| **SPACE** | 5 holistic dimensions | Surveys + tools | Both | Survey-heavy, 3–6 months to implement |
| **McKinsey** | Org productivity | DORA + SPACE + custom | Mixed | [Widely criticized](https://newsletter.pragmaticengineer.com/p/measuring-developer-productivity) for output theater |
| **LOC / Commits** | Activity volume | Git | Yes | Trivially gameable, penalizes refactoring |
| **Code Churn** | % of recent code rewritten | Git | No (team) | Arbitrary time window, context-blind |
| **Bus Factor** | Knowledge concentration risk | Git blame | No (team) | Only identifies risk, not impact |
| **Git analytics tools** (Pluralsight Flow, LinearB, etc.) | Activity & cycle time | Git + integrations | Both | Still activity-focused — measures *when*, not *whether it lasted* |
| **Engineering Impact Score** | **Code that survives over time** | **Git log + blame** | **Yes** | Accuracy depends on codebase design quality |

The core gap this model fills: **existing frameworks measure activity or velocity, not whether individual contributions actually lasted.** DORA tells you how fast code reaches production. This model tells you whether it was worth deploying.

Time-decayed survival is also naturally resistant to gaming — you can't inflate your score with busy work, because only code that remains in the codebase months later counts.

## Quick Start

### Install via Homebrew

```bash
brew tap machuz/tap
brew install eis
```

### Run

```bash
# Analyze current repo
eis analyze .

# Analyze multiple repos
eis analyze /path/to/repo1 /path/to/repo2

# Auto-discover git repos under a directory
eis analyze --recursive /path/to/workspace

# Search deeper (default depth: 2)
eis analyze --recursive --depth 3 ~/projects
```

### Configuration (Optional)

Create `eis.yaml` in your working directory:

```yaml
aliases:
  "John Smith": "john"
  "J. Smith": "john"

exclude_authors:
  - "dependabot[bot]"

weights:
  production: 0.15
  quality: 0.10
  survival: 0.25
  design: 0.20
  breadth: 0.10
  debt_cleanup: 0.15
  indispensability: 0.05

tau: 180
sample_size: 500
```

```bash
eis analyze --config eis.yaml --recursive ~/projects
```

### What You Get

- **Rankings table** with all 7 axis scores and total
- **Archetype classification** (Architect, Solid Cleaner, Mass Producer, Drifter, etc.)
- **Bus Factor risk map** showing modules with dangerous ownership concentration
- Color-coded output for quick visual scanning

### Supported Languages

Works out of the box with: Go, TypeScript/JavaScript, Python, Rust, Java, Ruby, C/C++ (including firmware/embedded), Scala, Haskell, OCaml. Additional extensions can be added via `blame_extensions` in `eis.yaml`.

### Requirements

- Git repos cloned locally
- That's it. No AI tokens, no API keys, no cloud dependencies

### Alternative: Claude Code

For deeper qualitative analysis (actionable insights, team recommendations), you can also use Claude Code:

```bash
claude "Follow the instructions in PROMPT.md to calculate Engineering Impact Scores for my team. Use config.yaml for configuration."
```

## CLI Options

```
eis analyze [flags] [path...]

Flags:
  --config <path>     Config file (default: eis.yaml)
  --recursive         Recursively find git repos under given paths
  --depth <n>         Max directory depth for recursive search (default: 2)
  --tau <days>        Survival decay parameter (default: 180)
  --sample <n>        Max files to blame per repo (default: 500)
  --workers <n>       Concurrent blame workers (default: 4)
```

## Configuration

See [`config.example.yaml`](config.example.yaml) for all options:

- **Aliases**: merge variant git author names into canonical names
- **Exclude authors**: filter out bots and non-human contributors
- **Architecture patterns**: define which files count as "design files" for the Design axis. Defaults:
  - Backend: `*/repository/*interface*`, `*/domainservice/`, `*/router.go`, `*/middleware/`, `di/*.go`
  - Frontend: `*/core/`, `*/stores/`, `*/hooks/`, `*/types/`
  - Override in `eis.yaml` to match your project structure (e.g., `*/proto/`, `*/migrations/`, `Makefile`)
- **Blame extensions**: file extensions for blame analysis
- **Weights**: customize axis weights (default: Survival 25%, Design 20%, Production 15%, Debt 15%, Quality 10%, Breadth 10%, Indispensability 5%)
- **Survival tau**: decay half-life in days (default: 180)
- **Debt threshold**: minimum events for debt score (default: 10)

## Blog Posts

- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/11/153212) — 日本語版フル記事（実測結果付き）
- [English / dev.to](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c) — English version with images and real-world rankings

## Roadmap

- [x] Scoring methodology and formulas
- [x] Claude Code analysis prompt
- [x] Data collection script
- [x] Configuration template
- [x] Standalone CLI tool (`eis`) — zero AI dependency
- [x] Homebrew distribution (`brew install eis`)
- [x] Recursive repo discovery (`--recursive`)
- [x] Author alias mapping via config
- [x] Concurrent blame analysis (worker pool)
- [ ] GitHub Action for automated quarterly tracking
- [ ] HTML dashboard visualization
- [ ] Multi-language commit message support for Quality detection
- [ ] JSON / CSV output format

## Support

If this metric helped you see your team differently, consider supporting the project:

- [GitHub Sponsors](https://github.com/sponsors/machuz)
- PayPay: `w_machu7`

## License

MIT

<p align="center">
  <img src="docs/images/logo-full.svg?v=3" alt="EIS — the Git Telescope" width="420">
</p>

<p align="right">
  <a href="docs/whitepaper.md"><img src="https://img.shields.io/badge/Whitepaper-EN-504945" alt="Whitepaper EN"></a>
  <a href="docs/whitepaper-ja.md"><img src="https://img.shields.io/badge/Whitepaper-JA-504945" alt="Whitepaper JA"></a>
  <a href="https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c"><img src="https://img.shields.io/badge/dev.to-Read%20Article-0A0A0A?logo=devdotto&logoColor=white" alt="dev.to"></a>
  <a href="https://ma2k8.hateblo.jp/entry/2026/03/11/153212"><img src="https://img.shields.io/badge/はてなブログ-記事を読む-00A4DE" alt="はてなブログ"></a>
  <a href="https://github.com/sponsors/machuz"><img src="https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github" alt="Sponsor"></a>
</p>

**Engineering Impact Score** (EIS, pronounced *"ace"*)

Git records commits.
EIS reveals the structure of the code universe.

### What the telescope reveals

Point it at a codebase, and hidden patterns emerge:

- **Architects** — who actually shaped the system's structure
- **Anchors** — the quiet forces holding the system together
- **Dark matter** — invisible forces stabilizing the code universe
- **Collapse risk** — bus-factor concentrations before they become emergencies
- **Entropy** — where code is decaying and who is fighting it

All from `git log` and `git blame`. Nothing else.

![Terminal Output](docs/images/terminal-output.svg?v=0.11.0)

## Why this exists

Most engineering metrics measure activity.

PR counts. Commit counts. Lines of code.

But activity is not impact.

An engineer who writes 10,000 lines that get rewritten next month left no gravity. An engineer who writes 500 lines that become the foundation of the system shaped its universe.

**EIS measures the gravity engineers leave in the codebase** — structure that survives, design that endures, and debt that gets cleaned up.

---

## How it works

EIS reads `git log` and `git blame` to compute 7 axes:

| Signal | What it captures |
|---|---|
| **Production** | Sustained output velocity |
| **Quality** | First-pass correctness (low fix/revert ratio) |
| **Survival** | Code that endures — exponential time-decay weighted |
| **Design** | Commits to architecture-defining files |
| **Breadth** | Cross-repository structural presence |
| **Debt Cleanup** | Fixing others' bugs vs. generating new ones |
| **Indispensability** | Modules where one engineer owns 80%+ of blame |

From these signals, EIS derives **Roles** (Architect, Anchor, Producer...), **Styles** (Builder, Emergent, Rescue...), and **States** (Active, Growing, Fragile...) — a 3-axis topology for each engineer.

**Survival is the core thesis.** Exponential time-decay ensures that "are you *still* writing durable code?" matters most. This is naturally resistant to gaming — busy work doesn't survive.

---

## Quick Start

```bash
# Install
go install github.com/machuz/engineering-impact-score/cmd/eis@latest
# or
brew tap machuz/tap && brew install eis

# Try it on your repo
eis analyze .
```

That's it. No AI tokens, no API keys, no cloud dependencies.

```bash
# Auto-discover repos under a directory
eis analyze --recursive ~/workspace

# With config (aliases, domain mapping, weights)
eis analyze --config eis.yaml --recursive ~/projects

# Export as JSON or CSV
eis analyze --format json --recursive ~/workspace > result.json

# Team-level analysis (aggregates individual scores)
eis team --recursive ~/workspace

# Team analysis with JSON output (paste into AI for deeper insights)
eis team --format json --recursive ~/workspace
```

## How This Differs from Existing Metrics

| Framework | What it measures | Signal source | Individual? | Key limitation |
|---|---|---|---|---|
| **DORA** | Deployment speed & stability | CI/CD pipeline | No (team) | Doesn't measure code quality or individual impact |
| **SPACE** | 5 holistic dimensions | Surveys + tools | Both | Survey-heavy, 3-6 months to implement |
| **McKinsey** | Org productivity | DORA + SPACE + custom | Mixed | [Widely criticized](https://newsletter.pragmaticengineer.com/p/measuring-developer-productivity) for output theater |
| **LOC / Commits** | Activity volume | Git | Yes | Trivially gameable, penalizes refactoring |
| **Code Churn** | % of recent code rewritten | Git | No (team) | Arbitrary time window, context-blind |
| **Bus Factor** | Knowledge concentration risk | Git blame | No (team) | Only identifies risk, not impact |
| **Git analytics tools** (Pluralsight Flow, LinearB, etc.) | Activity & cycle time | Git + integrations | Both | Still activity-focused — measures *when*, not *whether it lasted* |
| **Engineering Impact Score** | **Code that survives over time** | **Git log + blame** | **Yes** | Accuracy depends on codebase design quality |

The core gap this model fills: **existing frameworks measure activity or velocity, not whether individual contributions actually lasted.** DORA tells you how fast code reaches production. This model tells you whether it was worth deploying.

Time-decayed survival is also naturally resistant to gaming — you can't inflate your score with busy work, because only code that remains in the codebase months later counts.

## The 7 Axes

| Axis | Weight | What it measures |
|---|---|---|
| **Production** | 15% | Changes per day (absolute: configurable `production_daily_ref`, default 1000) |
| **First-pass Quality** | 10% | Low fix/revert commit ratio |
| **Code Survival** | **25%** | Recency-weighted blame survival (tau=180 exponential decay) |
| **Design** | 20% | Commits to architecture files |
| **Breadth** | 10% | Number of repositories contributed to |
| **Debt Cleanup** | 15% | Ratio of others' debt cleaned vs. own debt generated |
| **Indispensability** | 5% | Modules where you own 80%+ of blame lines (Bus Factor) |

**Code Survival is the core thesis** — exponential time decay ensures "are you *still* writing durable designs?" matters most.

**Gravity** (v0.10.0) is shown alongside scores but excluded from the total. It measures structural influence (`Indisp×0.4 + Breadth×0.3 + Design×0.3`) and is color-coded by health: green = durable influence, yellow = moderate, red = fragile dependency.

## Score Guide

| Score | Assessment | Approx. Hourly (JPY) | Approx. Total Comp (USD) |
|---|---|---|---|
| 80-100 | Irreplaceable core member | ¥12,000-20,000 | $250K-400K+ |
| 60-79 | Near-core | ¥9,000-15,000 | $180K-300K |
| **40-59** | **Senior equivalent (40+ is genuinely strong)** | **¥7,000-11,000** | **$140K-220K** |
| 30-39 | Mid-level | ¥6,000-9,000 | $100K-160K |
| 20-29 | Junior-Mid | ¥5,000-8,000 | $80K-120K |
| -19 | Junior | ¥3,500-6,000 | $60K-90K |

USD figures are rough estimates and vary significantly by market (SF vs. Midwest, US vs. Europe, etc.).

**40 = Senior.** This metric is deliberately harsh. Scoring 40+ across 7 relative axes requires serious, well-rounded ability.

> **Important: EIS measures impact on *this codebase*, not absolute engineering ability.** A high score means "on this codebase, this person's contributions are surviving, shaping architecture, and cleaning up debt." It does not mean they are a better engineer than someone with a lower score. Conversely, high Survival might mean the code can't be refactored away due to poor design — not that the code is good. If scores don't match your gut feeling, that's a signal worth investigating: it may reveal codebase design issues rather than people issues.

![Score Guide](docs/images/score-guide.svg?v=0.11.0)

## Engineer Topology (3-Axis Classification)

Instead of a single archetype label, EIS v0.9+ classifies each engineer along 3 independent axes. This gives a richer, more composable picture than a flat type.

### Role — what they contribute

| Role | Key Signals | Description |
|---|---|---|
| **Architect** | Design↑ RobustSurv↑ Breadth○ | Shapes system design with durable code under change pressure |
| **Anchor** | Qual↑ notLow(Prod) | Reliable quality contributor, not yet shaping architecture |
| **Cleaner** | Qual↑ Surv↑ Debt↑ | High quality, durable code, actively cleans others' debt |
| **Producer** | notLow(Prod) | Meaningful production output |
| **Specialist** | Surv↑ Breadth↓ | Deep in narrow area, high survival but low breadth |
| **—** | | No dominant role signal |

### Style — how they contribute

| Style | Key Signals | Description |
|---|---|---|
| **Builder** | Prod↑ Design↑ Debt○ | Designs, builds heavily, AND cleans up — the full package |
| **Resilient** | Prod↑ Surv↓ RobustSurv○ | Iterates heavily but what survives under pressure is durable |
| **Rescue** | Prod↑ Surv↓ Debt↑ | Takes over and rewrites inherited legacy code |
| **Churn** | Prod↑ Qual↓ Surv↓ gap≥30 | Constant rework — most commits are fixes or reverts |
| **Mass** | Prod↑ Surv↓ | High output but code doesn't survive |
| **Balanced** | Total≥30 | Steady contributor, no dominant pattern |
| **Spread** | Breadth↑ Prod↓ Surv↓ Design↓ | Wide presence, shallow depth everywhere |
| **—** | | No dominant style signal |

### State — lifecycle phase

| State | Key Signals | Risk |
|---|---|---|
| **Active** | Recent commits | — |
| **Growing** | Prod↓ Qual↑ | — |
| **Former** | RawSurv↑ Surv↓ Design/Indisp↑ | **⚠️ Handoff** |
| **Silent** | Prod↓ Surv↓ Debt↓ (≥100 commits) | **High** |
| **Fragile** | Surv↑ Prod↓ Qual<70 | **⚠️ Hidden** |
| **—** | | No dominant state signal |

### Reading the output

Each engineer gets a 3-label profile. Examples:

| Profile | Interpretation |
|---|---|
| Architect / Builder / Active | Core contributor: designs, builds, cleans, currently active |
| Producer / Mass / — | High output but code doesn't last |
| — / Spread / Silent | Wide but shallow, not contributing meaningfully |
| Anchor / Balanced / Growing | Reliable quality, steady pace, improving |
| — / — / Fragile | Code survives only due to low change pressure |

**Churn, Mass, and Spread styles look productive on individual metrics** but score low overall. Only multi-axis evaluation exposes them.

![Archetypes Radar](docs/images/archetypes-radar.svg?v=0.11.0)

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
# Absolute axes (cross-org comparable):
#   Production: min(changes_per_day / production_daily_ref * 100, 100)
#   Quality: 100 - fix_ratio (already 0-100)
#   Debt: bounded 0-100 scale

# Relative axes (normalized within domain):
#   Survival, Design, Breadth, Indispensability

# Scored per domain (Backend/Frontend/Infra/Firmware + custom domains, separately)
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

- **Domains are scored separately** (Backend/Frontend/Infra/Firmware by default, plus custom domains) — mixing them contaminates rankings; auto-detected from file extensions or configured explicitly
- **Hybrid scoring** — Production, Quality, and Debt use absolute scales (cross-org comparable); Survival, Design, Breadth, and Indispensability use relative normalization within domain
- **Debt threshold** — members with fewer than 10 debt events get a neutral score (50) to avoid extreme ratios
- **Accuracy scales with codebase design quality** — well-structured codebases (Clean Architecture, DDD) yield more meaningful scores. If the score doesn't match gut feeling, it may signal poor codebase structure rather than a metric problem

## CLI Options

```
eis analyze [flags] [path...]    Individual rankings
eis team [flags] [path...]       Team-level analysis

Shared flags:
  --config <path>     Config file (default: eis.yaml)
  --recursive         Recursively find git repos under given paths
  --depth <n>         Max directory depth for recursive search (default: 2)
  --format <fmt>      Output format: table, csv, json (default: table)
  --tau <days>        Survival decay parameter (default: 180)
  --sample <n>        Max files to blame per repo (default: 500)
  --workers <n>       Concurrent blame workers (default: 4)
  --domain <name>     Only analyze repos in this domain (e.g. Backend, Frontend, Mobile)
  --active-days <n>   Days to consider author active (default: 30)
  --pressure-mode     Change pressure mode: include (default) or ignore
```

### Team Analysis (`eis team`)

Aggregates individual scores into team-level health metrics and **5-axis team classification** (Structure / Culture / Phase / Risk / Character). Classification is influence-weighted — high-scoring members shape the team's identity more.

![Team Output](docs/images/team-output.svg?v=0.11.0)

| Health Axis | What it measures |
|---|---|
| **Complementarity** | Role diversity coverage (5 known roles) |
| **Growth Potential** | Growing members + mentor (Builder/Cleaner) presence |
| **Sustainability** | Inverse of risk state ratio (Former/Silent/Fragile) |
| **Debt Balance** | Average debt cleanup tendency (50 = neutral) |
| **Productivity Density** | Output per member, with small-team bonus |
| **Quality Consistency** | Average quality + low variance |
| **Risk Ratio** | % of members in risk states |

Structural metrics (AAR, Anchor Density, Architecture Coverage) and full classification details are covered in the [Chapter 2 blog posts](#blog-posts).

### `eis timeline` — Time-Series Analysis

Tracks how individual scores, roles, and team health evolve over time. Supports 3-month, 6-month, or yearly spans.

![Timeline Output](docs/images/timeline-html-output.png?v=0.11.0)

```bash
eis timeline --recursive ~/workspace                     # Default: last 4 quarters
eis timeline --span 6m --periods 0 --recursive ~/workspace  # Full history, 6-month spans
eis timeline --format html --output timeline.html ~/workspace  # Interactive HTML dashboard
eis timeline --format svg --output ./charts/ ~/workspace     # SVG images for blog posts
eis timeline --format ascii ~/workspace                      # Terminal line charts
```

## Configuration

See [`config.example.yaml`](config.example.yaml) for all options:

- **Domains**: explicit repo-to-domain mapping. Defaults are Backend/Frontend/Infra/Firmware; custom domains (e.g. Mobile, Data) can be added with `repos:` patterns and/or `extensions:` for auto-detection. Repos not listed use auto-detection from file extensions
- **Exclude repos**: skip specific repos from analysis
- **Production daily ref**: baseline for absolute Production scoring (default: 1000 changes/day = score 100)
- **Aliases**: merge variant git author names into canonical names
- **Exclude authors**: filter out bots and non-human contributors
- **Architecture patterns**: define which files count as "design files" for the Design axis. Defaults:
  - Backend: `*/repository/*interface*`, `*/domainservice/`, `*/router.go`, `*/middleware/`, `di/*.go`
  - Frontend: `*/core/`, `*/stores/`, `*/hooks/`, `*/types/`
  - Override in `eis.yaml` to match your project structure (e.g., `*/proto/`, `*/migrations/`, `Makefile`)
- **Blame extensions**: file extensions for blame analysis
- **Weights**: customize axis weights (default: Survival 25%, Design 20%, Production 15%, Debt 15%, Quality 10%, Breadth 10%, Indispensability 5%)
- **Survival tau**: decay half-life in days (default: 180)
- **Active days**: how recently an author must have committed to be marked active (default: 30)
- **Debt threshold**: minimum events for debt score (default: 10)
- **Teams**: named team definitions for `eis team` (optional — if omitted, each domain = one team)

### What You Get

- **Rankings table** with all 7 axis scores, total, and **Active** indicator (✓ = committed within last 30 days)
- **3-axis topology** (Role / Style / State) with confidence scores (0.0-1.0) — e.g., `Architect (1.00) / Builder (0.86) / Active (0.80)`
- **Bus Factor risk map** showing modules with dangerous ownership concentration
- Color-coded output for quick visual scanning
- **JSON / CSV export** (`--format json|csv`) for dashboards and programmatic use

### Supported Languages

Works out of the box with: Go, TypeScript/JavaScript, Python, Rust, Java, Ruby, C/C++ (including firmware/embedded), Scala, Haskell, OCaml. Additional extensions can be added via `blame_extensions` in `eis.yaml`.

### Alternative: Claude Code

For deeper qualitative analysis (actionable insights, team recommendations), you can also use Claude Code:

```bash
claude "Follow the instructions in PROMPT.md to calculate Engineering Impact Scores for my team. Use config.yaml for configuration."
```

## Blog Posts — git考古学 / Git Archaeology

**#1 — Individual Scoring**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/11/153212) — git考古学 #1：履歴だけでエンジニアの「戦闘力」を定量化する
- [English / dev.to](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c) — Git Archaeology #1: Measuring Engineering Impact from Git History Alone

**#2 — Team Analysis**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/13/060851) — git考古学 #2：エンジニアの「戦闘力」から、チームの「構造力」へ
- [English / dev.to](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f) — Git Archaeology #2: Beyond Individual Scores — Measuring Team Health

**#3 — Architect Evolution**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/14/135648) — git考古学 #3：Architectには流派がある——Git履歴が暴く進化の分岐モデル
- [English / dev.to](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga) — Git Archaeology #3: Two Paths to Architect — How Engineers Evolve Differently

**#4 — Backend Architect Concentration**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/14/155124) — git考古学 #4：Backend Architectは収束する——成仏という聖なる仕事
- [English / dev.to](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d) — Git Archaeology #4: Backend Architects Converge — The Sacred Work of Laying Souls to Rest

**#5 — Timeline**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/14/180329) — git考古学 #5：タイムライン——スコアは嘘をつかないし、遠慮も映る
- [English / dev.to](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5) — Git Archaeology #5: Timeline — Scores Don't Lie, and They Capture Hesitation Too

**#6 — Team Timeline & Evolution Models**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/14/184223) — git考古学 #6：チームは進化する——タイムラインが暴く組織の法則
- [English / dev.to](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4lei) — Git Archaeology #6: Teams Evolve — The Laws of Organization Revealed by Timelines

**#7 — Observing the Universe of Code**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/14/213413) — git考古学 #7：コードの宇宙を観測する
- [English / dev.to](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0) — Git Archaeology #7: Observing the Universe of Code

**#8 — Engineering Relativity**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/14/233602) — git考古学 #8：エンジニアリング相対性理論——同じエンジニアが異なるスコアを得る理由
- [English / dev.to](https://dev.to/machuz/git-archaeology-8-engineering-relativity-why-the-same-engineer-gets-different-scores-5dnl) — Git Archaeology #8: Engineering Relativity — Why the Same Engineer Gets Different Scores

**#9 — Origin: The Big Bang of Code Universes**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/15/054313) — git考古学 #9：Origin——コード宇宙のビッグバン
- [English / dev.to](https://dev.to/machuz/git-archaeology-9-collapse-good-architects-and-black-hole-engineers-1dcn) — Git Archaeology #9: Origin — The Big Bang of Code Universes

**#10 — Dark Matter: The Invisible Gravity**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/15/062608) — git考古学 #10：Dark Matter——見えない重力
- [English / dev.to](https://dev.to/machuz/git-archaeology-10-dark-matter-the-invisible-gravity-45ne) — Git Archaeology #10: Dark Matter — The Invisible Gravity

**#11 — Entropy: The Universe Always Tends Toward Disorder**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/15/062609) — git考古学 #11：Entropy——宇宙は常に無秩序に向かう
- [English / dev.to](https://dev.to/machuz/git-archaeology-11-entropy-the-universe-always-tends-toward-disorder-ak9) — Git Archaeology #11: Entropy — The Universe Always Tends Toward Disorder

**#12 — Collapse: Good Architects and Black Hole Engineers**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/15/062610) — git考古学 #12：Collapse——良いArchitectとBlack Hole Engineer
- [English / dev.to](https://dev.to/machuz/git-archaeology-12-collapse-good-architects-and-black-hole-engineers-3fed) — Git Archaeology #12: Collapse — Good Architects and Black Hole Engineers

**#13 — Cosmology of Code**
- [Japanese / はてなブログ](https://ma2k8.hateblo.jp/entry/2026/03/15/062611) — git考古学 #13：Cosmology of Code——コード宇宙論
- [English / dev.to](https://dev.to/machuz/git-archaeology-13-cosmology-of-code-dci) — Git Archaeology #13: Cosmology of Code

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
- [x] Domain separation (BE/FE/Infra/Firmware + custom) with auto-detection
- [x] Absolute scoring for Production (per-day rate) and Quality (fix ratio)
- [x] Configurable domain mapping, repo exclusion
- [x] JSON / CSV output format (`--format json|csv`)
- [x] Team-level analysis (`eis team`) with 7 health axes
- [ ] GitHub Action for automated quarterly tracking
- [x] Timeline analysis (`eis timeline`) with per-period scoring
- [x] Chart visualization (`--format ascii|html|svg`)
- [ ] Multi-language commit message support for Quality detection

## Special Thanks

- [@reizist](https://github.com/reizist) — identified that `exclude_file_patterns` was not applied to git log and blame targets
- [@ponsaaan](https://github.com/ponsaaan) — pointed out that `config.example.yaml` was outdated and mismatched with the current config structure. Debugged the submodule hang issue in debt analysis. Also the former architect whose well-designed code continues to serve as the foundation of our product. Living proof that good design creates common sense
- [@exoego](https://github.com/exoego) — implemented fully customizable domains ([#1](https://github.com/machuz/engineering-impact-score/pull/1)): custom domain definitions with extension mapping, `default_domains: false` for complete domain redefinition, and backward-compatible YAML parsing. Shipped with 900+ lines of comprehensive tests. Also improved config UX ([#2](https://github.com/machuz/engineering-impact-score/pull/2)): `--config` now raises an early error when the specified file doesn't exist, preventing silent fallback to defaults during long analyses

## Support

If this metric helped you see your team differently, consider supporting the project:

- [GitHub Sponsors](https://github.com/sponsors/machuz)
- PayPay: `w_machu7`

## License

MIT

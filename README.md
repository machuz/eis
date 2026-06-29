<p align="center">
  <img src="docs/images/logo-full.svg?v=4" alt="EIS — the Git Telescope" width="420">
</p>

<p align="center">
  <a href="https://orbitlens.io">
    <img src="https://api.orbitlens.io/card/machuz/orbit" alt="EIS Orbit" width="80%">
  </a>
  <br>
  <a href="https://orbitlens.io">
    <img src="https://api.orbitlens.io/card/machuz/team" alt="EIS Observatory" width="80%">
  </a>
</p>

<p align="right">
  <a href="https://orbitlens.io"><img src="https://api.orbitlens.io/badge/machuz/team" alt="EIS Aggregate"></a>
  &nbsp;
  <a href="docs/whitepaper.md"><img src="https://img.shields.io/badge/Whitepaper-EN-504945" alt="Whitepaper EN"></a>
  <a href="docs/whitepaper-ja.md"><img src="https://img.shields.io/badge/Whitepaper-JA-504945" alt="Whitepaper JA"></a>
  <a href="https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c"><img src="https://img.shields.io/badge/dev.to-Read%20Article-0A0A0A?logo=devdotto&logoColor=white" alt="dev.to"></a>
  <a href="https://github.com/sponsors/machuz"><img src="https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github" alt="Sponsor"></a>
</p>

**Engineering Impact Signal** (EIS, pronounced *"ace"*)

```bash
brew tap machuz/tap && brew install eis
eis analyze --recursive ~/your-workspace
```

Git records commits.
EIS reveals the structure of the code universe.

> *Software Cosmology* — the idea that codebases behave like universes: with gravity, entropy, collapse, and structure. EIS is the telescope used to observe them.

### What the telescope reveals

Point it at a codebase, and hidden patterns emerge:

- **Architects** — who actually shaped the system's structure
- **Anchors** — the quiet forces holding the system together
- **Dark matter** — invisible forces stabilizing the code universe
- **Collapse risk** — bus-factor concentrations before they become emergencies
- **Entropy** — where code is decaying and who is fighting it

All from `git log` and `git blame`. Nothing else.

![Terminal Output](docs/images/terminal-output.svg?v=0.12.0)

## Why this exists

Most engineering metrics measure activity.

PR counts. Commit counts. Lines of code.

But activity is not impact.

An engineer who writes 10,000 lines that get rewritten next month left no gravity. An engineer who writes 500 lines that become the foundation of the system shaped its universe.

**EIS measures the gravity engineers leave in the codebase** — structure that survives, design that endures, and debt that gets cleaned up.

EIS is not a performance metric. It is an observational instrument. Like Galileo's telescope, it doesn't decide what exists — it simply reveals structures that were already there.

### Prior art

The individual signals EIS uses — code ownership, bus factor, survival analysis, developer productivity — have been studied extensively. EIS was developed independently, not derived from these works, but the underlying ideas overlap. What EIS adds is the integration: a single weighted framework that combines these signals into a structural observation tool, mapped to a cosmological model of software.

---

## How it works

EIS reads `git log` and `git blame` to compute 7 axes:

| Signal | What it captures |
|---|---|
| **Production** | Sustained output velocity |
| **Catalysis** | Others build on your *surviving* foundation — enabling later contributors |
| **Survival** | Code that endures — exponential time-decay weighted blame |
| **Design** | Commits to architecture-defining files |
| **Breadth** | Effective number of modules shaped (survival-weighted module diversity) |
| **Debt Cleanup** | Fixing others' bugs vs. generating new ones |
| **Indispensability** | Modules where one engineer owns 80%+ of blame |

From these signals, EIS derives **Roles** (Architect, Anchor, Producer...), **Styles** (Builder, Emergent, Rescue...), and **States** (Active, Growing, Fragile...) — a 3-axis topology for each engineer.

**Survival is the core thesis.** Exponential time-decay ensures that "are you *still* writing durable code?" matters most. This is naturally resistant to gaming — busy work doesn't survive.

---

## Quick Start

```bash
# Install
go install github.com/machuz/eis/cmd/eis@latest
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

# Team-level analysis (aggregates individual signals)
eis team --recursive ~/workspace

# Team analysis with JSON output (paste into AI for deeper insights)
eis team --format json --recursive ~/workspace
```

## GitHub Actions

Run EIS in CI and upload the signals to the OrbitLens observatory. The repo is
analyzed on the runner and **only the signals are sent — your source code never
leaves the runner.** This is the path for GitHub Enterprise or any environment
that does not allow cloning repos to an external service.

```yaml
# .github/workflows/eis.yml
name: EIS
on: { push: { branches: [main] } }
jobs:
  upload:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }   # full history — EIS reads git log/blame
      - uses: machuz/eis@v2
        with:
          token: ${{ secrets.ACE_API_TOKEN }}
          # api-url: https://api.orbitlens.io   # default; set for staging/self-host
```

- **`token` (required)** — create one in OrbitLens Ace → **Settings → API tokens**
  and store it as the repo secret `ACE_API_TOKEN`.
- **`fetch-depth: 0` is required** — a shallow clone has no history, so EIS
  cannot read `git log`/`git blame`. The action fails fast if the checkout is
  shallow.
- Other inputs: `api-url` (default `https://api.orbitlens.io`; override for
  staging or self-host), `version` (default: latest stable release),
  `working-directory` (default `.`), and `args` (extra flags for `eis analyze`).

A full example workflow lives at
[`examples/github-actions/eis.yml`](examples/github-actions/eis.yml).

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
| **Engineering Impact Signal** | **Code that survives over time** | **Git log + blame** | **Yes** | Accuracy depends on codebase design quality |

The core gap this model fills: **existing frameworks measure activity or velocity, not whether individual contributions actually lasted.** DORA tells you how fast code reaches production. This model tells you whether it was worth deploying.

Time-decayed survival is also naturally resistant to gaming — you can't inflate your signal with busy work, because only code that remains in the codebase months later counts.

## The 7 Axes

| Axis | Weight | What it measures |
|---|---|---|
| **Production** | 10% | Changes per day (absolute: configurable `production_daily_ref`, default 1000) |
| **Catalysis** | 15% | Surviving mass others built on top of your still-living foundation — weighted by how much of your foundation remains |
| **Code Survival** | **25%** | Recency-weighted blame survival (tau=180 exponential decay); split 80/20 robust/dormant in Impact |
| **Design** | 20% | Commits to architecture files |
| **Breadth** | 10% | Effective number of modules shaped — the survival-weighted diversity (Hill number) of modules your living code spans |
| **Debt Cleanup** | 15% | Ratio of others' debt cleaned vs. own debt generated |
| **Indispensability** | 5% | Modules where you own 80%+ of blame lines (Bus Factor) |

**Code Survival is the core thesis** — exponential time decay ensures "are you *still* writing durable designs?" matters most.

**Gravity** is shown alongside signals but excluded from Impact. It measures structural influence — how much the system's shape depends on an engineer — and a good architect writes code that **survives under others' pressure** *and* that others **build on**. So gravity passes a structural **shape** (`0.45·Design + 0.25·Breadth + 0.30·Indisp`) through two NECESSARY gates: `catGate = 0.15 + 0.85·Catalysis/100` and `survGate = 0.15 + 0.85·RobustSurvival/100`, giving `Gravity = catGate · survGate · shape`. Shape is computed from change *history*, so a foundation that was later rewritten still reads high on shape — but the survival gate collapses it. **RobustSurvival** is *others-contested* survival: code counts only where authors **other than you** actively commit. So a foundation rewritten away (survival→0) collapses, **and so does code that "survives" only because nobody else has touched it** — a dead corner, or a module you alone churn. Both are untested by anyone but you, so neither earns structural dependence; a founder whose module others keep editing still passes. Catalysis (zero in a solo project) keeps a one-person project from minting gravity. Color-coded by health: green = durable influence, yellow = moderate, red = bus-factor risk.

## Signal Guide

**40 = Senior.** This metric is deliberately harsh. Reaching 40+ across 7 relative axes requires serious, well-rounded ability.

> **Important: EIS measures impact on *this codebase*, not absolute engineering ability.** A strong signal means "on this codebase, this person's contributions are surviving, shaping architecture, and cleaning up debt." It does not mean they are a better engineer than someone with a weak signal. If signals don't match your gut feeling, that's worth investigating: it may reveal codebase design issues rather than people issues.

EIS is not meant to replace human judgment. It reveals patterns that humans often sense but cannot quantify.

![Signal Guide](docs/images/score-guide.svg?v=0.12.0)

## Engineer Topology (3-Axis Classification)

Instead of a single archetype label, EIS v0.9+ classifies each engineer along 3 independent axes. This gives a richer, more composable picture than a flat type.

### Role — what they contribute

| Role | Key Signals | Description |
|---|---|---|
| **Architect** | Design↑ RobustSurv↑ Breadth○ | Shapes system design with durable code under change pressure |
| **Anchor** | Cat↑ notLow(Prod) | A catalyst others build on, with real output — not yet shaping architecture |
| **Cleaner** | Surv↑ Debt↑ | Durable code, actively cleans others' debt |
| **Producer** | notLow(Prod) | Meaningful production output |
| **Specialist** | Surv↑ Breadth↓ | Deep in narrow area, high survival but low breadth |
| **—** | | No dominant role signal |

### Style — how they contribute

| Style | Key Signals | Description |
|---|---|---|
| **Builder** | Prod↑ Design↑ Debt○ | Designs, builds heavily, AND cleans up — the full package |
| **Resilient** | Prod↑ Surv↓ RobustSurv○ | Iterates heavily but what survives under pressure is durable |
| **Rescue** | Prod↑ Surv↓ Debt↑ | Takes over and rewrites inherited legacy code |
| **Churn** | Prod↑ Surv↓ gap≥30 | Constant rework — high output, little survives |
| **Mass** | Prod↑ Surv↓ | High output but code doesn't survive |
| **Emergent** | Gravity↑ Prod○ RobustSurv↓ | Creating new structural gravity not yet battle-tested — a future Architect candidate |
| **Balanced** | Impact≥30 | Steady contributor, no dominant pattern |
| **Spread** | Breadth↑ Prod↓ Surv↓ Design↓ | Wide presence, shallow depth everywhere |
| **—** | | No dominant style signal |

### State — lifecycle phase

| State | Key Signals | Risk |
|---|---|---|
| **Active** | Recent commits | — |
| **Growing** | Prod↓ Cat↑ | — |
| **Former** | RawSurv↑ Surv↓ Design/Indisp↑ | **⚠️ Handoff** |
| **Silent** | Prod↓ Surv↓ Debt↓ (≥100 commits) | **High** |
| **Fragile** | Indisp↑ Prod↓ + dormant + untested | **⚠️ Hidden** |
| **—** | | No dominant state signal |

### Reading the output

Each engineer gets a 3-label profile. Examples:

| Profile | Interpretation |
|---|---|
| Architect / Builder / Active | Core contributor: designs, builds, cleans, currently active |
| Producer / Mass / — | High output but code doesn't last |
| — / Spread / Silent | Wide but shallow, not contributing meaningfully |
| Anchor / Balanced / Growing | Others build on their work, steady pace, on a growth trajectory |
| — / — / Fragile | Code survives only due to low change pressure |

**Churn, Mass, and Spread styles look productive on individual metrics** but show low impact overall. Only multi-axis observation exposes them.

![Archetypes Radar](docs/images/archetypes-radar.svg?v=0.12.0)

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

### Normalization & Impact

```python
# Absolute axes (cross-org comparable):
#   Production: min(changes_per_day / production_daily_ref * 100, 100)
#   Debt: bounded 0-100 scale

# Relative axes (max-normalized within domain; damped in pools < 3 contributors):
#   Catalysis, Survival, Design, Breadth, Indispensability

# Observed per domain (Backend/Frontend/Infra/Firmware + custom domains, separately)
# Survival is split 80/20 into robust/dormant; design is damped when neither
# robust survival nor production proves it; profiles that survive NOWHERE take a 0.8x penalty.
total = (
    norm_production * 0.10
    + norm_catalysis * 0.15
    + norm_survival * 0.25      # = robust * (0.25*0.8) + dormant * (0.25*0.2)
    + norm_design * 0.20
    + norm_breadth * 0.10
    + norm_debt_cleanup * 0.15
    + norm_indispensability * 0.05
)
```

## Module Topology (3-Axis)

EIS doesn't just profile engineers — it classifies **modules** too. Each module in the codebase is observed on 4 structural indicators and classified along 3 independent axes:

### 3 Axes

| Axis | What it measures | Classifications |
|---|---|---|
| **Coupling** | Boundary quality — implicit co-change dependencies | Isolated / Independent / Linked / Hub |
| **Vitality** | Change pressure × code survival × test coverage | Stable / Fragile / Warming / Turbulent / Critical / Dead |
| **Ownership** | Knowledge distribution across engineers | Distributed / Concentrated / Orphaned |

### 4 Structural Indicators (0-100)

| Indicator | What it measures | Calculation |
|---|---|---|
| **Boundary Integrity** | How clean the module boundary is | `(1 - avgCoupling) × 100` |
| **Change Absorption** | How well changes survive | Per-module time-decayed survival rate |
| **Knowledge Distribution** | How well ownership is spread | Based on ownership level + Shannon entropy |
| **Stability** | How infrequently the module changes | `(1 - percentileRank(pressure)) × 100` |

### Combination Examples

| Coupling | Vitality | Ownership | Interpretation |
|---|---|---|---|
| Independent | Stable | Distributed | Ideal. Maintain as-is |
| Hub | Critical | Concentrated | Worst case. Leaky boundary + high churn + single owner |
| Independent | Dead | Orphaned | Clean boundary but abandoned — safe to remove? |
| Independent | **Fragile** | Concentrated | Fossil fortress. Low churn + no tests + solo owner — invisible risk until the owner leaves |
| Hub | Stable | Distributed | Dependency center but healthy — possibly intentional design |
| Independent | Turbulent | Orphaned | Actively changing but owner has left — handoff needed |

The CLI shows only **anomalous modules** (Hub, Turbulent, Critical, Dead, Orphaned) with a summary line counting all modules. This focuses attention on structural risks rather than listing hundreds of healthy modules.

```
─── Module Topology (3-axis) ───
  587 modules — Coupling: 9 Hub, 64 Linked, 502 Independent, 12 Isolated | ...
Module                            Boundary  Absorption  Knowledge  Stability  Coupling  Vitality  Ownership
app/core/apperror                 94        11          6          1          Independent  Critical  Orphaned
app/domain/payment                97        26          80         7          Independent  Turbulent Distributed
```

Module topology requires `--pressure-mode include` (the default).

## Design Principles

- **Domains are observed separately** (Backend/Frontend/Infra/Firmware by default, plus custom domains) — mixing them contaminates rankings; auto-detected from file extensions or configured explicitly
- **Hybrid observation** — Production and Debt use absolute scales (cross-org comparable); Catalysis, Survival, Design, Breadth, and Indispensability use relative normalization within domain (damped in pools below 3 contributors, where a max of one or two people is not a real percentile)
- **Debt threshold** — members with fewer than 10 debt events get a neutral signal (50) to avoid extreme ratios
- **Comments don't count** — comment-only and blank lines in code files (Go/TS/Py/Ruby/SQL/etc.) are excluded from Production, Survival, Design, and Debt. You can't inflate your scores by spamming `//` comments. Prose files (`.md`, `.txt`, `.rst`) are counted verbatim so that documentation and research writing are fully preserved.
- **Untested code is worth half** — blame lines whose source file has no test coverage contribute to Survival at α=0.5 weight. Test coverage is inferred from sibling-pair naming (`foo.go` ↔ `foo_test.go`, `foo.test.ts`, `test_foo.py`, etc.) with a same-directory fallback. The `tested_survival` / `untested_survival` fields are exposed in JSON for downstream analysis; override the weight via `untested_survival_weight` in `eis.yaml`.
- **Accuracy scales with codebase design quality** — well-structured codebases (Clean Architecture, DDD) yield more meaningful signals. If the signal doesn't match gut feeling, it may indicate poor codebase structure rather than a metric problem

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

Aggregates individual signals into team-level health metrics and **5-axis team classification** (Structure / Culture / Phase / Risk / Character). Classification is influence-weighted — high-impact members shape the team's identity more.

![Team Output](docs/images/team-output.svg?v=0.12.0)

| Health Axis | What it measures |
|---|---|
| **Complementarity** | Role diversity coverage (5 known roles) |
| **Growth Potential** | Growing members + mentor (Builder/Cleaner) presence |
| **Sustainability** | Inverse of risk state ratio (Former/Silent/Fragile) |
| **Debt Balance** | Average debt cleanup tendency (50 = neutral) |
| **Productivity Density** | Output per member, with small-team bonus |
| **Catalysis Consistency** | Average catalysis + low variance (how evenly the team builds on each other) |
| **Risk Ratio** | % of members in risk states |

Structural metrics (AAR, Anchor Density, Architecture Coverage) and full classification details are covered in the [Chapter 2 blog posts](#blog-posts).

### `eis timeline` — Time-Series Analysis

Tracks how individual signals, roles, and team health evolve over time. Supports 3-month, 6-month, or yearly spans.

![Timeline Output](docs/images/timeline-html-output.png?v=0.12.0)

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
- **Production daily ref**: baseline for absolute Production observation (default: 1000 changes/day = signal 100)
- **Aliases**: merge variant git author names into canonical names
- **Exclude authors**: filter out bots and non-human contributors
- **Architecture patterns**: define which files count as "design files" for the Design axis. Segment-precise globbing: `*` matches exactly one path segment, `**` matches any depth, and a pattern ending in `/` matches files under that directory. Defaults:
  - Backend: `**/repository/*interface*`, `**/domainservice/`, `**/router.go`, `**/middleware/`, `**/di/*.go`
  - Frontend: `**/core/`, `*/stores/`, `*/hooks/`, `*/types/` (generic names stay single-level so framework routes like `app/(site)/stores/` are not miscounted)
  - Override in `eis.yaml` to match your project structure (e.g., `**/proto/`, `**/migrations/`, `Makefile`). Use `**` for markers at any depth; use `*` to anchor to one level.
- **Blame extensions**: file extensions for blame analysis
- **Weights**: customize axis weights (default: Survival 25%, Design 20%, Catalysis 15%, Debt 15%, Production 10%, Breadth 10%, Indispensability 5%)
- **Survival tau**: decay half-life in days (default: 180)
- **Active days**: how recently an author must have committed to be marked active (default: 30)
- **Debt threshold**: minimum events for debt signal (default: 10)
- **Teams**: named team definitions for `eis team` (optional — if omitted, each domain = one team)

### What You Get

- **Rankings table** with all 7 axis signals, Impact, and **Active** indicator (✓ = committed within last 30 days)
- **3-axis topology** (Role / Style / State) with confidence values (0.0-1.0) — e.g., `Architect (1.00) / Builder (0.86) / Active (0.80)`
- **Bus Factor risk map** showing modules with dangerous ownership concentration
- Color-coded output for quick visual scanning
- **JSON / CSV export** (`--format json|csv`) for dashboards and programmatic use

### Supported Languages

Works out of the box with: Go, TypeScript/JavaScript, Python, Rust, Java, Ruby, C/C++ (including firmware/embedded), Scala, Haskell, OCaml. Additional extensions can be added via `blame_extensions` in `eis.yaml`.

### Alternative: Claude Code

For deeper qualitative analysis (actionable insights, team recommendations), you can also use Claude Code:

```bash
claude "Follow the instructions in PROMPT.md to calculate Engineering Impact Signals for my team. Use config.yaml for configuration."
```

## Library

All three books are free to read on **OrbitLens Library**.

👉 **https://library.orbitlens.io/**

- **Book 1 — Psychological OS** (心理OS) — the internal operating principle for staying in motion by your own will
- **Book 2 — Structure-Driven Engineering Organization Theory** (構造駆動エンジニアリング組織論) — observation over evaluation, structure over people
- **Book 3 — Git Archaeology** (git考古学) — gravity, civilization, and the age of AI, read from the strata of code (17 chapters)

Each chapter is mirrored on [dev.to](https://dev.to/machuz) for English readers.

```
      ✦       *        ✧

       ╭────────╮
      │    ✦     │
       ╰────┬───╯
   .        │
            │
         ___│___
        /_______\

   ✧     the Git Telescope     ✦
```

### Software Cosmology

```
                      ✦ AI Era
                     /       \
                Culture     Stars (AI)
                    |          |
                Civilization   |
              /      |      \  |
       Architect   Anchor  Producer
              \      |      /
                 Gravity
                /       \
          Dark Matter   Entropy
               |           |
               |        Collapse
                \        /
              Code Universe
            /      |      \
       Origin    Stars   Relativity
                   |
                Timeline
                   |
             Git Telescope
                   |
              Git History
```

```
Physics             Software
─────────────────────────────
Matter        →     Code
Gravity       →     Reuse
Dark Matter   →     Anchor, PO, PdM, QA...
Entropy       →     Tech Debt
Stars         →     Engineers
Civilization  →     Team
Telescope     →     EIS
```

## Roadmap

- [x] Observation methodology and formulas
- [x] Claude Code analysis prompt
- [x] Data collection script
- [x] Configuration template
- [x] Standalone CLI tool (`eis`) — zero AI dependency
- [x] Homebrew distribution (`brew install eis`)
- [x] Recursive repo discovery (`--recursive`)
- [x] Author alias mapping via config
- [x] Concurrent blame analysis (worker pool)
- [x] Domain separation (BE/FE/Infra/Firmware + custom) with auto-detection
- [x] Absolute observation for Production (per-day rate) and Debt (bounded ratio)
- [x] **Catalysis axis** (others build on your surviving foundation) replacing first-pass Quality
- [x] **Gravity as a conjunctive gate** — `catGate(Catalysis) · survGate(RobustSurvival) · shape`, with others-contested survival and small-pool normalization damping
- [x] Configurable domain mapping, repo exclusion
- [x] JSON / CSV output format (`--format json|csv`)
- [x] Team-level analysis (`eis team`) with 7 health axes
- [x] **Module Topology** — 3-axis module classification (Coupling / Vitality / Ownership) with 4 structural indicators
- [ ] GitHub Action for automated quarterly tracking
- [x] Timeline analysis (`eis timeline`) with per-period profiling
- [x] Chart visualization (`--format ascii|html|svg`)
- [ ] **[OSS Gravity Map](research/oss-gravity-map/)** — empirical validation against 25 major OSS projects (React, Kubernetes, Terraform, Redis, Rust, etc.). Architect detection, hidden architect discovery, entropy fighter detection, collapse risk analysis. ~50,000 engineers across 8 languages. [Configs open for PRs](research/oss-gravity-map/CONTRIBUTING.md)

## Special Thanks

- [@reizist](https://github.com/reizist) — identified that `exclude_file_patterns` was not applied to git log and blame targets
- [@ponsaaan](https://github.com/ponsaaan) — pointed out that `config.example.yaml` was outdated and mismatched with the current config structure. Debugged the submodule hang issue in debt analysis. Also the former architect whose well-designed code continues to serve as the foundation of our product. Living proof that good design creates common sense
- [@exoego](https://github.com/exoego) — implemented fully customizable domains ([#1](https://github.com/machuz/eis/pull/1)): custom domain definitions with extension mapping, `default_domains: false` for complete domain redefinition, and backward-compatible YAML parsing. Shipped with 900+ lines of comprehensive tests. Also improved config UX ([#2](https://github.com/machuz/eis/pull/2)): `--config` now raises an early error when the specified file doesn't exist, preventing silent fallback to defaults during long analyses

## Support

If this metric helped you see your team differently, consider supporting the project:

- [GitHub Sponsors](https://github.com/sponsors/machuz)
- PayPay: `w_machu7`

## License

MIT

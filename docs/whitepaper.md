# Engineering Impact Signal: Quantifying Software Engineering Contributions from Git History

**Version 0.12.0** — March 2026

**Author:** machuz ([@machuz](https://github.com/machuz))

---

## Abstract

Engineering Impact Signal (EIS) is an open-source framework that quantifies individual and team-level software engineering contributions using only Git history data. Unlike existing approaches that rely on proxy metrics (commit counts, lines of code, PR throughput), EIS constructs a multi-axis observation model that captures *what kind* of contribution an engineer makes, *how* they contribute, and *where they are* in their professional lifecycle. The framework combines commit-based production metrics with `git blame`-based survival analysis and a novel change-pressure decomposition to distinguish code that endures under active development from code that merely persists in dormant modules.

At the team level, EIS aggregates individual signals into a 5-axis classification system that characterizes team structure, culture, lifecycle phase, risk profile, and overall character. A timeline analysis mode tracks these metrics across configurable time periods, enabling longitudinal observation of engineering organizations.

This paper presents the mathematical foundations, classification algorithms, and design rationale of EIS, along with its limitations and intended use cases.

**Keywords:** software engineering metrics, git analysis, code survival, team health, engineering observation, developer productivity

---

## 1. Introduction

### 1.1 The Problem

Measuring software engineering contributions is a long-standing challenge. Traditional metrics fail in predictable ways:

- **Commit count** rewards granularity of commits, not impact
- **Lines of code** penalizes refactoring and rewards verbosity
- **PR throughput** measures process compliance, not engineering quality
- **Story points** are subjective and team-specific

These metrics share a common flaw: they measure *activity*, not *impact*. An engineer who writes 10,000 lines that are deleted next quarter and an engineer who writes 500 lines that become the foundation of the system's architecture are indistinguishable by activity metrics.

### 1.2 Key Insight

Git repositories contain far more information than commit counts. Specifically, `git blame` reveals which code *survived* — which lines written by which authors still exist in the current codebase. Combined with commit metadata (timestamps, file paths, message patterns), this creates a rich signal about the nature and durability of contributions.

EIS exploits this insight through three mechanisms:

1. **Survival analysis**: Using `git blame` to measure how much of an engineer's code persists, with exponential time-decay weighting
2. **Change-pressure decomposition**: Splitting survival into *robust* (survived in frequently-modified modules) and *dormant* (survived in rarely-touched modules) components
3. **Architectural pattern detection**: Identifying contributions to structurally significant files (interfaces, routers, dependency injection, domain services)

### 1.3 Design Principles

#### Foundational Principles

1. **Local Universes**: Every codebase is its own universe. Impact must be understood in its local context — normalization, archetypes, and team classifications are all computed relative to the repository under observation.
2. **Observable Gravity**: Influence appears as gravity in code: survival, reuse, and structural pull. EIS measures this gravity, not activity.
3. **Evolutionary Trajectories**: Software evolves over time. Engineering value appears in the trajectory of that evolution — not in any single snapshot.

**The Principle of Observers.** EIS does not define value. It acts as an observational instrument, revealing structures that already exist in a codebase. Like a telescope, it does not change the universe. It only makes its gravity visible.

#### Technical Principles

1. **Git-only**: No integration with project management tools, CI systems, or code review platforms. The analysis requires only a Git repository.
2. **Multi-axis**: No single signal captures engineering contribution. EIS produces 7 individual axes and 5 team-level classification axes.
3. **Relative + Absolute**: Some metrics (Production) use absolute references for cross-team comparability; others (Design, Survival) use within-team normalization.
4. **Observable, not prescriptive**: EIS describes what happened in the codebase. It does not define what *should* happen.

---

## 2. Related Work

### 2.1 DORA Metrics

The DevOps Research and Assessment (DORA) framework measures four key metrics: deployment frequency, lead time for changes, change failure rate, and time to restore service. DORA operates at the team/organization level and focuses on delivery pipeline performance rather than individual contribution patterns. EIS complements DORA by providing the individual-level resolution that DORA intentionally avoids.

### 2.2 CodeScene

CodeScene performs behavioral code analysis, identifying hotspots, code health, and organizational patterns. It uses a proprietary observation model and focuses on code-level health rather than engineer-level classification. EIS differs in its explicit multi-axis classification system and its open-source, reproducible observation methodology.

### 2.3 git-fame and git-quick-stats

These tools provide basic attribution statistics (lines of code per author, commit counts). EIS builds on the same raw data but adds survival analysis, architectural pattern detection, and multi-axis profiling that these tools lack.

### 2.4 Academic Research

Nagappan et al. (2008) demonstrated that organizational metrics derived from version control predict software defects better than code metrics alone. Bird et al. (2011) showed that code ownership patterns correlate with software quality. EIS formalizes these insights into a practical, automated framework.

---

## 3. Individual Profiling Model

### 3.1 Overview

EIS computes 7 axes for each contributor in a repository:

| Axis | Signal | Scale | Source |
|------|--------|-------|--------|
| Production | Volume of code changes | Absolute | Commits |
| Quality | Inverse of fix ratio | Absolute | Commit messages |
| Survival | Durability of code (time-decayed) | Relative | `git blame` |
| Design | Changes to architectural files | Relative | Commits + patterns |
| Breadth | Diversity of modules touched | Relative | Commits |
| Debt Cleanup | Ratio of fixing others' code vs. creating debt | Absolute | `git blame` + commits |
| Indispensability | Bus factor — sole ownership of modules | Relative | `git blame` |

Each axis produces a signal in [0, 100].

### 3.2 Production

Production measures the raw volume of code changes.

**Calculation:**

$$\text{Production}_a = \sum_{c \in \text{commits}(a)} \sum_{f \in \text{files}(c)} (\text{insertions}_f + \text{deletions}_f)$$

**Normalization (Absolute):**

$$\text{Score}_a = \min\left(\frac{\text{Production}_a / \text{activeDays}_a}{\text{ProductionDailyRef}} \times 100,\; 100\right)$$

Where `ProductionDailyRef` (default: 1000 lines/day) provides a fixed baseline for cross-team comparability. Files matching exclusion patterns (lock files, generated code, swagger docs) are excluded.

**Rationale:** Production is deliberately kept as an absolute metric. A team of two engineers and a team of twenty should be comparable on per-person output.

### 3.3 Quality

Quality measures the inverse of an engineer's fix ratio.

**Calculation:**

$$\text{FixRatio}_a = \frac{|\{c \in \text{commits}(a) : \text{isFix}(c)\}|}{|\text{commits}(a) \setminus \text{merges}(a)|}$$

$$\text{Quality}_a = 100 - \text{FixRatio}_a \times 100$$

**Fix Detection:** A commit is classified as a fix if its message matches:

```
(?i)^[^\w]*(?:\[?\s*(?:fix|revert|hotfix)\s*\]?[:/\s])
```

or contains the Japanese word "修正" (fix/correction). Merge commits contribute to fix count but not to total count, preventing merge-heavy workflows from inflating quality signals.

**Rationale:** An engineer whose commits are predominantly fixes is spending time correcting mistakes (their own or others'). Quality is an absolute scale — it does not depend on team composition.

### 3.4 Survival

Survival is the central innovation of EIS. It measures how much of an engineer's code persists in the current codebase, weighted by time.

**Raw Survival:**

$$\text{RawSurvival}_a = |\{l \in \text{blame} : \text{author}(l) = a\}|$$

**Time-Decayed Survival:**

$$\text{Survival}_a = \sum_{l \in \text{blame}(a)} e^{-d_l / \tau}$$

Where $d_l$ is the age of line $l$ in days and $\tau$ is the decay constant (default: 180 days). This weighting ensures that recently-written surviving code is valued more highly than ancient code that may persist only due to inertia.

#### 3.4.1 Change-Pressure Decomposition

Not all surviving code is equal. Code in a module that receives 50 commits per quarter has been *tested by collaboration* — other engineers have worked around it, modified adjacent code, and it has survived. Code in a module that receives 1 commit per quarter may survive simply because no one has looked at it.

EIS introduces **change pressure** to distinguish these cases:

$$\text{Pressure}_m = \frac{|\text{commits touching module } m|}{|\text{blame lines in module } m|}$$

Using the median pressure as threshold:

- **Robust Survival**: Lines in modules where $\text{Pressure}_m \geq \text{median}$
- **Dormant Survival**: Lines in modules where $\text{Pressure}_m < \text{median}$

Both use the same exponential decay formula but are computed independently.

**Rationale:** This decomposition prevents the "forgotten code" problem — where an engineer signals highly on survival simply because they wrote code in a module that nobody modifies. Robust Survival specifically measures code that *endures under active development*.

### 3.5 Design

Design measures contributions to architecturally significant files.

**Calculation:**

$$\text{Design}_a = \sum_{c \in \text{commits}(a)} \sum_{f \in \text{archFiles}(c)} (\text{insertions}_f + \text{deletions}_f)$$

**Architecture Detection:** Files are classified as architectural if they match configurable glob patterns:

```
*/repository/*interface*    # Repository interfaces
*/domainservice/            # Domain services
*/router.go                 # Routing definitions
*/middleware/               # Middleware layers
di/*.go                     # Dependency injection
*/core/                     # Core modules
*/stores/                   # State management (frontend)
*/hooks/                    # React hooks (frontend)
*/types/                    # Type definitions
```

**Normalization:** Relative (max-based), so the highest Design contributor in the team signals 100.

**Rationale:** Not all code changes are equal. Changes to interfaces, dependency injection, and routing configurations have outsized structural impact compared to changes within a single module.

### 3.6 Breadth

Breadth measures the diversity of modules an engineer touches.

$$\text{Breadth}_a = |\{\text{unique modules touched by } a\}|$$

Normalized relative to the team maximum, capped by `BreadthMax` (default: 5).

**Rationale:** Engineers who contribute across many modules have different organizational impact than deep specialists. Neither is inherently better, but the distinction is meaningful for team composition analysis.

### 3.7 Debt Cleanup

Debt Cleanup measures whether an engineer cleans up others' technical debt or creates debt for others to clean.

**Algorithm:**

1. Identify fix commits (using Quality's fix detection regex)
2. Sample up to 500 fix commits
3. For each fix commit, run `git blame` on the parent commit to identify the original author of the changed lines
4. Count:
   - $\text{cleaned}_a$ = fix commits where author $a$ fixed code written by others
   - $\text{generated}_a$ = fix commits where others fixed code written by author $a$

**Signal:**

$$\text{DebtCleanup}_a = 50 + 50 \times \frac{\text{cleaned}_a - \text{generated}_a}{\text{cleaned}_a + \text{generated}_a}$$

- Signal = 0: Pure debt creator
- Signal = 50: Neutral (balanced or insufficient data)
- Signal = 100: Pure cleaner

Authors with fewer than `DebtThreshold` (default: 10) total interactions receive a neutral signal of 50.

**Rationale:** This metric captures a dimension invisible to other tools: whether an engineer's code tends to require fixes by others, or whether they tend to fix others' code. The formula is symmetric and ranges from pure creator to pure cleaner.

### 3.8 Indispensability

Indispensability measures bus factor risk at the individual level.

**Algorithm:**

1. Group `git blame` lines by module (first 2 path components, e.g., `app/domain`)
2. For each module, identify the top author by line count
3. Calculate ownership share: $\text{share} = \text{topCount} / \text{totalLines}$
4. Signal:
   - Critical ownership ($\text{share} \geq 0.80$): +1.0 per module
   - High ownership ($\text{share} \geq 0.60$): +0.5 per module

$$\text{Indispensability}_a = \text{criticalCount} \times 1.0 + \text{highCount} \times 0.5$$

Normalized relative to team maximum.

**Rationale:** High indispensability is a *risk indicator*, not an achievement metric. An engineer who owns 80%+ of multiple modules represents a single point of failure.

### 3.9 Normalization Strategy

EIS uses two normalization strategies:

**Max-Based Relative Normalization** (for Survival, Design, Breadth, Indispensability):

$$\text{Score}_a = \min\left(\frac{\text{raw}_a}{\max_b(\text{raw}_b)} \times 100,\; 100\right)$$

The highest contributor always signals 100. This is appropriate for metrics where the absolute value is meaningless outside the team context.

**Absolute Normalization** (for Production, Quality, Debt Cleanup):

Fixed reference points that allow cross-team comparison. Production uses a daily reference rate; Quality and Debt Cleanup are inherently bounded.

### 3.10 Impact

The impact is a weighted sum:

$$\text{Total} = \sum_{i} w_i \times \text{Score}_i$$

Default weights:

| Axis | Weight |
|------|--------|
| Production | 0.15 |
| Quality | 0.10 |
| Survival | 0.25 |
| Design | 0.20 |
| Breadth | 0.10 |
| Debt Cleanup | 0.15 |
| Indispensability | 0.05 |

When change-pressure data is available, Survival is split:

$$\text{Survival contribution} = w_{\text{survival}} \times (0.80 \times \text{RobustSurvival} + 0.20 \times \text{DormantSurvival})$$

Additionally, Design is damped by a proof factor:

$$\text{designDamping} = \max\left(\frac{\text{RobustSurvival}}{100} \times 0.8 + 0.2,\; \frac{\text{Production}}{100} \times 0.8 + 0.2\right)$$

$$\text{effectiveDesign} = \text{Design} \times \text{designDamping}$$

This prevents inflated design signals from engineers who own architectural files but neither produce actively nor have code that survives under pressure.

A penalty of 0.80× is applied to the impact if an engineer has zero Robust Survival, indicating their code has never been tested by collaboration.

#### Gravity Signal

A separate composite measures structural influence:

$$\text{Gravity} = 0.40 \times \text{Indispensability} + 0.30 \times \text{Breadth} + 0.30 \times \text{Design}$$

---

## 4. Individual Classification: The 3-Axis Topology

Raw signals are classified into three orthogonal axes describing the *nature* of an engineer's contribution.

### 4.1 Soft Matching Functions

Rather than hard thresholds, EIS uses sigmoid-like soft matching functions:

**highness(v)** — confidence that $v$ is "high" (≥60):

$$\text{highness}(v) = \begin{cases} 1.0 & v \geq 80 \\ 0.5 + \frac{v-60}{40} & 60 \leq v < 80 \\ \frac{v-40}{40} \times 0.3 & 40 \leq v < 60 \\ 0 & v < 40 \end{cases}$$

**lowness(v)** — confidence that $v$ is "low" (<30):

$$\text{lowness}(v) = \begin{cases} 1.0 & v < 10 \\ 0.5 + \frac{30-v}{40} & 10 \leq v < 30 \\ \frac{50-v}{40} \times 0.3 & 30 \leq v < 50 \\ 0 & v \geq 50 \end{cases}$$

**notLow(v)** — confidence that $v$ is "not low" (≥50):

$$\text{notLow}(v) = \begin{cases} 1.0 & v \geq 50 \\ 0.5 + \frac{v-30}{40} & 30 \leq v < 50 \\ \frac{v-10}{40} \times 0.3 & 10 \leq v < 30 \\ 0 & v < 10 \end{cases}$$

These functions produce continuous confidence values in [0, 1], avoiding the brittleness of hard cutoffs.

### 4.2 Role Axis — "What do they contribute?"

| Role | Confidence Formula | Interpretation |
|------|-------------------|----------------|
| **Architect** | $\min(\text{highness(Design)},\; \text{highness(Survival)},\; \text{notLow(Breadth)})$ | Shapes the structure that others build upon |
| **Anchor** | $\min(\text{highness(Quality)},\; \text{notLow(Production)})$ | Stabilizes quality across the codebase |
| **Cleaner** | $\min(\text{highness(Quality)},\; \text{highness(Survival)},\; \text{highness(DebtCleanup)})$ | Resolves others' technical debt |
| **Producer** | $\text{notLow(Production)}$ | Generates output and moves features forward |
| **Specialist** | $\min(\text{highness(Survival)},\; \text{lowness(Breadth)})$ | Deep expertise in a narrow domain |

When change-pressure data is available, Architect uses Robust Survival instead of total Survival (unless the engineer has high Production, in which case total Survival is used to avoid penalizing active builders).

**Selection:** The role with the highest confidence above the minimum threshold (0.10) is selected. In case of ties within a 0.15 margin, earlier rules take priority.

### 4.3 Style Axis — "How do they contribute?"

| Style | Confidence Formula | Interpretation |
|-------|-------------------|----------------|
| **Builder** | $\min(\text{high(Prod)},\; \text{high(Design)},\; \text{notLow(Debt)})$ | Designs, builds, and cleans up |
| **Resilient** | $\min(\text{high(Prod)},\; \text{low(Surv)},\; \text{notLow(RobustSurv)})$ | Iterates heavily; what survives pressure is durable |
| **Rescue** | $\min(\text{high(Prod)},\; \text{low(Surv)},\; \text{high(Debt)})$ | High output cleaning up legacy |
| **Churn** | $\min(\text{notLow(Prod)},\; \text{low(Quality)},\; \text{low(Surv)})$ | High output, constant rework |
| **Mass** | $\min(\text{high(Prod)},\; \text{low(Surv)})$ | High output but code doesn't last |
| **Emergent** | $\min(\text{high(Gravity)},\; \text{notLow(Prod)},\; \text{low(RobustSurv)})$ | Creating new structures not yet battle-tested |
| **Balanced** | $0.30$ (flat) | Steady contributor, no dominant pattern |
| **Spread** | $\min(\text{high(Breadth)},\; \text{low(Prod)},\; \text{low(Surv)},\; \text{low(Design)})$ | Wide presence, low depth |

### 4.4 State Axis — "Where are they in their lifecycle?"

| State | Confidence Formula | Interpretation |
|-------|-------------------|----------------|
| **Former** | $\min(\text{high(RawSurv)},\; \text{low(Surv)},\; \max(\text{high(Design)},\; \text{high(Indisp)}))$ | Code persists but engineer is inactive; was important |
| **Silent** | $\min(\text{low(Prod)},\; \text{low(Surv)},\; \text{low(Debt)})$ | All observed signals are low — no production, no survival, no debt cleanup. May indicate role mismatch or an environment that hasn't activated this person's strengths |
| **Fragile** | $0.85 + \frac{\text{dormantRatio} - 80}{200}$ | Code survives only where no one touches it |
| **Growing** | $\min(\text{low(Prod)},\; \text{high(Quality)})$ | Low volume, high quality — on growth trajectory |
| **Active** | $0.80$ if recently active | Currently contributing |

Fragile requires: dormant ratio ≥80%, Indispensability ≥60, Production <40. This identifies engineers whose high survival is illusory — their code persists in dead zones.

### 4.5 Composite Label

The three axes produce labels like:

- **Architect Builder Active** — actively designing and building durable structures
- **Producer Mass Active** — high output but code doesn't survive
- **Anchor Balanced Growing** — quality-focused, still developing breadth
- **Architect Emergent Active** — creating new architectural patterns not yet proven

---

## 5. Team-Level Analysis

### 5.1 Member Categorization

Team members are categorized into three tiers:

| Tier | Criteria | Used for |
|------|----------|----------|
| **Core** | `RecentlyActive` AND `Impact ≥ 20` | Computing averages and distributions |
| **Risk** | State ∈ {Former, Silent, Fragile} | Risk detection (always included) |
| **Peripheral** | All others | Excluded from metrics |

**Weighted ratios** are used for role/style distributions:

$$w_a = \max\left(\frac{\text{Total}_a}{100},\; 0.1\right)$$

$$\text{weightedRatio}(\text{predicate}) = \frac{\sum_{a : \text{pred}(a)} w_a}{\sum_a w_a}$$

Higher-impact members carry proportionally more weight in team-level calculations.

### 5.2 Health Metrics

Six health indicators provide a diagnostic view of team condition:

#### Complementarity

$$\text{Coverage} = \frac{|\text{unique roles}|}{5} \times 80$$

$$\text{Bonus} = 10 \cdot \mathbb{1}[\text{Architect}] + 5 \cdot \mathbb{1}[\text{Anchor}] + 5 \cdot \mathbb{1}[\text{Cleaner}]$$

$$\text{Complementarity} = \text{clamp}(\text{Coverage} + \text{Bonus},\; 0,\; 100)$$

Measures role diversity. A team with all five roles and the key trio (Architect, Anchor, Cleaner) reaches 100.

#### Growth Potential

$$\text{GrowthPotential} = \frac{|\text{Growing}|}{|\text{members}|} \times 60 + 20 \cdot \mathbb{1}[\text{Builder}] + 20 \cdot \mathbb{1}[\text{Cleaner}]$$

Teams with members actively developing new skills and with mentoring capacity (Builder, Cleaner) have higher growth potential.

#### Sustainability

$$\text{Sustainability} = (1 - \text{RiskRatio}) \times 80 + 20 \cdot \mathbb{1}[\text{Architect}]$$

Where RiskRatio = proportion of Former/Silent/Fragile members. Teams with low attrition and architectural leadership are sustainable.

#### Debt Balance

$$\text{DebtBalance} = \text{clamp}(\text{AvgDebtCleanup},\; 0,\; 100)$$

Direct average of individual Debt Cleanup signals. A team averaging 50 is neutral; above 50 is net cleaning, below is net creating.

#### Productivity Density

$$\text{ProductivityDensity} = \text{AvgProduction}_{\text{core}}$$

With small-team bonus: ×1.2 for teams ≤3, ×1.1 for teams ≤5 (when AvgProduction ≥50).

#### Quality Consistency

$$\text{QualityConsistency} = 0.6 \times \text{AvgQuality} + 0.4 \times \text{clamp}(100 - 2\sigma_{\text{Quality}},\; 0,\; 100)$$

Balances high average quality with low variance. A team where everyone has 80% quality signals higher than one with 95%/65% split.

### 5.3 Team Classification: 5-Axis System

#### Character (Composite Identity)

| Character | Galaxy Analogy | Key Criteria | Interpretation |
|-----------|---------------|-------------|----------------|
| **Spiral** | Spiral galaxy — strong core, active star formation | Arch. coverage >0.4, productivity >35, balanced AAR | Architecture drives production. Gravitational core and star formation coexist |
| **Elliptical** | Elliptical galaxy — mature, stable, few new stars | Arch. structure + Stability culture | Mature and change-resistant. Solid structure, low entropy |
| **Starburst** | Starburst galaxy — explosive star formation | Arch. structure + Builder culture | Rapid expansion. High energy, structure still forming |
| **Nebula** | Nebula — stellar nursery | Builder culture + Scaling/Emerging phase | Next-generation engineers developing. Conditions for star formation exist |
| **Irregular** | Irregular galaxy — no gravitational center | High productivity, low architecture coverage | Stars form everywhere with no cohesion. High output, no structural direction |
| **Cluster** | Star cluster — dense but weakly bound | Delivery team + Mass production culture | Productive but no gravitational structure to hold it together |
| **Collision** | Colliding galaxies — structural disruption | Firefighting culture | Forces collide. Structure disrupted, energy scattered |
| **Dwarf** | Dwarf galaxy — small but long-lived | Maintenance structure + Stability culture | Compact and stable. Maintains order with minimal resources |
| **Filament** | Cosmic filament — wide, thin structure | Exploration culture | Broad reach, thin depth. Probing the large-scale structure |

#### Structure (Role Composition)

| Structure | Key Criteria |
|-----------|-------------|
| **Architectural Engine** | ≥1 Architect, ≥2 Anchors, balanced AAR, low unstructured ratio |
| **Architectural Team** | ≥1 Architect, ≥1 Anchor, low unstructured ratio |
| **Architecture-Heavy** | ≥1 Architect, AAR >2.0 (design outpaces implementation) |
| **Emerging Architecture** | ≥1 Architect, high unstructured ratio |
| **Delivery Team** | >50% Producers |
| **Maintenance Team** | No Architects, ≥40% Anchors |
| **Unstructured** | >50% unclassified |

#### Phase (Lifecycle)

| Phase | Key Criteria |
|-------|-------------|
| **Emerging** | ≥40% Growing members |
| **Scaling** | 20-40% Growing, high Growth Potential |
| **Mature** | ≥80% Active, high Sustainability |
| **Stable** | ≥60% Active |
| **Legacy-Heavy** | ≥30% Risk members, high average signals, Architect present |
| **Declining** | ≥30% Risk members, low signals or no Architect |
| **Rebuilding** | Both Growing and Risk members present |

#### Risk (Primary Concern)

| Risk | Key Criteria |
|------|-------------|
| **Bus Factor** | ≤5 members, high average Indispensability |
| **Design Vacuum** | No Architects, low Complementarity |
| **Quality Drift** | Quality Consistency ≤60 |
| **Debt Spiral** | Debt Balance ≤45 |
| **Talent Drain** | Risk ratio ≥25% |
| **Healthy** | No significant risks detected |

---

## 6. Timeline Analysis

### 6.1 Period-Based Observation

EIS supports longitudinal analysis by dividing repository history into configurable time periods (default: 3-month spans).

**Algorithm:**

1. Collect all commits once from the repository
2. For each time period $[t_{\text{start}}, t_{\text{end}})$:
   a. Filter commits where $\text{date} \leq t_{\text{end}}$
   b. Find the boundary commit (latest commit at $t_{\text{end}}$)
   c. Run `git blame` at that boundary commit: `git blame <hash> -- <file>`
   d. Compute all metrics with $\text{refTime} = t_{\text{end}}$ (not `time.Now()`)
   e. Override `ActiveDays` to cover the full period span
3. Assemble per-author and per-team timelines

**Critical Design Decision:** The `ScoreAt(refTime)` function replaces `time.Now()` with the period's end time for all recency calculations. Without this, historical periods would incorrectly mark all members as inactive.

### 6.2 Transition Detection

For each consecutive period pair, EIS detects changes in Role, Style, and State:

```
If Role[t] ≠ Role[t-1] AND neither is "—" → Transition(Role, from, to, period)
If Style[t] ≠ Style[t-1] AND neither is "—" → Transition(Style, from, to, period)
If State[t] ≠ State[t-1] AND neither is "—" → Transition(State, from, to, period)
```

These transitions reveal career trajectories: "Producer -> Anchor" (quality focus developing), "Mass -> Builder" (learning to build durably), "Active -> Former" (departure).

### 6.3 Team Timeline

Team-level timelines track:

- Classification changes across all 5 axes
- Health metric trajectories
- Membership composition shifts
- Role/Style/State distributions per period

A common pattern observed: **Architectural Team → Maintenance Team → Architectural Engine** — the progression from single-architect dependency through a maintenance phase to distributed design capability.

---

## 7. Module-Level Analysis: The 3-Axis Module Topology

### 7.1 Motivation

The individual profiling model answers "who is strong?" but not "where is the system breaking?" Module-level analysis completes the picture: by classifying every module on 3 independent axes, EIS can identify structural risks invisible from engineer signals alone.

This mirrors the engineer topology design: just as one axis cannot distinguish an "Architect who is a Builder" from an "Architect who is Spread", a single module label cannot distinguish "Hub that is Stable" from "Hub that is Turbulent".

### 7.2 Structural Indicators

Four indicators measure each module on [0, 100]:

| Indicator | What it measures | Calculation |
|-----------|-----------------|-------------|
| **Boundary Integrity** | Absence of implicit co-change coupling | $(1 - \text{avgCoupling}_m) \times 100$ |
| **Change Absorption** | How well code survives in the module | Per-module time-decayed survival: $\frac{\sum_{l \in m} e^{-d_l/\tau}}{|l \in m|}$ |
| **Knowledge Distribution** | Ownership health | Base signal from ownership level (HEALTHY→80, FRAGMENTED→50, CONCENTRATED→25, SOLE_OWNER→10) ± entropy adjustment |
| **Stability** | Infrequency of changes | $(1 - \text{percentileRank}(\text{pressure}_m)) \times 100$ |

**Co-change coupling** is measured using the Jaccard coefficient between module pairs: for modules $A$ and $B$ that appear in the same commit, $\text{coupling}(A,B) = \frac{|A \cap B|}{|A \cup B|}$. A module's average coupling is the mean Jaccard across all pairs involving that module.

**Module survival** reuses the same exponential decay formula as individual survival (Section 3.4), but aggregates by module rather than by author.

### 7.3 The Three Axes

#### Coupling — Boundary Quality

| Classification | Meaning | Detection |
|---------------|---------|-----------|
| **Isolated** | No co-change pairs | Hard gate: CouplingPairCount = 0 |
| **Independent** | Low average coupling | highness(BoundaryIntegrity) |
| **Linked** | Moderate coupling (15-60%) | Mid-range avgCoupling |
| **Hub** | High average coupling | highness(avgCoupling × 100) |

Hub modules are implicit dependency centers — they change whenever other modules change, indicating leaky boundaries or shared concerns that should be made explicit.

#### Vitality — Life Force

| Classification | Meaning | Detection |
|---------------|---------|-----------|
| **Dead** | No commits, no active owner | Hard gate: commits=0 AND !ownerActive |
| **Critical** | Extreme pressure + very low survival | pressureLevel ≥ 80 AND absorption < 20 |
| **Turbulent** | High pressure + low survival | highness(pressure) ∧ lowness(absorption) |
| **Warming** | Moderate pressure + declining survival | 30 ≤ pressure < 70 AND absorption < 50 |
| **Stable** | Low pressure (healthy equilibrium) | highness(stability) |

Vitality classification **requires blame data**. Modules without blame lines (docs, configs, CI) can only be Dead or Stable — they cannot be Turbulent or Critical, because "no blame data" means "unknown survival", not "low survival".

#### Ownership — Knowledge Distribution

| Classification | Meaning | Detection |
|---------------|---------|-----------|
| **Orphaned** | Top author inactive + high ownership share | daysSince(topAuthor) > activeDays AND topShare ≥ 0.50 |
| **Concentrated** | Knowledge in one person's hands | SOLE_OWNER or CONCENTRATED level |
| **Distributed** | Healthy knowledge spread | HEALTHY or FRAGMENTED level |

### 7.4 Classification Engine

Module classification reuses the same `pickBest()` and `highness()`/`lowness()`/`notLow()` soft-match functions from the engineer topology (Section 4.1). Each axis is classified independently with confidence values in [0, 1].

### 7.5 Anomaly-Focused Display

The CLI shows only **anomalous modules** — those classified as Hub, Turbulent, Critical, Dead, or Orphaned on any axis. A summary line counts all modules across all classifications. This design focuses attention on structural risks rather than listing hundreds of healthy modules.

### 7.6 Combination Semantics

The power of 3-axis classification lies in combinations:

| Coupling | Vitality | Ownership | Risk Level | Action |
|----------|----------|-----------|------------|--------|
| Independent | Stable | Distributed | **None** | Maintain |
| Hub | Critical | Concentrated | **Maximum** | Immediate refactoring + knowledge transfer |
| Independent | Dead | Orphaned | **High** | Evaluate: remove or revive |
| Hub | Stable | Distributed | **Low** | Possibly intentional (shared library) |
| Independent | Turbulent | Orphaned | **High** | Owner handoff needed |
| Linked | Warming | Concentrated | **Medium** | Monitor; succession plan |

### 7.7 Relationship to Individual Profiling

Module topology complements, not replaces, individual profiling. Phase 3 of the module science roadmap will cross-reference the two: "This Critical × Orphaned module needs an Architect-type engineer assigned to it."

---

## 8. Limitations

### 7.1 Normalization Sensitivity

Relative normalization means that adding or removing a team member can change everyone's signals. The highest contributor always signals 100 on relative axes, making cross-team comparison impossible for those dimensions.

### 7.2 Commit Hygiene Dependency

Quality detection relies on commit message conventions (`fix:`, `revert:`, etc.). Teams with poor commit hygiene will have unreliable Quality signals. Squash-merge workflows may obscure individual contribution patterns.

### 7.3 Architecture Pattern Configuration

The default architecture patterns (`*/repository/*interface*`, `*/router.go`, etc.) reflect Clean Architecture conventions. Teams using different patterns must customize configuration for meaningful Design signals.

### 7.4 Monorepo Assumptions

Blame-based analysis assumes a single repository or uses `--recursive` mode to aggregate across repositories. The normalization strategy may not behave well for extremely heterogeneous monorepos.

### 7.5 Not a Performance Evaluation Tool

EIS is designed as an *observability* tool, not an evaluation tool. Using it for performance reviews without understanding its limitations would be harmful. Signals reflect what happened in the codebase, not the value of an engineer's contributions to the organization.

---

## 9. Use Cases

### 8.1 Team Health Diagnostics

A team lead can run `eis analyze --team` to understand:
- Whether design capability is concentrated (Bus Factor risk) or distributed (Architectural Engine)
- Whether the team is in maintenance mode or actively evolving
- Which health metrics are declining

### 8.2 Longitudinal Observation

`eis timeline` reveals patterns invisible in point-in-time snapshots:
- An engineer transitioning from Producer to Architect over 6 months
- A team's structure degrading after a key member's departure
- The "hesitation" pattern — an engineer whose signals dip when joining a new team, then recover

### 8.3 Hiring and Team Composition

Team-level metrics provide evidence-based answers to hiring questions:
- "Do we need another Architect or another Anchor?"
- "Is our Complementarity signal improving or declining?"
- "What would happen to our Structure classification if Engineer X left?"

### 8.4 AI-Assisted Analysis

The JSON and HTML output formats are designed for AI consumption. Feeding `eis timeline --format json` output to an LLM enables natural-language queries: "What happened to the backend team in 2024-H2?" The AI can correlate signal changes, role transitions, and health metric movements to formulate hypotheses.

---

## 10. Implementation

EIS is implemented in Go and distributed as a single binary.

```bash
brew tap machuz/tap && brew install eis

# Individual analysis
eis analyze ~/workspace/my-repo

# Team analysis
eis analyze --team ~/workspace/my-repo

# Timeline (3-month spans, last year)
eis timeline --format html --output timeline.html ~/workspace/my-repo

# Cross-repository analysis
eis analyze --recursive ~/workspace
```

**Performance:** For a repository with 500 tracked files and 4 time periods, analysis takes approximately 25 seconds (dominated by `git blame` operations). Blame is parallelized across configurable worker count.

**Source code:** [github.com/machuz/eis](https://github.com/machuz/eis)

---

## 11. Conclusion

EIS demonstrates that Git history contains far more information about engineering contributions than commonly extracted. By combining commit-based metrics with blame-based survival analysis and change-pressure decomposition, it is possible to construct a multi-dimensional view of individual and team-level engineering patterns.

The key contributions of this work are:

1. **Change-pressure decomposition** — distinguishing code that survives under active development from code that persists in dormant modules
2. **3-axis individual classification** — capturing what, how, and lifecycle state simultaneously
3. **5-axis team classification** — providing organizational diagnostics from code-level data
4. **3-axis module topology** — classifying structural health of modules (Coupling / Vitality / Ownership) using the same soft-match engine as engineer classification
5. **Soft matching functions** — avoiding hard thresholds that create classification artifacts

The framework is intentionally limited to Git data, making it universally applicable to any team using version control. Its limitations — normalization sensitivity, commit hygiene dependency, architecture pattern configuration — are acknowledged and manageable through configuration.

The ultimate insight is simple: **codebases have gravitational structures**. Some engineers write code that becomes the center around which other code is built, and that structure survives. EIS makes this gravity observable.

---

## References

1. Nagappan, N., Murphy, B., & Basili, V. (2008). The influence of organizational structure on software quality. *ICSE '08*.
2. Bird, C., Nagappan, N., Murphy, B., Gall, H., & Devanbu, P. (2011). Don't touch my code! Examining the effects of ownership on software quality. *ESEC/FSE '11*.
3. Forsgren, N., Humble, J., & Kim, G. (2018). *Accelerate: The Science of Lean Software and DevOps*. IT Revolution Press.
4. Tornhill, A. (2022). *Software Design X-Rays*. Pragmatic Bookshelf.

---

## Appendix A: Default Configuration

```yaml
tau: 180                    # Survival decay constant (days)
sample_size: 500            # Max fix commits sampled for Debt analysis
debt_threshold: 10          # Min interactions for Debt observation
breadth_max: 5              # Cap for Breadth axis
active_days: 30             # Window for "recently active"
blame_timeout: 120          # Seconds per file blame
production_daily_ref: 1000  # Baseline for Production observation

weights:
  production: 0.15
  quality: 0.10
  survival: 0.25
  design: 0.20
  breadth: 0.10
  debt_cleanup: 0.15
  indispensability: 0.05

bus_factor:
  critical: 0.80
  high: 0.60

exclude_file_patterns:
  - "package-lock.json"
  - "yarn.lock"
  - "go.sum"
  - "docs/swagger*"
  - "*generated*"
  - "mock_*"
  - "*.gen.*"

architecture_patterns:
  - "*/repository/*interface*"
  - "*/domainservice/"
  - "*/router.go"
  - "*/middleware/"
  - "di/*.go"
  - "*/core/"
  - "*/stores/"
  - "*/hooks/"
  - "*/types/"

blame_extensions:
  - "*.go"
  - "*.ts"
  - "*.tsx"
  - "*.py"
  - "*.rs"
  - "*.java"
  - "*.rb"
```

---

## Appendix B: Glossary

| Term | Definition |
|------|-----------|
| **AAR** | Architect-to-Anchor Ratio. Measures balance between design and stabilization roles. |
| **Architecture Coverage** | (Architects + Anchors) / MemberCount. Proportion of structurally-contributing members. |
| **Anchor Density** | Anchors / MemberCount. Proportion of quality-stabilizing members. |
| **Change Pressure** | Commits / Blame lines per module. Indicates how actively a module is developed. |
| **Core Member** | Recently active with Impact ≥ 20. Included in team averages. |
| **Gravity** | Composite of Indispensability, Breadth, and Design. Measures structural influence. |
| **Risk Member** | State ∈ {Former, Silent, Fragile}. Included in risk calculations. |
| **Robust Survival** | Blame lines in high-pressure modules, time-decayed. Code proven under collaboration. |
| **Dormant Survival** | Blame lines in low-pressure modules, time-decayed. Code untested by collaboration. |
| **Tau (τ)** | Exponential decay constant for Survival calculation. Default 180 days. |
| **Module Topology** | 3-axis classification of modules: Coupling, Vitality, Ownership. |
| **Co-change Coupling** | Jaccard coefficient measuring how often two modules change together. |
| **Boundary Integrity** | Module indicator: `(1 - avgCoupling) × 100`. Higher = cleaner boundary. |
| **Change Absorption** | Module indicator: per-module time-decayed survival rate. |

---

**License:** MIT

**Citation:**
```
@software{eis2026,
  title = {Engineering Impact Signal},
  author = {machuz},
  url = {https://github.com/machuz/eis},
  year = {2026}
}
```

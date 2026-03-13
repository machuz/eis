---
title: "Beyond Individual Scores: Measuring Team Health from Git History"
published: true
description: "Chapter 2 of Engineering Impact Score. Team-level analysis — complementarity, risk ratio, productivity density — all from git data you already have."
tags: opensource, productivity, git, teams
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/engineering-impact-framework-diagram-fixed.png
---

*Individual scores tell you who is strong. Team health tells you whether the team will still be strong next quarter.*

## Previously

In [Chapter 1](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c), I introduced a 7-axis scoring model for individual engineers, powered entirely by git history. The 3-axis topology (Role / Style / State) gave us a vocabulary for describing engineers: an *Architect* who writes durable design code, a *Cleaner* who quietly pays down debt, a *Producer* who ships volume.

But individual scores have a blind spot.

**Teams.**

## Why Individual Scores Aren't Enough

A team where every member scores 80+ isn't necessarily strong. If everyone is a Producer, nobody is shaping architecture. Nobody is paying down debt. The codebase ships fast and rots faster.

Conversely, a team averaging 50 points — but with one Architect, one Cleaner, and two Growing juniors — may be in a much healthier position. In six months, those juniors will level up, and the foundation is solid.

**A strong team is not the sum of individual scores. It's about composition and complementarity.**

## `eis team` — Team-Level Analysis

The new `eis team` command aggregates individual scores into team-level metrics.

```bash
# Simplest: domain = team (no config needed)
eis team --recursive ~/workspace

# With explicit team definitions
eis team --config eis.yaml --recursive ~/workspace

# JSON output (paste into AI for deeper analysis)
eis team --format json --recursive ~/workspace
```

If no `teams` section exists in config, each domain (Backend / Frontend / Infra) is treated as a single team. Zero config required.

```yaml
# eis.yaml (optional)
teams:
  backend-core:
    domain: Backend
    members: [alice, bob, charlie]
  frontend-app:
    domain: Frontend
    members: [dave, eve]
```

## Seven Team Health Axes

### 1. Complementarity — Role Diversity

How many of the 5 known roles are present? Architect gets the biggest bonus because design leadership is the most critical gap.

```
coverage = uniqueRoles / 5
bonus = Architect(+10) + Anchor(+5) + Cleaner(+5)
score = coverage × 80 + bonus  (clamped 0-100)
```

A team with only Producers and no Architect scores 16. A fully diverse team hits 100.

### 2. Growth Potential — Can Juniors Level Up Here?

Growing members + mentoring capacity.

```
score = growingRatio × 60 + Builder(+20) + Cleaner(+20)
```

Having Growing juniors is necessary but not sufficient. Without a Builder or Cleaner as a role model, growth stalls. Both must be present for the score to climb.

### 3. Sustainability — Inverse of Risk

What percentage of the team is in a risk state (Former, Silent, Fragile)?

```
riskRatio = (Former + Silent + Fragile) / memberCount
score = (1 - riskRatio) × 80 + Architect(+20)
```

Former members whose code still dominates blame. Silent members who haven't meaningfully contributed in months. Fragile members whose code survives only because nobody touches it. These are the hidden drags on team velocity.

### 4. Debt Balance — Self-Cleaning Tendency

Average Debt Cleanup score across members. 50 is neutral; above 50 means the team cleans more debt than it creates.

### 5. Productivity Density — Output Per Head

Average production score with a small-team bonus. Three people running a large-scale API server will show an abnormally high density — impressive, but also risky.

```
base = avg(members.Production)
bonus: ≤3 members && base ≥ 50 → ×1.2
        ≤5 members && base ≥ 50 → ×1.1
```

This quantifies the "this amount of code from this few people is insane" feeling.

### 6. Quality Consistency — Mean + Low Variance

A team averaging 80 quality with tight variance is healthy. A team averaging 80 but ranging from 95 to 40 is not — the low end drags reviews, and quality gates become performative.

```
score = avgQuality × 0.6 + (100 - stdev × 2) × 0.4
```

### 7. Risk Ratio — Blunt Truth

Percentage of members in Former, Silent, or Fragile state. Above 25% is a warning. Above 50% is a crisis.

## 5-Axis Team Classification — Reverse-Engineering Structure from Code (v0.10.0)

In Chapter 1, we introduced the 3-axis topology (Role / Style / State) for individual engineers. In v0.10.0, we **aggregate individual topologies into team-level classifications** across five axes.

![Team 5-Axis Classification Flow: Bottom-up structure discovery from Code → Engineer → Team](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/team-classification-flow.png)

The insight behind this: **we're reverse-engineering structure from code, through individual engineers.** Starting from raw git log and git blame, we read individual characteristics, then derive team structure bottom-up — code → engineer → team → organizational architecture.

### The Five Axes

| Axis | Derived from | Question |
|---|---|---|
| **Structure** | Member Role distribution | What structural roles exist on this team? |
| **Culture** | Member Style distribution | How does this team work? |
| **Phase** | Member State distribution | Where is this team in its lifecycle? |
| **Risk** | Health metrics | What risks does this team carry? |
| **Character** | Composite of above 4 | In one word, what kind of team is this? |

Character is a meta-classification synthesized from the other four — the team's "face" in a single label.

### Weighted Classification — Strong Members Paint the Team's Color

Classification uses **member Total score as influence weight**:

```
weight = member.Total / 100  (minimum 0.1)
```

An Architect scoring 90 and an Architect scoring 15 on the same team don't contribute equally to the team's character. The high-scorer shapes the team far more. Ethnographically, **strong performers with high output propagate more culture to the team**. The formula encodes this directly.

The minimum weight of 0.1 ensures that presence still matters — three Growing members at score 15 each still influence the team's Phase.

### Structure — From Role Distribution

| Label | Condition | Meaning |
|---|---|---|
| **Architectural Engine** | Architect+Anchor strong, AAR 0.3-0.8, high coverage | Design and quality engine firing on both cylinders |
| **Architectural Team** | Architect-heavy | Deep design bench |
| **Architecture-Heavy** | Architect-skewed, low Anchor | Design exists but implementation can't keep up |
| **Emerging Architecture** | Few Architects, mostly Anchor/Producer | Design culture is nascent |
| **Delivery Team** | Producer-dominant | Ship-focused |
| **Maintenance Team** | Cleaner/Anchor-dominant | Operations and stability focused |
| **Unstructured** | Mostly "—" roles | No clear structural identity |
| **Balanced** | No dominant pattern | Evenly distributed |

**AAR (Architect-to-Anchor Ratio)** is the key structural metric. Too many Architects and design outpaces implementation. Too many Anchors and stability dominates at the expense of innovation. The healthy range is 0.3–0.8.

### Culture — From Style Distribution

| Label | Dominant Styles | Meaning |
|---|---|---|
| **Builder** | Builder-heavy | Build-and-ship culture |
| **Stability** | Balanced/Resilient | Conservative, stability-oriented |
| **Mass Production** | Mass-heavy | Volume over durability |
| **Firefighting** | Churn/Rescue | Constant emergency response |
| **Exploration** | Spread-heavy | Wide exploration, thin depth |
| **Mixed** | No dominant pattern | Blended |

### Phase — From State Distribution

| Label | Dominant States | Meaning |
|---|---|---|
| **Emerging** | Growing-heavy | Growth phase |
| **Scaling** | Active + Growing | Expansion phase |
| **Mature** | Active-dominant | Mature, productive team |
| **Stable** | Active + Balanced | Steady state |
| **Declining** | Former/Silent | Talent attrition |
| **Rebuilding** | Active + Former mix | Rebuilding after departures |

### Risk — From Health Metrics

| Label | Condition | Meaning |
|---|---|---|
| **Design Vacuum** | Low Complementarity | No design leadership |
| **Talent Drain** | High Risk Ratio | Losing effective contributors |
| **Debt Spiral** | Low Debt Balance | Accumulating technical debt |
| **Quality Erosion** | Low Quality Consistency | Quality is degrading |
| **Healthy** | None of the above | Clean bill of health |

### Character — The Team's Identity

Synthesized from Structure × Culture × Phase × Risk plus structural metrics (AAR, Anchor Density, Productivity Density).

| Character | Key Conditions | Meaning |
|---|---|---|
| **Elite** | High SC, healthy AAR, high PD | Design strength meets production velocity |
| **Fortress** | Good Structure, stable Culture | Robust defensive team |
| **Pioneer** | Growth Phase, Builder Culture | Trailblazing into new territory |
| **Academy** | Growing members + Builder present | Active talent development |
| **Feature Factory** | Producer-dominant, no Architect | Ships features but design is adrift |
| **Guardian** | Anchor/Cleaner-dominant | Quality and maintenance guardians |
| **Firefighting** | Churn/Rescue culture | Perpetually fighting fires |

**SC (Structure-Culture complementarity)** measures how well Structure and Culture mesh. Architectural Engine + Builder Culture is the best combination. Delivery Team + Firefighting Culture is the worst.

### Structural Metrics

Three additional metrics quantify the team's structural skeleton:

**AAR (Architect-to-Anchor Ratio)**: Architect count ÷ Anchor count. Healthy range: 0.3–0.8. Too high = design overload (implementation can't keep up). Too low = stability saturation (no design innovation). Architects present with zero Anchors signals Architect Isolation.

**Anchor Density**: Anchors ÷ active members. How thick is the quality and stability foundation?

**Architecture Coverage**: (Architects + Anchors) ÷ total team. What percentage of the team is involved in design and quality?

These metrics reveal **structural quality** that Role distribution alone cannot show — the difference between a team that *has* architects and a team where architecture *actually works*.

## Patterns of Strong Teams

After running this on multiple teams — now with 5-axis classification and structural metrics — patterns emerge:

**Strong teams share:**

- Architect + Builder present (someone designs, someone implements design)
- 3+ role types (minimum: Architect / Anchor / Producer)
- 20%+ Growing ratio (juniors are developing)
- Risk Ratio near 0%
- Quality Consistency above 70

**Dangerous compositions:**

- Mass/Churn-heavy: high volume, low durability
- No Architect: nobody shapes the design layer → implicit decisions accumulate
- Silent accumulation: headcount says 8, effective contributors are 4
- Producer monoculture: everyone builds, nobody cleans

## Growth Model — Climbing Three Layers

EIS's Role classification maps to three layers of engineering growth.

```
  Design Layer        Architect
                      ↑ Design decisions are embedded in code as structure
  ──────────────────────────────────
  Stabilization Layer Anchor / Cleaner
                      ↑ Quality rises, code survives, others' debt gets cleaned
  ──────────────────────────────────
  Implementation Layer Producer / Growing
                       Write code. Ship it.
```

**Implementation Layer**: Write and ship code. Growing engineers start here. Production exists but Survival is still low.

**Stabilization Layer**: Quality improves, code starts surviving. You can fix other people's code too. Anchors and Cleaners live here.

**Design Layer**: You touch architecture files and shape the structure. Architects live here.

Growth means climbing these layers. EIS score trajectories make this observable:

- Survival rising → moving from Implementation to Stabilization
- Design rising → moving from Stabilization to Design
- DebtCleanup rising → expanding team contribution

In team context, **teams with high Growth Potential have environments where this climb is possible**. An Anchor at the Stabilization layer, an Architect at the Design layer. Role models exist, so Growing members can reach the next layer. Without role models, they keep spinning at Implementation.

And there's a downward direction too.

```
  Design Layer        → Former (design knowledge leaves with the person)
  Stabilization Layer → Silent (contributions stop), Fragile (only untouched code remains)
  Implementation Layer → Silent (stops writing)
```

Helping members climb the layers, and catching early when someone is falling into a Risk state. **That's the management job EIS makes visible.** Track score trajectories quarter over quarter, and you can see who's climbing, who's plateaued, and who's slipping — in numbers.

## A Sociological Observation

**Teams with a Builder or Cleaner grow people faster.** When a role model exists — someone whose code demonstrably survives and whose reviews teach — Growing members transition to Active at roughly double the rate.

**Teams without an Architect degrade over time.** Low Complementarity correlates with declining Quality Consistency over 6-month windows. Without a design compass, everyone codes in their own direction.

**Small-team anomalies are two-sided.** High Productivity Density is both a strength and a risk. If one person leaves a 3-person team scoring 80+ density, the impact is catastrophic.

## Member Tiers & Warnings (v0.10.3)

In v0.10.3, `eis team` introduces **three-tier member classification** and **automatic warnings**.

### Core / Risk / Peripheral

Not every git author is a "team member." A drive-by contributor who touched one file shouldn't dilute your team's health metrics. EIS now splits members into three tiers:

| Tier | Condition | Used for |
|---|---|---|
| **Core** | `RecentlyActive && Total >= 20` | Averages, ProductivityDensity, QualityConsistency |
| **Risk** | State in {Former, Silent, Fragile} | Distributions, RiskRatio, Classification |
| **Peripheral** | Everyone else | TotalMemberCount only |

This prevents cross-functional helpers from diluting metrics while keeping risk states visible. The header now shows: `4 core + 3 risk / 16 total`.

### Automatic Warnings

EIS detects dangerous metric combinations and surfaces them as plain-text warnings:

```
⚠ Warnings:
  43% risk ratio — 3 of 7 effective members are Former/Silent/Fragile
  Top contributor (machuz) accounts for 46% of core production — ProdDensity drops to 39 without them
  2 Silent members — headcount says 16 but effective contributors are 4
  Fragile gravity — okatechnology (Grav 68) has high influence but low robust survival (8)
```

Warning types:

- **Bus factor risk**: few core members carrying many repos
- **Risk ratio**: percentage of inactive/at-risk members
- **Top contributor concentration**: what happens if they leave (simulated ProdDensity drop)
- **Silent accumulation**: headcount vs. effective contributor gap
- **Gravity warnings**: fragile influence centers, low structural coverage despite Architect presence

### Phase Refinement

The Phase axis now distinguishes between truly declining teams and strong teams carrying historical weight:

| Label | Condition | Meaning |
|---|---|---|
| **Legacy-Heavy** | Risk high, but AvgTotal ≥ 40 + Architect present | Strong team with accumulated history |
| **Mature with Attrition** | Moderate risk (20-40%), active core still strong | Natural attrition from mature team |
| **Declining** | Risk high, weak core | Genuine talent drain |

A Backend team with an Architect scoring 90+ and two Silent former members isn't "Declining" — it's **Legacy-Heavy**. The distinction matters for planning.

## Real-World Results — Our Team

Running `eis team` on our actual product (12 Backend repos + 9 Frontend repos):

**Backend — Elite / Legacy-Heavy**:

- 4 core members carrying 12 repos, 3 risk members (2 Silent + 1 Former)
- Architect + 2 Anchors = AAR 0.50 (healthy range)
- ProdDensity 60 — decent for 4 people, but top contributor accounts for 46% of production
- `Legacy-Heavy` phase: not declining, but the historical weight is real

**Frontend — Pioneer / Mature**:

- 6 core members, 0 risk — everyone is active
- Architect + Anchor present, structural coverage 33%
- Sustain 100/100, RiskRatio 0% — clean bill of health
- Gravity warning: one member has high structural influence with low robust survival

The numbers tell a story:

- **Backend**: Strong but carrying historical weight. An Elite team by character, but fragile — one departure changes everything.
- **Frontend**: A Mature Pioneer. The Architect is functioning, Risk is 0%. One Gravity warning remains, but the team is structurally healthy.

**The numbers started to tell a story.** Not just "who is strong" but "what state is the team in, and what happens next."

## Good Design Creates Common Sense

The reason our Backend team is classified as Legacy-Heavy is clear: a former architect with enormous output left the team. The sheer volume of what they built means several modules remain that only they had touched. A large chunk of `git blame` points to a Former member.

And yet, the team hasn't collapsed.

Why? Because those modules were built on well-organized design. We received verbal handoffs, but there was no comprehensive documentation or complete knowledge transfer. Still, the design embedded in the code's structure gave the remaining engineers enough understanding to operate confidently.

**Strong design leaves knowledge in structure, not in people.** And that structure creates shared understanding across the team.

I think of this as "good design creates common sense." Great design doesn't necessarily require documentation or complete knowledge transfer. The code's structure itself communicates the module's intent and usage.

EIS currently measures quantitative signals — history, code survival rates, structural influence. This "common sense through design" — why a Former member's code still runs healthily — isn't directly observable yet.

But if it becomes possible, we could move beyond simple Legacy-Heavy warnings and distinguish between **"historically heavy but structurally sound"** and **"genuinely dangerous dependency structures."**

In practice, though, we may not need to measure that directly. A strong team will gradually replace Former members' code with their own, and Legacy-Heavy resolves itself over time. It converges toward where it should be. EIS can naturally capture that convergence through Survival trajectories and Risk Ratio changes.

## How to Use It

![eis team terminal output](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/team-output.png)

```bash
# Install
brew tap machuz/tap && brew install eis

# Team analysis
eis team --recursive ~/workspace

# JSON → paste into Claude / ChatGPT
eis team --format json --recursive ~/workspace | pbcopy
```

Deep insights are intentionally out of scope. The tool produces quantitative data; humans (or AI) interpret it. This separation is by design.

## Summary — Individual × Team

Chapter 1 answers "What kind of engineer is this person?"
Chapter 2 answers "What state is this team in?"

Together, they enable:

- **Hiring**: see which Role is missing → define the position
- **Team formation**: maximize complementarity
- **1-on-1s**: discuss growth direction based on score trajectories
- **Risk management**: catch Risk Ratio deterioration early

All from git history. No surveys. No additional tooling.

**What you can measure, you can improve. What you can't measure, you can only pray about.**

Let's turn team strength from a prayer into a metric.

---

**GitHub:** [machuz/engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLI tool, full methodology, and blog posts. Install with `brew tap machuz/tap && brew install eis`.

**Support:** If this helped you see your team differently — [❤️ Sponsor on GitHub](https://github.com/sponsors/machuz) / PayPay: `w_machu7`

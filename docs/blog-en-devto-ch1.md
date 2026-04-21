---
title: "Git Archaeology #1 — Measuring Engineering Impact from Git History Alone"
series: "Git Archaeology"
published: true
description: A 7-axis observation model that quantifies engineer impact using nothing but git log and git blame. Code survival, debt cleanup, bus factor — all from data you already have.
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/cover-ch1.png?v=4
---

*Why commit counts, PR counts, and lines of code fail to capture real engineering strength*

![7-axis signal visualization](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch1-iconic.png?v=4)

## From Scores to Signals

I never felt comfortable with scores.

I wanted to visualize things like: the effort someone puts in, the ability to create structures that last, the impact on a team. But these are influenced by too many variables — the state of the codebase, the direction of the project, team dynamics, even personal circumstances.

**There is no way to reduce all of this into a single, absolute score from code alone.**

However, I realized something important. While absolute scoring is impossible, we *can* observe **changes in strength.**

At one point, a certain engineer is driving the system forward. At another moment, someone else takes that role. This is not a score. **This is a signal.**

You can't assign a fixed value to a human or a system. But you can observe: where momentum exists, where pressure is building, where stability is forming or breaking.

Scores try to evaluate. **Signals reveal.**

I jokingly call it an engineer's **"combat power."** The formal name is **Engineering Impact Signal** — **EIS**, pronounced *"ace."* But what it actually measures is something more precise:

> **observable structural signals recorded in the codebase itself**

---

## The Core Idea: Code That Survives Matters Most

The strongest engineers don't just write code. They write code that **continues to exist months later** without needing to be rewritten.

So the most important signal in this model is **Code Survival**.

But even survival must be handled carefully. Raw git blame favors early contributors. To fix this, the model applies **time-decayed survival** — recent code counts far more than ancient code.

| Age | Weight |
|---|---|
| 7 days | 0.96 |
| 30 days | 0.85 |
| 90 days | 0.61 |
| 180 days | 0.37 |
| 1 year | 0.13 |
| 2 years | 0.02 |

![Time-decayed Survival Weight curve showing exponential decay over 730 days](https://raw.githubusercontent.com/machuz/eis/main/docs/images/survival-decay-curve.png?v=5)

This means departed team members' signals naturally decay over time. It approximates **who is currently writing durable code**, not who wrote the most code historically.

### Dormant vs Robust — The Key Distinction

Code "surviving" can mean two very different things:

- **Dormant Survival**: code remains in an untouched module. Not durable — just undisturbed
- **Robust Survival**: code remains in **files where other engineers are actively making changes**. Only code that survives under real change pressure is counted

This distinction is EIS's most important innovation. An engineer with low overall Survival but decent Robust Survival is iterating heavily while producing change-resistant code (**Resilient** style). Conversely, high Survival but low Robust Survival signals code that survives only because nobody touches it (**Fragile** state).

**Time-decayed survival resists gaming.** You can't inflate your impact with busy work — only code that remains in the codebase months later counts. And the debt cleanup axis makes it structurally impossible to achieve high impact by generating work for others.

---

## 7 Axes of Engineering Impact

![EIS Framework Overview](https://raw.githubusercontent.com/machuz/eis/main/docs/images/engineering-impact-framework-diagram-fixed.png?v=5)
*Git history flows into 7-axis signals, 3-axis topology (Role/Style/State), and Gravity.*

| Axis | Weight | What it captures |
|---|---|---|
| Production | 15% | Changes per day (absolute scale) |
| Quality | 10% | Low rate of fix/revert commits |
| Survival | **25%** | Code that still exists today (time-decayed) |
| Design | 20% | Contributions to architecture files |
| Breadth | 10% | Number of repositories touched |
| Debt Cleanup | 15% | Fixing issues created by others |
| Indispensability | 5% | Bus-factor risk |

**Survival gets the highest weight (25%)** — it's the core thesis: *are you writing designs that last?*

Two axes deserve special attention:

**Debt Cleanup** — when I added this metric, the "silent hero" on my team became visible. Someone who quietly fixed everyone else's bugs, all the time. This is exactly the kind of person that traditional metrics make invisible.

**Design** — frequent commits to architecture files (repository interfaces, domain services, routers, middleware) signal architectural involvement. Not whether a decision was *correct*, but **who participates in shaping the structure**. These patterns must be customized per project — the configuration step itself becomes a design conversation.

> For detailed formulas and normalization: [Whitepaper](https://github.com/machuz/eis)

---

## Impact and the Observation Model

Signals are weighted into a single **Impact** value:

```
impact =
  production       × 0.15
  + quality        × 0.10
  + survival       × 0.25
  + design         × 0.20
  + breadth        × 0.10
  + debt_cleanup   × 0.15
  + indispensability × 0.05
```

The scale is intentionally strict:

| Impact | Assessment |
|---|---|
| 80+ | Supernova. 1–2 per team at most |
| 60–79 | Near-core. Strong |
| **40–59** | **Senior-level. 40+ is genuinely strong** |
| 30–39 | Mid-level |
| 20–29 | Junior–Mid |
| <20 | Junior |

**40 = Senior.** Putting up decent numbers across seven axes simultaneously requires serious, well-rounded ability. An engineer in the 40s can compete in any market.

**Critical caveat:** EIS measures **impact on *this* codebase**, not absolute engineering ability. High Survival might even mean the code can't be refactored away. If the observations don't match your gut feeling, that's worth investigating — it may reveal codebase design issues rather than people issues. (We call this [Engineering Relativity](https://dev.to/machuz/git-archaeology-8-engineering-relativity-why-the-same-engineer-gets-different-scores-5dnl).)

![Impact Guide](https://raw.githubusercontent.com/machuz/eis/main/docs/images/score-guide.png?v=5)

---

## 3-Axis Topology

Once signals are calculated, recognizable patterns emerge. EIS decomposes engineer topology into **three independent axes**:

![Engineer Topology](https://raw.githubusercontent.com/machuz/eis/main/docs/images/engineering-archetypes-paper-figure.png?v=5)

### Role — *What* they contribute

| Role | Signal | Description |
|---|---|---|
| **Architect** | Design↑ Surv↑ | Shapes system structure. Code survives because the design is sound |
| **Anchor** | Qual↑ Surv↑ Debt↑ | Stabilizes the codebase. Writes durable code and quietly cleans up others' bugs |
| **Cleaner** | Debt↑ | Primarily fixes debt generated by others |
| **Producer** | Prod↑ | High output. Whether that output is *good* depends on Style and State |
| **Specialist** | Indisp↑ Breadth↓ | Deep expertise in a narrow area |

### Style — *How* they contribute

| Style | Signal | Description |
|---|---|---|
| **Builder** | Prod↑ Surv↑ Design↑ | Designs, builds heavily, AND maintains |
| **Resilient** | Prod↑ RobustSurv○ | Iterates heavily, but what survives under change pressure is durable |
| **Rescue** | Prod↑ Surv↓ Debt↑ | Taking over and cleaning up legacy code |
| **Churn** | Prod○ Qual↓ Surv↓ | Production exists but survival does not |
| **Mass** | Prod↑ Surv↓ | Writes a lot, nothing survives |
| **Balanced** | Even distribution | Well-rounded |
| **Spread** | Breadth↑ Prod↓ Surv↓ | Wide presence, zero depth |

### State — *Lifecycle phase*

| State | Signal | Description |
|---|---|---|
| **Active** | Recent commits, Surv↑ | Currently writing durable code |
| **Growing** | Qual↑ Prod↓ | Low output but high quality. Leveling up |
| **Former** | Raw Surv↑ Surv↓ | Code persists but author is inactive. Handoff priority |
| **Silent** | Prod↓ Surv↓ Debt↓ | All signals low. May indicate role mismatch or environment that hasn't activated this person's strengths |
| **Fragile** | Surv↑ Prod↓ Qual<70 | Code survives only because nobody changes it |

> Deep dive into topology evolution: [Chapter 3](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga) and [Chapter 4](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d)

---

## Engineer Gravity

Impact tells you *how strong*. Topology tells you *what kind*. **Gravity** tells you how much structural influence this person exerts.

```
Gravity = Indispensability × 0.40 + Breadth × 0.30 + Design × 0.30
```

High Gravity isn't automatically good — it has a **health dimension**:

```
health = Quality × 0.6 + RobustSurvival × 0.4

Gravity < 20  → dim gray  (low influence)
health ≥ 60   → green     (healthy gravity)
health ≥ 40   → yellow    (moderate)
health < 40   → red       (fragile gravity — dangerous)
```

Red gravity means **"the system depends on this person AND the code is fragile."** The most dangerous combination, instantly visible.

---

## Real-World Results

I ran this on my own team (14 repos, 10+ engineers). The signals matched the team's gut feeling almost perfectly.

![Backend Rankings](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch1-backend-table.png?v=4)

**R.S.** — Production 17 doesn't turn heads. But Survival 50 (2nd on the team) means their recent code stays. Debt Cleanup 88 means they're quietly fixing everyone else's bugs. **This is exactly the kind of person that Debt Cleanup was designed to surface.** The Anchor role captures this perfectly.

**Y.Y.** — Design 67, Breadth 81. The original architect. Indispensability 100 — more modules are still attributed to this person than to any active member. Topology reads `Architect / — / Former` — the Role persists in the code even after departure. This is the signal that [Chapter 4](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d) calls "a soul that needs to be laid to rest" — a handoff priority.

**Z.** — Impact 24.9. Breadth was the only elevated signal. Topology reads `— / Spread / —` — wide presence but no depth. An early observation of this pattern could have informed better role alignment.

---

## What Existing Metrics Miss

DORA measures deployment velocity. SPACE uses surveys. Git analytics tools track *when* code was written. **None of them ask whether the code actually survived.**

EIS fills that gap: time-decayed survival + robust/dormant separation + debt cleanup tracking. From data you already have.

---

## Limitations and Honesty

This model is **not** a measure of human worth. It estimates technical influence observable in a codebase.

Engineers contribute in ways git cannot capture: mentoring, domain expertise, documentation, psychological safety. EIS captures what git records — nothing more.

Low impact doesn't mean a weak engineer. Ambiguous specs, organizational friction, and poor planning all reduce signals. If the entire team shows low impact, examine the organization before examining individuals.

The model's accuracy scales with codebase design quality. In chaotic codebases, high Survival might just mean dead code. **The metric's low accuracy is itself a signal.**

---

## Try It

```bash
❯ brew tap machuz/tap && brew install eis
❯ eis analyze --recursive ~/projects
```

Zero AI tokens. Zero API keys. Just `git log` and `git blame`.

![Terminal Output](https://raw.githubusercontent.com/machuz/eis/main/docs/images/terminal-output.png?v=0.11.0)

The real value comes from **tracking changes over time**. If Survival rises quarter-over-quarter, design skills are growing. If Debt Cleanup rises, team contribution is increasing.

> Full methodology: [Whitepaper](https://github.com/machuz/eis) · [README](https://github.com/machuz/eis)

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png?v=2)

**GitHub**: [eis](https://github.com/machuz/eis) — CLI tool, formulas, and methodology all open source. `brew tap machuz/tap && brew install eis` to install.


If this was useful: [Sponsor on GitHub](https://github.com/sponsors/machuz)

---


[Chapter 2: Team Health →](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f)

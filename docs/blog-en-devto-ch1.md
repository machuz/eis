---
title: "Git Archaeology #1 — Measuring Engineering Impact from Git History Alone"
published: true
description: A 7-axis observation model that quantifies engineer impact using nothing but git log and git blame. Code survival, debt cleanup, bus factor — all from data you already have.
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/cover-ch1.png?v=4
---

*Why commit counts, PR counts, and lines of code fail to capture real engineering strength*

![7-axis score visualization](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch1-iconic.png?v=4)

I lead backend and infrastructure for a mid-sized product team. As the team grew, something kept nagging at me: **how do you quantify how strong an engineer actually is?**

Commit count? Lines of code? Number of PRs? All of these are easy to measure — and easy to misinterpret.

A typo fix and a system-wide architectural change both count as "one PR." A generated lockfile can add thousands of lines. Commit habits vary wildly between engineers.

Yet inside every team, people still have a sense of who the strongest engineers are.

> "This person writes code that lasts."
> "That person touches everything but somehow nothing improves."

Those intuitions exist, but they are rarely measurable. In salary negotiations, "I just feel like they're strong" doesn't hold up. My hiring instincts have been solid — I've never made a bad hire — but after someone joins, I wanted a way to **quantitatively track how much they actually deliver**.

I also wanted proof of something I already felt: that my current team is one of the strongest I've ever worked with.

The trigger was simple: **I knew this team was strong. But I had no words to explain it.**

I wanted to say "this team is seriously good" — but I had no evidence. Not volume of voice, not politics, but **facts recorded in code**. That's what I wanted to speak with.

So one evening, drinking and pair-programming with Claude Code, I built an observation model that uses **nothing but git history** — and the results matched my gut feeling with eerie accuracy.

I jokingly call it an engineer's **"combat power."** The formal name is **Engineering Impact Signal** — **EIS**, pronounced *"ace."* But what it actually measures is something more precise:

> **observable technical impact recorded in the codebase itself**

---

## The Core Idea: Code That Survives Matters Most

The strongest engineers don't just write code.

They write code that **continues to exist months later** without needing to be rewritten.

So the most important signal in this model is **Code Survival**.

But even survival must be handled carefully. Raw git blame favors early contributors. Someone who wrote a lot of code three years ago may dominate blame history even if they haven't contributed since.

To fix this, the model applies **time-decayed survival**. Recent code counts far more than ancient code.

| Age | Weight |
|---|---|
| 7 days | 0.96 |
| 30 days | 0.85 |
| 90 days | 0.61 |
| 180 days | 0.37 |
| 1 year | 0.13 |
| 2 years | 0.02 |

![Time-decayed Survival Weight curve showing exponential decay over 730 days](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/survival-decay-curve.png?v=5)
*Time-decayed survival gives much more weight to recently written code than legacy code that simply remains untouched.*

This means departed team members' signals naturally decay over time — solving the problem of someone who wrote a ton of code during the founding era dominating the leaderboard forever.

It approximates **who is currently writing durable code**, not who wrote the most code historically.

---

## The 7 Axes of Engineering Impact

![EIS Framework Overview: Git history flows into 7-axis scores, 3-axis topology (Role/Style/State), and Gravity](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/engineering-impact-framework-diagram-fixed.png?v=5)
*The Engineering Impact Signal aggregates seven observable signals derived from git history into impact, topology, and structural gravity.*

| Axis | Weight | What it captures |
|---|---|---|
| Production | 15% | Changes per day (absolute: configurable `production_daily_ref`, default 1000) |
| Quality | 10% | Low rate of fix/revert commits |
| Survival | **25%** | Code that still exists today (time-decayed) |
| Design | 20% | Contributions to architecture files |
| Breadth | 10% | Number of repositories touched |
| Debt Cleanup | 15% | Fixing issues created by others |
| Indispensability | 5% | Bus-factor risk |

I started with 5 axes, but real-world measurement revealed two blind spots: "the person who quietly cleans up everyone else's bugs" and "the person whose departure would kill the project." That's why Debt Cleanup and Indispensability were added.

**Survival gets the highest weight (25%)** because it's the core thesis of this model: *are you still writing designs that last?*

Quality is weighted low (10%) because commit-message-based detection is a rough proxy. Indispensability is low (5%) because it conflates "strong, therefore needed" with "nobody bothered to take over." The weight design is intentional.

**BE / FE / Infra are scored separately.** Without this separation, backend code volume contaminates frontend rankings and vice versa.

---

## Production: Changes Per Day, Absolute Scale

Commit counts are unreliable. Some engineers make large commits. Others split work into many small ones. An engineer who changes 100 lines in one commit and one who makes 100 single-line commits shouldn't be profiled the same.

We measure `insertions + deletions` per day, using **absolute profiling**:

```
production_score = min(changes_per_day / production_daily_ref × 100, 100)
```

The daily rate is computed from total changes divided by the span between the author's first and last commit. The reference (`production_daily_ref`, default 1000) is configurable.

Why absolute instead of relative? With relative profiling, **a single-person domain always gives that person a signal of 100**, even if they're slow. Absolute profiling makes production comparable across organizations and domains.

Auto-generated files are excluded:

- `package-lock.json`, `yarn.lock` — library updates move tens of thousands of lines
- `docs/swagger*` — auto-generated
- `mock_*`, `*.gen.*` — code generation

---

## Quality: The Fix Ratio

```
quality    = 100 - fix_ratio
fix_ratio  = fix_commits / total_commits × 100
```

Fix commits are detected using patterns like `fix`, `revert`, `hotfix`, plus language-specific keywords (e.g. `修正` in Japanese teams — if you don't catch these, accuracy drops for non-English teams).

A high fix ratio usually means **code that needs frequent correction after being written**. It's a reasonable proxy for first-pass correctness.

But large refactors and proactive improvements also trigger "fix" patterns, which is why **Quality is intentionally weighted at only 10%**. Interpret it alongside Design — if someone has low Quality but high Design, the "fixes" are likely aggressive architectural improvements, which is healthy.

---

## Survival: The Most Important Metric

This is the heart of the model.

Naive git blame gives high signals to "someone who wrote a lot of code three years ago and hasn't done anything since." That's wrong. What we want to know is: **are you actively writing good designs right now?**

![Survival Calculation](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch1-code-survival.png?v=4)

Each surviving line contributes a decayed weight to its author. Engineers whose code remains stable accumulate high survival signals. Engineers whose code is constantly rewritten do not.

Raw blame (without decay) is still useful separately — it shows "who built the foundation of this codebase." But for the combat power impact, only time-decayed survival is used.

### Dormant vs Robust — Separating What "Survived" Actually Means

Code "surviving" can mean two very different things:

- **Dormant Survival**: code remains in an untouched module. Not durable — just undisturbed
- **Robust Survival**: code remains in **files where other engineers are actively making changes**. Only code that survives under real change pressure is counted

Standard Survival doesn't distinguish between the two. EIS separates them.

This distinction drives the Style classification. An engineer with low overall Survival but decent Robust Survival is iterating heavily while producing change-resistant code (**Resilient** style). Conversely, high Survival but low Robust Survival suggests code that survives only because nobody touches it (**Fragile** state).

---

## Design Influence

Architectural work tends to appear in specific areas of a codebase.

**Backend:** repository interfaces, domain services, routers, middleware, dependency injection

**Frontend:** core logic, shared stores, hooks, type systems

Frequent commits in these areas signal **architectural involvement**. It's not measuring whether a design decision was *correct* — it's measuring **who participates in shaping the system structure**. Someone who never touches architecture files is unlikely to be making good design decisions. As an approximation, it works surprisingly well.

**These patterns must be customized per project.** The defaults assume Clean Architecture / DDD directory structures. If your architecture lives in Protobuf definitions, Terraform configs, or migration files, add those patterns — otherwise the Design axis will be blind to your actual architecture.

There's a meta-insight here: **defining which files are "architecture files" is itself an act of articulating your team's design philosophy.** Asking "where do design decisions manifest in our codebase?" forces you to clarify your team's architectural boundaries. The configuration step becomes a design conversation.

---

## Breadth: Operating Across the System

How many repositories does an engineer contribute to? Simple, but it cleanly separates engineers who only operate in their own silo from those who understand the system holistically.

---

## Debt Cleanup: The Quiet Heroes

One of the most revealing metrics. When a fix commit modifies code, we check **who originally wrote those lines**.

![Debt Cleanup Tracking](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch1-code-debt.png?v=4)

The moment I added this metric, the "silent hero" on my team became visible — someone who quietly fixed everyone else's bugs, all the time. Conversely, the "high-output engineer who generates fix work for everyone around them" also became impossible to ignore.

**Note:** Members with fewer than 10 total debt events (generated + cleaned) get a neutral signal of 50, because small samples produce extreme ratios.

---

## Indispensability: Bus Factor Risk

![Indispensability Detection](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch1-code-indispensability.png?v=4)

If one engineer owns more than 80% of the lines in a module, that module becomes a **bus factor risk**. High indispensability means both expertise *and* organizational fragility — which is why the weight is only 5%. It's as much a **handoff priority alert** as it is an evaluation metric.

---

## Observation Model

Metrics use a **hybrid approach**:

**Absolute axes** (comparable across organizations):

- Production: `min(changes_per_day / production_daily_ref * 100, 100)`
- Quality: `100 - fix_ratio` (directly on 0-100 scale)
- Debt Cleanup: bounded 0-100 scale

**Relative axes** (normalized within domain):

- Survival, Design, Breadth, Indispensability: top person gets 100

Profiled per domain (Backend / Frontend / Infra / Firmware separately). Domain is auto-detected from file extensions or configured explicitly.

```
score =
  production       × 0.15
  + quality        × 0.10
  + survival       × 0.25
  + design         × 0.20
  + breadth        × 0.10
  + debt_cleanup   × 0.15
  + indispensability × 0.05
```

The scale is intentionally strict.

| Impact | Assessment | Approx. Total Comp (USD) |
|---|---|---|
| 80+ | Supernova. 1–2 per team at most | $250K–400K+ |
| 60–79 | Near-core. Strong | $180K–300K |
| **40–59** | **Senior-level. 40+ is genuinely strong** | **$140K–220K** |
| 30–39 | Mid-level | $100K–160K |
| 20–29 | Junior–Mid | $80K–120K |
| <20 | Junior | $60K–90K |

Comp figures are rough estimates and vary significantly by market (SF vs. Midwest, US vs. Europe, etc.).

**40 = Senior.** If that seems low, consider what it takes: with relative observation across 7 axes, just putting up decent numbers across the board requires serious, well-rounded ability. Production, quality, survival, design, breadth, debt cleanup — doing well in all of them simultaneously is structurally difficult. If your senior hits 40, that's *normal*. An engineer in the 40s can compete in any market.

**One critical caveat.** EIS measures **impact on *this* codebase**, not absolute engineering ability. High impact means "on this codebase, this person's code is surviving, shaping architecture, and cleaning up debt." It does *not* mean they are a better engineer than someone with lower impact. High Survival might even mean the code can't be refactored away because the design is poor — not that the code itself is good. If the observations don't match your gut feeling, that's a signal worth investigating: it may reveal codebase design issues rather than people issues. The real question is whether someone can maintain their impact under *good* design — that's where true ability shows.

![Impact Guide](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/score-guide.png?v=5)

---

## Patterns That Emerge: The 3-Axis Topology

Once signals are calculated, recognizable patterns appear in the 7-axis distribution. In earlier versions of this model, we assigned a single "archetype" label — Architect-Builder, Mass Producer, Silent Killer, etc. It was intuitive, but it had a fundamental problem: **real engineers don't fit neatly into one box.**

An engineer can be an Architect in *role*, a Builder in *style*, and Active in *state* — all at the same time. Cramming that into "Architect-Builder (0.85)" loses information. Worse, a single label creates false equivalences: two "Producers" might have completely different styles and lifecycle phases.

Starting in v0.9.0, the model decomposes engineer topology into **three independent axes**:

![Engineer Topology in Production–Survival Space: 11 archetypes plotted across four quadrants](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/engineering-archetypes-paper-figure.png?v=5)
*Different engineer profiles emerge naturally when production and time-decayed survival are plotted together.*

### Axis 1: Role — *What* they contribute

Role captures the engineer's **primary contribution type** based on their 7-axis signal distribution.

| Role | Signal | Description |
|---|---|---|
| **Architect** | Design↑ Surv↑ | Shapes system structure. Code survives because the design is sound |
| **Anchor** | Qual↑ Surv↑ Debt↑ | Stabilizes the codebase. Writes durable code and quietly cleans up others' bugs |
| **Cleaner** | Debt↑ | Primarily fixes debt generated by others. The quiet hero |
| **Producer** | Prod↑ | High output. Whether that output is *good* depends on Style and State |
| **Specialist** | Indisp↑ Breadth↓ | Deep expertise in a narrow area. Bus factor risk, but valuable |
| **—** | No dominant signal | No single role dominates. Often combined with Balanced style |

### Axis 2: Style — *How* they contribute

Style captures the engineer's **working pattern** — independent of what they produce.

| Style | Signal | Description |
|---|---|---|
| **Builder** | Prod↑ Surv↑ Design↑ | The full package. Designs, builds heavily, AND maintains. If this person leaves, the product stalls |
| **Resilient** | Prod↑ RobustSurv○ | Iterates heavily — writes, rewrites, experiments — but what survives under change pressure is durable |
| **Rescue** | Prod↑ Surv↓ Debt↑ | Actively taking over and cleaning up legacy code. Low survival isn't from writing bad code but from rewriting inherited debt |
| **Churn** | Prod○ Qual↓ Surv↓ gap≥30 | A constant stream of rework. Most commits are fixes or reverts. Producing *churn*, not value |
| **Mass** | Prod↑ Surv↓ | Writes a lot, nothing survives |
| **Balanced** | Even distribution | No extreme peaks or valleys. Well-rounded |
| **Spread** | Breadth↑ Prod↓ Surv↓ | Wide presence across repos, zero depth. Shows up everywhere, produces little |
| **—** | No dominant pattern | No clear working style detected |

### Axis 3: State — *Lifecycle phase*

State captures **where the engineer is in their trajectory** relative to the codebase.

| State | Signal | Description |
|---|---|---|
| **Active** | Recent commits, Surv↑ | Currently writing durable code. Healthy and engaged |
| **Growing** | Qual↑ Prod↓ Design↓ | Low output but high quality. Writing carefully, learning the ropes. If production and design increase over time, this person is leveling up |
| **Former** | Raw Surv↑ Surv↓ | Code still exists (high raw survival) but the author is no longer active (low decayed survival). A **handoff priority alert** — modules they built need knowledge transfer |
| **Silent** | Prod↓ Surv↓ Debt↓ | All observed signals are low — no production, no survival, no debt cleanup. This pattern is easy to miss because it produces no visible problems. May indicate a role mismatch, a need for different project assignment, or an environment that hasn't activated this person's strengths |
| **Fragile** | Surv↑ Prod↓ Qual<70 | Code survives not because it's well-written, but because nobody is changing it. A hidden risk that will collapse under change pressure |
| **—** | No lifecycle signal | No clear state detected |

### Why This Matters

**Churn, Mass, Spread, and Silent patterns show low impact overall but can look impressive on individual signals (or fly under the radar entirely).** Organizations that evaluate on production alone or breadth alone will miss these multi-axis patterns — and may not align recognition with the types of contribution that sustain the system. Only multi-axis observation reveals the full picture. Rescue style is a notable exception — low survival looks alarming, but high debt cleanup reveals active legacy rescue rather than new debt generation. Resilient style is another positive exception — low total survival resembles Mass, but decent robust survival reveals iteration toward durable code. Fragile state is a subtle case — high survival looks reassuring, but combined with low production and mediocre quality, it signals dormant code that will break under change pressure.

![Topology Radar Chart — 9 of 288 theoretical combinations (Role×Style×State = 6×8×6)](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/archetypes-radar.png?v=1.0.0)

---

## Engineer Gravity — Structural Influence (v0.10.0)

The 7-axis impact tells you *how strong* an engineer is. The 3-axis topology tells you *what kind* of engineer they are. But there's another dimension: **how much structural influence does this person exert on the codebase?**

That's what **Gravity** measures.

```
Gravity = Indispensability × 0.40 + Breadth × 0.30 + Design × 0.30
```

It combines module ownership (bus factor risk), cross-cutting reach (breadth), and architectural involvement (design) into a single number that approximates **structural pull** — how much of the system's shape is determined by this person.

**But high Gravity isn't automatically good.** Gravity has a health dimension.

A high-quality engineer with high Gravity is exerting *healthy structural influence* — their designs form the backbone of the system and those designs are durable. A low-quality engineer with high Gravity is a *fragile structural dependency* — the system depends on them, but the code itself is brittle.

This is expressed through color:

```
health = Quality × 0.6 + RobustSurvival × 0.4

Gravity < 20  → dim gray  (low influence)
health ≥ 60   → green     (healthy gravity)
health ≥ 40   → yellow    (moderate)
health < 40   → red       (fragile gravity)
```

In practice, this reveals striking patterns. On my backend team, my own Gravity is 97 (green) — Design 100, Survival 100, and Indispensability 43 indicate well-distributed structural influence backed by durable code. On another domain, one member has Gravity 100 (red) — extremely high Indispensability means they own the vast majority of modules, but quality is lacking. **"If this person leaves, everything collapses" AND "the code itself is fragile"** — the most dangerous combination, instantly visible as red gravity.

Gravity is intentionally excluded from Impact. Impact answers "how strong is this engineer?" Gravity answers a different question: "how much structural influence do they have, and is it healthy?" They're orthogonal dimensions. In terminal output, Gravity appears as a color-coded column next to each member's signals.

We're **reverse-engineering structure from code, through individual engineers** — Gravity is the first step in reading team architecture from git history.

---

## Real-World Results

I ran this on my own team (14 repos, 10+ engineers including departed members). Here are anonymized excerpts.

Yes, I rank highest. I'm the tech lead and I designed the metric — if the person making architectural decisions *didn't* top the leaderboard, that would be a red flag, not a feature. It also helps that I log the most hours on this codebase by a wide margin, so take the top spot with a grain of salt. The real validation is elsewhere: the rankings for everyone else matched the team's gut feeling almost perfectly.

### Backend Rankings (Excerpt)

![Backend Scores](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch1-backend-table.png?v=4)

† Insufficient sample (fewer than 10 fix-commit involvements). Neutral value 50 used in Impact calculation.

**Z.** — Impact 24.9. Breadth was the only elevated signal — Production 6, Design 4, Survival nearly zero. The topology reads `— / Spread / —`. This is a profile where presence is wide but depth is absent. An early observation of this pattern could have informed better role alignment — perhaps pairing with an Architect to focus contributions into a narrower, higher-impact area.

**Y.Y.** — Design 67, Breadth 81. The original architect of the codebase's early structure. Indispensability 100 — the highest on the team — indicates that more modules are still attributed to this person than to any active member. Time decay brought Survival to 12, but the codebase's structural foundation still reflects their design decisions. The topology reads `Architect / — / Former` — the Role persists in the code even after the engineer's departure. This is the signal that Chapter 4 calls "a soul that needs to be laid to rest" — not a criticism, but a handoff priority.

**R.S.** — Production 17 doesn't turn heads. But Survival 50 (2nd on the team) means their recent code stays. Debt Cleanup 88 means they're quietly fixing everyone else's bugs. **This is exactly the kind of person that Debt Cleanup was designed to surface.** The Anchor role captures this perfectly.

My own Quality of 57 is low, reflecting aggressive architectural changes (introducing DelegateProcess layers, designing PartProcess abstractions, iterating on domain models in code). Combined with Design 100, it reads as proactive improvement rather than sloppiness. The Builder style reflects that I'm not just designing — I'm building heavily too.

### Frontend Rankings (Excerpt)

![Frontend Scores](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch1-frontend-table.png?v=4)

**R.M.** owns essentially all of the new product's FE repository core library (200K+ lines). Indispensability 100 — same structural risk as Y.Y. on the backend. **If this person leaves, there is nobody who can make frontend design decisions.** Debt Cleanup 39 is mid-range within FE, but with 129 self-fixes, it shows a self-contained work style: write it, fix it yourself.

**X.**: Quality 18 — 82% of their commits were fixes or corrections. Survival ≈ 0, Debt Cleanup 0. High volume of production, high volume of fixes, but the code did not survive. `Producer / Churn / Former` — the topology describes a pattern where production exists but survival does not. This signal may indicate that pair programming, design review support, or a structured mentoring environment could help convert that production energy into durable code.

---

## Why Revenue Metrics Miss Engineering Health

This section is for executives and managers.

"Revenue is growing, so engineering must be fine" — this is a dangerous assumption. Revenue measures **product-market fit**, not **engineering health**.

Think of it this way: revenue is a car's speed. Engineering health is the engine's condition. An engine can be failing and still produce speed — if you're going downhill. Speed alone doesn't tell you the engine is healthy.

To understand engineering health, you need signals that revenue doesn't capture:

- **Code durability** — are you rewriting the same features every quarter?
- **Technical debt accumulation** — does adding 1 feature generate 2 bug fixes?
- **Bus factor** — how many modules die if one person leaves?
- **Design concentration** — is all architecture decided by one person?

Git history contains all of these signals.

**Even with revenue growing, if Survival decline + Debt increase + Bus Factor concentration are progressing simultaneously, the organization will collapse at scale.**

---

## How This Differs from DORA, SPACE, and Other Metrics

If you're familiar with engineering productivity frameworks, you might wonder: how does this relate to what already exists?

| Framework | What it measures | Key limitation |
|---|---|---|
| **DORA** | Deployment speed & stability | Team-level. Doesn't measure code quality or individual impact |
| **SPACE** | 5 holistic dimensions (surveys + tools) | Survey-heavy, 3–6 months to implement |
| **LOC / Commits** | Activity volume | Trivially gameable, penalizes refactoring |
| **Code Churn** | % of recent code rewritten | Context-blind — can't distinguish refactoring from instability |
| **Bus Factor** | Knowledge concentration risk | Only identifies risk, not impact |
| **Git analytics tools** (Pluralsight Flow, LinearB) | Activity & cycle time | Measures *when* code was written, not *whether it lasted* |

**The gap:** existing frameworks measure activity or velocity. None of them ask the question *"did this individual's code actually survive?"*

DORA tells you how fast code reaches production. This model tells you whether it was worth deploying.

Time-decayed survival is also naturally resistant to gaming. You can't inflate your impact with busy work — only code that remains in the codebase months later counts. And the debt cleanup axis makes it structurally impossible to achieve high impact by generating work for others.

---

## Accuracy Scales with Design Quality

This model has an interesting property: **higher codebase design quality yields higher observation accuracy**.

In well-structured codebases (Clean Architecture, DDD), the assumptions hold: "touching design files = making design decisions" is true, and high Survival means "the design withstood change pressure."

In chaotic codebases, high Survival might just mean "dead code nobody touches." And low Design signal might just mean "architecture files don't exist as a clear category."

**The metric's low accuracy is itself a signal of poor design.** If the observations don't match your gut feeling, the problem may not be the model — it may be that your codebase structure can't withstand measurement. Investing in design is itself infrastructure for improving observation accuracy.

---

## Observations Reflect the Organization, Not Just the Individual

One critical nuance: **low impact doesn't necessarily mean a weak engineer.**

Consider:

- **Ambiguous or frequently changing specs** — even a strong engineer will accumulate fix commits and rewrites. Quality and Survival drop not because of poor coding, but because of poor planning upstream
- **Slow design reviews or decision bottlenecks** — engineers can't start their next task. Low Production isn't slowness — it's organizational friction
- **Misalignment between product requirements and engineering** — features get built, then rebuilt because "that's not what we meant." Survival drops not from bad design, but from bad communication

In other words, this metric captures **the environment an engineer operates in** — spec quality, planning precision, decision-making speed — not just their individual ability. If the *entire team* shows low impact, that's a signal to examine organizational processes before blaming individual engineers.

The flip side is equally true: **improve planning and spec quality, and engineer impact will naturally rise.** The observation is simultaneously an engineer's report card and an organizational health barometer.

From the engineer's perspective, raising your impact requires more than just writing what you're told. You need to evaluate whether the spec makes sense, push back when it doesn't, engage in design discussions, and deeply understand the product before writing code. Durable code only emerges from correct understanding and correct design decisions. This metric naturally rewards **engineers who care about the product and engage from the spec level, not just the code level.** Do that, and your impact will grow. Maybe that's what real "combat power" actually is.

---

## Who Makes the Decisions at Your Company?

If you're reading this and thinking "I should try this on my team," here's a question worth asking:

**What is the combat power ranking of the person making engineering decisions at your company?**

If you want to build a genuinely strong engineering organization, the person making architectural decisions **must** be among the highest impact members on the team. Why? Because the quality of design decisions shows up in code. The architect should be someone who writes code personally, and whose code survives. That's the only reliable way to ensure design quality.

When there's a gap between decision-making authority and observed structural impact, a misalignment emerges: **engineers with strong Design and Survival signals may find their contributions overridden by decisions that don't reflect the codebase's structural reality.** This is a structural signal worth investigating — not to assign blame, but to ask whether decision-making authority and structural contribution are aligned.

If you observe this pattern, the telescope is showing you something about your organization's structure — worth a conversation, not a verdict.

---

## Limitations

This model is not perfect.

- Commit messages affect fix detection accuracy → Quality is weighted at only 10%
- Blame analysis can be expensive on large repos → sample up to 500 files
- Frontend refactors reduce survival signals → separate BE/FE profiling
- Copy-paste inflates production → offset by survival and design
- Debt cleanup depends on sample size → threshold: <10 events = neutral value 50
- Indispensability conflates "strong" with "nobody took over" → weight is only 5%

Not universal. But as **validation of your intuition**, it works far better than "just a feeling." At minimum, it's 100x better than guessing.

---

## This Is Not a Measure of Human Worth

This model does **not** measure a person's value.

It estimates **technical influence observable in a codebase**.

Engineers contribute in ways git cannot capture: mentoring, domain expertise, documentation, psychological safety, team culture. Those contributions matter enormously. This observation only quantifies what git records — nothing more. It's not universal. But zero measurement is infinitely worse.

---

## Measure Every Quarter

This model's real value comes from **tracking changes over time**.

If Survival is rising quarter-over-quarter, that engineer's **design skills are growing**. If fix ratio is dropping, **first-pass quality is improving**. If Debt Cleanup is rising, **team contribution is increasing**.

Conversely, if only Production is rising while everything else flatlines — output increased, but quality didn't.

Numbers don't lie.

---

## Try It: The CLI Tool

This started as a blog post and a Claude Code experiment. The formulas are now baked into a standalone CLI — **zero AI tokens, zero API keys, zero cloud dependency.**

```bash
❯ brew tap machuz/tap
❯ brew install eis
```

```bash
# Analyze current repo
❯ eis analyze .

# Auto-discover repos under a directory
❯ eis analyze --recursive ~/projects

# With config
❯ eis analyze --config eis.yaml --recursive ~/projects
```

Runs in seconds to minutes. Color-coded output with 7-axis signals, 3-axis topology (Role / Style / State), and Bus Factor risks — right in your terminal.

![Terminal Output](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/terminal-output.png?v=0.11.0)

## Final Thought

Git quietly records the real history of a system. If we analyze that history carefully, we can understand who builds durable systems, who stabilizes them, and where organizational risk exists.

Numbers will never tell the whole story. But **quantify what you can, then qualitatively supplement what you can't.** That order matters.

As AI-driven development grows, this model could also measure **the rigidity of AI-generated code**. Is AI a debt factory or a cleaner? Git records humans and AI equally. Quantitatively tracking AI-written code survival and debt ratio — that feels like the next phase of engineering evaluation.

If you run this model on your own codebase, the results may surprise you. The numbers are more honest than you think.

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full.png?v=2)

**GitHub**: [eis](https://github.com/machuz/eis) — CLI tool, formulas, and methodology all open source. `brew tap machuz/tap && brew install eis` to install.


If this was useful: [❤️ Sponsor on GitHub](https://github.com/sponsors/machuz)

---

### Series

- [Chapter 0: What If Git History Could Tell You Who Your Strongest Engineers Are?](https://dev.to/machuz/git-archaeology-0-what-if-git-history-could-tell-you-who-your-strongest-engineers-are-5397)
- **Chapter 1: Measuring Engineering Impact from Git History Alone**
- [Chapter 2: Beyond Individual Scores: Measuring Team Health from Git History](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f)
- [Chapter 3: Two Paths to Architect: How Engineers Evolve Differently](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga)
- [Chapter 4: Backend Architects Converge: The Sacred Work of Laying Souls to Rest](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d)
- [Chapter 5: Timeline: Scores Don't Lie, and They Capture Hesitation Too](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5)
- [Chapter 6: Teams Evolve: The Laws of Organization Revealed by Timelines](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4lei)
- [Chapter 7: Observing the Universe of Code](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0)
- [Chapter 8: Engineering Relativity: Why the Same Engineer Gets Different Scores](https://dev.to/machuz/git-archaeology-8-engineering-relativity-why-the-same-engineer-gets-different-scores-5dnl)
- [Chapter 9: Origin: The Big Bang of Code Universes](https://dev.to/machuz/git-archaeology-9-collapse-good-architects-and-black-hole-engineers-1dcn)
- [Chapter 10: Dark Matter: The Invisible Gravity](https://dev.to/machuz/git-archaeology-10-dark-matter-the-invisible-gravity-45ne)
- [Chapter 11: Entropy: The Universe Always Tends Toward Disorder](https://dev.to/machuz/git-archaeology-11-entropy-the-universe-always-tends-toward-disorder-ak9)
- [Chapter 12: Collapse: Good Architects and Black Hole Engineers](https://dev.to/machuz/git-archaeology-12-collapse-good-architects-and-black-hole-engineers-3fed)
- [Chapter 13: Cosmology of Code](https://dev.to/machuz/git-archaeology-13-cosmology-of-code-dci)
- [Chapter 14: Civilization — Why Only Some Codebases Become Civilizations](https://dev.to/machuz/git-archaeology-14-civilization-why-only-some-codebases-become-civilizations-2nl3)
- [Chapter 15: AI Creates Stars, Not Gravity](https://dev.to/machuz/git-archaeology-15-ai-creates-stars-not-gravity-4i05)
- [Final Chapter: The Engineers Who Shape Gravity](https://dev.to/machuz/git-archaeology-16-the-engineers-who-shape-gravity-3fmi)

---

[Chapter 2: Team Health →](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f)

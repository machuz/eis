---
title: "DORA, SPACE, LinearB — and the Question They All Leave Unanswered"
published: true
description: "DORA measures deployment speed. SPACE surveys developer experience. LinearB tracks cycle time. None of them ask: does your code still exist? A comparison of engineering metrics frameworks — and what git blame reveals that they can't."
tags: productivity, devops, git, opensource
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/cover-comparison.png?v=1
---

*Every framework measures something real. The question is what it leaves in the dark.*

---

## The Metrics Landscape in 2026

Engineering leaders have more measurement frameworks than ever:

- **DORA** — the gold standard for delivery performance
- **SPACE** — a multi-dimensional developer productivity framework from Microsoft Research
- **LinearB** — a commercial platform tracking engineering workflow metrics
- **CodeScene** — behavioral code analysis with hotspot detection
- **git-fame / git-quick-stats** — basic git attribution tools

Each captures something genuine. But after using them across multiple teams, I kept running into the same blind spot: **none of them tell you whether the code an engineer wrote is still standing.**

That question — *does your code survive?* — is what led me to build [EIS (Engineering Impact Signal)](https://github.com/machuz/eis). This article isn't about declaring a winner. It's about understanding what each framework sees, what it misses, and where the gaps are.

---

## DORA: The Pipeline Telescope

**What it measures:** Deployment frequency, lead time for changes, change failure rate, mean time to restore service.

**What it gets right:** DORA gave the industry its first empirically-backed framework for delivery performance. The four metrics are well-defined, measurable, and correlated with organizational outcomes. The annual State of DevOps reports provide longitudinal evidence.

**What it leaves in the dark:**

- **Individual resolution is zero.** DORA operates at the team/organization level by design. You can tell *whether* the team is shipping fast, but not *who* is building the structures that enable that speed.
- **Code durability is invisible.** A team can have elite DORA metrics while accumulating massive technical debt. Fast deployment says nothing about whether the code survives next quarter.
- **Architecture is unobserved.** Who shaped the system? Who holds the design knowledge? DORA doesn't ask.

**Where it fits:** DORA is a delivery health check. It answers "how fast and safely does this team ship?" — a critical question, but a narrow one.

| | DORA | EIS |
|---|---|---|
| **Scope** | Team / org | Individual + team |
| **Data source** | CI/CD pipeline | Git history only |
| **Core question** | "How fast do we ship?" | "Does the code survive?" |
| **Individual resolution** | None | 7-axis per engineer |
| **Code survival** | Not measured | Core thesis (25% weight) |
| **Architecture visibility** | None | Design axis + module topology |
| **Setup** | CI/CD integration required | `brew install eis && eis` |

---

## SPACE: The Survey Telescope

**What it measures:** Satisfaction & well-being, Performance, Activity, Communication & collaboration, Efficiency & flow. Proposed by Microsoft Research, GitHub, and the University of Victoria.

**What it gets right:** SPACE acknowledged that productivity is multi-dimensional — that a single metric will always be gamed or misinterpreted. The framework legitimized subjective measures (satisfaction, flow) alongside objective ones. This was an important intellectual step.

**What it leaves in the dark:**

- **It's a framework, not a tool.** SPACE describes *what categories* to measure but doesn't prescribe *how*. Operationalizing it requires surveys, telemetry integration, and significant organizational effort.
- **Survey data decays fast.** Developer satisfaction measured in Q1 might not reflect Q2 reality. Surveys are snapshots with social-desirability bias.
- **Code-level observation is absent.** SPACE's "Performance" dimension includes code review velocity and code quality, but doesn't specify measurement methods. Whether code survives, who owns the architecture, bus factor risk — these aren't part of the framework.
- **Individual signals are politically sensitive.** Using SPACE for individual-level measurement is explicitly discouraged by its authors, leaving the individual contribution question unanswered.

**Where it fits:** SPACE is the most thoughtful meta-framework for thinking about productivity. It's the right starting point for deciding *what* to measure. But it needs tools to provide the actual measurements.

| | SPACE | EIS |
|---|---|---|
| **Nature** | Conceptual framework | Measurement tool |
| **Data source** | Surveys + telemetry + code metrics | Git history only |
| **Core question** | "What dimensions matter?" | "What trace did you leave?" |
| **Operationalization** | Requires significant setup | Single CLI command |
| **Subjectivity** | Surveys included | Fully objective (git data) |
| **Code survival** | Not specified | Core thesis |
| **Architecture** | Not specified | Design axis + module topology |

---

## LinearB: The Workflow Telescope

**What it measures:** Cycle time, coding time, review time, deployment frequency, PR size, developer workload. A commercial platform with Git and project management integrations.

**What it gets right:** LinearB provides actionable workflow analytics. It identifies bottlenecks in the development cycle — long review times, oversized PRs, uneven workload distribution. The benchmarking feature lets teams compare against industry data.

**What it leaves in the dark:**

- **Activity over impact.** LinearB measures *when* and *how fast* code was written and reviewed — not *whether it lasted*. An engineer who ships 50 small PRs that all get rewritten next sprint looks productive.
- **Code durability is absent.** No survival analysis. No distinction between code that endures and code that churns.
- **Architecture is opaque.** Who contributes to architectural decisions? Who holds critical module knowledge? LinearB tracks workflow, not structural influence.
- **Proprietary and SaaS-only.** Requires connecting your repositories and project management tools. The observation methodology isn't open for inspection.

**Where it fits:** LinearB is an engineering workflow optimizer. It makes development processes visible and identifies friction points. It answers "where is our process slow?" — not "what is our codebase's structural health?"

| | LinearB | EIS |
|---|---|---|
| **Scope** | Team workflow | Individual + team structure |
| **Data source** | Git + Jira/Linear + CI/CD | Git history only |
| **Core question** | "Where is our process slow?" | "What structural impact did you leave?" |
| **Code survival** | Not measured | Core thesis (25% weight) |
| **Pricing** | Free tier + paid plans | Free and open source |
| **Setup** | SaaS integration | Single CLI command |
| **Architecture** | Not measured | Design axis + module topology |

---

## CodeScene: The Behavioral Telescope

**What it measures:** Code health, hotspots, code-level complexity trends, organizational coupling, knowledge distribution.

**What it gets right:** CodeScene is the closest to EIS in philosophy. It analyzes behavioral patterns in code, identifies hotspots (frequently changed complex code), and detects organizational patterns. The X-ray feature shows complexity trends within files. It asks questions about the code itself, not just the process.

**What it leaves in the dark:**

- **Engineer topology is limited.** CodeScene identifies knowledge silos and organizational patterns, but doesn't classify engineers along multiple axes (Role, Style, State). You can see *that* knowledge is concentrated, but not *what kind* of engineer holds it.
- **Survival analysis differs.** CodeScene tracks code age and change frequency, but doesn't apply time-decayed survival weighting. The distinction between robust survival (enduring under change pressure) and dormant survival (persisting because nobody touches it) is absent.
- **Proprietary methodology.** The scoring algorithms aren't fully open. EIS publishes every formula in the [whitepaper](https://github.com/machuz/eis/blob/main/docs/whitepaper.md).
- **Module topology is different.** CodeScene maps organizational coupling and hotspots. EIS maps change pressure, co-change coupling, module survival, and ownership fragmentation — then classifies modules along 3 axes (Coupling, Vitality, Ownership).

**Where it fits:** CodeScene is strong for code-level health monitoring and organizational risk detection. It operates at a different resolution than EIS — more focused on the code's health, less on the engineer's structural fingerprint.

| | CodeScene | EIS |
|---|---|---|
| **Focus** | Code health + org patterns | Engineer impact + structural topology |
| **Data source** | Git + optional integrations | Git history only |
| **Core question** | "Where is the code unhealthy?" | "Who shaped the structure?" |
| **Engineer classification** | Knowledge distribution | 3-axis topology (Role/Style/State) |
| **Code survival** | Code age tracking | Time-decayed survival with robust/dormant split |
| **Pricing** | Commercial (paid) | Free and open source |
| **Methodology** | Proprietary | Fully open ([whitepaper](https://github.com/machuz/eis/blob/main/docs/whitepaper.md)) |

---

## git-fame and git-quick-stats: The Raw Data

**What they measure:** Lines of code per author, commit counts, file-level attribution.

These tools are useful for quick summaries, but they measure *volume*, not *impact*. An engineer who wrote 50,000 lines of generated code dominates the rankings. An engineer who wrote 500 lines of architecture that the entire system depends on is invisible.

EIS starts from the same raw data (`git log`, `git blame`) but adds:

- **Time-decayed survival** — old code weighs less than recent code
- **Robust vs. dormant distinction** — surviving under change pressure vs. surviving in an untouched corner
- **Debt cleanup tracking** — who fixes other people's bugs
- **Architectural pattern detection** — contributions to structurally significant files
- **3-axis engineer topology** — not just "how much" but "what kind"

---

## The Gap: What None of Them Measure

Here's the fundamental question that existing frameworks leave unanswered:

> **Of all the code this engineer wrote, how much is still standing — and is it standing because it's good, or because nobody touches it?**

This is the **survival question**, and it changes everything.

Consider two engineers:

- **Engineer A** ships 200 PRs per quarter. DORA metrics are elite. LinearB shows fast cycle time. But 6 months later, 80% of their code has been rewritten by others.
- **Engineer B** ships 40 PRs per quarter. Modest by any activity metric. But their code forms the structural backbone of the system — still there after two years, surviving under active change pressure.

DORA sees a fast team. LinearB sees high throughput. SPACE might capture satisfaction. **But none of them see that Engineer B is the gravitational center of the codebase.**

EIS sees it. Because it asks the question that requires `git blame` to answer: *what survived?*

---

## Complementary, Not Competitive

These frameworks aren't rivals. They're telescopes pointed at different parts of the sky.

```
DORA        → "Is the delivery pipeline healthy?"
SPACE       → "What dimensions of productivity matter?"
LinearB     → "Where are the workflow bottlenecks?"
CodeScene   → "Where is the code unhealthy?"
EIS         → "Who shaped the structure — and does it endure?"
```

A mature engineering organization might use DORA for delivery health, LinearB for process optimization, and EIS for understanding *who is actually building the structures that everything else runs on.*

The most dangerous state is measuring only activity. Teams that optimize for deployment frequency and cycle time without observing code survival can ship fast while accumulating invisible structural debt. The code churns. The architecture erodes. And the engineers who quietly hold it all together — the [Anchors](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c), the [Cleaners](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c), the [Dark Matter](https://dev.to/machuz/git-archaeology-10-dark-matter-the-invisible-gravity-45ne) — remain invisible.

---

## The Full Comparison

| Dimension | DORA | SPACE | LinearB | CodeScene | EIS |
|---|---|---|---|---|---|
| **Scope** | Team/org | Framework | Team workflow | Code + org | Individual + team + module |
| **Data source** | CI/CD | Surveys + mixed | Git + PM tools | Git + optional | Git only |
| **Individual resolution** | None | Discouraged | Workflow per dev | Knowledge maps | 7-axis + 3-axis topology |
| **Code survival** | No | Not specified | No | Code age | Time-decayed + robust/dormant |
| **Architecture** | No | Not specified | No | Hotspots | Design axis + module topology |
| **Bus factor** | No | Not specified | No | Knowledge silos | Indispensability axis |
| **Debt tracking** | No | Not specified | No | Code health | Debt Cleanup axis |
| **Engineer classification** | No | No | No | Limited | Role / Style / State |
| **Module classification** | No | No | No | Hotspots | Coupling / Vitality / Ownership |
| **Setup cost** | CI/CD integration | Surveys + tooling | SaaS integration | SaaS / on-prem | `brew install eis` |
| **Pricing** | Free (metrics) | Free (framework) | Freemium | Commercial | Free and open source |
| **Methodology** | Published | Published | Proprietary | Proprietary | [Fully open](https://github.com/machuz/eis/blob/main/docs/whitepaper.md) |

---

## Try It Yourself

```bash
brew tap machuz/tap && brew install eis
cd your-repo
eis analyze .
```

No API keys. No AI tokens. No SaaS integration. Just `git log` and `git blame`.

Point the telescope at your codebase and see what the other frameworks can't show you: **who shaped the structure, and does it endure.**

> Full methodology: [Whitepaper](https://github.com/machuz/eis/blob/main/docs/whitepaper.md) · [GitHub](https://github.com/machuz/eis)

---

### The Git Archaeology Series

This article is a standalone comparison piece. For the full exploration of what git history reveals about engineering teams:

- [Chapter 0: What If Git History Could Tell You Who Your Strongest Engineers Are?](https://dev.to/machuz/git-archaeology-0-what-if-git-history-could-tell-you-who-your-strongest-engineers-are-5397)
- [Chapter 1: Measuring Engineering Impact from Git History Alone](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)
- [Full series (17 chapters)](https://dev.to/machuz/git-archaeology-0-what-if-git-history-could-tell-you-who-your-strongest-engineers-are-5397)

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png)

**GitHub**: [eis](https://github.com/machuz/eis) — CLI tool, formulas, and methodology all open source.

If this was useful: [Sponsor on GitHub](https://github.com/sponsors/machuz)

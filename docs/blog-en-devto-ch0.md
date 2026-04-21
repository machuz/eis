---
title: "Git Archaeology #0 — What If Git History Could Tell You Who Your Strongest Engineers Are?"
series: "Git Archaeology"
published: true
description: "A 3-minute intro to Engineering Impact Signal — an OSS CLI that quantifies engineering impact from git log and git blame alone. No API keys, no AI tokens."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/cover-ch0.png?v=1
---

*git log + git blame. That's all it takes.*

---

## What Is This?

**Engineering Impact Signal (EIS, pronounced "ace")** is an open-source CLI tool that quantifies engineering impact from Git history alone.

No external APIs. No AI tokens. Just `git log` and `git blame`.

```bash
brew tap machuz/tap && brew install eis
cd your-repo
eis
```

That's it. Here's what you get:

![Terminal Output](https://raw.githubusercontent.com/machuz/eis/main/docs/images/terminal-output.png?v=0.11.0)

---

## Why I Built This

Commit counts. PR counts. Lines of code. Easy to measure — and all meaningless.

A typo fix and a system-wide architecture change both count as "one PR." A generated lockfile adds thousands of lines. Commit frequency varies wildly between engineers.

Yet inside every team, people sense whose structural signal is strongest.

> "That person writes code that lasts."
> "That person touches everything but nothing improves."

Those intuitions exist, but they're not measurable. I wanted to find a way to **observe that gut feeling as numbers** — drawn from a source that can't be gamed by politics: **the git history itself.**

---

## The Telescope

Over the years, I've recruited several strong engineers by reaching out personally. I'm grateful that many of them said yes.

Why did they come? I don't think it was just the tech stack or compensation.

**"This person actually sees my work."** — I believe that's what they felt.

For engineers, having your technical contributions properly observed matters deeply. Not commit counts or PR counts, but **whether your code survives, whether you contribute to architecture, whether you clean up debt** — having someone who sees that substance.

I had that eye. At least, that's my self-assessment.

EIS is that **observer's eye, turned into an open-source telescope.**

Anyone can use it. Point it at any team. Through the lens of git history — a lens that cannot lie.

---

## Building a Structure Where Sincere Makers Win

Engineers in Japan are paid less than their peers in other countries.

It's not because they lack skill. I believe it comes from a cultural tendency to let the work speak for itself — rather than asserting their own value. They write code in silence, fix architecture in silence, clean up debt in silence — and that work stays **invisible**. Invisible means unheard. And when it's unheard, the structure pulls the soul of their work toward whoever speaks the loudest.

Whenever I've sensed that happening, I've resisted hard enough to warp the team's magnetic field.

I want to build a world where the work of people who sincerely face their craft **becomes visible**.

As I built EIS, I started to see the shape of what I really wanted to do: **create a structure where people who sincerely face their craft are the ones who win.** The moment I gained that self-awareness, the energy exploded, and I was able to bring this tool this far in a short period of time.

And now I'm thinking about what comes after the telescope.

A telescope **observes** the universe. But observation alone doesn't change an engineer's life. You need to **interpret** the observation data, **propose** a universe that fits them — the right codebase, the right team, the right organization — and **show** them a stable orbit within that universe. Only then does "a structure where sincere makers win" become real.

That's the next step for EIS. Turning the telescope into an observatory.

---

## Making Architecture a Science

Being strong at math. Being strong at algorithms. Being strong at language specifications.

These all have theory forged in academia over decades. Computational complexity, type theory, formal verification — they can guarantee correctness through mathematical proof. They stand on scientific foundations.

Academic attempts at software architecture have existed for over 30 years. Architecture Description Languages and evaluation methods have been proposed. But **they haven't converged into a unified theory.** The efforts remain fragmented and haven't made it into practice.

"What is good design?" "Is this team's structure healthy?" — the industry has best practices and rules of thumb for these questions, but very little quantitative language.

And now AI writes enormous amounts of code.

**The value of writing code is declining in relative terms. What matters most is structure.** What structure do you place code on top of? Does that structure withstand change? Where does the team's knowledge accumulate?

If EIS can become a **tool for making architecture a science**, I'd be glad.

And now it observes more than just people. To observe **modules themselves**, we designed 4 new metrics:

| Metric | What it measures |
|---|---|
| **Change Pressure** | Change frequency ÷ code volume per module. Higher pressure = more structural stress |
| **Co-change Coupling** | Module pairs that change together. Detects implicit coupling invisible in import graphs |
| **Module Survival** | Time-decayed survival rate of code within a module |
| **Ownership Fragmentation** | How knowledge is distributed across a module. Measured via Shannon entropy |

These metrics combine to classify every module along 3 independent axes — Coupling (boundary quality), Vitality (change pressure × survival), and Ownership (knowledge distribution). Invisible structural risks become observable data:

- `Hub × Critical × Orphaned` — a module at the center of implicit dependencies, under extreme change pressure, with no active owner. Maximum risk.
- `Independent × Stable × Distributed` — a well-bounded module with healthy ownership. The ideal state.

The telescope now observes **both the stars (engineers) and the space they inhabit (modules).**

---

## 7 Axes of Impact

EIS observes engineers across 7 axes:

| Axis | Weight | What it measures |
|---|---|---|
| Production | 15% | Volume of changes |
| Quality | 10% | First-time quality (low fix/revert rate) |
| **Survival** | **25%** | **Does your code still exist today? (time-decayed)** |
| Design | 20% | Contributions to architecture files |
| Breadth | 10% | Cross-repository activity |
| Debt Cleanup | 15% | Cleaning up other people's debt |
| Indispensability | 5% | Module ownership (bus factor) |

The most important axis is **Survival**. Is the code you wrote still there after 6 months? After a year?

Writing a lot means nothing if it gets rewritten next month. **Engineers who write code that lasts are the strong ones.**

---

## 3-Axis Archetypes

Beyond signals, EIS classifies engineers along three independent axes:

**Role** — What you contribute
- Architect: designs the structure
- Anchor: guards quality
- Cleaner: pays down debt
- Producer: generates volume
- Specialist: deep in one area

**Style** — How you contribute
- Builder: builds and designs simultaneously
- Resilient: rebuilds after destruction
- Rescue: pays down others' debt
- Churn: high volume, low survival
- Mass: mass production with low survival
- Balanced: even across all axes
- Spread: touches everything but lacks depth

**State** — Lifecycle phase
- Former: left the team, but their code remains
- Silent: low activity, low survival (detected only for experienced engineers)
- Fragile: code survives only because nobody touches it
- Growing: low volume, but high quality
- Active: currently contributing

From these classifications, **team structure becomes visible.**

---

## What This Reveals

Here's what EIS has surfaced in real teams:

- **A departed Architect's code still makes up 30% of the codebase** (Former detection)
- **Code that survives only because nobody touches it** — not because it's good (Fragile detection)
- **No Producers on the team** — the layer that generates volume on top of structure is empty (Producer Vacuum)
- **Architect Bus Factor = 1** — all design knowledge concentrated in one person
- **136 Orphaned modules** — owners have left, nobody holds the knowledge (Module Topology)
- **12 Critical modules** — high change pressure + code doesn't survive. Structural time bombs (Module Topology)

Cold git history tells **team stories** you didn't know you had. And module topology tells **where the system is breaking**, not just who is strong.

---

## Validated on the OSS Universe

The best way to verify a telescope works is to **point it at stars whose positions are already known.**

We ran EIS against **29 open-source repositories spanning 55,343 engineers.** React, Kubernetes, Rails, Laravel, esbuild, Rust — projects whose structures are common knowledge.

The results matched community intuition:

- **esbuild**: Evan Wallace hits 100 on every axis. Gravity concentration: 92.5% — exactly the "Evan built it alone" consensus
- **Rails**: 6 engineers with Design above 35. A civilization that distributed design authority over 20 years — DHH, Jeremy Kemper, Rafael Franca, and others
- **Laravel**: Taylor Otwell at 100, every other top-10 contributor below Design 4 — the "Taylor's creation" consensus, quantified
- **React**: 5 generations of architect transitions over 10 years — Paul O'Shannessy → Dan Abramov → Brian Vaughn → Sebastian Markbåge → Jorge Cabiedes. The Gravity Map revealed dynamics that diverge from commit counts. The most striking case is Jorge Cabiedes. With just **82 commits**, he reached Gravity 60 — entering the top 3 gravity zone among 2,010 engineers. Design 100, Production 100, Debt Cleanup 96. Meanwhile, Dan Abramov has **1,890 commits** with Gravity 51.1. Twenty-three times more commits, yet less gravitational pull. Jorge generates nearly the same gravity as Brian Vaughn (1,627 commits, Gravity 61.6) with one-twentieth the commit volume. Jorge's style classification is **Builder** — someone who designs, constructs, and cleans up. He appeared suddenly in the 2025 gravity distribution, and Design 100 is a number only Jorge holds among the top 5 (Sebastian Markbåge 7, Dan Abramov 10, Brian Vaughn 4, Andrew Clark 20). A codebase's gravity field is shaped not by the **volume** of commits but by the **quality of structural engagement**. Even 82 commits, when concentrated on architecture files, when the code survives, when it cleans up debt — exert stronger gravity than 1,890 commits. The force invisible when scrolling git log is etched into the strata of git blame. There's another story that only emerges by tracing the timeline. Sebastian Markbåge. From 2016 to 2023, Sebastian's gravity was **always there, but never at the center**. 68 → 72 → 33 → 60 → 66 → 63 → 63 → 38. As gravitational centers shifted — Paul O'Shannessy's field, Dan Abramov's field, Brian Vaughn's field — Sebastian remained within the top-5 gravity zone but never once became the gravitational pole. Then 2024 — Gravity **100**. The center of the gravity field quietly shifted to Sebastian. What happened? Sebastian's role classification is **Cleaner**. Not a Producer churning out features, not an Architect drawing blueprints. **Someone who cleans up structural debt left by others and maintains the codebase's integrity**. Debt Cleanup 54, Quality 96.2, and Indispensability **100** — meaning the highest module ownership ratio in all of React's codebase: without Sebastian's code, it doesn't stand. Line up the 8 years of gravity readings and there's nothing flashy. But his code **kept surviving** (Survival 79.2). Dan Abramov with 1,890 commits at Survival 0.1, Brian Vaughn with 1,627 commits at Survival 0.1 — as waves of rewrites swept away code, Sebastian's 1,495 commits remained in the strata at Survival 79.2. As other engineers departed and code was rewritten, the **ratio of code that didn't disappear** grew. In 2024, a threshold was crossed. The quietly accumulated strata became the gravity field itself. You can't see this in a snapshot. Only by tracing the annual gravity distribution over time does this story surface: "A Cleaner who was never at the gravitational center for 8 years became the gravitational pole after tectonic shifts in the codebase." This is the strength of the Gravity Map's temporal axis
- **Kubernetes**: Gravity concentration 0.8%. Structure distributed across 5,000+ contributors

An even more interesting finding: **gravity concentration varies 4.8× across language families.**

| Language category | Gravity concentration | Structural physics |
|---|---|---|
| Go (anti-framework culture) | 16.4% | Concentrated in few architects |
| Rust / Scala (expressive) | 6.7% | Type systems distribute structure |
| Rails / Laravel (framework-driven) | 5.1% | Frameworks absorb structure |
| C / C++ (systems) | 3.4% | Most distributed |

Here's the critical point: **this is not about which structure is "correct."**

esbuild's 92.5% concentration isn't "bad design" — at a scale where one person can hold the entire system in their head, it may be optimal. Kubernetes' 0.8% distribution isn't "better because it's distributed" — at 5,000+ contributors, distribution is inevitable, and that itself is a design outcome.

What EIS observes is **the physics of structure**, not a judgment of quality. A telescope describes the shape of galaxies. It doesn't claim spiral galaxies are "better" than elliptical ones.

### Top 50: The Brightest Stars in the OSS Universe

We also mapped the **Gravity distribution of the top 50 engineers** — structural influence — across all 29 projects.

> [OSS Gravity Map — Top 50 Engineers](https://machuz.github.io/eis/research/oss-gravity-map/analysis/top50.html)

Salvatore Sanfilippo (Redis), Alexey Milovidov (ClickHouse), Ritchie Vink (Polars) — their gravity saturates the scale. But the more remarkable finding was the **440 engineers the world has never heard of.** They don't give conference talks. They don't have mass Twitter followings. Yet when we traced the gravitational field lines through the codebase, there they were — quietly holding the architecture together. We call them **Hidden Architects.**

**A note on cross-universe comparison.** Gravity is a *relative signal within each repository*, not an absolute value across repositories. Josh Goldberg's Gravity 100 in eslint and Jordan Liggitt's Gravity 77.3 in Kubernetes are observations from **different universes** — they cannot be directly compared. This is Engineering Relativity (Ch. 8) in action.

However, the distortion is partially mitigated by Gravity's composition. Its three axes — module ownership ratio, design involvement ratio, and cross-cutting reach — are **proportion-based signals**, not absolute volume. Owning 80% of modules in a 50-module project and owning 80% in a 500-module project both register the same Indispensability signal. The ranking captures *who shaped the gravitational field of their universe*, not who works in the "biggest" universe.

Think of it as mapping the brightest star in each galaxy. Some galaxies are larger than others, but in every galaxy, the star that shapes the gravitational field is observable.

> Full analysis: [OSS Gravity Map](https://machuz.github.io/eis/research/oss-gravity-map/analysis/oss-gravity-map-en.html)

---

## What This Is NOT

> *We don't measure engineers. We reveal how software actually works.*

This series uses the word "combat power" (戦闘力) to describe impact. It's a catchy metaphor borrowed from Dragon Ball — but it carries a dangerous implication: that engineers can be ranked on a single axis of strength.

**They can't. And EIS doesn't try to.**

So what *does* it measure? Simple: **in this codebase, how much did you build, how much influence did you leave, and how much of what you wrote is still standing?** That's it. Not "how good an engineer are you" — but "what trace did you leave in this particular universe of code."

True engineering excellence can only be quantified by traces left across *multiple* universes. High impact in one codebase is a local observation. Consistent high impact across different codebases, different teams, different domains — that's reproducible gravity. That's the difference between a bright star in one galaxy and a force of nature.

A few things to keep in mind:

**EIS measures codebase impact, not engineering ability.** An impact of 40 means "on this codebase, this person's code is surviving, shaping architecture, and cleaning up debt." It does *not* mean they are objectively a better engineer than someone at 30. Move them to a different codebase, and the observations might invert. (We call this [Engineering Relativity](https://dev.to/machuz/git-archaeology-8-engineering-relativity-why-the-same-engineer-gets-different-scores-5dnl).)

**Signals without context are dangerous.** A low Survival signal might mean poor design — or it might mean the engineer is actively rewriting legacy code (Rescue style). A strong signal in a poorly designed codebase might mean "nobody can refactor your code away." Always interpret with context.

**Non-code contributions are invisible to git.** Code review quality, mentoring, documentation, psychological safety, domain expertise — these matter enormously but leave no trace in `git log`. EIS captures what git records, nothing more. Using it as a complete evaluation of an engineer would be harmful and wrong.

**It's not a surveillance tool.** EIS is a telescope — it reveals structures that already exist. It doesn't create hierarchies. If it's used to rank and punish rather than to understand and improve, it has failed its purpose.

**Time-decayed survival resists gaming.** You can't inflate your impact with busy work. Only code that remains in the codebase months later counts. The debt cleanup axis makes it structurally impossible to achieve high impact by generating work for others.

The telescope measures the brightness of stars. It doesn't decide which stars deserve to exist.

---

## The Series

This blog series — **Git Archaeology** — applies EIS to real teams and explores what the numbers reveal.

1. **[Measuring Engineering Impact from Git History Alone](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)** — Full 7-axis observation design
2. **[Beyond Individual Scores: Measuring Team Health from Git History](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f)**
3. **[Two Paths to Architect: How Engineers Evolve Differently](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga)**
4. **[Backend Architects Converge: The Sacred Work of Laying Souls to Rest](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d)**
5. **[Timeline: Scores Don't Lie, and They Capture Hesitation Too](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5)**
6. **[Teams Evolve: The Laws of Organization Revealed by Timelines](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4lei)**
7. **[Observing the Universe of Code](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0)**
8. **[Engineering Relativity: Why the Same Engineer Gets Different Scores](https://dev.to/machuz/git-archaeology-8-engineering-relativity-why-the-same-engineer-gets-different-scores-5dnl)**
9. **[Origin: The Big Bang of Code Universes](https://dev.to/machuz/git-archaeology-9-collapse-good-architects-and-black-hole-engineers-1dcn)**
10. **[Dark Matter: The Invisible Gravity](https://dev.to/machuz/git-archaeology-10-dark-matter-the-invisible-gravity-45ne)**
11. **[Entropy: The Universe Always Tends Toward Disorder](https://dev.to/machuz/git-archaeology-11-entropy-the-universe-always-tends-toward-disorder-ak9)**
12. **[Collapse: Good Architects and Black Hole Engineers](https://dev.to/machuz/git-archaeology-12-collapse-good-architects-and-black-hole-engineers-3fed)**
13. **[Cosmology of Code](https://dev.to/machuz/git-archaeology-13-cosmology-of-code-dci)**
14. **[Civilization: Why Only Some Codebases Become Civilizations](https://dev.to/machuz/git-archaeology-14-civilization-why-only-some-codebases-become-civilizations-1fe3)**
15. **[AI Creates Stars, Not Gravity](https://dev.to/machuz/git-archaeology-15-the-age-of-ai-the-starburst-that-code-universes-were-never-prepared-for-o7k)**
16. **[The Engineers Who Shape Gravity](https://dev.to/machuz/git-archaeology-16-the-engineers-who-shape-gravity-3fmi)**

### Series

- **Chapter 0: What If Git History Could Tell You Who Your Strongest Engineers Are?**
- [Chapter 1: Measuring Engineering Impact from Git History Alone](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)
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

[Chapter 1: Individual Observation →](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)

---

## Install

```bash
# Homebrew
brew tap machuz/tap && brew install eis

# Go
go install github.com/machuz/eis/cmd/eis@latest
```

**GitHub**: [eis](https://github.com/machuz/eis)

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png)

---
title: "Git Archaeology #6 — Teams Evolve: The Laws of Organization Revealed by Timelines"
published: true
description: "Chapter 6 of Engineering Impact Signal. When individual changes ripple into team-level shifts, patterns emerge — and those patterns have laws."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/cover-ch6.png?v=4
---

*When individuals change, teams change. Team timelines reveal that these changes follow laws.*

![Team evolution stages — Formation, Growth, Maturity, Scaling](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-iconic.png?v=4)

## Previously

In [Chapter 5](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5), I used `eis timeline` to trace individual timelines — Y.Y.'s departure and machuz's architectural permeation, R.M.'s "hesitation" and return.

But those were individual stories. **A team is a collection of individuals.** When individuals change, the team's nature changes too.

`eis timeline` also auto-generates team-level aggregations. This chapter reads those team timelines, then extracts the **laws of engineer evolution** visible in the data.

---

## Team Timeline: What It Shows

The team timeline classifies the entire team for each period:

- **Character** — Team personality (Spiral, Dwarf, Cluster, Lenticular, Filament, Collision)
- **Structure** — Organization (Architectural Engine, Delivery Team, Maintenance Team, Unstructured)
- **Culture** — Culture (Builder, Stability, Exploration, Firefighting)
- **Phase** — Phase (Mature, Emerging, Declining)
- **Risk** — Risk (Healthy, Design Vacuum, Quality Drift)

Plus Health metrics (Complementarity, Growth Potential, Sustainability, etc.) and Signal Averages per period.

**Individual Role/Style changes surface as team Character/Structure changes.** That's the insight.

---

## Real Data: The Backend Team's Transformation

![Backend Team Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-backend-team-timeline.png?v=4)

(2024-H2 through 2025-H2 lack classification due to insufficient member count. Comparing 2024-H2 and 2026-H1.)

**Lenticular → Spiral. Unstructured → Architectural Engine. Declining → Mature. Design Vacuum → Healthy.**

Every axis improved.

Why? As Chapter 5 showed:

- **2024**: Y.Y. sustained the structure as Architect Builder
- **2025-H1**: Y.Y. declined from Anchor → Fragile. No structure owner
- **2025-H2**: machuz reached Architect Builder
- **2026-H1**: Team reached Spiral / Architectural Engine / Mature / Healthy

**A generational transition at the individual level manifested as a team-level transformation.**

Y.Y.'s departure temporarily created a Design Vacuum. But machuz inherited the Architect role, new members joined, and the team reached a healthier state than before.

Looking at Signal Averages:

![Backend Signals](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-backend-scores.png?v=4)

Design at 36.4 is still low. Because Architect responsibility is concentrated in machuz alone. Other members' Design signals are mostly 0–30.

**The Spiral team's next challenge: distributing design capability.**

---

## Real Data: The Frontend Team's Evolution

Frontend has longer data coverage, making transitions easier to read.

![Frontend Team Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-frontend-team-timeline.png?v=4)

First thing that stands out: **Declining → Mature only happened in 2026-H1.** Later than Backend.

Then **Culture: Stability → Builder**. This is largely R.M.'s influence (Architect from day one, as covered in Chapter 5). R.M.'s continuous involvement in design files shifted team culture from Stability (defensive) to Builder (offensive).

Meanwhile, Risk shifted from **Quality Drift → Design Vacuum**. This looks like a deterioration, but the meaning is different:

- **Quality Drift**: Quality varies across members (many Producers)
- **Design Vacuum**: Designers are scarce (Architect departed or absent)

O. settled into Producer in 2025-H2, and R.M. sometimes oscillates to Anchor. **There aren't always two Architects simultaneously**, hence the Design Vacuum risk.

One more interesting pattern in Frontend:

![Frontend Team Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-frontend-team-timeline.png?v=4)

**In 2025-H1, the team briefly became Cluster / Delivery Team, then reverted.**

What happened? 2025-H1 was when R.M. reached 83.8 as Architect — not only designing but also shipping at high volume, single-handedly driving both design and implementation. O. was simultaneously Anchor (54.3), delivering steadily. **The designer was producing at high throughput while the Anchor sustained delivery.** The team temporarily exhibited Cluster (high-throughput) / Delivery Team characteristics.

But that "maximum output" was temporary. The next half, it reverted to Dwarf / Maintenance.

**Team character fluctuates with individual output on a quarterly basis.** The smaller the team, the more pronounced this effect.

---

## Cross-Domain: Infra and Firmware

### Infra: Filament / Emerging

![Infra & Firmware](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-infra-firmware.png?v=4)

Filament / Exploration / Emerging. **A team still taking shape.**

Every member profiles as Growing or Spread. Zero Architects. Design Vacuum is inevitable.

But Phase: Emerging means **growth is happening**. Not Declining.

### Firmware: Collision

Firmware has only two members. Character: Collision. Culture: Firefighting.

Production 100, Quality 84.6. **Productive, but no design.**

This is a team "responding to problems as they arise." No bandwidth to build structure. No Architect, Design at 0.

**The smaller the team, the more a single person's entry or exit fundamentally changes the team's nature.** If one Architect joined Firmware, it would shift from Collision to Builder.

---

## Now the Main Event: Evolution Models

When you lay individual timelines alongside team timelines, **laws** emerge.

Who evolves how, under what conditions do Roles change, and how does it affect the team?

Here are the evolution models extracted from our timeline data.

---

### Model 1: There's More Than One Path to Architect Manifestation

machuz's Backend timeline:

![machuz Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-machuz-timeline.png?v=4)

**Anchor → Producer → Architect.**

This looks like a "growth staircase," but the reality is different. machuz already had extensive Architect experience from other teams. However, this team had a predecessor Architect (Y.Y.) already in place.

What machuz did was **respect the predecessor's structure while improving it and shipping features at high volume**. The Anchor phase was about understanding the existing structure. The Producer phase was about producing extensively on top of it. Throughout this process, machuz's own architectural vision gradually permeated the codebase.

When Y.Y.'s signals began declining in 2025-H2, machuz's architecture became the structural backbone. EIS captured this as Architect Builder.

This isn't "growing into an Architect." It's **"an existing architectural vision permeating the codebase until the numbers reflect it."**

Meanwhile, R.M.'s Frontend timeline (from Chapter 5):

![Architect by Q4](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-data-architect-quarter.png?v=4)

**Architect by their second quarter.** R.M. also had Architect experience externally. But unlike machuz, they shortened the adaptation period and immediately began designing with their own architectural approach.

**Even among experienced joiners, the speed of manifestation differs.**

1. **Permeation type** — Respect the predecessor's structure, produce on top of it, and gradually infuse your own design (machuz's path)
2. **Immediate type** — Brief Anchor phase, then start designing with your own architecture immediately (R.M.'s path)

The permeation type takes longer but maintains continuity with existing structure. The immediate type is fast but carries **collision risk with the team**, as Chapter 5 showed.

---

### Model 2: Backend Architects Concentrate, Frontend Architects Flow

Touched on in Chapter 5, but timelines make it clearer.

**Backend:**

![BE Architects](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-be-architects.png?v=4)

As an observation, **there was never a period with two simultaneous Architects.** Y.Y. stepped down before machuz's architecture became dominant. Whether this reflects a structural constraint of Backend's single design axis (one DB schema, one API convention) or simply the timing of architectural permeation is hard to determine from this sample alone. But at minimum, BE Architects tend to concentrate.

**Frontend:**

![FE Architects](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-fe-architects.png?v=4)

In Frontend, when R.M. became Architect, O. dropped to Anchor. **At first glance, it looks like the same "one seat" pattern.**

But quarterly data (Chapter 5) reveals:

![Simultaneous Architects](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-data-simultaneous.png?v=4)

**In 2025-Q2, two Architects existed simultaneously.**

A phenomenon not observed in Backend happened in Frontend.

The hypothesis: **Backend shares a single structure.** Database schema, API design conventions, shared libraries. Design decisions converge on a single axis, so Architects tend to concentrate.

**Frontend supports parallel structures.** Component design, state management, routing. Independent design decisions are possible in each domain, allowing multiple Architects to coexist.

That said, Backend's "one at a time" pattern could be a sample size limitation. With a larger team, multiple Backend Architects might coexist.

The **Design Vacuum risk** in Frontend's team timeline is the flip side of this fluidity. Even if multiple Architects are possible, they aren't always present simultaneously. The moment one transitions to Producer, Design Vacuum appears.

---

### Model 3: Producer Is Metabolism

Look at O.'s transitions again:

![O. Transitions](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-data-engineer-j-transitions.png?v=4)

**An Architect who finishes building structure becomes a Producer.**

This isn't regression. **It's metabolism.**

An Architect builds structure. The structure is complete. Then the Architect's work diminishes. Changes to design files become unnecessary, Design signal drops, and Role naturally shifts to Producer.

Then they enter a phase of "producing on top of the structure."

The same pattern can be predicted for machuz's Backend. Currently maintaining 92 as Architect Builder, but once the structure stabilizes, machuz will likely transition to Producer too. **When that time comes, the next Architect will be needed.**

This is where Backend's team Health metric becomes meaningful:

![Growth Potential](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-data-health.png?v=4)

Growth Potential at 20. **The seeds of the next-generation Architect are still weak.** This is Backend's medium-term risk.

---

### Model 4: The Founding Architect Lifecycle

In the Frontend 6-month timeline, one engineer has a uniquely dramatic trajectory.

![X.](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-engineer-k.png?v=4)

**Impact 87.8 in 2024-H1. Architect Builder.** Production 81, Survival 100, Design 100.

The next half: Impact 14.6. After that, effectively zero.

This is the **founding Architect lifecycle**.

X. built Frontend's initial structure. In 2024-H1, the codebase's blame was dominated by X.. Architect Builder was the natural result.

But as the team grew and other engineers (R.M., O.) joined and began rewriting the structure, X.'s Survival dropped rapidly. Blame lines were replaced by other members.

**A founding Architect's impact drops as the team grows.**

This isn't failure. It's **proof of success.** Other engineers are building on top of the structure you created alone. That's why the impact drops.

![Gravity Transfer](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-gravity-transfer.png?v=4)

**X.'s impact transferred to Engineers I and J.** The total isn't conserved, but the generational transfer of structural influence is clear.

This is a different kind of "exit" from Y.Y.'s departure (Chapter 5). Y.Y. exited due to **team departure**. X. exited due to **team growth**.

EIS captures both.

---

### Model 5: Builder Is a Prerequisite for Architect

Timeline data shows that **reaching Architect almost always requires passing through Builder**:

![Evolution Paths](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-evolution-paths.png?v=4)

machuz passed through Anchor Builder before reaching Architect. R.M. went Architect Balanced → Architect Builder. O. also went Architect Balanced → Architect Builder.

**Engineers who can't become Builders don't become Architects.**

What is a Builder? An engineer who writes new code that survives. Not modifying existing code (Balanced), but adding new structure.

This aligns with the essence of "designing." Design isn't "modifying what exists" — it's "creating new structure."

It's no coincidence that the moment Frontend's Culture shifted from Stability → Builder in the team timeline, the Phase changed from Declining → Mature.

**Teams without Builder culture can't mature.** Defense alone leads to decline.

---

### Model 6: Producer Vacuum

What happens when a team has no Producers?

Look at Backend 2024-H2. machuz as Anchor Builder (76.4), Y.Y. as Architect Builder (84.1). But zero Producers.

That period's team classification:

![Producer Vacuum](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-data-producer-vacuum.png?v=4)

**The Architect builds structure, the Anchor maintains it. But nobody is producing on top of that structure.**

Structure without production. That's Producer Vacuum.

Compare with Backend 2026-H1:

![Effective Members](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-data-effective-members.png?v=4)

machuz (Architect Builder) plus multiple Anchors/Producers. People producing on top of the structure. That's why it reached Spiral / Architectural Engine.

**Architects alone can't make a team function.** You need Producers who use the structure the Architect built. Only then does the team work.

---

### Model 7: The Producer Phase Is Architectural Groundwork

Look at machuz's timeline one more time:

![machuz Phases](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-data-machuz-phases.png?v=4)

**Architect Builder appeared in the numbers after passing through a Producer phase in 2025-H1.**

machuz's Producer phase wasn't a "learning period." It was **a period of shipping features at high volume on top of the predecessor Architect's (Y.Y.'s) structure, while gradually weaving in their own design philosophy through improvements**.

By using the existing structure extensively, you internalize both its limitations and its possibilities. And within those improvements, you embed your own architecture.

When 92.5 as Architect Builder appeared in 2025-H2, that was the moment the permeation crossed a threshold. Y.Y.'s signal decline (the structural backbone shifting) and machuz's Design 100 (design changes reaching architecture files) happened simultaneously.

**The Producer phase can serve as groundwork for permeating your architecture into the codebase.** On the surface it looks like Producer work, but beneath the surface, a design philosophy transplant is underway.

Without this "groundwork" — starting to design without deep understanding of the existing structure — risks "armchair design." Without experience using the structure, you build structures that are hard to use.

R.M. could become Architect immediately after joining because, we can infer, **they had extensive experience with similar architectures at their previous position.** They had used comparable structures extensively elsewhere. That's why they could quickly identify "what needs to be designed" in a new environment. The groundwork period was unnecessary.

---

## The Big Picture

Here are the evolution models extracted from timelines:

![Evolution Model](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch6-evolution-model.png?v=4)

These laws were inductively derived from our team's real data. Whether the same laws hold for other teams is unknown.

But with `eis timeline`, **you can discover your own team's laws yourself.**

---

## Using Team Timelines

### 1. Organizational Review

```bash
❯ eis timeline --span 6m --periods 0 --recursive ~/workspace
```

Generate half-year team timelines and track Phase / Risk trends. If Declining persists, something needs to change.

### 2. Architect Planning

When a team has no Architect (Design Vacuum), use timelines to identify who could become one:

- Do they have Builder experience?
- Have they been through a Producer phase?
- Have they understood the structure as an Anchor?

### 3. Detecting Producer Vacuum

If a team has an Architect but isn't functioning, suspect Producer Vacuum. Structure exists but nobody's producing on it. Check individual timelines — if Producers are absent, reassess resource allocation.

### 4. Evaluating Founding Architects Correctly

If you find a founding member whose signals are declining, **determine whether it's failure or success.** Signal decline due to team growth should be recognized as an achievement, not a problem.

---

## What This Discovery Means

Chapter 5 gained the time axis. Chapter 6 **gained laws**.

Snapshots show "now." Timelines show "change." Laws predict "what happens next."

- When machuz transitions to Producer, the next Architect will be needed
- If Frontend's Design Vacuum persists, either wait for R.M.'s Architect return or develop a new Architect
- For Infra to progress from Emerging → Mature, it needs Builders first

**Deriving laws from cold numbers, then reading the future from those laws.** That's the true power of the timeline.

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full.png?v=2)

**GitHub**: [eis](https://github.com/machuz/eis) — CLI tool, formulas, and methodology all open source. `brew tap machuz/tap && brew install eis` to install.


If this was useful: [❤️ Sponsor on GitHub](https://github.com/sponsors/machuz)

---

### Series

- [Chapter 0: What If Git History Could Tell You Who Your Strongest Engineers Are?](https://dev.to/machuz/git-archaeology-0-what-if-git-history-could-tell-you-who-your-strongest-engineers-are-5397)
- [Chapter 1: Measuring Engineering Impact from Git History Alone](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)
- [Chapter 2: Beyond Individual Scores: Measuring Team Health from Git History](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f)
- [Chapter 3: Two Paths to Architect: How Engineers Evolve Differently](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga)
- [Chapter 4: Backend Architects Converge: The Sacred Work of Laying Souls to Rest](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d)
- [Chapter 5: Timeline: Signals Don't Lie, and They Capture Hesitation Too](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5)
- **Chapter 6: Teams Evolve: The Laws of Organization Revealed by Timelines**
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

← [Chapter 5: Timeline](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5) | [Chapter 7: Observing the Universe of Code →](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0)

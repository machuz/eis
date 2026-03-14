---
title: "Git Archaeology #6 — Teams Evolve: The Laws of Organization Revealed by Timelines"
published: true
description: "Chapter 6 of Engineering Impact Score. When individual changes ripple into team-level shifts, patterns emerge — and those patterns have laws."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/cover-ch6.svg
---

*When individuals change, teams change. Team timelines reveal that these changes follow laws.*

## Previously

In [Chapter 5](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5), I used `eis timeline` to trace individual timelines — Engineer F's departure and machuz's architectural permeation, Engineer I's "hesitation" and return.

But those were individual stories. **A team is a collection of individuals.** When individuals change, the team's nature changes too.

`eis timeline` also auto-generates team-level aggregations. This chapter reads those team timelines, then extracts the **laws of engineer evolution** visible in the data.

---

## Team Timeline: What It Shows

The team timeline classifies the entire team for each period:

- **Character** — Team personality (Elite, Guardian, Factory, Balanced, Explorer, Firefighting)
- **Structure** — Organization (Architectural Engine, Delivery Team, Maintenance Team, Unstructured)
- **Culture** — Culture (Builder, Stability, Exploration, Firefighting)
- **Phase** — Phase (Mature, Emerging, Declining)
- **Risk** — Risk (Healthy, Design Vacuum, Quality Drift)

Plus Health metrics (Complementarity, Growth Potential, Sustainability, etc.) and Score Averages per period.

**Individual Role/Style changes surface as team Character/Structure changes.** That's the insight.

---

## Real Data: The Backend Team's Transformation

![Backend Team Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch6-backend-team-timeline.svg)

(2024-H2 through 2025-H2 lack classification due to insufficient member count. Comparing 2024-H2 and 2026-H1.)

**Balanced → Elite. Unstructured → Architectural Engine. Declining → Mature. Design Vacuum → Healthy.**

Every axis improved.

Why? As Chapter 5 showed:

- **2024**: Engineer F sustained the structure as Architect Builder
- **2025-H1**: Engineer F declined from Anchor → Fragile. No structure owner
- **2025-H2**: machuz reached Architect Builder
- **2026-H1**: Team reached Elite / Architectural Engine / Mature / Healthy

**A generational transition at the individual level manifested as a team-level transformation.**

Engineer F's departure temporarily created a Design Vacuum. But machuz inherited the Architect role, new members joined, and the team reached a healthier state than before.

Looking at Score Averages:

![Backend Scores](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch6-backend-scores.svg)

Design at 36.4 is still low. Because Architect responsibility is concentrated in machuz alone. Other members' Design scores are mostly 0–30.

**The Elite team's next challenge: distributing design capability.**

---

## Real Data: The Frontend Team's Evolution

Frontend has longer data coverage, making transitions easier to read.

![Frontend Team Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch6-frontend-team-timeline.svg)

First thing that stands out: **Declining → Mature only happened in 2026-H1.** Later than Backend.

Then **Culture: Stability → Builder**. This is largely Engineer I's influence (Architect from day one, as covered in Chapter 5). Engineer I's continuous involvement in design files shifted team culture from Stability (defensive) to Builder (offensive).

Meanwhile, Risk shifted from **Quality Drift → Design Vacuum**. This looks like a deterioration, but the meaning is different:

- **Quality Drift**: Quality varies across members (many Producers)
- **Design Vacuum**: Designers are scarce (Architect departed or absent)

Engineer J settled into Producer in 2025-H2, and Engineer I sometimes oscillates to Anchor. **There aren't always two Architects simultaneously**, hence the Design Vacuum risk.

One more interesting pattern in Frontend:

![Frontend Team Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch6-frontend-team-timeline.svg)

**In 2025-H1, the team briefly became Factory / Delivery Team, then reverted.**

What happened? 2025-H1 was when Engineer I scored 83.8 as Architect — not only designing but also shipping at high volume, single-handedly driving both design and implementation. Engineer J was simultaneously Anchor (54.3), delivering steadily. **The designer was producing at high throughput while the Anchor sustained delivery.** The team temporarily exhibited Factory (high-throughput) / Delivery Team characteristics.

But that "maximum output" was temporary. The next half, it reverted to Guardian / Maintenance.

**Team character fluctuates with individual output on a quarterly basis.** The smaller the team, the more pronounced this effect.

---

## Cross-Domain: Infra and Firmware

### Infra: Explorer / Emerging

![Infra & Firmware](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch6-infra-firmware.svg)

Explorer / Exploration / Emerging. **A team still taking shape.**

Every member scores as Growing or Spread. Zero Architects. Design Vacuum is inevitable.

But Phase: Emerging means **growth is happening**. Not Declining.

### Firmware: Firefighting

Firmware has only two members. Character: Firefighting. Culture: Firefighting.

Production 100, Quality 84.6. **Productive, but no design.**

This is a team "responding to problems as they arise." No bandwidth to build structure. No Architect, Design at 0.

**The smaller the team, the more a single person's entry or exit fundamentally changes the team's nature.** If one Architect joined Firmware, it would shift from Firefighting to Builder.

---

## Now the Main Event: Evolution Models

When you lay individual timelines alongside team timelines, **laws** emerge.

Who evolves how, under what conditions do Roles change, and how does it affect the team?

Here are the evolution models extracted from our timeline data.

---

### Model 1: There's More Than One Path to Architect Manifestation

machuz's Backend timeline:

![machuz Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch6-machuz-timeline.svg)

**Anchor → Producer → Architect.**

This looks like a "growth staircase," but the reality is different. machuz already had extensive Architect experience from other teams. However, this team had a predecessor Architect (Engineer F) already in place.

What machuz did was **respect the predecessor's structure while improving it and shipping features at high volume**. The Anchor phase was about understanding the existing structure. The Producer phase was about producing extensively on top of it. Throughout this process, machuz's own architectural vision gradually permeated the codebase.

When Engineer F's scores began declining in 2025-H2, machuz's architecture became the structural backbone. EIS captured this as Architect Builder.

This isn't "growing into an Architect." It's **"an existing architectural vision permeating the codebase until the numbers reflect it."**

Meanwhile, Engineer I's Frontend timeline (from Chapter 5):

```
2024-Q3     56.1  Anchor   Balanced
2024-Q4     75.7  Architect Balanced
```

**Architect by their second quarter.** Engineer I also had Architect experience externally. But unlike machuz, they shortened the adaptation period and immediately began designing with their own architectural approach.

**Even among experienced joiners, the speed of manifestation differs.**

1. **Permeation type** — Respect the predecessor's structure, produce on top of it, and gradually infuse your own design (machuz's path)
2. **Immediate type** — Brief Anchor phase, then start designing with your own architecture immediately (Engineer I's path)

The permeation type takes longer but maintains continuity with existing structure. The immediate type is fast but carries **collision risk with the team**, as Chapter 5 showed.

---

### Model 2: Backend Architects Concentrate, Frontend Architects Flow

Touched on in Chapter 5, but timelines make it clearer.

**Backend:**

![BE Architects](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch6-be-architects.svg)

As an observation, **there was never a period with two simultaneous Architects.** Engineer F stepped down before machuz's architecture became dominant. Whether this reflects a structural constraint of Backend's single design axis (one DB schema, one API convention) or simply the timing of architectural permeation is hard to determine from this sample alone. But at minimum, BE Architects tend to concentrate.

**Frontend:**

![FE Architects](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch6-fe-architects.svg)

In Frontend, when Engineer I became Architect, Engineer J dropped to Anchor. **At first glance, it looks like the same "one seat" pattern.**

But quarterly data (Chapter 5) reveals:

```
2025-Q2   Engineer I  73.2  Architect    Engineer J  63.8  Architect
```

**In 2025-Q2, two Architects existed simultaneously.**

A phenomenon not observed in Backend happened in Frontend.

The hypothesis: **Backend shares a single structure.** Database schema, API design conventions, shared libraries. Design decisions converge on a single axis, so Architects tend to concentrate.

**Frontend supports parallel structures.** Component design, state management, routing. Independent design decisions are possible in each domain, allowing multiple Architects to coexist.

That said, Backend's "one at a time" pattern could be a sample size limitation. With a larger team, multiple Backend Architects might coexist.

The **Design Vacuum risk** in Frontend's team timeline is the flip side of this fluidity. Even if multiple Architects are possible, they aren't always present simultaneously. The moment one transitions to Producer, Design Vacuum appears.

---

### Model 3: Producer Is Metabolism

Look at Engineer J's transitions again:

```
Architect → Anchor → Architect → Producer → Producer → Producer
```

**An Architect who finishes building structure becomes a Producer.**

This isn't regression. **It's metabolism.**

An Architect builds structure. The structure is complete. Then the Architect's work diminishes. Changes to design files become unnecessary, Design score drops, and Role naturally shifts to Producer.

Then they enter a phase of "producing on top of the structure."

The same pattern can be predicted for machuz's Backend. Currently maintaining 92 as Architect Builder, but once the structure stabilizes, machuz will likely transition to Producer too. **When that time comes, the next Architect will be needed.**

This is where Backend's team Health metric becomes meaningful:

```
Growth Potential: 20.0
```

Growth Potential at 20. **The seeds of the next-generation Architect are still weak.** This is Backend's medium-term risk.

---

### Model 4: The Founding Architect Lifecycle

In the Frontend 6-month timeline, one engineer has a uniquely dramatic trajectory.

```
--- Engineer K (Frontend) ---
2024-H1     87.8  Architect Builder
2024-H2     14.6  — —
2025-H1      7.1  — — Silent
2025-H2      3.2  — —
2026-H1      3.2  — —
```

**Total 87.8 in 2024-H1. Architect Builder.** Production 81, Survival 100, Design 100.

The next half: Total 14.6. After that, effectively zero.

This is the **founding Architect lifecycle**.

Engineer K built Frontend's initial structure. In 2024-H1, the codebase's blame was dominated by Engineer K. Architect Builder was the natural result.

But as the team grew and other engineers (Engineer I, Engineer J) joined and began rewriting the structure, Engineer K's Survival dropped rapidly. Blame lines were replaced by other members.

**A founding Architect's score drops as the team grows.**

This isn't failure. It's **proof of success.** Other engineers are building on top of the structure you created alone. That's why the score drops.

```
Engineer K:  87.8 → 14.6 → 7.1 → 3.2 → 3.2
Engineer I:   — → 72.7 → 83.8 → 85.1 → 78.1
Engineer J:  25.9 → 74.9 → 54.3 → 38.6 → 54.2
```

**Engineer K's score transferred to Engineers I and J.** The total isn't conserved, but the generational transfer of structural influence is clear.

This is a different kind of "exit" from Engineer F's departure (Chapter 5). Engineer F exited due to **team departure**. Engineer K exited due to **team growth**.

EIS captures both.

---

### Model 5: Builder Is a Prerequisite for Architect

Timeline data shows that **reaching Architect almost always requires passing through Builder**:

```
machuz:     Anchor → Anchor Builder → Producer Balanced → Architect Builder
Engineer I: Anchor Balanced → Architect Balanced → Architect Builder
Engineer J: Anchor Growing → Architect Balanced → Architect Builder
Engineer F: (first appearance) Architect Builder
```

machuz passed through Anchor Builder before reaching Architect. Engineer I went Architect Balanced → Architect Builder. Engineer J also went Architect Balanced → Architect Builder.

**Engineers who can't become Builders don't become Architects.**

What is a Builder? An engineer who writes new code that survives. Not modifying existing code (Balanced), but adding new structure.

This aligns with the essence of "designing." Design isn't "modifying what exists" — it's "creating new structure."

It's no coincidence that the moment Frontend's Culture shifted from Stability → Builder in the team timeline, the Phase changed from Declining → Mature.

**Teams without Builder culture can't mature.** Defense alone leads to decline.

---

### Model 6: Producer Vacuum

What happens when a team has no Producers?

Look at Backend 2024-H2. machuz as Anchor Builder (76.4), Engineer F as Architect Builder (84.1). But zero Producers.

That period's team classification:

```
Character: Balanced
Phase: Declining
```

**The Architect builds structure, the Anchor maintains it. But nobody is producing on top of that structure.**

Structure without production. That's Producer Vacuum.

Compare with Backend 2026-H1:

```
Effective Members: 5
Total: 48.3 (average)
```

machuz (Architect Builder) plus multiple Anchors/Producers. People producing on top of the structure. That's why it reached Elite / Architectural Engine.

**Architects alone can't make a team function.** You need Producers who use the structure the Architect built. Only then does the team work.

---

### Model 7: The Producer Phase Is Architectural Groundwork

Look at machuz's timeline one more time:

```
2024-H2     76.4  Anchor  Builder
2025-H1     58.4  Producer Balanced
2025-H2     92.5  Architect Builder
```

**Architect Builder appeared in the numbers after passing through a Producer phase in 2025-H1.**

machuz's Producer phase wasn't a "learning period." It was **a period of shipping features at high volume on top of the predecessor Architect's (Engineer F's) structure, while gradually weaving in their own design philosophy through improvements**.

By using the existing structure extensively, you internalize both its limitations and its possibilities. And within those improvements, you embed your own architecture.

When 92.5 as Architect Builder appeared in 2025-H2, that was the moment the permeation crossed a threshold. Engineer F's score decline (the structural backbone shifting) and machuz's Design 100 (design changes reaching architecture files) happened simultaneously.

**The Producer phase can serve as groundwork for permeating your architecture into the codebase.** On the surface it looks like Producer work, but beneath the surface, a design philosophy transplant is underway.

Without this "groundwork" — starting to design without deep understanding of the existing structure — risks "armchair design." Without experience using the structure, you build structures that are hard to use.

Engineer I could become Architect immediately after joining because, we can infer, **they had extensive experience with similar architectures at their previous position.** They had used comparable structures extensively elsewhere. That's why they could quickly identify "what needs to be designed" in a new environment. The groundwork period was unnecessary.

---

## The Big Picture

Here are the evolution models extracted from timelines:

```
┌──────────────────────────────────────────────────────────┐
│              Evolution Model Overview                    │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  [Growing] → [Anchor] → [Producer] → [Architect]        │
│                  ↑            │            │             │
│                  │            │   structure │             │
│                  │            ←─────────────┘             │
│                  │       (metabolism: back to Producer)   │
│                                                          │
│  * Permeation: respect existing, produce, then permeate   │
│                [Anchor] → [Producer] → [Architect]        │
│  * Immediate:  [Anchor(brief)] → [Architect] direct       │
│  * Founding:   [Architect] → score decline (= success)    │
│                                                          │
├──────────────────────────────────────────────────────────┤
│  BE: Architect seats tend to concentrate (observed: 1)    │
│  FE: Architect seats are fluid (observed: 1–2)           │
├──────────────────────────────────────────────────────────┤
│  Builder prerequisite: Builder experience needed for      │
│                        Architect                         │
│  Producer fuel: using structure deeply powers design      │
│  Producer Vacuum: no producers = structure sits idle      │
└──────────────────────────────────────────────────────────┘
```

These laws were inductively derived from our team's real data. Whether the same laws hold for other teams is unknown.

But with `eis timeline`, **you can discover your own team's laws yourself.**

---

## Using Team Timelines

### 1. Organizational Review

```bash
eis timeline --span 6m --periods 0 --recursive ~/workspace
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

If you find a founding member whose scores are declining, **determine whether it's failure or success.** Score decline due to team growth should be recognized as an achievement, not a problem.

---

## What This Discovery Means

Chapter 5 gained the time axis. Chapter 6 **gained laws**.

Snapshots show "now." Timelines show "change." Laws predict "what happens next."

- When machuz transitions to Producer, the next Architect will be needed
- If Frontend's Design Vacuum persists, either wait for Engineer I's Architect return or develop a new Architect
- For Infra to progress from Emerging → Mature, it needs Builders first

**Deriving laws from cold numbers, then reading the future from those laws.** That's the true power of the timeline.

---

**GitHub:** [machuz/engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLI tool, formulas, and methodology are all open source. Install with `brew tap machuz/tap && brew install eis`.

If this article was useful:

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

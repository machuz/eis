---
title: "Git Archaeology #3 — Two Paths to Architect: How Engineers Evolve Differently"
published: true
description: "Chapter 3 of Engineering Impact Score. Architects don't come from one mold — inheritance vs. emergence, and why cold numbers tell the most human stories."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/cover-ch3.png
---

*Cold numbers, it turns out, tell the most human stories.*

## Previously

In [Chapter 1](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c), I introduced a 7-axis scoring model for individual engineers. In [Chapter 2](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-51il), I aggregated those scores into team-level health metrics.

But staring at the data long enough, I noticed something else.

**Architects don't all come from the same mold.**

There are two distinct evolution paths. And understanding them changes how you read the numbers — and the people behind them.

---

## Score Intuition

Before diving in, let me share the gut-feel calibration I've developed after running EIS across many engineers:

- **10–20**: Contractors who didn't work out, or hires we regretted
- **Low 20s**: Engineers I felt "meh" about
- **30+**: People I actively wanted to work with
- **40+**: Senior-level

The scores consistently matched my intuition.

**40 is senior.** That's how strict this model is. Hitting 40+ across seven axes requires sustained, high-quality contribution. Real senior engineering is hard.

---

## A Team Portrait

Looking at one frontend team's EIS data, I started seeing the evolution paths emerge.

---

**Engineer A** — Architect / Builder / Active

The structural center of the team.

![Engineer Archetypes](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-engineer-profiles.png?v=2)

This is the team's gravitational core.

---

**Engineer B** — Anchor / Mass / Active

Producing like crazy. But Robust is only 11.

Most of their code gets rewritten by Engineer A.

But **they keep producing until 11% survives**.

Anchor means they're trying to guard quality. Mass means their precision is still rough. **The struggle and the effort — both show up in the numbers.**

This isn't just "immature." They're in the early stage of an Anchor-to-Architect path. They're building the instincts for **Inheritance Architect**.

---

**Engineer C** — Producer / Balanced / Active

The one who listens to everyone, synthesizes feedback, and writes code that reflects the team's consensus.

That's why their code is robust.

**Producer means they're shipping real code** — not just attending meetings. The label doesn't attach unless you're actually producing. It's proof of standing shoulder-to-shoulder with the team.

They haven't created gravitational code yet — the kind everyone else builds on top of. But they're pushing the team forward as an adaptive Producer.

---

**Engineer D** — Producer / Emergent / Active (High Gravity)

The only one outside the Architect **generating real gravity**.

Engineers who can create gravity are rare. Genuinely rare.

But Dormant is high — **much of their code hasn't been touched by others yet**. They built ahead of the team.

And Robust is low.

**This is the number that shows Engineer D's struggle most clearly.** Their gravitational field is getting overwritten. The code they shaped is being reshaped.

But it's not all gone. Some of it remains. You can't feel that intuitively — but the metrics pick it up.

In fact, while writing this article and analyzing Engineer D's state, I realized something.

"High Gravity + producing code + but low Robust" — this combination wasn't captured by existing Styles. It was being labeled Balanced, but that's wrong. **This is exactly what an Emergent Architect candidate looks like.**

So in v0.11.0, I added **Style: Emergent**.

Emergent means "not yet fully formed, but coming into being."

- Proposing new structure, currently colliding with existing structure
- Not yet battle-tested by the team (low Robust), but generating gravity
- The pre-stage before evolving into Architect

**The insight from this article became an evolution of the metrics themselves.**

---

**Engineer E** — Producer / Churn / —

The more they write, the more debt they create.

Their code is systematically replaced.

**Style: Churn says it all.** Producing, but not surviving. This is another reality the metrics surface.

---

## Numbers Tell Stories

Here's what struck me looking at this data.

**These metrics visualize the stories that accumulate day by day.**

Engineer B: "Producing like mad, but only 11% survives. And yet — still producing." That's pain. But it's also grit. Both show up in the numbers.

Engineer D: "Has gravity. But keeps getting overwritten. But not completely erased." That's pain too. But something remains. Something you can't feel intuitively — but the metrics catch it.

**Cold numbers turn out to be the most emotional.**

Daily effort etched into the codebase. Struggle. Persistence. Things that slowly remain. These metrics make that visible.

---

## Two Kinds of Gravity

Here's an important nuance.

**Gravity comes in two flavors.**

EIS calculates Gravity as:

![Gravity Formula](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch1-formula-gravity.png?v=2)

For Engineer D:

![Engineer D Gravity Calculation](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-gravity-calc-d.png?v=2)

Most of that 68 comes from **Indispensability (sole ownership) and Breadth (spread)**. Design is only 5. This isn't "structural center" — it's "widely held by one person."

For Engineer A:

![Engineer A Gravity Calculation](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-gravity-calc-a.png?v=2)

Design 100 drives this Gravity. They're creating the structural center.

---

Two types of gravity:

**Structural Gravity**

Design-based. Creating the center of the architecture. Other code radiates outward from there.

**Occupancy Gravity**

Indispensability-based. Widely held by one person. Nobody else has touched it.

---

When we say "engineers who create gravity are rare" or "great engineers generate gravity wherever they go" — **these two types are qualitatively different**.

Engineer A has structural gravity. Engineer D has occupancy gravity.

But this isn't a dismissal of Engineer D.

**Occupancy gravity is the precursor to structural gravity.**

First you hold wide territory. Nobody else has touched it yet. Collisions and rewrites happen. Then Design rises.

The path is: **Occupancy Gravity → Structural Gravity**.

Emergent Architects walk exactly this road.

---

## The Danger of Untested Code

There's another important state EIS makes visible.

**State: Fragile.**

When code "survives," it can mean two different things:

- **Robust Survival**: Code remains in files that others frequently modify — survived under pressure
- **Dormant Survival**: Code remains in modules nobody touches — simply untouched

The latter isn't durability. **It's just neglect.**

Fragile state is detected when:

- Dormant ratio is 80%+ (almost nobody touches it)
- Indispensability is high (held by one person)
- Production is low (not actively producing)

In other words: **"Code remains, but only because nobody touched it."**

The moment change pressure arrives, this code may collapse. It survives not because it's high quality, but because **it hasn't been tested**.

High Survival feels reassuring, but without checking the Dormant ratio, you miss the truth. EIS separates "dormant code" from "code that survived change pressure" — making both visible.

---

## What the Team Metrics Show

This team's metrics include a warning:

![Data Warning](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-data-warning.png?v=2)

**Engineer D's struggle surfaces as a Warning.** High gravity, but low robust survival.

Yet this team is strong:

![Team Metrics](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-data-team-metrics.png?v=2)

Why?

Because **the current structure and the next structure coexist**.

---

## Why These Three Roles Exist

Step back and think about it from an ecological perspective.

A codebase ecosystem requires three roles:

- **Architect** = **Creates** structure. Shapes the terrain, builds an environment where others can thrive
- **Anchor** = **Maintains** structure. Stabilizes the soil, prevents ecosystem collapse
- **Producer** = **Extends** structure. Thrives on existing ground, generates user value

Remove any one and the ecosystem breaks. Architects alone create structure but no features. Producers alone build features but structure crumbles. Without Anchors, both structure and features rot over time.

**A healthy codebase has all three coexisting.**

---

## Two Evolution Paths

Now for the main thesis.

**Architects come in two schools.**

---

**Inheritance Architect**

Evolves from Anchor.

Traits:

- Deep understanding of existing structure
- Knows the real-world constraints
- Better at refinement than destruction
- Strengthens the system without breaking it

This path is especially strong in **backend**.

Backend tends to have clearer responsibility separation, domain boundaries, and "closer to correct" designs.

So **copying good structure, guarding it, and purifying it** has high value.

---

**Emergent Architect**

Evolves from High-Gravity Producer.

Traits:

- Creates new structure rather than inheriting existing one
- Early friction is high
- Gets overwritten, collides with others
- But eventually creates a new center

This path is especially interesting in **frontend**.

Frontend doesn't have the same "one correct answer" as backend.

---

## The Beauty of Frontend

This is important.

Frontend — unlike backend — deals with UX, layout, state management, interaction, abstraction granularity. **Multiple aesthetics coexist.**

So a frontend Architect who only **inherits the correct answer** is weak.

What's needed is someone who **brings a different structural proposal, clashes with existing structure, and forges a stronger gravitational field**.

**In frontend, the collision of ideas is beautiful.**

---

## Reading Engineer D

That earlier pattern:

![Engineer D Pattern](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-data-pattern.png?v=2)

isn't just immaturity.

It can also mean: **proposing new structure that competes with the existing one**.

So "Architect candidate" is imprecise. More accurately: **Emergent Architect candidate**.

---

## Reading Engineer B

Engineer B shows:

![Engineer B Pattern](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-data-anchor-mass.png?v=2)

This doesn't mean "unlikely to become Architect."

It means: **has the foundation for Inheritance Architect, but style is still rough**.

- Positioned on the side of guarding structure
- Has quality awareness
- But still producing with too much collision

They're at the **early stage of Inheritance Architect**.

---

## What Comes After the Struggle

These metrics reflect effort and impact on the current codebase.

How things evolve from here depends on the ecosystem.

But engineers who can produce this kind of gravity — even if their score temporarily drops, even if they go through painful stretches — **will eventually settle into a score commensurate with their intellectual potential**.

If anyone on this team will hit 60–80 and challenge Engineer A's dominance through domain knowledge and craft — **it's Engineer D and no one else**.

---

## Career Models Aren't Linear

Here's the synthesis.

**Backend-style evolution:**

![Backend Evolution Path](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-diagram-be-evolution.png?v=2)

**Frontend-style evolution:**

![Frontend Evolution Path](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-diagram-fe-evolution.png?v=2)

Career paths aren't a single line. They're a **branching evolution model**.

And the two types aren't opposites — they're **complementary**.

The ideal team probably looks like:

- Inheritance Architect
- Emergent Architect candidate
- Anchor
- Producer

In other words: **the power to guard, the power to break and rebuild, the power to stabilize, the power to ship**.

---

## The Danger of Producer-Only Teams

On the flip side, there's a dangerous team composition.

**Everyone is a Producer.**

It looks active. Code ships daily. PRs merge constantly.

But:

- **Nobody touches the design layer** → tacit knowledge accumulates
- **Nobody pays down debt** → code rots in 3 months
- **Nobody guards quality** → fix rate keeps climbing

Everyone writes, nobody cleans up. Everyone pushes forward, nobody solidifies the foundation.

EIS makes this instantly visible:

![Producer-Only Warning](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-producer-warning.png?v=2)

No Architect. No Anchor. No Cleaner.

**Complementarity score tanks.**

Production is high, but Survival is low across the board. Quality is scattered. In six months, you're left with a mountain of tech debt.

What this team needs isn't another Producer. It's **an Architect — or at least an Anchor**.

---

## This Team's Composition

With this lens, the team looks cleaner:

- **Engineer A**: Architect holding the current structure
- **Engineer B**: Anchor who can grow toward Inheritance Architect
- **Engineer C**: Adaptive Producer pushing the team forward
- **Engineer D**: Emergent Architect candidate with high gravity

This isn't "one Architect + followers."

**The current structure and the next structure coexist.**

That's strong.

---

## What This Discovery Means

This discovery elevates EIS's career model.

Not a single line — a **branching evolution model**.

And that's also a theory of team design.

---

**GitHub:** [machuz/engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLI tool, formulas, and methodology. `brew tap machuz/tap && brew install eis` and you're set.

If this was useful:

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

---
title: "Git Archaeology #7 — Observing the Universe of Code"
published: true
description: "Chapter 7 of Engineering Impact Score. Codebases have gravitational structures. Great engineers don't just write code — they bend the gravity of codebases."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/engineering-impact-framework-diagram-fixed.png
---

*Great engineers don't just write code. They bend the gravity of codebases.*

## Previously

In [Chapter 6](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4jf3), I read team timelines and found that individual Role/Style changes surface as team-level Character/Structure shifts — and those shifts follow laws.

But that was still about "reading analysis results."

This chapter is different. It's about the sensation I arrived at after building EIS — **the universe-like structure of codebases**, and what it means to make that structure observable.

---

## HTML Dashboard: The Experience of Looking at Data

First, the practical part. `eis timeline --format html` now outputs an interactive dashboard that's actually useful.

![Timeline HTML Dashboard](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/timeline-html-output.png?v=0.11.0)

```bash
eis timeline --format html --output timeline.html --recursive ~/workspace
```

Chart.js-based line charts show individual and team score trajectories, health metrics, membership composition, and classification changes at a glance. Tooltips display Role/Style/State with confidence scores. Transition markers highlight exactly when changes happened.

For a quick terminal check, there's also `--format ascii`:

![Timeline ASCII Output](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/timeline-ascii-output.png?v=0.11.0)

What makes this powerful is that **you can look at this screen alongside an AI**.

Open the HTML in a browser, ask an AI "What happened to this team in 2024-H2?" The AI can see the numbers, the role transitions, the health metric movements — formulate hypotheses from the data and offer interpretations. This kind of experience was difficult with terminal output alone.

The team Health Metrics view is particularly interesting. Complementarity, Growth Potential, Sustainability, Debt Balance — you can see at a glance how these evolve across periods.

![Team Health Metrics](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/team-health-metrics-output.png?v=0.11.0)

---

## The Universe of Code

Building EIS, the sensation I ultimately arrived at was surprisingly simple.

**A codebase has a universe-like structure.**

There is gravity.

Other code starts gathering around certain code. Abstractions pass through it. Design is built around it as a center.

And then a **structural center** is born.

In EIS terms, this is an **Architect**. High Design axis, high Survival axis. Meaning: code written by this engineer becomes the center around which other code is built — and **that structure survives**.

This is the gravity of a codebase.

---

## A Universe with Only One Gravity

In many teams, there is only one gravity.

A single Architect builds the central structure of the codebase. Decides the API design philosophy. Defines the granularity of abstractions. Solidifies the directory structure.

This is a strong structure. It has consistency. No ambiguity.

But it is simultaneously a **very fragile structure**.

In EIS metrics, such teams have distinctive signatures:

- **Risk: Bus Factor** — one departure collapses the team structure
- **Structure: Maintenance Team** rather than Architectural Engine — with only one designer, everyone else just maintains the existing structure
- **Anchor Density: Low** — few members stabilize quality

If that gravity disappears — the Architect transfers, quits, moves to another project — the codebase disperses at once. Design consistency crumbles. Structure weakens.

Engineer F's departure in Chapter 5 was the moment this could have happened. But it didn't. machuz reached Architect Builder at the same timing — a **generational shift of gravity** occurred. This was partly fortunate, partly structurally inevitable. But what if machuz hadn't been there? The team would have become "a universe that lost its gravity."

---

## A Universe with Multiple Gravities

In truly strong teams, something else happens.

**New gravity is born.**

The existing Architect maintains the structure. Around them, **emergent Architects** begin creating new gravitational centers.

Designs collide. Abstraction granularity is debated. Implementation approaches clash.

At first glance, it might look like **conflict**.

But structurally, this is the **evolution of the code universe**.

In EIS team timelines, this evolution is traceable through concrete metrics:

```
Classification:
  Period         2024-H1            2024-H2             2025-H1
  Character      Elite              Guardian            Elite
  Structure      Architectural Team Maintenance Team    Architectural Engine
  Risk           Bus Factor         Design Vacuum       Healthy
```

**Architectural Team → Maintenance Team → Architectural Engine**

Initially, one Architect supported the structure (Architectural Team). As that person's influence faded, the structure entered maintenance mode (Maintenance Team). But as multiple members grew their Design axis, the team evolved into a structure where design capability was distributed (Architectural Engine).

This is "a universe with multiple gravities."

---

## Four Forces

In teams with multiple gravities, four forces operate simultaneously:

| Force | Carrier | Observable in EIS |
|---|---|---|
| **The force to protect** | Architect Builder | High Design + High Survival. Maintains existing structure, guards consistency |
| **The force to break** | Emergent Architect | High Design + High Production. Introduces new abstractions, updates existing structure |
| **The force to stabilize** | Anchor | High Quality + High Survival. Fixes bugs, writes tests, raises the quality floor |
| **The force to build** | Producer | High Production. Moves features forward, creates user value |

When all four are present, the team's Classification becomes **Elite / Architectural Engine / Builder / Mature / Healthy**.

Conversely, when any one is missing, it surfaces as Risk:

- Missing the force to protect → **Design Vacuum**
- Missing the force to stabilize → **Quality Drift**
- Missing the force to break → **Stagnation**
- Missing the force to build → **Declining**

---

## Structural Impact, Not Volume of Voice

Writing this far, I notice it has interesting implications for engineering evaluation.

In the software world, there are **engineers who are skilled but quiet**.

They understand good design. They write good code. But they don't talk much in meetings. Documentation is minimal. Presentations aren't their thing.

Conversely, there are **engineers whose voice is loud but whose code leaves no gravity**.

They articulate direction. They speak up in design reviews. But look at the actual codebase — their code isn't at the center of other code. Low Survival axis. Low Design axis.

Which one is **actually moving the codebase**?

That answer lives not in discussions but in code. Git history contains not just *who wrote code*, but **who created the gravity of the codebase**.

What EIS tries to do is see engineers by **structural impact, not volume of voice**.

---

## Proving Your Team's Strength — The Hiring Context

This "observability" has another practical application: **hiring**.

In engineering hiring, it's easy to say "our team has strong technical culture." But almost no team can **back that up with data**.

With EIS team timelines, you can:

- **Classification: Elite / Architectural Engine / Mature** — "Our team has distributed design capability. We don't depend on any single individual."
- **Health: Complementarity 0.85** — "Our members' skills are complementary with low overlap bias."
- **Risk: Healthy** — "No Bus Factor, Design Vacuum, or Quality Drift risks."
- **Phase: Sustained Mature** — "Stable without stagnating."

Instead of telling candidates "this is a technically interesting team," you **show them the graphs**. Show score trajectories. Show how the team has evolved over time, backed by data.

This works in reverse too. Candidates can use it to evaluate teams. "Is this team actually growing?" "Does it have a design culture?" — the EIS dashboard can answer that.

Hiring is matching. Being able to show your team's real capability honestly and quantitatively is valuable for both sides.

---

## What It Means to Make Something Observable

The history of physics is also a history of observation.

Planets orbited the sun before telescopes existed. But it was only when this became **observable** that we could understand the structure, make predictions, and put it to use.

Codebase structure is the same.

Who is the Architect? Where is the gravity? Is the team evolving or declining?

These things **already exist** in Git history. They just weren't observable.

EIS is an attempt to make the universe of code **a little more observable**.

---

## Great engineers don't just write code. They bend the gravity of codebases.

A great engineer is not simply someone who writes code.

They might be **someone who bends the gravity of codebases**.

---

### Series

- [Chapter 1: Measuring Real Engineering Impact from Git History Alone](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)
- [Chapter 2: Team Analysis — How Individual Scores Reveal Team Health](https://dev.to/machuz/git-archaeology-2-team-analysis-how-individual-scores-reveal-team-health-4e6a)
- [Chapter 3: Archetypes — Behind the Numbers, There Are People](https://dev.to/machuz/git-archaeology-3-archetypes-behind-the-numbers-there-are-people-48kl)
- [Chapter 4: The Normalization Trap — "Top in a Repo" Means Nothing](https://dev.to/machuz/git-archaeology-4-the-normalization-trap-top-of-a-repo-means-nothing-11c5)
- [Chapter 5: Timeline — Scores Don't Lie, and They Capture Hesitation Too](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5)
- [Chapter 6: Teams Evolve — The Laws of Organization Revealed by Timelines](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4jf3)
- **Chapter 7: Observing the Universe of Code** (this post)

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score)

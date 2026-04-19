---
title: "Structure-Driven Engineering Organization Theory #0 — Why Existing Org Theory Doesn't Work"
series: "Structure-Driven Engineering Organization Theory"
published: true
description: "Stop evaluating people. Start observing structure. A new engineering org theory built on Git Archaeology, layered organizations, and the isomorphism between product and team."
tags: management, leadership, engineering, culture
---

*Change hiring. Change the evaluation system. Run more 1-on-1s. The organization keeps breaking in the same place — because what you're observing is still "people."*

*Observation, not evaluation. Structure, not people. Design, not emotion.*

---

> **Scope of this chapter**: thinking layer (shifting how we look at organizations from evaluation to observation).

## What This Book Is About

Engineering org theory has, for a long time, taken **people** as its subject.

Hiring, evaluations, 1-on-1s, feedback, career paths, culture fit — the objects being handled are all people. And the lever pulled to change the organization has always been the person.

This book steps off that subject, once.

What we handle is not "people" but **the structures people produce**, **the layers those structures run on**, and **the transformations that cut through those layers**. People stand on top of these — but the object of observation and design is structure, not the person.

The title — *Structure-Driven Engineering Organization Theory* — is a direct statement of that. The core of **Structure-Driven** is to describe the organization as "**Principle ── transformation ── Structure ── transformation ── Implementation**," and to design with awareness of **3 layers and the 2 transformations that run through them**. Structure is what appears as the *result* of transformation between principle and implementation — the state where thought has taken form as architecture, org chart, or information design. Designing an organization in a structure-driven way means building one where these transformations keep running. Chapter 4 develops this in full.

### What "structure" refers to in this book

This book uses the word "structure" frequently. To avoid conflation, four distinct objects are named upfront:

1. **The structure of code** — chapter 2. Read along three axes: Design / Survival / Change Pressure.
2. **The structural model of people** — chapter 3. The Role × Style × State three-axis topology.
3. **The layer structure of organizations and their transformations** — chapter 4. Implementation / Structure / Principle — 3 layers + 2 transformations.
4. **The structure of products** — chapter 5. Carries the same 3 layers + 2 transformations as the organization, **isomorphically** (screens & APIs / information architecture / worldview).

The four are different objects, but the book's central claim is that **all four can be handled with the same "observation → language → design" practice.** Items 3 and 4 in particular (organization and product) are **isomorphic structures** — they break and heal symmetrically; chapter 5 develops this. When "structure" appears in the text, infer from context which of the four it points to.

## Where Existing Org Theory Breaks

Let's locate the exact point where conventional org theory stops functioning. The theory itself isn't wrong; **its inputs aren't trustworthy**.

### 1. Evaluation Handles Impressions, Not Observations

Engineering evaluations usually look like this:

- Self-assessment
- The manager's observation notes
- Memories from 1-on-1s
- 360-degree review, averaged subjectively

All valuable signals. But all of them are **data that has already passed through a human perceptual filter**.

- People with loud voices are over-valued
- People silently holding the complex parts together are invisible
- People who move dramatically in code review stick in the mind
- People who have quietly held down the system's roots don't come up in conversation

An evaluation that doesn't distinguish "observed" from "impressed" is treating noise as signal. However elegant your career ladder, **noisy input produces noisy decisions**.

### 2. Management Can't Observe "Effort"

"They're working hard but not producing results."
"They're producing results, but they're burning out."
"Just having them around makes the team tighter."

Field managers handle this kind of intuition every day. But there's no way to translate it into **a language that transfers to other people**.

"Effort" is an internal state, not a structure. States live inside a person and can't be observed from outside. A management practice built on something unobservable eventually collapses into **total dependence on the manager's interpretive ability**. Swap the manager and the same engineer's evaluation changes. As the organization grows, this variance expands exponentially.

### 3. Process Is Placed Where Structure Should Be Observed

Scrum, OKRs, career frameworks, engineering ladders — none of these are bad. They're excellent **substitute devices** for making decisions in the absence of observation.

The failure mode is mistaking the substitute for the thing itself.

- If velocity goes up, the team is getting better (believed without checking)
- If OKRs are hit, the organization is moving forward (believed without checking)
- If ladder checkboxes are filled, they're senior (believed without checking)

Substitute devices are **scaffolding placed because the structure underneath can't be seen**. The moment the scaffolding becomes the goal, the organization optimizes for "filling scaffolding." The people in the field know this. But without an alternative, they keep filling.

---

What all three failure modes share is this: **there is no means of observation, or the organization is pretending to observe when it isn't**.

So the starting point for org theory isn't "improving evaluation accuracy" or "polishing process." It's **acquiring observation itself**.

## The "Effort Can't Be Observed" Problem

A concrete example.

A team has three engineers: A, B, and C. All self-identify as senior. Six months in, evaluation time rolls around.

- **A** commits frequently every week. Many PRs, visible in review. Speaks up in meetings.
- **B** commits a few times a month at most. But files B touches rarely get rewritten.
- **C** doesn't build much new functionality. Mostly fixes others' bugs and cleans up debt.

Under conventional evaluation, A stands out overwhelmingly, B looks "low-contribution," and C is dismissed as "unflashy."

Now observe the three **structurally**, and a different picture emerges:

- A produces volume, but most of what A writes is replaced by others within 3 months (short-lived code)
- B commits rarely, but a high fraction of B's code survives beyond 6 months (long-lived code — effectively an *Anchor*)
- C doesn't ship new features, but system-wide bug density drops after C has touched the code (a *Cleaner*)

This isn't hypothetical. Feed only `git log` and `git blame` in, compute code survival, first-pass quality, and debt cleanup rate, and this kind of structure becomes **mechanically visible**.

This is the critical point:

**"Effort" can't be observed. But "contribution to structure" can.**

Conventional org theory has been trying to handle unobservable effort. Structure-Driven Engineering Organization Theory starts from observable structural contribution. They look like similar problems, but the inputs are different. Different inputs yield different decisions.

## What Happens in an Organization That Fails to Observe

When the A/B/C gap **goes unobserved for years**, the organization develops a concrete kind of pain. This isn't abstract — it's a chain reaction happening in real workplaces.

It proceeds in this order:

1. **Observation fails**, so people with a track record of structurally-sound work (the Bs and Cs) don't get surfaced.
2. Instead, **loud voices, meeting-room presence, and short-term visibility** (the As) get the promotions. Without observational data, decisions fall back on impression.
3. Once promoted, their judgment — through no malice — tends to **drift away from structure**. Their own history isn't built on reading structure. So architectural decisions, hiring bars, refactoring priorities — all get pulled toward volume and charisma.
4. **For engineers who actually read structure, this drift becomes daily pain.** "This design will break in a year," they say, and get back: "Your argument is too abstract." The evidence always sides with the loud voice.
5. Eventually the Bs and Cs **leave quietly**. Anchors and Cleaners don't make scenes. Because they don't, the reason for their departure never reaches leadership. When it does, it gets filed as "wasn't a fit."
6. What remains is the loud voices, and **the people who adapted to the loud voices**. The organization has optimized itself in a direction that doesn't serve structure.

Once this loop starts spinning, it accelerates. The loud voices become the next evaluators. Loud-voice criteria select the next promotees. Structure-readers drift further away. An organization that was only **missing observation** becomes **a machine for expelling the structurally-competent**.

This is the strongest practical reason to introduce observation.

Observation isn't about "fair evaluation" or "management efficiency." **It exists so organizations don't lose the people who are structurally right.** People with the history to make sound structural decisions need to be findable, promotable, and empowered — independently of the volume of their voice.

When observation is in place, at least this question becomes askable:

> "Over the last three years, which code did this promotion candidate leave behind, and on which layer have they been fighting?"

Being able to ask that question, in itself, protects the organization.

## Subjective Evaluation vs. Structural Observation

A clarification to avoid misreading.

**Structural observation doesn't reject subjective evaluation.** Human judgment is irreplaceable as a signal source. Especially for the emotional and ethical dimensions that don't show up in structure — "how they support the team," "how they treat juniors" — only human eyes can see these.

But laying subjective evaluation **on top of** structural observation is completely different from making decisions on subjective evaluation alone.

- Former: look at the structure first, then add the human value not captured by numbers
- Latter: judge by impression without knowing the structure

This book is about the shift from latter to former. Structural observation comes first. Then, what only human eyes can see is seen by human eyes.

### Three Layers of Observation

When observing an organization, distinguish three layers.

| Layer | What's observed | How to observe | Examples |
|---|---|---|---|
| **Behavior** | Who did what | Logs, calendar, Slack | Commit frequency, meeting hours |
| **Output** | What was produced | Artifacts, docs | Features shipped, bugs filed |
| **Accumulation** | What remained, how it's connected | Git Archaeology, dependency graphs | Surviving code, owned modules |

Most organizations stop at Behavior and Output. They **don't observe the Accumulation layer**, so "effort" and "contribution to the structure that remains" can't be told apart.

The core thesis of Structure-Driven Engineering Organization Theory:

> **Every organizational intervention should be decomposed across the three layers — Behavior, Output, Accumulation — before it's made.**

In 1-on-1s, evaluations, hiring — don't blend the three. Blended, the conversation doesn't align. Split, the same topic surfaces as three distinct decisions.

## The Stance of This Book

Five foundational premises.

1. **Observation, not evaluation.** Don't make decisions about what you can't observe.
2. **Structure, not people.** Don't try to change people; read the structures people produced, and intervene on the structure.
3. **Organizations run on layers.** Abstraction and implementation live on different layers. Blend the layers and no amount of talent will prevent the organization from spinning its wheels.
4. **Product and organization are isomorphic.** Product UX badness is congruent with organizational layer badness. Fixing one can reveal how to fix the other.
5. **Structure, not emotion, drives decisions.** Not because emotion is dismissed — because emotion and structure must not be blended. The emotionally-right call and the structurally-right call can be handled separately.

These five may sound cold to some readers. The intent isn't coldness. The practical goal is exactly the opposite: **observation lets us treat people more fairly**.

People who slipped through the subjective-evaluation net and were genuinely contributing finally become visible. People who were over-valued subjectively become visible for what they *aren't* securing. Both are healthy information for the organization.

## The Structure of This Book

The path forward:

| Ch | Title | Covered |
|---|---|---|
| 1 | The Concept of Observation | Splitting evaluation from observation; introducing EIS |
| 2 | What Remains Is Structure | Treating code as structure, not artifact |
| 3 | People, Read as Types | Anchor / Producer / Mass and related archetypes |
| 4 | Organizational Layers | Implementation / Structure / Principle — 3 layers + 2 transformations, transformers and transformation coaches |
| 5 | Product-Organization Isomorphism | Product's three layers, symptoms of the isomorphism, symmetric intervention design |
| 6 | Designing Interventions (1-on-1 / pair programming) | Decomposing across Behavior / Output / Accumulation |
| 7 | Making Culture | Language makes culture. Evaluating transformation itself. |
| 8 | Conditions for a Structure-Driven Engineering Organization | Reproducibility, observability, self-correction |
| 9 | Connecting to OrbitLens | The moment observation becomes SaaS |
| 10 | Conclusion — Building an OS, Not an Organization | Design over management |

Chapters are written to be readable independently, but 1 → 3 → 4 → 6 forms a single logical spine. Read that sequence in order and the connection surfaces.

## What Changes in the Field

Each chapter ends with a "what changes in the field" section — to prevent the book from drifting into pure abstraction.

At this intro level, here's what changes overall:

- **1-on-1 content shifts.** "How's everything going?" becomes "Over the last three months, here's the fraction of your code still alive, and here's where it's declining — this is the layer boundary where it's being lost."
- **Evaluation's subject shifts.** "Are they working hard?" becomes "What structure have they contributed to?"
- **Hiring criteria shift.** You name whether you're short on Anchors or short on Producers before opening the req.
- **The manager's job shifts.** From "watching people" to "designing structure and placing people on the right layer."
- **How your own work looks to you shifts.** Without going through subjectivity, you can see what you've left behind and which layer you've been fighting on.

For engineering managers and tech leads, these are changes to **tomorrow's work**. The payoff from observability isn't a distant future — it kicks in immediately.

## Before You Continue

This book isn't written as a continuation of existing engineering org theory. You *could* translate its ideas into existing vocabulary (engagement, psychological safety, velocity, ladders), but doing so **gets pulled back into the gravity of that vocabulary** and ends up in the same frame as before.

So new vocabulary is introduced deliberately:

- **EIS (Engineering Impact Signal)** — an observation index reading structure from Git history
- **Git Archaeology** — the method of excavating the strata of code to read structure
- **Anchor / Producer / Mass** — structural archetypes of people
- **Implementation / Structure / Principle** — organizational layer names (the 3-layer model)
- **Transformation** (Principle ↔ Structure / Structure ↔ Implementation) — the movement that cuts through the layers
- **Transformer / transformation coach** — the person who carries out transformation (former), and the person who grows others' transformation capability (latter)
- **Isomorphism** — the design principle that the layer structure of an organization and the layer structure of its product break and heal symmetrically

**All of these will be defined in subsequent chapters. It's fine not to understand them yet.**

### On Git Archaeology

"Git Archaeology" is the design philosophy and implementation behind EIS, already published as a separate book.

- [**git-archaeology — Observing Gravity, Civilization, and the Future of AI from the Strata of Code**](https://zenn.dev/machuz/books/git-archaeology) (Zenn Book, Japanese) — English dev.to series also available.

Reading it first deepens the EIS context later, but **it's not required**. All terms this book needs will be redefined in Chapter 1 and Chapter 2. Consult it only when you're curious.

### Relationship to Existing Org Theory

This book doesn't reject existing org theory. Instead, it tries to **place much of that theory back on top of observation**.

- **Conway's Law (organizational structure mirrors architecture)** — this book provides a means to **see Conway's Law-described correlations weekly through Git history × EIS**. Its contribution sits on the side of making the theory observable.
- **Team Topologies (designing organizations through flow efficiency and cognitive load)** — the layer structure here (Implementation / Structure / Principle) is a different cut from Team Topologies' team types and is meant to be used together. Layer is the **abstraction-level** axis; Team Topologies is the **responsibility-type** axis.
- **Staff Engineer literature (Will Larson and others)** — this book's Role × Layer takes Staff Engineer's archetype discussion and generalizes it. Tech Lead / Architect / Solver / Right Hand variants are adopted as-is.
- **Wardley Mapping (strategy via the evolution stage of components)** — pairs naturally with this book's AI-era reshaping prediction. This book's scope is narrower: it stays on people and organizational placement.
- **Systems theory and process philosophy lineage** — this book's core intuition, "**the middle is not a layer but a movement**" (what lives between Principle ↔ Structure ↔ Implementation is not a place but transformation), shares heritage with Luhmann's social systems theory (society as a chain of communications), Deleuze's philosophy of difference (relation and movement over substance), and general systems theory / cybernetics (feedback loops, not states, are the real object). The originality here is **operationalizing** that intuition into concrete engineering-organization design principles and observation methods — not claiming it as metaphysical invention.

The difference from these is that this book **assumes observation as a routine operation**. Many org theories are correct as theories but never required **data you could check weekly** on the floor. This book fills that gap with EIS. Put another way: it tries to **turn org theory into a science you can measure every week.**

The point of introducing new vocabulary is singular: **to give names to things that couldn't be observed before**. The moment a name exists, conversation becomes possible. The moment conversation is possible, design becomes possible. The moment design is possible, it becomes culture.

> It's not: language first, observation next, structure last.
> It's: **observation first, language next, structure last, culture as the consequence.**

In this order, we redesign the organization.

---

In the next chapter, we decompose the concept of "observation" itself. Evaluation, measurement, monitoring, observation — how do we separate the overlapping words? Why is EIS *observation* and not *evaluation*? Miss this, and the rest of the book collapses back into conventional evaluation theory.

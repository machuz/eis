---
title: "Structure-Driven Engineering Organization Theory #6 — Designing Interventions (1-on-1 / Pair Programming)"
series: "Structure-Driven Engineering Organization Theory"
published: true
description: "Slot weekly 1-on-1s into the calendar, ask 'how are things going?', and the same person keeps getting stuck in the same place. The intervention is trapped at the behavior layer. Real interventions hit structure, not emotion."
tags: management, leadership, engineering, productivity
---

*Six months of weekly "how are things going?" 1-on-1s won't move the organization one millimeter — the intervention is trapped at the behavior layer. The moment the conversation opens that way, it doesn't touch accumulation by a single gram.*

*Cut with structure, not with emotion. That's the only intervention that moves organizations.*

---

> **Scope of this chapter**: design layer (decomposing interventions into three targets — behavior, output, structure) + practice layer (redesigning 1-on-1s, pair programming, code review, and reorganizations).

### How this starts on the floor

Wednesday afternoon 1-on-1. The EM asks: "How are things?" The engineer thinks for a second: "Well, a few things going on... I'm working on it." The EM follows up: "Anything you're stuck on?" "Yeah, my last PR is taking forever to get reviewed. Motivation-wise it's been rough." "That's tough — don't worry about it, take it easy this week."

The conversation is warm. The EM cares. They'll do this again next week. And **six months later, the same engineer will be reporting the same blockage.**

What's missing in that exchange? **The intervention is confined to the behavior layer.** The concern, the reassurance, the next meeting — all of it pushes on "what you're doing this week." The *quality of what this person produces* and *what has accumulated from their work* — the **output** and **accumulation** layers — never came up once.

---

## The Three Layers of Intervention

![Three layers of intervention — accumulation shapes behavior and output](https://cdn.jsdelivr.net/gh/machuz/eis@main/docs/images/blog/sdo-ch6-three-intervention-layers-en.svg)

Interventions aimed at an organization split into three layers, based on what they target — **behavior** (what someone is doing), **output** (what got produced), and **accumulation** (what remains, and how it connects into the conditions of the next behavior).

This is the core of structure-driven thinking — **these three layers are not equal**. **Accumulation shapes the conditions under which behavior and output form.** What has remained, and how it's connected, regulates what people on the ground actually do, and determines the quality of what comes out. So if you only intervene at the behavior layer, and accumulation doesn't shift, behavior reverts on its own.

Read in the other direction: **change accumulation, and behavior and output change on their own.** That's the causal direction this book calls "structure-driven." Traditional interventions fail because they run the causality the wrong way — change behavior, hope output improves, hope accumulation follows. The order is inverted.

### Behavior layer

**What someone is doing day to day.**

- Meetings, commits, review time, Slack comments, deploy frequency, on-call hours
- Easy to observe, easy to change, effects measurable quickly
- But — **changing behavior alone doesn't leave anything at the output or accumulation layers.**

"Let's cut meetings." "Let's be more casual on Slack." "Let's do more 1-on-1s." These are behavior-layer interventions. They have their place, but they rarely move organizations on their own.

### Output layer

**What came out of the person.**

- Features, docs, decisions, RFCs, design proposals, mentoring trails
- One step heavier than behavior
- But — **output that nobody uses doesn't become structure.**

An RFC that nobody cites. A feature nobody uses after release. Minutes that don't get referenced. **Output that doesn't hook into later decisions or later implementation doesn't accumulate into structure.**

### Accumulation layer

**What remained. How things are connected.**

- The seven EIS axes, layer placement, transformation paths, shared conventions, team shape, hiring bar
- The heaviest layer. The hardest to change.
- But — **when this changes, behavior and output change on their own.**

The accumulation layer is the state that appears **as the accumulated result of individual behaviors and outputs** — and at the same time, **the ground that decides the conditions for the next generation of behavior and output.** It's hard to touch directly, but the moment it shifts, every behavior and every output hanging off it changes meaning.

To design interventions in a structure-driven way, start from "**what do we want to accumulate?**" — not from "what behaviors should we change?"

### The principle: don't mix layers

**Every intervention must declare which layer it targets.** Mixed interventions fail.

- "Let's improve communication" → behavior / output / structure all mashed together. No one can tell where the effect is supposed to land
- "**Increase 1-on-1 frequency**" → behavior layer
- "**Create a rule that every decision gets written down**" → output layer
- "**Reconfigure layer placement**" → accumulation layer

Vague interventions end with "it kind of felt like it helped." Layer-tagged interventions can be **verified in the next observation cycle**.

---

## Redesigning the 1-on-1

The traditional 1-on-1 opens emotion-first: "How are things?" "Anything you're stuck on?" The moment you enter through that door, **the intervention is pinned to the behavior layer**. Sharing a feeling is a behavior. Conversations don't naturally descend from there into structure.

Redesign: run the order **structure → output → behavior + emotion.**

```
Opening 15 minutes — accumulation layer share
  - Last 3 months of EIS signals
    - Production / Survival / Design / Cleanup / Breadth / Quality / Indispensability
  - Which layer they sit on (Implementation / Structure / Principle)
    and which transformations they're carrying
  - Their role in the team (Anchor / Producer / Cleaner / Specialist …)

Middle 15 minutes — output layer
  - Recent major outputs and the process that led to that structure
  - RFCs written, design calls made, mentoring done, review patterns

Closing 15 minutes — behavior layer + emotion
  - Next steps, blockers, fatigue, motivation
  - Spoken on top of the structure mapped out in the first 30 minutes
```

### Why this order

Putting emotion **last** is not because emotion is unimportant. It's the opposite — **separating emotion from structure lets you handle emotion itself with more care.**

If you open with emotion, the mood of the moment colors the entire 1-on-1. "They seem down today; we'll do the structural stuff next time." Which means **structural stuff is perpetually postponed.**

When you open with structure, emotion lands *on top of* structure. "The code stands. The Role has shifted toward Anchor. And lately they've been tired and sleeping badly." The **same** statement of fatigue reads completely differently in this frame. Emotion is no longer "a separate problem" — it's "a phenomenon inside the structure," and can be handled that way.

> **From the field**
>
> This pattern repeats — a pair is doing weekly 30-minute 1-on-1s. They've been doing it for a year. And yet, **neither the EM nor the engineer has reached consensus on the engineer's career direction or their place in the organization.**
>
> The reason is simple. Every 1-on-1, the topic drifts back to "this week's problem" or "how I'm feeling." **Not once in a year has structure come up.** A year of time is consumed.
>
> Emotion is valuable. But handling emotion for a year doesn't move a career. The only thing that eventually deepens the handling of emotion is securing time for structure first.

### Observation is for understanding, not evaluation

Putting structure at the front of a 1-on-1 invariably raises the question: *are we building a panopticon?* Quite the opposite. Opening with structure first is meant to **understand which direction the person has been investing time and energy in** — before anything else.

- Production is low for a month — maybe they weren't slacking; maybe they were spending that time on Design or Survival trial-and-error.
- Survival is low in a stretch — maybe the quality didn't collapse; maybe the situation genuinely demanded speed.
- Breadth is wide while Design is thin — maybe they aren't scattering; maybe they're deliberately exploring multiple domains in an early phase.

**Don't judge from the signals alone.** The signals are a map for reading **the direction of someone's effort**. Only once that direction is understood can an intervention land in a way that fits their effort. Interventions made without that understanding collapse back into top-down evaluation — **exactly the posture this book is trying to escape.**

Opening a 1-on-1 with structure isn't there to check "what are you doing?" It's a way of saying **"I see what you've been facing into — I've read it."**

### Make it your own — a question

> In the 1-on-1s you ran last week —
>
> - In how many of them did you spend **15+ minutes on the accumulation layer**?
> - Could you concretely cite the **outputs** from the other person's last three months?
> - Did any 1-on-1 that opened with "how are things?" end **still trapped at the behavior layer**?

### Try this in tomorrow's 1-on-1

> Spend the **first 15 minutes on accumulation only.**
>
> - If EIS is set up, share the person's last 3 months of signals (Production / Survival / Design / Cleanup / Breadth / Quality / Indispensability) as-is
> - If EIS isn't there, ask only: "What has this person **left** over three months? Where is that output **cited** or **referenced**?"
> - Emotions, blockers, next actions — push them to the back 30 minutes
>
> One try is enough to feel whether structure can ride into the conversation at all. That alone changes what a 1-on-1 means.

---

## Redesigning Pair Programming

Traditional pairing: pair by skill gap. Senior teaches junior. Expert shadows beginner. That's a **behavior-layer intervention** — sharing screen time, watching the expert's hands, answering questions.

Redesign: **pick the pair based on the layer-movement you're aiming for**, and say it out loud.

- "**Growing an Anchor**" → Anchor-shaped senior × promising junior. Transfer the Structure ↔ Implementation transformation.
- "**Succession for the Cleaner**" → existing Cleaner × someone who can face down debt. Pass on the cleanup form.
- "**Producer-speed maintenance**" → two Producers pair up. They pull each other's coding pace along, and speed is preserved.
- "**Scaling transformation capability**" → Anchor carrying Structure ↔ Implementation × a strong Implementation-layer Producer. Install transformation capability into the Producer.

The pair isn't chosen by **individual skill**, it's chosen by **the layer movement we want to cause.** When the intent is stated, pair programming becomes **a structure-layer intervention** — the person's type changes, the role changes, the team's placement changes.

---

## Redesigning Code Review

In many organizations, code review has become a behavior-layer intervention — "LGTM," "nit: typo," "this is personal preference, but…". Response time and comment count get measured, and that passes for "a team with active review."

Real code review is a **structure-layer intervention.**

### Redesign the review vocabulary

Instead of "LGTM," state **evaluation along the structural axes**:

- **Design**: does this change contribute to the codebase's Design layer (its architectural center), or is it a utility on the surface?
- **Survival**: will this land as robust code, or is it a shortcut likely to be rewritten in three months?
- **Cleanup**: is this sweeping someone else's debt, or is it just rewriting to your taste?
- **Quality**: how likely is this change to produce a fix afterward?

With this vocabulary, **the review itself becomes a structural observation record.** Run it alongside the EIS signals, and you can say things like "this Anchor's reviews actually lift the Design layer."

### Visually separate "nits" from "structural concerns"

If every comment lands as an equal item, **nits drown structural concerns.** Most review fatigue comes from this.

- 🟢 **Nit**: style, naming, minor polish. Does not block.
- 🟡 **Non-blocking suggestion**: worth discussing, but this PR can ship without it.
- 🔴 **Blocking**: structural concern. Affects Design / Survival / Quality.

Just separating them visually is enough to lift review from behavior layer to accumulation layer.

---

## Redesigning Reorganization

"We're reorganizing" usually ends as a refactor of the tree diagram — boxes move, reporting lines re-draw, new titles get handed out. This is an **output-layer intervention.** The new org chart is the output; it can be evaluated at that moment. And **none of the structure has changed.**

Redesign: **define the target of the reorg as a change in the accumulation layer.**

- Before rewriting the org chart, **draw the layer map** — who sits on which layer, who carries which transformation
- State the goal in **layer-thinness / missing-transformation** language: "the Structure layer is thin," "the Principle ↔ Structure transformation has stalled"
- Don't move boxes — **place a transformer / preserve a code connection / separate a role**. Write down what is *expected to change*, in structural vocabulary

Most reorgs skip this mapping. The boxes moved, but the transformers are stuck in the same place — **the chart changed, the structure didn't.** This is the parent of chapter 4's "implementation is fast but direction wobbles" / "we hired more seniors but nothing's moving" symptoms.

---

## Reintervening on "Talented But Spinning"

Chapter 3 and chapter 4 referenced the "looks capable but spins" pattern. Decompose it across the three intervention layers and the intervention changes shape.

Common observation: **behavior is busy, output is thin, structure shows Breadth ↑ / Survival ↓ / Design ↓** (broad, shallow, scattered).

- **Behavior layer** — meeting attendance, participation, review turnaround all at or above average
- **Output layer** — code volume exists but nothing that's landed as a feature
- **Accumulation layer** — on EIS, only Breadth is high; Survival and Design are low (wide, shallow scatter)

Telling this person "work harder" or "focus more" without seeing the accumulation is a **behavior-layer intervention.** They're already busy. More activity doesn't tighten the scatter.

The correct intervention is in the **accumulation layer**:

- **Move the placement**: Breadth is excessive — reposition them on a specific layer or domain
- **Change the type**: a scattered Producer may fit better as a Specialist settled into one area
- **Install transformation capability through pairing**: pair them with someone who can carry the Structure ↔ Implementation transformation

**Don't change the person — change the placement.** The chapter-4 principle becomes a concrete intervention here.

---

## What Changes in the Field

Adding this three-layer intervention vocabulary to the organization shifts the following:

1. **1-on-1 fatigue drops.** Emotion and structure aren't mixed, so both get handled with care.
2. **Interventions become measurable.** The next observation cycle (next EIS run) shows whether the structural signals moved.
3. **The language shifts from "people problem" to "placement problem."** Not "A is spinning" but "A is a Spread type with Breadth ↑ / Survival ↓ — placement isn't right for the structure layer."
4. **"Draw the layer map before you rewrite the org chart" becomes the default.** Before moving boxes, people agree on who should carry which transformation.
5. **Code review becomes a structural observation record.** LGTM counts get replaced by contribution to the Design layer.

Cutting with structure instead of emotion — that's where this chapter lands.

> **The reason an organization doesn't change, even after years of 1-on-1s, is not that the people are bad or lack talent. It's that the intervention has stayed pinned to the behavior layer — accumulation was never touched.**

---

## What's Next

We've now assembled the vocabulary to observe an organization, describe its structure, and design interventions into it. But if every intervention still depends on **an individual's interpretive skill**, the organization reverts the moment that individual leaves.

The moment an intervention becomes the **language** of the organization — not the skill of one person — **it becomes culture**. Once culture, the structure holds even when the intervener changes.

Next chapter: **making culture**. How the book's vocabulary (EIS, the three layers, transformation, Role × Style × State) gets installed into the organization's daily conversation. Culture isn't the sharing of values — it's **the sharing of language.**

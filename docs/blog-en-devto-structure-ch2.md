---
title: "Structure-Driven Organization Theory #2 — Output as Structure"
published: true
description: "Measure code not at the moment it's written, but by what happened after. What remains is structure."
tags: engineering, leadership, git, career
---

*Measure code not at the moment it's written, but by what happened after. What remains is structure.*

---

## Is Code an Artifact, or a Structure?

Ask most organizations "what's this engineer's output?" and you get:

- Features shipped
- PRs closed
- Lines of code written
- Tickets resolved

All of them describe **volume existing at the moment of writing**. On the time axis, only the transition from zero to positive is observed.

But code keeps moving after it's written. It gets rewritten, inherited, deleted, survives, breaks. Volume at the moment of writing shows **nothing about the life of the code afterward**.

This book switches the viewpoint here. **Code is not an artifact; it's a structure.** Reading it as structure means following what remains, what disappears, and what gets rewritten along the time axis.

## The Limits of Artifact View

Treating code as artifact creates concrete distortions.

### 1. Short-Lived and Long-Lived Code Are Indistinguishable

A wrote 1,000 lines; B wrote 100 lines. Under traditional metrics, A contributed 10× more.

But six months later, A's code is almost entirely rewritten, and B's is all still there. The rewrites weren't wasted — they served as scaffolding for the next developer to explore on. Still, in terms of **remaining code**, B's 100 lines carry **greater structural contribution**.

Artifact view can't make this distinction. Result: in workplaces where only the moment of writing is noticed, mass-producing short-lived code becomes rational strategy.

### 2. Refactors and Cleanups Don't Get Credited

Cleaning up someone else's past debt doesn't ship new functionality. Line count doesn't go up — it may decrease. It looks unflashy on a ticket board.

But **without people who do this, the organization rots**. A codebase, unless deliberately cleaned, only grows in entropy over time.

Artifact view misses this kind of contribution. The cleaners end up as "unflashy but necessary people," receiving vague gratitude from the beneficiaries. With structural view, you can observe **how much debt was reduced** by a cleanup.

### 3. The Gradient of Rewrites Isn't Visible

Every "rewrite" has a different meaning:

- Good design, and the next generation rewrites while inheriting it
- Bad design, noticed later, and the next generation rewrites it
- Writer and inheritor intentions misalign, and it gets rewritten repeatedly

These three don't collapse into a single number. **How rewriting is happening** needs to be seen structurally.

Artifact view captures only the fact that a rewrite happened. Structural view extracts **information usable for the next decision**.

## Three Axes for Reading Code as Structure

This book uses three axes to read code as structure.

### Design

**Which layer's files did the person contribute to, and how much?**

An organization's codebase has a center of gravity. Entry points, auth foundations, domain models, data schemas, shared libraries — these are depended on by many others. Changes here **move the design's center of gravity**.

The Design axis weights contributions by the centrality of the files touched. Editing 1,000 lines of peripheral utilities doesn't move the center. Editing 10 lines of a core file does.

EIS takes an architecture-file list from config and weights contributions accordingly.

### Robust

**Time-decayed survival volume.**

Code survival isn't linear. 100% right after writing, 95% a month later, 60% a year later — it doesn't decay that smoothly. Sometimes there are short bursts of heavy rewriting, sometimes long plateaus of stability.

Robust is defined as **survival volume multiplied by a time-decay function**. In EIS, the form is `exp(-days / τ)`, decaying with days elapsed since the write. Code that survives longer weighs more.

This isn't "old code worship." It's not about being old — **the fact of having survived time** is itself information.

### Survival

**The fraction still attributed to the writer right now, after others' rewrites.**

Survival is the simplest. `git blame` gives current attribution; aggregate by original author. Of 100 lines written six months ago, 30 still blame to me — Survival is 30%.

Survival carries the most information when combined with Design contribution. **Survival on peripheral files** and **survival on central files** mean different things.

### How the Three Axes Relate

The three aren't independent. They fill each other in.

- High Design, low Survival → Touched the center but couldn't anchor it
- Low Design, high Survival → Wrote long-lived code in peripheral areas
- High Design and high Survival → Carried the organization's skeleton — an Anchor

Evaluating on one axis alone always distorts. **Seeing the three simultaneously is what makes the structure speakable**.

## Why Code Gets Rewritten

Read as structure, "rewriting" isn't the enemy; it's the information source. Look at the **gradient** of rewrites, not the rate.

Four main causes:

### 1. Design Error

The original design didn't survive later requirements. This happens to everyone. Perfect design doesn't exist. The question is whether you can **distinguish** "rewritten because of design error" from "survived because the design was right."

Robust Survival makes the distinction visible. A region rewritten heavily over a short period signals design-error rewriting. A region rewriting gradually over a long period is natural evolution.

### 2. Requirements Change

Business requirements shifted, and the code followed. This isn't a design error. Being able to follow is itself evidence of **design flexibility**.

Organizations where Design-layer changes **propagate appropriately with each requirement shift** are structurally healthy. Organizations where the Design layer sits isolated, not reflecting requirement changes, are structurally rotting.

### 3. Style Mismatch

Writer and inheritor had different preferences. This is closest to a meaningless rewrite. A small amount is fine, but **an organization where this is the main cause of rewrites has a political codebase**.

Style alignment opportunities should be taken in pairing and in continuous review. Taking them in rewrites is too late.

### 4. New Understanding

The writer realized later: "this would have been better written this way." Refactoring. Less a necessary evil, more **an inevitable part of growing the structure**.

This kind of rewrite shows up as Debt Cleanup signal — volume of cleaning up others' past debt with current understanding. People high on this axis keep the codebase healthy.

Don't collapse rewriting into one number. **Read the gradient** — these four become distinguishable.

## Git Archaeology: A Concrete Sketch

How do you actually compute Design / Robust / Survival from a Git repo? The procedure at a conceptual level:

```
1. git log --numstat --all           # All commit change volumes
   → (author, file, +/-) per commit

2. git blame <file> at HEAD          # Current attribution per file
   → For surviving lines, who wrote them

3. Apply exp(-days/τ) per line       # Time-decay factor
   → Robust Survival

4. Design weight: weight each commit
   by architecture-central file list
   → Design contribution

5. Cleanup detection:
   Detect commits rewriting the
   previous (other) commit
   → Debt Cleanup

6. Ownership map:
   Aggregate blame ratios per module
   → Indispensability (bus factor)
```

EIS computes all of the above mechanically. Seconds to minutes per repo. Even cross-organization, half a day is plenty.

**Not spending time on observation** is critical. A heavy observational device doesn't survive operational use. EIS is designed to be light enough to run **weekly or monthly** as a routine.

## A Concrete Case: The Same Person Wears a Different Face in a Different Codebase

Continue the A/B/C example from the previous chapter, but move it into a different codebase.

Say B (the Anchor type, high Survival) joins a young-founder-led startup. The codebase is only two years old; all code is still mid-rewrite. Measuring Survival produces mostly noise.

Here, B doesn't function as an "Anchor" in the original definition. But **no one has touched this organization's Design layer** — schema design, architectural direction, auth foundations are underbuilt.

When B goes to work on that, the Design axis climbs sharply within three months. Survival doesn't show up yet, but **a history of Design contribution** is now on the record. Another six months in, B's Design-layer code begins to get rewritten by others who inherit it. Survival starts climbing.

This is what "**reading structurally, the same person wears a different face depending on the organization**" means. You can't compare A's effort vs. B's effort in absolute terms. But **which axis a person contributes to in which phase of which organization** is visible through structural observation.

## What Changes in the Field

When code becomes readable as structure, these change:

1. **Code review vocabulary shifts.** "This code is clean" becomes "This code contributes to the Design layer / seems likely to be Robust."
2. **Flashy pre-launch contribution separates from contribution that survives six months later.** Short-term and long-term output become visible on different indicators.
3. **Refactoring becomes evaluable for the first time.** With a Debt Cleanup axis, "this quarter we reduced debt by 30% through cleanup" becomes sayable.
4. **Staffing a new project changes.** Whether you need an Anchor, a Cleaner, or a Producer becomes an axis-level decision.
5. **Post-departure evaluation of what someone left behind shifts.** "How much of the code someone who left wrote is still alive?" becomes an ex-post metric for their contribution to the organization.

The fifth is especially important. Conventional organizations are **built to forget the contributions of people who left**. With structural observation, what someone left in the Design layer can be honored in numbers, over time. This is also proper respect for the person.

## What's Next

We now have axes for reading code as structure. The next question: **who writes the code visible on those axes?**

View people as "ability" and you can't observe them from outside. View people as "types of contribution to structure" and you can. The same person contributes to structure in wildly different axis distributions.

Next chapter: three basic types — Anchor / Producer / Mass — and their derivatives. These aren't personality types. They're **patterns emerging from structural observation**. Read them with that in mind.

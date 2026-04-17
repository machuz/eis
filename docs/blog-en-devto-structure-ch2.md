---
title: "Structure-Driven Engineering Organization Theory #2 — Output as Structure"
published: true
description: "Measure code not at the moment it's written, but by what happened after. What remains is structure."
tags: engineering, leadership, git, career
---

*Measure code not at the moment it's written, but by what happened after. What remains is structure.*

---

> **Scope of this chapter**: design layer (defining the observational axes for reading code as structure rather than artifact).

### How this starts on the floor

X's first PR was textbook clean. Naming conventions perfect, tests in place, every reviewer LGTM'd it on sight. The same week, Y, joining the company in the same cohort, shipped a PR with shakier naming and abbreviated variable names; reviewers added "let's refactor this later." Y agreed and moved on.

Six months later. The clean code X wrote in that PR has been almost entirely refactored away. Y's code, including the parts that drew refactor comments, is almost entirely still there. "Clean X" and "rough-but-shipped Y" — by the measure of contribution-to-structure, the ranking has flipped.

This chapter assembles the vocabulary that separates **how clean it looked at the moment of writing** from **how it survives after the writing.**

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

This book uses three axes to read code as structure — **Design**, **Survival**, and **Change Pressure**. Survival itself breaks into two sub-measures.

### Design

**Which layer's files did the person contribute to, and how much?**

An organization's codebase has a center of gravity. Entry points, auth foundations, domain models, data schemas, shared libraries — these are depended on by many others. Changes here **move the design's center of gravity**.

The Design axis weights contributions by the centrality of the files touched. Editing 1,000 lines of peripheral utilities doesn't move the center. Editing 10 lines of a core file does.

EIS takes an architecture-file list from config and weights contributions accordingly.

### Survival

**How much of what the person wrote is still there.** Two sub-measures, with different meanings.

#### Raw Survival

**The fraction still attributed to the writer right now, after others' rewrites.**

Raw Survival is the simplest. `git blame` gives current attribution; aggregate by original author. Of 100 lines written six months ago, 30 still blame to me — Raw Survival is 30%.

#### Robust Survival

**Survival volume with time-decay weighting.**

Raw Survival naively counts "lines still there." By itself, a line written five years ago and a line written last week each count as one line. That means **time-in-tenure inflates the number** — you can game it by sticking around long enough.

Robust Survival fixes this by weighting every surviving line by **how long ago it was written**. In EIS, the weight takes the form `exp(-days / τ)` (`days` = elapsed days, `τ` = decay constant). **Older lines weigh less.**

Two design intentions:

1. **Game-resistance**: tenure alone can't pile up the score. Contribution from old lines fades over time.
2. **Weight on the present**: codebases evolve. Weighting **recently-written lines that are still alive** higher lets contribution to the *current* structure show through.

#### Why Split Raw and Robust

Raw Survival alone can't distinguish "**code that's been there a long time**" from "**code written recently that's still in use**." Putting Robust next to Raw reveals the **temporal breakdown** of survival for the first time.

But "been there a long time" doesn't automatically mean "contributed to structure." It may simply be **surviving because no one touches it**. Raw and Robust side by side can't fully answer that. What settles it is the next axis — change pressure.

### Change Pressure

**How much change is concentrated in a given area.**

The definition is simple:

```
change_pressure = commits_touching_module / module_LOC
```

Divide the commits that touched the module over a period by the module's size (LOC). A large module with many commits may still have ordinary pressure; a small module with the same commit count has high pressure.

Change pressure matters because it can **flip the interpretation of Survival**:

- High Survival in a **high-pressure** region → the code is **enduring under active change pressure**. Real skeleton.
- High Survival in a **low-pressure** region → the code is **not enduring; it's just untouched**. Possibly frozen debt.

Evaluate Survival alone, and the second kind of writer gets treated as skeleton-builder by mistake. Layering change pressure on top lets you tell **robust skeleton** from **frozen-and-avoided territory**.

### How the Three Axes Relate

The three aren't independent. They fill each other in.

- Design ↑ + Survival ↓ → Touched the center but couldn't anchor it
- Design ↓ + Survival ↑ → Wrote long-lived code in peripheral areas
- Design ↑ + Robust Survival ↑ + **Change Pressure ↑** → **a real Anchor** (skeleton enduring under pressure)
- Design ↑ + Raw Survival ↑ + **Change Pressure ↓** → **Fragile skeleton** (surviving only because nothing touches it)

Evaluating on one axis alone always distorts. **Seeing the three simultaneously is what makes the structure speakable**.

> Change pressure is the axis that turns EIS from a mere "impact metric" into an **engineering risk detector**. In this book it's one of three axes; in the risk-detection context it plays a lead role on its own.

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

How do you actually compute Design / Survival / Change Pressure from a Git repo? The procedure at a conceptual level:

```
1. git log --numstat --all           # All commit change volumes
   → (author, file, +/-) per commit

2. git blame <file> at HEAD          # Current attribution per file
   → Raw Survival
     (surviving lines, by original author)

3. Apply exp(-days/τ) per line       # Time-decay factor
   → Robust Survival
     (Raw Survival weighted by time decay)

4. Design weight: weight each commit
   by architecture-central file list
   → Design contribution

5. Change Pressure:
   per module, compute
   commits_touching_module / module_LOC
   → Change Pressure
     (separates "endured" from "frozen" Survival)

6. Cleanup detection:
   Detect commits rewriting the
   previous (other) commit
   → Debt Cleanup

7. Ownership map:
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

1. **Code review vocabulary shifts.** "This code is clean" becomes "This code contributes to the Design layer / seems likely to survive robustly / is written in a high-pressure region."
2. **Flashy pre-launch contribution separates from contribution that survives six months later.** Short-term and long-term output become visible on different indicators.
3. **Refactoring becomes evaluable for the first time.** With a Debt Cleanup axis, "this quarter we reduced debt by 30% through cleanup" becomes sayable.
4. **Staffing a new project changes.** Whether you need an Anchor, a Cleaner, or a Producer becomes an axis-level decision.
5. **Post-departure evaluation of what someone left behind shifts.** "How much of the code someone who left wrote is still alive?" becomes an ex-post metric for their contribution to the organization.

The fifth is especially important. Conventional organizations are **built to forget the contributions of people who left**. With structural observation, what someone left in the Design layer can be honored in numbers, over time. This is also proper respect for the person.

## What's Next

We now have axes for reading code as structure. The next question: **who writes the code visible on those axes?**

View people as "ability" and you can't observe them from outside. View people as "types of contribution to structure" and you can. The same person contributes to structure in wildly different axis distributions.

Next chapter: three basic types — Anchor / Producer / Mass — and their derivatives. These aren't personality types. They're **patterns emerging from structural observation**. Read them with that in mind.

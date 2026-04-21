---
title: "Git Archaeology #8 — Engineering Relativity: Why the Same Engineer Gets Different Signals"
series: "Git Archaeology"
published: true
description: "Chapter 8 of Engineering Impact Signal. The same engineer produces different EIS signals in different codebases — and that's not a bug, it's physics."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/cover-ch8.png?v=4
---

*The same object is lighter on the Moon and heavier on Jupiter. The same thing happens in codebases.*

![Same engineer, different signals across repos](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch8-iconic.png?v=4)

## Previously

In [Chapter 7](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0), I talked about the universe-like structure of codebases — gravity, four forces, and "seasoned, good gravity."

This chapter is about another fundamental property of that gravity.

---

## Gravity Changes with the Universe

Looking at EIS results across different codebases, I noticed something.

**Gravity changes depending on the universe.**

EIS measures "how much gravity you created" in a codebase. But gravity has one critical property:

**It depends on the space it exists in.**

In physics, Earth, the Moon, and Jupiter each have different gravitational fields. The same object becomes lighter or heavier depending on where it is.

The same phenomenon occurs in codebases.

**The same engineer gets different EIS signals in different codebases.**

---

## Mature Universes and Young Universes

In a mature codebase:

- Structure is stable
- Architects already exist
- Abstractions are well-established
- "Seasoned, good gravity" is already present

In such environments, creating new gravity is not easy. The stronger the existing structure, the more energy it takes to shift the center. **EIS signals are harder to raise.**

In a structurally weak codebase:

- No central structure exists
- Design is fragmented
- Abstractions are lacking

In such environments, new gravity forms easily. The first person to introduce decent design becomes an Architect overnight. **EIS signals are easier to raise.**

---

## EIS Is Not an Absolute Value

This means EIS is **not an absolute value**.

EIS is determined not by an engineer's ability alone, but by the **interaction between the engineer and the codebase's gravitational field**.

This is, in a sense —

**Engineering Relativity.**

The same engineer, in a different universe, produces different gravity.

---

## The Trap of Raw Numbers

This has important implications for engineering evaluation.

Imagine an engineer whose signals look like this:

![Repo Scores](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch8-repo-scores.png?v=4)

Naturally, 60 looks "better."

But if **Repo A has an extremely strong gravitational field** — multiple Architects, highly refined structure, battle-tested abstractions — then **35 in that context may actually be remarkable**.

There's a "normalization trap" here. EIS's relative normalization means the top contributor in each team reaches 100 — so the top signal in one repo might be mediocre in another. But this chapter's point is more fundamental than normalization mechanics. Normalization is a calculation issue; Engineering Relativity is a **structural** issue.

**The codebase itself changes the *meaning* of the signal.**

That's Engineering Relativity.

Let me be explicit about something important:

**EIS does not directly measure an engineer's ability. It measures their impact within a code universe.**

Ability and impact are different things. A highly capable engineer may show modest impact in a universe with strong existing gravity. An average engineer may show outsized impact in a young universe. What EIS measures is "how much gravity did this engineer create in this universe" — not "how talented is this engineer."

---

## Reading EIS with Relativity in Mind

How do you account for this relativity when reading EIS? Here are some approaches.

### 1. Check Team Classification

Look at `eis analyze --team`:

![Structure Comparison](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch8-structure-comparison.png?v=4)

Impact: 40 inside an Architectural Engine and Impact: 40 inside an Unstructured team have completely different meanings.

### 2. Look at Architect Density

The more Architects on a team, the harder it is to raise your Design axis. This is a natural consequence of relative normalization. Reaching Design: 60 in a team with three Architects is likely harder than reaching Design: 100 in a team with none.

### 3. Use `--per-repo` for Cross-Repo Analysis

```bash
❯ eis analyze --recursive --per-repo ~/workspace
```

The `--per-repo` flag profiles each repository independently and produces a cross-repo comparison table. Producer in one repo, Architect in another — that pattern reveals adaptability and latent capability.

### 4. Watch "Gravitational Field Changes" in Timelines

```bash
❯ eis timeline --span 6m --periods 0 --recursive ~/workspace
```

Codebase structure isn't static. Member departures, refactoring, new features — these shift the gravitational field. In timelines, you can distinguish "engineers whose signals rise when structure weakens" from "engineers who maintain stable signals regardless of structural strength."

---

## The Reproducibility of Architects

Looking at EIS across multiple codebases, you notice a certain type of engineer exists. **Engineers who create gravity no matter what universe they're in.**

Different codebase. Different team. Different tech stack. They still build **structural centers**.

This might be called **Architect Reproducibility**.

When you analyze an entire workspace with `--recursive --per-repo`, an engineer who is consistently Architect across multiple repositories has "general-purpose design capability" that doesn't depend on any specific codebase.

Conversely, an engineer who is Architect in only one repository is creating gravity within that repository's specific context. This is also valuable, but it's a different kind of strength.

EIS `--per-repo` analysis makes this reproducibility **numerically verifiable**:

![Per-Repo Breakdown](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ch8-per-repo-breakdown.png?v=4)

---

## Gravitational Lensing: When Others' Signals Reveal Your Gravity

There's a subtler phenomenon worth noting — one borrowed from astrophysics.

In physics, you can detect massive objects not by looking at them directly, but by observing how they bend the light of objects behind them. This is **gravitational lensing**.

In codebases, something similar happens. An Architect's gravity is sometimes most visible not in their own signals, but in how it **shapes everyone else's signals**.

When a strong Architect is present:
- Other engineers' Survival signals may be lower (the Architect's code dominates blame)
- The team's Design axis distribution is skewed (one person absorbs most architectural changes)
- New joiners' signals reveal a characteristic "ramp-up curve" — they start low and gradually contribute to the existing structure

When that Architect leaves:
- Multiple engineers' signals shift simultaneously
- Design Vacuum risk appears
- The "flattening" of signal distributions indicates the loss of a gravitational center

You can observe this in `eis timeline --team`: the moment a gravitational center disappears, the entire team's metrics ripple. **The gravity was real — you just needed to look at its effects on others to see its full shape.**

---

## The Laws of Physics Are Not Uniform Across Universes

Engineering Relativity has one more deep implication.

**Each universe has its own laws of physics.**

In a universe built on a certain framework, structure is implicitly delegated to the framework itself. There's less design freedom, but a small team can bootstrap a universe quickly. Observed through EIS, the Design axis tends to be low across the board, with Production dominating.

In a universe built with a language whose type system is highly expressive, design decisions are explicitly inscribed in code. Interface design, constraints expressed through types, layers of abstraction — all of it is recorded in commits. Observed through EIS, the Design axis tends to be higher, and Survival stabilizes.

This is not about superiority. **It's about different laws of physics.**

In a small-to-medium universe where complexity is manageable, delegating structure to a framework is the right call. The universe runs efficiently with minimal gravity. Most engineers know this intuitively.

But when the universe expands, complexity explodes, and many engineers begin working on it simultaneously — **implicit structure can no longer maintain order.** Explicit design decisions, inscribed in code, become necessary to resist entropy.

The question of "which laws of physics suit which scale of universe" has been an **aerial battle** for years. "That tech choice was right." "No, it was wrong." — backed by nothing but experience and gut feeling.

EIS might bring observational data to this aerial battle.

By observing universes with different physical laws side by side — comparing Design axis, Survival axis, Robust Survival, and team structure tendencies — it may become possible to test hypotheses like "beyond a certain scale, type system expressiveness has a significant impact on Survival" using commit light.

Furthermore — language and type system choices **influence culture**. A team whose culture is to express constraints through types and a team whose culture is to guarantee correctness through tests will produce Architects with different characteristics and Entropy Fighters with different behaviors. The laws of physics of the universe shape the ecosystem of engineers who live within it.

This is still a hypothesis. But the results of observing 29 OSS projects and 55,000 engineers across universes are [beginning to show glimpses](https://github.com/machuz/eis/blob/main/research/oss-gravity-map/analysis/cross-language-gravity.md). Gravity concentration varies by **4.8x** between language families.

---

## Great engineers create gravity in every universe.

Truly great engineers create gravity in every universe.

But that gravity looks different depending on the universe.

**That's Engineering Relativity.**

---


← [Chapter 7: Observing the Universe of Code](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0) | [Chapter 9: Origin →](https://dev.to/machuz/git-archaeology-9-collapse-good-architects-and-black-hole-engineers-1dcn)

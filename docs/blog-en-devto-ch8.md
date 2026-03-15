---
title: "Git Archaeology #8 — Engineering Relativity: Why the Same Engineer Gets Different Scores"
published: true
description: "Chapter 8 of Engineering Impact Score. The same engineer produces different EIS scores in different codebases — and that's not a bug, it's physics."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/cover-ch8.png?v=4
---

*The same object is lighter on the Moon and heavier on Jupiter. The same thing happens in codebases.*

![Same engineer, different scores across repos](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch8-iconic.png?v=4)

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

**The same engineer gets different EIS scores in different codebases.**

---

## Mature Universes and Young Universes

In a mature codebase:

- Structure is stable
- Architects already exist
- Abstractions are well-established
- "Seasoned, good gravity" is already present

In such environments, creating new gravity is not easy. The stronger the existing structure, the more energy it takes to shift the center. **EIS scores are harder to raise.**

In a structurally weak codebase:

- No central structure exists
- Design is fragmented
- Abstractions are lacking

In such environments, new gravity forms easily. The first person to introduce decent design becomes an Architect overnight. **EIS scores are easier to raise.**

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

Imagine an engineer whose scores look like this:

![Repo Scores](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch8-repo-scores.png?v=4)

Naturally, 60 looks "better."

But if **Repo A has an extremely strong gravitational field** — multiple Architects, highly refined structure, battle-tested abstractions — then **35 in that context may actually be remarkable**.

There's a "normalization trap" here. EIS's relative normalization means the top contributor in each team scores 100 — so the top score in one repo might be mediocre in another. But this chapter's point is more fundamental than normalization mechanics. Normalization is a calculation issue; Engineering Relativity is a **structural** issue.

**The codebase itself changes the *meaning* of the score.**

That's Engineering Relativity.

Let me be explicit about something important:

**EIS does not directly measure an engineer's ability. It measures their impact within a code universe.**

Ability and impact are different things. A highly capable engineer may show modest impact in a universe with strong existing gravity. An average engineer may show outsized impact in a young universe. What EIS measures is "how much gravity did this engineer create in this universe" — not "how talented is this engineer."

---

## Reading EIS with Relativity in Mind

How do you account for this relativity when reading EIS? Here are some approaches.

### 1. Check Team Classification

Look at `eis analyze --team`:

![Structure Comparison](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch8-structure-comparison.png?v=4)

Total: 40 inside an Architectural Engine and Total: 40 inside an Unstructured team have completely different meanings.

### 2. Look at Architect Density

The more Architects on a team, the harder it is to raise your Design axis. This is a natural consequence of relative normalization. Scoring Design: 60 in a team with three Architects is likely harder than scoring Design: 100 in a team with none.

### 3. Use `--per-repo` for Cross-Repo Analysis

![Per-Repo Analysis](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch8-bash-per-repo.png?v=4)

The `--per-repo` flag scores each repository independently and produces a cross-repo comparison table. Producer in one repo, Architect in another — that pattern reveals adaptability and latent capability.

### 4. Watch "Gravitational Field Changes" in Timelines

![Timeline Command](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch8-bash-timeline.png?v=4)

Codebase structure isn't static. Member departures, refactoring, new features — these shift the gravitational field. In timelines, you can distinguish "engineers whose scores rise when structure weakens" from "engineers who maintain stable scores regardless of structural strength."

---

## The Reproducibility of Architects

Looking at EIS across multiple codebases, you notice a certain type of engineer exists. **Engineers who create gravity no matter what universe they're in.**

Different codebase. Different team. Different tech stack. They still build **structural centers**.

This might be called **Architect Reproducibility**.

When you analyze an entire workspace with `--recursive --per-repo`, an engineer who is consistently Architect across multiple repositories has "general-purpose design capability" that doesn't depend on any specific codebase.

Conversely, an engineer who is Architect in only one repository is creating gravity within that repository's specific context. This is also valuable, but it's a different kind of strength.

EIS `--per-repo` analysis makes this reproducibility **numerically verifiable**:

![Per-Repo Breakdown](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch8-per-repo-breakdown.png?v=4)

---

## Gravitational Lensing: When Others' Scores Reveal Your Gravity

There's a subtler phenomenon worth noting — one borrowed from astrophysics.

In physics, you can detect massive objects not by looking at them directly, but by observing how they bend the light of objects behind them. This is **gravitational lensing**.

In codebases, something similar happens. An Architect's gravity is sometimes most visible not in their own scores, but in how it **shapes everyone else's scores**.

When a strong Architect is present:
- Other engineers' Survival scores may be lower (the Architect's code dominates blame)
- The team's Design axis distribution is skewed (one person absorbs most architectural changes)
- New joiners' scores reveal a characteristic "ramp-up curve" — they start low and gradually contribute to the existing structure

When that Architect leaves:
- Multiple engineers' scores shift simultaneously
- Design Vacuum risk appears
- The "flattening" of score distributions signals the loss of a gravitational center

You can observe this in `eis timeline --team`: the moment a gravitational center disappears, the entire team's metrics ripple. **The gravity was real — you just needed to look at its effects on others to see its full shape.**

---

## Great engineers create gravity in every universe.

Truly great engineers create gravity in every universe.

But that gravity looks different depending on the universe.

**That's Engineering Relativity.**

---

### Series

- [Chapter 1: Measuring Engineering Impact from Git History Alone](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)
- [Chapter 2: Beyond Individual Scores: Measuring Team Health from Git History](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f)
- [Chapter 3: Two Paths to Architect: How Engineers Evolve Differently](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga)
- [Chapter 4: Backend Architects Converge: The Sacred Work of Laying Souls to Rest](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d)
- [Chapter 5: Timeline: Scores Don't Lie, and They Capture Hesitation Too](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5)
- [Chapter 6: Teams Evolve: The Laws of Organization Revealed by Timelines](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4lei)
- [Chapter 7: Observing the Universe of Code](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0)
- **Chapter 8: Engineering Relativity: Why the Same Engineer Gets Different Scores**
- [Chapter 9: Origin: The Big Bang of Code Universes](https://dev.to/machuz/git-archaeology-9-collapse-good-architects-and-black-hole-engineers-1dcn)
- [Chapter 10: Dark Matter: The Invisible Gravity](https://dev.to/machuz/git-archaeology-10-dark-matter-the-invisible-gravity-45ne)
- [Chapter 11: Entropy: The Universe Always Tends Toward Disorder](https://dev.to/machuz/git-archaeology-11-entropy-the-universe-always-tends-toward-disorder-ak9)
- [Chapter 12: Collapse: Good Architects and Black Hole Engineers](https://dev.to/machuz/git-archaeology-12-collapse-good-architects-and-black-hole-engineers-3fed)
- [Chapter 13: Cosmology of Code](https://dev.to/machuz/git-archaeology-13-cosmology-of-code-dci)

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full.png?v=2)

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLI tool, formulas, and methodology all open source. `brew tap machuz/tap && brew install eis` to install.


If this was useful: [❤️ Sponsor on GitHub](https://github.com/sponsors/machuz)

---

← [Chapter 7: Observing the Universe of Code](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0) | [Chapter 9: Origin →](https://dev.to/machuz/git-archaeology-9-collapse-good-architects-and-black-hole-engineers-1dcn)

---
title: "Git Archaeology #12 — Collapse: Good Architects and Black Hole Engineers"
published: true
description: "Chapter 12 of Engineering Impact Score. Not all strong gravity is good gravity. Some engineers create structure that survives their departure. Others create gravity that collapses without them."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/cover-ch12.png?v=4
---

*The universe has another property. Collapse.*

![Good Architect vs Black Hole Engineer](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch12-iconic.png?v=4)

## Previously

In [Chapter 11](https://dev.to/machuz/git-archaeology-11-entropy-the-universe-always-tends-toward-disorder-ak9), I wrote about Entropy — the universe always tends toward disorder.

This chapter is about another property of gravity.

Collapse.

---

## Stars Are Not Forever

The universe has another property.

Collapse.

Stars are not forever. Galaxies are not forever.

When gravity breaks down, the structure of the universe changes in an instant.

The same phenomenon occurs in codebases.

---

## When an Architect Leaves

Architects create universes.

They define design, create abstractions, organize dependencies, and build gravitational centers.

But here's the critical point.

**A truly great Architect designs for "the universe after they're gone."**

In a universe built by a good Architect, order is maintained even after they leave.

Because the structure remains.

The gravitational field of the design persists in the universe.

---

## Black Hole Engineer

But not all strong gravity is good gravity.

The universe has black holes. Black holes have extremely strong gravity. But their gravity doesn't create structure — it **swallows everything**.

The same type of engineer exists in code universes.

**Black Hole Engineer.**

Their characteristics:

![Black Hole Pattern](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch12-black-hole-pattern.png?v=4)

High technical skill. High output. Strong influence.

But — they don't create structure.

Instead — dependency concentrates.

---

## A Black Hole Universe

Around a Black Hole Engineer, this happens:

Massive services. Massive utilities. Massive modules.

Work concentrates, dependencies concentrate, code concentrates.

The result — **the center of the universe becomes one engineer.**

![Good Architect vs Black Hole](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch12-good-vs-blackhole.png?v=4)

A good Architect distributes gravity. Leaves structure. Gives the universe order.

A Black Hole Engineer concentrates gravity. Becomes the center of the universe themselves.

---

## Collapse

The problem is when that engineer leaves.

When the black hole disappears, the center of the universe disappears.

What happens then?

![Collapse Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch12-collapse-timeline.png?v=4)

Design decisions stop. Dependencies break. Nobody can touch the code.

The code universe collapses in an instant.

---

## Good Gravity

A good Architect is different from a black hole.

They don't concentrate gravity. They distribute structure.

They share abstractions, clarify boundaries, and leave order in the universe.

So — even after they leave, the universe doesn't collapse.

This is **seasoned, good gravity**.

Remember Chapter 4's "laying souls to rest." A good Architect can be laid to rest. Their code lives on after departure. Survival 100 is proof that the structure remains.

**A Black Hole Engineer cannot be laid to rest.**

Because the code universe collapses the moment they leave.

---

## Using EIS to Prevent Collapse

### 1. Monitor Bus Factor

Teams where `eis analyze --team` shows a Bus Factor near 1 are at risk of Black Hole collapse.

![Bus Factor Check](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch12-bash-team.png?v=4)

Bus Factor = 1 means "if one person leaves, it collapses." This is the clearest sign of a Black Hole.

### 2. Detect Indispensability Concentration

Use `--per-repo` to examine individual score distributions.

![Per-Repo Detection](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch12-bash-per-repo.png?v=4)

One person with extremely high Indispensability while everyone else is extremely low — this distribution is the signature of a Black Hole.

### 3. Watch for "One Person Stays Architect Forever" in Timelines

![Timeline Detection](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch12-bash-timeline.png?v=4)

A good Architect's timeline shows an Architect → Producer transition (like Engineer J in Chapter 5). Once the structure is built, they produce on top of it.

A Black Hole Engineer's timeline shows **permanent Architect**. They never release the structure. They keep concentrating gravity.

### 4. Judge Gravity Quality Through Surrounding Scores

Use the gravitational lensing effect from Chapter 8.

Around a good Architect:
- Teammates' Design scores gradually rise (they learn the structure and start contributing)
- New joiners ramp up quickly (the structure is clear and understandable)

Around a Black Hole Engineer:
- Teammates' Design scores stay low (they can't touch — or don't dare touch — the structure)
- New joiners ramp up slowly (you have to ask one person to understand anything)

**The quality of gravity is reflected in the surrounding scores.**

---

## Preventing Collapse Is a Leader's Job

EIS can detect collapse risk. But detection alone doesn't prevent collapse.

Preventing collapse is a leader's job.

Specifically:

- When you find Bus Factor = 1, **deliberately expand pair work and code review scope**
- When you find Indispensability concentration, **create time for that engineer to teach**
- When you find a permanent Architect pattern, **build mechanisms to distribute design decisions**

EIS shows you the universe's structure. How to reshape that structure is a human decision.

---

## Regeneration After Collapse — Engineers Who Can Replace a Black Hole

Collapse isn't necessarily the end.

Just as Engineer I created a new universe in Chapter 5, there are engineers who can bring new gravity to a collapsed universe.

These engineers have specific traits:

- **Architect Reproducibility** (Chapter 8). They can create structure in any universe
- They can **read** existing gravitational fields. They understand collapsed structures and grasp what was lost
- They choose designs that **distribute** gravity. They don't repeat the Black Hole mistake

In timelines, the pattern looks like this:

When such an engineer joins a post-collapse team:
- Team classification recovers from Unstructured → Guardian → Balanced
- Bus Factor rises from 1 to 2, then 3
- Multiple members' Design scores start rising simultaneously

**Only an Architect who distributes structure can turn collapse into regeneration.**

What's needed to replace a Black Hole isn't the same strength of gravity. It's a **different quality** of gravity.

---

## Stars are not forever. That's why structure matters.

In the universe, when a star dies, the elements it created remain. Iron, oxygen, carbon — all forged in the star's nuclear fusion.

A good Architect is the same. What remains after they leave isn't code — it's **structure**.

What a Black Hole Engineer leaves behind is — void.

**Stars are not forever. That's why structure matters.**

---

### Series

- [Chapter 1: Measuring Engineering Impact from Git History Alone](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)
- [Chapter 2: Beyond Individual Scores: Measuring Team Health from Git History](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f)
- [Chapter 3: Two Paths to Architect: How Engineers Evolve Differently](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga)
- [Chapter 4: Backend Architects Converge: The Sacred Work of Laying Souls to Rest](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d)
- [Chapter 5: Timeline: Scores Don't Lie, and They Capture Hesitation Too](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5)
- [Chapter 6: Teams Evolve: The Laws of Organization Revealed by Timelines](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4lei)
- [Chapter 7: Observing the Universe of Code](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0)
- [Chapter 8: Engineering Relativity: Why the Same Engineer Gets Different Scores](https://dev.to/machuz/git-archaeology-8-engineering-relativity-why-the-same-engineer-gets-different-scores-5dnl)
- [Chapter 9: Origin: The Big Bang of Code Universes](https://dev.to/machuz/git-archaeology-9-collapse-good-architects-and-black-hole-engineers-1dcn)
- [Chapter 10: Dark Matter: The Invisible Gravity](https://dev.to/machuz/git-archaeology-10-dark-matter-the-invisible-gravity-45ne)
- [Chapter 11: Entropy: The Universe Always Tends Toward Disorder](https://dev.to/machuz/git-archaeology-11-entropy-the-universe-always-tends-toward-disorder-ak9)
- **Chapter 12: Collapse: Good Architects and Black Hole Engineers**
- [Chapter 13: Cosmology of Code](https://dev.to/machuz/git-archaeology-13-cosmology-of-code-dci)

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full.png)

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLI tool, formulas, and methodology all open source. `brew tap machuz/tap && brew install eis` to install.


If this was useful: [❤️ Sponsor on GitHub](https://github.com/sponsors/machuz)

---

← [Chapter 11: Entropy](https://dev.to/machuz/git-archaeology-11-entropy-the-universe-always-tends-toward-disorder-ak9) | [Chapter 13: Cosmology of Code →](https://dev.to/machuz/git-archaeology-13-cosmology-of-code-dci)

---
title: "Git Archaeology — A Complete Theory of Software Universes"
published: true
description: "16 chapters condensed: how git log and git blame reveal engineer gravity, team health, software cosmology, and why AI creates stars but not gravity."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/cover-ch0.png?v=1
---

*What if your git history could tell you who really shaped your codebase?*

Over 17 chapters (0–16), I built a theory of software — starting from a simple question and ending with a cosmology. This is the complete arc, condensed into one read.

---

## 0. The Telescope

It started with a frustration: **"This team is strong. But I had no words to explain it."**

Over the years, I've recruited several strong engineers by reaching out personally. Why did they come? I believe it's because they felt: **"This person actually sees my work."** Not commit counts — but whether code survives, whether you contribute to architecture, whether you clean up debt.

I had that observer's eye. EIS is that **eye turned into an open-source telescope.**

Math has complexity theory. Programming languages have type theory. But software architecture still lacks a unified scientific foundation — 30 years of academic attempts (ADLs, ATAM) remain fragmented and haven't made it into practice. In the age of AI, what matters most is structure. If EIS can become a tool for **making architecture a science**, I'd be glad.

So I built this telescope. Using nothing but `git log` and `git blame`, I quantified engineering impact across 7 axes:

| Axis | What it measures |
|---|---|
| **Production** | Volume of output (time-decayed) |
| **Quality** | Self-revision discipline |
| **Survival** | How long code lives without being rewritten |
| **Design** | Structural influence — files touched by others |
| **Breadth** | Reach across the codebase |
| **Debt Cleanup** | Cleaning up others' debt |
| **Indispensability** | How much of the codebase you "own" via blame |

From these, three topologies emerge: **Role** (Architect / Anchor / Producer), **Style** (Builder / Balanced / Mass), and **State** (Active / Growing / Fragile / Former).

The numbers were eerily accurate. Silent heroes surfaced. Hidden risks became visible.

> Quantify what you can. Qualitatively supplement what you can't. That order matters.

*Deep dive: [Chapter 0 — Introduction](https://dev.to/machuz/git-archaeology-0-what-if-git-history-could-tell-you-who-your-strongest-engineers-are-5397), [Chapter 1](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c), [Chapter 2](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f)*

---

## II. Evolution

When I added timelines — quarterly snapshots of signals — stories emerged.

An engineer's Role shifts from Producer to Anchor to Architect. Another's signals plateau — not from stagnation, but from **strategic patience**. A departure trajectory becomes visible three quarters before anyone notices.

Cold numbers are what reflect the hottest stories.

From these timelines, I extracted evolution laws:

- **Builder is a prerequisite for Architect** — you cannot design what you haven't built
- **Producer is metabolism, not regression** — sometimes the best Architects go back to producing
- **Backend Architects converge; Frontend Architects branch** — different gravitational physics
- **Departed Architects leave "souls" in the code** — laying them to rest through Debt Cleanup is sacred work

*Deep dive: [Chapter 3](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga), [Chapter 4](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d), [Chapter 5](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5), [Chapter 6](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4lei)*

---

## III. Cosmology

The deeper I looked, the more codebases looked like universes.

| Physics | Software |
|---|---|
| Big Bang | First commit — initial conditions determine everything |
| Stars | Engineers |
| Gravity | Structural influence — great engineers bend the gravity of codebases |
| Dark Matter | Work invisible in commits: reviews, design discussions, mentoring, culture |
| Entropy | Code rot — left alone, code always tends toward disorder |
| Black Holes | Engineers who concentrate dependency instead of distributing structure |
| Collapse | What happens when a Black Hole Engineer leaves |

This isn't metaphor. It's **structural correspondence**.

**Gravity** is the central concept. Not all code is equal. An engineer who creates a module boundary that 50 files depend on has generated gravity — a structural force that shapes everything around it.

**Dark matter** is what the telescope can't see. Culture, mentoring, design discussions, planning — these never appear in commits, but they determine the entire structure of the universe. A telescope must know its own limits.

**Entropy** is the default. Software always rots. Development is fundamentally a battle against entropy. Every EIS axis maps to either increasing entropy (Production) or fighting it (Quality, Design, Survival).

**Collapse** is what happens when gravity concentrates instead of distributes. A Black Hole Engineer writes great code — but when they leave, the codebase collapses instantly. A good Architect designs for the universe after they're gone.

> Stars are not forever. That's why structure matters.

*Deep dive: [Chapter 7](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0), [Chapter 8](https://dev.to/machuz/git-archaeology-8-engineering-relativity-why-the-same-engineer-gets-different-scores-5dnl), [Chapter 9](https://dev.to/machuz/git-archaeology-9-collapse-good-architects-and-black-hole-engineers-1dcn), [Chapter 10](https://dev.to/machuz/git-archaeology-10-dark-matter-the-invisible-gravity-45ne), [Chapter 11](https://dev.to/machuz/git-archaeology-11-entropy-the-universe-always-tends-toward-disorder-ak9), [Chapter 12](https://dev.to/machuz/git-archaeology-12-collapse-good-architects-and-black-hole-engineers-3fed)*

---

## IV. Civilization

Most codebases die within a few years. Entropy wins. The team changes. Knowledge scatters. Someone says "let's rewrite from scratch."

But a few survive. **Linux. Git. PostgreSQL. React.** Their creators left, contributors turned over across generations, and the structure persisted. These are not repositories. They are **civilizations**.

Civilization requires three roles:

```
Civilization =
  Architect  → creates gravity (structure)
  + Anchor   → maintains order (stability)
  + Producer → expands territory (growth)
```

Remove any one and the equation breaks:

| Missing | Result |
|---|---|
| No Architect | Growth without structure — entropy wins |
| No Anchor | Beautiful but fragile — collapses when Architect leaves |
| No Producer | Structure without growth — fossilization |

The most important engineers build systems that don't need them. That's the civilization test: **does the structure survive after the Architect leaves?**

*Deep dive: [Chapter 13](https://dev.to/machuz/git-archaeology-13-cosmology-of-code-dci), [Chapter 14](https://dev.to/machuz/git-archaeology-14-civilization-why-only-some-codebases-become-civilizations-2nl3)*

---

## V. AI Creates Stars, Not Gravity

AI is a starburst — generating code at unprecedented rates.

But code without structure is entropy. No matter how many stars form, without gravity, no galaxy is born.

In the age of AI, the scarcest engineering capability shifts:

| Era | Scarcest capability |
|---|---|
| Pre-AI | Writing code (Production) |
| Post-AI | Generating gravity (Design, Survival) |

The engineers who thrive are not those whose primary differentiator is implementation speed. They are the ones who generate gravity — **Code Architects** who create structure, and **Code Custodians** who fight entropy.

AI becomes a **gravity amplifier**. An Architect who once shaped one codebase can now shape ten. The muscle that matters is not the muscle for writing code — it's the muscle for generating gravity.

> AI creates stars. But engineers are the ones who shape gravity.

*Deep dive: [Chapter 15](https://dev.to/machuz/git-archaeology-15-ai-creates-stars-not-gravity-4i05)*

---

## VI. The Engineers Who Shape Gravity

Software engineering exists between two kinds of time.

Git remembers the past. AI imagines the future.

Between them, engineers shape gravity — creating structure, creating order, keeping the system from collapsing.

Where gravity exists, code is not mere fragments — it becomes structure. When structure emerges, systems persist beyond time. That is not just a repository. **It becomes a civilization.**

```
Git remembers the past.
AI imagines the future.

Between them, engineers shape gravity.

And from that gravity,
software civilizations emerge.
```

*Deep dive: [Chapter 16 — Final](https://dev.to/machuz/git-archaeology-16-the-engineers-who-shape-gravity-3fmi)*

---

Does your code universe have gravity?

Point the telescope and see.

```bash
❯ brew tap machuz/tap && brew install eis
❯ eis analyze --recursive ~/your-workspace
```

```
      ✦       *        ✧

       ╭────────╮
      │    ✦     │
       ╰────┬───╯
   .        │
            │
         ___│___
        /_______\

   ✧     the Git Telescope     ✦
```

---

### Full Series

- **[Chapter 0: What If Git History Could Tell You Who Your Strongest Engineers Are?](https://dev.to/machuz/git-archaeology-0-what-if-git-history-could-tell-you-who-your-strongest-engineers-are-5397)** — Introduction
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
- [Chapter 12: Collapse: Good Architects and Black Hole Engineers](https://dev.to/machuz/git-archaeology-12-collapse-good-architects-and-black-hole-engineers-3fed)
- [Chapter 13: Cosmology of Code](https://dev.to/machuz/git-archaeology-13-cosmology-of-code-dci)
- [Chapter 14: Civilization — Why Only Some Codebases Become Civilizations](https://dev.to/machuz/git-archaeology-14-civilization-why-only-some-codebases-become-civilizations-2nl3)
- [Chapter 15: AI Creates Stars, Not Gravity](https://dev.to/machuz/git-archaeology-15-ai-creates-stars-not-gravity-4i05)
- [Chapter 16: The Engineers Who Shape Gravity](https://dev.to/machuz/git-archaeology-16-the-engineers-who-shape-gravity-3fmi) — **Final**

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png?v=2)

**GitHub**: [eis](https://github.com/machuz/eis) — CLI tool, formulas, and methodology all open source. `brew tap machuz/tap && brew install eis` to install.

If this was useful: [Sponsor on GitHub](https://github.com/sponsors/machuz)

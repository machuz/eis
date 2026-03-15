---
title: "Git Archaeology #4 — Backend Architects Converge: The Sacred Work of Laying Souls to Rest"
published: true
description: "Chapter 4 of Engineering Impact Score. Backend teams evolve differently — and departed Architects leave souls that must be laid to rest."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/cover-ch4.png?v=4
---

*Departed Architects leave souls in the codebase. Laying them to rest is sacred work.*

![Convergence — multiple paths to one architectural destination](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-iconic.png?v=4)

## Previously

In [Chapter 3](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga), I explored a frontend team and discovered two evolution paths for Architects:

- **Inheritance Architect**: Anchor → guards and refines existing structure
- **Emergent Architect**: High-Gravity Producer → creates new gravitational fields

In frontend, multiple aesthetics of structure can coexist. This means multiple Architect candidates can emerge. **Structural competition** happens.

But when I ran EIS on a Backend team, I saw a completely different landscape.

---

## The Backend Team Portrait

Here's the Backend team's data:

![Backend Team](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-backend-team.png?v=4)

And the team metrics:

![Team Classification](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-team-classification.png?v=4)

**1 Architect. 3 Anchors. 0 Producers.**

This is a completely different structure from frontend.

---

## Why Backend Architects Concentrate

In frontend, multiple Architects can emerge.

In backend, that rarely happens.

The reason is simple.

**Backend has "correct answers" for structure.**

For example, backend systems often have stable patterns like:

- Domain layer
- Presentation layer
- UseCase layer
- Infrastructure layer
- Transaction boundaries
- Event boundaries

These design patterns have reusable structures across many systems.

In other words:

| FE | Multiple aesthetics of structure exist |
|---|---|
| BE | Correct answers tend to exist |

Because of this difference, **Backend Architects concentrate rather than proliferate**.

Think of it like a solar system:

![Backend Structure](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-structure.png?v=4)

This structure isn't unusual for Backend teams.

---

## The Soul of a Departed Architect

Look at Y.Y.:

![Y.Y. profile](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-data-engineer-f.png?v=4)

Y.Y. is a **departed Architect**.

But even with time decay applied, **they're still ranked #2 on the team**.

What does this mean?

**The assets they created are still massively present in the codebase.**

---

### What Former State Means

In EIS, Former state is only detected **when significant assets remain**.

Former state conditions:

- Raw Survival (no time decay) is high
- Survival (with time decay) is low
- AND Design or Indispensability is high

This means: **"This person once shaped the structure. They're no longer active. But their code remains."**

You don't become Former just by leaving. **Only those who left assets worth leaving** become Former.

Y.Y. meets exactly these conditions.

---

### Still #2 Despite Time Decay

This is the emotional part.

EIS uses **time-decayed Survival**. Code from 2 years ago has a weight of just 0.02.

And yet Y.Y. is still #2.

This means **they continue to have massive influence on the codebase even after departure**.

This isn't "the afterimage of someone who was once strong."

**It's real assets that still support the codebase today.**

---

## Laying the Former to Rest

But Former state has another meaning.

**It's something that needs to be laid to rest.**

The team metrics show:

![Phase and Risk indicators](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-data-phase-risk.png?v=4)

Legacy-Heavy means **strong but historically heavy team**.

Departed members' code remains in large quantities. That's not inherently bad — good design is why it remains.

---

### Good Design Creates Common Sense

Here's an important fact.

**This team hasn't collapsed.**

Several modules that only Y.Y. touched still exist. Most of git blame belongs to the Former member.

Yet the team functions normally.

Why?

**Because those modules were built under well-organized design.**

We received verbal handoffs. But complete documentation or knowledge transfer didn't happen. Still, **the design embedded as structure in the code gives later engineers a certain understanding**.

This is what I call "good design creates common sense."

Excellent design doesn't necessarily require complete documentation or knowledge transfer. **The structure of the code itself communicates the module's intent and usage.**

Strong design leaves knowledge in structure, not in people. And that structure creates shared understanding across the team.

---

### Legacy-Heavy Converges

EIS currently handles quantitative metrics like history and code survival rates. "Common sense through design" — why Former members' code still functions healthily — can't be directly observed yet.

If that became possible, we could distinguish between "structurally healthy despite heavy history" and "truly dangerous dependency structures."

But we may not need to measure that far.

**Strong teams gradually replace Former members' code with their own. Legacy-Heavy resolves over time. It converges toward the right state.**

EIS can naturally capture this convergence process through Survival trends and Risk Ratio changes.

---

However, as time passes, **the risk that fewer people understand that code** increases.

That's why laying the Former to rest becomes necessary.

---

### Souls Get Absorbed into Debt

What does laying a Former to rest mean?

It means **being absorbed into Debt Cleanup**.

![Team average Debt Cleanup score](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-data-debt-avg.png?v=4)

Active members touch Former's code. Understand it. Fix it. Rewrite it.

Then Former's blame lines decrease. They're replaced by active members' lines.

This is **laying to rest**.

EIS visualizes this process:

- Former's Survival gradually decreases
- Active members' Debt Cleanup increases
- Team's Legacy-Heavy degree decreases

**This is sacred work.**

The departed Architect's soul, through active members' hands, gradually dissolves into the codebase.

---

## The True Meaning of Anchor

This Backend team has 3 Anchors:

- Y.Y. (Former)
- P.
- R.S.

Anchor isn't just "quality guardian."

**Anchor means structure-understanding engineer.**

Anchor traits:

- Deep understanding of existing structure
- Doesn't break structure
- Guards quality
- Maintains system integrity

In other words, Anchor is **an engineer who understands structure and produces on top of it**.

And in Backend, this Anchor becomes **the evolution path to Architect**.

---

## Backend Architect Evolution

In frontend:

![Frontend evolution path: Producer to Emergent Architect](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-diagram-fe-evolution.png?v=4)

But in Backend, this is more common:

![Backend evolution path: Producer to Inheritance Architect](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch3-diagram-be-evolution.png?v=4)

This is the **Convergent Evolution Model**.

---

### The Reproducing Architect

Backend Architects are characterized by **being able to reproduce structure**.

Say a design like this worked in one codebase:

![Design pattern: Domain + Presentation + UseCase + Infrastructure](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-data-design-pattern.png?v=4)

A skilled Backend Architect can **reproduce that structure in a different system**.

So Backend Architect means:

- Someone who creates structure
- **AND someone who can reproduce structure**

If FE Architects are the type who "create emergent new structures," Backend Architects are the type who "**purify and reproduce structures**."

---

## Producer Vacuum

Another interesting point.

This team has **no Producers**.

![Role distribution: Architect 1, Anchor 3, Producer 0](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-data-role-count.png?v=4)

Producer means "someone who doesn't fully understand structure but produces on top of it."

Without Producers, team structure looks like:

![Producer vacuum diagram](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-diagram-producer-vacuum.png?v=4)

This is a state I'd call **Producer Vacuum**.

With Producers:

![Three-layer structure: Architect, Anchor, Producer](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-diagram-three-layer.png?v=4)

A three-layer structure forms.

This is the most stable form for Backend teams.

---

## Architect Bus Factor = 1

EIS raises warnings for this team:

![Bus factor warning](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-data-bus-factor.png?v=4)

This means **Architect Bus Factor = 1**.

If the Architect leaves, team productivity density drops significantly.

This is the most typical risk for Backend teams.

In frontend, Emergent Architect candidates can mitigate this problem.

But in Backend, **because correct structural answers exist, Architects tend to concentrate**.

Result: Bus Factor risk increases.

---

## Prescription for This Team

What does this Backend team need?

1. **Fill the Producer layer**: Allow reverse flow from Anchor → Producer. Create an environment where people can produce without fully understanding structure
2. **Accelerate laying Former to rest**: Active members take over departed Architect's code. Consciously increase Debt Cleanup
3. **Open the Anchor → Architect path**: Let structure-understanding Anchors participate in design decisions

#2 is especially important.

Legacy-Heavy state isn't "bad." It's proof that **good design remains**.

But understanding it, taking it over, and rewriting when needed — **someone has to do that work**.

That work is visualized as Debt Cleanup.

---

## FE vs BE Evolution Models Compared

Summary:

| | FE | BE |
|---|---|---|
| Structure | Multiple aesthetics exist | Correct answers tend to exist |
| Architect | Can distribute | Tends to concentrate |
| Evolution Model | Branching (Emergent or Inheritance) | Convergent (Inheritance dominant) |
| Risk | Structural collision | Bus Factor concentration |

Neither is better.

**The nature of the domain determines the evolution model.**

EIS visualizes these differences and can suggest appropriate prescriptions for each.

---

## Being Honest

My score of 92.4 at #1 feels good — like my ability as an Architect is being validated.

But honestly, **I also work an absurd number of hours**.

And there's another thing that makes me happy.

**I could only achieve this Architect score because I inherited solid design from Y.Y. and made it even better.**

Backend has time-tested good design passed down through generations — structures refined by those who came before. I learned them and built on them. Y.Y. also learned them and built on them. When we met, we already shared the same understanding. That's why common sense emerged and inheritance became possible.

On top of that foundation, I added `delegateProcess` and `partProcess` — not textbook patterns, but ones I'm convinced are useful. I enhanced existing concepts that already had business impact. I organized and modeled new core business functionality well.

That accumulation is what these numbers represent.

---

I worked with P. before, at another company. We merged our design knowledge back then. So we share common understanding of each other's design approach. Whatever their current score shows, I know they have horsepower. **I know they have the ability to hit 80+ on our team.**

R.S. is doing really good work. When I invited them to join this team, they said "I lack confidence because I don't have much experience." But from my experience, I was absolutely certain they could do good work. So I invited them. **And now, they've become an Anchor.** An Anchor on an Elite team. That's something to brag about anywhere. Deeply moving.

---

And even more honestly, **what I really want to brag about is the team**.

![Elite team classification](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch4-data-elite.png?v=4)

That this team gets an Elite classification. That it functions despite carrying Legacy-Heavy weight. That it's moving forward while laying departed Architects' assets to rest.

**One of my motivations for building EIS? I wanted to brag about this team.**

I wanted to say "my team is strong." But "somehow strong" isn't convincing.

And critically, **I wanted to brag on a playing field where bias can't enter**.

When someone subjectively says "that person is strong," it might just be the evaluator's bias. Maybe the person who's good at politics gets high ratings.

But git history doesn't lie.

Does the code survive? Are they involved in design? Are they creating debt, or cleaning it up?

**Reverse-engineering from history leaves no room for bias.**

So I built the numbers on this playing field.

Of course, this isn't everything.

**Whether the product that emerges is good, whether it fits the market — that needs to be judged on a different axis.**

EIS measures technical impact on the codebase. But product value isn't determined by code quality alone. Is it reaching users? Is it viable as a business? That's a different conversation.

EIS visualizes "strength as a technical organization." What you build on top of that is a separate decision.

And those numbers matched my intuition. The team really was strong.

---

## Spec Changes and Robust

A team member asked:

> Do spec changes that overwrite code get reflected in Robust scores?

**Yes, they do.**

When spec changes cause code rewrites, the original author's Robust score drops.

This means **planner precision also affects these scores**.

If specs are unstable, code gets rewritten. If planning is sloppy, no matter how good the code an engineer writes, it disappears.

That this team achieves strong Robust scores isn't just an engineering achievement. **It's the achievement of the entire development organization, including planners.**

---

### Renovations and Scores

Another question came up:

> If we do a major renovation, will the people who wrote pre-renovation code see their scores drop?

**Yes.**

If renovation replaces previous code, the previous authors' scores drop.

But here's the key point.

**Can you adapt?**

- Can you weave your opinions into the post-renovation design?
- Can you produce scores even on top of the new good design?
- More broadly, **can you produce scores across multiple codebases**, whether they're good or bad?

That's what real engineering strength is.

Scoring high on one codebase is possible if you're lucky with your environment.

But **being able to adapt and generate gravity on any codebase, any team** — that's what a real engineer looks like.

EIS measures impact on a single codebase. But if you repeat that measurement across multiple codebases, you start seeing **reproducibility that transcends environment**.

---

## What This Discovery Means

Chapter 3 showed frontend's branching evolution model.

Chapter 4 showed backend's convergent evolution model.

And one more concept emerged: **laying the Former to rest**.

Departed Architects' souls remain in the codebase.

Active members take them over, understand them, rewrite them.

**That sacred work is visualized as Debt Cleanup.**

Cold numbers, it turns out, tell the most human stories.

That might be the essence of EIS.

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full.png?v=2)

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLI tool, formulas, and methodology all open source. `brew tap machuz/tap && brew install eis` to install.


If this was useful: [❤️ Sponsor on GitHub](https://github.com/sponsors/machuz)

---

### Series

- [Chapter 1: Measuring Engineering Impact from Git History Alone](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)
- [Chapter 2: Beyond Individual Scores: Measuring Team Health from Git History](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f)
- [Chapter 3: Two Paths to Architect: How Engineers Evolve Differently](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga)
- **Chapter 4: Backend Architects Converge: The Sacred Work of Laying Souls to Rest**
- [Chapter 5: Timeline: Scores Don't Lie, and They Capture Hesitation Too](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5)
- [Chapter 6: Teams Evolve: The Laws of Organization Revealed by Timelines](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4lei)
- [Chapter 7: Observing the Universe of Code](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0)
- [Chapter 8: Engineering Relativity: Why the Same P.ets Different Scores](https://dev.to/machuz/git-archaeology-8-engineering-relativity-why-the-same-engineer-gets-different-scores-5dnl)
- [Chapter 9: Origin: The Big Bang of Code Universes](https://dev.to/machuz/git-archaeology-9-collapse-good-architects-and-black-hole-engineers-1dcn)
- [Chapter 10: Dark Matter: The Invisible Gravity](https://dev.to/machuz/git-archaeology-10-dark-matter-the-invisible-gravity-45ne)
- [Chapter 11: Entropy: The Universe Always Tends Toward Disorder](https://dev.to/machuz/git-archaeology-11-entropy-the-universe-always-tends-toward-disorder-ak9)
- [Chapter 12: Collapse: Good Architects and Black Hole Engineers](https://dev.to/machuz/git-archaeology-12-collapse-good-architects-and-black-hole-engineers-3fed)
- [Chapter 13: Cosmology of Code](https://dev.to/machuz/git-archaeology-13-cosmology-of-code-dci)
- [Chapter 14: Civilization — Why Only Some Codebases Become Civilizations](https://dev.to/machuz/git-archaeology-14-civilization-why-only-some-codebases-become-civilizations-2nl3)
- [Chapter 15: AI Creates Stars, Not Gravity](https://dev.to/machuz/git-archaeology-15-ai-creates-stars-not-gravity-4i05)
- [Chapter 16: The Engineers Who Shape Gravity](https://dev.to/machuz/git-archaeology-16-the-engineers-who-shape-gravity-3fmi) — **Final**

---

← [Chapter 3: Two Paths to Architect](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga) | [Chapter 5: Timeline →](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5)

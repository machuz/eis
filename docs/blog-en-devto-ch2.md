---
title: "Git Archaeology #2 — Beyond Individual Signals: Measuring Team Health from Git History"
series: "Git Archaeology"
published: true
description: "Chapter 2 of Engineering Impact Signal. Team-level analysis — complementarity, risk ratio, productivity density — all from git data you already have."
tags: opensource, productivity, git, teams
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/cover-ch2.png?v=4
---

*Individual signals tell you whose signal is strong. Team health tells you whether the team will still be strong next quarter.*

![Team structure and health radar](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch2-iconic.png?v=4)

## Why Individual Signals Aren't Enough

A team where every member hits 80+ isn't necessarily strong. If everyone is a Producer, nobody is shaping architecture. Nobody is paying down debt. The codebase ships fast and rots faster.

Conversely, a team averaging 50 — but with one Architect, one Cleaner, and two Growing juniors — may be in a much healthier position.

**A strong team is not the sum of individual signals. It's about composition and complementarity.**

---

## Why Revenue Doesn't Show Engineering Health

"Revenue is growing, so engineering must be fine" — a dangerous assumption. Revenue measures **product-market fit**, not **engineering health**.

Revenue is a car's speed. Engineering health is the engine's condition. An engine can be failing and still produce speed — if you're going downhill.

Git history contains signals that revenue can't:

- **Code durability** — are you rewriting the same features every quarter?
- **Technical debt** — does adding 1 feature generate 2 bug fixes?
- **Bus factor** — how many modules die if one person leaves?

**Even with revenue growing, if Survival decline + Debt increase + Bus Factor concentration are progressing simultaneously, the organization will collapse at scale.**

---

## Seven Team Health Axes

`eis team` aggregates individual signals into team-level health:

```bash
❯ eis team --recursive ~/workspace
```

| Axis | What it measures | Key insight |
|---|---|---|
| **Complementarity** | Role diversity (Architect, Anchor, Cleaner, Producer, Specialist) | A team with only Producers gets 16. Full diversity hits 100 |
| **Growth Potential** | Growing members + Builder/Cleaner role models present | Without role models, juniors can't level up |
| **Sustainability** | Inverse of risk states (Former, Silent, Fragile) | Hidden drags on team velocity |
| **Debt Balance** | Average Debt Cleanup. Above 50 = team cleans more than it creates | Self-cleaning tendency |
| **Productivity Density** | Output per head, with small-team bonus | "This output from this few people" |
| **Quality Consistency** | Mean quality + low variance | A team averaging 80 but ranging 95–40 is not healthy |
| **Risk Ratio** | % in Former/Silent/Fragile state | Above 25% = warning. Above 50% = crisis |

> Formulas for each axis: [Whitepaper](https://github.com/machuz/eis)

---

## Team Classification — Galaxy Morphology

EIS classifies teams along **five axes**, derived bottom-up from individual topologies:

![Team Classification Flow](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/team-classification-flow.png?v=4)

| Axis | Derived from | Question |
|---|---|---|
| **Structure** | Role distribution | What structural roles exist? |
| **Culture** | Style distribution | How does this team work? |
| **Phase** | State distribution | Where in its lifecycle? |
| **Risk** | Health metrics | What risks does it carry? |
| **Character** | Composite of above 4 | What galaxy is this team? |

Character uses **galaxy morphology** — because a telescope describes the shape of galaxies, not their quality:

| Character | Galaxy | Meaning |
|---|---|---|
| **Spiral** | Spiral galaxy | Strong core + active star formation. Architecture and production both firing |
| **Elliptical** | Elliptical galaxy | Mature, stable, change-resistant. Low entropy |
| **Starburst** | Starburst galaxy | Explosive growth. High energy, structure still forming |
| **Nebula** | Stellar nursery | Next-generation engineers developing |
| **Irregular** | Irregular galaxy | No gravitational center. High output, no direction |
| **Dwarf** | Dwarf galaxy | Small but long-lived. Steady quality |
| **Collision** | Colliding galaxies | Structural disruption. Constantly firefighting |

> Full galaxy guide with astronomical explanations: [Galaxy Morphology Guide](https://orbit-d8x.pages.dev/galaxy-guide.html)

Classification is **weighted by Impact** — an Architect at 90 shapes the team's character far more than an Architect at 15. Strong signal carriers propagate more culture.

---

## Growth Model

EIS's Role classification maps to three layers:

![Growth Model](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch2-diagram-growth-model.png?v=4)

**Implementation** → **Stabilization** → **Design**

- Survival rising → climbing from Implementation to Stabilization
- Design rising → climbing from Stabilization to Design
- DebtCleanup rising → expanding team contribution

Teams with high Growth Potential have environments where this climb is possible — role models at each layer. Without them, Growing members spin at Implementation.

![Decline Model](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch2-diagram-decline.png?v=4)

**Teams with a Builder or Cleaner grow people faster.** When a role model exists, Growing members transition to Active at roughly double the rate. **Teams without an Architect degrade over time.**

---

## Member Tiers

Not every git author is a "team member." EIS splits members into three tiers:

| Tier | Condition | Used for |
|---|---|---|
| **Core** | `RecentlyActive && Impact >= 20` | Averages, Density, Consistency |
| **Risk** | Former / Silent / Fragile | RiskRatio, Classification |
| **Peripheral** | Everyone else | Count only |

The header shows `4 core + 3 risk / 16 total`. Drive-by contributors don't dilute metrics. Silent members are detected.

EIS also surfaces **automatic warnings** — bus factor risk, silent accumulation, gravity fragility, top-contributor concentration.

![Team Warnings](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch2-warnings.png?v=4)

---

## Real-World Results

Running `eis team` on our product (12 Backend repos + 9 Frontend repos):

**Backend — Spiral / Legacy-Heavy**:

- 4 core members carrying 12 repos, 3 risk members (2 Silent + 1 Former)
- Architect + 2 Anchors = AAR 0.50 (healthy range)
- `Legacy-Heavy` phase: not declining, but the historical weight is real

**Frontend — Starburst / Mature**:

- 6 core members, 0 risk — everyone is active
- Architect + Anchor present, Risk 0%
- One Gravity warning remains, but structurally healthy

**The numbers tell a story.** Not just "whose signal is strong" but "what state is the team in, and what happens next."

---

## Good Design Creates Common Sense

Our Backend is Legacy-Heavy because a former architect left. Several modules remain that only they had touched.

And yet, the team hasn't collapsed.

Why? Because those modules were built on well-organized design. No comprehensive documentation. No complete knowledge transfer. But **the design embedded in the code's structure gave the remaining engineers enough understanding to operate confidently.**

Strong design leaves knowledge in structure, not in people. A strong team will gradually replace Former members' code with their own, and Legacy-Heavy resolves itself. EIS captures that convergence through Survival trajectories.

---

## Try It

```bash
❯ brew tap machuz/tap && brew install eis
❯ eis team --recursive ~/workspace

# JSON → paste into AI for deeper analysis
❯ eis team --format json --recursive ~/workspace | pbcopy
```

Chapter 1 answers "What kind of engineer is this person?"
Chapter 2 answers "What state is this team in?"

Together: hiring (which Role is missing), team formation (maximize complementarity), 1-on-1s (impact trajectories), risk management (catch deterioration early).

All from git history. No surveys. No additional tooling.

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full.png?v=2)

**GitHub**: [eis](https://github.com/machuz/eis) — CLI tool, formulas, and methodology all open source. `brew tap machuz/tap && brew install eis` to install.


If this was useful: [Sponsor on GitHub](https://github.com/sponsors/machuz)

---

### Series

- [Chapter 0: What If Git History Could Tell You Who Your Strongest Engineers Are?](https://dev.to/machuz/git-archaeology-0-what-if-git-history-could-tell-you-who-your-strongest-engineers-are-5397)
- [Chapter 1: Measuring Engineering Impact from Git History Alone](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)
- **Chapter 2: Beyond Individual Signals: Measuring Team Health from Git History**
- [Chapter 3: Two Paths to Architect: How Engineers Evolve Differently](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga)
- [Chapter 4: Backend Architects Converge: The Sacred Work of Laying Souls to Rest](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d)
- [Chapter 5: Timeline: Signals Don't Lie, and They Capture Hesitation Too](https://dev.to/machuz/git-archaeology-5-timeline-scores-dont-lie-and-they-capture-hesitation-too-1gi5)
- [Chapter 6: Teams Evolve: The Laws of Organization Revealed by Timelines](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4lei)
- [Chapter 7: Observing the Universe of Code](https://dev.to/machuz/git-archaeology-7-observing-the-universe-of-code-1op0)
- [Chapter 8: Engineering Relativity: Why the Same Engineer Gets Different Signals](https://dev.to/machuz/git-archaeology-8-engineering-relativity-why-the-same-engineer-gets-different-scores-5dnl)
- [Chapter 9: Origin: The Big Bang of Code Universes](https://dev.to/machuz/git-archaeology-9-collapse-good-architects-and-black-hole-engineers-1dcn)
- [Chapter 10: Dark Matter: The Invisible Gravity](https://dev.to/machuz/git-archaeology-10-dark-matter-the-invisible-gravity-45ne)
- [Chapter 11: Entropy: The Universe Always Tends Toward Disorder](https://dev.to/machuz/git-archaeology-11-entropy-the-universe-always-tends-toward-disorder-ak9)
- [Chapter 12: Collapse: Good Architects and Black Hole Engineers](https://dev.to/machuz/git-archaeology-12-collapse-good-architects-and-black-hole-engineers-3fed)
- [Chapter 13: Cosmology of Code](https://dev.to/machuz/git-archaeology-13-cosmology-of-code-dci)
- [Chapter 14: Civilization — Why Only Some Codebases Become Civilizations](https://dev.to/machuz/git-archaeology-14-civilization-why-only-some-codebases-become-civilizations-2nl3)
- [Chapter 15: AI Creates Stars, Not Gravity](https://dev.to/machuz/git-archaeology-15-ai-creates-stars-not-gravity-4i05)
- [Final Chapter: The Engineers Who Shape Gravity](https://dev.to/machuz/git-archaeology-16-the-engineers-who-shape-gravity-3fmi)

---

← [Chapter 1: Measuring Engineering Impact](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c) | [Chapter 3: Two Paths to Architect →](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga)

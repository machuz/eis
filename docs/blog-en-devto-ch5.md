---
title: "Git Archaeology #5 — Timeline: Scores Don't Lie, and They Capture Hesitation Too"
published: true
description: "Chapter 5 of Engineering Impact Score. When you line up quarterly snapshots, numbers start telling stories — including the ones people don't talk about."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/cover-ch5.png?v=4
---

*When you line up quarterly snapshots, numbers start telling stories.*

![Timeline with role transitions](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-iconic.png?v=4)

## Previously

In [Chapter 4](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d), I explored the Backend team's Architect concentration and the concept of "laying departed Architects' souls to rest."

But that analysis had a limitation. **It was a single point-in-time snapshot.**

Engineers change. They grow. They hesitate. When team dynamics shift, the way they engage with the codebase shifts too.

**To see that change, you need a timeline.**

---

## `eis timeline` — Gaining the Time Axis

I added a `timeline` command to EIS.

```bash
# Default: last 4 quarters in 3-month spans
❯ eis timeline --recursive ~/workspace

# From 2024, quarterly
❯ eis timeline --span 3m --since 2024-01-01 --recursive ~/workspace

# Half-year spans, full history
❯ eis timeline --span 6m --periods 0 --recursive ~/workspace

# Specific members only
❯ eis timeline --author alice,bob --recursive ~/workspace

# JSON output
❯ eis timeline --format json --recursive ~/workspace
```

The mechanism is simple:

1. Collect all commits once
2. Slice commits by period boundaries (every 3 months)
3. For each period, run `git blame <boundary-commit> -- <file>` to reconstruct blame state at that point in time
4. Run the scoring pipeline for each period

**"Collect once, slice many" strategy.** Commit collection happens once. Only blame runs per period.

This gives you each member's Score, Role, Style, and State lined up quarter by quarter. Change becomes visible.

---

## Real Data: The Frontend Team Timeline

Let me walk through our Frontend team's timeline from 2024-Q3 onward.

I'll focus on three key members.

---

### Y.Y.'s Arc: An Architect Speaks Even After Departure

![Y.Y. Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-engineer-f-timeline.png?v=4)

**In the first half of 2024, Y.Y. was putting up numbers on par with machuz.**

Total above 90. Architect Builder. Production 100, Design 100, Survival 100.

This isn't just "strong." **This person was the codebase's architect, full stop.**

They dipped to Producer in 2024-Q3, but snapped right back to Architect Builder in Q4. That wobble just means "one quarter without design involvement" — and the instant recovery proves the depth of structural understanding underneath.

Starting in 2025, scores begin a steady decline. Architect → Anchor → Fragile → Former.

**This is the trajectory of a departure.**

But notice: even at 2025-Q2, Survival is still 100. The code remains. The design lives on.

The subject of Chapter 4's "laying souls to rest" is exactly this person. And this timeline shows at a glance **just how much there is to lay to rest**.

---

### O.'s Arc: They Were an Architect Builder

![O. Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-engineer-j-timeline.png?v=4)

**2024-Q4: O., Total 91.7. Architect Builder.**

This number is extraordinary. When you consider that machuz (Backend) was at 64.1 in the same quarter, **O. had the highest structural influence across all teams that quarter**.

Design 100. Production 100. Survival 96.

In other words, that quarter's Frontend structure was **built by O.**.

The Role transitions afterward are fascinating:

![Role Transitions](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-data-role-transitions.png?v=4)

After building the structure as an Architect Builder, they shifted to Anchor, briefly returned to Architect, and finally settled into Producer.

This means "the Architect's work is done." The structure was built. Now they produce on top of it.

**A healthy transition.**

---

### R.M.'s Arc: Architect From Day One

![R.M. Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-engineer-i-timeline.png?v=4)

**Architect by their second quarter. Consistently in Architect territory ever since.**

They joined in 2024-Q3, starting as Anchor. By the next quarter, they'd ascended to Architect.

This is what "they've been doing Architect work since day one" looks like in data.

Total: 75.7 → 87.5 → 73.2 → 72.4 → 81.7 → 78.1. **Consistently above 70.**

Design 100 in multiple quarters. This means architectural file changes were happening continuously.

And here's where it gets interesting — **a curious wobble**.

---

### The 2025-Q3 "Hesitation"

![2025-Q3 Hesitation](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-data-hesitation.png?v=4)

In 2025-Q3, they dropped from Architect to Anchor. Style shifted from Builder to Balanced.

Total barely changed (73.2 → 72.4). Production went up (67 → 73). Quality went up (91 → 97).

**Their ability didn't decline. Only their design involvement did.**

Design: 100 → 73. That's what caused the "demotion" to Anchor.

What happened that quarter?

**They clashed with the team.**

Specifically, there was a disagreement over Frontend architecture direction.

R.M. had been driving design decisions since joining. That design philosophy didn't align with existing team members at certain points.

The result: **they deliberately reduced their involvement in design decisions**.

EIS captured this precisely:

- Design: 100 → 73 (fewer commits to architecture files)
- Style: Builder → Balanced (from creating structure to adapting to existing structure)
- Role: Architect → Anchor (from designer to structure maintainer)

**Numbers capture hesitation.**

---

### `--per-repo` Reveals the True Structure of "Hesitation"

The timeline alone tells us "design involvement decreased." But `eis analyze --recursive --per-repo` decomposes the picture to individual repositories — and the structure becomes far more precise.

Here's R.M.'s per-repo commit distribution:

![Per-Repo Commits](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-per-repo-commits.png?v=4)

**The true shape of Q3's "hesitation" emerges.**

In Q3, R.M. poured 274 commits into Repo B — their highest-ever quarter in that repo. Production didn't decline — it increased. But it was **production on top of already-established design**, not work that moved the architecture itself.

That's why Design dropped from 100 to 73. You can ship massive volume in an existing repo, but if you're not moving the structural center, the Design axis won't register it.

---

### The Conversation: "Wait for Me"

Behind this "hesitation" was a conversation.

R.M. had a design vision they'd been carrying — a conviction about how Frontend should be structured. I trusted their design instinct and their technical ability.

But the timing wasn't right yet.

**"Once the engineering org has proven itself and the business side is ready to go all-in on engineering, I'll hand it to you. Wait for me."**

I was certain their design was right. I was certain the whole team would benefit from a structure built around their strengths. But as a startup, the business side wasn't yet fully invested in engineering. The ambition to become a tech company was there — but the timing hadn't arrived. So — please, wait a little longer.

Q3's "hesitation" wasn't just friction. It was also **strategic patience**.

---

### A New Universe

![New Universe](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-data-new-universe.png?v=4)

In 2025-Q4, the new product launched.

I handed it to R.M..

Honestly, whether we could hit the tight schedule was uncertain. But given I's implementation speed and ability, and with me covering the gaps, I judged the odds of success were far from low. I told the business side we'd go with a new engineering structure — and got their buy-in.

Commits to existing repos dropped to single digits. In their place: 1,352 commits to the new repo. **1,352 commits in three months.** The following quarter, another 1,333. That's 2,685 commits in six months — +257,362 / -202,390 lines of hand-written TS/TSX (excluding API auto-generation).

This number is extraordinary. One engineer producing this volume in half a year can't be explained by raw productivity alone.

![Transition](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-data-transition.png?v=4)

Design: 73 → 100. Anchor → Architect.

**On a greenfield, the designer's true nature exploded.**

R.M. was a rare breed — an engineer who could also design. Not just code architecture, but visual design.

Initially, the new product was expected to follow the existing design language. But R.M. wanted to build from scratch.

**Two weeks later**, they came back with a prototype: dark theme, mobile-responsive, beautiful visuals, a side-pane architecture that enabled rich expressiveness — a level of polish that the existing design's incremental extension could never have reached.

Something interesting happened. An excellent external designer was involved in the project. When this designer saw R.M.'s design work, they redefined their own role. Rather than competing on visual design, they **stepped back and focused on what they did best — deep information architecture expertise**.

The result: engineering design skill and information design expertise locked together. The new product became a major success. The team decided to adopt the new codebase as the design reference going forward.

---

### Reinterpreting "Hesitation"

With `--per-repo`, the "hesitation" gains three dimensions:

1. **Surface**: Reduced design involvement due to team friction
2. **Structure**: Heavy production work concentrated in existing repos (274 commits)
3. **Context**: Waiting for the right moment to be entrusted with a new product

Q4 connected everything. A new universe was born, and R.M. created gravity in it.

This is the very phenomenon Chapter 8 calls "Engineering Relativity" — **the same engineer produces different gravity in different universes**. In a mature gravitational field, R.M. was an Anchor. In a brand new universe, they became Architect instantly.

Their ability didn't change. **The universe changed.**

---

### And Then, the Return

![Return](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-data-return.png?v=4)

Design stayed at 100 the following quarter too. The structure-building in web-admin continued.

Through friction, through strategic patience, they found their stage in a new universe. This "step back, then emerge somewhere new" pattern signals **Architect maturity**.

A young Architect who hits friction either retreats or bulldozes. A mature Architect **pulls back to read the team's reaction, then re-engages**. And a good leader **reads the timing and prepares the stage**.

R.M.'s timeline and `--per-repo` record that maturation process — quarter by quarter, repository by repository.

And here's another telling number. R.S., an existing team member, started joining R.M.'s new universe. R.S.'s commits in web-admin jumped from 86 in Q4 to 283 in Q1 — more than tripling. They were adapting to the gravitational field R.M. created, and accelerating. The new structure wasn't just one person's output — it was beginning to function as the team's gravity.

---

## Transitions: Change at a Glance

`eis timeline` auto-detects changes:

![Transitions](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-transitions.png?v=4)

Just lining up Role and Style changes tells you what happened.

Architect → Anchor → Architect → Producer.

**Build → stabilize → build again → done, now produce.**

You can read that the Architect's work is finished just from the transitions.

---

## Y.Y. vs. machuz

Lining up the timelines reveals one more thing:

![Comparison](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-comparison-table.png?v=4)

**The moment Y.Y. exits, machuz's architecture becomes the structural backbone.**

2025-Q3. The quarter Y.Y. dropped to 49.1, machuz hit 93.2 as Architect Builder.

This isn't coincidence.

The structure described in Chapter 4 — "Backend Architects concentrate" — is visible right here. **There was never a period where two Backend Architects existed simultaneously.** While Y.Y. held the Architect Builder position, machuz was still an Anchor. machuz reached Architect only after Y.Y.'s scores declined.

Whether this is a structural consequence of Backend's single design axis (DB schema, API conventions) or simply a matter of growth timing is hard to determine from this sample alone. But at minimum, **it happened as a generational transition** in this team.

The timeline visualizes this transition.

---

## What the Timeline Reveals

Things invisible in a point-in-time snapshot become visible in a timeline:

| Point-in-time Snapshot | Timeline |
|---|---|
| "Strong now" | "When they became strong" |
| "They're an Architect" | "When they became an Architect" |
| Can't read "hesitation" | "Design involvement temporarily decreased" is visible |
| Departure = data loss | The departure trajectory remains |
| Team structure = static | Team structure = dynamic (generational transitions visible) |

**Numbers don't lie. And they capture hesitation too.**

R.M.'s "step back" in 2025-Q3 was probably a conscious decision. But it's preserved as a quarter's worth of data, and only when you line it up against adjacent quarters do you think "ah, that's when it happened."

Y.Y.'s exit and machuz's architectural permeation — without a timeline, you can only say "this is the current structure." With a timeline, you can pinpoint "this generational shift happened in 2025-Q3."

---

## Practical Tips

Some practical uses for timelines:

### 1. Material for 1:1s

```bash
❯ eis timeline --author alice --recursive ~/workspace
```

Pull up a member's individual timeline at the start of a 1:1. "Your Design dropped this quarter. What happened?"

Numbers aren't for attacking. **They're conversation starters.**

### 2. Hire Retrospectives

Check a new member's timeline 3-6 months after joining. If you see a Growing → Active transition, success. If Role/Style are still blank after six months, onboarding needs attention.

### 3. Departure Signal Detection

The Active → Fragile → Former transition pattern tells you the departure trajectory. **In theory, you can intervene at the Active → Fragile stage.**

Y.Y.'s case, however, was not a typical departure pattern. They didn't leave by choice — they were pulled away when an **inter-company investment relationship was dissolved**. This wasn't personal dissatisfaction or motivation loss. It was Fragile → Former driven by external factors.

Yet EIS captured the change accurately. Regardless of the reason, **if codebase involvement drops, the numbers reflect it**. When Fragile appears, you check "why" — whether it's voluntary departure preparation or an external factor, the numbers alone can't tell you. But **the fact that change is occurring is detectable**.

What matters is that it works even in atypical cases. For standard departure signals, you'd likely catch them even earlier.

### 4. Team Timeline for Organizational Evolution

`eis timeline` also auto-generates team-level timelines:

![Team Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/ch5-team-timeline.png?v=4)

Guardian → Balanced → Elite. Declining → Mature. Design Vacuum → Healthy.

**You can see the team becoming healthier over time.**

---

## What This Discovery Means

Chapter 1 built snapshots. Chapter 2 looked at teams. Chapter 3 explored Architect lineages. Chapter 4 explored laying souls to rest.

Chapter 5 **gains the time axis**.

Snapshots show "now." Timelines show "why it's like this now."

Y.Y. built the structure. machuz became Architect on top of it. O. built a structure, then settled into producing. R.M. stepped back, then stepped forward again.

**All of it was preserved in the numbers.**

Cold numbers tell the most compelling stories. That's the essence of the timeline.

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full.png?v=2)

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLI tool, formulas, and methodology all open source. `brew tap machuz/tap && brew install eis` to install.


If this was useful: [❤️ Sponsor on GitHub](https://github.com/sponsors/machuz)

---

### Series

- [Chapter 1: Measuring Engineering Impact from Git History Alone](https://dev.to/machuz/measuring-engineering-impact-from-git-history-alone-f6c)
- [Chapter 2: Beyond Individual Scores: Measuring Team Health from Git History](https://dev.to/machuz/beyond-individual-scores-measuring-team-health-from-git-history-3n9f)
- [Chapter 3: Two Paths to Architect: How Engineers Evolve Differently](https://dev.to/machuz/two-paths-to-architect-how-engineers-evolve-differently-1ga)
- [Chapter 4: Backend Architects Converge: The Sacred Work of Laying Souls to Rest](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d)
- **Chapter 5: Timeline: Scores Don't Lie, and They Capture Hesitation Too**
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

---

← [Chapter 4: Backend Architects Converge](https://dev.to/machuz/backend-architects-converge-the-sacred-work-of-laying-souls-to-rest-m6d) | [Chapter 6: Teams Evolve →](https://dev.to/machuz/git-archaeology-6-teams-evolve-the-laws-of-organization-revealed-by-timelines-4lei)

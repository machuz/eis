---
title: "Git Archaeology #5 — Timeline: Scores Don't Lie, and They Capture Hesitation Too"
published: true
description: "Chapter 5 of Engineering Impact Score. When you line up quarterly snapshots, numbers start telling stories — including the ones people don't talk about."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/cover-ch5.svg
---

*When you line up quarterly snapshots, numbers start telling stories.*

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
eis timeline --recursive ~/workspace

# From 2024, quarterly
eis timeline --span 3m --since 2024-01-01 --recursive ~/workspace

# Half-year spans, full history
eis timeline --span 6m --periods 0 --recursive ~/workspace

# Specific members only
eis timeline --author alice,bob --recursive ~/workspace

# JSON output (for feeding into AI analysis)
eis timeline --format json --recursive ~/workspace
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

### Engineer F's Arc: An Architect Speaks Even After Departure

![Engineer F Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-engineer-f-timeline.svg)

**In the first half of 2024, Engineer F was putting up numbers on par with machuz.**

Total above 90. Architect Builder. Production 100, Design 100, Survival 100.

This isn't just "strong." **This person was the codebase's architect, full stop.**

They dipped to Producer in 2024-Q3, but snapped right back to Architect Builder in Q4. That wobble just means "one quarter without design involvement" — and the instant recovery proves the depth of structural understanding underneath.

Starting in 2025, scores begin a steady decline. Architect → Anchor → Fragile → Former.

**This is the trajectory of a departure.**

But notice: even at 2025-Q2, Survival is still 100. The code remains. The design lives on.

The subject of Chapter 4's "laying souls to rest" is exactly this person. And this timeline shows at a glance **just how much there is to lay to rest**.

---

### Engineer J's Arc: They Were an Architect Builder

![Engineer J Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-engineer-j-timeline.svg)

**2024-Q4: Engineer J, Total 91.7. Architect Builder.**

This number is extraordinary. When you consider that machuz (Backend) was at 64.1 in the same quarter, **Engineer J had the highest structural influence across all teams that quarter**.

Design 100. Production 100. Survival 96.

In other words, that quarter's Frontend structure was **built by Engineer J**.

The Role transitions afterward are fascinating:

```
Architect → Anchor → Architect → Producer → Producer → Producer
```

After building the structure as an Architect Builder, they shifted to Anchor, briefly returned to Architect, and finally settled into Producer.

This means "the Architect's work is done." The structure was built. Now they produce on top of it.

**A healthy transition.**

---

### Engineer I's Arc: Architect From Day One

![Engineer I Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-engineer-i-timeline.svg)

**Architect by their second quarter. Consistently in Architect territory ever since.**

They joined in 2024-Q3, starting as Anchor. By the next quarter, they'd ascended to Architect.

This is what "they've been doing Architect work since day one" looks like in data.

Total: 75.7 → 87.5 → 73.2 → 72.4 → 81.7 → 78.1. **Consistently above 70.**

Design 100 in multiple quarters. This means architectural file changes were happening continuously.

And here's where it gets interesting — **a curious wobble**.

---

### The 2025-Q3 "Hesitation"

```
2025-Q2 (Apr)        73.2    67    91   100     100  Architect    Builder
2025-Q3 (Jul)        72.4    73    97   100      73  Anchor       Balanced     ← here
2025-Q4 (Oct)        81.7   100    68   100     100  Architect    Balanced
```

In 2025-Q3, they dropped from Architect to Anchor. Style shifted from Builder to Balanced.

Total barely changed (73.2 → 72.4). Production went up (67 → 73). Quality went up (91 → 97).

**Their ability didn't decline. Only their design involvement did.**

Design: 100 → 73. That's what caused the "demotion" to Anchor.

What happened that quarter?

**They clashed with the team.**

Specifically, there was a disagreement over Frontend architecture direction.

Engineer I had been driving design decisions since joining. That design philosophy didn't align with existing team members at certain points.

The result: **they deliberately reduced their involvement in design decisions**.

EIS captured this precisely:

- Design: 100 → 73 (fewer commits to architecture files)
- Style: Builder → Balanced (from creating structure to adapting to existing structure)
- Role: Architect → Anchor (from designer to structure maintainer)

**Numbers capture hesitation.**

---

### `--per-repo` Reveals the True Structure of "Hesitation"

The timeline alone tells us "design involvement decreased." But `eis analyze --recursive --per-repo` decomposes the picture to individual repositories — and the structure becomes far more precise.

Here's Engineer I's per-repo commit distribution:

![Per-Repo Commits](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-per-repo-commits.svg)

**The true shape of Q3's "hesitation" emerges.**

In Q3, Engineer I poured 274 commits into Repo B — their highest-ever quarter in that repo. Production didn't decline — it increased. But it was **production on top of already-established design**, not work that moved the architecture itself.

That's why Design dropped from 100 to 73. You can ship massive volume in an existing repo, but if you're not moving the structural center, the Design axis won't register it.

---

### The Conversation: "Wait for Me"

Behind this "hesitation" was a conversation.

Engineer I had a design vision they'd been carrying — a conviction about how Frontend should be structured. I trusted their design instinct and their technical ability.

But the timing wasn't right yet.

**"Once the engineering org has proven itself and the business side is ready to go all-in on engineering, I'll hand it to you. Wait for me."**

I was certain their design was right. I was certain the whole team would benefit from a structure built around their strengths. But as a startup, the business side wasn't yet fully invested in engineering. The ambition to become a tech company was there — but the timing hadn't arrived. So — please, wait a little longer.

Q3's "hesitation" wasn't just friction. It was also **strategic patience**.

---

### A New Universe

```
Quarter      Repo A (existing)   Repo B (existing)   Repo C (new)
2025-Q4             5                   5             1,352          ← here
```

In 2025-Q4, the new product launched.

I handed it to Engineer I.

Commits to existing repos dropped to single digits. In their place: 1,352 commits to the new repo. **1,352 commits in three months.** The following quarter, another 1,333. That's 2,685 commits in six months.

This number is extraordinary. One engineer producing this volume in half a year can't be explained by raw productivity alone.

```
2025-Q3 (Jul)        72.4    73    97   100      73  Anchor       Balanced     ← producing in existing structure
2025-Q4 (Oct)        81.7   100    68   100     100  Architect    Balanced     ← creating new structure
```

Design: 73 → 100. Anchor → Architect.

**On a greenfield, the designer's true nature exploded.**

Engineer I was a rare breed — an engineer who could also design. Not just code architecture, but visual design.

Initially, the new product was expected to follow the existing design language. But Engineer I wanted to build from scratch.

**Two weeks later**, they came back with a prototype: dark theme, mobile-responsive, beautiful visuals, a side-pane architecture that enabled rich expressiveness — a level of polish that the existing design's incremental extension could never have reached.

Something interesting happened. An excellent external designer was involved in the project. When this designer saw Engineer I's design work, they redefined their own role. Rather than competing on visual design, they **stepped back and focused on what they did best — deep information architecture expertise**.

The result: engineering design skill and information design expertise locked together. The new product became a major success. The team decided to adopt the new codebase as the design reference going forward.

---

### Reinterpreting "Hesitation"

With `--per-repo`, the "hesitation" gains three dimensions:

1. **Surface**: Reduced design involvement due to team friction
2. **Structure**: Heavy production work concentrated in existing repos (274 commits)
3. **Context**: Waiting for the right moment to be entrusted with a new product

Q4 connected everything. A new universe was born, and Engineer I created gravity in it.

This is the very phenomenon Chapter 8 calls "Engineering Relativity" — **the same engineer produces different gravity in different universes**. In a mature gravitational field, Engineer I was an Anchor. In a brand new universe, they became Architect instantly.

Their ability didn't change. **The universe changed.**

---

### And Then, the Return

```
2025-Q4 (Oct)        81.7   100    68   100     100  Architect    Balanced
2026-Q1 (Jan)        78.1   100    84    83     100  Anchor       Builder      Active
```

Design stayed at 100 the following quarter too. The structure-building in web-admin continued.

Through friction, through strategic patience, they found their stage in a new universe. This "step back, then emerge somewhere new" pattern signals **Architect maturity**.

A young Architect who hits friction either retreats or bulldozes. A mature Architect **pulls back to read the team's reaction, then re-engages**. And a good leader **reads the timing and prepares the stage**.

Engineer I's timeline and `--per-repo` record that maturation process — quarter by quarter, repository by repository.

---

## Transitions: Change at a Glance

`eis timeline` auto-detects changes:

![Transitions](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-transitions.svg)

Just lining up Role and Style changes tells you what happened.

Architect → Anchor → Architect → Producer.

**Build → stabilize → build again → done, now produce.**

You can read that the Architect's work is finished just from the transitions.

---

## Engineer F vs. machuz

Lining up the timelines reveals one more thing:

![Comparison](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-comparison-table.svg)

**The moment Engineer F exits, machuz's architecture becomes the structural backbone.**

2025-Q3. The quarter Engineer F dropped to 49.1, machuz hit 93.2 as Architect Builder.

This isn't coincidence.

The structure described in Chapter 4 — "Backend Architects concentrate" — is visible right here. **There was never a period where two Backend Architects existed simultaneously.** While Engineer F held the Architect Builder position, machuz was still an Anchor. machuz reached Architect only after Engineer F's scores declined.

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

Engineer I's "step back" in 2025-Q3 was probably a conscious decision. But it's preserved as a quarter's worth of data, and only when you line it up against adjacent quarters do you think "ah, that's when it happened."

Engineer F's exit and machuz's architectural permeation — without a timeline, you can only say "this is the current structure." With a timeline, you can pinpoint "this generational shift happened in 2025-Q3."

---

## Practical Tips

Some practical uses for timelines:

### 1. Material for 1:1s

```bash
eis timeline --author alice --recursive ~/workspace
```

Pull up a member's individual timeline at the start of a 1:1. "Your Design dropped this quarter. What happened?"

Numbers aren't for attacking. **They're conversation starters.**

### 2. Hire Retrospectives

Check a new member's timeline 3-6 months after joining. If you see a Growing → Active transition, success. If Role/Style are still blank after six months, onboarding needs attention.

### 3. Departure Signal Detection

The Active → Fragile → Former transition pattern tells you the departure trajectory. **In theory, you can intervene at the Active → Fragile stage.**

Engineer F's case, however, was not a typical departure pattern. They didn't leave by choice — they were pulled away when an **inter-company investment relationship was dissolved**. This wasn't personal dissatisfaction or motivation loss. It was Fragile → Former driven by external factors.

Yet EIS captured the change accurately. Regardless of the reason, **if codebase involvement drops, the numbers reflect it**. When Fragile appears, you check "why" — whether it's voluntary departure preparation or an external factor, the numbers alone can't tell you. But **the fact that change is occurring is detectable**.

What matters is that it works even in atypical cases. For standard departure signals, you'd likely catch them even earlier.

### 4. Team Timeline for Organizational Evolution

`eis timeline` also auto-generates team-level timelines:

![Team Timeline](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/ch5-team-timeline.svg)

Guardian → Balanced → Elite. Declining → Mature. Design Vacuum → Healthy.

**You can see the team becoming healthier over time.**

---

## What This Discovery Means

Chapter 1 built snapshots. Chapter 2 looked at teams. Chapter 3 explored Architect lineages. Chapter 4 explored laying souls to rest.

Chapter 5 **gains the time axis**.

Snapshots show "now." Timelines show "why it's like this now."

Engineer F built the structure. machuz became Architect on top of it. Engineer J built a structure, then settled into producing. Engineer I stepped back, then stepped forward again.

**All of it was preserved in the numbers.**

Cold numbers tell the most compelling stories. That's the essence of the timeline.

---

**GitHub:** [machuz/engineering-impact-score](https://github.com/machuz/engineering-impact-score) — CLI tool, formulas, and methodology are all open source. Install with `brew tap machuz/tap && brew install eis`.

If this article was useful:

[![Sponsor](https://img.shields.io/badge/Sponsor-❤-ea4aaa?logo=github&style=for-the-badge)](https://github.com/sponsors/machuz)

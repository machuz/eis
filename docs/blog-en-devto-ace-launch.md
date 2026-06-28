---
title: "When AI writes a thousand lines a minute, which code actually holds? (Launching OrbitLens Ace)"
series: "OrbitLens Ace"
published: true
description: "Git keeps two records. Everyone reads activity (commits, PRs, lines); almost nobody reads what survived. Once AI makes code infinite, the surviving layer is the only signal left that means anything. EIS reads it; OrbitLens Ace interprets it into a chronicle, not a scoreboard."
tags: opensource, ai, git, saas
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/cover-ace-launch.png?v=1
---

*AI can write infinite code. Whether it lasts is a different question.*

---

Every git repository holds two records.

One everyone reads: the commit log, PR counts, lines changed. Activity. Who moved a lot.
One almost nobody reads: whether a line of code is still standing six months later, and who its blame finally resolves to. That one is structure.

Management usually watches the first and gets blindsided by the second. "If she leaves, nobody understands that subsystem." That collapse was written in the second record the whole time. We just weren't reading it.

---

## Why surviving code

Activity can be inflated. Split the commits. Let a lockfile add a few thousand lines. Pad the PR count. Busyness is easy to perform.

Survival is not. Whether your code outlives you isn't something you can fake overnight; it takes months to know. A typo fix and an architecture change are both "one PR," but the code still standing half a year later can't lie about which was which.

[EIS](https://github.com/machuz/eis) is an open-source CLI that reads exactly that layer. `git log` and `git blame`, nothing else, observed across 7 axes and 3 topology dimensions, printed as JSON. No external API, no AI tokens. `brew install`, run `eis`. Every formula sits in the [whitepaper](https://github.com/machuz/eis/blob/main/docs/whitepaper.md), so nothing is hidden inside.

---

## AI made activity meaningless

This is the part that bites right now.

The speed of writing code left human hands. Ask a model and a thousand lines arrive in minutes. Commit counts and lines changed no longer measure how much a person moved; they measure how much a tool moved. Activity metrics were always shaky. AI finished them off.

So what's left? What survives. Of the thousand lines AI just produced, the share that gets rewritten next week has a survival of zero. Only the code still standing in six months is doing structural work. When generation becomes infinitely cheap, the scarce thing isn't output. It's what stays.

And survival can't be gamed, not even by AI, because time isn't the only thing that decides it. Whether a piece of code lasts is settled by the ecosystem it lands in: whether the next person builds on it or tears it out. The other developers' hands give the answer. So it isn't an absolute ruler. It shifts by environment. A conservative team might keep mediocre-but-bug-free code untouched; an aggressive one might throw the same code out by next month. Survival is a relative signal, valid only inside that particular context.

But that is exactly why you can't fake it. It isn't self-reported and it isn't volume; it's the accumulated choices of everyone around you. And someone whose code survives across many different contexts is probably someone who can genuinely design. A lucky fit to one environment's habits doesn't explain that.

One more thing. Observation is the opposite of generation. The more the world fills with AI-written code, the more valuable an instrument that just reads the facts already cut into the ground. EIS uses no model. It doesn't infer, so it can't hallucinate. It reads what git already wrote down. In an age of generation, it stays an instrument of observation.

---

## But read it wrong and it breaks

This is where it goes wrong.

The moment you read the survival signal from above and put a number next to a name, it becomes a scoreboard. And a scoreboard always distorts what it measures. People start optimizing the number, and the texture of the work flattens underneath it. Engineering metrics die right here, turned into ranking and pressure.

So the question was never "can you read structure from git." It's "can you read it without dropping into evaluation." The instrument that gathers the signal isn't enough. How you read it decides whether it helps or harms.

That is why EIS and Ace are two instruments, not one.

- **EIS is the telescope. It observes.** It reads git and returns signals. No recommendation, no prediction, no evaluation. It stays with the facts.
- **Ace is the observatory. It reads the observation.** But into a chronicle, not a ranking.

If the instrument that gathers the signal also decides what it means, you can no longer tell where fact ends and opinion begins. Keep the telescope on the facts and let the observatory carry the interpretation. You can trust both because there is a line between them.

---

## What the observatory reads

Read these as the structure git was hiding, not as a feature list.

**Structural Summary. Numbers into language.** "Survival 23" on its own is useless. Weak design, or a legacy rewrite in progress? Ace reads the whole signal field and tells you, in prose, what's standing and what's coming apart.

**Conway's Law check. Org chart vs. reality.** Ace places the engineer topology next to the module topology and shows where they drift. The team that supposedly owns a service isn't the team writing it. Knowledge quietly migrated somewhere your org chart never noticed.

**Collapse risk. Before it's an incident.** Modules that survive only because nobody touches them. Modules orphaned when an owner left. A bus factor of one. The risk was in the git history from the start. Ace surfaces it before it turns into a crisis.

**The Organizational Chronicle. The heart of it.** Not a scoreboard. It doesn't rank engineers. It records what the codebase has lived through: the migrations it survived, the architect who shaped a subsystem and moved on, the module that turned fragile after an ownership change.

A scoreboard teaches people to game it. A chronicle is something a team grows attached to. We're after the second. The score is still there if you want to look closer, but it's a lens you pick up, never the headline.

---

## Two doors, one universe

**The telescope (open source):**

```bash
brew install machuz/tap/eis
cd your-repo
eis analyze .
```

No account, no integration. Point it at any repo and read the signals yourself.

**The observatory (SaaS):** [ace.orbitlens.io](https://ace.orbitlens.io). Connect a GitHub account and Ace runs the observation continuously, then reads it back for you.

EIS, the telescope, stays free and open source forever. No seat limits, no trial clock, no open core with the load-bearing parts behind a wall. We don't want to charge for the idea. We want it to travel, and an idea behind a paywall doesn't get far.

Ace, the observatory, is what carries a price, because running observation continuously and reading it back is real infrastructure. It's priced to be reachable, not to gate. **Free** at $0 (up to five people, public repos unlimited, six months of history). **Pro** at $7/mo (one developer, full history, the complete Gravity Certificate). **Nova** at $39/mo for eight seats, plus $18/seat beyond, for a private org that wants the org-level read. Most engineering-analytics platforms run $15 to $50 per developer per month, often behind "contact sales." Ace is flat for the first eight seats, and the engine beneath it is free. The Gravity Certificate, the trace that travels with a person, is free to mint and verify on every plan. Proving your own work shouldn't cost you anything.

---

## One honest thing

There's a failure mode we sit with. The moment engineering gets measured from above, the way an organization surveys its people, what the people on the ground actually feel tends to flatten. A number arrives, and the texture of what someone did disappears under it.

We're trying to build the opposite, and we're not sure we've managed it. Signals are signals, not verdicts. The things git can't see, we don't measure, and we say so out loud: mentoring, the design argument that never became a commit, the calm someone held during an incident. That's the dark matter the telescope can't reach. A low signal isn't a small contribution. More often it's a signal about the organization, not the person.

Measuring people is dangerous even with good intent, and we haven't found how to do it without ever diminishing the frontline's own sense of its work. So this is a real request. If you read your own signals, or your team's, and something feels off, flattened, unfair, missing the thing that actually mattered, tell us. That gap is where the instrument has to get better.

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png)

**GitHub**: [eis](https://github.com/machuz/eis) — the telescope, fully open source.
**Observatory**: [ace.orbitlens.io](https://ace.orbitlens.io)
**Library**: [library.orbitlens.io](https://library.orbitlens.io) — the theory and the shelf behind EIS (Git Archaeology, whitepaper).

If this was useful: [Sponsor on GitHub](https://github.com/sponsors/machuz)

---
title: "The Telescope Now Has an Observatory — Introducing OrbitLens Ace"
series: "OrbitLens Ace"
published: true
description: "EIS observes git history and prints 7-axis signals. But observation waits for interpretation. OrbitLens Ace is the observatory built on top of the telescope — structural summary, Conway's Law verification, collapse risk, and an organizational chronicle that records rather than scores."
tags: opensource, saas, productivity, git
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/cover-ace-launch.png?v=1
---

*A telescope observes. The observation means something only once it is read.*

---

## What Just Happened

[EIS (Engineering Impact Signal)](https://github.com/machuz/eis) is a telescope. It's an open-source CLI that reads `git log` and `git blame`, observes engineering across 7 axes and 3 topology dimensions, and prints the result as JSON. No external APIs, no AI tokens. Just the strata already recorded in git.

A telescope observes the universe, but raw light by itself doesn't tell you what you're looking at.

Today the telescope gets an observatory: **[OrbitLens Ace](https://ace.orbitlens.io)** is live.

---

## The Boundary: Telescope vs. Observatory

EIS and Ace are not the same instrument, and keeping them separate matters.

**EIS is the telescope.** It observes. It reads what git already recorded and refuses to do more: no recommendation, no prediction, no inference. Point it at a repository and it returns signals — who shaped the structure, whose code survived, where the gravity sits. The methodology is fully open, every formula in the [whitepaper](https://github.com/machuz/eis/blob/main/docs/whitepaper.md). A telescope you can take apart.

**Ace is the observatory.** It interprets. On top of EIS signals, it reads structure into language. A Structural Summary in prose. A verification of whether your team boundaries match your module boundaries (Conway's Law). A map of which modules are quietly fragile, and an organizational chronicle of what the codebase has lived through.

The telescope gathers light. The observatory reads the sky.

The reason to keep them apart is the same reason an observatory doesn't bend its mirrors: observation has to stay clean for interpretation to be trustworthy. If the instrument that gathers the signal also decides what the signal means, you can no longer tell where the observation ends and the opinion begins. EIS stays a telescope so Ace can be honest about being an observatory.

---

## What the Observatory Reads

### Structural Summary — light into language

EIS prints numbers. Numbers without context are dangerous: a low Survival signal might mean weak design, or it might mean someone is mid-rewrite of legacy code. Ace reads the signal field and renders it as prose — not a score, but a description of what's standing and what isn't.

### Conway's Law verification — people against modules

Your org chart says one thing. Your module boundaries say another. Ace places EIS's engineer topology next to its module topology and observes where they agree and where they drift. You see whether the team that owns a service is actually the team writing it, or whether knowledge has quietly migrated somewhere your org chart never noticed.

### Module topology + collapse risk

Some modules are healthy. Some are fragile, surviving not because they're good but because nobody touches them. Ace surfaces which modules sit under high change pressure with no active owner, and which carry a bus factor of one. The risk was always there in the git history. The observatory makes it observable before it becomes a crisis.

### The Organizational Chronicle — the heart of it

This is the part that matters most, so let me be precise about what it is and isn't.

The Chronicle is **not a scoreboard.** It doesn't rank your engineers. It records what your codebase has been through: the migrations it survived, the architect who shaped a subsystem and moved on, the module that turned fragile after an ownership change. It's a record of structural events, written so a team can look back and recognize its own history.

We built it this way on purpose. The most common failure mode of engineering metrics is that they get turned into evaluation, a number next to a name, used to rank and punish. EIS resists that by design — time-decayed survival can't be gamed by busy work — and Ace extends the resistance. The Chronicle observes what the codebase lived through, not how good each person is. The score is demoted to a lens you can pick up when you want to look closer, never the headline.

Observation over evaluation. A scoreboard is something people learn to game; a chronicle is something a team grows attached to.

---

## How to Start

Two doors, same universe.

**The telescope (open source):**

```bash
brew install machuz/tap/eis
cd your-repo
eis analyze .
```

No account, no integration. Point it at any repository and read the signals yourself.

**The observatory (SaaS):**

> [ace.orbitlens.io](https://ace.orbitlens.io)

Connect a GitHub account, and Ace runs the observation continuously, then reads it for you — structural summary, Conway verification, collapse risk, and the chronicle.

---

## The Telescope Is Free. Forever.

EIS, the telescope, is and stays completely free and open source. No seat limits, no trial clock, no "open core" with the load-bearing parts behind a wall. The whole methodology sits in the [whitepaper](https://github.com/machuz/eis/blob/main/docs/whitepaper.md), and the binary is one `brew install` away. An open telescope is the point: anyone can aim it at their own code and read the signals without asking us for anything.

The aim was never to charge for the idea. It's for the idea to spread — that engineering can be observed from git, that survival outlasts activity, that the quiet structural work should stop being invisible. An idea behind a paywall doesn't travel far.

The observatory — Ace, the SaaS — is what carries a price, because running observation continuously and reading it back is real infrastructure. It's priced to be reachable, not to gate:

- **Free** — $0. A solo developer, or a team up to five. Public repos unlimited, a few private, six months of history.
- **Pro** — $7/mo. One developer, full history, the complete Gravity Certificate.
- **Nova** — $39/mo base (eight seats included) + $18/seat beyond — for a private organization that wants the org-level reading: Conway verification, collapse risk, per-member ownership.

For contrast: most engineering-analytics platforms price per developer, commonly in the $15–50 per-developer-per-month range, often gated behind "contact sales." Ace's organization tier is a flat base for the first eight seats, and the engine beneath it is free. We'd rather more teams observe their codebase than fewer teams pay more to.

And the Gravity Certificate, the trace that travels with a person, is always free to mint and verify on any plan. Proving your own work shouldn't cost you anything.

---

## Why It's Built This Way

A telescope alone observes, but observation doesn't change an engineer's working life. You need to read the observation, recognize the structure inside it, and let a team see its own history clearly enough to act on it.

So the arc runs end to end: the telescope observes, the observatory interprets, and the work of people who quietly hold a codebase together stops being invisible.

The light was always there in git. Now there's an instrument that reads it.

---

## A Note on Measuring People — and a Request

There's a failure mode we sit with constantly. The moment engineering gets measured from above — from the altitude of management, an organization surveying its people the way a state surveys citizens — what the people on the ground actually feel tends to flatten. A number arrives, and the texture of what someone did gets lost underneath it.

We're trying to build the opposite of that, and we're not sure we've managed it. Signals are signals, not verdicts. The things git can't see stay unmeasured: mentoring, the design argument that never became a commit, the calm someone holds during an incident. We'd rather say so out loud than pretend otherwise. That's the dark matter the telescope knows it can't reach. A low signal isn't a small contribution; often it's a signal about the organization, not the person.

Measuring people is dangerous even with good intent, and we don't think we've found how to do it without ever diminishing the frontline's own sense of its work. So this is a real request. If you read your own signals, or your team's, and something feels off — flattened, unfair, missing the thing that actually mattered — tell us. That gap is where the instrument has to get better.

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png)

**GitHub**: [eis](https://github.com/machuz/eis) — the telescope, fully open source.
**Observatory**: [ace.orbitlens.io](https://ace.orbitlens.io)
**Library**: [library.orbitlens.io](https://library.orbitlens.io) — the theory and the shelf behind EIS (Git Archaeology, whitepaper).

If this was useful: [Sponsor on GitHub](https://github.com/sponsors/machuz)

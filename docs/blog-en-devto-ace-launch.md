---
title: "The Telescope Now Has an Observatory — Introducing OrbitLens Ace"
series: "OrbitLens Ace"
published: true
description: "EIS observes git history and prints 7-axis signals. But observation waits for interpretation. OrbitLens Ace is the observatory built on top of the telescope — structural summary, Conway's Law verification, collapse risk, and an organizational chronicle that records rather than scores."
tags: opensource, saas, productivity, git
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/cover-ace-launch.png?v=1
---

*A telescope observes. But observation waits to be read.*

---

## What Just Happened

[EIS (Engineering Impact Signal)](https://github.com/machuz/eis) is a telescope — an open-source CLI that reads `git log` and `git blame`, observes engineering across 7 axes and 3 topology dimensions, and prints the result as JSON. No external APIs. No AI tokens. Just the strata already recorded in git.

A telescope observes the universe. But raw light, by itself, doesn't tell you what you're looking at.

Today the telescope gets an observatory: **[OrbitLens Ace](https://ace.orbitlens.io)** is live.

---

## The Boundary: Telescope vs. Observatory

EIS and Ace are not the same instrument, and keeping them separate matters.

**EIS is the telescope.** It observes. It reads what git already recorded and refuses to do more — no recommendation, no prediction, no inference. Point it at a repository and it returns signals: who shaped the structure, whose code survived, where the gravity sits. The methodology is fully open, every formula in the [whitepaper](https://github.com/machuz/eis/blob/main/docs/whitepaper.md). A telescope you can take apart.

**Ace is the observatory.** It interprets. On top of EIS signals, it reads structure into language: a Structural Summary in prose, a verification of whether your team boundaries match your module boundaries (Conway's Law), a map of which modules are quietly fragile, and an organizational chronicle of what the codebase has lived through.

The telescope gathers light. The observatory reads the sky.

The reason to keep them apart is the same reason an observatory doesn't bend its mirrors: **observation has to stay clean for interpretation to be trustworthy.** If the instrument that gathers the signal also decides what the signal means, you can no longer tell where the observation ends and the opinion begins. EIS stays a telescope so Ace can be honest about being an observatory.

---

## What the Observatory Reads

### Structural Summary — light into language

EIS prints numbers. Numbers without context are dangerous — a low Survival signal might mean weak design, or it might mean someone is mid-rewrite of legacy code. Ace reads the signal field and renders it as prose: not a score, but a description of what's standing and what isn't.

### Conway's Law verification — people against modules

Your org chart says one thing. Your module boundaries say another. Ace places EIS's engineer topology next to its module topology and observes where they agree and where they drift. You see whether the team that owns a service is actually the team writing it — or whether knowledge has quietly migrated somewhere your org chart never noticed.

### Module topology + collapse risk

Some modules are healthy. Some are fragile — surviving not because they're good, but because nobody touches them. Ace surfaces which modules sit under high change pressure with no active owner, and which carry a bus factor of one. The risk was always there in the git history. The observatory makes it observable before it becomes a crisis.

### The Organizational Chronicle — the heart of it

This is the part that matters most, and it's worth being precise about what it is and isn't.

The Chronicle is **not a scoreboard.** It doesn't rank your engineers. It records what your codebase has been through — the migrations it survived, the architect who shaped a subsystem and moved on, the module that turned fragile after an ownership change. It's a record of structural events, written so a team can look back and recognize its own history.

We built it this way on purpose. The most common failure mode of engineering metrics is that they get turned into evaluation — a number next to a name, used to rank and punish. EIS resists that by design (time-decayed survival can't be gamed by busy work), and Ace extends the resistance: **the Chronicle observes what the codebase lived through, not how good each person is.** The score is demoted to a lens you can pick up when you want to look closer — never the headline.

Observation over evaluation. A chronicle is something a team grows attached to. A scoreboard is something people learn to game.

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

Pricing comes in three plans — **Free**, **Pro**, and **Nova** — covering a single repository up to a full organization. Pricing may shift as the product settles, so the current numbers live on the [pricing page](https://ace.orbitlens.io) rather than here.

---

## Why It's Built This Way

A telescope alone observes, but observation doesn't change an engineer's working life. You need to read the observation, recognize the structure inside it, and let a team see its own history clearly enough to act.

That's the whole arc: the telescope observes, the observatory interprets, and the work of people who quietly hold a codebase together stops being invisible.

The light was always there in git. Now there's an instrument that reads it.

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png)

**GitHub**: [eis](https://github.com/machuz/eis) — the telescope, fully open source.
**Observatory**: [ace.orbitlens.io](https://ace.orbitlens.io)

If this was useful: [Sponsor on GitHub](https://github.com/sponsors/machuz)

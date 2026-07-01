---
title: "Inside OrbitLens Ace: A Tour of the Observatory"
series: "OrbitLens Ace"
published: true
description: "A screen-by-screen tour of OrbitLens Ace: the observatory dashboard, Star Detail, module topology with collapse risk, the organizational chronicle with its Slack connector, Ambient mode, and the Gravity Certificate. Each screen framed by one question: what becomes observable?"
tags: opensource, ai, git, saas
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/cover-ace-features.png?v=2
---

*Every screen answers one question: what becomes observable here?*

---

[OrbitLens Ace](https://ace.orbitlens.io) is the observatory built on top of EIS, the open-source git telescope. The telescope observes: 7 axes, 3 topology dimensions, printed as JSON. The observatory reads that observation back as structure you can see.

This is a tour. Each screen below is a live galaxy in Ace, with the organization name and private module names redacted. The structure is real; the identifiers are blurred. For each one, the only question is: what becomes observable?

---

## The Observatory Dashboard: 7 axes at a glance

![Observatory dashboard: 7-axis signals (identifiers redacted)](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-observatory.png?v=1)

The dashboard is where the light lands first. Every observed engineer sits on the same 7 axes: Production, Catalysis, Survival, Design, Breadth, Debt Cleanup, Indispensability. The shape of a contribution becomes visible at a glance, not just how much of it there was.

Now that AI writes a thousand lines in minutes, a commit count measures the tool more than the person. So the axes that hold shape rather than volume matter more, not less.

**What becomes observable:** who shaped the structure versus who generated volume. An engineer with 1,890 commits and a Survival near zero looks busy on a commit graph and quiet here. An engineer with 82 commits but Design at the ceiling stands out. The axes hold a difference that a commit count flattens.

---

## Star Detail: the radar, the insight, the structural summary

![Star Detail: radar + structural summary (identifiers redacted)](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-star-detail.png?v=1)

Focus the telescope on a single star and the 7 axes open into a radar. Around it sit the engineer's topology classification (Role, Style, State) and a **Structural Summary** written in prose.

The summary is what to dwell on. Numbers without context invite misreading. A low Survival might mean weak design, or it might mean the engineer is mid-rewrite of legacy. The summary reads the whole signal field at once and describes what's actually standing. Light turned into language.

**What becomes observable:** not "how good is this person," but what kind of trace they left in this particular universe of code. A Cleaner who guards integrity reads differently from a Producer who generates volume, and the radar shows the difference rather than collapsing it into one number.

---

## Module Topology + Collapse Risk: where the structure is breaking

![Module topology: collapse risk and bus factor (identifiers redacted)](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-module-topology.png?v=1)

The telescope observes more than people. It observes the space they inhabit. Every module sits on three axes: Coupling (boundary quality), Vitality (change pressure × survival), and Ownership (knowledge distribution).

Ace reads the dangerous combinations. A module under high change pressure whose code doesn't survive is a structural time bomb. A module that survives only because nobody touches it is a **Fragile** fortress, fine until the day someone has to change it. A module whose owner has left is **Orphaned**, a bus factor that already went to zero.

**What becomes observable:** where the system is breaking, before it becomes a crisis. Not "this engineer is weak" but "this module carries a bus factor of one, and the person who held it is gone."

---

## The Organizational Chronicle: what the codebase lived through

![Organizational chronicle: structural events over time (identifiers redacted)](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-chronicle.png?v=1)

The Chronicle is deliberately **not a scoreboard.** It records structural events along a timeline: a migration the codebase survived, an architect who shaped a subsystem and moved on, a module that turned Fragile after an ownership change.

What it records is what the codebase has been through, not how good each person is. People learn to game a scoreboard. A team grows attached to a chronicle. The signal is still there if you want to look closer, but it's a lens you pick up, never the headline.

**What becomes observable:** the team's own history, written clearly enough to recognize. Observation over evaluation.

### Slack connector: the `:orbitlens_chronicle:` reaction

![Slack connector: chronicle reaction (identifiers redacted)](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-slack-connector.png?v=1)

A chronicle that only Ace writes would be a thinner record than the one a team actually carries. So the Chronicle has a connector: react to a Slack message with `:orbitlens_chronicle:` and that moment is placed onto the timeline. The day a hard migration finally landed. The decision a thread settled. Git records the structure; the connector lets a team annotate the dark matter git can't reach.

### Weekly digest: the codebase, once a week

Once a week, Ace reads the recent observation and places the week's structural events into the chronicle: a module that crossed into Fragile, an owner who went quiet, a survival ratio that shifted. It isn't an activity report. It's a record of what changed in the structure.

**What becomes observable:** the slow tectonics of a codebase, on a cadence a team can actually hold in its head.

---

## Ambient: the codebase, always in view

![Ambient mode: standing display (identifiers redacted)](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-ambient.png?v=1)

Ambient is the observatory left running on a screen, a standing display for a stand-up or an office wall. It keeps the structure quietly in view: the modules under pressure, the recent chronicle entries, the shape of the team's gravity field. It isn't a dashboard you open to investigate. It's a sky you glance at and stay oriented by.

**What becomes observable:** structure as something a team lives alongside, not something it audits once a quarter.

---

## Gravity Certificate: a trace that travels

![Gravity Certificate: portable record of structural impact (identifiers redacted)](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/png/ace-certificate.png?v=1)

An engineer's structural impact lives in a codebase's git history, and stays there when they move on. The Gravity Certificate lets that observation travel with the person: a record of the gravity they held, the modules they owned, the architecture they shaped, observed from git rather than asserted on a résumé.

It's careful by design. It's a record of what the code shows, observed from one universe, not a universal ranking of engineering ability. High gravity in one codebase is a local observation, and the certificate says exactly that and no more.

**What becomes observable:** the quiet structural work that usually stays invisible, the kind that holds a system together and never makes it onto a commit-count leaderboard.

---

## The Whole Arc

The telescope observes, the observatory reads. From the dashboard down to the certificate, the through-line never changes: **observation over evaluation.** Ace shows a team what its codebase has lived through and where its structure is bending, and it demotes the signal to a lens, so the record stays something a team wants to look at rather than something it learns to game.

The telescope (open source) is completely free. It runs locally, no account needed:

```bash
brew install machuz/tap/eis
```

The observatory (SaaS) is a few clicks. Open [ace.orbitlens.io](https://ace.orbitlens.io), log in with GitHub, pick the repos you want to watch, and every screen above renders in your browser.

Point the lens at your own organization, at your personal repos, at an open-source project you're curious about.

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png)

**GitHub**: [eis](https://github.com/machuz/eis) · **Observatory**: [ace.orbitlens.io](https://ace.orbitlens.io) · **Library**: [library.orbitlens.io](https://library.orbitlens.io)

The telescope is free and open source, forever. If this was useful: [Sponsor on GitHub](https://github.com/sponsors/machuz)

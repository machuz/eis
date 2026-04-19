---
title: "Structure-Driven Engineering Organization Theory #1 — The Concept of Observation"
series: "Structure-Driven Engineering Organization Theory"
published: true
description: "Evaluation assigns values. Observation reads structure. Get the entry point wrong and every destination collapses back into evaluation."
tags: management, leadership, engineering, observability
---

> *Part 1 of the **Structure-Driven Engineering Organization Theory** series. New here? → [Start from #0: Why Existing Org Theory Doesn't Work](https://dev.to/machuz/structure-driven-organization-theory-0-why-existing-org-theory-doesnt-work-1101)*

*"She's working hard." "Not really." The same person, and evaluators disagree. As long as we keep calling that "a difference in subjective view," an organization can never be measured.*

*Evaluation is price-setting. Observation is reading. Get the entry point wrong and wherever you arrive, you end up back at evaluation.*

---

> **Scope of this chapter**: thinking layer (defining what "observation" means) + design layer (showing why EIS qualifies as observation).

### How this starts on the floor

A quarterly review meeting. A manager says: "A has grown a lot in the past six months — commit count up, PR count up, more vocal in reviews. Mark them 'exceeds expectations.'" Another manager exhales quietly: "Actually, B has been quietly cleaning up around A. A's PRs go through because B handles the work before and after." A third cuts in: "Numbers are numbers. A gets the rating. We can talk about B somewhere else."

Everyone in the room is saying something true. No one is lying. And yet, what comes out of the meeting is missing something obvious — B's contribution falls through the cracks of the evaluation entirely.

That dropped piece is where this chapter starts. The reason is simple: **what isn't observed can't be discussed.**

---

## Why Start With Observation

As the intro laid out, everything in this book begins with observation. Observation first, language next, structure last, culture as the consequence.

But "observation" is also an everyday word. Because of that, **what the word actually points to drifts at the very first step**. If a reader walks into the rest of the book carrying the assumption "observation just means evaluation, roughly," the entire book will be consumed as another evaluation framework.

To prevent that, this chapter decomposes the concept of observation itself. We separate it from neighboring words, fix its position, and then show why EIS (Engineering Impact Signal) qualifies as *observation*.

## Separating the Neighboring Words

When talking about organizations, people conflate every word that means "measure." Evaluation, measurement, monitoring, observation — they all look like they describe the same act. But **the input and output are different**.

| Word | Input | Output | Subjectivity |
|---|---|---|---|
| **Evaluation** | Facts + criteria + evaluator | Delta against criteria | High (evaluator's subjectivity) |
| **Measurement** | Target + instrument | Numbers | Low (depends on instrument precision) |
| **Monitoring** | Time-series numbers | Threshold alerts | Low (threshold design is subjective) |
| **Observation (human-centric)** | Target + observer | Interpreted facts | Medium (observer's selection bias) |
| **Observation (structural)** | Structure of target | A reading of structure | Low (structure is externalized) |

The last two rows matter here. In English both translate as "observation," but Japanese distinguishes them with separate words (*kansatsu* / *kansoku*). The distinction is worth importing.

- **Human-centric observation** filters facts through the observer's gaze. "She's late a lot" is this kind of observation. Without an observer who notices lateness, the fact doesn't get recorded.
- **Structural observation** reads the **structure** of the target directly. The value is the same whether anyone is watching. "This code has 80% survival 6 months after it was written" is structural observation. With a defined algorithm, any computer produces the same answer.

This distinction is decisive for handling organizations. Decisions made on human-centric observation **change when you change the observer**. Decisions made on structural observation don't.

Conventional management has run almost entirely on human-centric observation. This book places structural observation at the foundation and layers human-centric observation on top.

## How Evaluation and Observation Differ

One more level down. If we miss this, the rest of the book gets absorbed as "a new evaluation metric."

### Evaluation: The Operation of Assigning Value to Facts

Evaluation takes a fact, applies external **criteria** to it, and outputs a value.

- Fact: "A made 120 commits in six months."
- Criterion: "Senior engineers should commit 20+ times per month."
- Evaluation: "A exceeds the criterion."

What matters: **evaluation isn't the fact itself, it's the delta between fact and criterion**. Change the criterion and the same fact produces a different evaluation. Someone always makes the criterion. Whatever someone makes contains their subjectivity.

This doesn't make evaluation bad. Decisions need criteria. The problem is **forgetting where the criteria came from and trusting only the evaluated value**. Once that happens, criteria become gameable.

### Observation: Reading Structure As It Is

Observation reads structure directly without bringing criteria in.

- Observation: "30% of the code A wrote six months ago is still alive right now."
- Observation: "75% of the code B wrote six months ago is still alive right now."

At this point, **no criterion has been introduced**. The statement just describes what is.

Apply a criterion like "above X% is senior" and it becomes evaluation. Don't, and the structural fact itself still carries decision-relevant information. Observation is the **prerequisite data** for evaluation, not evaluation itself.

### Why This Matters in the Field

In an evaluation-only organization, criteria become politics. Who sets them, how they're phrased, how they're gamed — these start outweighing the actual work.

In an organization with observation at the base, facts get shared first. "This module has had 40% of its lines rewritten over the past year." "This person's code survival is 70%." Criteria get discussed on top of that. Criteria can change; observational data doesn't. Criteria become easier to change, and **the argumentative footing doesn't shift**.

## EIS Is Observation

With that framing, here's why EIS counts as *observation*.

EIS takes `git log` and `git blame` as input and computes seven axes:

1. **Production** — Changed lines (daily-scale)
2. **Quality** — First-pass quality (absence of follow-up fixes)
3. **Survival** — Time-decayed code survival rate
4. **Design** — Contribution to architecturally-central files
5. **Breadth** — Span across repositories
6. **Debt Cleanup** — How much of others' debt was swept
7. **Indispensability** — Module ownership rate (bus factor)

All of these are **values computed mechanically from the objective structure of Git history**. Anyone running them gets the same numbers. No "good/bad" judgment is included at this point.

EIS qualifies as observation because three conditions hold:

1. **The input is structure.** Git history is fact engraved in the repository, not filtered through human subjectivity.
2. **The output is reproducible.** Apply the same algorithm to the same repo and anyone gets the same value.
3. **It carries no criterion.** The value *is* the output; "good or bad" is a separate layer.

With these three, EIS runs as observation rather than evaluation. Only after an organization looks at EIS output and applies its own criterion ("in this phase, this axis matters") does it become evaluation.

In other words, **EIS is neutral with respect to criteria**. Different teams, different phases can apply different criteria on top. The observational data doesn't change. This is the practical strength of putting observation at the base.

### The Three Layers of Observation, and Where EIS Sits

Restating the three layers from the intro:

| Layer | What's observed | Example |
|---|---|---|
| **Behavior** | Who did what | Commit frequency, meeting hours |
| **Output** | What was produced | Features shipped, bugs filed |
| **Structure** | What remained, how it's connected | Surviving code, owned modules |

EIS is a device for reading the **Structure layer**. Behavior/Output indicators (commit count, PR count, release count) can be used alongside, but alone they can't see "contribution to structure." The point of using EIS is to hold a **numerically-tractable observational instrument at the bottom of the three-layer stack**.

## The Ethics of Observation

Introducing observation always hits the question: **isn't observation just surveillance?**

This must be handled honestly. Observation *can* tip over into surveillance. These three principles keep it from doing so.

### 1. Choose the Aggregation Unit Correctly

Observation can run at the individual, team, or module level. **The ethical character of the system depends on which level you operate at.**

- **Individual** — Strongest information, but should be used as **material for the person's own career judgment**. Using it as a cross-person ranking tips toward surveillance fast.
- **Team** — Ideal for reading organizational health. No individual identification.
- **Module** — Reads risk in the codebase itself. Not about people.

In this book, individual observational results are treated as **information the person themselves owns**. Managers read it alongside them. Individual raw data should not be shown on an org-wide dashboard.

### 2. Don't Make Observation the Sole Basis for HR Decisions

Observation is **material** for HR decisions, not the **sole basis**. "The observation says X, therefore we promote" is dangerous. Observation supports human judgment; it doesn't replace it.

This echoes the intro's "subjective evaluation vs. structural observation." Structural observation goes first; subjective evaluation layers on top. Either reversed, or either alone, doesn't function.

### 3. Name the Operational Owner

Observational data can't be left in an "anyone can see it" state. Specify who has access, who interprets, and who feeds it back to the person.

- Minimize access rights (the person + their direct manager + an org-health owner)
- Raw data at the individual level, aggregated data anonymized
- Interpretation shared in 1-on-1s, never flashed on a dashboard for a broad audience

Hold these three and observation won't become surveillance. Drop them and even the best observational metric will tip over. **The ethics sit more in how observation is operated than in what's observed**.

### Surveillance or State-Sharing — Governance Is the Watershed

Pressing further: **the same observational data becomes "surveillance" or "state-sharing" depending on the organization's governance**.

A design feature of EIS matters here. EIS has **high game-resistance**. Time-decayed Survival doesn't rise from working frantically — only surviving code counts. The other axes are similarly resistant to short-term behavior or loud voices. To game the metric, **you have to write good structure**. That's the only path.

This game-resistance is what lets the surveillance/state-sharing split be determined by **governance** rather than by the metric.

- **Strongly hierarchical organizations** — A layer with position and evaluative authority uses observational data from below as "management material." However carefully observation is designed, data here tends to act as **an instrument of evaluation and power**. The three principles above exist for this context.
- **Flat, autonomous organizations** — Where positions don't exist (or are light), each person autonomously chooses where to contribute from their strengths, and they push each other to think hard about structure and build it together. Here, observational data functions as **state-sharing for everyone**. "I'm weak on this axis, strong on this one right now" — this becomes material for the individual's own decisions *and* material for the team's conversation.

Observation running in the latter kind of organization doesn't become surveillance. Rather than anyone managing anyone, it becomes **a device for driving the whole team's conversation about structure**. EIS's game-resistance is the prerequisite that sustains this state. Put a low-game-resistance metric into a flat organization and you end up with a "people who game the metric" / "people who don't" asymmetry that eats the flatness.

This isn't an argument about which organizational form is right. The claim is: **whether observation becomes surveillance or state-sharing is decided not by the metric itself but by the governance of the organization it runs in**. The person introducing the observational tool has to make that judgment call about their own organization. The three principles above are safety devices for running observation inside a hierarchy; in a flat organization, the lightness of the operation itself does the ethical work.

## What Organizations Can Stop Carrying Once Observation Is In Place

Put observation at the base and much of what was running as a necessary workaround can be **set down**.

- **Commit-count and PR-count rankings** — Once contribution to structure is visible, you stop having to chase volume
- **Total dependence on managerial interpretive skill** — Interpretation still matters, but **facts that precede interpretation** arrive through a separate channel
- **The "roughly excellent / roughly meh" atmosphere** — Conversations start from observed values, so impression-based talk decreases
- **"Velocity is up, so the team is healthy"** — Looking at the Structure layer often reveals velocity and health not aligning

You don't have to drop all of it at once. The real change is that, to the extent observation enters, **these fall out of the center of the conversation**.

## What Changes in the Field

Five things change immediately when observation enters an organization:

1. **1-on-1 openings shift from "checking an evaluation" to "sharing an observation."** Before the manager brings criteria in, facts are looked at together with the person.
2. **Evaluation discussions change order.** "How should we evaluate this person?" is preceded by "What has been observed about this person?" Disagreements across evaluators collapse by half once observation is shared.
3. **Changing criteria gets cheaper.** With observational data in hand, "what would shift if we changed the criterion?" becomes simulable. The political cost of changing criteria drops.
4. **Hiring requirements reshape.** "Hire a senior" becomes "this axis is weak; we need someone strong in this axis."
5. **Self-understanding crystallizes.** Structural observation also serves as a mirror for the person themselves — "what have I left behind?" The impression-based self-evaluation gets replaced by a structural one.

These won't shift in a week. They start shifting **around the 3-month mark**, once two or three observational cycles have piled up and the conversation has become structural in tone.

## What's Next

This chapter defined observation and showed why EIS qualifies. But the primary object of observation — **code** — is still treated by most organizations as an artifact.

Artifact-view code is measured by the volume present at the moment of writing. This book treats code as **structure**. What happens after the code is written matters more than what exists the moment it's written.

Next chapter: what does it mean to *read code as structure*? We introduce three axes — Design / Robust / Survival — and sketch the concrete Git Archaeology procedure.

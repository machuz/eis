---
title: "Structure-Driven Engineering Organization Theory #3 — A Structural Model of People"
published: true
description: "Ability lives inside, so it can't be seen. Contribution to structure remains outside, so it can be seen. Reframe people not as 'abilities' but as 'types.'"
tags: management, leadership, engineering, career
---

*Ability lives inside, so it can't be seen. Contribution to structure remains outside, so it can be seen. Reframe people not as abilities but as types.*

---

## The Limit of Seeing People as "Ability"

Engineering organizations discuss people through the language of **ability**.

- "Strong design ability"
- "Great problem-solving ability"
- "Lacking communication ability"

Natural everyday phrases, and precisely because they're natural, **what's actually being argued stays vague**. Ability is an internal attribute of the person. What's inside can't be observed from outside. Arguments grounded in the unobservable eventually collapse into "exchanging impressions."

In this book, we suspend the "ability" framing. Instead, we capture people via **types of contribution to structure** (archetypes). Types can be read from past code and Git history. Because they're observable, they can carry an actual conversation.

Worth emphasizing: these types are **not personality classifications**. They aren't MBTI-style "finding the person's true essence." They're observational results for **how this person contributed to structure, in this period, in this organization, on this layer**. Change the organization and the type changes. Change the phase and the type changes.

Types are not fixed labels. They're **observational snapshots**.

## The Three Basic Types

Introduce three basic types. These are defined by combinations of two observational axes (Production and Survival).

### Anchor

**Places long-lived, high-quality code and builds a skeleton that others don't rewrite.**

- Signal: Quality ↑, Survival ↑, Design contribution present
- Field picture:
  - "Who wrote this part?" "X" "Yeah, figures…"
  - In code review, gives feedback that passes the design backbone through
  - Doesn't mind leaving things in docs
  - Not flashy, but anywhere they touch becomes organizational bone

Anchors create **organizational gravity**. As long as they're there, the codebase holds a structural center. When an Anchor leaves, that region loses gravity for years.

Anchors don't make noise. They aren't flashy. They aren't necessarily the type that shines in meetings. So **without observation, they get overlooked** — this is the typical failure mode.

### Producer

**Drives volume and speed forward. Carries short-term momentum.**

- Signal: Production ↑, Survival and Design not constrained
- Field picture:
  - High sprint commit count
  - Most relied upon immediately before launch
  - Good at "at least getting something that works out the door"
  - Doesn't particularly care whether the code is still there in six months — neither do they

Producers create **organizational speed**. There are phases where Producers are needed more than Anchors. A product launch with a fixed market date, for example. In that situation, all Anchors means time disappears into design discussion. Producers driving forward means **working code — the thing design discussions actually need** — gets built first.

Producer isn't a bad type. Without Producers, nothing moves. But **an organization of only Producers has nothing left six months later**.

### Mass

**Produces volume, but nothing remains.**

- Signal: Production ↑, Survival ↓, Quality ↓-leaning
- Field picture:
  - Looks busy every day
  - Commits often, but three months later almost all is rewritten
  - Frequent back-and-forth in review
  - Their own exhaustion and the organizational impact don't feel balanced

Mass is **not a negative type**. Worth emphasizing because this is easily misread.

Mass is often a signal that **the person has been pushed into this type *in this organization***. The same person may function as an Anchor or Producer in a different organization or on a different layer. A long-running Mass state is less about the person's fault and more often about **a placement failure or a role-design failure**.

Mass also serves as an organizational warning signal. When commit volume and team velocity don't align and several people are in Mass state, it may indicate **the organization is structurally built to produce Mass**.

## Derived Types

Around the three basic types, several derivatives show up often.

### Cleaner

**Protects structure by cleaning others' debt.**

- Signal: Debt Cleanup ↑, Quality ↑, Survival ↑
- Field picture:
  - Takes on others' bug fixes as their own tasks
  - Voluntarily ships refactoring PRs
  - "I cleaned this part up" is a recurring phrase
  - Few new features, but after their hand passes through, the codebase feels lighter

Cleaners are **the device that lowers organizational entropy**. Organizations with and without a Cleaner diverge visibly in codebase quality over equal time.

Cleaners are hard to see in artifact view — typical. No new features, no line-count increase, unflashy tickets. But organizations that lose their Cleaner **rot visibly within six months**. This can only be named once observation is in place.

### Specialist

**Deep in one area, doesn't spread to others.**

- Signal: Survival ↑ and Breadth ↓
- Field picture:
  - No one knows their area better
  - Limited interest in other areas
  - The quality of their area is overwhelming
  - Poor fit for rotation or outreach

Specialists are Anchors of a specific area. Not broad-shallow, but narrow-deep. This produces value in **technically complex domains** — database schemas, security, payments infrastructure, infrastructure automation. Without a Specialist, technical debt in those domains snowballs.

Caution: **Specialists are easily undervalued on low Breadth**. Don't short-circuit "low cross-repo span = low contribution." Narrow is fine, if deep.

### Architect (Inheritance-type, Preliminary)

**Someone whose design gets inherited and extended by others' code.**

- Signal: Design ↑, Robust Survival ↑, Breadth present
- What "preliminary" means here:
  - Anchor = "the code they wrote remains"
  - Architect = "**their design direction propagates into code others write**"

An Anchor's own lines remain. An Architect's design permeates lines they never wrote. This is the state where **design has become a shared language**.

Architect is deeply tied to chapter 4's layer structure and chapter 7's culture discussion. Treating it as a standalone archetype has less precision than placing it inside the layer + culture context. Here it's positioned as "the phase beyond Anchor."

## The Structure of "Talented But Spinning"

Organizations often have **people whose talent is obvious, but who are somehow spinning in place**. Neither the person nor those around them can articulate what's happening.

Read structurally, patterns emerge.

### Spread — Touching Widely and Shallowly, Leaving Nothing

- Signal: Production ↑, Breadth ↑, Survival ↓, Design ↓
- Reading: **Producing volume, spanning places. Yet nothing remains, and the center hasn't been touched.**

Spread-type people look serious and active. They show up to meetings, touch many repos, get involved across teams. But six months later, when you count code, their trace is thin.

This state **isn't a lack of ability**. More often, it's a **layer mismatch**. Someone in Spread state is placed on the implementation layer when they should be fighting on the intermediate layer. Or given a cross-cutting role when they should be going deep in a specific domain.

The intervention isn't **making them try harder**. It's changing placement.

#### Some People Play Spread Strategically

One more distinction worth drawing. Some engineers who *look* like Spread on the signals are **doing it strategically**.

Complex requirements, or a product vision that resists verbalization, sometimes **can't be communicated to a team through conversation alone**. The hands stop moving before the design intent lands. Certain seniors, in these situations, write **a rough, working skeleton across the relevant area themselves**. They don't chase quality in the details — they **hand quality off to the engineers who come next**.

This is neither Mass nor Churn. The structure of the act is different:

- The person isn't aiming for "my code survives"
- What they write functions as **a strike surface for the team's real discussion**
- Value only completes when subsequent engineers (Anchors, Cleaners) layer the real skeleton on top

On observational signals, this looks like Spread (Production ↑, Breadth ↑, Survival ↓, Design ↓). The tell is: **is the person handing off to the people behind them**, and **are Survival / Design rising in those same modules afterward?** If the team's Robust Survival is climbing in the regions they touched, what you're looking at is **intended Spread** — team-accelerating scaffolding.

This kind of contributor is hard to identify on observation alone. Combine observation with direct conversation and with how Survival evolves around them. Mistakenly intervening as "Spread → reassign" risks **destroying the organization's drive engine**.

### Fragile — Only Survival Is High, Without Change Pressure

- Signal: Survival ↑, Production ↓, Quality < 70
- Reading: **It remains, but only because nothing's been touched. Not surviving from quality — surviving from nobody daring to touch it.**

Fragile easily reads as Anchor in the data because Survival is high. But what's actually happening is **frozen code that's wanted-to-be-touched-but-can't-be**.

Watch out for this type. An organization full of Fragile is **carrying unmodifiable debt**. The writer of the debt, without intending to "leave" anything, gets treated as an Anchor because their code "just happened to survive."

With observation, Survival can be split into **Raw Survival** and **Robust Survival**. Fragile shows high Raw Survival but low Robust Survival (time-decayed). This separates Anchor from Fragile.

### Churn — High Volume, Low Survival, Constant Rewrites

- Signal: Production ↑, Survival ↓, Quality ↓, Prod–Surv gap ≥ 30
- Reading: **What they write gets rewritten by others immediately, or they rewrite themselves.**

Churn looks like Mass but differs. Mass is "producing volume that doesn't remain." Churn is "**every write necessitates further rewrites**." A Churn writer often unintentionally increases others' workload.

Intervention on Churn must be careful. The person is working hard. But their effect on the codebase may be **purely negative**.

Pairing and apprenticeship under an Anchor work here. Churn is often **a type that changes with training**.

## Types Are Not Fixed

What to emphasize repeatedly in this chapter: **these types are not the person's essence**.

The same person:

- Was an Anchor in Organization A
- Became a Producer in Organization B
- Fell into Spread in Organization C

The person didn't change. **The layer structure and placement of the organization differed**.

In Organization A, they were placed on a layer matching their strengths (intermediate-layer centers, implementation-layer depth). In Organization B, matching a speed-needed phase, they took implementation-layer Producer work. In Organization C, they were given a cross-cutting role and collapsed into Spread.

Types are **observational results of person-organization interaction**. So the target of intervention isn't the person — it's the connection between person and organization. This is what chapter 4 onward develops as **layer structure and placement design**.

## What Changes in the Field

Capture people as types, and organizational conversations shift.

1. **Hiring plans concretize.** "Hire a senior" becomes "we're short on intermediate-layer Anchors" / "we want implementation-layer Cleaners."
2. **Evaluation's subject shifts.** "Who is better" becomes "do we have the types needed for this phase?"
3. **Placement changes become discussable.** "This person is in Spread state; moving them to a different layer might unlock them" becomes a real sentence.
4. **Reading attrition risk shifts.** An Anchor leaving means that region loses gravity for years. Invisible in volume metrics, but speakable through structural observation.
5. **Self-understanding becomes structural.** "I'm Producer-leaning right now; mid-term I want to shift toward Anchor" becomes thinkable. Career becomes self-manageable.

The fifth is the **benefit to the person themselves** that this book offers. Conventional evaluation was a "manager's impression" black box from the person's point of view. With structural observation, they can read their own career. This is the real substance of empowerment.

## What's Next

The language for seeing people as types is now in place. But types alone don't drive organizations. **Types × layers** is what functions.

An Anchor on the implementation layer and an Anchor on the intermediate layer have different organizational impact. A Producer placed on the principle layer spins out.

Next chapter: defining the layer structure of organizations. Implementation / Intermediate / Principle, and the **translator** role that connects them. From here, this chapter's type theory connects to real-world practice as **placement design**.

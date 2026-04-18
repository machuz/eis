---
title: "Structure-Driven Engineering Organization Theory #7 — Making Culture"
published: true
description: "The moment you declare 'our values are X, Y, Z' on a slide, the culture is already dead. Culture isn't what you define — it's the vocabulary that gets used in everyday conversation. Making transformation evaluable is how culture takes hold."
tags: management, leadership, culture, engineering
---

*The moment you declare "our values are X, Y, Z" on a slide, the culture is already dead. The words pinned to the wall are used exactly nowhere outside that meeting.*

*Culture isn't something you define. It's the vocabulary that gets used in everyday conversation.*

---

> **Scope of this chapter**: thinking layer (redefining culture from "shared values" to "shared language") + design layer (making transformation itself an evaluation target, and using culture to distinguish the Structure layer from the Middle layer).

### How this starts on the floor

**Scene A — the Values all-hands**

At an all-hands, the CEO flips to a slide. "Our values: Ownership, Excellence, Ship it." Applause. The next week, a different topic: "code quality feels like it's been slipping." "But speed also matters." "Is that an Ownership issue?" "More like Excellence, I think?" "Or maybe we should prioritize Ship it?" — the discussion scatters. The values are touched as **slogans at the top of the agenda**, never as **material for the decision.**

**Scene B — a team where structural vocabulary flies**

A different team, reviewing a feature. One engineer: "Does this change contribute to the Design layer?" "No — utility territory. Not Robust either." "Then refactor the Design layer first, then pile this on — Survival will be higher that way." "Agreed. Breadth is already too wide. And we should decide, Anchor-style, who guards the design." The conversation isn't mushy. **The vocabulary has settled in.**

---

![Without Culture vs With Culture — does the meeting converge?](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/sdo-ch7-culture-contrast-en.svg)

The same topic — "code quality" — scatters into feelings in Scene A and converges along structural axes in Scene B. The difference? **The available vocabulary is different.**

Scene A only has the **wall vocabulary** (Ownership / Excellence / Ship it) — no vocabulary that **descends into day-to-day judgment.** Scene B has vocabulary that descends (Design / Survival / Robust / Breadth / Anchor). That's what this book calls "culture."

## Culture isn't a definition — it's how the vocabulary gets used

Values, missions, slogans, credos — these are **the result of culture**, or its shadow, not culture itself. Pin them on the wall, broadcast them widely; if they never reach the floor's conversations, they don't exist.

What culture actually is: **the vocabulary used in everyday conversation in this organization, and how it's used.**

This definition is measurable in one way: **can a new hire use this vocabulary naturally within a week?** A week in, they can write "contributes to the Design layer" in a code review, say "I want to talk about the accumulation layer today" in a 1-on-1, ask "which layer is that?" in a meeting — if they're there, the culture has taken hold.

The opposite: a year passes, the wall values never get cited in a meeting, new hires never use those words. That's a sign the culture **doesn't exist.**

## "Which layer is that?" culture

Teams where the vocabulary has settled have one telltale sentence: **"Which layer is that?"**

At the start of a meeting, when discussion starts to mix, someone inserts: "Is this a Principle-layer discussion? Structure? Implementation?" "Is that a behavior-layer problem, or an accumulation-layer problem?" A mixed discussion that had been rolling forward gets **linguistically stopped** — that's the signature of a working culture.

One person who can stop the discussion is enough. Everyone doesn't have to be able to. That one person will stop discussions this way for several months, and the rest of the team starts asking the same question naturally. **Language propagates.**

But if the only person who can stop is a single person, the culture vanishes the moment they leave. **Designing for the number of people who can stop the discussion** is designing for culture's durability.

---

> **Make it your own — a question**
>
> In the last week, in your team —
>
> - How many times was the question **"which layer is that?"** asked in meetings?
> - Did a new hire use structural vocabulary naturally **within a week**? Or have they been there three months and still never cite the wall values?
> - How many people in the team can **stop the discussion** that way?

## Making transformation evaluable

Traditional evaluation looks only at **output and behavior**. Code volume, feature releases, meeting attendance, 1-on-1 frequency — these are easy to observe.

**Transformation is invisible to traditional evaluation.** A transformer's work gets recorded thinly as "ran a meeting," "wrote a doc," "reviewed code." It never climbs the evaluation ladder. The transformer burns out. They leave. The organization loses the transformation without noticing.

To make transformation a target of evaluation as part of culture, you need **vocabulary for observing the signals of transformation.**

![Making transformation evaluable — two axes](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/sdo-ch7-evaluation-axes-en.svg)

### Transformer signals

- **Whether reviews on others' commits/PRs produce structural improvements** (Design / Survival rising after the review lands)
- **How often inter-layer documents (RFCs, design principles, decision records) get updated**
- **How many times per meeting "linguistic realignment" happens** — stopping mixed discussion with "which layer is that?" and bringing it back in
- **The rate of organizational slowdown during a transformer's absence** (the cost of their unavailability, visible ex post)

These aren't fully quantifiable, but they raise the resolution by several orders over "fuzzy contribution." At minimum, the evaluation vocabulary shifts from "ran a meeting" to "fixed the Principle-layer decision into a Structure-layer form how many times."

### Transformation coaches need a different evaluation axis

The **transformation coach** introduced in chapter 4 (Bill Campbell) needs to be evaluated separately from transformers.

- Transformer: runs daily transformations. Signals are relatively observable.
- Transformation coach: extends *others'* transformation capability. **EIS's seven axes can't capture this.**

A transformation coach's contribution can only be measured in **the changes in the people they worked with.** "After six months with this leader, the leader can now carry the Principle ↔ Structure transformation on their own." This kind of observation doesn't show up in structural signals. Only human judgment can read it.

That's exactly why culture has to **explicitly make a place to evaluate the work of transformation coaches.** Otherwise, "unobservable = unevaluated," and the role quietly disappears from the organization.

## Reading people by backing, not by title

Culture has one more function — it **distinguishes people in the Structure layer from people in the Middle layer.**

At some point in the industry, the title "VPoE" became trendy for engineering organizations. The title spread **before** the role's actual substance was understood. Organizations ended up with "VPoEs" who weren't doing what a VPoE is supposed to do. The tragedy repeated across many companies.

**We must not repeat that with the words "Structure layer" or "transformer."** The more the structure-driven vocabulary spreads, the more people will claim those titles. This is exactly why culture matters here.

![Structure layer vs Middle layer — read by backing, not by title](https://raw.githubusercontent.com/machuz/eis/main/docs/images/blog/sdo-ch7-three-middle-moves-en.svg)

### Separating "Structure layer" from "Middle layer"

This book uses the two terms with **deliberately different meanings**:

- **Structure layer** — people who carry substance. Their behavior is **backed** by one of the three types below.
- **Middle layer** — people who hold only the title, with no substantive backing. The target that a structure-driven culture excludes.

In chapter 4, we retired "Middle layer" as a frame for the middle band of an organization (renamed to "Structure layer"). Here we **bring it back with a different meaning** — the pejorative kind. Someone who sits at a middle title but doesn't move. The static connotation of "layer" fits this usage precisely.

### Three types of Structure-layer work

Being in the Structure layer — that is, having real backing — splits into **three kinds**:

1. **Code-backed** — visible in EIS. Comes with Design ↑, Survival ↑, Cleanup ↑. The "hands-on-code while guarding structure" pattern. Anchor, Cleaner, and Inheritance Architect sit here.
2. **Character / trust-based** — the Bill Campbell type, the **transformation coach**. Not visible in EIS, but discernible from **the changes in the people they worked with**. External coaches, senior mentors, product-side trust bridges.
3. **Non-code contribution** — translation, decision bridging, articulating the Principle layer. Observable through **document update history and meeting records**. The Principle ↔ Structure transformer lives here.

When one of these three backs someone's behavior, they're **in the Structure layer**. In an organization with culture, which of the three applies surfaces naturally in conversation — "A is type 1," "B only functions as type 2," "C is carrying type 3."

### Middle layer — excluded by culture

And **behavior with none of the three backings** is just a title being claimed. This book calls it the **Middle layer** — someone in a middle-ish title position, with none of the Structure-layer substance.

In a structure-driven culture, this surfaces on its own — "claims to be a VPoE, but we can't find an observation that fits any of the three." The discrimination becomes possible in day-to-day conversation.

**How long someone can sit in the Middle layer on the strength of a title alone is a measure of how shallow the culture is.** Deep culture makes title-only occupancy unsustainable. That's the mechanism by which culture protects the organization's health.

## The lifespan of language

The culture's vocabulary isn't fixed. **When reality changes, the vocabulary needs updating.**

- New realities need **new words** (e.g., transformation coach, accumulation layer)
- Old words get **deprecated** (e.g., "intermediate layer" → split into "structure layer" + "transformation")
- Unused words get **removed** (a value nobody uses, still pinned to the wall, is a sign of rot)

Vocabulary management uses the same discipline as **software refactoring**:

- Quarterly, inventory the vocabulary the organization actually uses
- Add newly settled words to the list
- Deprecate or remove unused words
- For words whose usage is drifting, pin down the correct usage in one place

"We used to call it that" stops being an obstacle when the language itself is managed. **Intentionally maintaining the lifespan of vocabulary** is a condition of a healthy culture.

## The self-correcting property of culture

Even a settled culture isn't permanently correct. Vocabulary containing mistakes can settle too. **The real strength of a culture is whether it can self-correct when it's wrong.**

The indicator is simple: **is "which layer is that?" asked from inside the team?**

- Executives get asked "isn't that principle out of date?" **from below**
- A value from three years ago gets pushed back on by new hires: "Is this actually being used on the ground?"
- A term's definition gets flagged by the team as no longer matching the reality

These questions surfacing **naturally, from inside** = the culture's self-correction is working. This becomes a core indicator of chapter 8's "conditions of a structure-driven organization."

Conversely, a culture that changes **only when executives or outside consultants say so** isn't really a culture — it's compliance with whichever order is current. The moment the orders stop, the confusion returns.

---

## What Changes in the Field

Designing culture as "shared vocabulary" changes the following:

1. **Onboarding shifts from "preach values" to "hand over vocabulary."** New hires get the EIS seven axes, the three layers, Role × Style × State, transformation, behavior, output, accumulation, transformation coach — in their first week. A week later, observe whether they can participate in conversations using that vocabulary.
2. **Culture work shifts from "events and slogans" to "adding and removing vocabulary."** Offsites, credos, value cards don't *make* culture — they visualize its result. Actual culture work is **maintaining the list of daily vocabulary** and **running the practice of deprecating unused words.**
3. **"Structure layer" vs "Middle layer" becomes discernible in conversation.** With the three-types distinction (code-backed / character-backed / non-code-backed), someone in the **Middle layer** (title-only, no backing) surfaces naturally in daily conversation. No special HR process needed — the culture self-cleans.
4. **Transformers and transformation coaches finally get an evaluation axis.** Evaluation beyond "they ran some meetings" becomes possible. Transformation signals get observed; the transformation coach's "changes in the people they worked with" gets judged by human eyes.
5. **Culture becomes self-correcting.** Culture updates from the floor's "isn't that out of date?" instead of from executive fiat. That's the core of **self-correcting property** — taken up in the next chapter.

## What's Next

Up to here, we've assembled the language for observing an organization, describing its structure, designing interventions on it, and sharing that language as culture. One question remains: **what are the conditions under which all of this holds together as an organization?**

The next chapter sums up the **conditions for a structure-driven organization**. **Reproducibility** (interventions don't depend on one person and can be repeated), **observability** (what's happening stays visible), **self-correction** (when it's wrong, the inside fixes it). Only when all three are present does the organization function as an OS.

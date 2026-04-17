---
title: "Structure-Driven Engineering Organization Theory #4 — The Layer Structure of Organizations"
published: true
description: "Org charts show reporting lines but hide abstraction levels. Three layers — Implementation / Intermediate / Principle — and the translators who keep them connected."
tags: management, leadership, engineering, architecture
---

*People in the same room, speaking the same words, talking past each other. Usually because they're on different layers, but treating it as one conversation.*

---

> **Scope of this chapter**: design layer (describing organizations in three layers) + field layer (how to read symptoms that show up on the floor).

### How this starts on the floor

A quarterly strategy meeting. The CEO says: "Next quarter, we bet on user experience. That's where we differentiate." The room nods. No one objects.

An hour after the meeting, conversation starts in the engineering Slack: "User experience — concretely, what are we building?" "What about the core refactor?" "Priorities keep shifting, can't keep up." The CEO communicated a clear strategy. The floor experiences "no instructions came down."

The CEO didn't lie or slack. The engineers aren't lazy. **They were talking at different abstraction levels, and there was no one between them to translate.**

This chapter redescribes the organization as **three layers of differing abstraction**.

---

## An Organization Isn't a Single Plane

Draw an org chart. Almost always, it ends up as a **tree**. CEO over directors, directors over managers, managers over members. Hierarchy.

Tree org charts carry essential information: who reports to whom, scope of responsibility, where evaluation authority sits. You can't run an organization without it.

But tree org charts are **fatally missing one piece of information** — **abstraction level**.

The CEO and the tech lead are probably in the same meeting room, talking about the same product. But when the CEO is asking "how do we change the world with this product?", the tech lead is asking "how do we migrate this repo's schema?". **Same language, different layer.**

When the participants don't notice the layer mismatch, the conversation spins. The CEO's question gets answered with "well, there's a problem with the migration script…", and the tech lead's question gets answered with "we need a grander vision." Both responses are valid in their own layer; neither lands.

The tree org chart can't display this kind of misalignment. Because it can't display it, no one addresses it. This book reframes the organization here — **not as a tree, but as layers**.

## Three Layers

An organization runs on three layers of differing abstraction.

### Implementation

The layer of **what actually moves**.

- What to build, how to build it, how to fix it, how to operate it
- Code, features, bug fixes, deploys, on-call, customer support
- Maximum concreteness. **Hands move; results follow.**

People in this layer ultimately judge "does it work or not." Beautiful design that doesn't work is worthless; ugly code that works has value — that's the working axis here.

### Intermediate

The layer that **converses with both abstract and concrete**.

- How to structure, how to connect, how to make it withstand change
- Architecture, product design, operational design, security, data modeling
- Can't be done with abstract alone or concrete alone. **Holding both vocabularies is the requirement.**

The intermediate layer translates the "why" descending from Principle into "how" the Implementation layer can act on, and translates "doesn't work" / "breaks" rising from Implementation into "constraints" the Principle layer can judge against. **Translation's main battlefield is here.**

There's a third function the Intermediate layer carries: **keeping the fire alive** — translating between layers so that decisions made in one layer don't extinguish the morale of another. Covered in the "The Intermediate Layer Keeps the Fire Alive" section below.

### Principle

The layer of **strategy, vision, worldview**.

- Why we're building it, what we're not building, who we're building for, where we're betting
- The product's worldview, the root of the value proposition, the source of competitive edge, the organization's mission
- Concreteness is lowest. But **it governs everything else**.

People in this layer ask "is this worth building" first. Even if implementation is possible, no value → no go. Even if implementation is hard, real value → go. That's the working axis.

> These three layers are **not a fixed authority hierarchy**. Not vertical — **abstraction level**. The Principle layer isn't superior or distant. It just **uses different grammar**.

CEO isn't necessarily on the Principle layer. CTO isn't necessarily on the Intermediate layer. **Title and layer are different things.** Title is determined by tree position. Layer is determined by what abstraction level the person actually decides at, day to day.

## The Role of Translators

When layers differ, people speak **different languages**. The Principle layer's vocabulary (market, value, worldview) and the Implementation layer's vocabulary (modules, dependencies, latency) have different grammars and different referents.

Pushing a message down or up untranslated doesn't work:

- **Principle's "worldview" pushed directly to Implementation** → arrives as undecidability of specifications. "OK, but which column do I add?"
- **Implementation's "operational concern" pushed directly to Principle** → gets dismissed as tunnel vision. "That's a small detail; look at the bigger picture."

It's not bad faith. **Translation is just missing.**

### What a Translator Is

A translator is someone who **moves between both vocabularies and converts grammar while preserving the decision**.

- Can decompose Principle's intent down to a granularity Implementation can act on
- Can aggregate Implementation's constraints up to a granularity Principle can judge with
- Both sides feel "the essence wasn't lost"

This is similar to translating between English and Japanese. Good translation isn't mechanical word-substitution — it **reconstructs the speaker's intent in a different language**. Bad translation has correct word correspondences and missing intent.

Typical translators in organizations:

- Strong **tech leads** (Intermediate ↔ Implementation)
- Strong **product managers** (Intermediate ↔ Principle, sometimes Implementation too)
- Strong **engineering directors** (Intermediate as a whole ↔ Principle)
- Strong **Architects** (live in Intermediate, translate to both Implementation and Principle)

> Without translators, large organizations don't function. In organizations without translators, Principle's intent doesn't reach the floor and the floor's voice doesn't reach the executives. **Absence of translation isn't a communication shortage — it's a structural gap.**

### The Intermediate Layer Keeps the Fire Alive

The role of translation isn't only converting grammar. There's a second critical function: **keeping the fire alive.**

By "fire" we mean the floor's morale, momentum, trust in their own work, conviction in what they're building — the **internal energy that drives doing the work.** An organization's productivity ultimately rests on the total amount of this fire.

#### Decisions made in isolation tend to extinguish fire in other layers

When Principle decides in isolation — strategic pivot, roadmap reshuffle, priority change — every decision is rational *within Principle*. But when those decisions reach Implementation untranslated, they land as: "**six months of my work is being thrown away**" / "**my judgment wasn't needed by the organization**." The fire goes out.

When Implementation raises a voice in isolation — "this design will break," "this deadline is impossible" — the concerns are legitimate *within Implementation*. But when they reach Principle untranslated, they get processed as: "**engineers are complaining**" / "**more negativity**." The voice's own fire goes out, and the next person to raise a concern stops bothering.

Each layer's isolated decision is correct from inside that layer. **What's wrong is the absence of translation on the path between layers.** Without translation, a correct decision in one layer arrives in another layer in a form that puts out the fire.

#### Intermediate sits on the propagation path of fire

People in the Intermediate layer can **predict in advance** how Principle's decision will land in Implementation. So instead of pushing it down as-is, they deliver it in a form that doesn't extinguish fire:

- Instead of "six months of my work is being thrown away," they deliver: "What you built (X) remains as the foundation for the new direction (Y). Y stands on top of X."
- Instead of "more negativity," they elevate: "The Implementation-side concern can be reconstructed as decision input Z for Principle."

This isn't "messaging" or "softer communication." It's **redesigning the decision itself so meaning isn't lost between layers.**

#### Fire can be amplified, not just preserved

Good translation doesn't only **prevent** fire from going out — it can **amplify** it.

- Convert small Implementation wins (bug fix, refactor, test additions) into a form Principle can read: "This accumulated cleanup will speed up next-quarter feature delivery by 30%."
- Translate Principle's vision into a granularity Implementation hands can engage with: "The first line of code on this screen connects to the moment a user first understands their own organization."

Carrying the felt sense that the work matters — that act itself amplifies fire.

> **The strongest Intermediate-layer figure — Bill Campbell (*Trillion Dollar Coach*)**
>
> Bill Campbell, who coached the leaders of Apple, Google, and Intuit for decades, **wrote no code and set no strategy. He was an Intermediate-layer figure specialized purely in amplifying the fire of the people around him.** He shaped what leaders said, resolved conflicts between them, built bridges of trust — from that single function, the resulting value was reasonably called "a trillion dollars."
>
> His title was "external coach." **An Intermediate-layer figure who never appeared on the org chart**, producing outsized value through a single function — fire amplification — exactly the kind of person this book's frame can name and value as the strongest exemplar of the Intermediate layer.

#### The real cost of layer mismatch

In organizations with a hollow Intermediate layer, each layer is making locally-correct decisions, yet **the fire of the organization as a whole gradually goes down.** "Everyone is working in good faith, yet somehow we don't move forward" — what most organizations call this state is, almost certainly, this.

> Layers have no hierarchy. But **they can extinguish the fire of other layers, or amplify it.** The Intermediate layer sits on that propagation path.

### Translators Are Hard to Evaluate

Translator work doesn't directly produce visible artifacts. They don't write code, they don't set strategy — they **connect the two**.

In artifact view, translator contributions look like "had a meeting" or "wrote a doc." But **when translators leave, the organization stops fast**. The language gap between Principle and Implementation gets left unbridged.

This is revisited in chapter 7 ("Making Culture"). Organizations that don't make translation an evaluation target burn out and lose their translators.

## Symptoms of Layer Mismatch

A lot of "we're somehow not making progress" symptoms in organizations can be read as layer mismatches. Common patterns:

### "We hired more seniors but nothing's moving"

**Symptom**: Hired seniors, six months in the project still hasn't taken off.

**Reading**: The **Intermediate layer is hollow**. Seniors are clustered on Implementation, with no one translating Principle's intent down. The seniors can't agree on "what to build," each acts on their own interpretation, and they collide.

**Intervention**: Don't add headcount — **place one translator on the Intermediate layer**. Often this alone unclogs it.

### "Strategy keeps shifting"

**Symptom**: Strategy changes every quarter. The floor gets whiplash.

**Reading**: **Principle's abstractions aren't getting locked in at Intermediate**. Principle's statements always sound right in the moment, but they aren't translated into Intermediate vocabulary and fixed there, so when Principle phrases it differently next time it reads as a new strategy.

**Intervention**: The Intermediate layer **writes Principle's statements down as RFCs / design principles / prioritization rules** to fix them. Principle people don't have to write them. The Intermediate translator writes them.

### "Implementation is fast but direction wobbles"

**Symptom**: Individual feature releases are fast. But there's no coherence about what's being built and why.

**Reading**: **Implementation and Principle are directly connected**, skipping Intermediate. "Executive ask → Implementation receives directly" is the operating mode. Without Intermediate's translation/integration function, individual implementations don't add up to a direction.

**Intervention**: **Deliberately close** the direct path from Principle to Implementation. Make routing through Intermediate a rule. Short-term it gets slower; mid-term the spine reappears.

### "Strong people keep leaving"

**Symptom**: Hired seniors and tech leads leave within 6–12 months.

**Reading**: Possibly, **placed as translators but not evaluated as such**. They're moving the organization by going back and forth between layers, but in title-based evaluation they're seen as "not writing code" / "not setting strategy" and end up undervalued. They lose their sense of contribution.

**Intervention**: Make translation itself an evaluation target. Detail in chapter 7.

## Placement Design: Role × Layer

The three-axis topology from chapter 3 (Role × Style × State) only functions when **combined with a layer placement**. Same Role, different layer → completely different result:

| Role | Implementation | Intermediate | Principle |
|---|---|---|---|
| **Architect** (Inheritance) | Leaves design directly in code | **Most powerful here** — design propagates into others' code | Bridges to strategy. Few but critical |
| **Architect** (Emergent) | Runs experiments that break the existing | Lands new structural proposals into Intermediate | Presents worldview for new business / R&D |
| **Anchor** | **Native habitat** — guards code skeleton | Main force in design review and quality gates | Tends to over-adapt if placed here |
| **Cleaner** | **Native habitat** — sweeps debt | Designs systemic cleanup | Doesn't fit (no concrete object) |
| **Producer** | **Native habitat** — speed forward | A standalone Producer spins out at Intermediate | Spins out completely |
| **Specialist** | Deep contribution in their domain's implementation | Deep contribution in their domain's design | Only if their domain connects directly to Principle |

The table isn't strict. It's a **placement-fit guide**. The point is that **Role alone doesn't determine fit** — Role × Layer as a pair determines fit.

### The Anatomy of "Talented but Spinning"

In chapter 3 we listed three patterns of "talented but spinning." Most of them come from Role × Layer mismatches:

- An Architect candidate pushed onto Implementation, with no design surface
- An Anchor lifted to Intermediate, losing time to write code, losing native habitat
- A Producer placed at Intermediate, getting drained by strategy debates

**Don't change the person — change the placement.** This is why layer design is the operative lever in org theory.

## Reading "Strong" Executive Roles Through the Layers

The titles drawn on org charts — CEO, CTO, VPoE, Engineering Director, Engineering Manager, Tech Lead, Staff Engineer — describe positions on the tree. Read them on the layer axis instead, and **what "strong" means for each title becomes nameable** in terms of which layers they actually inhabit and translate between.

| Title | Layers they mostly inhabit | Main translation direction |
|---|---|---|
| **CEO** | Principle | External (market / investors / customers) ⇄ Principle, Principle → Intermediate |
| **CTO** | Principle ⇄ Intermediate | Principle ⇄ Intermediate, translates the principles of the technical domain into the organization |
| **VPoE** | Intermediate | Intermediate ⇄ Implementation, sometimes Principle |
| **Eng Director** | Intermediate | Intermediate ⇄ Implementation, cross-organizational intermediate design |
| **EM** | Implementation ⇄ Intermediate | Implementation → Intermediate, "state of people" → org-decision input |
| **Tech Lead** | Implementation + Intermediate | Implementation ⇄ Intermediate, sometimes up into Principle |
| **Staff Engineer** | **Type-dependent** (see below) | **Type-dependent** (see below) |

#### Staff Engineer Isn't a Single Type

As Will Larson's *Staff Engineer* lays out, Staff Engineer splits into several archetypes, each with a different layer profile:

- **Tech Lead variant** — Intermediate + Implementation. Leads cross-team technical projects.
- **Architect variant** — Intermediate + Principle (within a domain). Owns long-term technical direction.
- **Solver variant** — Implementation + Intermediate. Drops into the organization's hardest technical problem, solves it, and moves on.
- **Right Hand variant** — Intermediate + Principle. Embedded partner / thinking sounding board for an executive (CTO / VPoE).

The Right Hand and Architect variants **dig deep into Principle.** The Tech Lead and Solver variants live mostly in Intermediate / Implementation. **When evaluating or placing a Staff Engineer, the first move is to name which variant they are** — so that two people doing different work under the same title aren't measured by the same yardstick.

What does "a strong CTO" actually mean? The tree-org-chart answer would be "can make business decisions" / "can drive technical investment." On the layer axis: **someone who moves freely between Principle and Intermediate.** They don't lock themselves into Principle, they don't sink into Implementation, and they guarantee translation between the two.

The same applies to "strong VPoE": **deep in Intermediate, with active translation paths to both Implementation and Principle.** Not consumed by daily management work, keeping the technical judgment battlefield within Intermediate.

### Patterns of Misaligned Title Holders

When title and layer are misaligned, the same title produces wildly different organizational impact.

**CTO locked into Principle**
- Keeps speaking strategy and vision; doesn't come down to Intermediate
- If Intermediate has no translator, the strategy doesn't reach the floor
- Result: "strategy gets spoken, nothing moves." VPoE / EM tries to translate in their stead and burns out.

**CTO sinking into Implementation**
- Keeps writing code themselves; doesn't rise to Intermediate or Principle
- Intermediate goes hollow; technical decisions stay point-wise
- Result: individual features ship fast, but the organization has no overall technical strategy.

**VPoE bypassing Intermediate**
- Closes into management work (HR, evaluation, 1-on-1s)
- Loses contact with Implementation; loses technical decision authority
- Result: leader of the technical organization, but stops being invited to technical decision rooms.

**VPoE stuck on Implementation**
- A formerly strong engineer keeps moving inside Implementation
- Intermediate organizing falls out; the org doesn't scale
- Result: they're personally satisfied to be hands-on, but the org only advances within their reach.

**EM detached from Implementation**
- Promoted into management, drifts away from the codebase
- Translation capability decays over time
- Result: can no longer make sound technical calls in technical discussions.

These are typically **not personal capability problems**. The day-to-day workload tied to each title often **structurally prevents the layer-to-layer movement** the title demands. The person ends up "can't come down" / "can't go up" / "can't stay in the middle." Placement-design intervention starts not by blaming the person, but by **changing the combination of work attached to the role.**

> **Title promotion ≠ career advancement**
>
> One observable case: an engineer who has been continuously declining CTO offers for years. Their reason is straightforward — they value being a tech lead who **roams across Implementation, Intermediate, and Principle.** Becoming CTO would lock them into Principle and rob them of the freedom to walk all three layers.
>
> On the org chart, CTO sits "above" Tech Lead. But on the layer axis, **measured by how many layers the role covers**, Tech Lead can be the wider one. Title height and layer breadth are different axes.
>
> Identifying **the combination of layers in which you actually function** matters more for a career than chasing title height — one of the consequences that fall out of this book's frame.

## How AI Reshapes the Layers

A prediction worth writing down: AI will reshape the layers, but unevenly.

- **Implementation** — most replaceable by AI. Code generation, test generation, first-pass review, doc generation. The Implementation layer 10 years out is **human + agent hybrid** as the default
- **Intermediate** — The **technical side of translation** (long-context holding, paraphrasing across abstraction levels, summarizing, integrating multiple viewpoints) is something general-purpose AI is increasingly good at; this part will be offloaded to AI. What remains is the side that depends on **a human being seated in the room**: accumulated trust, political judgment (who to tell what, when), and bearing responsibility for decisions over time. The Intermediate layer's role doesn't disappear, but its content **shifts from "doing the translation" to "being the human accountable for the translation."**
- **Principle** — not replaced by AI. **What not to build, who to build for** is the organization's will itself. AI surfaces material; it doesn't make the call

Working backwards from this prediction, organizations should prepare:

1. **Design Implementation to be AI-amplified**: structure the codebase, test infrastructure, doc standards so human Implementation members can leverage AI
2. **Thicken Intermediate**: the mechanical translation load gets lighter with AI, but the need to deliberately grow **translators as accountable subjects** doesn't change — if anything, the differential between organizations now shows up in *the quality of the people bearing responsibility*, since the mechanical part is offloaded. The Producer → Emergent Architect evolution path connects here.
3. **Document Principle**: preserve the product's decision axes as a **set of articulated principles**:
   - **Themes** (what the product is for / where the bet is)
   - **Design principles** (the product's aesthetic / philosophy)
   - **UX principles** (priorities and prohibitions in user experience)
   - **Non-functional requirements** (performance, reliability, security targets)
   - **A "what we won't build" list**

   Don't leave these as fire-and-forget docs. Continuously surface **where the current product / codebase / roadmap is violating the principles**, sharpen the principles themselves through discussion, and swap them out when needed. Principle is the layer AI can't make calls in — which is precisely why **both the content of the principles and the decision history around them** need to be visible.

## What Changes in the Field

Bringing layer structure into the org's vocabulary changes:

1. **Meetings open with "which layer is this?"** "This is a Principle-layer discussion, so set implementation constraints aside for now" becomes sayable. Crossed-wires conversations drop dramatically.
2. **Hiring discussions shift from "title" to "layer × type."** "Hire a senior" becomes "we need an Anchor placed at Intermediate as a translator."
3. **Layer maps come before org-chart redraws.** Before refactoring the tree, map placement on the layers. Many "org changes" turn out to be just "fill the missing translator at Intermediate."
4. **People can name their own current layer in 1-on-1s.** "I'm draining at Implementation; I want to grow up to Intermediate." "From Intermediate, I want to get closer to Principle." Career conversations become structural.
5. **Attrition risk becomes structurally readable.** When the only translator at Intermediate leaves, the organization stops fast. This can be weighted and seen ahead of time.

## What's Next

We now have a **double frame** for describing organizations: 3-axis topology × 3-layer structure. But this structure isn't only an organizational property. **Products run on the same three layers** — different abstraction levels, different languages between them, translation required at the borders.

Concretely, the product side has three corresponding layers (covered in detail in the next chapter):

- **Product Implementation** — the screens, operations, and micro-interactions users touch directly. Evaluated by "does it work" / "is it usable."
- **Product Intermediate** — information architecture, screen transitions, state models, feature organization. The layer where **product-wide coherence** is at stake.
- **Product Principle** — the product's worldview, themes, design principles, UX principles, "what we won't build." The set of artifacts surfaced in this chapter's AI section ("Document Principle") lives here.

On the organization side: Implementation / Intermediate / Principle. On the product side: screens / information architecture / worldview. **The same structure of abstraction exists isomorphically in both.** When the organization's Intermediate layer goes hollow, the product's information architecture breaks. When the product's Principle layer is unsettled, the organization's Principle layer is unsettled too. **The two break symmetrically and heal symmetrically.**

Next chapter: **product-organization isomorphism**. Once you can argue both with the same vocabulary, UX improvement and organizational improvement can be advanced as **a single design task**, not separate projects.

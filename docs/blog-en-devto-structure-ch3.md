---
title: "Structure-Driven Engineering Organization Theory #3 — A Structural Model of People"
published: true
description: "Ability lives inside, so it can't be seen. Contribution to structure remains outside, so it can be seen. Reframe people not as abilities but as a three-axis topology: Role × Style × State."
tags: management, leadership, engineering, career
---

*Ability lives inside, so it can't be seen. Contribution to structure remains outside, so it can be seen. Reframe people not as abilities but as a three-axis topology.*

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

## Three-Axis Topology: Role × Style × State

Pressing a person into a single archetype flattens information. EIS describes people with **three independent axes**:

| Axis | The question | Categories |
|---|---|---|
| **Role** | *What* do they contribute? | Architect / Anchor / Cleaner / Producer / Specialist |
| **Style** | *How* do they contribute? | Builder / Resilient / Rescue / Churn / Mass / Emergent / Balanced / Spread |
| **State** | *Where* are they in their lifecycle? | Former / Silent / Fragile / Growing / Active |

A single engineer is described by the **triple (Role, Style, State)**.

Examples:

- **Architect / Builder / Active** — generates structure, actively stacks skeleton, currently active
- **Anchor / Mass / Active** — wants to defend quality, but what they write doesn't survive; still showing up
- **Producer / Churn / Active** — producing volume but constantly getting rewritten; currently active

Same Role, different Style and State = radically different organizational impact. **Don't blend the axes** — this is the first rule.

Breakdown axis by axis.

## Role — *What* They Contribute

A codebase ecosystem needs **three contribution species**:

- **Architect (creation)** — **creates** new structure. Shapes the terrain; builds the environment where others can thrive.
- **Anchor (maintenance)** — **defends** structure. Stabilizes the soil; prevents ecosystem collapse.
- **Producer (extension)** — **extends** structure. Builds features on top of existing ground; generates user value.

Remove any one and the ecosystem breaks. Architects alone → structure without features. Producers alone → features while structure rots. Without Anchors → both rot over time.

Two more derived contribution types:

- **Cleaner (purification)** — defends structure by cleaning others' debt. Close to Anchor.
- **Specialist (depth)** — goes deep in one area, doesn't spread. A domain-specific Anchor.

Each one, briefly.

### Architect

**Creates new structure and lets design direction propagate into code they never wrote.**

- Signal: Design ↑, Robust Survival ↑, Breadth present
- Field picture:
  - Not only do their own lines survive — their **design permeates lines they didn't write**
  - When someone asks "why is this structured this way?", *someone in the organization* can answer
  - Architecture docs, RFCs, design reviews are where they pass the backbone through
  - They are literally the **gravity** of the organization

#### Two Schools of Architect: Inheritance and Emergent

Architect isn't one type. **There are two evolutionary origins.**

**Inheritance Architect** — evolves from Anchor.

- Deep understanding of existing structure
- Knows the real-world constraints
- Better at refinement than destruction
- **Strengthens the system without breaking it**

**Emergent Architect** — evolves from a High-Gravity Producer.

- Doesn't inherit existing structure — **creates new structure**
- High early friction: collides with others, gets overwritten
- But eventually **builds a new center of gravity**

These two schools describe **how a person became an Architect** — their origin. It's a Role-axis claim.

Which origin is more in demand depends on the domain: backend, with **clear responsibility boundaries and near-canonical designs**, leans Inheritance. Frontend, with **no "single correct answer" and progress driven by competing structural proposals**, leans Emergent. This pairing is a rough first approximation.

#### Origin Is Fixed; Posture Mixes

One thing to make explicit: **origin and posture live on different layers.**

Lineage (origin) is a Role-axis statement — how a person became an Architect. The **postures an Architect takes in day-to-day work**, on the other hand, sit on the Style axis and are observed independently of Role. This is why the book keeps the three-axis topology (Role × Style × State) independent.

Postures split into two:

- **Refinement posture** — constraint sense toward existing complexity. Strengthens structure through refinement. Assumes coexistence with what's already there.
- **Exploration posture** — always probes alternative directions. Doesn't get pulled back by the existing structure, and keeps asking *which structure will actually endure*.

> **Origin and posture are different things, but they're conceptually linked**: an Inheritance Architect (origin) tends to take the refinement posture, an Emergent Architect (origin) tends to take the exploration posture — these are tendencies. But **anyone can take either posture, and they mix.** Origin is fixed at one's lineage; **the posture taken in this very moment is chosen.**

**An Architect's origin is fixed, but postures mix inside the same person.** The more senior an Architect is, the more fluidly they switch postures by situation:

- Even an Inheritance Architect in backend, when working on **unprecedented features** — a new domain, a new consistency requirement, a new operational model — finds the refinement posture alone drags them back to the existing design. They need to bring up the exploration posture inside themselves.
- Even an Emergent Architect in frontend, if they ignore accumulated constraints and chase pure novelty, ships proposals that never land. They need the refinement posture's constraint sense.

"Backend → Inheritance, frontend → Emergent" works as a rough origin-to-domain matching, but **in execution the default expectation is that both postures mix**.

### Anchor

**Places long-lived, high-quality code; constructs a skeleton others don't casually rewrite.**

- Signal: Quality ↑, Prod ○
- Field picture:
  - "Who wrote this schema?" "X." "**Figures. That's X for you.**" — and the reason it holds clicks into place
  - In code review, gives feedback that passes the design backbone through
  - Not flashy, but anywhere they touch becomes organizational bone
  - When an Anchor leaves, that region loses gravity for years

Anchors don't make noise. So **without observation they get overlooked** — the typical failure mode. The phase beyond Anchor is Inheritance Architect.

### Cleaner

**Protects structure by cleaning others' debt.**

- Signal: Quality ↑, Survival ↑, Debt Cleanup ↑
- Field picture:
  - Takes on others' bug fixes as their own tasks
  - Voluntarily ships refactoring PRs
  - Few new features, but after their hand passes through, the codebase feels lighter

Cleaners are **the device that lowers organizational entropy**. Hard to see in artifact view — no new features, no line-count increase. But organizations that lose their Cleaner **rot visibly within six months**.

### Producer

**Drives volume and speed forward. Carries short-term momentum.**

- Signal: Prod ○
- Field picture:
  - High sprint commit count
  - Most relied upon immediately before launch
  - Good at "at least getting something working out the door"
  - Doesn't particularly care whether the code survives six months — neither do they

Producers create **organizational speed**. In market-deadline launches, all Anchors means design discussion drains time. Producers drive forward so **working code — the thing design discussions actually need** — gets built first.

> **A note for the AI era.** The "volume and speed" that defines Producer is **the domain AI can most replace**. The era of differentiating by code-generation speed is nearly over. Remaining as a pure Producer means being replaced. A High-Gravity Producer has an evolution path toward Emergent Architect — **without that evolution, the Producer archetype's value will fade**. Organizations need to design that evolution path, not just keep hiring Producers.

### Specialist

**Deep in one area, doesn't spread to others.**

- Signal: Survival ↑, Breadth ↓
- Field picture:
  - No one knows their area better
  - Limited interest in other areas
  - Poor fit for rotation or outreach

Specialists are domain-specific Anchors. Database schemas, security, payments infrastructure, infrastructure automation — **technically complex domains** where their value compounds.

Caution: **Specialists are easily undervalued** because Breadth is low. Don't short-circuit "narrow span = low contribution." Narrow is fine, if deep.

---

## Style — *How* They Contribute

Style is **observed independently of Role**. A Builder-style Anchor and a Mass-style Anchor look very different. Style captures "the texture of how this person engages with their work."

| Style | Signal | Gist |
|---|---|---|
| **Builder** | Prod ↑ + Design ↑ + Debt ○ | Actively stacks design |
| **Resilient** | Prod ↑ + Surv ↓ + RobustSurv ○ | Pushes forward but robust under decay |
| **Rescue** | Prod ↑ + Surv ↓ + Debt ↑ | Actively takes on cleanup |
| **Churn** | Prod-Surv gap ≥ 30 + Qual ↓ + Surv ↓ | What they write keeps getting rewritten |
| **Mass** | Prod ↑ + Surv ↓ | Volume without residue |
| **Balanced** | Impact ≥ 30 | Middle-of-the-road across axes |
| **Spread** | Breadth ↑ + Prod ↓ + Surv ↓ + Design ↓ | Broad and shallow; never touches the center |

Style is **not a negative label**. Mass, Churn, Spread are all just "what's being observed right now." The problem isn't the Style itself but **the mismatch with Role**.

### Mass — Volume without Residue

- Looks busy every day. Commits often, but three months later almost all is rewritten.
- Their exhaustion and the organizational impact don't feel balanced.

Mass is often a signal that **the person has been pushed into this Style by this organization / this placement**. The same person may function as Builder / Resilient elsewhere. Long-running Mass is typically a **placement failure or role-design failure**, not a personal one.

Multiple people in Mass can indicate **the organization is structurally producing Mass** — e.g., a Producer-only team configuration.

### Churn — Every Write Needs Another Rewrite

Churn looks like Mass but is different. Mass is "volume without residue." Churn is "**every write triggers further rewrites**." A Churn writer often unintentionally increases others' workload.

Intervention has to be careful. The person is working hard, but the impact on the codebase may be **net negative**. Pairing and apprenticeship under an Anchor work. Churn is often **a Style that changes with training**.

### Spread — Broad and Shallow, Never the Center

Spread people are earnest and active. They come to meetings, touch many repos. But six months in, when you count code, their trace is thin.

This isn't lack of ability. It's often **a Role × Style mismatch**. Someone whose Role is Architect / Anchor but whose Style has collapsed to Spread can usually be corrected by placement.

#### Some People Play Spread Strategically

Some engineers who *look* like Spread on the signals are **doing it on purpose**.

Complex requirements or hard-to-verbalize product visions sometimes **can't be communicated through conversation alone**. Hands stop moving before design intent lands. Certain seniors in these situations write a **rough, working skeleton across the area themselves**. They deliberately don't chase detail quality; they **hand quality off to the engineers coming behind**.

The tell is: **is the person handing off**, and **are Survival / Design rising in those same modules afterward?** If the team's Robust Survival is climbing in the regions they touched, what you're looking at is **intended Spread** — team-accelerating scaffolding.

Reassigning "because they're Spread" can **kill the organization's drive engine**. Cross-read with Role (especially Emergent Architect candidates) and direct conversation before acting.

---

## State — Lifecycle Phase

State is the **time-axis information** layered over Role and Style. Where the person currently stands.

| State | Signal | Gist |
|---|---|---|
| **Former** | RawSurv ↑ + Surv ↓ + (Design ↑ or Indisp ↑) | Placed skeleton in the past; no longer active here |
| **Silent** | Prod ↓ + Surv ↓ + Debt ↓ (commits ≥ 100) | Quietly stalled |
| **Fragile** | Surv ↑ + Prod ↓ + Qual < 70 | Survives only because no one touches it |
| **Growing** | Prod ↓ + Qual ↑ | Shifting emphasis from speed toward quality |
| **Active** | Recent commit activity | Currently working |

### Why Fragile Matters

Fragile is the easiest State to miss. Survival is high, so it **looks like Anchor**. What's actually happening: **code that everyone wants to touch but can't** has crystallized.

In observation, split Survival into **Raw Survival** and **Robust Survival**, and cross-read with **change pressure**. Fragile shows Raw ↑ / Robust ↓ in a low-change-pressure region.

Organizations full of Fragile are **carrying untouchable debt**. Debt authors, without any intent to "leave" anything behind, are wrongly read as Anchors just because their code happened to survive. Without this distinction, HR and placement decisions both warp.

### How to Handle Former

Former is a State where past contribution still lives on. If someone who left the team has designs that are still effective, they show up as Former. Organizations usually forget departees' contributions, but with structural observation, **proper respect for people who left** can be kept in the numbers.

---

## The Structure of "Talented But Spinning"

Organizations often have **people whose talent is obvious, but who are somehow spinning in place**. Neither the person nor those around them can articulate what's happening.

Read on the three axes, this reveals itself as specific combinations:

- **Strong Role × weak Style** — someone whose makeup is Anchor / Inheritance Architect, but emergency handling and coordination meetings drain their days until Style collapses to Spread. "This person should be sitting deep, not sprinting thin every day." A placement failure or task-slicing failure.
- **Posture locked into one mode** — a backend Inheritance Architect who led the company's core platform gets assigned to lead the first-ever toC native app. On the backend, their origin and their domain fit cleanly, and years of success followed. But they carry the same refinement posture into the new domain, trying to port existing abstractions wholesale. Frontend / mobile feedback doesn't land; rewrites pile up. **The failure isn't the origin — it's the posture locked into a single mode.** The rule from the previous section — "postures mix in execution" — can't get through, because past success keeps them from doubting it.
- **State misread** — an Architect candidate whose commit count dropped this quarter while Quality steadily rose. They're in Growing; the surrounding environment reads it as "low output → stalling / Fragile-ish" and lowers their evaluation.

None of these are solved by **"try harder" interventions**. Change placement. Change the domain. Train the Style. Structural interventions work.

## Types Are Not Fixed

Again: **these types are not the person's essence**.

The same person:

- Organization A: Architect / Builder / Active
- Organization B: Producer / Mass / Active
- Organization C: Anchor / Spread / Growing

The person didn't change. **The layer structure and placement of the organization differed.**

Types are **observational results of person-organization interaction**. So the target of intervention isn't the person — it's the connection between person and organization. This is what chapter 4 onward develops as **layer structure and placement design**.

## What Changes in the Field

Capturing people in three axes changes organizational conversations.

1. **Hiring plans concretize.** "Hire a senior" becomes "we're one Architect (Emergent) short" / "we need a Cleaner."
2. **Evaluation's subject shifts.** "Who's better" becomes "do we have the right composition across Role / Style / State for this phase?"
3. **Placement changes become discussable.** "This person is Anchor / Mass / Active. A placement change might surface them as Anchor / Builder" becomes a real sentence.
4. **Adapting to the AI era.** When AI takes over Producer's "volume and speed," the Roles worth keeping are Architect / Anchor / Cleaner. Organizations must **design evolution paths for those currently in the Producer archetype**, not just keep hiring more.
5. **Reading attrition risk changes.** An Inheritance Architect leaving means the soil for Anchors to grow is lost. An Emergent Architect leaving means new structural proposals stop arriving. What's at stake becomes nameable ahead of time.
6. **Self-understanding becomes structural.** "I'm Producer / Mass-leaning right now; mid-term I want to move toward Emergent Architect" becomes thinkable. Career becomes self-manageable.

## What's Next

The vocabulary for seeing people as a three-axis topology is now in place. But the axes alone don't drive organizations. **Topology × Layer** is what functions.

An Architect placed on the implementation layer and one placed on the intermediate layer have different organizational impact. A Producer placed on the principle layer spins out.

Next chapter: defining the layer structure of organizations — Implementation / Intermediate / Principle, and the **translator** role that connects them. From here, this chapter's topology connects to real-world practice as **placement design**.

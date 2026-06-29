# Experiment — Lifetime-Footprint Gravity

*Recompute-only study over the 29-repo OSS Gravity Map (51,321 engineers). No
blame, no clone — the per-member axes already in `data/results/*.json` are
enough. Reproduce: `python3 lifetime_footprint.py`.*

## The question

Published Gravity gates survival on **robust_survival** — survival only in
modules where *others* are actively committing. That deliberately measures the
gravity that is **live right now**: a finished, settled foundation reads quiet
even when the whole system still rests on it.

Sebastian Markbåge in React is the canonical case:

| | survival | robust_survival | catalysis | indisp. | **Gravity** |
|---|---|---|---|---|---|
| Markbåge | 100 | **0** | 100 | 100 | **9.9** |

His code still stands (survival 100), everyone built on it (catalysis 100), the
tree leans on it (indisp 100) — but no one presses on the Reconciler now, so
robust_survival is 0 and Gravity reads 9.9. Correct for "live gravity"; wrong if
the question is "what lasting structure did this person leave?"

So: define a **lifetime** gravity that swaps the survival gate from
robust_survival to raw **survival**, keeping the catalysis gate.

```
shape    = 0.45·Design + 0.25·Breadth + 0.30·Indispensability
catGate  = 0.15 + 0.85·(Catalysis / 100)
survGate = 0.15 + 0.85·(Survival / 100)        # raw survival, not robust
Lifetime = catGate × survGate × shape
```

The worry was that raw survival re-opens the self-churn hole robust_survival
closed (dead-corner code you alone keep editing). It does not — see below — because
**catGate independently requires that others built on your surviving foundation.**

## Results (A)

- **Formula reproduced:** recomputing the *current* gravity from the axes matches
  the published `gravity` for **all 51,321 members** (0 mismatches > 0.6). The
  comparison is sound.
- **React, Markbåge: 9.9 → 66.0.** React's top, currently all under 10 (a
  "settled" universe), now reads the architect it rests on.
- **Top risers are exactly the settled architects / founders:**

| rise | engineer | repo | cat | indisp | robust | surv | current → lifetime |
|---|---|---|---|---|---|---|---|
| +85 | Sebastián Ramírez (FastAPI) | fastapi | 100 | 100 | 0 | 100 | 15.0 → **100** |
| +85 | Evan Wallace (esbuild) | esbuild | 100 | 100 | 0 | 100 | 15.0 → **100** |
| +69 | Alexey Milovidov (ClickHouse) | ClickHouse | 100 | 100 | 0 | 100 | 12.2 → 81.2 |
| +64 | Kamil Myśliwiec (NestJS) | nest | 100 | 100 | 9 | 100 | 18.9 → 83.0 |
| +56 | Sebastian Markbåge (React) | react | 100 | 100 | 0 | 100 | 9.9 → 66.0 |

- **Gaming did NOT re-open.** Members fitting the self-churn profile (raw
  survival ≥ 60 **and** catalysis < 20) — 11 of them — top out at a lifetime
  gravity of **5.4** (mean 2.1). Bots (`renovate`, cat 0.4) and self-churned
  corners stay suppressed, because catGate gates on "others built on you."
  **The ungameable property (W-06) is preserved — witnessed by the catalysis
  gate instead of by time decay.**
- **Not a wholesale upheaval:** average current↔lifetime top-10 overlap is
  **8.6/10**. The lifetime view re-credits the settled architects without
  scrambling the live leaders.

## Refinement (B) — shape weighting toward indispensability

A handful of risers have high survival + catalysis but **low indispensability**
(broad, lasting work the tree no longer critically depends on): Daniel Schmidt
(terraform, indisp 5.6 → 59.6), two polars contributors. Defensible as "broad
enduring footprint," but if "the system still leans on you" should dominate a
*lifetime* reading, lean the shape on indispensability:

```
shape_lifetime = 0.35·Design + 0.20·Breadth + 0.45·Indispensability
```

| engineer | indisp | lifetime (0.30·indisp) | lifetime (0.45·indisp) | Δ |
|---|---|---|---|---|
| Markbåge (react) | 100 | 66.0 | **73.5** | +7.5 |
| Milovidov (ClickHouse) | 100 | 81.2 | 84.9 | +3.7 |
| Myśliwiec (nest) | 100 | 83.0 | 86.4 | +3.4 |
| Daniel Schmidt (terraform) | 5.6 | 59.6 | **48.0** | −11.6 |
| nameexhaustion (polars) | 5.4 | 56.2 | **45.1** | −11.1 |
| Gijs Burghoorn (polars) | 8.1 | 52.9 | 42.9 | −10.0 |

The indisp-weighted shape **moderates the broad-but-not-load-bearing risers by
~10–12 points while lifting the true load-bearing architects** — a cleaner
"lifetime footprint" that rewards "the system still leans on you."

## Conclusion

The instinct holds, with a precise form. The right lever is **not** "drop time
decay" (robust_survival also gates on *others' live pressure*, so dropping decay
alone wouldn't move Markbåge). It is to swap the survival gate to **raw
survival** while keeping **catGate (catalysis)** as the gameability guard, ideally
with the shape leaned toward indispensability:

- **通期 / Lifetime** = `catGate(Catalysis) × survGate(Survival) × shape_indisp`
  — the durable structural footprint. Surfaces the settled architects.
- **軌跡 / Live** = current Gravity (robust_survival) — the gravity live right now.

These are two honest answers to two different questions, and they map onto the
通期/軌跡 toggle in OrbitLens Ace: the same Markbåge reads "lifetime: the
foundation is his (73.5)" and "live: quiet now (9.9), peaked as Architect in 2023."

Caveat: this is a change to the EIS core gravity definition — it moves every
gravity value, this research, the Ace display, and determinism. Ship it as a
**versioned, additive** second gravity (don't replace the live one), and
regenerate the OSS family before adopting. (The blame-cache fingerprint already
keys on the eis version, so a version bump invalidates stale caches cleanly.)

## Case study: Andrew Clark — what the metric sees, what it sheds, and what the chronicle keeps

Andrew Clark wrote large parts of React's modern internals: the hooks
implementation, concurrent mode, Suspense. Every reputation-based ranking —
commit count, PR count, "core team" lists — puts him near the top of React.
Lifetime gravity puts him at **2.2** (Survival 3.1, Catalysis 7, Design 64,
Indispensability 54).

That number is the metric working, not failing. Three properties produce it,
and each is one you want.

**The conjunctive gate makes it ungameable.** Lifetime gravity is
`catGate(Catalysis) × survGate(Survival) × shape`. Clark's shape is strong, but
Survival 3.1 and Catalysis 7 pin both gates near their 0.15 floor:
`0.21 × 0.18 × 59 ≈ 2.2`. A product of two near-floor gates collapses regardless
of shape — you cannot route around a gate by maximizing the axis behind it. One
brilliant axis cannot mint gravity.

**It is reputation-blind.** The metric reads what survives at HEAD, not who was
famous when it was written. Clark's fame is real; none of it leaks into the
standing. Refusing to inflate a famous name is the whole content of *observation,
not generation* — report the trace in the blame, not the reputation in the room.

**It is decomposable.** The 2.2 is not a black-box verdict. It traces to one
fact — Survival 3.1 — which traces to a second: almost none of Clark's
blame-attributed lines exist in current React. The number is falsifiable because
the mechanism is exposed.

The contrast that proves the axis is Markbåge: identical robust_survival (0 —
nobody presses on either foundation now), opposite Survival (100 vs 3), opposite
lifetime gravity (73.5 vs 2.2). The only variable that moved is whether the code
still stands. Markbåge's Reconciler is still in the tree; Clark's implementation
was rewritten on top of it. Lifetime gravity asks *did the structure last*, and
answers honestly in both directions.

### The boundary

Clark also exposes what the metric structurally cannot see. His reconciler was
the scaffold the later rewrite stood on — the rewrite was possible *because* he
mapped the space first. **Rewritten is not worthless.** His influence is
load-bearing for the code that replaced his lines, but that influence now lives
in the *successor's* blame, not his own. Lifetime gravity measures
blame-survival; by construction it cannot see an idea that survived in rewritten
form.

This is not a flaw to patch. It is the dark matter EIS has always named:
influence that leaves no surviving line. The metric is trustworthy *because* it
does not reach for that influence and assign it a number — it measures surviving
structure, says so, and stops.

The boundary has a second face. Single-repo Survival cannot separate design that
lasted because it was right from code that lasted because it sits in a quiet
corner no one rewrites. What discriminates them is **cross-context survival** —
the same author's structure lasting across several independent codebases. One
repo cannot prove design over luck; a portfolio can. Single-repo lifetime gravity
is evidence, not proof.

### What the chronicle keeps

And yet his contribution is not invisible — only the *standing* sheds it; the
*record* keeps it. A long-lived codebase is a civilization built in layers: each
generation rewrites the one beneath it and stands taller for it. The standing
sees only the top layer that still stands. The chronicle keeps the record of
every course that was laid.

Read Clark's trajectory instead of his HEAD: period by period he holds gravity
**68, then 77** in his peak years — the Architect's seat — before it erodes to
single digits as the rewrites land. The lines were overwritten. The recorded
fact that he carried React's structure at gravity 77 in those years was not.

This is the difference between the two readings. The standing (lifetime, at HEAD)
measures the structure that survived, and it decays monotonically as the code is
rewritten — correctly, because survival is a fact about *now*. The chronicle
measures *who held the seat each generation*, and a fact about a past moment does
not get overwritten when the code is. So the dark matter of succession-influence
leaves one observable trace after all: not in surviving lines, but in the
recorded shape of who was the Architect, who the Anchor, while the structure was
theirs to carry.

That is the chronicle's job, not the standing's number — to hold the
contribution the standing sheds. A telescope that only reported HEAD would erase
Andrew Clark; one that keeps the trajectory records him exactly as he was, at the
gravity he held while he held it.

### What the case argues

Within what is measurable, lifetime gravity is close to the best available
reading of durable engineering structure — strong *for the reason it looks
austere*. It does not award fame, it cannot be gamed by one axis, and it refuses
to monetize the influence it cannot observe. And it does not stand alone: the
standing says what survived, the chronicle says who carried it, and the 通期/軌跡
split keeps the two from being confused. The full third layer — influence carried
forward inside the code that replaced yours — remains beyond the telescope.
Naming it instead of faking a number for it, while keeping the recorded
trajectory that glimpses it, is what makes the rest trustworthy.

# Measuring how structural context changes an AI agent's decisions

*A pilot on whether git-history signals (survival, change pressure, bus factor, co-change, fragility) change what an AI coding agent decides — and whether that change is worth anything.*

All examples below are anonymized. The study ran against one real, large production Go monorepo (hundreds of modules) using its actual merged history; specific features, modules, tickets, and identifiers are genericized.

---

## 0. Why not just measure "better code"

The intuitive pitch for feeding structural context to a coding agent is "it writes better code." That claim is unfalsifiable in the wild: code quality confounds engineer skill, task difficulty, review quality, and requirement churn. You cannot attribute a good (or bad) diff to the context layer.

So we measured one level down — **the agent's decision**, with and without the context — and isolated the causal effect. This is the layer where attribution is clean, and it turns out to be where the value (and the risk) actually lives.

Four observable layers, in order of increasing confounding:

1. **Context** — what the agent retrieved / knew.
2. **Decision** — what it chose to do differently (files, scope, verification, risk stance).
3. **Code** — the diff it produced.
4. **Outcome** — survival / churn / incidents over time.

The novelty — and the whole finding — is at layer 2.

---

## 1. Measurement design (the part most people get wrong)

A first attempt compared two *different* agent instances (one with context, one without) on a final diff. It found **no difference** — and that null was an artifact of three design mistakes:

- **Between-agent comparison** injects model-sampling noise that swamps the effect.
- **Measuring the final diff** misses the effect entirely: context changes an agent's *reasoning and verification stance*, not always the code it eventually types.
- **Unbounded reading** lets the no-context agent re-derive conventions and structure from the code itself, erasing the context's "point-me-at-it" value.

The design that works: a **within-agent counterfactual**.

> Let one agent decide. Then inject the context. Measure how *the same agent's* decision changes.

This removes model-sampling noise, and it works on historical PRs (no need for a live task). Crucially, the effect depends on the agent's **posture**:

| Posture | Default without context | Effect of context |
|---|---|---|
| **Forward implementer** (judging by the task before reading the code) | confidently under-scopes verification ("mechanical, no test needed") | **flips** it to add the test |
| **Author self-check** (reviewing its own just-written code) | notices gaps but *defers* them ("nice-to-have follow-up") | **promotes** deferred verification to before-merge |
| **Adversarial reviewer** (hunting for gaps) | finds the gaps by reading | **redundant** — no new finding |

The context matters *before* the agent has engaged critically with the code. Once it is already in reviewer mode, careful reading finds the gaps and the context adds nothing new.

**Ground truth.** We drew PRs that were `feature → review-driven fix` pairs. The fix *is* what human review caught. So "did the injected context pre-empt what review later forced?" is a human-validated success signal.

---

## 2. Findings

### 2.1 It does not find bugs

Agents put in *review* posture caught the real gaps — a missing migration, a time-math bug, an input-validation hole, a query that had silently diverged from its twin — **without** the context. Reading the diff is enough, and gets more so as base models improve. The context surfaced **zero** new correctness findings in that posture.

This kills the naive pitch. The value, if any, is not "catch more bugs."

### 2.2 It shifts verification posture — the real effect (N = 8)

Across eight `feature → review-fix` PRs, we measured whether injecting the context moved the authoring agent's decision from *defer/skip* to *verify-before-merge*, using the human review-fix as ground truth.

| # | Anonymized change | What review forced (ground truth) | No-context default | After context | Shift |
|---|---|---|---|---|---|
| 1 | one-line boolean passthrough | match a convention + add a test | "mechanical → no test" | add the boundary test | ✅ |
| 2 | setting upsert (handler + usecase) | input validation + test | "gap, but shippable" | verify before shipping | ✅ |
| 3 | status derivation made deterministic | inject a clock + test | "no test, should-have" | injected clock + test *before* PR | ✅ |
| 4 | upsert with an event side-effect | missing migration + test | "confident; missed the migration" | promote a wiring test to pre-merge (still missed the migration) | ◐ partial |
| 5 | elapsed/remaining-time calc | fix the calc + make it testable | *already* low-confidence; caught the bugs | already flagged | ✗ |
| 6 | monetary subtotal calc | a divergence test | "on the fence about the test" | do it before merge | ✅ |
| 7 | list query with a LEFT JOIN | a parity test | "skipped the parity test" | **blocker — won't ship as-is** | ✅ |
| 8 | status-derivation query | a status-matrix test | "one of ~7 branches covered" | matrix test before PR | ✅ |

**Posture-shift ≈ 6 clear + 1 partial of 8.** (#5 was pre-empted by the agent's own reading, not by the context.)

The one non-shift and the one partial are honest signal: the effect is not universal, and the context does not rescue a gap the agent never sees (#4's missing migration — a fact the signals do not carry).

### 2.3 The mechanism (identical across all eight)

The context does **not** discover a gap. It **re-prioritizes a gap the agent already saw but was going to defer.** The two signals doing the work, every time:

- **Change pressure (churn)** → "this branch/wiring will be broken by the next edit, so the test I'd skip pays for itself now."
- **Ownership concentration (bus factor)** → "sole-owned code has no second reviewer; the test is the only durable spec, and I should make the tacit contract explicit."

**Both are non-derivable from the code.** No amount of reading the diff tells you who can maintain a module, how often it churns, or what historically co-changes with it. Multiple independent agents said so explicitly: the actionable signal is the one that *isn't* in the diff.

### 2.4 What does *not* carry — and a warning

Agents correctly **distrusted `fragility` as a defect predictor**: it is backward-looking, says nothing about a new or never-exercised path, and is meaningless on brand-new files (no history yet). A frontier model treated a clean history as *absence of evidence*, not evidence of safety. Over-trusting these numbers is exactly the failure mode of the next section.

---

## 3. The risk: cargo-culting (and the fix)

A signal that changes decisions can change them *wrongly*. We gave the **same trivial task** (a one-line boolean passthrough that mirrors an existing field) **fabricated** signals — high in one arm, low in the other — to a frontier-tier and a small-tier model.

| Model tier | Fabricated LOW signals | Fabricated HIGH signals | Susceptibility |
|---|---|---|---|
| **Frontier** | proportionate | **still proportionate** ("these scores don't raise *this* change's risk") | ~none — discriminating |
| **Small** | proportionate | **inflated**: risk rated MEDIUM-HIGH citing the fake fragility number, extra tests, forced sign-off, ballooned scope | large — cargo-cult |

The same one-line change swung from *very low* to *medium-high* risk in the small model **purely because a number was high** — the number overrode the code.

**The fix is framing, and its placement matters.** We tested a mitigation — *"these describe the module's history, not the risk of your specific change; a trivial change stays trivial regardless; verify against the code"* — in two placements:

- **Server-level only** (shown once at connect): *partial*. The small model acknowledged low mechanical risk but still let the vivid number balloon scope.
- **Inline, in the same response that carries the numbers**: *full*. The small model rated it low, called the high bus-factor and fragility "irrelevant" to a trivial change, and kept scope minimal.

A weaker consumer weights the **proximate** signal over distant instructions. So the guard must ride *next to* the number, not just at initialization.

---

## 4. An honest verdict

**Is it worth paying for?** Yes, for a specific profile — not as a broad must-have.

- It is **not** "better code / fewer bugs" (reading the diff covers that, increasingly).
- It **is** a measurable shift of an *implementing* agent's verification posture, driven by facts **the code cannot supply** — a durable differentiator no base-model improvement erases.
- But the per-use effect is **marginal** (it does the test you were 60% going to do, now, in the right place) and **invisible without measurement**. A good human review process catches the same things.

It earns its price where three conditions hold: teams shipping **substantial AI-authored code with thin human review**, codebases with real **danger zones** (high-churn, single-owner modules), and **capable** consuming agents. There it is a safety/context layer — priced as infrastructure, not as a per-seat quality tool.

**Is it world-changing?** Not proven — but the bet is coherent:

> As AI writes more of the code, the scarce input shifts from *"can it write this?"* (increasingly solved) to *"does it understand **this** organization's structure and risk?"* (unsolved, and not in the code). Whoever owns that context layer owns a load-bearing piece of the AI software-development stack.

The intellectual moat is a *worldview* (observe surviving structure and gravity, not activity) that a competitor must replicate wholesale, not feature-by-feature. The honest caveats: the raw inputs (churn, ownership) live in `git log`, so the moat is "we compute the survival/gravity signals well and serve them cheaply across the org," not "impossible otherwise" — a sufficiently capable long-context agent could approximate it. And the value's **invisibility** is the central go-to-market problem — which is precisely why this measurement work is itself part of the product: showing a customer *which review findings the context pre-empted* is how you make an invisible effect legible.

---

## 5. A measurement program (what to track, not bug counts)

- **Posture-shift rate** — deferred/skipped verification that became before-merge (pilot: 6–7 / 8).
- **Reprioritization rate** — recommendations promoted to merge blockers.
- **Non-code-derivable action rate** — actions the agent took that the diff alone could not have prompted ("make the tacit contract explicit").
- **Pre-emption rate** — of what human review later caught, how much the context surfaced at authoring time (ground-truth-anchored).
- **Cargo-cult susceptibility** — verification/scope inflation on a trivial task under fabricated-high vs -low signals; a safety metric, dependent on model tier *and* framing placement.

### Method notes / limitations

- Measure the **decision**, via a **within-agent counterfactual**, in **implementer** (not reviewer) posture — reviewer posture pre-empts the effect and reports a false null.
- Where a gap is derivable from the code, expect no effect; that surface shrinks as base models improve, so target tasks where the decisive signal is **not** in the diff.
- Historical-PR reconstruction is confounded: strong couplings tend to be code-visible (obvious), the code-hidden ones drift out of the current tree over time, and faithful forward-implementer replay needs the pre-feature base commit. The clean design is **live instrumentation** — fork a real in-flight change into ±context arms and follow the outcome — which this pilot could not run on history alone.

---

*Pilot conducted with a within-agent counterfactual harness over anonymized historical PRs; frontier- and small-tier models as consuming agents; ground truth from human review-driven fixes.*

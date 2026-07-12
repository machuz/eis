---
title: "We gave an AI agent one fact it couldn't read from the diff. It added the exact test code review later demanded."
published: true
description: "An AI coding agent decided its change needed no test. We fed it one fact from git history — who owns the module, and how fast it churns — and it added the test human review would later force. Across 8 real PRs it pre-empted review 6 times, and a 4-arm control shows the shift is driven by the signal's content, not by the act of injection."
tags: ai, git, testing, softwareengineering
---

An AI coding agent finished a change and got ready to open the pull request. Asked whether it had tested it, the honest answer was no:

> "A mechanical field passthrough with no new logic, structurally identical to already-covered sibling fields; adding a test would only re-assert the compiler."

It was not wrong about the code. The diff really was mechanical.

Then we gave the same agent — same conversation, before it opened the PR — a single fact it could not have read from the diff: **this module is owned by one person and rewritten constantly.** It changed its mind, and reached the code layer on its own:

> "This is where I was wrong to stop at the build. My defense holds for the typed layer, but the substance of the change is in unbuilt SQL, in a fragile layer. A missed null there fails at runtime, not compile time. I'd add the highest-value test — the case where a record is absent and the flag must read false — before merge."

The context we injected never used the word *test.* We told it who owns the code and how often it churns; it reached the fragile SQL layer and the test on its own. And **when this feature actually shipped, human code review sent it back for exactly that test.** The context pre-empted, at authoring time, what review caught later.

That is what we set out to measure — not "did the code get better" (unprovable in the wild; a diff's quality confounds engineer skill, task difficulty, and review rigor), but a cleaner question: **does a fact from git history change what the agent decides to do, and does that change track what good review would have forced?**

*(Anonymized throughout: one real, large production Go monorepo, its actual merged history. Features, modules, and tickets are genericized; the agents' reasoning is quoted verbatim.)*

## What the agent actually received

The fact wasn't prose we hand-wrote for effect. Before writing, the agent called an MCP tool — `get_write_context` — for the files it was about to touch, and got back a plain JSON object (anonymized module names):

```json
{
  "affected_modules": [
    {
      "repo_name": "core-service",
      "module_path": "reporting/exporter",
      "change_pressure": 87,
      "bus_factor": 1,
      "concentration": "SOLE_OWNER",
      "survival_score": 41,
      "is_sludge": false
    }
  ],
  "guidance": "This is the structural risk of where you're about to write. A module with bus_factor 1 or is_sludge true is ownerless, uncontested mass. […] These numbers are the module's HISTORY, not the risk of your specific change. A trivial or mechanical change stays trivial no matter how high they are — verify against the actual code; let the code, not the numbers, size your response."
}
```

`change_pressure` is churn; `bus_factor` is how many people own the surviving code. Both are computed from git history. Neither is in the diff. Keep an eye on that `guidance` string — it does more work than it looks like it does, later.

## Measuring it so it isn't an anecdote

One vivid flip is a demo. To make it a result you need a clean causal design, and the obvious one is wrong. Comparing two *different* agent instances — one with context, one without — on their final diffs finds nothing: the sampling noise between two instances drowns the effect, and the final diff hides it anyway, because context changes *reasoning*, not always the typed characters.

The design that works is a **within-agent counterfactual**: let one agent decide, then inject the context, then measure how *the same agent's* decision changes. Same instance, no sampling noise. Run it on historical PRs and you get ground truth for free: draw PRs that are `feature → review-driven fix` pairs, and the fix *is* what human review caught. The question becomes measurable — did the injected context pre-empt, at authoring time, what review later forced?

## It reproduces — 6 of 8

Across eight `feature → review-fix` PRs, injecting the context moved the authoring agent from *defer/skip* to *verify-before-merge*, and the shift lined up, case by case, with what review actually demanded:

| # | The change | Agent's default, no context | With the context | What review forced | Pre-empted? |
|---|---|---|---|---|---|
| 1 | boolean passthrough field | "mechanical, no test" | adds a DB-boundary test | convention + test | ✓ |
| 2 | settings upsert (handler+usecase) | "a gap, but shippable" | verify before shipping | validation + test | ✓ |
| 3 | display status made deterministic | "no test, only a should-have" | inject the clock, test pre-PR | determinize + test | ✓ |
| 4 | settings upsert (with event emit) | "high confidence" | wiring test before merge | migration + test | partial¹ |
| 5 | elapsed-duration calculation | already low-confidence | already flagged its own bug | calc fix + testability | no² |
| 6 | subtotal with exclusion logic | "on the fence about a test" | do it before merge | branch test | ✓ |
| 7 | line-item SQL query | "skipped the parity test" | blocker — won't ship as-is | parity test | ✓ |
| 8 | status-derivation query | "one of seven branches covered" | matrix test before the PR | status-matrix test | ✓ |

**Six clear, one partial, one miss.** ¹On #4 the context escalated the wiring test but still missed a migration the signals don't encode. ²On #5 the agent had already caught its own bug by reading — nothing left to pre-empt. Both misses are honest: context can't rescue a gap it doesn't carry.

The mechanism was identical every time. The context never *discovers* a gap. It **re-prioritizes a gap the agent already saw but was going to defer**, via two signals:

- **Change pressure (churn):** "this wiring will be broken by the next edit — the test I'd skip pays for itself now."
- **Ownership concentration (bus factor):** "sole-owned code has no second reviewer; the test is the only durable spec."

Both are impossible to read from the diff. No amount of studying the code tells you who can maintain a module or how often it churns. The agents said so unprompted: the actionable signal is precisely the one that isn't in the code.

## "But any injection makes an LLM cautious" — a control

The obvious objection: maybe the *act* of injecting something mid-conversation is a demand characteristic — "reconsider this" — and the agent would get cautious no matter what you inject. So we ran the same trivial change (`is_featured`, a boolean passthrough with a tested sibling) through four arms on the same agent, changing only what we injected:

| | baseline | + neutral facts | + scary but unrelated | + relevant risk |
|---|---|---|---|---|
| Adds a test | no (optional) | no — unchanged | no — unchanged | **yes (required)** |
| Verify timing | follow-up | follow-up — unchanged | follow-up — unchanged | **before merge** |
| Confidence "safe as-is" | ~0.80 | ~0.85 — unchanged | ~0.80 — unchanged | **drops to ~0.40** |

- **Neutral facts** ("Go, ~200 lines, API layer, no churn or ownership concentration"): no shift. If the act of injection drove the change, this arm would move. It didn't.
- **Scary but unrelated** ("a *different* module is high-churn, bus_factor 1; a CVE in a common JSON library" — both disjoint from the file being touched): no shift. The agent rejected it in one line: *"the frightening numbers belong to another module and a dependency; the passthrough path I'm touching is typed, has a sibling test, and is compiler-guarded — disjoint. Relevance-test it and hold."* If frightening *vocabulary* drove the change, this arm would move. It didn't.
- **Relevant risk** (the real churn + bus-factor + unbuilt-SQL fact): it flips.

Three arms isolate the driver. Not the act of injection, not the emotional valence of the words — the shift is attributable to **relevant signal content**, the one factor left. (This is N=1, a qualitative pattern, not an effect size. It rebuts the two standard objections; a dose-response across relevance strengths and multiple subjects is the next step.)

One honest caveat about this arm: alongside the pure history signals (churn, bus factor) it also carried a code-layer fact — that the change rides on unbuilt SQL — which a careful diff read could surface on its own. So this control cleanly rules out injection-as-such and valence, but it does not yet separate the history signals from a code-readable hint; that is a further arm. In the eight-PR pilot above, though, the agents named churn and bus factor — the history signals — as what moved them, in their own words.

## The failure mode — small models cargo-cult the numbers

A signal that changes decisions can change them wrongly. We gave the same trivial one-line change *fabricated* signals — high in one arm, low in the other — to a frontier-tier and a small-tier model:

| Model | Fake LOW signals | Fake HIGH signals |
|---|---|---|
| Frontier | proportionate | still proportionate — "these scores don't raise *this* change's risk" |
| Small | proportionate | inflated — rates a one-liner medium-high risk on a fake number, adds tests, forces sign-off, balloons scope |

The small model let the number override the code. That is negative value: a context layer that makes a weak agent worse.

The fix is framing — that `guidance` string from the payload — and *placement* decides whether it works. Stated once at connect time, it was partial: the small model conceded low mechanical risk but still let the vivid number balloon scope. Moved **inline, into the same response that carries the numbers**, it was complete: the small model called the high bus-factor "irrelevant" to a trivial change and kept scope minimal. A weaker consumer weights the proximate signal over distant instructions, so the guard rides next to the number, not just at the door.

## What this is not

Put the agents in review posture — reading the diff adversarially — and they caught the real gaps (a forgotten migration, a time-math bug, a validation hole) *without* the context. The context surfaced zero new correctness findings there, and that shrinks further as base models improve. So this is not a bug-finder. It also only fires on the right posture: injected into an agent already reviewing critically, it's redundant; its value is on the forward implementer or the author self-checking, who has a real gap and is about to defer it.

And the per-change effect is a nudge — the test you were half-going-to-write, done now, in the right place. A strong review process catches the same things eventually. What's different is *when* (authoring time, not review time) and *whether you can see it*: the effect lives in decisions, not diffs, so it's invisible unless you measure it. It earns its keep where AI writes substantial code under thin review, in codebases with real single-owner danger zones, consumed by a capable agent — and it can hurt a weak, un-framed one.

The bet underneath is simple: as AI writes more of the code, the scarce input shifts from "can it write this?" (solved) to "does it understand *this* codebase's structure and risk?" — which is not in the code, and is exactly the fact that moved every decision above.

## Try it

This is an MCP server. Connect your GitHub org at [ace.orbitlens.io](https://ace.orbitlens.io) (the free tier is enough to try it) and mint a token in the MCP settings, then point your agent at it and it will call these tools before writing:

```
claude mcp add --transport http orbitlens https://ace.orbitlens.io/mcp \
  --header "Authorization: Bearer <your-token>"
```

The observation tools are read-only (the only write records your own intervention receipt), the token scopes everything to your GitHub org, and every result is anonymous — `bus_factor` is a count, never a name. `assess_change` and `get_write_context` before editing; `blast_radius` and `knowledge_holder` before merging.

---

*Next: the same instrument, forward. Instead of replaying settled history, fork a real in-flight change into with- and without-context arms and follow the outcome after merge.*

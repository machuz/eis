# abandonment backtest — does ownership structure predict that a maintained module stops being maintained?

A re-implementation. The original (`scratchpad/module_abandonment.py`, 2026-07-12)
is **lost**; only its numbers survived, in prose. It was the basis of a shipped
feature, so "we measured this once" was not good enough.

## The question

Given a module that is **currently being maintained**, does its ownership
structure at time *T* predict that it stops receiving commits over the next *H*
periods?

Not "does it collapse". Not "does its survival mass fall". **Does anyone commit
to it.**

## The trap this design exists to avoid

An earlier version of this study used

    collapse(m, t) = survival_mass(m, t+H) < 0.2 × survival_mass(m, t)

That is a **decay artifact**. EIS survival is time-decayed blame mass: a module
that receives no new commits loses ~15.6% of its mass per period (observed period
ratio median 0.844, half-life ≈ 4.1 months). So *any* un-refilled module crosses
*any* fixed threshold on a fixed schedule — 106 of 137 "collapses" were pure
decay, and the headline "≈ 1 year of lead time" turned out to be

    log(0.2) / log(0.844) ≈ 9.5 periods

i.e. a restatement of the decay constant, not a property of bus factor.

**A thresholded decaying quantity cannot be an outcome.** The outcome here is
threshold-free:

    abandoned(m, t, H) = zero commits to m in [t, t+H)

and it is computed from raw `git log`, never from EIS.

## Design

| | |
|---|---|
| unit | (module *m*, anchor period *t*) |
| inclusion | *m* is **currently maintained**: ≥ 3 commits in [t−3, t) |
| predictors @ t | from history ≤ t only (see below) |
| outcome | abandoned over [t, t+H), H ∈ {6, 12, 18} |
| metric | ROC-AUC, sign-oriented so > 0.5 always means "predicts abandonment" |

The inclusion filter is what makes the question non-trivial: we ask **whether a
live module dies**, not whether a dead module is dead.

### Predictors

From `module_survival_by_author` at period *t* (point-in-time, scored at the
window boundary — not a HEAD read):

- `n_hold` — distinct authors holding surviving mass (**owner_count**)
- `hhi` — Σ share² (concentration)
- `top_share` — largest single share (bus-factor proxy)
- `survival` — total surviving mass (**level** — the confound to beat)

From raw git:

- `n_recent` — distinct authors who committed to *this module* in [t−3, t)
- `churn` — commits in [t−3, t) (the activity baseline; repowise / CodeScene family)

> `n_recent` must be per-module. Counting repo-wide active authors gives every
> module the same value at time *t* — a repo-activity proxy confounded with
> calendar time, not ownership. The first draft of this script had that bug and it
> produced a confident, wrong answer.

### Non-source modules are excluded, and that is load-bearing

`prettier` resolves ~470 "modules", most of them **test-fixture directories**
(`tests/new_expression`, `tests/refi`, …). A fixture is written once by one
person and never touched again — so it has `owner_count == 1` *and* is
guaranteed "abandoned". Leave those in and low owner_count predicts abandonment
at **AUC 0.711, module-clustered CI [0.664, 0.750], p = 0.0005**. Drop them and
the same repo gives **0.602, CI [0.328, 0.853], p = 0.46** — null.

The strongest ownership result in the corpus was directory granularity, not
ownership. This is the same class of error as the decay artifact: a number that
is real, robust, significant, and about something other than what it is named
after.

`--keep-non-source` reproduces the artifact on demand. The exclusion list mirrors
calibration ② of the structural-debt spec ("core is decided by intent").

### predictor ⊥ outcome

Predictors come from `eis timeline`. The outcome comes from `git log --numstat`.
Two independent pipelines, so no shared machinery can manufacture the
correlation. Same hinge as `../oss-gravity-map/analysis/survival-backtest`.

## Two statistics that do the real work

Units are (module × period), so they are massively non-independent — one module
contributes dozens of rows and its fate is one draw. A pooled AUC alone is not
evidence.

- **cluster bootstrap** — resample *modules* with replacement → 95% CI.
  If the CI includes 0.5, there is no finding.
- **module permutation** — swap whole modules' predictor trajectories between
  modules, keeping each module's outcome. This preserves both marginals and
  destroys only the pairing. `p = P(AUC_perm ≥ AUC_obs)`. If the association is
  really "some modules are just busier", this keeps it and p stays high.

## Running

```sh
go build -o /tmp/eis ./cmd/eis
EIS=/tmp/eis research/abandonment-backtest/run.sh react express esbuild ...
python3 research/abandonment-backtest/abandonment.py --pooled
```

Reuses the clones and curated configs in `../oss-gravity-map` — nothing is
re-cloned. `data/` is gitignored and fully regenerable.

Seeded (`SEED = 20260722`): the same inputs give the same CI and p. A calibration
run that cannot be reproduced to the digit is not a calibration run — that is
the whole reason this directory exists.

## Results

Recorded per run in the OrbitLens calibration ledger (`docs/calibration/runs/`),
not here, so that claim status and evidence stay in one place.

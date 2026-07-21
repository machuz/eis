# Collapse backtest — does EIS predict maintenance collapse?

A predictive-validity study, run on 8 OSS repos, with an honest (mostly negative)
conclusion. The question: does an EIS-observable signal — structural
**non-conformance**, survival-weighted **concentration** (bus-factor), or
**churn** — predict that a module *collapses* (dies / gets abandoned) after the
person who held it leaves? And does any EIS signal beat plain churn (the
repowise / CodeScene lineage)?

**Short answer: no.** Across three outcome definitions and a placebo control,
no signal predicts *departure-caused* collapse. The value of EIS is in
**observing** realized durability, not in **predicting** collapse.

## What was tested

- **Corpus**: 8 OSS repos (express, redis, eslint, prettier, fastapi, nest,
  vite, esbuild), full-history monthly EIS timeline + analyze.
- **Substrate**: per-`(module, author, month)` survival, from
  `eis timeline --format json` emitting `module_survival_by_author` (#366) +
  per-module churn/cochange/ownership from `eis analyze`.
- **Design**: placebo-controlled cohorts — `treat` (anchor = an author's
  departure) vs `control` (anchor = a mid-tenure month, no departure). A signal
  is only *departure-driven* if its correlation is stronger in treat than control.

## Findings (each saved with full numbers in the commit history / notes)

1. **Conformance does not predict collapse.** Both proxies — co-change
   path-dispersion and token-entropy naturalness — are null or wrong-signed,
   across a survival-mass outcome and an abandonment outcome. The "structural
   inconsistency → collapse" moat hypothesis is **rejected**.
2. **Concentration (bus-factor) is mechanically confounded.** It correlates with
   "collapse" about equally in the placebo (no departure) as in real departures
   (departure effect ≈ +0.04). High concentration makes total survival drop more
   *by arithmetic*, not because the maintainer left.
3. **Churn is the only placebo-passing signal — and it's repowise's, not EIS's.**
   For the abandonment outcome, churn's departure effect is the largest (~+0.33),
   but weak, clustered, and modest.
4. **Defect prediction is a volume tautology** (`defect_test.py`). fix-commit
   count is ~0.90 explained by raw commit volume; controlling for volume, churn's
   contribution vanishes; and defect *rate* (fixes per change) is unpredictable
   (~uniform ≈ 21%). "Predict which module gets bug-fixes" ≈ "predict which module
   has many commits."
5. **Successor-absence ranks at-risk modules, but isn't a causal predictor.** A
   module whose holder has no active co-holder is abandoned ~2.5× more after
   departure — but successor-existence predicts continued activity almost as much
   during normal tenure, so it's mostly "lively multi-person modules stay lively."
   Useful as a **feature** (rank at-risk modules on departure), not as a
   prediction moat.
6. **Substance = survival × catalysis, not robust-survival × catalysis.**
   `robust_survival × catalysis` zeros out stable foundational code (e.g. a
   10k-commit maintainer whose core is now stable scores 0). Catalysis, not the
   "robust" qualifier, is what filters neglect-survival. See `substance.py`.

**Meta-conclusion.** Predicting maintenance collapse from git structure appears
near-unpredictable for the whole category (EIS confounded; repowise
volume-tautological). EIS should sell *observation of realized durability*
(survival × catalysis), not *prediction*. The likely reason the effect is absent
on OSS: community successor availability suppresses bus-factor collapse; the
phenomenon may only manifest in private codebases with no successor.

## How to run

Requires an `eis` built from `main` (>= #366, which emits
`module_survival_by_author` in `timeline --format json`):

```sh
go build -o eis ./cmd/eis           # from the eis repo root
export EIS=$PWD/eis                  # point run.sh at it (or put eis on PATH)

cd research/collapse-backtest
./run.sh                            # clones the 8 repos, writes data/*.json  (~20 min, ~1 GB)
python3 naturalness.py              # per-module token-entropy conformance
python3 module_activity.py          # per-module monthly commit activity (for abandonment)
python3 backtest.py                 # the placebo-controlled head-to-head
python3 defect_test.py              # is repowise-style defect prediction a volume tautology?
python3 substance.py                # survival x catalysis contributor score
```

`SPEC.md` is the original design spec (outcome definitions, cohorts, baselines).
`data/` and the OSS clones are gitignored (regenerable via `run.sh`).

## Files

- `run.sh` — generate the observation substrate (clone → analyze + timeline).
- `backtest.py` — departure detection, three outcomes, placebo cohorts, head-to-head.
- `naturalness.py` — token-entropy naturalness (conformance proxy).
- `module_activity.py` — per-`(module, month)` commit counts (abandonment outcome).
- `defect_test.py` — churn vs volume decomposition of defect prediction.
- `substance.py` — `survival × catalysis` per contributor, and why not `robust`.

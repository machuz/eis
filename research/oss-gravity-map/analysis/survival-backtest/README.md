# survival×catalysis — predictive-validity backtest

Establishes external predictive validity for EIS **gravity** (survival × catalysis),
directly comparable to code-health tools that publish a defect-prediction ROC-AUC
(e.g. repowise's ~0.74), and demonstrates the sharpened thesis **"survival ≠ gravity"**:
the *catalysis* gate (others building on your surviving code) carries forward-predictive
information that raw survival, code age, and activity/churn do not.

## The question

Does a contribution's **gravity at time T**, computed from history ≤ T only, predict
whether that code actually **lasts** over `(T, HEAD]`?

We test three claims:

1. **Own-axis validity.** gravity@T predicts survival-to-HEAD with ROC-AUC ≥ ~0.7.
2. **survival ≠ gravity.** Adding the catalysis gate (others-contested build-on) to raw
   survival gives a positive incremental ΔAUC — the conjunction beats the single axis.
3. **activity ≠ durability.** Churn / activity (the CodeScene/repowise family's substrate)
   is a *weaker* durability predictor than gravity — the game-resistance claim, in data.

## Design (leakage-free, decoupled predictor and outcome)

The credibility hinge: the **predictor** and the **outcome** are computed by two
*independent* pipelines, so no shared machinery can manufacture the correlation.

### Predictor @ T — from EIS (`eis timeline`)

`eis timeline --format json` emits, per period ending at `T`, a point-in-time snapshot:
`module_survival_by_author : {module : {author : decayed_surviving_mass}}`, scored at the
window's boundary commit using only commits ≤ `window.End` (see `pkg/timeline/run.go`:
"resolves its own boundary commit at window.End … scores against window.End"). This is a
true as-of-T computation — not a HEAD read.

From that snapshot, at the chosen anchor `T`:

- **module gravity** = Σ over authors of the module's surviving mass.
- **module catalysis / contest** = number of distinct authors carrying surviving mass in
  the module (and the Hill-number spread) — "how many build on it".
- **module concentration** = top author's share (bus-factor proxy).
- **author gravity** = Σ over modules of that author's surviving mass; **breadth** = spread.

Baselines (also as-of-T, from `git log`/`git blame` ≤ T):

- **churn** — commits (and changed LOC) touching the module in a trailing window ≤ T
  (the activity substrate; stands in for the repowise/CodeScene family here — the
  complexity/AST proxy is a **phase-2** add, see below).
- **raw survival** — module/author surviving mass at T *without* the catalysis gate.
- **age** — days since the module/author first appeared.

### Outcome @ HEAD — from raw git (independent of EIS)

Cohort survival of the ≤ T code, git-of-theseus style:

- `sha_T` = last commit with author-date ≤ T.
- **cohort(m)** = lines in module `m`'s files at `sha_T` (blame `sha_T`).
- **survivors(m)** = lines in `m` at HEAD whose blame commit has author-date ≤ T (i.e. of
  today's code, the part that originated ≤ T and was never overwritten).
- **survival_ratio(m)** = survivors(m) / cohort(m)  ∈ [0, 1].

Person-level is the same with "lines authored by `a`" (blame author = last toucher, so a
line only counts as `a`'s survivor if it reached HEAD unmodified — the correct notion of
"a's code lasted").

This uses only `git`, never EIS's survival decay, so predictor⊥outcome.

## Metric

- **ROC-AUC** on a binarized outcome (a module/author "fails" when survival_ratio falls in
  the bottom band, e.g. < 0.5 of its cohort persists) — the number that lines up against
  repowise's published 0.74.
- **Spearman ρ** on the continuous survival_ratio — threshold-free robustness check.
- **Incremental ΔAUC** from adding catalysis to raw survival (claim 2), reported per repo
  and pooled across the 30-repo corpus, with a DeLong / bootstrap CI.

## Corpus

The 30-repo OSS gravity-map corpus (`../../configs/*.yaml`, clones in
`../../data/repos/`) — react, kubernetes, rails, envoy, clickhouse, rust, git, … — diverse
in language, age, and domain. Each repo already carries an `architecture_patterns` config
so module resolution and the design axis are meaningful.

## Anchors / horizon

Anchor `T` = a period end roughly `Δ` before HEAD (default Δ = 3y). Multiple anchors
(rolling origin) can be added; phase 1 uses one anchor per repo to validate the pipeline,
then scales.

## Phasing

- **Phase 1 (this dir):** module + person level; predictors = gravity, raw survival, churn,
  age, concentration. No AST. Proves claims 1–3 with churn as the activity baseline.
- **Phase 2 (deferred):** throwaway complexity/AST proxy (McCabe / nesting / LCOM) via
  tree-sitter, scoped to the languages repowise supports, for a literal head-to-head
  against a Code-Health-style score. Deferred because multi-language AST is the heavy part
  and is not needed to establish claims 1–3.

## Files

- `extract_predictors.py` — run `eis timeline`, parse, emit per-module & per-author
  predictors @ T (+ churn/age baselines).
- `outcome_cohort_survival.py` — raw-git cohort survival ≤T → HEAD, per module & author.
- `backtest.py` — join, binarize, ROC-AUC + Spearman + incremental ΔAUC, per-repo & pooled.
- `run.sh` — orchestrate one repo (or the corpus) end to end.

Not shipped in the `eis` CLI — this is offline research (telescope stays the telescope).

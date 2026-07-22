# person backtest — does surviving stock predict future durable output?

The person-axis counterpart to `../abandonment-backtest`. The original pilot
(`scratchpad/backtest2.py`, 2026-07-09) is lost; its headline survives in prose:

> survival_T → future survival = **+0.62**, activity_T → future survival = **+0.24**.
> Survival predicts 2.6× better.

That number is the most-used sales copy in the deck. It is also inflated by
construction, and this directory exists to say by how much.

## The trap

EIS survival is time-decayed blame mass, half-life ≈ 4.1 months. The pilot's
outcome was future survival **stock** at T+3 months. About 60% of the mass in
the outcome is *the same mass* as in the predictor, still decaying. A quantity
correlating with its own decayed self is not a prediction.

The module axis had this disease in a different costume — a thresholded decaying
outcome, where "≈1 year of lead time" turned out to be `log(0.2)/log(0.844)`.
The fix there was to make the outcome threshold-free and raw-git sourced. Here
the same move is available without leaving the mass panel:

    outcome = NEW mass the author adds during (t, t+H]

**Decay can only remove mass. Mass that appears was authored.** So the outcome
shares nothing with the predictor's stock — the mechanical channel is closed by
construction rather than argued about in a limitations section.

## Design

| | |
|---|---|
| unit | (author *a*, anchor period *t*) |
| inclusion | *a* is currently active: added new mass in [t−3, t) |
| predictors @t | `stock` (surviving mass), `flow` (new mass in [t−3, t)), `n_modules`, `top_share` |
| outcome | `future_flow` = new mass added over (t, t+H], H ∈ {3, 6, 12} |
| metric | ROC-AUC on `productive` = future_flow > 0, sign-oriented |

## The quadrant claim, made rigorous

The pilot's real content was never the correlation — it was the quadrant:
"busy but not surviving" produced ~1/2.6 of "quiet but surviving". That is a
claim about **stock discriminating at fixed activity**. So the test that matters
is the stratified one (`--stratify`): inside each `flow` tercile, does `stock`
still separate? A headline AUC that evaporates within strata is the confound
talking, not the thesis.

## Raw activity (the 6th CSV column) — what makes C-02 testable

`flow` is mass added, so it is durability-filtered and **cannot** express "busy
but their code does not survive": that person has `flow ≈ 0` and reads as quiet.
Raw commit counts can, so the loader takes an optional 6th column and
`--activity commits` puts it in the stratified and quadrant views.

Joining raw git to the mass panel is possible because `resolveAuthor` falls back
to `fnv1a64(lower(trim(author))) | (1<<62)` for unresolved authors — a pure
function of the author string, reproducible outside the service — and because the
author string EIS emits is the plain commit name. On react that alone joins 930
of 1052 panel keys; the persistent identity cache covers a few more.

**Rows whose key cannot be joined must be dropped, never passed through as
commits = 0.** An unjoined author would be mislabelled "quiet", and the unjoined
set is not random: identity resolution sweeps only the first
`ListRepoCommitAuthors` pages (~300 commits), so what is missing is *recent,
high-volume* contributors — precisely the "busy" population the claim is about.
On react that is 88.6% of panel keys but only 79.2% of commits.

## Older note: what the mass-only version cannot test

`flow` is `max(0, mass_t − mass_{t−1})` — **surviving** mass added, not commits.
So it is already durability-filtered, and that breaks the one comparison the
pilot's headline was about.

"Busy but their code doesn't survive" is exactly the person whose commits are
overwritten fast. In this measure that person has `flow ≈ 0`, so they land in the
quiet-and-not-surviving cell, not the busy-and-not-surviving one. **The quadrant
computed here is not the pilot's quadrant**, and its ratio should not be reported
as a correction of the pilot's 1/2.6 — the two measure different populations.

Testing "activity ≠ durability" properly needs raw per-author commit counts
aligned to the same identities as the mass panel. The observation tables key
authors by `github_user_id`; raw git keys them by email, and the ingest path
drops `author.login`, so there is no join. That is the blocker, and it is a
data-model gap rather than a study-design one.

What this directory *can* answer is narrower and still worth having: does
accumulated surviving stock predict future durable output, separably from
recent output.

## Known bias, and which way it cuts

`flow` is `max(0, mass_t − mass_{t−1})`. Mass also falls through decay and
through other people overwriting your lines, so this **floors** new production —
and floors it hardest for authors carrying a large decaying stock. That biases
*against* the stock → output claim. The direction is therefore reportable; the
magnitude is not.

Gaps in the panel are treated as *no observation*, not as zero mass, so a
missing period cannot manufacture a drop.

## Guards

Author-clustered bootstrap CI, author permutation test, and — the lesson that
cost a merged PR on 2026-07-22 — **cluster and event counts printed next to
every CI**. A null with a single-digit cluster count means "cannot tell", not
"no effect", and must not be written up as a refutation.

## Input

A CSV, no header, one row per (author, period):

    YYYY-MM,author_key,n_modules,total_mass,max_module_mass

`author_key` is expected to be a pseudonym — the exporter should hash the real
identifier so it never leaves the deployment. Nothing here needs the identity;
it only needs a stable grouping key.

    python3 person.py --csv panel.csv --label react --stratify

## Results

Recorded in the OrbitLens calibration ledger (`docs/calibration/runs/`), not
here, so claim status and evidence stay in one place.

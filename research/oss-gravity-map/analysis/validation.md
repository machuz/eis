# Does EIS reflect reality? — an independent, time-aware validation

EIS assigns each engineer a *gravity*: how much durable, load-bearing code others
build on top of. A natural worry is that this is circular — that gravity merely
restates whatever EIS already measures. So the question worth asking is simple.
Does gravity line up with who the world *independently* calls an architect, and
does it line up at the *right time*?

This note answers both, against ground-truth that EIS never sees.

**Headline.** Across 25 repositories, gravity ranks a declared architect above a
non-architect with mean **AUC = 0.928** (95% CI 0.887–0.962; null = 0.500). Its
top-20 recovers **69%** of independently-declared architects (95% CI 60–77%) — a
**29×** enrichment over a random top-20 (4.2 expected hits, 122 observed). And the
architects whose static gravity has since faded are caught, correctly, at high
gravity in their own era. The metric discriminates, and it discriminates in time.

## Method

**Ground-truth.** For 25 OSS repositories we collected, from GitHub and from each
project's own governance files, a list of `architect_candidates` — the logins that
appear as code owners, release authors, and sustained top contributors. This list
is assembled with no reference to any EIS output (`scripts/fetch-ground-truth.sh`,
`data/ground-truth/*.json`).

**Bridge.** `architect_candidates` are GitHub logins; EIS reports author display
names. We bridge login → name through `data/alias-map.json` (a per-repo authoritative
map) with a fallback to the cached GitHub user profile. Candidates that cannot be
bridged are dropped from the denominator rather than counted as misses — a
conservative choice that can only lower the reported recall.

**Recall@20.** For each repo we ask: of the bridged architect candidates, how many
land in EIS's static gravity top-20?

**Temporal check.** Static gravity is a single number over all history. But code
gets superseded — a founder's gravity *should* decay once their code is rewritten.
So for repos where we also have a per-period timeline, we ask a sharper question:
- For candidates EIS *catches* — does their timeline gravity peak in their actual era?
- For candidates EIS *misses* — did they ever peak high in any period, or were they
  consistently low?

A low static gravity is only a real miss if the person never load-bore durable code
in *any* era. If their timeline peaked and then faded, EIS captured them correctly
and the static low is a time-faithful fade, not an error.

**Statistics.** Three tests, no library dependencies (`scripts/validate-stats.py`,
stdlib only, seed = 42):
- *Ranking* — AUC (= P[architect outranks non-architect], ties credited 0.5), per
  repo and as a mean over repos, with a one-sided normal-approx Mann-Whitney p
  against the 0.5 null and a cluster bootstrap (resampling whole repos, B = 5000)
  for the 95% CI. AUC uses the **full** ranking, so it does not depend on the
  arbitrary "top-20" cutoff.
- *Top-list enrichment* — an exact hypergeometric test per repo: against a
  population of N ranked engineers with K architects, how surprising is it to draw
  `hits` architects in a top-20 sample? Reported as fold-enrichment over the chance
  expectation and a one-sided P[X ≥ observed].
- *Temporal separation* — a one-sided Mann-Whitney comparing the per-period gravity
  peaks of caught architects against missed ones.

Everything below is reproduced by `scripts/validate-temporal.py` (tables) and
`scripts/validate-stats.py` (the tests above).

## Result 0 — the metric discriminates (AUC = 0.93)

Treating it as a ranking problem — does a higher gravity mean a higher chance of
being a declared architect? — sidesteps the arbitrariness of any cutoff.

| statistic | value | null |
|---|---|---|
| mean AUC over 25 repos | **0.928** (95% CI 0.887–0.962) | 0.500 |
| micro recall@20 | **69%** (95% CI 60–77%) | — |
| top-20 fold-enrichment | **29.1×** (122 obs vs 4.2 expected) | 1.0× |

Per-repo AUC ranges from 0.997 (arrow) down to 0.59 (fastapi); 23 of 25 repos sit
above 0.88. The hypergeometric enrichment is individually significant at p < 1e-3
in **21 of 25** repos; the four exceptions are all small-K cases (envoy K=4 at
p=1.2e-3, esbuild, fastapi, nest K=2) where a top-20 sample simply cannot resolve a
two- to four-name ground-truth. No test is run that the data cannot support.

## Result 1 — recall@20 = 69%

Across 25 repositories, EIS's static gravity top-20 recovers **122 of 177**
(micro-averaged **69%**) of the independently-declared architect candidates.

| tier | repos |
|---|---|
| 100% | express, prometheus |
| 83–89% | loki, superset, polars, spring-boot, argo-cd, duckdb, prettier, vite |
| 67–78% | arrow, terraform, ClickHouse, eslint, swc |
| 33–62% | grafana, phoenix, react, redis, envoy, rust, kubernetes |
| ≤29% | esbuild, fastapi, nest |

An earlier pass that used raw `CODEOWNERS` mentions as ground-truth scored ~49%.
The gap is instructive: `CODEOWNERS` over-includes reviewers and small-directory
owners who never wrote foundational code. Against a precise architect list, recall
rises to 69% — EIS is *more discriminating* than paper ownership, not less accurate.

## Result 2 — EIS credits architects in their era

This is the part that matters most. The candidates whose static gravity is now
*low* are not missed — their timeline peaks them, correctly, in the era their code
was load-bearing, and fades them once it was superseded.

| repo | engineer | per-period peak | static (now) |
|---|---|---|---|
| grafana | **Torkel Ödegaard** (Grafana founder) | **94** @2017 (Architect) | 1.6 |
| prometheus | **Matt T. Proud** (co-founder) | 90 @2012 (Architect) | 0.7 |
| react | Sebastian Markbåge | 89 @2023 (Architect) | 9.9 |
| prometheus | Fabian Reinartz | 88 @2016 (Architect) | 1.3 |
| react | Andrew Clark | 77 @2021 (Architect) | 1.9 |
| grafana | Ryan McKinley | 73 @2024 (Architect) | 14.9 |
| prometheus | Julius Volz (co-founder) | 60 @2013 (Architect) | 5.9 |
| prometheus | Ganesh Vernekar | 49 @2020 (Cleaner) | 1.5 |

And in the language repos, the same pattern holds for the people who literally
defined the language: Martin Odersky peaks at gravity 99 in Scala's 2005 founding
era; Paul Phillips, the prolific Scala 2 compiler architect who later left, peaks
at 96 in 2011 and is at 1.8 today; Nicolas Stucki peaks at 100 in Scala 3's 2019.

Gravity is not "who wrote the most" — it is "whose code is load-bearing *now*."
The timeline shows the dual: "whose code *was* load-bearing then." A founder's
gravity decaying from 94 to 1.6 is not the metric failing to see them; it is the
metric seeing, accurately, that their code has since been rebuilt by others.

## Result 3 — the misses are consistent, not lossy

For the architect candidates EIS's static top-20 *does* miss, the timeline tells
the same story it tells statically: they were low in every era too.

| repo | engineer | static | per-period peak |
|---|---|---|---|
| react | Sophie Alpert | 0.1 | 32 @2015 |
| react | Dominic Gannaway | 0.2 | 10 @2018 |
| grafana | Marcus Efraimsson | 0.5 | 8 @2018 |
| grafana | Daniel Lee | 0.1 | 2 @2017 |
| envoy | Mike Schore | 0.3 | 0 @2016 |

The separation is statistical, not anecdotal. Across the repos with a timeline,
the caught architects' per-period peaks (n=18, median 32) sit well above the missed
ones' (n=8, median 2): one-sided Mann-Whitney p = 1.6e-3, AUC = 0.87 — a caught
architect's era-peak outranks a missed one's 87% of the time. The metric's
verdict is the same whether read statically or period-by-period.

None of the missed candidates ever peaked above 32. These are real and valued contributors — but their
contribution was management, review, build/CI infrastructure, or code that was
quickly superseded, not durable load-bearing source. A worked example: Lizan Zhou,
a long-standing Envoy maintainer with 452 commits, spends ~80% of them in
`ci/`, `bazel/`, `.azure-pipelines/`, `.bazelrc`; his surviving source footprint
is 66 lines across the sampled HTTP core, against Matt Klein's 646 in the same
files. EIS ranks him low on *code* gravity because that is what he is — low on
code gravity, high on build maintainership.

The diagnosis is consistent across both axes: EIS's discrimination is stable over
time. It does not flicker. People it credits are caught at their peak; people it
ranks low were low across all eras.

## The honest boundary

EIS measures **durable code structural gravity**. By construction it does not
credit non-code contribution — code review, release management, CI and build
maintenance, project leadership — as architecture. This is a scope statement, not
a defect: the lens is built to see one thing clearly rather than everything
faintly. The 31% of "misses" against an architect list are dominated by exactly
these non-code roles, plus bots and superseded code. Where an organization wants
those roles surfaced, that belongs to a separate signal, not to gravity.

One operational caveat worth stating: gravity's Design and Catalysis components
depend on the repo's `architecture_patterns` configuration. With no config, EIS
cannot locate the architectural surface and gravity collapses toward zero. The
config is the part of the telescope that must be aimed; an unaimed telescope sees
nothing, and that is observer error, not absence of signal.

## Reproducing

```bash
scripts/fetch-ground-truth.sh        # refresh data/ground-truth/*.json
python3 scripts/validate-temporal.py # recall@20 + temporal credit + miss tables
python3 scripts/validate-stats.py    # AUC, bootstrap CI, hypergeometric, Mann-Whitney

# the temporal section additionally needs per-period timelines, which are large
# (8–23 MB each) and therefore git-ignored — regenerate the ones it reads with:
for r in react grafana prometheus envoy scala scala3; do
  eis timeline --format json --span 1y --periods 0 --fast-log --sample 500 \
    --config configs/$r.yaml data/repos/$r > data/results/$r-timeline.json
done
```

The static figures (AUC, recall@20, enrichment) read only the committed
`data/results/*.json` and `data/ground-truth/*.json`; with the seeded, stdlib-only
scripts they reproduce exactly. The temporal figures additionally read the
regenerated timelines above.

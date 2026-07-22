#!/usr/bin/env python3
"""Does ownership structure predict that a *currently maintained* module stops
being maintained?

This is a re-implementation. The original (scratchpad/module_abandonment.py,
2026-07-12) is lost; only its numbers survive in prose. See
orbitlens docs/calibration/claims/C-08.

## The trap this design exists to avoid

An earlier version used outcome = "survival mass drops below 20%". That is a
*decay artifact*: EIS survival is time-decayed blame mass, so an un-refilled
module loses ~15.6%/period (observed period ratio median 0.844, half-life
~4.1 months) and crosses any fixed threshold on a schedule. 106 of 137
"collapses" were pure decay, and the celebrated "~1 year lead time" was just
log(0.2)/log(0.844) ≈ 9.5 periods. Thresholded survival cannot be an outcome.

So the outcome here is **threshold-free and decay-independent**:

    abandoned(m, t, H)  =  zero commits to m in [t, t+H)

Commits are counted from raw git (module_activity.py), never from EIS survival.
A module cannot become "abandoned" by sitting still and decaying — it has to
actually stop receiving commits.

## Design

unit          (module m, anchor period t)
inclusion     m is *currently maintained*: >= MIN_BEFORE commits in [t-W, t)
              (asking "does a live module die" — not "is a dead module dead")
predictors@t  computed from history <= t only:
                n_hold    distinct authors holding surviving mass  (owner_count)
                n_recent  distinct authors who committed in [t-W, t)
                hhi       sum of squared mass shares   (concentration)
                top_share largest single mass share    (bus-factor proxy)
                survival  total surviving mass         (level; the confound)
                churn     commits in [t-W, t)          (activity baseline)
outcome       abandoned over the next H periods (H = 6, 12, 18)
metric        ROC-AUC, direction-signed so >0.5 always means "predicts abandonment"

## Why the two extra statistics matter

Units are (module x period), so they are massively non-independent: one module
contributes many rows and its fate is one draw. Two guards:

  cluster bootstrap  resample *modules* with replacement -> 95% CI. A pooled
                     AUC whose module-clustered CI includes 0.5 is not evidence.
  module permutation swap whole modules' predictor trajectories between modules,
                     keeping each module's outcome. Breaks the module<->predictor
                     link while preserving both marginals. p = P(AUC_perm >= obs).
"""

from __future__ import annotations

import argparse
import glob
import json
import os
import random
from collections import defaultdict

DATA = os.environ.get(
    "EIS_BT_DATA", os.path.join(os.path.dirname(os.path.abspath(__file__)), "data"))

W = 3               # trailing window (periods) defining "currently maintained"
MIN_BEFORE = 3      # commits in the trailing window to count as maintained
HORIZONS = (6, 12, 18)
N_BOOT = 2000
N_PERM = 2000
SEED = 20260722     # fixed: these scripts must be re-runnable to the same numbers

# Modules that are not maintained source. Excluding these is not cosmetic — it
# is the difference between a finding and a tautology.
#
# prettier resolves ~470 "modules", most of them test-fixture directories
# (tests/new_expression, tests/refi, ...). A fixture is written once by one
# person and never touched again: owner_count == 1 AND guaranteed "abandoned".
# Leaving them in makes low owner_count predict abandonment at AUC 0.711,
# p=0.0005 — an artifact of directory granularity, not of ownership.
#
# Same rule as calibration (2) of the structural-debt spec: examples / docs /
# website are excluded by default, and an explicit architecture declaration is
# what puts a directory back in ("core is decided by intent").
NON_SOURCE = frozenset((
    "test", "tests", "spec", "specs", "__tests__", "fixture", "fixtures",
    "testdata", "e2e", "integration", "benchmark", "benchmarks", "bench",
    "doc", "docs", "website", "site", "example", "examples", "samples",
    "changelog", "changelog_unreleased", ".github", ".circleci",
    "scripts", "tools", "vendor", "third_party", "node_modules", "fuzz",
))


def is_source_module(module):
    return not any(p.lower() in NON_SOURCE for p in module.split("/"))


# predictor -> does a LOW value predict abandonment?
#   n_hold / n_recent / survival / churn : fewer/less -> more likely abandoned
#   hhi / top_share                      : more concentrated -> more likely abandoned
PREDICTORS = {
    "n_hold": "low",
    "n_recent": "low",
    "hhi": "high",
    "top_share": "high",
    "survival": "low",
    "churn": "low",
}


# ---------------------------------------------------------------- loading


def load_stream(path):
    """eis emits a stream of concatenated JSON docs, one per domain."""
    dec = json.JSONDecoder()
    s = open(path).read()
    i, out = 0, []
    while i < len(s):
        while i < len(s) and s[i] in " \n\t\r":
            i += 1
        if i >= len(s):
            break
        o, j = dec.raw_decode(s, i)
        out.append(o)
        i = j
    return out


def midx(label):
    y, m = label.split("-")
    return int(y) * 12 + int(m) - 1


def wsum(months, lo, hi):
    return sum(c for mo, c in months.items() if lo <= midx(mo) < hi)


# ---------------------------------------------------------------- panel


def build_panel_from_saas(repo, horizon, keep_non_source=False):
    """Same units, but predictors read from a SaaS-exported panel.

    On a 13-year monorepo, `eis timeline --span 1m` is not tractable locally
    (see README). A deployment has already computed that panel and stored it,
    so panel_csv.py exports it to <repo>.panel.json and we read the predictors
    from there instead of recomputing them.

    The outcome is unchanged: raw `git log` via activity.json. Predictor and
    outcome still come from two independent pipelines — swapping the predictor
    *source* does not weaken that, it only removes the local recompute.
    """
    panel_path = os.path.join(DATA, f"{repo}.panel.json")
    act_path = os.path.join(DATA, f"{repo}.activity.json")
    if not (os.path.exists(panel_path) and os.path.exists(act_path)):
        return []
    panel = json.load(open(panel_path))
    act = json.load(open(act_path))
    if not panel or not act:
        return []
    who_path = os.path.join(DATA, f"{repo}.committers.json")
    who = json.load(open(who_path)) if os.path.exists(who_path) else {}
    last_idx = max(midx(mo) for mm in act.values() for mo in mm)

    units = []
    for r in panel:
        module = r["module"]
        if module == "." or (not keep_non_source and not is_source_module(module)):
            continue
        t = midx(r["month"])
        if t + horizon > last_idx + 1:
            continue
        months = act.get(module, {})
        before = wsum(months, t - W, t)
        if before < MIN_BEFORE:
            continue
        after = wsum(months, t, t + horizon)

        mod_who = who.get(module, {})
        recent = {a for k in range(t - W, t)
                  for a in mod_who.get(f"{k // 12:04d}-{k % 12 + 1:02d}", ())}

        units.append({
            "repo": repo,
            "module": f"{repo}:{module}",
            "t": t,
            "n_hold": r["n_hold"],
            "n_recent": len(recent),
            "hhi": r["hhi"],
            "top_share": r["top_share"],
            "survival": r["survival"],
            "churn": before,
            "abandoned": 1 if after == 0 else 0,
        })
    return units


def build_panel(repo, horizon, keep_non_source=False):
    """-> list of unit dicts, one per (module, anchor period)."""
    tl_path = os.path.join(DATA, f"{repo}.timeline.json")
    act_path = os.path.join(DATA, f"{repo}.activity.json")
    if not (os.path.exists(tl_path) and os.path.exists(act_path)):
        return []

    act = json.load(open(act_path))          # module -> {"YYYY-MM": commits}
    if not act:
        return []
    who_path = os.path.join(DATA, f"{repo}.committers.json")
    who = json.load(open(who_path)) if os.path.exists(who_path) else {}
    all_months = {mo for mm in act.values() for mo in mm}
    last_idx = max(midx(mo) for mo in all_months)

    units = []
    for dom in load_stream(tl_path):
        periods = dom.get("periods") or []
        if not periods:
            continue

        for p in periods:
            label = p.get("label")
            if not label:
                continue
            t = midx(label)
            # need a full future window inside the observed range
            if t + horizon > last_idx + 1:
                continue

            msba = p.get("module_survival_by_author") or {}
            for module, by_author in msba.items():
                if module == "." or not by_author:
                    continue
                if not keep_non_source and not is_source_module(module):
                    continue
                months = act.get(module, {})
                before = wsum(months, t - W, t)
                if before < MIN_BEFORE:
                    continue                   # not currently maintained
                after = wsum(months, t, t + horizon)

                mass = {a: v for a, v in by_author.items() if v > 0}
                total = sum(mass.values())
                if total <= 0:
                    continue
                shares = [v / total for v in mass.values()]

                # per-module committers in the trailing window — "who could
                # maintain THIS module", not "who is active in the repo"
                mod_who = who.get(module, {})
                recent = {
                    a
                    for k in range(t - W, t)
                    for a in mod_who.get(f"{k // 12:04d}-{k % 12 + 1:02d}", ())
                }

                units.append({
                    "repo": repo,
                    "module": f"{repo}:{module}",
                    "t": t,
                    "n_hold": len(mass),
                    "n_recent": len(recent),
                    "hhi": sum(s * s for s in shares),
                    "top_share": max(shares),
                    "survival": total,
                    "churn": before,
                    "abandoned": 1 if after == 0 else 0,
                })
    return units


# ---------------------------------------------------------------- metrics


class Ranked:
    """Bucketed AUC over a fixed multiset of predictor values.

    Resampling (bootstrap, permutation) needs tens of thousands of AUCs over
    the *same* set of values. Sorting each time is O(n log n) per iteration and
    dominates everything. Instead bucket the distinct values once; each AUC is
    then a single O(units + buckets) pass:

        AUC = [ sum_i pos_i * cum_neg_below_i + 0.5 * sum_i pos_i * neg_i ]
              / (n_pos * n_neg)

    which is exactly the tie-corrected Mann-Whitney statistic.
    """

    def __init__(self, values):
        vals = sorted(set(values))
        self.index = {v: i for i, v in enumerate(vals)}
        self.k = len(vals)

    def auc(self, buckets, labels):
        pos_c = [0] * self.k
        neg_c = [0] * self.k
        npos = nneg = 0
        for b, y in zip(buckets, labels):
            if y:
                pos_c[b] += 1
                npos += 1
            else:
                neg_c[b] += 1
                nneg += 1
        if npos == 0 or nneg == 0:
            return None
        below = 0
        acc = 0.0
        for i in range(self.k):
            p = pos_c[i]
            if p:
                acc += p * below + 0.5 * p * neg_c[i]
            below += neg_c[i]
        return acc / (npos * nneg)


def _orient(key, a):
    """>0.5 always means 'predicts abandonment', whichever way the axis runs."""
    if a is None:
        return None
    return 1.0 - a if PREDICTORS[key] == "low" else a


def signed_auc(units, key):
    r = Ranked([u[key] for u in units])
    return _orient(key, r.auc([r.index[u[key]] for u in units],
                              [u["abandoned"] for u in units]))


def cluster_bootstrap(units, key, rng, n=N_BOOT, cluster="module"):
    r = Ranked([u[key] for u in units])
    by_c = defaultdict(list)
    for u in units:
        by_c[u[cluster]].append((r.index[u[key]], u["abandoned"]))
    keys = list(by_c)
    if len(keys) < 2:
        return None
    out = []
    for _ in range(n):
        buckets, labels = [], []
        for _ in range(len(keys)):
            for b, y in by_c[rng.choice(keys)]:
                buckets.append(b)
                labels.append(y)
        a = _orient(key, r.auc(buckets, labels))
        if a is not None:
            out.append(a)
    if len(out) < n // 2:
        return None
    out.sort()
    return out[int(0.025 * len(out))], out[int(0.975 * len(out))]


def module_permutation_p(units, key, rng, n=N_PERM):
    """Swap whole modules' predictor trajectories between modules.

    Preserves each module's outcome sequence and each module's predictor
    sequence; destroys only the pairing. If the association is really an
    artifact of "some modules are just busier", permuting keeps that and the
    p-value stays high.
    """
    r = Ranked([u[key] for u in units])
    by_mod = defaultdict(list)
    for u in units:
        by_mod[u["module"]].append(u)
    mods = list(by_mod)
    if len(mods) < 3:
        return None
    obs = signed_auc(units, key)
    if obs is None:
        return None

    donors = [[r.index[u[key]] for u in by_mod[m]] for m in mods]
    labels = [u["abandoned"] for m in mods for u in by_mod[m]]
    lens = [len(by_mod[m]) for m in mods]

    order = list(range(len(mods)))
    hits = 0
    for _ in range(n):
        rng.shuffle(order)
        buckets = []
        for mi, src_i in enumerate(order):
            src = donors[src_i]
            need = lens[mi]
            if len(src) >= need:
                buckets.extend(src[:need])
            else:                              # cycle if the donor is shorter
                buckets.extend(src[k % len(src)] for k in range(need))
        a = _orient(key, r.auc(buckets, labels))
        if a is not None and a >= obs:
            hits += 1
    return (hits + 1) / (n + 1)


# ---------------------------------------------------------------- report


def analyse(units, horizon, rng, label, full_stats=True, cluster="module"):
    n = len(units)
    pos = sum(u["abandoned"] for u in units)
    mods = len({u["module"] for u in units})
    print(f"\n### {label}  H={horizon}")
    print(f"units={n}  modules={mods}  abandoned={pos}  base_rate={pos / n:.3f}"
          if n else f"units=0")
    if n == 0 or pos == 0 or pos == n:
        print("  (degenerate — no discrimination possible)")
        return {}

    rows = {}
    print(f"  {'predictor':<10} {'AUC':>6}  {cluster+' 95% CI':>18}  {'perm p':>8}")
    for key in PREDICTORS:
        a = signed_auc(units, key)
        if a is None:
            continue
        ci = cluster_bootstrap(units, key, rng, cluster=cluster) if full_stats else None
        p = module_permutation_p(units, key, rng) if full_stats else None
        ci_s = f"[{ci[0]:.3f}, {ci[1]:.3f}]" if ci else "—"
        p_s = f"{p:.4f}" if p is not None else "—"
        flag = ""
        if ci and ci[0] > 0.5:
            flag = "  *"
        print(f"  {key:<10} {a:6.3f}  {ci_s:>18}  {p_s:>8}{flag}")
        rows[key] = {"auc": a, "ci": ci, "p": p}
    return rows


def leave_one_repo_out(pooled, horizon, repos):
    """Is the pooled signal one repo wearing a corpus as a disguise?

    With a handful of repos the repo-clustered bootstrap CI is too wide to say
    anything (5 clusters). Dropping each repo in turn is the blunt version of
    the same question and it is readable: if one row moves the AUC a lot, the
    "corpus" result is that repo.
    """
    print(f"\n### leave-one-repo-out  H={horizon}")
    base = {k: signed_auc(pooled, k) for k in PREDICTORS}
    hdr = "  ".join(f"{k:>9}" for k in PREDICTORS)
    print(f"  {'dropped':<12} {hdr}")
    print(f"  {'(none)':<12} " + "  ".join(
        f"{base[k]:9.3f}" if base[k] is not None else f"{'—':>9}" for k in PREDICTORS))
    for r in repos:
        rest = [u for u in pooled if u["repo"] != r]
        if not rest:
            continue
        vals = {k: signed_auc(rest, k) for k in PREDICTORS}
        n_ab = sum(u["abandoned"] for u in rest)
        if n_ab == 0 or n_ab == len(rest):
            continue
        print(f"  −{r:<11} " + "  ".join(
            f"{vals[k]:9.3f}" if vals[k] is not None else f"{'—':>9}" for k in PREDICTORS))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--repos", nargs="*", help="repo shortnames (default: all in data/)")
    ap.add_argument("--pooled", action="store_true", help="also analyse all repos pooled")
    ap.add_argument("--quick", action="store_true", help="skip bootstrap/permutation")
    ap.add_argument("--json-out", help="write results as JSON")
    ap.add_argument("--data", help="data dir (default $EIS_BT_DATA or ./data)")
    ap.add_argument("--loro", action="store_true", help="leave-one-repo-out table")
    ap.add_argument("--keep-non-source", action="store_true",
                    help="do NOT drop tests/docs/website modules (shows the artifact)")
    args = ap.parse_args()

    global DATA
    if args.data:
        DATA = args.data

    repos = args.repos or sorted({
        os.path.basename(p).split(".")[0]
        for p in (glob.glob(os.path.join(DATA, "*.timeline.json"))
                  + glob.glob(os.path.join(DATA, "*.panel.json")))
    })
    if not repos:
        print(f"no timeline data in {DATA} — run run.sh first")
        return 1

    rng = random.Random(SEED)
    results = {}
    for horizon in HORIZONS:
        pooled = []
        for repo in repos:
            if os.path.exists(os.path.join(DATA, f"{repo}.panel.json")):
                units = build_panel_from_saas(repo, horizon, args.keep_non_source)
            else:
                units = build_panel(repo, horizon, args.keep_non_source)
            pooled.extend(units)
            if len(repos) > 1 and not args.pooled:
                continue
            if len(repos) == 1:
                results[f"{repo}@H{horizon}"] = analyse(
                    units, horizon, rng, repo, full_stats=not args.quick)
        if len(repos) > 1:
            # Pooling adds a second level of non-independence: repos differ in
            # size, age and culture. A module-clustered CI does not absorb that,
            # so report the repo-clustered CI too — it is the conservative one.
            for cl in ("module", "repo"):
                results[f"pooled@H{horizon}/{cl}"] = analyse(
                    pooled, horizon, rng, f"POOLED({len(repos)} repos)",
                    full_stats=not args.quick, cluster=cl)
            if args.loro:
                leave_one_repo_out(pooled, horizon, repos)

    if args.json_out:
        def clean(o):
            if isinstance(o, dict):
                return {k: clean(v) for k, v in o.items()}
            if isinstance(o, tuple):
                return list(o)
            return o
        json.dump(clean(results), open(args.json_out, "w"), indent=2)
        print(f"\nwrote {args.json_out}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

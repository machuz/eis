#!/usr/bin/env python3
"""Does an author's surviving *stock* predict their future durable *output*
better than their recent activity does?

This is the person-axis counterpart to ../abandonment-backtest. The original
person-axis pilot (scratchpad/backtest2.py, 2026-07-09) is lost; its headline
was "survival_T -> future survival = +0.62 vs activity_T -> +0.24, so survival
predicts 2.6x better".

## The trap this design exists to avoid

That number is inflated by construction.

EIS survival is time-decayed blame mass with a ~4.1-month half-life. The pilot's
outcome was future survival *stock* at T+3 months. Roughly 60% of the mass in
the outcome is literally the same mass as in the predictor, still decaying. A
correlation between a quantity and its own decayed self is not a prediction —
and the ledger already flagged it ("survival の自己持続は時間減衰設計で一部構造的").

The module axis had the same disease in a different costume: a thresholded
decaying outcome. The fix there was to make the outcome threshold-free and
sourced from raw git. The fix here is the same move applied to the person axis:

    outcome = NEW mass the author adds during (t, t+H]

Decay can only *reduce* mass. Mass that appears is mass that was authored. So
the outcome shares nothing with the predictor's stock — the mechanical channel
is closed rather than argued about.

## Design

unit          (author a, anchor period t), a is currently active:
              added new mass in [t-W, t)
predictors@t  stock      total surviving mass at t          <- "survival"
              flow       new mass added in [t-W, t)         <- "activity"
              n_modules  modules where a holds mass         <- "breadth"
              top_share  a's largest single-module share
outcome       future_flow = new mass added over (t, t+H]
              binarized as `productive` = future_flow > 0
metric        ROC-AUC, sign-oriented so >0.5 means "predicts future output"

## The quadrant claim, made rigorous

The pilot's real content was not the correlation but the quadrant: "busy but not
surviving" contributed ~1/2.6 of "quiet but surviving". That is a claim about
stock discriminating *at fixed activity*. So the test that matters is
stratified: inside each flow tercile, does stock still separate? `--stratify`
reports exactly that. A headline AUC that evaporates within strata is the
confound talking.

## Known bias, and which way it cuts

`flow` is measured as max(0, mass_t - mass_{t-1}): mass also falls through decay
and through other people overwriting your lines, so this floors new production,
and it floors it *hardest* for authors carrying a large decaying stock. That
biases against the stock->output claim. Conservative, so the direction is
reportable; the magnitude is not.

## Guards

Same as the module axis, plus the lesson that cost a merged PR on 2026-07-22:

  author bootstrap    resample authors with replacement -> 95% CI
  author permutation  swap whole authors' predictor trajectories between authors
  reported counts     authors (clusters) and events are printed next to every CI.
                      A null with a single-digit cluster count is "cannot tell",
                      not "no effect", and must not be written up as a refutation.
"""

from __future__ import annotations

import argparse
import json
import os
import random
from collections import defaultdict

DATA = os.environ.get(
    "EIS_BT_DATA", os.path.join(os.path.dirname(os.path.abspath(__file__)), "data"))

W = 3
HORIZONS = (3, 6, 12)
N_BOOT = 2000
N_PERM = 2000
SEED = 20260722
MIN_FLOW = 1e-9

PREDICTORS = {
    "stock": "high",        # more surviving mass -> more future output?
    "flow": "high",         # more recent output  -> more future output?
    "n_modules": "high",
    "top_share": "low",
}


def midx(label):
    y, m = label.split("-")
    return int(y) * 12 + int(m) - 1


def load_person_panel(path):
    """CSV, no header: YYYY-MM,author_key,n_modules,total_mass,max_module_mass"""
    rows = defaultdict(dict)                     # author -> t -> record
    for line in open(path):
        line = line.strip()
        if not line:
            continue
        parts = line.split(",")
        if len(parts) != 5:
            continue
        ym, who, n_mod, tot, mx = parts
        try:
            t = midx(ym)
            tot_f, mx_f = float(tot), float(mx)
        except ValueError:
            continue
        rows[who][t] = {
            "stock": tot_f,
            "n_modules": int(n_mod),
            "top_share": (mx_f / tot_f) if tot_f > 0 else 1.0,
        }
    return rows


def build_units(panel, horizon):
    units = []
    for who, by_t in panel.items():
        ts = sorted(by_t)
        if len(ts) < 2:
            continue
        last_t = ts[-1]
        # new mass added in each period: decay can only subtract, so a rise is
        # authorship. Periods with no row are treated as no observation, not
        # as zero mass, so gaps do not manufacture a drop.
        delta = {}
        for i, t in enumerate(ts):
            if i == 0:
                delta[t] = by_t[t]["stock"]      # first appearance = all new
            else:
                delta[t] = max(0.0, by_t[t]["stock"] - by_t[ts[i - 1]]["stock"])

        for t in ts:
            if t + horizon > last_t:
                continue                          # need a full future window
            flow = sum(d for k, d in delta.items() if t - W <= k < t)
            if flow <= MIN_FLOW:
                continue                          # not currently active
            future = sum(d for k, d in delta.items() if t < k <= t + horizon)
            r = by_t[t]
            units.append({
                "author": who,
                "t": t,
                "stock": r["stock"],
                "flow": flow,
                "n_modules": r["n_modules"],
                "top_share": r["top_share"],
                "future_flow": future,
                "productive": 1 if future > MIN_FLOW else 0,
            })
    return units


class Ranked:
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
    if a is None:
        return None
    return 1.0 - a if PREDICTORS[key] == "low" else a


def signed_auc(units, key):
    r = Ranked([u[key] for u in units])
    return _orient(key, r.auc([r.index[u[key]] for u in units],
                              [u["productive"] for u in units]))


def cluster_bootstrap(units, key, rng, n=N_BOOT):
    r = Ranked([u[key] for u in units])
    by_a = defaultdict(list)
    for u in units:
        by_a[u["author"]].append((r.index[u[key]], u["productive"]))
    keys = list(by_a)
    if len(keys) < 2:
        return None
    out = []
    for _ in range(n):
        buckets, labels = [], []
        for _ in range(len(keys)):
            for b, y in by_a[rng.choice(keys)]:
                buckets.append(b)
                labels.append(y)
        a = _orient(key, r.auc(buckets, labels))
        if a is not None:
            out.append(a)
    if len(out) < n // 2:
        return None
    out.sort()
    return out[int(0.025 * len(out))], out[int(0.975 * len(out))]


def permutation_p(units, key, rng, n=N_PERM):
    r = Ranked([u[key] for u in units])
    by_a = defaultdict(list)
    for u in units:
        by_a[u["author"]].append(u)
    authors = list(by_a)
    if len(authors) < 3:
        return None
    obs = signed_auc(units, key)
    if obs is None:
        return None
    donors = [[r.index[u[key]] for u in by_a[a]] for a in authors]
    labels = [u["productive"] for a in authors for u in by_a[a]]
    lens = [len(by_a[a]) for a in authors]
    order = list(range(len(authors)))
    hits = 0
    for _ in range(n):
        rng.shuffle(order)
        buckets = []
        for ai, src_i in enumerate(order):
            src = donors[src_i]
            need = lens[ai]
            if len(src) >= need:
                buckets.extend(src[:need])
            else:
                buckets.extend(src[k % len(src)] for k in range(need))
        a = _orient(key, r.auc(buckets, labels))
        if a is not None and a >= obs:
            hits += 1
    return (hits + 1) / (n + 1)


def report(units, horizon, rng, label, full=True):
    n = len(units)
    pos = sum(u["productive"] for u in units)
    authors = len({u["author"] for u in units})
    print(f"\n### {label}  H={horizon}")
    if n == 0:
        print("  units=0 (degenerate)")
        return
    # counts first, deliberately: a CI is unreadable without its cluster count
    print(f"units={n}  authors(clusters)={authors}  productive={pos}  "
          f"base_rate={pos / n:.3f}")
    if pos == 0 or pos == n:
        print("  (degenerate — no discrimination possible)")
        return
    print(f"  {'predictor':<10} {'AUC':>6}  {'author 95% CI':>18}  {'perm p':>8}")
    for key in PREDICTORS:
        a = signed_auc(units, key)
        if a is None:
            continue
        ci = cluster_bootstrap(units, key, rng) if full else None
        p = permutation_p(units, key, rng) if full else None
        ci_s = f"[{ci[0]:.3f}, {ci[1]:.3f}]" if ci else "—"
        p_s = f"{p:.4f}" if p is not None else "—"
        star = "  *" if ci and ci[0] > 0.5 else ""
        print(f"  {key:<10} {a:6.3f}  {ci_s:>18}  {p_s:>8}{star}")


def stratified(units, horizon, rng, by="flow", target="stock"):
    """Does `target` still discriminate at fixed `by`? (the quadrant claim)"""
    vals = sorted(u[by] for u in units)
    if len(vals) < 9:
        return
    lo, hi = vals[len(vals) // 3], vals[2 * len(vals) // 3]
    bands = [("low", lambda v: v <= lo),
             ("mid", lambda v: lo < v <= hi),
             ("high", lambda v: v > hi)]
    print(f"\n### stratified by {by} — does {target} still separate?  H={horizon}")
    print(f"  {'band':<6} {'units':>6} {'authors':>8} {'events':>7} "
          f"{target+' AUC':>12}  {'author 95% CI':>18}")
    for name, pred in bands:
        sub = [u for u in units if pred(u[by])]
        ev = sum(u["productive"] for u in sub)
        au = len({u["author"] for u in sub})
        if not sub or ev == 0 or ev == len(sub):
            print(f"  {name:<6} {len(sub):>6} {au:>8} {ev:>7} {'degenerate':>12}")
            continue
        a = signed_auc(sub, target)
        ci = cluster_bootstrap(sub, target, rng)
        ci_s = f"[{ci[0]:.3f}, {ci[1]:.3f}]" if ci else "—"
        print(f"  {name:<6} {len(sub):>6} {au:>8} {ev:>7} {a:>12.3f}  {ci_s:>18}")


def quadrants(units, horizon, label="stock", other="flow"):
    """The sales copy's shape: mean future output in the 4 (stock x flow) cells.

    The pilot reported 54.1 vs 20.6 here and turned it into "1/2.6". Reproducing
    the *shape* lets that sentence be restated with a number that survives the
    decoupled outcome, instead of being quietly dropped.
    """
    if len(units) < 8:
        return
    sv = sorted(u[label] for u in units)
    fv = sorted(u[other] for u in units)
    s_med, f_med = sv[len(sv) // 2], fv[len(fv) // 2]
    cells = {}
    for u in units:
        key = ("stock+" if u[label] > s_med else "stock-",
               "flow+" if u[other] > f_med else "flow-")
        cells.setdefault(key, []).append(u["future_flow"])
    print(f"\n### quadrants (median split) — mean future output  H={horizon}")
    print(f"  {'':<8} {'flow+ (busy)':>16} {'flow- (quiet)':>16}")
    for s_lab in ("stock+", "stock-"):
        row = []
        for f_lab in ("flow+", "flow-"):
            vals = cells.get((s_lab, f_lab), [])
            row.append(f"{sum(vals) / len(vals):.2f} (n={len(vals)})" if vals else "—")
        print(f"  {s_lab:<8} {row[0]:>16} {row[1]:>16}")
    busy_dead = cells.get(("stock-", "flow+"), [])
    quiet_alive = cells.get(("stock+", "flow-"), [])
    if busy_dead and quiet_alive:
        a = sum(busy_dead) / len(busy_dead)
        b = sum(quiet_alive) / len(quiet_alive)
        if a > 0:
            print(f"  busy-but-not-surviving / quiet-but-surviving = {a / b:.2f}"
                  f"  (pilot reported 1/2.6 = 0.38)")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--csv", required=True, help="person panel CSV")
    ap.add_argument("--label", default="panel")
    ap.add_argument("--quick", action="store_true")
    ap.add_argument("--stratify", action="store_true",
                    help="also run the fixed-activity contrast (the quadrant claim)")
    args = ap.parse_args()

    panel = load_person_panel(args.csv)
    rng = random.Random(SEED)
    for horizon in HORIZONS:
        units = build_units(panel, horizon)
        report(units, horizon, rng, args.label, full=not args.quick)
        if args.stratify and units:
            stratified(units, horizon, rng)
            quadrants(units, horizon)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

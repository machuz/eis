#!/usr/bin/env python3
"""Join predictors @T with cohort-survival outcomes @HEAD and score predictive
validity. Pure Python (no numpy/sklearn): exact Spearman ρ and ROC-AUC (via the
Mann-Whitney U identity). Pools across repos.

The three claims (README.md):
  1. own-axis validity  — gravity@T predicts survival-to-HEAD (AUC / ρ well above 0.5)
  2. survival != gravity — the contest/catalysis feature beats raw_survival
                           (contest_mass = surviving mass others build on = survival×catalysis)
  3. activity != durability — churn is a weaker durability predictor than gravity

Usage:
  backtest.py --level module --min-cohort 30 --decay-threshold 0.7 \
      --pair repoA_mods.csv:repoA_out_mods.csv --pair repoB_mods.csv:repoB_out_mods.csv
  backtest.py --level author --min-cohort 50 --decay-threshold 0.7 --pair ...:...
"""
import argparse
import csv
import math
from collections import defaultdict

MODULE_FEATURES = ["raw_survival", "contest_mass", "spread", "author_count",
                   "churn_commits", "churn_loc", "age_days"]
AUTHOR_FEATURES = ["raw_gravity", "module_count", "top_module_share"]


def load_join(pred_csv, out_csv, key):
    outcome = {}
    with open(out_csv) as f:
        for r in csv.DictReader(f):
            if r["survival_ratio"] != "":
                outcome[r[key]] = (float(r["survival_ratio"]), int(r["cohort"]))
    rows = []
    with open(pred_csv) as f:
        for r in csv.DictReader(f):
            k = r[key]
            if k not in outcome:
                continue
            ratio, cohort = outcome[k]
            row = {"key": k, "survival_ratio": ratio, "cohort": cohort}
            for fld, v in r.items():
                if fld == key:
                    continue
                try:
                    row[fld] = float(v)
                except (ValueError, TypeError):
                    row[fld] = 0.0
            rows.append(row)
    return rows


def rank(values):
    """Fractional (average) ranks, for Spearman."""
    order = sorted(range(len(values)), key=lambda i: values[i])
    ranks = [0.0] * len(values)
    i = 0
    while i < len(order):
        j = i
        while j + 1 < len(order) and values[order[j + 1]] == values[order[i]]:
            j += 1
        avg = (i + j) / 2.0 + 1.0
        for k in range(i, j + 1):
            ranks[order[k]] = avg
        i = j + 1
    return ranks


def spearman(x, y):
    n = len(x)
    if n < 3:
        return float("nan"), float("nan")
    rx, ry = rank(x), rank(y)
    mx, my = sum(rx) / n, sum(ry) / n
    sxy = sum((a - mx) * (b - my) for a, b in zip(rx, ry))
    sxx = sum((a - mx) ** 2 for a in rx)
    syy = sum((b - my) ** 2 for b in ry)
    if sxx == 0 or syy == 0:
        return float("nan"), float("nan")
    rho = sxy / math.sqrt(sxx * syy)
    # approximate two-sided p via t distribution
    if abs(rho) >= 1.0:
        return rho, 0.0
    t = rho * math.sqrt((n - 2) / (1 - rho * rho))
    p = 2 * (1 - _t_cdf(abs(t), n - 2))
    return rho, p


def _t_cdf(t, df):
    # regularized incomplete beta via continued fraction (good enough for reporting)
    x = df / (df + t * t)
    ib = 0.5 * _betai(df / 2.0, 0.5, x)
    return 1 - ib if t > 0 else ib


def _betai(a, b, x):
    if x <= 0:
        return 0.0
    if x >= 1:
        return 1.0
    lbeta = math.lgamma(a) + math.lgamma(b) - math.lgamma(a + b)
    front = math.exp(math.log(x) * a + math.log(1 - x) * b - lbeta) / a
    # Lentz continued fraction
    f, c, d = 1.0, 1.0, 0.0
    for i in range(0, 200):
        m = i // 2
        if i == 0:
            num = 1.0
        elif i % 2 == 0:
            num = (m * (b - m) * x) / ((a + 2 * m - 1) * (a + 2 * m))
        else:
            num = -((a + m) * (a + b + m) * x) / ((a + 2 * m) * (a + 2 * m + 1))
        d = 1.0 + num * d
        if abs(d) < 1e-30:
            d = 1e-30
        d = 1.0 / d
        c = 1.0 + num / c
        if abs(c) < 1e-30:
            c = 1e-30
        f *= d * c
        if abs(1.0 - d * c) < 1e-8:
            break
    return front * (f - 1.0)


def auc(scores, labels):
    """ROC-AUC via Mann-Whitney U with tie handling. labels: 1=positive."""
    pos = [s for s, l in zip(scores, labels) if l == 1]
    neg = [s for s, l in zip(scores, labels) if l == 0]
    if not pos or not neg:
        return float("nan")
    r = rank(scores)
    rpos = sum(ri for ri, l in zip(r, labels) if l == 1)
    u = rpos - len(pos) * (len(pos) + 1) / 2.0
    return u / (len(pos) * len(neg))


def summarize(rows, features, decay_threshold, min_cohort):
    rows = [r for r in rows if r["cohort"] >= min_cohort]
    n = len(rows)
    y = [r["survival_ratio"] for r in rows]
    # binary: 1 = DECAYED (did not last) — the event we want to predict, so a
    # predictor that is LOW for decayed code gives AUC<0.5; we report AUC of the
    # predictor for "survives" so higher-gravity => higher AUC is the natural read.
    survives = [1 if r["survival_ratio"] >= decay_threshold else 0 for r in rows]
    base = sum(survives) / n if n else float("nan")
    out = {"n": n, "survive_rate": base}
    for f in features:
        x = [r.get(f, 0.0) for r in rows]
        rho, p = spearman(x, y)
        a = auc(x, survives)   # AUC of predictor discriminating survives(1) vs decayed(0)
        out[f] = {"spearman": rho, "p": p, "auc": a}
    return out, rows


def stratified_contest(rows, strat="raw_survival", feat="spread", nbins=3):
    """Claim-2 robustness: within raw_survival strata, does contest/spread still
    rank-correlate with survival? If yes, catalysis adds info beyond survival."""
    rows = [r for r in rows if feat in r and strat in r]
    if len(rows) < nbins * 5:
        return None
    rows = sorted(rows, key=lambda r: r[strat])
    per = len(rows) // nbins
    out = []
    for b in range(nbins):
        chunk = rows[b * per: (b + 1) * per] if b < nbins - 1 else rows[b * per:]
        rho, p = spearman([r[feat] for r in chunk],
                          [r["survival_ratio"] for r in chunk])
        out.append((len(chunk), rho, p))
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--level", choices=["module", "author"], required=True)
    ap.add_argument("--pair", action="append", required=True,
                    help="predictor.csv:outcome.csv")
    ap.add_argument("--min-cohort", type=int, default=30)
    ap.add_argument("--decay-threshold", type=float, default=0.7)
    args = ap.parse_args()

    key = "module" if args.level == "module" else "author"
    feats = MODULE_FEATURES if args.level == "module" else AUTHOR_FEATURES

    pooled = []
    per_repo = []
    for pair in args.pair:
        pred, out = pair.split(":")
        rows = load_join(pred, out, key)
        pooled.extend(rows)
        s, _ = summarize(rows, feats, args.decay_threshold, args.min_cohort)
        per_repo.append((pred, s))

    print(f"\n=== {args.level.upper()} LEVEL  (min_cohort={args.min_cohort}, "
          f"survive>= {args.decay_threshold}) ===")
    ps, prows = summarize(pooled, feats, args.decay_threshold, args.min_cohort)
    print(f"POOLED n={ps['n']}  survive_rate={ps['survive_rate']:.3f}")
    print(f"{'feature':<16} {'spearman':>9} {'p':>9} {'auc':>7}")
    # order: gravity-ish features first for readability
    for f in feats:
        d = ps[f]
        star = "***" if (d['p'] == d['p'] and d['p'] < 0.001) else \
               "**" if (d['p'] == d['p'] and d['p'] < 0.01) else \
               "*" if (d['p'] == d['p'] and d['p'] < 0.05) else ""
        print(f"{f:<16} {d['spearman']:>9.3f} {d['p']:>9.4f} {d['auc']:>7.3f} {star}")

    if args.level == "module":
        strat = stratified_contest(prows)
        if strat:
            print("\nClaim-2 (survival!=gravity): contest/spread ~ survival WITHIN "
                  "raw_survival terciles:")
            for i, (nb, rho, p) in enumerate(strat):
                print(f"  tercile {i+1}: n={nb:<5} spearman(spread, outcome)={rho:.3f} (p={p:.3f})")

    if len(per_repo) > 1:
        print("\nper-repo AUC (survives) for the headline features:")
        head = feats[:3]
        print(f"{'repo':<28} " + " ".join(f"{h:>12}" for h in head))
        for pred, s in per_repo:
            name = pred.split("/")[-1].replace("_mods.csv", "").replace("_auths.csv", "")
            print(f"{name:<28} " +
                  " ".join(f"{s[h]['auc']:>12.3f}" for h in head))


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Collapse backtest — Phase 4: ABANDONMENT outcome + placebo control.

Outcome (v3): post-anchor commit-rate collapse of a module ("nobody touches it
anymore"), NOT survival-mass. Future-activity based -> not mechanically tied to
ownership share, so the placebo control can isolate a real departure effect.

For anchor month t0, module m (active before, mb>=3 commits in the prior 12mo):
  mod_ret  = after / (before + after + 1)        # in [0,~1); low = went quiet
  repo_ret = repo_after / (repo_before+repo_after+1)
  abandonment = repo_ret - mod_ret               # >0 = m went quiet more than the repo

Cohorts: treat (t0 = departure) vs control (t0 = mid-tenure, no departure).
Real departure-driven abandonment iff treat correlation > control.
"""
import json, glob, math, os, sys
from collections import defaultdict

HORIZON = 12
MIN_ACTIVE = 6
SIDE = 6
EPS = 1e-9
MIN_HOLD = 0.05
MIN_BEFORE = 3          # module must have >=3 commits in the prior window
DATA = os.path.join(os.path.dirname(__file__), "data")


def load_domains(path):
    dec = json.JSONDecoder(); s = open(path).read(); i = 0; objs = []
    while i < len(s):
        while i < len(s) and s[i] in " \n\t\r": i += 1
        if i >= len(s): break
        o, j = dec.raw_decode(s, i); objs.append(o); i = j
    return objs


def midx(lbl):
    y, m = lbl.split("-"); return int(y) * 12 + int(m) - 1


def wsum(d, lo, hi):
    return sum(c for mm, c in d.items() if lo <= midx(mm) < hi)


def pathdist(x, y):
    xs, ys = x.split("/"), y.split("/"); c = 0
    for a, b in zip(xs, ys):
        if a == b: c += 1
        else: break
    return (len(xs) - c) + (len(ys) - c)


def cochange_dispersion(ad):
    cc = ad.get("cochange") or []
    pairs = cc[0]["pairs"] if cc and isinstance(cc, list) and cc[0].get("pairs") else []
    part = defaultdict(list)
    for pr in pairs:
        a, b, c = pr["module_a"], pr["module_b"], pr.get("coupling", 0)
        part[a].append((b, c)); part[b].append((a, c))
    return {m: sum(c*pathdist(m, p) for p, c in ps)/sum(c for _, c in ps)
            for m, ps in part.items() if sum(c for _, c in ps) > 0}


def churn_by_module(ad):
    return {m["module"]: m.get("change_pressure", 0) for m in (ad.get("module_scores") or [])}


def process_repo(repo, tl_domains, an_domains):
    disp = {}; churn = {}
    for ad in an_domains:
        disp.update(cochange_dispersion(ad)); churn.update(churn_by_module(ad))
    nat = _maybe(os.path.join(DATA, f"{repo}.naturalness.json"))
    act = _maybe(os.path.join(DATA, f"{repo}.activity.json")) or {}
    repo_month = defaultdict(int)
    for mm in act.values():
        for mo, c in mm.items():
            repo_month[mo] += c

    rows = []
    for dom in tl_domains:
        if "authors" not in dom or "periods" not in dom:
            continue
        labels = [p["label"] for p in dom["periods"]]
        idx = {l: k for k, l in enumerate(labels)}
        n = len(labels)
        msba = {p["label"]: (p.get("module_survival_by_author") or {}) for p in dom["periods"]}
        active = defaultdict(set)
        for a in (dom.get("authors") or []):
            for p in (a.get("periods") or []):
                if p.get("commits", 0) > 0:
                    active[a["author"]].add(p["label"])

        def total(month, module):
            return sum((msba.get(month, {}).get(module, {}) or {}).values())

        def emit(author, ti, cohort):
            t0 = labels[ti]; i0 = midx(t0)
            rb = wsum(repo_month, i0-HORIZON, i0); ra = wsum(repo_month, i0, i0+HORIZON)
            repo_ret = ra/(rb+ra+1)
            for module, am in (msba.get(t0, {}) or {}).items():
                h0 = am.get(author, 0.0); tot0 = total(t0, module)
                if h0 < EPS or tot0 < EPS:
                    continue
                share0 = h0/tot0
                if share0 < MIN_HOLD:
                    continue
                d = act.get(module, {})
                mb = wsum(d, i0-HORIZON, i0)
                if mb < MIN_BEFORE:       # module wasn't active before -> abandonment undefined
                    continue
                ma = wsum(d, i0, i0+HORIZON)
                mod_ret = ma/(mb+ma+1)
                # SUCCESSOR: other co-holders of the module (survival share >=5%)
                # who remain active in [t0, t0+HORIZON] -> someone can take it over.
                coh = [a for a, mass in am.items()
                       if a != author and mass/tot0 >= 0.05]
                succ = 0
                for a in coh:
                    if any(i0 <= midx(lbl) < i0+HORIZON for lbl in active.get(a, ())):
                        succ = 1; break
                rows.append({
                    "cohort": cohort, "repo": repo, "author": author, "module": module,
                    "concentration": share0,
                    "dispersion": disp.get(module, 0.0),
                    "naturalness": (nat or {}).get(module),
                    "churn": churn.get(module, 0.0),
                    "successor": succ,                     # 1 = an active co-holder remains
                    "n_coholders": len(coh),
                    "abandonment": repo_ret - mod_ret,     # >0 = went quiet more than repo
                    "frozen": 1 if ma == 0 else 0,
                })

        for author, months in active.items():
            if len(months) < MIN_ACTIVE:
                continue
            mi = sorted(idx[m] for m in months); li = mi[-1]; mset = set(mi)
            if li <= n-1-HORIZON:
                emit(author, li, "treat")
            for ti in mi:
                if ti < SIDE or ti > n-1-HORIZON-SIDE:
                    continue
                if any(ti-SIDE <= j < ti for j in mset) and any(ti+HORIZON <= j <= ti+HORIZON+SIDE for j in mset):
                    emit(author, ti, "control"); break
    return rows


def _maybe(p):
    return json.load(open(p)) if os.path.exists(p) else None


def pearson(xs, ys):
    n = len(xs)
    if n < 3: return float("nan")
    mx, my = sum(xs)/n, sum(ys)/n
    num = sum((x-mx)*(y-my) for x, y in zip(xs, ys))
    dx = math.sqrt(sum((x-mx)**2 for x in xs)); dy = math.sqrt(sum((y-my)**2 for y in ys))
    return num/(dx*dy) if dx > 0 and dy > 0 else float("nan")


def partial(x, y, z):
    rxy, rxz, ryz = pearson(x, y), pearson(x, z), pearson(y, z)
    return (rxy - rxz*ryz)/math.sqrt(max(1e-9, (1-rxz**2)*(1-ryz**2)))


def rcol(rows, key):
    rr = [r for r in rows if r.get(key) is not None]
    y = [r["abandonment"] for r in rr]; x = [r[key] for r in rr]
    return pearson(x, y), len(rr)


def report(rows, tag):
    print(f"\n### {tag}: n={len(rows)}  (frozen: {sum(r['frozen'] for r in rows)})")
    out = {}
    for key in ["successor", "n_coholders", "concentration", "churn", "dispersion", "naturalness"]:
        r, m = rcol(rows, key)
        out[key] = r
        print(f"  r({key:<14}, abandonment) = {r:+.3f}   (n={m})")
    # mean abandonment split by successor existence
    s1 = [r["abandonment"] for r in rows if r.get("successor") == 1]
    s0 = [r["abandonment"] for r in rows if r.get("successor") == 0]
    if s1 and s0:
        print(f"  -> mean abandonment: successor={sum(s1)/len(s1):+.3f} (n={len(s1)})  "
              f"NO-successor={sum(s0)/len(s0):+.3f} (n={len(s0)})")
    return out


def main():
    repos = sorted({os.path.basename(p).split(".")[0]
                    for p in glob.glob(os.path.join(DATA, "*.timeline.json"))})
    allr = []
    for repo in repos:
        tlp = os.path.join(DATA, f"{repo}.timeline.json")
        anp = os.path.join(DATA, f"{repo}.analyze.json")
        if not (os.path.exists(tlp) and os.path.getsize(tlp) > 0):
            continue
        an = []
        for doc in load_domains(anp):
            an.extend((doc.get("domains") or []) if isinstance(doc, dict) else [])
        try:
            rows = process_repo(repo, load_domains(tlp), an)
        except Exception as e:
            print(f"[{repo}] ERR {e}", file=sys.stderr); continue
        print(f"[{repo}] treat={sum(1 for r in rows if r['cohort']=='treat')} "
              f"control={sum(1 for r in rows if r['cohort']=='control')}")
        allr.extend(rows)

    treat = [r for r in allr if r["cohort"] == "treat"]
    control = [r for r in allr if r["cohort"] == "control"]
    print("\n=== ABANDONMENT outcome, PLACEBO-CONTROLLED ===")
    t = report(treat, "TREAT (departure)")
    c = report(control, "CONTROL (mid-tenure, no departure)")
    print("\n-- departure effect = treat - control --")
    for key in ["successor", "n_coholders", "concentration", "churn", "dispersion", "naturalness"]:
        print(f"  {key:<14}: {t[key]:+.3f} - {c[key]:+.3f} = {t[key]-c[key]:+.3f}")

    print("\n-- partial correlations on TREAT --")
    rr = [r for r in treat if r.get("naturalness") is not None]
    conc = [r["concentration"] for r in treat]; chu = [r["churn"] for r in treat]; ab = [r["abandonment"] for r in treat]
    print(f"  partial r(concentration, abandonment | churn) = {partial(conc, ab, chu):+.3f}")
    print(f"  partial r(churn, abandonment | concentration) = {partial(chu, ab, conc):+.3f}")


if __name__ == "__main__":
    main()

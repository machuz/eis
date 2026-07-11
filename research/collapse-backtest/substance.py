#!/usr/bin/env python3
"""Substance = robust_survival x catalysis, computed on 8 OSS repos.

Aggregated per contributor (we measure the CODE's fate, then group by who
produced it; identity is never a feature). Key questions for the X post:
  - Does raw output (commits/lines) buy Substance?  r(commits, Substance)
  - Does the x-gate matter? how many "inert survivors" (survived, no catalysis)
  - How skewed is Substance? (scarce = the point)
"""
import json, glob, os, math
from collections import defaultdict

DATA = os.path.join(os.path.dirname(__file__), "data")


def load(path):
    dec = json.JSONDecoder(); s = open(path).read(); i = 0; out = []
    while i < len(s):
        while i < len(s) and s[i] in " \n\t\r": i += 1
        if i >= len(s): break
        o, j = dec.raw_decode(s, i); out.append(o); i = j
    return out


def pearson(xs, ys):
    n = len(xs)
    if n < 3: return float("nan")
    mx, my = sum(xs)/n, sum(ys)/n
    num = sum((x-mx)*(y-my) for x, y in zip(xs, ys))
    dx = math.sqrt(sum((x-mx)**2 for x in xs)); dy = math.sqrt(sum((y-my)**2 for y in ys))
    return num/(dx*dy) if dx > 0 and dy > 0 else float("nan")


rows = []
seen_fields = set()
for anp in sorted(glob.glob(os.path.join(DATA, "*.analyze.json"))):
    repo = os.path.basename(anp).split(".")[0]
    for doc in load(anp):
        for dom in (doc.get("domains") or []):
            for m in (dom.get("members") or []):
                seen_fields |= set(m.keys())
                commits = m.get("commits", 0) or 0
                rs = m.get("robust_survival", 0) or 0
                cat = m.get("catalysis", 0) or 0
                surv = m.get("survival", 0) or 0
                grav = m.get("gravity", 0) or 0
                la = m.get("lines_added", 0) or 0
                # A: robust survival x catalysis   B: survival x catalysis (catalysis filters neglect)
                rows.append(dict(repo=repo, member=m.get("member", "?"), commits=commits,
                                 lines_added=la, survival=surv, robust=rs, catalysis=cat,
                                 gravity=grav, substance=rs*cat/100.0, substance2=surv*cat/100.0))

print("member fields present:", "robust_survival" in seen_fields, "catalysis" in seen_fields)

rows = [r for r in rows if r["commits"] >= 20 or r["survival"] > 0]
n = len(rows)
print(f"contributors (>=20 commits or survival>0): {n}")

def cnt(f): return sum(1 for r in rows if f(r))
print(f"  robust_survival>0: {cnt(lambda r: r['robust']>0)}  "
      f"catalysis>0: {cnt(lambda r: r['catalysis']>0)}  "
      f"both>0 (any Substance): {cnt(lambda r: r['robust']>0 and r['catalysis']>0)}")

def gini(vals):
    v = sorted(vals); s = sum(v); c = 0.0
    for i, x in enumerate(v, 1):
        c += i * x
    return (2*c)/(len(v)*s) - (len(v)+1)/len(v) if s > 0 else 0.0

sub = [r["substance"] for r in rows]; com = [r["commits"] for r in rows]; ln = [r["lines_added"] for r in rows]

def topshare(key, k=0.10):
    ordered = sorted(rows, key=lambda r: -r[key])
    tot = sum(r[key] for r in rows)
    return sum(r[key] for r in ordered[:max(1, int(k*n))]) / tot if tot > 0 else 0

print(f"\n-- concentration: how much sits in the top 10% of contributors --")
print(f"  commits : top10% hold {topshare('commits')*100:.0f}%   (Gini {gini(com):.2f})")
print(f"  Substance: top10% hold {topshare('substance')*100:.0f}%   (Gini {gini(sub):.2f})")

print(f"\n-- does raw output buy Substance? --")
print(f"  r(commits,     Substance) = {pearson(com, sub):+.3f}   (R^2={pearson(com,sub)**2*100:.0f}%)")
print(f"  r(lines_added, Substance) = {pearson(ln, sub):+.3f}")

# among code that robustly survived (top quartile robust), how many catalyzed?
survived = sorted(rows, key=lambda r: -r["robust"])[:max(4, n//4)]
med_cat = sorted(r["catalysis"] for r in rows)[n//2]
inert = [r for r in survived if r["catalysis"] <= med_cat]
print(f"\n-- why the x-gate: of the {len(survived)} contributors whose code most robustly survived,")
print(f"   {len(inert)} ({len(inert)*100//len(survived)}%) catalyzed at/below the median -> survived but inert")

top = sorted(rows, key=lambda r: -r["substance"])
print(f"\n-- top Substance (anonymized) : Substance | robust | catalysis | commits --")
for r in top[:5]:
    print(f"   repo={r['repo']:<9} S={r['substance']:5.1f}  R={r['robust']:5.1f}  C={r['catalysis']:5.1f}  commits={r['commits']}")
print("   ...")
# a busy-but-low-substance example
busy = sorted(rows, key=lambda r: -r["commits"])
for r in busy:
    if r["substance"] < 1 and r["commits"] > 50:
        print(f"   BUSY-BUT-THIN repo={r['repo']:<9} S={r['substance']:5.1f}  R={r['robust']:5.1f}  C={r['catalysis']:5.1f}  commits={r['commits']}")
        break

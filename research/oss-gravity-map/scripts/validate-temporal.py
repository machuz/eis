#!/usr/bin/env python3
"""Validate EIS against independent ground-truth, with a temporal (per-period) dimension.

Two questions:
  1. recall@20 — of a repo's independent architect_candidates (GitHub logins from
     CODEOWNERS / release authors / top contributors), how many land in EIS's static
     gravity top-20?
  2. temporal correctness — for architect_candidates that EIS's *static* top-20 misses,
     did they ever peak high in the *per-period* timeline (i.e. was EIS right to credit
     them in their era and fade them once their code was superseded)?

The point of (2): a low static gravity is only a "miss" if the person never load-bore
durable code in any era. If their timeline peaked, EIS captured them correctly and the
static low is a *time-faithful fade*, not an error.

Bridge: architect_candidates are GitHub logins; EIS reports author display names.
We bridge via data/alias-map.json (per-repo {login: name}) then github-users-cache.json.
"""
import json, os, sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

def first_object(path):
    """Read the first top-level JSON object from a file that may have trailing data."""
    raw = open(path).read()
    depth = 0; ins = False; esc = False; end = 0
    for i, c in enumerate(raw):
        if esc: esc = False; continue
        if c == '\\' and ins: esc = True; continue
        if c == '"': ins = not ins; continue
        if ins: continue
        if c == '{': depth += 1
        elif c == '}':
            depth -= 1
            if depth == 0: end = i + 1; break
    return json.loads(raw[:end])

alias = json.load(open(os.path.join(ROOT, 'data/alias-map.json')))
try:
    ucache = json.load(open(os.path.join(ROOT, 'data/github-users-cache.json')))
except FileNotFoundError:
    ucache = {}

def login_to_name(repo, login):
    lo = login.lower()
    for k, v in alias.get(repo, {}).items():
        if k.lower() == lo: return v
    u = ucache.get(login) or ucache.get(lo)
    if isinstance(u, dict) and u.get('name'): return u['name']
    return None

def static_members(repo):
    p = os.path.join(ROOT, f'data/results/{repo}.json')
    if not os.path.exists(p): return None
    doms = json.load(open(p))['domains']
    m = {}
    for d in doms:
        for x in d['members']:
            nm = x.get('member') or ''
            m[nm] = max(m.get(nm, 0), x.get('gravity', 0) or 0)
    return m

def perperiod_peaks(repo):
    p = os.path.join(ROOT, f'data/results/{repo}-timeline.json')
    if not os.path.exists(p): return None
    tl = first_object(p)
    labels = [pp['label'] for pp in tl['periods']]
    out = {}
    for a in tl['authors']:
        peaks = [(pp.get('gravity', 0) or 0, labels[i], pp.get('role'))
                 for i, pp in enumerate(a['periods'])]
        out[a['author']] = max(peaks, key=lambda x: x[0])
    return out

def gt_repos():
    d = os.path.join(ROOT, 'data/ground-truth')
    return sorted(f[:-5] for f in os.listdir(d) if f.endswith('.json'))

def main():
    recall_rows = []
    temporal_rows = []
    hit_peak_rows = []
    for gt_name in gt_repos():
        gt = json.load(open(os.path.join(ROOT, f'data/ground-truth/{gt_name}.json')))
        # ground-truth repo name may differ from results filename; try a few
        cands = [gt_name, gt.get('name', ''), gt['repo'].split('/')[-1]]
        repo = next((c for c in cands if c and static_members(c) is not None), None)
        if not repo: continue
        st = static_members(repo)
        top20 = {nm for nm, _ in sorted(st.items(), key=lambda x: -x[1])[:20]}
        archs = gt.get('architect_candidates') or []
        if not archs: continue
        hit = miss_named = bridged = 0
        peaks = perperiod_peaks(repo)
        for lg in archs:
            nm = login_to_name(repo, lg)
            if not nm or nm not in st:
                continue  # cannot bridge — exclude from denominator (conservative)
            bridged += 1
            if nm in top20:
                hit += 1
                if peaks is not None:
                    pk, yr, role = peaks.get(nm, (0, '-', None))
                    if pk >= 40:
                        hit_peak_rows.append((repo, nm, st[nm], pk, yr, role))
            else:
                miss_named += 1
                if peaks is not None:
                    pk, yr, role = peaks.get(nm, (0, '-', None))
                    temporal_rows.append((repo, nm, st[nm], pk, yr, role))
        if bridged:
            recall_rows.append((repo, hit, bridged, hit / bridged))

    print("=== recall@20 vs independent architect_candidates (bridged logins only) ===")
    tot_h = tot_b = 0
    for repo, hit, b, r in sorted(recall_rows, key=lambda x: -x[3]):
        print(f"  {repo:14} {hit:2}/{b:<2} = {r*100:4.0f}%")
        tot_h += hit; tot_b += b
    if tot_b:
        print(f"  {'OVERALL':14} {tot_h:2}/{tot_b:<2} = {tot_h/tot_b*100:4.0f}%  "
              f"(micro-avg across {len(recall_rows)} repos)")

    if hit_peak_rows:
        print("\n=== temporal credit: confirmed architects — EIS peaks them in their era ===")
        for repo, nm, sg, pk, yr, role in sorted(hit_peak_rows, key=lambda x: -x[3]):
            print(f"  {repo:11} {nm[:22]:22} per-period peak {pk:3.0f} @{yr} ({role}) "
                  f"| static now {sg:4.1f}")

    if temporal_rows:
        print("\n=== temporal check: static-missed architect_candidates — did they peak per-period? ===")
        caught = faint = 0
        for repo, nm, sg, pk, yr, role in sorted(temporal_rows, key=lambda x: -x[3]):
            v = "CAUGHT in era" if pk >= 40 else ("partial" if pk >= 15 else "never high")
            if pk >= 40: caught += 1
            elif pk < 15: faint += 1
            mark = "  <- time-faithful fade" if pk >= 40 else ""
            print(f"  {repo:11} {nm[:22]:22} static {sg:5.1f} | per-period peak {pk:3.0f} @{yr} ({role}){mark}")
        print(f"  summary: caught-in-era(>=40)={caught}  never-high(<15)={faint}  "
              f"total-missed-with-timeline={len(temporal_rows)}")

if __name__ == '__main__':
    main()

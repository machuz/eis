#!/usr/bin/env python3
"""Lifetime-footprint gravity experiment.

The published Gravity uses a survival GATE built on robust_survival — survival
only in modules where OTHERS are actively committing. That measures gravity that
is *live right now*: a finished, settled foundation (no one currently pressing on
it) reads quiet even when the whole system still rests on it. Sebastian Markbåge
in React is the canonical case: Indispensability 100, Catalysis 100, raw survival
100, yet Gravity 9.9 because robust_survival is 0.

This script recomputes a second, "lifetime" gravity that swaps the survival gate
from robust_survival to raw survival (does your code still stand, regardless of
whether others are pressing on it *now*), keeping the Catalysis gate. The
question: does it surface the settled architects without re-opening the
self-churn gaming hole that robust_survival was built to close?

Pure recompute from the per-member axes already in data/results/*.json — no blame,
no clone. Run: python3 lifetime_footprint.py
"""
import json, glob, os

def gate(x):       return 0.15 + 0.85 * (x / 100.0)
def shape(m, w):   return w[0]*m.get('design', 0) + w[1]*m.get('breadth', 0) + w[2]*m.get('indispensability', 0)

SHAPE_CUR  = (0.45, 0.25, 0.30)   # published shape weights (design, breadth, indisp)
SHAPE_INDISP = (0.35, 0.20, 0.45) # B) lifetime variant: lean harder on "the system leans on you"

def g_current(m):  return gate(m.get('catalysis', 0)) * gate(m.get('robust_survival', 0)) * shape(m, SHAPE_CUR)
def g_lifetime(m): return gate(m.get('catalysis', 0)) * gate(m.get('survival', 0))        * shape(m, SHAPE_CUR)
def g_life_indisp(m): return gate(m.get('catalysis', 0)) * gate(m.get('survival', 0))     * shape(m, SHAPE_INDISP)

def load_members(fp):
    return [m for dom in json.load(open(fp))['domains'] for m in dom['members']]

def main():
    repos = sorted(glob.glob(os.path.join(os.path.dirname(__file__), '..', 'data', 'results', '*.json')))
    mismatch = total = overlap = nrepo = 0
    risers, gaming = [], []
    for fp in repos:
        repo = os.path.basename(fp)[:-5]
        mem = load_members(fp)
        if not mem:
            continue
        nrepo += 1
        for m in mem:
            total += 1
            gc, gl, gli = g_current(m), g_lifetime(m), g_life_indisp(m)
            m['_gc'], m['_gl'], m['_gli'] = gc, gl, gli
            if 'gravity' in m and abs(gc - m['gravity']) > 0.6:
                mismatch += 1
            risers.append((gl - gc, repo, m.get('member', '?'), m.get('catalysis', 0),
                           m.get('indispensability', 0), m.get('robust_survival', 0),
                           m.get('survival', 0), gc, gl, gli))
            if m.get('survival', 0) >= 60 and m.get('catalysis', 0) < 20:   # self-churn risk profile
                gaming.append((m.get('member', '?'), repo, m.get('survival', 0), m.get('catalysis', 0), gc, gl))
        cur10 = {m['member'] for m in sorted(mem, key=lambda x: -x['_gc'])[:10]}
        life10 = {m['member'] for m in sorted(mem, key=lambda x: -x['_gl'])[:10]}
        overlap += len(cur10 & life10)

    print(f"repos={nrepo} members={total} formula_mismatch(>0.6)={mismatch}  avg top10 overlap cur↔lifetime={overlap/nrepo:.1f}/10")

    react = load_members([r for r in repos if r.endswith('react.json')][0])
    for m in react:
        m['_gc'], m['_gl'] = g_current(m), g_lifetime(m)
    mk = next(m for m in react if 'Markb' in m['member'])
    print(f"\nReact Markbåge: current {mk['_gc']:.1f} -> lifetime {mk['_gl']:.1f}")

    print("\nBiggest risers (lifetime - current), top 12:")
    for d, repo, name, cat, ind, rob, sur, gc, gl, gli in sorted(risers, reverse=True)[:12]:
        print(f"  +{d:5.1f}  {name[:22]:22} {repo:11} cat={cat:3} indisp={ind:5} robust={rob:4} surv={sur:4}  cur {gc:5.1f} -> life {gl:5.1f} (indisp-wt {gli:5.1f})")

    glv = [g[5] for g in gaming]
    print(f"\nGAMING CHECK (surv>=60 & catalysis<20): n={len(gaming)} lifetime gravity max={max(glv):.1f} mean={sum(glv)/len(glv):.1f}")

    # B) indisp-weighting effect on the low-indispensability risers
    print("\nB) indisp-weight effect — risers with indispensability < 30 (broad-but-not-load-bearing):")
    low = [r for r in sorted(risers, reverse=True)[:40] if r[4] < 30]
    for d, repo, name, cat, ind, rob, sur, gc, gl, gli in low[:8]:
        print(f"  {name[:22]:22} {repo:11} indisp={ind:5}  life {gl:5.1f} -> indisp-wt {gli:5.1f}  (Δ {gli-gl:+.1f})")

if __name__ == '__main__':
    main()

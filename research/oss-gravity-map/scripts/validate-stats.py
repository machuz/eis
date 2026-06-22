#!/usr/bin/env python3
"""Statistical validation of EIS gravity against independent ground-truth.

Beyond the point-estimate recall@20, this answers the questions a referee asks:

  1. RANKING (AUC). Does gravity rank declared architects above non-architects?
     AUC = P(a random architect outranks a random non-architect), ties = 0.5.
     Null = 0.5. Computed per repo and pooled, with a cluster bootstrap 95% CI.
  2. TOP-LIST ENRICHMENT (hypergeometric). Is hitting 122/177 architects in the
     top-20 more than a random top-20 would catch? Fold-enrichment + exact p per
     repo (one-sided P[X >= observed]), and a pooled expected-vs-observed.
  3. TEMPORAL SEPARATION (Mann-Whitney). Do the architects EIS catches peak higher
     in the per-period timeline than the ones it misses? One-sided U test.

Pure stdlib (no scipy): math.comb for hypergeometric, normal-approx Mann-Whitney
with tie correction, seeded cluster bootstrap. Reproducible: seed = 42.
"""
import json, os, math, random

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
random.seed(42)

# ---- bridge (shared with validate-temporal.py) ----
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
    m = {}
    for d in json.load(open(p))['domains']:
        for x in d['members']:
            nm = x.get('member') or ''
            m[nm] = max(m.get(nm, 0), x.get('gravity', 0) or 0)
    return m

def first_object(path):
    raw = open(path).read(); depth = 0; ins = False; esc = False; end = 0
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

def perperiod_peaks(repo):
    p = os.path.join(ROOT, f'data/results/{repo}-timeline.json')
    if not os.path.exists(p): return None
    tl = first_object(p)
    return {a['author']: max((pp.get('gravity', 0) or 0) for pp in a['periods'])
            for a in tl['authors']}

def gt_repos():
    d = os.path.join(ROOT, 'data/ground-truth')
    return sorted(f[:-5] for f in os.listdir(d) if f.endswith('.json'))

# ---- statistics (stdlib only) ----
def norm_sf(z):  # one-sided upper-tail of standard normal
    return 0.5 * math.erfc(z / math.sqrt(2))

def auc_and_u(pos_scores, neg_scores):
    """AUC with tie=0.5 and Mann-Whitney U + normal-approx one-sided p (pos>neg)."""
    n1, n2 = len(pos_scores), len(neg_scores)
    if n1 == 0 or n2 == 0: return None
    allv = sorted(pos_scores + neg_scores)
    # rank with ties averaged
    ranks = {}; i = 0
    while i < len(allv):
        j = i
        while j + 1 < len(allv) and allv[j + 1] == allv[i]: j += 1
        r = (i + j) / 2 + 1
        ranks[allv[i]] = r
        i = j + 1
    R1 = sum(ranks[v] for v in pos_scores)
    U1 = R1 - n1 * (n1 + 1) / 2
    auc = U1 / (n1 * n2)
    # tie-corrected variance
    from collections import Counter
    tie = Counter(allv); N = n1 + n2
    tcorr = sum(t**3 - t for t in tie.values())
    mu = n1 * n2 / 2
    var = n1 * n2 / 12 * ((N + 1) - tcorr / (N * (N - 1)))
    z = (U1 - mu) / math.sqrt(var) if var > 0 else 0.0
    return auc, U1, z, norm_sf(z)

def hyper_sf(N, K, n, x):  # P[X >= x], X ~ Hypergeometric(N,K,n)
    lo, hi = max(0, n - (N - K)), min(n, K)
    tot = math.comb(N, n)
    return sum(math.comb(K, k) * math.comb(N - K, n - k) for k in range(max(x, lo), hi + 1)) / tot

# ---- collect ----
per_repo = []          # (repo, N, K, hits, recall, auc, auc_p, exp20, fold, hyp_p)
caught_peaks, missed_peaks = [], []
pos_all, neg_all = [], []   # for a sanity pooled AUC

for gt_name in gt_repos():
    gt = json.load(open(os.path.join(ROOT, f'data/ground-truth/{gt_name}.json')))
    repo = next((c for c in (gt_name, gt.get('name', ''), gt['repo'].split('/')[-1])
                 if c and static_members(c) is not None), None)
    if not repo: continue
    st = static_members(repo)
    archs = gt.get('architect_candidates') or []
    cand_names = {login_to_name(repo, lg) for lg in archs}
    cand_names = {n for n in cand_names if n and n in st}
    K = len(cand_names)
    if K == 0: continue
    N = len(st)
    ranked = sorted(st.items(), key=lambda x: -x[1])
    top20 = {nm for nm, _ in ranked[:20]}
    hits = len(cand_names & top20)
    pos = [st[n] for n in cand_names]
    neg = [v for n, v in st.items() if n not in cand_names]
    a = auc_and_u(pos, neg)
    auc, _, _, ap = a if a else (None, None, None, None)
    pos_all += [(repo, v) for v in pos]; neg_all += [(repo, v) for v in neg]
    exp20 = 20 * K / N
    fold = (hits / exp20) if exp20 else float('inf')
    hp = hyper_sf(N, K, min(20, N), hits)
    per_repo.append((repo, N, K, hits, hits / K, auc, ap, exp20, fold, hp))
    peaks = perperiod_peaks(repo)
    if peaks:
        for n in cand_names:
            pk = peaks.get(n, 0)
            (caught_peaks if n in top20 else missed_peaks).append(pk)

# ---- report ----
print("=== Ranking discrimination: AUC (gravity ranks architects above the rest) ===")
print(f"{'repo':14}{'N':>6}{'K':>4}{'hit@20':>7}{'AUC':>7}{'p(AUC>0.5)':>12}")
aucs = []
for repo, N, K, hits, rec, auc, ap, e, f, hp in sorted(per_repo, key=lambda x: -(x[5] or 0)):
    aucs.append(auc)
    pstr = f"{ap:.1e}" if ap is not None else "—"
    print(f"{repo:14}{N:>6}{K:>4}{hits:>4}/{K:<2}{auc:>7.3f}{pstr:>12}")
mean_auc = sum(aucs) / len(aucs)

# cluster bootstrap CI on mean AUC and on micro recall
def boot_ci(stat_fn, B=5000):
    vals = []
    for _ in range(B):
        samp = [per_repo[random.randrange(len(per_repo))] for _ in per_repo]
        vals.append(stat_fn(samp))
    vals.sort()
    return vals[int(0.025 * B)], vals[int(0.975 * B)]

lo_a, hi_a = boot_ci(lambda s: sum(r[5] for r in s) / len(s))
micro = sum(r[3] for r in per_repo) / sum(r[2] for r in per_repo)
lo_r, hi_r = boot_ci(lambda s: sum(r[3] for r in s) / sum(r[2] for r in s))
print(f"\nmean AUC over {len(per_repo)} repos = {mean_auc:.3f}  (95% cluster-bootstrap CI "
      f"{lo_a:.3f}–{hi_a:.3f}; null = 0.500)")
print(f"micro recall@20            = {micro*100:.0f}%   (95% CI {lo_r*100:.0f}–{hi_r*100:.0f}%)")

print("\n=== Top-20 enrichment vs a random top-20 (hypergeometric) ===")
exp_tot = sum(r[7] for r in per_repo); obs_tot = sum(r[3] for r in per_repo)
print(f"observed hits = {obs_tot} / {sum(r[2] for r in per_repo)}   "
      f"expected by chance = {exp_tot:.1f}   fold-enrichment = {obs_tot/exp_tot:.1f}x")
sig = sum(1 for r in per_repo if r[9] < 1e-3)
weak = [(r[0], r[3], r[2], r[9]) for r in per_repo if r[9] >= 1e-3]
print(f"per-repo one-sided p < 1e-3 in {sig}/{len(per_repo)} repos.")
if weak:
    print("  not individually significant: " +
          ", ".join(f"{w[0]}({w[1]}/{w[2]}, p={w[3]:.1e})" for w in weak))

if caught_peaks and missed_peaks:
    res = auc_and_u(caught_peaks, missed_peaks)
    auc_t, U, z, p = res
    cm = sorted(caught_peaks)[len(caught_peaks)//2]
    mm = sorted(missed_peaks)[len(missed_peaks)//2]
    print("\n=== Temporal separation: per-period peak of caught vs missed architects ===")
    print(f"caught (n={len(caught_peaks)}) median peak = {cm:.0f} ; "
          f"missed (n={len(missed_peaks)}) median peak = {mm:.0f}")
    print(f"Mann-Whitney one-sided p = {p:.1e}  (AUC = {auc_t:.3f}: a caught architect's "
          f"era-peak outranks a missed one's {auc_t*100:.0f}% of the time)")

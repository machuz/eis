#!/usr/bin/env python3
"""Is repowise-style defect prediction real, or a tautology of size/volume?

For each module: count total commits and fix-commits (SZZ-lite: fix-ish subject).
Then ask:
  A. r(churn, fix_count)          -- the repowise-style headline (expected +)
  B. r(total_commits, fix_count)  -- the volume tautology baseline (expected ++)
  C. partial r(churn, fix_count | total_commits)  -- does churn ADD beyond volume?
  D. r(churn, fix_RATE=fix/total) -- does churn predict defect-PRONENESS, or only volume?
If C and D collapse toward 0, "churn predicts defects" is mostly "active modules
get more commits of every kind, including fixes" = near-tautological.
"""
import json, os, re, subprocess, glob, math, sys
from collections import defaultdict

BASE = os.path.dirname(__file__)
DATA = os.path.join(BASE, "data")
FIX_RE = re.compile(r"\b(fix(e[ds]|ing)?|bug|bugfix|hotfix|defect|regress\w*|crash\w*|broken|fault\w*|leak|npe|segfault)\b", re.I)


def _load(path):
    dec = json.JSONDecoder(); s = open(path).read(); i = 0; out = []
    while i < len(s):
        while i < len(s) and s[i] in " \n\t\r": i += 1
        if i >= len(s): break
        o, j = dec.raw_decode(s, i); out.append(o); i = j
    return out


def module_meta(analyze_path):
    """module -> churn(change_pressure); also sorted module names."""
    churn = {}
    for doc in _load(analyze_path):
        for dom in (doc.get("domains") or []):
            for m in (dom.get("module_scores") or []):
                churn[m["module"]] = m.get("change_pressure", 0)
    mods = sorted([m for m in churn if m != "."], key=len, reverse=True)
    return churn, mods


def assign(rel, mods):
    for m in mods:
        if rel == m or rel.startswith(m + "/"):
            return m
    return "."


def counts(repo, mods):
    repo_dir = os.path.join(BASE, repo)
    p = subprocess.run(["git", "-C", repo_dir, "log", "--no-merges",
                        "--pretty=format:@%at%x1f%s", "--numstat"],
                       capture_output=True, text=True, errors="ignore")
    tot = defaultdict(int); fix = defaultdict(int)
    is_fix = False; touched = set()

    def flush():
        for m in touched:
            tot[m] += 1
            if is_fix:
                fix[m] += 1

    for line in p.stdout.splitlines():
        if line.startswith("@"):
            if touched:
                flush()
            subj = line.split("\x1f", 1)[1] if "\x1f" in line else ""
            is_fix = bool(FIX_RE.search(subj))
            touched = set()
        elif line.strip():
            parts = line.split("\t")
            if len(parts) == 3:
                touched.add(assign(parts[2], mods))
    if touched:
        flush()
    return tot, fix


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


def main():
    rows = []
    repos = sorted({os.path.basename(p).split(".")[0]
                    for p in glob.glob(os.path.join(DATA, "*.analyze.json"))})
    for repo in repos:
        anp = os.path.join(DATA, f"{repo}.analyze.json")
        if not os.path.isdir(os.path.join(BASE, repo)):
            continue
        churn, mods = module_meta(anp)
        tot, fix = counts(repo, mods)
        r = 0
        for m in tot:
            if m == "." or tot[m] < 10:      # need enough commits to have a stable rate
                continue
            if m not in churn:
                continue
            rows.append({"repo": repo, "module": m, "total": tot[m],
                         "fix": fix[m], "rate": fix[m]/tot[m], "churn": churn[m]})
            r += 1
        print(f"[{repo}] modules with >=10 commits: {r}  total_fix_share={sum(fix.values())/max(1,sum(tot.values())):.2f}")

    print(f"\n=== POOLED: {len(rows)} modules across {len(repos)} repos ===")
    churn = [x["churn"] for x in rows]; total = [x["total"] for x in rows]
    fix = [x["fix"] for x in rows]; rate = [x["rate"] for x in rows]
    print("\n-- Is churn's defect prediction real, or volume tautology? --")
    print(f"  A. r(churn,  fix_count)                   = {pearson(churn, fix):+.3f}   (repowise-style headline)")
    print(f"  B. r(total_commits, fix_count)            = {pearson(total, fix):+.3f}   (pure volume baseline)")
    print(f"  C. partial r(churn, fix_count | total)    = {partial(churn, fix, total):+.3f}   (churn's ADD beyond volume)")
    print(f"  D. r(churn,  fix_RATE=fix/total)          = {pearson(churn, rate):+.3f}   (defect-proneness, not volume)")
    print(f"     r(total_commits, fix_RATE)             = {pearson(total, rate):+.3f}")
    print(f"\n  mean fix-rate = {sum(rate)/len(rate):.3f}  (fraction of commits that are fixes)")


if __name__ == "__main__":
    main()

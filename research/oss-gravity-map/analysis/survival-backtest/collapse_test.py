#!/usr/bin/env python3
"""Track-2 test: does a module's contest structure at T predict COLLAPSE after its
owner departs? The sharpened thesis (memory #57): sole-owned code collapses when
the owner leaves; contested (others-built-on) code survives. Cohort survival
unconditionally washed this out — here we CONDITION on owner departure, where the
bus-factor effect should actually manifest.

Per repo + anchor T (reuses the predictor/outcome CSVs, plus two cheap git-log
passes — no timeline re-run):
  owner(m)        = top committer of module m by lines changed <= T
  departed(m)     = owner's LAST commit over full history is > gap days before HEAD
  contested(m)    = predictor spread / author_count (others build on it)
  survival(m)     = outcome cohort survival ratio to HEAD
Claim: AMONG departed-owner modules, contest predicts survival (sole-owned
collapses); among stayed-owner modules the effect is weak. The differentiated
signal is the contest x departure interaction.

Usage: collapse_test.py --corpus <dir> --repos r1,r2,... --gap-days 365
"""
import argparse
import csv
import json
import math
import subprocess
from collections import defaultdict


def git_out(repo, *args):
    return subprocess.run(["git", "-C", repo, *args], capture_output=True,
                          text=True, errors="replace", check=True).stdout


def module_of(path, module_paths):
    best = ""
    for mp in module_paths:
        if mp and (path == mp or path.startswith(mp + "/")) and len(mp) > len(best):
            best = mp
    return best or None


def module_owner_and_author_last(repo, anchor, module_paths):
    """One pass <=T for module->top-committer(by lines); one full pass for each
    author's last commit epoch."""
    # module -> author -> lines, up to T
    mono = git_out(repo, "log", f"--until={anchor}", "--no-merges", "--numstat",
                   "--format=%x01%an")
    lines_by = defaultdict(lambda: defaultdict(int))
    cur = None
    for ln in mono.splitlines():
        if ln.startswith("\x01"):
            cur = ln[1:]
            continue
        if not ln.strip() or cur is None:
            continue
        p = ln.split("\t")
        if len(p) != 3:
            continue
        add, dele, path = p
        mod = module_of(path, module_paths)
        if mod is None:
            continue
        try:
            lines_by[mod][cur] += (0 if add == "-" else int(add)) + (0 if dele == "-" else int(dele))
        except ValueError:
            pass
    owner = {}
    for mod, authors in lines_by.items():
        if authors:
            owner[mod] = max(authors.items(), key=lambda kv: kv[1])[0]
    # author -> last commit epoch (full history)
    full = git_out(repo, "log", "--no-merges", "--format=%an%x09%at")
    last = {}
    for ln in full.splitlines():
        if "\t" not in ln:
            continue
        name, at = ln.rsplit("\t", 1)
        try:
            at = int(at)
        except ValueError:
            continue
        if name not in last or at > last[name]:
            last[name] = at
    return owner, last


def head_epoch(repo):
    return int(git_out(repo, "log", "-1", "--format=%at").strip())


def rank(v):
    o = sorted(range(len(v)), key=lambda i: v[i]); r = [0.0] * len(v); i = 0
    while i < len(o):
        j = i
        while j + 1 < len(o) and v[o[j + 1]] == v[o[i]]:
            j += 1
        for k in range(i, j + 1):
            r[o[k]] = (i + j) / 2 + 1
        i = j + 1
    return r


def spearman(x, y):
    n = len(x)
    if n < 4:
        return float("nan"), n
    rx, ry = rank(x), rank(y); mx = sum(rx) / n; my = sum(ry) / n
    sxy = sum((a - mx) * (b - my) for a, b in zip(rx, ry))
    sxx = sum((a - mx) ** 2 for a in rx); syy = sum((b - my) ** 2 for b in ry)
    return (sxy / math.sqrt(sxx * syy) if sxx and syy else float("nan")), n


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--corpus", required=True)
    ap.add_argument("--repos", required=True)
    ap.add_argument("--bt", default="/tmp/bt")
    ap.add_argument("--gap-days", type=int, default=365)
    ap.add_argument("--min-cohort", type=int, default=20)
    ap.add_argument("--sole-share", type=float, default=0.7,
                    help="top_share above this = sole-owned")
    args = ap.parse_args()

    pooled = []            # dicts across all repos
    per_repo_rho = defaultdict(list)
    for r in args.repos.split(","):
        meta = json.load(open(f"{args.bt}/{r}_meta.json"))
        anchor = meta["anchor_end"]
        repo = f"{args.corpus}/data/repos/{r}"
        outcome = {}
        for x in csv.DictReader(open(f"{args.bt}/{r}_out_mods.csv")):
            if x["survival_ratio"] != "":
                outcome[x["module"]] = (float(x["survival_ratio"]), int(x["cohort"]))
        pred = {}
        for x in csv.DictReader(open(f"{args.bt}/{r}_mods.csv")):
            pred[x["module"]] = x
        module_paths = list(pred.keys())
        owner, last = module_owner_and_author_last(repo, anchor, module_paths)
        he = head_epoch(repo)
        cutoff = he - args.gap_days * 86400

        rows = []
        for mod, (surv, cohort) in outcome.items():
            if cohort < args.min_cohort or mod not in owner or mod not in pred:
                continue
            o = owner[mod]
            departed = last.get(o, 0) < cutoff
            p = pred[mod]
            rows.append({
                "repo": r, "module": mod, "survival": surv,
                "spread": float(p["spread"]), "author_count": float(p["author_count"]),
                "top_share": float(p["top_share"]), "departed": departed,
            })
        pooled.extend(rows)
        # within-repo: among departed-owner modules, contest ~ survival
        dep = [z for z in rows if z["departed"]]
        if len(dep) >= 4:
            rho, n = spearman([z["spread"] for z in dep], [z["survival"] for z in dep])
            if rho == rho:
                per_repo_rho["spread|departed"].append(rho)
            rho2, _ = spearman([z["top_share"] for z in dep], [z["survival"] for z in dep])
            if rho2 == rho2:
                per_repo_rho["top_share|departed"].append(rho2)

    dep = [z for z in pooled if z["departed"]]
    sta = [z for z in pooled if not z["departed"]]
    print(f"\n=== TRACK-2 COLLAPSE-ON-DEPARTURE (gap={args.gap_days}d, min_cohort={args.min_cohort}) ===")
    print(f"pooled modules: {len(pooled)}  |  owner-departed: {len(dep)}  owner-stayed: {len(sta)}")

    def block(name, rows):
        if len(rows) < 4:
            print(f"\n[{name}] n={len(rows)} (too few)")
            return
        y = [z["survival"] for z in rows]
        for feat in ("spread", "author_count", "top_share"):
            rho, n = spearman([z[feat] for z in rows], y)
            print(f"[{name}] spearman({feat:<13}, survival) = {rho:+.3f}  (n={n})")
    block("owner DEPARTED", dep)
    block("owner STAYED  ", sta)

    # 2x2 mean survival: departed/stayed x sole/contested
    print("\nmean survival_ratio  (sole = top_share > "
          f"{args.sole_share}):")
    print(f"{'':<16}{'sole-owned':>12}{'contested':>12}")
    for grp, rows in (("departed", dep), ("stayed", sta)):
        sole = [z["survival"] for z in rows if z["top_share"] > args.sole_share]
        cont = [z["survival"] for z in rows if z["top_share"] <= args.sole_share]
        ms = f"{sum(sole)/len(sole):.3f}" if sole else "  -"
        mc = f"{sum(cont)/len(cont):.3f}" if cont else "  -"
        print(f"{grp:<16}{ms:>12}({len(sole):>3}){mc:>12}({len(cont):>3})")

    print("\nwithin-repo mean rho (avoids Simpson):")
    for k, v in per_repo_rho.items():
        if v:
            print(f"  {k}: mean={sum(v)/len(v):+.3f}  (n_repos={len(v)}, "
                  f"sign-consistent {max(sum(1 for x in v if x>0), sum(1 for x in v if x<0))}/{len(v)})")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Per-module monthly commit activity from git (for the abandonment outcome).

For each repo: parse `git log --numstat`, bucket each commit into its month
and attribute it to every module (longest-prefix of a touched file) it touched.
Output data/<repo>.activity.json = {module: {"YYYY-MM": commit_count}}.

Abandonment (measured in backtest.py) = a module's commit rate cratering after
an anchor month. Unlike survival mass, this is future-activity based -> not
mechanically tied to ownership share, so the placebo control can isolate a
real departure effect.
"""
import json, os, re, subprocess, sys, glob, time
from collections import defaultdict

BASE = os.path.dirname(__file__)
DATA = os.path.join(BASE, "data")


def _load(path):
    dec = json.JSONDecoder(); s = open(path).read(); i = 0; out = []
    while i < len(s):
        while i < len(s) and s[i] in " \n\t\r": i += 1
        if i >= len(s): break
        o, j = dec.raw_decode(s, i); out.append(o); i = j
    return out


def module_names(analyze_path):
    mods = set()
    for doc in _load(analyze_path):
        for dom in (doc.get("domains") or []):
            for m in (dom.get("module_scores") or []):
                mods.add(m["module"])
    return sorted([m for m in mods if m != "."], key=len, reverse=True)


def assign(relpath, sorted_mods):
    for m in sorted_mods:
        if relpath == m or relpath.startswith(m + "/"):
            return m
    return "."


def process(repo):
    anp = os.path.join(DATA, f"{repo}.analyze.json")
    repo_dir = os.path.join(BASE, repo)
    if not (os.path.exists(anp) and os.path.isdir(repo_dir)):
        return None
    mods = module_names(anp)
    # @<unix_ts> marks a commit; following lines are "added<TAB>deleted<TAB>path"
    p = subprocess.run(
        ["git", "-C", repo_dir, "log", "--no-merges", "--pretty=format:@%at", "--numstat"],
        capture_output=True, text=True, errors="ignore")
    act = defaultdict(lambda: defaultdict(int))   # module -> month -> commit count
    month = None; touched = set()

    def flush():
        for m in touched:
            act[m][month] += 1

    for line in p.stdout.splitlines():
        if line.startswith("@"):
            if month is not None and touched:
                flush()
            ts = int(line[1:])
            g = time.gmtime(ts)
            month = f"{g.tm_year:04d}-{g.tm_mon:02d}"
            touched = set()
        elif line.strip():
            parts = line.split("\t")
            if len(parts) == 3:
                touched.add(assign(parts[2], mods))
    if month is not None and touched:
        flush()

    out = {m: dict(mm) for m, mm in act.items()}
    return out


def main():
    repos = sorted({os.path.basename(p).split(".")[0]
                    for p in glob.glob(os.path.join(DATA, "*.analyze.json"))})
    for repo in repos:
        out = process(repo)
        if out is None:
            print(f"[{repo}] skip"); continue
        json.dump(out, open(os.path.join(DATA, f"{repo}.activity.json"), "w"))
        months = set()
        for mm in out.values():
            months |= set(mm.keys())
        print(f"[{repo}] modules={len(out)}  months={len(months)}  "
              f"total_commits={sum(sum(mm.values()) for mm in out.values())}")


if __name__ == "__main__":
    main()

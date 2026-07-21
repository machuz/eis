#!/usr/bin/env python3
"""Per-module monthly commit counts, straight from git.

This is the outcome substrate, and it is deliberately computed from raw git
rather than from EIS. The outcome ("did anyone commit to this module") must not
share machinery with the predictors (EIS surviving mass), or the correlation is
manufactured. Same predictor-outcome decoupling as survival-backtest.

Writes two files per repo:
  data/<repo>.activity.json = {module: {"YYYY-MM": commit_count}}
  data/<repo>.committers.json = {module: {"YYYY-MM": [author_email, ...]}}

The second one is what makes "how many people could maintain this module" a
*per-module* quantity. Counting repo-wide active authors instead gives every
module at time t the same value, which is a repo-activity proxy confounded with
calendar time — not ownership.

Module assignment = longest matching prefix of a touched path, using the module
list from `eis analyze` so it matches what the timeline reports.

    REPOS=../oss-gravity-map/data/repos python3 activity.py
"""

from __future__ import annotations

import glob
import json
import os
import subprocess
import time
from collections import defaultdict

HERE = os.path.dirname(os.path.abspath(__file__))
DATA = os.environ.get("EIS_BT_DATA", os.path.join(HERE, "data"))
REPOS = os.environ.get("REPOS", os.path.join(HERE, "..", "oss-gravity-map", "data", "repos"))


def load_stream(path):
    dec = json.JSONDecoder()
    s = open(path).read()
    i, out = 0, []
    while i < len(s):
        while i < len(s) and s[i] in " \n\t\r":
            i += 1
        if i >= len(s):
            break
        o, j = dec.raw_decode(s, i)
        out.append(o)
        i = j
    return out


def module_names(analyze_path):
    mods = set()
    for doc in load_stream(analyze_path):
        for dom in doc.get("domains") or []:
            for m in dom.get("module_scores") or []:
                mods.add(m["module"])
    # longest first so the prefix match picks the most specific module
    return sorted((m for m in mods if m != "."), key=len, reverse=True)


def assign(path, mods):
    for m in mods:
        if path == m or path.startswith(m + "/"):
            return m
    return "."


def process(repo):
    analyze = os.path.join(DATA, f"{repo}.analyze.json")
    repo_dir = os.path.join(REPOS, repo)
    if not os.path.exists(analyze) or not os.path.isdir(repo_dir):
        return None
    mods = module_names(analyze)

    p = subprocess.run(
        ["git", "-C", repo_dir, "log", "--no-merges",
         "--pretty=format:@%at%x09%ae", "--numstat"],
        capture_output=True, text=True, errors="ignore")

    act = defaultdict(lambda: defaultdict(int))          # module -> month -> commits
    who = defaultdict(lambda: defaultdict(set))          # module -> month -> {author}
    month, author, touched = None, None, set()

    def flush():
        for m in touched:
            act[m][month] += 1
            who[m][month].add(author)

    for line in p.stdout.splitlines():
        if line.startswith("@"):
            if month is not None and touched:
                flush()
            head = line[1:].split("\t")
            g = time.gmtime(int(head[0]))
            month = f"{g.tm_year:04d}-{g.tm_mon:02d}"
            author = head[1] if len(head) > 1 else "?"
            touched = set()
        elif line.strip():
            parts = line.split("\t")
            if len(parts) == 3:
                touched.add(assign(parts[2], mods))
    if month is not None and touched:
        flush()

    return ({m: dict(mm) for m, mm in act.items()},
            {m: {mo: sorted(s) for mo, s in mm.items()} for m, mm in who.items()})


def main():
    repos = sorted({os.path.basename(p).split(".")[0]
                    for p in glob.glob(os.path.join(DATA, "*.analyze.json"))})
    for repo in repos:
        res = process(repo)
        if res is None:
            print(f"[{repo}] skip (no analyze.json or repo dir)")
            continue
        act, who = res
        json.dump(act, open(os.path.join(DATA, f"{repo}.activity.json"), "w"))
        json.dump(who, open(os.path.join(DATA, f"{repo}.committers.json"), "w"))
        months = {mo for mm in act.values() for mo in mm}
        print(f"[{repo}] modules={len(act)} months={len(months)} "
              f"commits={sum(sum(mm.values()) for mm in act.values())}")


if __name__ == "__main__":
    main()

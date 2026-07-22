#!/usr/bin/env python3
"""Build the abandonment panel from a pre-aggregated CSV instead of `eis timeline`.

Why this exists: on a 13-year monorepo, running `eis timeline --span 1m` locally
is not tractable (analyze alone spawns ~9 concurrent git blames at ~2min each,
1.1GB RSS, before ~160 periods even start). But a SaaS deployment has already
computed exactly that panel and stored it, so the predictors can be read out
instead of recomputed.

Input CSV (no header), one row per (module, period):

    YYYY-MM,n_hold,survival_total,survival_sumsq,survival_max,module_path

`hhi` and `top_share` are recovered from the aggregates:

    hhi       = sum(g_i^2) / (sum g_i)^2
    top_share = max(g_i)   /  sum g_i

so the exporter never has to ship per-author rows.

The outcome still comes from raw `git log` on a local clone (activity.py),
never from the same source as the predictors. That decoupling is the whole
credibility of the design and it survives the change of predictor source.

    python3 panel_csv.py --csv panel.csv --repo-dir /path/to/clone --name react
    # writes <name>.activity.json / <name>.committers.json / <name>.panel.json
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import time
from collections import defaultdict

HERE = os.path.dirname(os.path.abspath(__file__))
DATA = os.environ.get("EIS_BT_DATA", os.path.join(HERE, "data"))


def assign(path, mods):
    for m in mods:
        if path == m or path.startswith(m + "/"):
            return m
    return "."


def git_activity(repo_dir, mods):
    """Per-module monthly commits and committers, from raw git."""
    p = subprocess.run(
        ["git", "-C", repo_dir, "log", "--no-merges",
         "--pretty=format:@%at%x09%ae", "--numstat"],
        capture_output=True, text=True, errors="ignore")
    act = defaultdict(lambda: defaultdict(int))
    who = defaultdict(lambda: defaultdict(set))
    month = author = None
    touched = set()

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
    ap = argparse.ArgumentParser()
    ap.add_argument("--csv", required=True)
    ap.add_argument("--repo-dir", required=True)
    ap.add_argument("--name", required=True)
    args = ap.parse_args()

    panel = []
    for line in open(args.csv):
        line = line.strip()
        if not line:
            continue
        parts = line.split(",", 5)
        if len(parts) != 6:
            continue
        ym, n_hold, tot, sumsq, mx, module = parts
        try:
            tot_f, sumsq_f, mx_f = float(tot), float(sumsq), float(mx)
        except ValueError:
            continue                       # header or garbage row
        if tot_f <= 0:
            continue
        panel.append({
            "month": ym,
            "module": module,
            "n_hold": int(n_hold),
            "survival": tot_f,
            "hhi": sumsq_f / (tot_f * tot_f),
            "top_share": mx_f / tot_f,
        })

    mods = sorted({r["module"] for r in panel if r["module"] != "."},
                  key=len, reverse=True)
    act, who = git_activity(args.repo_dir, mods)

    os.makedirs(DATA, exist_ok=True)
    json.dump(panel, open(os.path.join(DATA, f"{args.name}.panel.json"), "w"))
    json.dump(act, open(os.path.join(DATA, f"{args.name}.activity.json"), "w"))
    json.dump(who, open(os.path.join(DATA, f"{args.name}.committers.json"), "w"))

    months = {r["month"] for r in panel}
    print(f"[{args.name}] panel rows={len(panel)} modules={len(mods)} "
          f"periods={len(months)} git_months={len({mo for mm in act.values() for mo in mm})}")


if __name__ == "__main__":
    main()

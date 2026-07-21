#!/usr/bin/env python3
"""Augment an author predictor CSV with tenure_days and commit_count as of T —
the person-level confounds (age/volume), analogous to module age_days. Cheap: one
`git log` pass <= T, no blame. Lets backtest.py test whether author gravity
predicts survival BEYOND 'senior people have older, larger footprints'.

Usage: add_author_tenure.py --repo <path> --anchor-date YYYY-MM-DD --csv auths.csv
"""
import argparse
import csv
import subprocess
from collections import defaultdict


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo", required=True)
    ap.add_argument("--anchor-date", required=True)
    ap.add_argument("--csv", required=True)
    args = ap.parse_args()

    T = int(subprocess.run(["date", "-j", "-f", "%Y-%m-%d", args.anchor_date, "+%s"],
                           capture_output=True, text=True).stdout.strip() or 0)
    if not T:  # GNU date fallback
        T = int(subprocess.run(["date", "-d", args.anchor_date, "+%s"],
                               capture_output=True, text=True).stdout.strip())

    r = subprocess.run(["git", "-C", args.repo, "log", f"--until={args.anchor_date}",
                        "--no-merges", "--format=%an%x09%at"],
                       capture_output=True, text=True, errors="replace", check=True)
    first = {}
    count = defaultdict(int)
    for line in r.stdout.splitlines():
        if "\t" not in line:
            continue
        name, at = line.rsplit("\t", 1)
        try:
            at = int(at)
        except ValueError:
            continue
        count[name] += 1
        if name not in first or at < first[name]:
            first[name] = at

    rows = list(csv.DictReader(open(args.csv)))
    fields = rows[0].keys() if rows else ["author"]
    out_fields = list(fields) + ["tenure_days", "commit_count"]
    with open(args.csv, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=out_fields)
        w.writeheader()
        for row in rows:
            a = row["author"]
            row["tenure_days"] = (T - first[a]) // 86400 if a in first else 0
            row["commit_count"] = count.get(a, 0)
            w.writerow(row)
    print(f"{args.csv}: tenure/volume added for {len(rows)} authors")


if __name__ == "__main__":
    main()

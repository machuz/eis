#!/usr/bin/env python3
"""Outcome side of the backtest: cohort survival of <=T code to HEAD, from raw git.

Deliberately independent of EIS (uses only `git`), so the predictor (EIS gravity)
and the outcome (git cohort survival) share no machinery — the correlation, if any,
cannot be an artifact of one computation. git-of-theseus style. See README.md.

For a repo and anchor date T:
  sha_T        = last commit with commit-date <= T
  cohort       = every line of every file present at sha_T (author from `git blame sha_T`)
  survives     = that exact line still present at HEAD, via `git blame --reverse sha_T..HEAD`
                 (a cohort line survives iff its reverse-blame commit is the HEAD tip)
  ratio(m)     = survivors(m) / cohort(m)   in [0,1]   (survivors are a SUBSET of cohort
                 by construction — normal + reverse blame run on the SAME sha_T line set)
Author-level is the same keyed on the sha_T blame author (a line counts as an
author's survivor only if their original line reached HEAD — the right notion of
"lasted"). Using one shared line set guarantees ratio in [0,1].

Usage:
  outcome_cohort_survival.py --repo <path> --anchor-date YYYY-MM-DD \
      --modules <predictor mods.csv, for the module-path set> \
      --out-modules out_mods.csv [--with-authors --out-authors out_auths.csv]
"""
import argparse
import csv
import subprocess
import sys
from collections import defaultdict
from datetime import datetime, timezone


def git(repo, *args, check=True):
    return subprocess.run(["git", "-C", repo, *args],
                          capture_output=True, text=True, check=check)


def sha_at(repo, until):
    r = git(repo, "rev-list", "-1", f"--before={until} 23:59:59", "HEAD")
    sha = r.stdout.strip()
    if not sha:
        raise SystemExit(f"no commit <= {until}")
    return sha


def module_of(path, module_paths):
    best = ""
    for mp in module_paths:
        if mp and (path == mp or path.startswith(mp + "/")) and len(mp) > len(best):
            best = mp
    return best or None


def files_at(repo, sha):
    r = git(repo, "ls-tree", "-r", "--name-only", sha)
    return [f for f in r.stdout.splitlines() if f]


def cohort_authors(repo, sha, path):
    """Per-line author of `path` as it existed at sha (the cohort). One entry per
    line, in file order. [] if unblameable."""
    r = git(repo, "blame", "--line-porcelain", "-w", sha, "--", path, check=False)
    if r.returncode != 0:
        return []
    out, author = [], None
    for ln in r.stdout.split("\n"):
        if ln.startswith("author "):
            author = ln[7:]
        elif ln.startswith("\t"):
            out.append(author)
            author = None
    return out


def survives_flags(repo, sha_T, head_sha, path):
    """For each line of `path` at sha_T, True iff it still exists at HEAD, via
    reverse blame over sha_T..HEAD. One entry per line, in the SAME file order as
    cohort_authors(sha_T), so the two zip line-for-line. [] if unblameable."""
    r = git(repo, "blame", "--reverse", "-w", "--line-porcelain",
            f"{sha_T}..{head_sha}", "--", path, check=False)
    if r.returncode != 0:
        return []
    # In reverse blame each line-block's leading 40-hex sha is the LAST commit in
    # which the line survived; the HEAD tip means it lasted all the way.
    flags = []
    for ln in r.stdout.split("\n"):
        if len(ln) >= 40 and all(c in "0123456789abcdef" for c in ln[:40]) and \
           (len(ln) == 40 or ln[40] == " "):
            flags.append(ln[:40] == head_sha)
    return flags


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo", required=True)
    ap.add_argument("--anchor-date", required=True)
    ap.add_argument("--modules", required=True,
                    help="predictor modules.csv — its 'module' column defines the module set")
    ap.add_argument("--out-modules", required=True)
    ap.add_argument("--with-authors", action="store_true")
    ap.add_argument("--out-authors", default="")
    args = ap.parse_args()

    sha_T = sha_at(args.repo, args.anchor_date)
    head_sha = git(args.repo, "rev-parse", "HEAD").stdout.strip()

    with open(args.modules) as f:
        module_paths = [row["module"] for row in csv.DictReader(f)]

    files_T = files_at(args.repo, sha_T)

    cohort_mod = defaultdict(int)
    surv_mod = defaultdict(int)
    cohort_author = defaultdict(int)
    surv_author = defaultdict(int)

    for path in files_T:
        mod = module_of(path, module_paths)
        if mod is None and not args.with_authors:
            continue
        authors = cohort_authors(args.repo, sha_T, path)
        if not authors:
            continue
        flags = survives_flags(args.repo, sha_T, head_sha, path)
        # Reverse blame should return one flag per cohort line; if the counts
        # disagree (rename edge cases), align by the shorter length so survivors
        # can never exceed cohort.
        n = min(len(authors), len(flags)) if flags else 0
        for i in range(len(authors)):
            alive = i < n and flags[i]
            if mod is not None:
                cohort_mod[mod] += 1
                if alive:
                    surv_mod[mod] += 1
            if args.with_authors and authors[i] is not None:
                cohort_author[authors[i]] += 1
                if alive:
                    surv_author[authors[i]] += 1

    with open(args.out_modules, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["module", "cohort", "survivors", "survival_ratio"])
        for mod in sorted(set(cohort_mod) | set(surv_mod)):
            c, s = cohort_mod.get(mod, 0), surv_mod.get(mod, 0)
            w.writerow([mod, c, s, (s / c) if c > 0 else ""])

    if args.with_authors and args.out_authors:
        with open(args.out_authors, "w", newline="") as f:
            w = csv.writer(f)
            w.writerow(["author", "cohort", "survivors", "survival_ratio"])
            for a in sorted(set(cohort_author) | set(surv_author)):
                c, s = cohort_author.get(a, 0), surv_author.get(a, 0)
                w.writerow([a, c, s, (s / c) if c > 0 else ""])

    sys.stderr.write(f"sha_T={sha_T[:10]} files@T={len(files_T)} "
                     f"modules={len(cohort_mod)} authors={len(cohort_author)}\n")


if __name__ == "__main__":
    main()

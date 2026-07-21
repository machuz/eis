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
    # errors="replace": blame output of non-UTF-8 / binary source files must not
    # crash the walk — the author/porcelain markers we parse are ASCII, and the
    # tab-prefixed content lines are only counted, never interpreted.
    return subprocess.run(["git", "-C", repo, *args],
                          capture_output=True, text=True, errors="replace", check=check)


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


def show_linecount(repo, sha, path):
    """Lines of `path` at `sha` — the cohort size (all lines existed at T, so all
    are <=T by construction). Cheap: one `git show`, no blame."""
    r = git(repo, "show", f"{sha}:{path}", check=False)
    if r.returncode != 0:
        return 0
    s = r.stdout
    if not s:
        return 0
    return s.count("\n") + (0 if s.endswith("\n") else 1)


def blame_lines(repo, rev, path):
    """(author, author_time_epoch) per line of `path` at `rev`. [] if unblameable.
    Forward blame — fast. For survivors we filter author-time <= T; because git
    blame keeps a line with its own file (no cross-file moves without -M), a file's
    <=T lines at HEAD are a subset of that same file's lines at sha_T, so per-file
    survivors <= cohort and the ratio stays in [0,1]."""
    r = git(repo, "blame", "--line-porcelain", "-w", rev, "--", path, check=False)
    if r.returncode != 0:
        return []
    out, author, atime = [], None, None
    for ln in r.stdout.split("\n"):
        if ln.startswith("author "):
            author = ln[7:]
        elif ln.startswith("author-time "):
            try:
                atime = int(ln[12:])
            except ValueError:
                atime = None
        elif ln.startswith("\t"):
            out.append((author, atime))
            author, atime = None, None
    return out


def sample(seq, cap):
    """Deterministic even-stride subsample of at most `cap` items (no RNG — keeps
    the backtest reproducible, W-02)."""
    if cap <= 0 or len(seq) <= cap:
        return seq
    step = len(seq) / cap
    return [seq[int(i * step)] for i in range(cap)]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo", required=True)
    ap.add_argument("--anchor-date", required=True)
    ap.add_argument("--modules", required=True,
                    help="predictor modules.csv — its 'module' column defines the module set")
    ap.add_argument("--out-modules", required=True)
    ap.add_argument("--with-authors", action="store_true")
    ap.add_argument("--out-authors", default="")
    ap.add_argument("--sample-files", type=int, default=0,
                    help="cap files blamed (0=all); even-stride subsample for big repos")
    args = ap.parse_args()

    sha_T = sha_at(args.repo, args.anchor_date)
    T_epoch = int(datetime.strptime(args.anchor_date, "%Y-%m-%d")
                  .replace(tzinfo=timezone.utc).timestamp())

    with open(args.modules) as f:
        module_paths = [row["module"] for row in csv.DictReader(f)]

    files_T = files_at(args.repo, sha_T)
    head_files = set(files_at(args.repo, "HEAD"))
    files_T = sample(files_T, args.sample_files)

    cohort_mod = defaultdict(int)
    surv_mod = defaultdict(int)
    cohort_author = defaultdict(int)
    surv_author = defaultdict(int)

    for path in files_T:
        mod = module_of(path, module_paths)
        want_mod = mod is not None
        if not want_mod and not args.with_authors:
            continue
        # cohort = lines at sha_T (cheap show); author cohort needs sha_T blame.
        if args.with_authors:
            shaT = blame_lines(args.repo, sha_T, path)
            cohort_lines = len(shaT)
            for author, _ in shaT:
                if author is not None:
                    cohort_author[author] += 1
        else:
            cohort_lines = show_linecount(args.repo, sha_T, path)
        if want_mod:
            cohort_mod[mod] += cohort_lines
        # survivors = HEAD lines (same file) authored <= T. One forward blame.
        if path in head_files:
            for author, atime in blame_lines(args.repo, "HEAD", path):
                if atime is None or atime > T_epoch:
                    continue
                if want_mod:
                    surv_mod[mod] += 1
                if args.with_authors and author is not None:
                    surv_author[author] += 1

    with open(args.out_modules, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["module", "cohort", "survivors", "survival_ratio"])
        for mod in sorted(set(cohort_mod) | set(surv_mod)):
            c, s = cohort_mod.get(mod, 0), surv_mod.get(mod, 0)
            # clamp: rare cross-file rename (blame without -M) can nudge a module's
            # survivors just past its cohort; the true ratio is capped at 1.0.
            w.writerow([mod, c, s, min(1.0, s / c) if c > 0 else ""])

    if args.with_authors and args.out_authors:
        with open(args.out_authors, "w", newline="") as f:
            w = csv.writer(f)
            w.writerow(["author", "cohort", "survivors", "survival_ratio"])
            for a in sorted(set(cohort_author) | set(surv_author)):
                c, s = cohort_author.get(a, 0), surv_author.get(a, 0)
                w.writerow([a, c, s, min(1.0, s / c) if c > 0 else ""])

    sys.stderr.write(f"sha_T={sha_T[:10]} files@T={len(files_T)} "
                     f"modules={len(cohort_mod)} authors={len(cohort_author)}\n")


if __name__ == "__main__":
    main()

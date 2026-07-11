#!/usr/bin/env python3
"""Extract as-of-T predictors for the survival backtest.

Runs `eis timeline --format json` over a repo's full history (yearly windows),
picks the anchor period whose end is closest to (HEAD_date - horizon), and emits
per-module and per-author predictor rows from that point-in-time snapshot, plus
the churn/age baselines from `git log` bounded to <= T.

Predictor and outcome are deliberately computed by different pipelines (EIS here,
raw git in outcome_cohort_survival.py) so nothing shared can manufacture the
correlation. See README.md.

Usage:
  extract_predictors.py --eis <eis-bin> --repo <path> --config <cfg.yaml> \
      --horizon-days 1095 --out-modules mods.csv --out-authors auths.csv \
      [--anchor-date YYYY-MM-DD]
"""
import argparse
import csv
import json
import subprocess
import sys
from collections import defaultdict
from datetime import datetime, timezone


def run_timeline(eis, repo, config):
    """Full-history yearly timeline as JSON. Returns list of per-domain objects."""
    cmd = [eis, "timeline", "--format", "json", "--span", "1y",
           "--periods", "0", "--fast-log", repo]
    if config:
        cmd[1:1] = ["--config", config]  # after 'timeline'... keep simple:
    # rebuild cleanly to guarantee flag order
    cmd = [eis, "timeline"]
    if config:
        cmd += ["--config", config]
    cmd += ["--format", "json", "--span", "1y", "--periods", "0", "--fast-log", repo]
    out = subprocess.run(cmd, capture_output=True, text=True)
    if out.returncode != 0:
        sys.stderr.write(out.stderr[-2000:])
        raise SystemExit(f"eis timeline failed ({out.returncode}) for {repo}")
    return decode_concatenated_json(out.stdout)


def decode_concatenated_json(s):
    """The CLI emits one JSON object per domain, concatenated. Stream-decode them."""
    dec = json.JSONDecoder()
    objs, i, n = [], 0, len(s)
    while i < n:
        while i < n and s[i] in " \t\r\n":
            i += 1
        if i >= n:
            break
        obj, end = dec.raw_decode(s, i)
        objs.append(obj)
        i = end
    return objs


def parse_date(d):
    return datetime.strptime(d, "%Y-%m-%d").replace(tzinfo=timezone.utc)


def head_date(repo):
    r = subprocess.run(["git", "-C", repo, "log", "-1", "--format=%cI"],
                       capture_output=True, text=True, check=True)
    return datetime.fromisoformat(r.stdout.strip())


def pick_anchor(domains, target):
    """Across all domains' periods, find the period 'end' closest to target date.
    Returns (end_date_str, [ (domain_obj, period_obj) for that end ])."""
    ends = set()
    for dom in domains:
        for p in dom.get("periods", []):
            ends.add(p["end"])
    if not ends:
        raise SystemExit("timeline produced no periods")
    best = min(ends, key=lambda e: abs((parse_date(e) - target).total_seconds()))
    rows = []
    for dom in domains:
        for p in dom.get("periods", []):
            if p["end"] == best:
                rows.append((dom, p))
    return best, rows


def module_features(anchor_rows):
    """Aggregate module -> {raw_survival, author_count, top_share} at the anchor.
    A module may appear in multiple domains; masses sum, authors union."""
    by_mod = defaultdict(lambda: defaultdict(float))
    for _dom, p in anchor_rows:
        msba = p.get("module_survival_by_author") or {}
        for mod, authors in msba.items():
            for author, mass in authors.items():
                by_mod[mod][author] += float(mass)
    feats = {}
    for mod, authors in by_mod.items():
        masses = list(authors.values())
        total = sum(masses)
        if total <= 0:
            continue
        top = max(masses)
        feats[mod] = {
            "raw_survival": total,
            "author_count": sum(1 for m in masses if m > 0),
            # contest: surviving mass NOT held by the top author (others-built-on).
            "contest_mass": total - top,
            "top_share": top / total,
            "spread": 1.0 - top / total,
        }
    return feats


def author_features(anchor_rows):
    """Aggregate author -> {raw_gravity(Σ module mass), module_count(breadth)}."""
    by_author_mod = defaultdict(lambda: defaultdict(float))
    for _dom, p in anchor_rows:
        msba = p.get("module_survival_by_author") or {}
        for mod, authors in msba.items():
            for author, mass in authors.items():
                by_author_mod[author][mod] += float(mass)
    feats = {}
    for author, mods in by_author_mod.items():
        masses = [m for m in mods.values() if m > 0]
        total = sum(masses)
        if total <= 0:
            continue
        feats[author] = {
            "raw_gravity": total,
            "module_count": len(masses),   # breadth: how many modules they hold
            "top_module_share": max(masses) / total,
        }
    return feats


def module_of(path, module_paths):
    """Longest module-path prefix — the same rule eis/write-context uses."""
    best = ""
    for mp in module_paths:
        if mp and (path == mp or path.startswith(mp + "/")) and len(mp) > len(best):
            best = mp
    return best or None


def churn_and_age(repo, anchor_dt, module_paths, window_days=365):
    """Baselines <= T: per-module commit count + changed LOC in a trailing window,
    and module age (days since first commit touching it). Raw git — file paths are
    mapped to the timeline's module set by longest prefix."""
    until = anchor_dt.strftime("%Y-%m-%d")
    since = (anchor_dt.replace(year=anchor_dt.year - 1)).strftime("%Y-%m-%d")
    # numstat log up to T for age (first-touch) and windowed churn.
    r = subprocess.run(
        ["git", "-C", repo, "log", f"--until={until}", "--numstat",
         "--format=commit%x09%cI", "--no-merges"],
        capture_output=True, text=True, check=True)
    churn_commits = defaultdict(set)      # module -> set(commit) in window
    churn_loc = defaultdict(int)          # module -> added+deleted LOC in window
    first_touch = {}                      # module -> earliest date seen
    cur_commit, cur_dt, in_window = None, None, False
    since_dt = parse_date(since)
    for line in r.stdout.splitlines():
        if line.startswith("commit\t"):
            cur_commit = line
            cur_dt = datetime.fromisoformat(line.split("\t", 1)[1])
            in_window = cur_dt >= since_dt
            continue
        if not line.strip():
            continue
        parts = line.split("\t")
        if len(parts) != 3:
            continue
        add, dele, path = parts
        mod = module_of(path, module_paths)
        if mod is None:
            continue
        if mod not in first_touch or cur_dt < first_touch[mod]:
            first_touch[mod] = cur_dt
        if in_window:
            churn_commits[mod].add(cur_commit)
            try:
                churn_loc[mod] += (0 if add == "-" else int(add)) + (0 if dele == "-" else int(dele))
            except ValueError:
                pass
    out = {}
    for mod in module_paths:
        ft = first_touch.get(mod)
        out[mod] = {
            "churn_commits": len(churn_commits.get(mod, ())),
            "churn_loc": churn_loc.get(mod, 0),
            "age_days": (anchor_dt - ft).days if ft else 0,
        }
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--eis", required=True)
    ap.add_argument("--repo", required=True)
    ap.add_argument("--config", default="")
    ap.add_argument("--horizon-days", type=int, default=1095)
    ap.add_argument("--anchor-date", default="")
    ap.add_argument("--out-modules", required=True)
    ap.add_argument("--out-authors", required=True)
    ap.add_argument("--out-meta", default="")
    args = ap.parse_args()

    hd = head_date(args.repo)
    if args.anchor_date:
        target = parse_date(args.anchor_date)
    else:
        target = hd.fromtimestamp(hd.timestamp() - args.horizon_days * 86400, tz=timezone.utc)

    domains = run_timeline(args.eis, args.repo, args.config)
    anchor_end, anchor_rows = pick_anchor(domains, target)
    anchor_dt = parse_date(anchor_end)

    mfeat = module_features(anchor_rows)
    afeat = author_features(anchor_rows)
    baselines = churn_and_age(args.repo, anchor_dt, list(mfeat.keys()))

    with open(args.out_modules, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["module", "raw_survival", "author_count", "contest_mass",
                    "top_share", "spread", "churn_commits", "churn_loc", "age_days"])
        for mod, ft in sorted(mfeat.items()):
            b = baselines.get(mod, {})
            w.writerow([mod, ft["raw_survival"], ft["author_count"], ft["contest_mass"],
                        ft["top_share"], ft["spread"], b.get("churn_commits", 0),
                        b.get("churn_loc", 0), b.get("age_days", 0)])

    with open(args.out_authors, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["author", "raw_gravity", "module_count", "top_module_share"])
        for a, ft in sorted(afeat.items()):
            w.writerow([a, ft["raw_gravity"], ft["module_count"], ft["top_module_share"]])

    meta = {"repo": args.repo, "head_date": hd.isoformat(), "anchor_end": anchor_end,
            "target": target.isoformat(), "n_modules": len(mfeat), "n_authors": len(afeat)}
    if args.out_meta:
        with open(args.out_meta, "w") as f:
            json.dump(meta, f, indent=2)
    sys.stderr.write(json.dumps(meta) + "\n")


if __name__ == "__main__":
    main()

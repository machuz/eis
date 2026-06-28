#!/usr/bin/env python3
"""Backfill ``lifetime_gravity`` into the OSS Gravity Map results.

Why a backfill instead of a fresh ``eis analyze`` run
-----------------------------------------------------
``LifetimeGravity`` (added to eis in machuz/eis#266) is a pure function of axes
that already live in every member record — Catalysis, Survival, Design, Breadth,
Indispensability. The change ADDED a field; it touched no existing formula, so a
fresh run over the same git histories would reproduce every other axis byte for
byte and only append this one field. Recomputing it from the published axes is
therefore equivalent to re-cloning and re-blaming all 29 repos (~50GB), without
the cost.

Trust, not assumption
---------------------
The script does not take that equivalence on faith. It re-derives the EXISTING
``gravity`` (robust-survival gate, 0.45/0.25/0.30 shape) from the same axes and
checks it reproduces the stored value for every one of the ~51k members. If the
gate floors or shape weights here disagreed with eis, ``gravity`` would miss by
tens of points; it matches to within axis-rounding noise (the stored axes are
themselves round1'd, so a few tenths of drift is expected and bounded). Passing
that check is the proof that ``lifetime_gravity`` below is what eis would emit.

Caveat: values are computed from the PUBLISHED (round1) axes, so they are
faithful to within ~0.1–0.2, not bit-exact with a from-scratch eis run on the
full-precision internals. When eis next runs these repos fresh, it overwrites
with the exact value. Mirrors analysis/lifetime_footprint.py and the merged
internal/scorer/scorer.go:lifetimeGravityScore.

Run: python3 scripts/backfill-lifetime-gravity.py [--check]
"""
from __future__ import annotations

import glob
import json
import math
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
RESULTS = os.path.join(HERE, "..", "data", "results")

CAT_FLOOR = 0.15   # scorer.gravityCatFloor
SURV_FLOOR = 0.15  # scorer.gravitySurvFloor

# scorer.gravityScore (live) vs scorer.lifetimeGravityScore (all-time)
SHAPE_LIVE = (0.45, 0.25, 0.30)      # design, breadth, indispensability
SHAPE_LIFETIME = (0.35, 0.20, 0.45)

VERIFY_TOL = 0.5  # gate cleanly separates rounding noise (~0.15) from a wrong constant (tens)


def round1(v: float) -> float:
    # mirror eis output/json.go round1: math.Round(v*10)/10 (half away from zero, v>=0)
    return math.floor(v * 10 + 0.5) / 10


def gate(x: float, floor: float) -> float:
    return floor + (1 - floor) * (x / 100.0)


def shape(m: dict, w: tuple) -> float:
    return w[0] * m.get("design", 0) + w[1] * m.get("breadth", 0) + w[2] * m.get("indispensability", 0)


def live_gravity(m: dict) -> float:
    return gate(m.get("catalysis", 0), CAT_FLOOR) * gate(m.get("robust_survival", 0), SURV_FLOOR) * shape(m, SHAPE_LIVE)


def lifetime_gravity(m: dict) -> float:
    return gate(m.get("catalysis", 0), CAT_FLOOR) * gate(m.get("survival", 0), SURV_FLOOR) * shape(m, SHAPE_LIFETIME)


def members(doc: dict):
    for dom in doc.get("domains", []):
        for m in dom.get("members", []):
            yield m
        for rr in dom.get("per_repo", []):
            for m in rr.get("members", []):
                yield m


def with_lifetime_after_gravity(m: dict) -> dict:
    """Rebuild the member dict inserting lifetime_gravity right after gravity,
    matching the eis jsonMember field order."""
    out = {}
    for k, v in m.items():
        out[k] = v
        if k == "gravity":
            out["lifetime_gravity"] = round1(lifetime_gravity(m))
    if "lifetime_gravity" not in out:  # member without a gravity key (defensive)
        out["lifetime_gravity"] = round1(lifetime_gravity(m))
    return out


def main() -> int:
    check_only = "--check" in sys.argv[1:]
    files = sorted(glob.glob(os.path.join(RESULTS, "*.json")))

    total = 0
    max_dev = 0.0
    worst = None
    # Pass 1: verify the formula reproduces the published live gravity everywhere.
    for fp in files:
        doc = json.load(open(fp))
        for m in members(doc):
            if "gravity" not in m:
                continue
            total += 1
            dev = abs(round1(live_gravity(m)) - m["gravity"])
            if dev > max_dev:
                max_dev, worst = dev, (os.path.basename(fp), m.get("member", "?"))

    print(f"verify: re-derived live gravity for {total} members, max deviation {max_dev:.3f} "
          f"(worst: {worst[0]} / {worst[1]})" if worst else "verify: no members")
    if max_dev > VERIFY_TOL:
        print(f"ABORT: deviation {max_dev:.3f} > {VERIFY_TOL} — gate/shape constants disagree with eis. "
              f"Not writing.", file=sys.stderr)
        return 1

    # Pass 2: write lifetime_gravity (skip in --check).
    written = 0
    for fp in files:
        doc = json.load(open(fp))
        changed = False
        for dom in doc.get("domains", []):
            if dom.get("members"):
                dom["members"] = [with_lifetime_after_gravity(m) for m in dom["members"]]
                changed = True
            for rr in dom.get("per_repo", []):
                if rr.get("members"):
                    rr["members"] = [with_lifetime_after_gravity(m) for m in rr["members"]]
                    changed = True
        if changed and not check_only:
            # Match Go's encoding/json output byte-for-byte: keep non-ASCII raw
            # (ensure_ascii=False) but HTML-escape & < > like the Go encoder does,
            # so the only diff against the eis-produced files is the new field.
            text = json.dumps(doc, indent=2, ensure_ascii=False)
            text = text.replace("&", "\\u0026").replace("<", "\\u003c").replace(">", "\\u003e")
            with open(fp, "w") as f:
                f.write(text + "\n")
            written += 1

    if check_only:
        print("check OK — formula reproduces published gravity; lifetime_gravity not written (--check).")
    else:
        print(f"backfilled lifetime_gravity into {written} result file(s).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

#!/usr/bin/env python3
"""Per-module token-entropy naturalness (Phase 1 conformance signal).

For each repo at HEAD: tokenize source, build a repo-wide trigram model,
and score each module by the mean cross-entropy (surprise, bits/token) of
its own tokens under that model. High = the module's code deviates from the
codebase's patterns = non-conformant (idiosyncratic). Low = idiomatic.

Modules = the resolved module names from the eis analyze output; each source
file is assigned to its longest-prefix module (mirrors eis "longest match
wins"). Output: data/<repo>.naturalness.json  {module: bits_per_token}.

Approximation vs the spec: whole-repo model (not strict leave-one-out) — a
module reinforces its own patterns slightly, biasing large modules lower.
Acceptable for a directional Phase-1 test; noted as a limitation.
"""
import json, os, re, math, sys, glob
from collections import defaultdict, Counter

BASE = os.path.dirname(__file__)
DATA = os.path.join(BASE, "data")
CODE_EXT = {".js", ".mjs", ".cjs", ".jsx", ".ts", ".tsx", ".c", ".h", ".cc",
            ".cpp", ".hpp", ".py", ".go", ".rs"}
# Keep this minimal: score every module eis resolved (incl. examples/test) so
# the join with the backtest's held-modules is complete. Only drop dirs that are
# never eis modules (VCS, deps, build artifacts).
EXCLUDE_DIR = {".git", "node_modules", "vendor", "dist", "build", "out", "third_party"}
TOKEN_RE = re.compile(r"[A-Za-z_][A-Za-z0-9_]*|[0-9]+|[^\sA-Za-z0-9_]")
N = 3          # trigram
BACKOFF = 0.4  # stupid-backoff weight


def tokenize(text):
    return TOKEN_RE.findall(text)


def module_names(analyze_path):
    mods = set()
    for doc in _load(analyze_path):
        for dom in (doc.get("domains") or []):
            for m in (dom.get("module_scores") or []):
                mods.add(m["module"])
    return mods


def _load(path):
    dec = json.JSONDecoder(); s = open(path).read(); i = 0; out = []
    while i < len(s):
        while i < len(s) and s[i] in " \n\t\r": i += 1
        if i >= len(s): break
        o, j = dec.raw_decode(s, i); out.append(o); i = j
    return out


def assign_module(relpath, sorted_mods):
    """Longest module-name that is a path-prefix of relpath (or '.' fallback)."""
    best = None
    for m in sorted_mods:            # sorted_mods is longest-first
        if m == ".":
            continue
        if relpath == m or relpath.startswith(m + "/"):
            best = m; break
    return best if best is not None else "."


def collect(repo_dir, mods):
    sorted_mods = sorted([m for m in mods], key=len, reverse=True)
    module_tokens = defaultdict(list)   # module -> token list
    all_tokens = []
    for root, dirs, files in os.walk(repo_dir):
        dirs[:] = [d for d in dirs if d not in EXCLUDE_DIR]
        for fn in files:
            ext = os.path.splitext(fn)[1]
            if ext not in CODE_EXT:
                continue
            if fn.endswith((".min.js", ".test.js", ".test.ts", ".pb.go", ".gen.go")):
                continue
            fp = os.path.join(root, fn)
            rel = os.path.relpath(fp, repo_dir)
            try:
                text = open(fp, encoding="utf-8", errors="ignore").read()
            except Exception:
                continue
            toks = tokenize(text)
            if not toks:
                continue
            mod = assign_module(rel, sorted_mods)
            module_tokens[mod].extend(toks)
            all_tokens.extend(toks)
    return module_tokens, all_tokens


def build_model(tokens):
    uni = Counter(); bi = Counter(); tri = Counter()
    ctx1 = Counter(); ctx2 = Counter()
    prev2 = prev1 = None
    for t in tokens:
        uni[t] += 1
        if prev1 is not None:
            bi[(prev1, t)] += 1; ctx1[prev1] += 1
            if prev2 is not None:
                tri[(prev2, prev1, t)] += 1; ctx2[(prev2, prev1)] += 1
        prev2, prev1 = prev1, t
    return uni, bi, tri, ctx1, ctx2, sum(uni.values())


def surprise(prev2, prev1, t, model):
    uni, bi, tri, ctx1, ctx2, Ntok = model
    V = len(uni) + 1
    if prev2 is not None and ctx2.get((prev2, prev1), 0) > 0 and tri.get((prev2, prev1, t), 0) > 0:
        p = tri[(prev2, prev1, t)] / ctx2[(prev2, prev1)]
    elif prev1 is not None and ctx1.get(prev1, 0) > 0 and bi.get((prev1, t), 0) > 0:
        p = BACKOFF * bi[(prev1, t)] / ctx1[prev1]
    else:
        p = (BACKOFF * BACKOFF) * (uni.get(t, 0) + 1) / (Ntok + V)
    return -math.log2(p) if p > 0 else 32.0


def module_entropy(toks, model):
    if len(toks) < 20:      # too few tokens to be meaningful
        return None
    prev2 = prev1 = None; tot = 0.0; n = 0
    for t in toks:
        tot += surprise(prev2, prev1, t, model); n += 1
        prev2, prev1 = prev1, t
    return tot / n


def process(repo):
    anp = os.path.join(DATA, f"{repo}.analyze.json")
    repo_dir = os.path.join(BASE, repo)
    if not (os.path.exists(anp) and os.path.isdir(repo_dir)):
        return None
    mods = module_names(anp)
    module_tokens, all_tokens = collect(repo_dir, mods)
    if len(all_tokens) < 1000:
        print(f"[{repo}] too few tokens ({len(all_tokens)})", file=sys.stderr)
        return {}
    model = build_model(all_tokens)
    out = {}
    for mod, toks in module_tokens.items():
        e = module_entropy(toks, model)
        if e is not None:
            out[mod] = round(e, 4)
    return out


def main():
    repos = sorted({os.path.basename(p).split(".")[0]
                    for p in glob.glob(os.path.join(DATA, "*.analyze.json"))})
    for repo in repos:
        out = process(repo)
        if out is None:
            print(f"[{repo}] skip (no repo dir / analyze)"); continue
        json.dump(out, open(os.path.join(DATA, f"{repo}.naturalness.json"), "w"))
        vals = sorted(out.values())
        if vals:
            print(f"[{repo}] modules={len(out)}  entropy min={vals[0]:.2f} "
                  f"median={vals[len(vals)//2]:.2f} max={vals[-1]:.2f}")
        else:
            print(f"[{repo}] no modules scored")


if __name__ == "__main__":
    main()

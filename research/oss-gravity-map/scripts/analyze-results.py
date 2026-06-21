#!/usr/bin/env python3
"""
OSS Gravity Map — Analysis Engine

Compares EIS gravity rankings against ground truth (known architects,
maintainers, release authors) to validate the Software Cosmology model.

IMPORTANT: Gravity scores are normalized per-repository ("per-universe").
Cross-universe comparison requires adjustment for universe size.

Gravity = catGate(Catalysis) × survGate(RobustSurvival) × shape(Design, Breadth, Indispensability)
  (structural influence — both surviving-under-others-pressure AND catalysis are necessary)

Adjusted Gravity = Gravity × ln(N) / ln(N_max)
  (N = engineers in repo, N_max = largest repo — accounts for universe scale)

Outputs:
  1. Per-Universe Champions — Top engineers within each project
  2. Cross-Universe Top 50 — Adjusted gravity ranking across all projects
  3. Hidden Architects — High gravity but NOT in maintainer list
  4. Entropy Fighters — High Debt Cleanup + High Robust Survival
  5. Collapse Risk — Gravity concentration (bus factor proxy)

Usage:
  python3 analyze-results.py [--results-dir DIR] [--ground-truth-dir DIR] [--output-dir DIR]
"""

import json
import math
import os
import sys
import argparse
from pathlib import Path
from dataclasses import dataclass, field, asdict
from typing import Optional


# =============================================================================
# Data structures
# =============================================================================

@dataclass
class EngineerScore:
    """EIS scores for a single engineer in a single repo."""
    author: str
    repo: str
    total: float = 0.0
    production: float = 0.0
    quality: float = 0.0
    survival: float = 0.0
    robust_survival: float = 0.0
    dormant_survival: float = 0.0
    design: float = 0.0
    breadth: float = 0.0
    debt_cleanup: float = 0.0
    indispensability: float = 0.0
    role: str = ""
    role_confidence: float = 0.0
    style: str = ""
    style_confidence: float = 0.0
    state: str = ""
    state_confidence: float = 0.0
    commits: int = 0
    # From EIS v2.13.0: Gravity = catGate(Catalysis)*survGate(RobustSurvival)*shape
    gravity: float = 0.0


@dataclass
class RepoAnalysis:
    """Analysis results for a single repository."""
    name: str
    repo: str
    engineers: list = field(default_factory=list)
    universe_size: int = 0
    # Architect detection
    top_k_overlap: float = 0.0
    architect_recall: float = 0.0
    spearman_rho: float = 0.0
    # Hidden architects
    hidden_architects: list = field(default_factory=list)
    # Entropy fighters
    entropy_fighters: list = field(default_factory=list)
    # Collapse risk
    gravity_concentration: float = 0.0  # top-3 gravity / total gravity
    top_3_gravity_share: float = 0.0


@dataclass
class GlobalGravityEntry:
    """An engineer's gravity across the entire OSS universe."""
    author: str
    repo: str
    gravity: float = 0.0           # local (per-universe) gravity
    adjusted_gravity: float = 0.0  # cross-universe adjusted gravity
    universe_size: int = 0         # number of engineers in this repo
    total: float = 0.0
    role: str = ""
    style: str = ""
    is_known_architect: bool = False
    is_maintainer: bool = False


# =============================================================================
# Loading
# =============================================================================

def load_eis_results(results_dir: str) -> dict[str, list[EngineerScore]]:
    """Load EIS JSON results for all repos."""
    all_results = {}

    for f in Path(results_dir).glob("*.json"):
        repo_name = f.stem
        # Skip timeline files
        if "timeline" in repo_name:
            continue
        try:
            with open(f) as fh:
                data = json.load(fh)
        except (json.JSONDecodeError, IOError) as e:
            print(f"  WARN: could not load {f}: {e}", file=sys.stderr)
            continue

        engineers = []
        domains = data.get("domains", [])
        for domain in domains:
            for member in domain.get("members", []):
                eng = EngineerScore(
                    author=member.get("member", ""),
                    repo=repo_name,
                    # EIS renamed the composite score "total" → "impact"; keep a
                    # fallback so older result JSONs still load.
                    total=member.get("impact", member.get("total", 0)),
                    production=member.get("production", 0),
                    quality=member.get("quality", 0),
                    survival=member.get("survival", 0),
                    robust_survival=member.get("robust_survival", 0),
                    dormant_survival=member.get("dormant_survival", 0),
                    design=member.get("design", 0),
                    breadth=member.get("breadth", 0),
                    debt_cleanup=member.get("debt_cleanup", 0),
                    indispensability=member.get("indispensability", 0),
                    role=member.get("role", ""),
                    role_confidence=member.get("role_confidence", 0),
                    style=member.get("style", ""),
                    style_confidence=member.get("style_confidence", 0),
                    state=member.get("state", ""),
                    state_confidence=member.get("state_confidence", 0),
                    commits=member.get("commits", 0),
                    gravity=member.get("gravity", 0),
                )
                engineers.append(eng)

        all_results[repo_name] = engineers

    return all_results


def load_ground_truth(gt_dir: str) -> dict[str, dict]:
    """Load ground truth JSON for all repos."""
    all_gt = {}
    for f in Path(gt_dir).glob("*.json"):
        repo_name = f.stem
        try:
            with open(f) as fh:
                all_gt[repo_name] = json.load(fh)
        except (json.JSONDecodeError, IOError):
            continue
    return all_gt


def load_known_architects(dataset_path: str) -> dict[str, list[str]]:
    """Load known architects from dataset.yaml."""
    try:
        import yaml
        with open(dataset_path) as f:
            data = yaml.safe_load(f)
    except ImportError:
        print("  WARN: PyYAML not installed, skipping known architects", file=sys.stderr)
        return {}

    architects = {}
    for entry in data.get("tier1", []):
        name = entry.get("name", "")
        architects[name] = entry.get("known_architects", [])
    return architects


def load_user_name_cache(path: str) -> dict[str, str]:
    """Load the GitHub login → display-name cache. Ground truth and known
    architects are GitHub *logins* (e.g. "gaearon") while EIS author names are
    git *display names* (e.g. "Dan Abramov"); this map bridges the two so
    architect recall and known-architect tagging match on either form."""
    try:
        with open(path) as f:
            raw = json.load(f)
    except (json.JSONDecodeError, IOError):
        return {}
    cache = {}
    for login, v in raw.items():
        nm = v if isinstance(v, str) else (v.get("name") if isinstance(v, dict) else None)
        if login and nm:
            cache[login.lower()] = nm
    return cache


def architect_keys(login: str, name_cache: dict[str, str]) -> set[str]:
    """Normalized match keys for one known-architect login: the login itself and
    its resolved display name (if known). Either matching an EIS author counts."""
    keys = {normalize_author(login)}
    nm = name_cache.get(login.lower()) if name_cache else None
    if nm:
        keys.add(normalize_author(nm))
    return keys


def load_alias_maps(configs_dir: str) -> dict[str, dict[str, str]]:
    """Load alias configs to build login→canonical name maps per repo."""
    try:
        import yaml
    except ImportError:
        return {}

    alias_maps = {}
    configs_path = Path(configs_dir)
    for f in configs_path.glob("*.yaml"):
        repo_name = f.stem
        with open(f) as fh:
            config = yaml.safe_load(fh)
        if not config or "aliases" not in config:
            continue
        alias_maps[repo_name] = {}
        for alias_key, canonical in config["aliases"].items():
            alias_maps[repo_name][normalize_author(alias_key)] = canonical
            alias_maps[repo_name][normalize_author(canonical)] = canonical
    return alias_maps


def resolve_gt_name(login: str, repo_name: str, alias_maps: dict) -> str:
    """Resolve a GitHub login to the canonical EIS author name using alias maps."""
    repo_aliases = alias_maps.get(repo_name, {})
    canonical = repo_aliases.get(normalize_author(login))
    if canonical:
        return canonical
    return login


# =============================================================================
# Analysis
# =============================================================================

def normalize_author(author: str) -> str:
    """Normalize git author for matching."""
    return author.lower().strip()


def compute_top_k_overlap(eis_top: list[str], gt_top: list[str], k: int = 10) -> float:
    """Fraction of EIS top-k that appear in ground truth top-k."""
    eis_set = set(normalize_author(a) for a in eis_top[:k])
    gt_set = set(normalize_author(a) for a in gt_top[:k])
    if not gt_set:
        return 0.0
    return len(eis_set & gt_set) / len(gt_set)


def compute_recall(eis_top: list[str], architects: list[str], k: int = 20,
                   name_cache: dict[str, str] = None) -> float:
    """Fraction of known architects appearing in EIS top-k by gravity. Each
    architect (a GitHub login) matches if either its login or its resolved
    display name is among the top-k EIS author names."""
    if not architects:
        return 0.0
    eis_set = set(normalize_author(a) for a in eis_top[:k])
    found = sum(1 for a in architects if architect_keys(a, name_cache or {}) & eis_set)
    return found / len(architects)


def spearman_rank(ranking_a: list[str], ranking_b: list[str]) -> float:
    """Spearman rank correlation between two rankings, over the members they
    share. Ranks are taken WITHIN the common set (dense 0..n-1) — using each
    name's absolute position in its full list instead lets the d² terms exceed
    the n(n²-1) denominator and pushes ρ far outside [-1, 1]."""
    pos_a = {normalize_author(a): i for i, a in enumerate(ranking_a)}
    pos_b = {normalize_author(b): i for i, b in enumerate(ranking_b)}
    common = set(pos_a) & set(pos_b)

    if len(common) < 3:
        return 0.0

    # Re-rank densely within the common set, preserving each list's order.
    order_a = sorted(common, key=lambda c: pos_a[c])
    order_b = sorted(common, key=lambda c: pos_b[c])
    rank_a = {c: i for i, c in enumerate(order_a)}
    rank_b = {c: i for i, c in enumerate(order_b)}

    n = len(common)
    d_sq_sum = sum((rank_a[c] - rank_b[c]) ** 2 for c in common)
    return 1 - (6 * d_sq_sum) / (n * (n**2 - 1))


def find_hidden_architects(engineers: list[EngineerScore],
                           maintainers: list[str],
                           k: int = 20) -> list[EngineerScore]:
    """Find high-gravity engineers NOT in maintainer/contributor top list."""
    maintainer_set = set(normalize_author(m) for m in maintainers)
    sorted_eng = sorted(engineers, key=lambda e: e.gravity, reverse=True)

    hidden = []
    for eng in sorted_eng[:k]:
        if normalize_author(eng.author) not in maintainer_set:
            hidden.append(eng)
    return hidden


def find_entropy_fighters(engineers: list[EngineerScore]) -> list[EngineerScore]:
    """Find engineers with high Debt Cleanup + high Robust Survival.

    Entropy Fighters are the people who fight the second law of thermodynamics
    in code: they clean up debt AND their cleanup survives over time.
    Uses robust_survival (change-pressure-aware) rather than raw survival.
    """
    fighters = []
    for eng in engineers:
        # Use robust_survival when available, fall back to survival
        robust = eng.robust_survival if eng.robust_survival > 0 else eng.survival
        if eng.debt_cleanup >= 50 and robust >= 30:
            fighters.append(eng)
    # Sort by debt_cleanup + robust_survival
    fighters.sort(key=lambda e: e.debt_cleanup + max(e.robust_survival, e.survival), reverse=True)
    return fighters


def compute_gravity_concentration(engineers: list[EngineerScore], top_n: int = 3) -> float:
    """Gravity concentration: top-N gravity / total gravity."""
    if not engineers:
        return 0.0
    sorted_eng = sorted(engineers, key=lambda e: e.gravity, reverse=True)
    total_gravity = sum(e.gravity for e in engineers)
    if total_gravity == 0:
        return 0.0
    top_gravity = sum(e.gravity for e in sorted_eng[:top_n])
    return top_gravity / total_gravity


# =============================================================================
# Per-repo analysis
# =============================================================================

def analyze_repo(repo_name: str,
                 engineers: list[EngineerScore],
                 ground_truth: Optional[dict],
                 known_architects: list[str],
                 alias_maps: dict = None,
                 name_cache: dict = None) -> RepoAnalysis:
    """Full analysis for a single repository."""
    analysis = RepoAnalysis(name=repo_name, repo=repo_name)
    analysis.engineers = engineers
    analysis.universe_size = len(engineers)

    if not engineers:
        return analysis

    sorted_by_gravity = sorted(engineers, key=lambda e: e.gravity, reverse=True)
    eis_top = [e.author for e in sorted_by_gravity]

    if ground_truth:
        _alias_maps = alias_maps or {}
        gt_contributors = [resolve_gt_name(c["login"], repo_name, _alias_maps)
                          for c in ground_truth.get("top_contributors", [])
                          if c.get("type") == "User"]
        maintainers = [resolve_gt_name(m, repo_name, _alias_maps)
                      for m in ground_truth.get("architect_candidates", [])]
        release_authors = [resolve_gt_name(r, repo_name, _alias_maps)
                          for r in ground_truth.get("release_authors", [])]

        analysis.top_k_overlap = compute_top_k_overlap(eis_top, gt_contributors, k=10)
        analysis.spearman_rho = spearman_rank(eis_top[:30], gt_contributors[:30])

        all_known = set(normalize_author(m) for m in (maintainers + release_authors))
        analysis.hidden_architects = find_hidden_architects(
            engineers, list(all_known), k=20
        )

    if known_architects:
        analysis.architect_recall = compute_recall(eis_top, known_architects, k=20,
                                                   name_cache=name_cache)

    analysis.entropy_fighters = find_entropy_fighters(engineers)

    analysis.gravity_concentration = compute_gravity_concentration(engineers, top_n=3)
    total_grav = sum(e.gravity for e in engineers)
    top3_grav = sum(e.gravity for e in sorted_by_gravity[:3])
    analysis.top_3_gravity_share = top3_grav / total_grav if total_grav > 0 else 0

    return analysis


# =============================================================================
# Global gravity map (cross-universe)
# =============================================================================

def build_global_gravity_map(all_results: dict[str, list[EngineerScore]],
                             all_gt: dict[str, dict],
                             known_architects: dict[str, list[str]],
                             alias_maps: dict = None,
                             name_cache: dict = None) -> list[GlobalGravityEntry]:
    """Build cross-project gravity ranking with universe-size adjustment.

    Adjusted Gravity = gravity × ln(N) / ln(N_max)

    This accounts for the fact that being the gravity leader in a 5,000-person
    universe (Kubernetes) is structurally more significant than being the leader
    in a 125-person universe (esbuild), even if both score 100 locally.
    """
    entries = []
    _alias_maps = alias_maps or {}
    _name_cache = name_cache or {}

    # Find largest universe for normalization
    n_max = max(len(engineers) for engineers in all_results.values()) if all_results else 1
    ln_n_max = math.log(n_max) if n_max > 1 else 1.0

    for repo_name, engineers in all_results.items():
        gt = all_gt.get(repo_name, {})
        maintainers = set(normalize_author(resolve_gt_name(m, repo_name, _alias_maps))
                         for m in gt.get("architect_candidates", []))
        # Known architects are logins; expand each to {login, display name} so an
        # EIS author (a display name) can match. Union across all architects.
        repo_architects = set()
        for a in known_architects.get(repo_name, []):
            repo_architects |= architect_keys(a, _name_cache)

        n = len(engineers)
        ln_n = math.log(n) if n > 1 else 0.0
        scale_factor = ln_n / ln_n_max  # 0..1, larger universes → closer to 1

        for eng in engineers:
            # Filter noise: minimum total score
            if eng.total < 20:
                continue

            adjusted = eng.gravity * scale_factor

            entry = GlobalGravityEntry(
                author=eng.author,
                repo=repo_name,
                gravity=eng.gravity,
                adjusted_gravity=adjusted,
                universe_size=n,
                total=eng.total,
                role=eng.role,
                style=eng.style,
                is_known_architect=normalize_author(eng.author) in repo_architects,
                is_maintainer=normalize_author(eng.author) in maintainers,
            )
            entries.append(entry)

    # Sort by ADJUSTED gravity for cross-universe comparison
    entries.sort(key=lambda e: e.adjusted_gravity, reverse=True)
    return entries


# =============================================================================
# Output
# =============================================================================

def write_markdown_report(analyses: list[RepoAnalysis],
                          global_map: list[GlobalGravityEntry],
                          output_dir: str):
    """Write the full analysis report as Markdown."""
    out = Path(output_dir) / "RESULTS.md"

    with open(out, "w") as f:
        f.write("# OSS Gravity Map — Results\n\n")
        f.write("*Generated by EIS (Engineering Impact Score)*\n\n")
        f.write("---\n\n")

        # === Methodology note ===
        f.write("## Methodology\n\n")
        f.write("### What is Gravity?\n\n")
        f.write("In the Software Cosmology model, **Gravity** measures structural influence — ")
        f.write("a structural *shape* passed through two necessary gates, so neither surviving ")
        f.write("code alone nor catalysis alone can mint it:\n\n")
        f.write("```\n")
        f.write("shape    = 0.45·Design + 0.25·Breadth + 0.30·Indispensability\n")
        f.write("catGate  = 0.15 + 0.85·(Catalysis / 100)\n")
        f.write("survGate = 0.15 + 0.85·(RobustSurvival / 100)\n")
        f.write("Gravity  = catGate × survGate × shape\n")
        f.write("```\n\n")
        f.write("- **Design / Breadth / Indispensability**: the structural shape — what you shaped, ")
        f.write("how widely, and how much the surviving codebase leans on it\n")
        f.write("- **survGate (RobustSurvival)**: did the code last *under others' pressure*? ")
        f.write("Robust survival counts only lines surviving in modules where authors OTHER than ")
        f.write("you actively commit — dead-corner or self-churned code does not qualify\n")
        f.write("- **catGate (Catalysis)**: did others build on your surviving foundation? ")
        f.write("The one axis necessarily zero for a solo author, so it witnesses that the ")
        f.write("structure observably leans on someone\n\n")
        f.write("### Cross-Universe Comparison\n\n")
        f.write("Gravity scores are **normalized per-repository** (per-universe). ")
        f.write("A Gravity of 100 in Express (391 engineers) is not directly comparable ")
        f.write("to a Gravity of 77 in Kubernetes (5,217 engineers).\n\n")
        f.write("For cross-universe ranking, we apply a **universe-size adjustment**:\n\n")
        f.write("```\n")
        f.write("Adjusted Gravity = Gravity × ln(N) / ln(N_max)\n")
        f.write("```\n\n")
        f.write("where N = engineers in the repo, N_max = largest repo analyzed. ")
        f.write("This gives appropriate weight to engineers who hold structural influence ")
        f.write("in larger, more complex ecosystems.\n\n")
        f.write("---\n\n")

        # === What We Found ===
        n_projects = len(analyses)
        n_engineers = sum(a.universe_size for a in analyses)
        n_hidden = sum(len(a.hidden_architects) for a in analyses)
        n_entropy = sum(len(a.entropy_fighters) for a in analyses)
        f.write("## What We Found\n\n")
        f.write(f"We pointed the Git Telescope at {n_projects} of the most built-upon open-source projects — ")
        f.write("React, Kubernetes, Terraform, Redis, Rust, ClickHouse, and the rest — ")
        f.write(f"and read the structure left behind by {n_engineers:,} engineers.\n\n")
        f.write("The question was never who commits the most. It was quieter: when thousands of people pass ")
        f.write("through a codebase over a decade, whose structure does the rest of the system end up leaning on — ")
        f.write("and does it still hold once others start pressing on it?\n\n")
        f.write("Under the two honest gates — code that survives where *others* keep committing, and a foundation ")
        f.write("others actually build on — the names that rise are quiet ones once you see them. ")
        f.write("Jose Valim (Phoenix), Taylor Otwell (Laravel), Donny / kdy1 (swc), fisker Cheung (Prettier), ")
        f.write("Ritchie Vink (Polars), Josh Goldberg (ESLint): their gravity saturates their universe. ")
        f.write("Not because they are famous, and not because they were there first — being first earns nothing ")
        f.write("once your share erodes — but because the surviving shape of the system still rests on them ")
        f.write("while everyone else keeps editing around it.\n\n")
        f.write("The gates also stay silent where founding alone would have flattered. A founder whose modules ")
        f.write("no one else contests reads as quiet, not central: the telescope cannot see influence that no one ")
        f.write("leans on. Sebastian Markbåge sits at the top of React's gravity, yet the data calls him a Cleaner — ")
        f.write("more of his energy holds the structure intact than adds new surface. That is not a demotion. ")
        f.write("It is the instrument declining to confuse ownership with gravity.\n\n")
        f.write(f"And then there are the {n_hidden} engineers the map surfaces who never appear in a maintainer ")
        f.write("list — the **Hidden Architects**. They give no talks and hold no titles, yet the field lines run ")
        f.write("straight through their code: the part written years ago that the system still rests on today.\n\n")
        f.write(f"Alongside them, {n_entropy} **Entropy Fighters** — high on both Debt Cleanup and robust survival — ")
        f.write("quietly hold back the second law in code that would otherwise drift.\n\n")
        f.write("This map is not a leaderboard. It leaves a question open: once the noise of activity is filtered ")
        f.write("out, what does a system actually lean on — and would anyone notice if that quietly left?\n\n")
        f.write("---\n\n")

        # === Per-Universe Champions ===
        f.write("## Per-Universe Champions\n\n")
        f.write("Each project is its own universe with independently normalized scores. ")
        f.write("These are the gravity leaders within their respective ecosystems.\n\n")

        for a in sorted(analyses, key=lambda x: x.universe_size, reverse=True):
            if not a.engineers:
                continue
            sorted_eng = sorted(a.engineers, key=lambda e: e.gravity, reverse=True)
            champion = sorted_eng[0] if sorted_eng else None
            if not champion:
                continue
            f.write(f"### {a.name} ({a.universe_size:,} engineers)\n\n")
            f.write(f"| Rank | Author | Gravity | Design | Breadth | Indisp. | Role | Style |\n")
            f.write(f"|------|--------|---------|--------|---------|---------|------|-------|\n")
            for i, eng in enumerate(sorted_eng[:5]):
                f.write(f"| {i+1} | {eng.author} | {eng.gravity:.1f} | {eng.design:.0f} | {eng.breadth:.0f} | {eng.indispensability:.0f} | {eng.role} | {eng.style} |\n")
            f.write("\n")

            if a.entropy_fighters:
                f.write(f"*Entropy Fighters*: ")
                names = [f"**{e.author}** (debt={e.debt_cleanup:.0f}, robust_surv={e.robust_survival:.0f})"
                        for e in a.entropy_fighters[:3]]
                f.write(", ".join(names))
                f.write("\n\n")

            f.write(f"Gravity Concentration (top-3): {a.gravity_concentration:.1%} · ")
            f.write(f"Top-10 Overlap with GitHub: {a.top_k_overlap:.0%}\n\n")
            f.write("---\n\n")

        # === Cross-Universe Adjusted Top 50 ===
        f.write("## Cross-Universe Top 50 (Adjusted Gravity)\n\n")
        f.write("> Scores adjusted by `ln(universe_size) / ln(max_universe_size)` ")
        f.write("to account for the structural weight of larger ecosystems.\n\n")
        f.write("| Rank | Engineer | Project | Universe | Gravity (local) | Adjusted | Role | Status |\n")
        f.write("|------|----------|---------|----------|-----------------|----------|------|--------|\n")
        for i, entry in enumerate(global_map[:50]):
            status = "Known" if entry.is_known_architect else ("Maintainer" if entry.is_maintainer else "**Hidden**")
            f.write(f"| {i+1} | {entry.author} | {entry.repo} | {entry.universe_size:,} | "
                    f"{entry.gravity:.1f} | {entry.adjusted_gravity:.1f} | {entry.role} | {status} |\n")
        f.write("\n---\n\n")

        # === Aggregate statistics ===
        f.write("## Aggregate Validation Metrics\n\n")
        valid = [a for a in analyses if a.engineers]
        if valid:
            avg_overlap = sum(a.top_k_overlap for a in valid) / len(valid)
            avg_recall = sum(a.architect_recall for a in valid if a.architect_recall > 0)
            recall_count = sum(1 for a in valid if a.architect_recall > 0)
            avg_recall = avg_recall / recall_count if recall_count > 0 else 0
            avg_spearman = sum(a.spearman_rho for a in valid) / len(valid)
            avg_concentration = sum(a.gravity_concentration for a in valid) / len(valid)

            f.write(f"- **Avg Top-10 Overlap**: {avg_overlap:.1%}\n")
            f.write(f"- **Avg Architect Recall (Tier 1)**: {avg_recall:.1%}\n")
            f.write(f"- **Avg Spearman ρ**: {avg_spearman:.3f}\n")
            f.write(f"- **Avg Gravity Concentration**: {avg_concentration:.1%}\n")
            f.write(f"- **Projects analyzed**: {len(valid)}\n")
            total_engineers = sum(len(a.engineers) for a in valid)
            f.write(f"- **Total engineers**: {total_engineers:,}\n")
            total_hidden = sum(len(a.hidden_architects) for a in valid)
            f.write(f"- **Hidden architects found**: {total_hidden}\n")
            total_entropy = sum(len(a.entropy_fighters) for a in valid)
            f.write(f"- **Entropy fighters found**: {total_entropy}\n")
            f.write(f"- **Largest universe**: {max(a.universe_size for a in valid):,} engineers\n")
            f.write(f"- **Smallest universe**: {min(a.universe_size for a in valid):,} engineers\n")

        f.write("\n")

    print(f"Report written to {out}")


def write_json_output(analyses: list[RepoAnalysis],
                      global_map: list[GlobalGravityEntry],
                      output_dir: str):
    """Write structured JSON output for further analysis."""
    out = Path(output_dir) / "results.json"

    data = {
        "methodology": {
            "gravity_formula": "Indispensability * 0.40 + Breadth * 0.30 + Design * 0.30",
            "adjusted_gravity_formula": "gravity * ln(universe_size) / ln(max_universe_size)",
            "note": "Gravity is normalized per-repo. Use adjusted_gravity for cross-repo comparison."
        },
        "global_gravity_map": [asdict(e) for e in global_map[:100]],
        "per_repo": {}
    }

    for a in analyses:
        if not a.engineers:
            continue
        sorted_eng = sorted(a.engineers, key=lambda e: e.gravity, reverse=True)
        data["per_repo"][a.name] = {
            "universe_size": a.universe_size,
            "top_k_overlap": a.top_k_overlap,
            "architect_recall": a.architect_recall,
            "spearman_rho": a.spearman_rho,
            "gravity_concentration": a.gravity_concentration,
            "engineer_count": len(a.engineers),
            "top_20": [asdict(e) for e in sorted_eng[:20]],
            "hidden_architects": [asdict(e) for e in a.hidden_architects[:10]],
            "entropy_fighters": [asdict(e) for e in a.entropy_fighters[:10]],
        }

    with open(out, "w") as f:
        json.dump(data, f, indent=2)

    print(f"JSON output written to {out}")


# =============================================================================
# Main
# =============================================================================

def main():
    parser = argparse.ArgumentParser(description="OSS Gravity Map Analysis")
    parser.add_argument("--results-dir", default="data/results",
                       help="Directory with EIS JSON results")
    parser.add_argument("--ground-truth-dir", default="data/ground-truth",
                       help="Directory with ground truth JSON")
    parser.add_argument("--dataset", default="dataset.yaml",
                       help="Dataset YAML with known architects")
    parser.add_argument("--output-dir", default="analysis",
                       help="Output directory")
    parser.add_argument("--configs-dir", default="configs",
                       help="Directory with per-repo eis.yaml configs")
    parser.add_argument("--alias-map", default="data/alias-map.json",
                       help="Path to alias-map.json (from build-alias-map.py)")
    parser.add_argument("--user-cache", default="data/github-users-cache.json",
                       help="GitHub login → display-name cache (bridges ground-truth logins to EIS author names)")
    args = parser.parse_args()

    print("=== OSS Gravity Map Analysis ===")
    print()

    # Load data
    print("Loading EIS results...")
    all_results = load_eis_results(args.results_dir)
    print(f"  Loaded {len(all_results)} repos")

    print("Loading ground truth...")
    all_gt = load_ground_truth(args.ground_truth_dir)
    print(f"  Loaded {len(all_gt)} repos")

    print("Loading known architects...")
    known_architects = load_known_architects(args.dataset)
    print(f"  Loaded architects for {len(known_architects)} Tier 1 repos")

    print("Loading alias maps...")
    alias_maps = load_alias_maps(args.configs_dir)
    print(f"  Loaded aliases for {len(alias_maps)} repos")

    name_cache = load_user_name_cache(args.user_cache)
    print(f"  Loaded {len(name_cache)} GitHub login→name mappings")

    # Merge alias-map.json (from build-alias-map.py) if it exists
    if os.path.exists(args.alias_map):
        print(f"Loading alias map from {args.alias_map}...")
        try:
            with open(args.alias_map) as f:
                external_aliases = json.load(f)
            merged_count = 0
            for repo_name, mappings in external_aliases.items():
                if repo_name not in alias_maps:
                    alias_maps[repo_name] = {}
                for login, canonical_name in mappings.items():
                    norm_login = normalize_author(login)
                    if norm_login not in alias_maps[repo_name]:
                        alias_maps[repo_name][norm_login] = canonical_name
                        alias_maps[repo_name][normalize_author(canonical_name)] = canonical_name
                        merged_count += 1
            print(f"  Merged {merged_count} aliases from alias map")
        except (json.JSONDecodeError, IOError) as e:
            print(f"  WARN: could not load alias map: {e}", file=sys.stderr)
    else:
        print(f"  No alias map found at {args.alias_map} (run build-alias-map.py to generate)")

    if not all_results:
        print("\nERROR: No EIS results found. Run analyze-repos.sh first.")
        sys.exit(1)

    # Per-repo analysis
    print("\nAnalyzing repos...")
    analyses = []
    for repo_name, engineers in sorted(all_results.items()):
        gt = all_gt.get(repo_name)
        architects = known_architects.get(repo_name, [])
        analysis = analyze_repo(repo_name, engineers, gt, architects, alias_maps, name_cache)
        analyses.append(analysis)
        n_eng = len(engineers)
        n_hidden = len(analysis.hidden_architects)
        n_entropy = len(analysis.entropy_fighters)
        print(f"  {repo_name}: {n_eng} engineers, "
              f"overlap={analysis.top_k_overlap:.0%}, "
              f"recall={analysis.architect_recall:.0%}, "
              f"hidden={n_hidden}, entropy_fighters={n_entropy}, "
              f"concentration={analysis.gravity_concentration:.0%}")

    # Global gravity map
    print("\nBuilding global gravity map (with universe-size adjustment)...")
    global_map = build_global_gravity_map(all_results, all_gt, known_architects, alias_maps, name_cache)
    print(f"  Total entries: {len(global_map)}")
    if global_map:
        print(f"  Top adjusted: {global_map[0].author} ({global_map[0].repo}) = {global_map[0].adjusted_gravity:.1f}")

    # Output
    print("\nWriting reports...")
    os.makedirs(args.output_dir, exist_ok=True)
    write_markdown_report(analyses, global_map, args.output_dir)
    write_json_output(analyses, global_map, args.output_dir)

    print("\nDone!")


if __name__ == "__main__":
    main()

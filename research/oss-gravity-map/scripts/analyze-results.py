#!/usr/bin/env python3
"""
OSS Gravity Map — Analysis Engine

Compares EIS gravity rankings against ground truth (known architects,
maintainers, release authors) to validate the Software Cosmology model.

Outputs:
  1. Architect Detection — Top-k overlap, recall, Spearman correlation
  2. Hidden Architects — High EIS gravity but NOT in maintainer list
  3. Entropy Fighters — High Debt Cleanup + High Robust Survival
  4. Collapse Risk — Gravity concentration (bus factor proxy)
  5. Global Gravity Map — Cross-project ranking

Usage:
  python3 analyze-results.py [--results-dir DIR] [--ground-truth-dir DIR] [--output-dir DIR]
"""

import json
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
    # Computed
    gravity: float = 0.0  # survival + design + debt_cleanup (structural influence)


@dataclass
class RepoAnalysis:
    """Analysis results for a single repository."""
    name: str
    repo: str
    engineers: list = field(default_factory=list)
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
    gravity: float = 0.0
    total: float = 0.0
    role: str = ""
    style: str = ""
    is_known_architect: bool = False
    is_maintainer: bool = False


# =============================================================================
# Loading
# =============================================================================

def load_eis_results(results_dir: str) -> dict[str, list[EngineerScore]]:
    """Load EIS JSON results for all repos.

    EIS JSON format (flat per member):
    {
      "domains": [{
        "name": "...",
        "members": [{
          "rank": 1,
          "member": "Author Name",
          "production": 50.0,
          "quality": 80.0,
          "survival": 60.0,
          "design": 40.0,
          "breadth": 30.0,
          "debt_cleanup": 20.0,
          "indispensability": 10.0,
          "gravity": 45.0,
          "total": 55.0,
          "role": "Architect",
          "role_confidence": 0.8,
          "style": "Builder",
          "style_confidence": 0.6,
          "state": "Active",
          "state_confidence": 0.9
        }]
      }]
    }
    """
    all_results = {}

    for f in Path(results_dir).glob("*.json"):
        repo_name = f.stem
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
                    total=member.get("total", 0),
                    production=member.get("production", 0),
                    quality=member.get("quality", 0),
                    survival=member.get("survival", 0),
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
                    gravity=member.get("gravity", 0),
                )
                # If gravity not in output, compute it
                if eng.gravity == 0 and (eng.survival + eng.design + eng.debt_cleanup) > 0:
                    eng.gravity = eng.survival + eng.design + eng.debt_cleanup
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
        # Fallback: manual parse for simple structure
        print("  WARN: PyYAML not installed, skipping known architects", file=sys.stderr)
        return {}

    architects = {}
    for entry in data.get("tier1", []):
        name = entry.get("name", "")
        architects[name] = entry.get("known_architects", [])
    return architects


# =============================================================================
# Analysis
# =============================================================================

def normalize_author(author: str) -> str:
    """Normalize git author to GitHub-like username for matching."""
    # EIS uses git log author names, ground truth uses GitHub logins
    # This is a known limitation — we do best-effort matching
    return author.lower().strip()


def compute_top_k_overlap(eis_top: list[str], gt_top: list[str], k: int = 10) -> float:
    """Fraction of EIS top-k that appear in ground truth top-k."""
    eis_set = set(normalize_author(a) for a in eis_top[:k])
    gt_set = set(normalize_author(a) for a in gt_top[:k])
    if not gt_set:
        return 0.0
    return len(eis_set & gt_set) / len(gt_set)


def compute_recall(eis_top: list[str], architects: list[str], k: int = 20) -> float:
    """How many known architects appear in EIS top-k."""
    eis_set = set(normalize_author(a) for a in eis_top[:k])
    arch_set = set(normalize_author(a) for a in architects)
    if not arch_set:
        return 0.0
    return len(eis_set & arch_set) / len(arch_set)


def spearman_rank(ranking_a: list[str], ranking_b: list[str]) -> float:
    """Compute Spearman rank correlation between two rankings."""
    # Only compare engineers present in both lists
    set_a = set(normalize_author(a) for a in ranking_a)
    set_b = set(normalize_author(a) for a in ranking_b)
    common = set_a & set_b

    if len(common) < 3:
        return 0.0

    rank_a = {normalize_author(a): i for i, a in enumerate(ranking_a)}
    rank_b = {normalize_author(b): i for i, b in enumerate(ranking_b)}

    n = len(common)
    d_sq_sum = sum((rank_a[c] - rank_b[c]) ** 2 for c in common)

    # Spearman: 1 - 6*sum(d^2) / (n*(n^2-1))
    if n * (n**2 - 1) == 0:
        return 0.0
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
    """Find engineers with high Debt Cleanup + high Robust Survival."""
    fighters = []
    for eng in engineers:
        if eng.debt_cleanup >= 50 and eng.survival >= 40:
            fighters.append(eng)
    # Sort by debt_cleanup + survival
    fighters.sort(key=lambda e: e.debt_cleanup + e.survival, reverse=True)
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
                 known_architects: list[str]) -> RepoAnalysis:
    """Full analysis for a single repository."""
    analysis = RepoAnalysis(name=repo_name, repo=repo_name)
    analysis.engineers = engineers

    if not engineers:
        return analysis

    # Sort by gravity
    sorted_by_gravity = sorted(engineers, key=lambda e: e.gravity, reverse=True)
    eis_top = [e.author for e in sorted_by_gravity]

    # Ground truth comparisons
    if ground_truth:
        gt_contributors = [c["login"] for c in ground_truth.get("top_contributors", [])
                          if c.get("type") == "User"]
        maintainers = ground_truth.get("architect_candidates", [])
        release_authors = ground_truth.get("release_authors", [])

        # Top-k overlap (EIS gravity top-10 vs GitHub commit top-10)
        analysis.top_k_overlap = compute_top_k_overlap(eis_top, gt_contributors, k=10)

        # Spearman correlation
        analysis.spearman_rho = spearman_rank(eis_top[:30], gt_contributors[:30])

        # Hidden architects
        all_known = set(normalize_author(m) for m in (maintainers + release_authors))
        analysis.hidden_architects = find_hidden_architects(
            engineers, list(all_known), k=20
        )

    # Architect recall (against manually curated list)
    if known_architects:
        analysis.architect_recall = compute_recall(eis_top, known_architects, k=20)

    # Entropy fighters
    analysis.entropy_fighters = find_entropy_fighters(engineers)

    # Collapse risk
    analysis.gravity_concentration = compute_gravity_concentration(engineers, top_n=3)
    total_grav = sum(e.gravity for e in engineers)
    top3_grav = sum(e.gravity for e in sorted_by_gravity[:3])
    analysis.top_3_gravity_share = top3_grav / total_grav if total_grav > 0 else 0

    return analysis


# =============================================================================
# Global gravity map
# =============================================================================

def build_global_gravity_map(all_results: dict[str, list[EngineerScore]],
                             all_gt: dict[str, dict],
                             known_architects: dict[str, list[str]]) -> list[GlobalGravityEntry]:
    """Build the cross-project gravity ranking."""
    entries = []

    for repo_name, engineers in all_results.items():
        gt = all_gt.get(repo_name, {})
        maintainers = set(normalize_author(m) for m in gt.get("architect_candidates", []))
        repo_architects = set(normalize_author(a) for a in known_architects.get(repo_name, []))

        for eng in engineers:
            entry = GlobalGravityEntry(
                author=eng.author,
                repo=repo_name,
                gravity=eng.gravity,
                total=eng.total,
                role=eng.role,
                style=eng.style,
                is_known_architect=normalize_author(eng.author) in repo_architects,
                is_maintainer=normalize_author(eng.author) in maintainers,
            )
            entries.append(entry)

    entries.sort(key=lambda e: e.gravity, reverse=True)
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

        # === Global Gravity Map ===
        f.write("## Top 50 OSS Gravity Engineers\n\n")
        f.write("| Rank | Engineer | Project | Gravity | Total | Role | Known Architect? |\n")
        f.write("|------|----------|---------|---------|-------|------|------------------|\n")
        for i, entry in enumerate(global_map[:50]):
            known = "Yes" if entry.is_known_architect else ("Maintainer" if entry.is_maintainer else "**Hidden**")
            f.write(f"| {i+1} | {entry.author} | {entry.repo} | {entry.gravity:.1f} | {entry.total:.1f} | {entry.role} | {known} |\n")
        f.write("\n---\n\n")

        # === Per-repo summaries ===
        f.write("## Per-Project Analysis\n\n")

        for a in analyses:
            if not a.engineers:
                continue

            f.write(f"### {a.name}\n\n")

            # Metrics
            f.write(f"- **Top-10 Overlap**: {a.top_k_overlap:.1%}\n")
            f.write(f"- **Architect Recall**: {a.architect_recall:.1%}\n")
            f.write(f"- **Spearman ρ**: {a.spearman_rho:.3f}\n")
            f.write(f"- **Gravity Concentration (top-3)**: {a.gravity_concentration:.1%}\n")
            f.write(f"- **Total engineers analyzed**: {len(a.engineers)}\n\n")

            # Top 10 by gravity
            sorted_eng = sorted(a.engineers, key=lambda e: e.gravity, reverse=True)
            f.write("**Top 10 Gravity Engineers:**\n\n")
            f.write("| Rank | Author | Gravity | Surv | Design | Debt | Role | Style |\n")
            f.write("|------|--------|---------|------|--------|------|------|-------|\n")
            for i, eng in enumerate(sorted_eng[:10]):
                f.write(f"| {i+1} | {eng.author} | {eng.gravity:.1f} | {eng.survival:.0f} | {eng.design:.0f} | {eng.debt_cleanup:.0f} | {eng.role} | {eng.style} |\n")
            f.write("\n")

            # Hidden architects
            if a.hidden_architects:
                f.write(f"**Hidden Architects ({len(a.hidden_architects)}):**\n\n")
                for eng in a.hidden_architects[:5]:
                    f.write(f"- **{eng.author}** — gravity {eng.gravity:.1f}, role: {eng.role}, style: {eng.style}\n")
                f.write("\n")

            # Entropy fighters
            if a.entropy_fighters:
                f.write(f"**Entropy Fighters ({len(a.entropy_fighters)}):**\n\n")
                for eng in a.entropy_fighters[:5]:
                    f.write(f"- **{eng.author}** — debt_cleanup: {eng.debt_cleanup:.0f}, survival: {eng.survival:.0f}\n")
                f.write("\n")

            f.write("---\n\n")

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

        f.write("\n")

    print(f"Report written to {out}")


def write_json_output(analyses: list[RepoAnalysis],
                      global_map: list[GlobalGravityEntry],
                      output_dir: str):
    """Write structured JSON output for further analysis."""
    out = Path(output_dir) / "results.json"

    data = {
        "global_gravity_map": [asdict(e) for e in global_map[:100]],
        "per_repo": {}
    }

    for a in analyses:
        if not a.engineers:
            continue
        sorted_eng = sorted(a.engineers, key=lambda e: e.gravity, reverse=True)
        data["per_repo"][a.name] = {
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

    if not all_results:
        print("\nERROR: No EIS results found. Run analyze-repos.sh first.")
        sys.exit(1)

    # Per-repo analysis
    print("\nAnalyzing repos...")
    analyses = []
    for repo_name, engineers in sorted(all_results.items()):
        gt = all_gt.get(repo_name)
        architects = known_architects.get(repo_name, [])
        analysis = analyze_repo(repo_name, engineers, gt, architects)
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
    print("\nBuilding global gravity map...")
    global_map = build_global_gravity_map(all_results, all_gt, known_architects)
    print(f"  Total entries: {len(global_map)}")

    # Output
    print("\nWriting reports...")
    os.makedirs(args.output_dir, exist_ok=True)
    write_markdown_report(analyses, global_map, args.output_dir)
    write_json_output(analyses, global_map, args.output_dir)

    print("\nDone!")


if __name__ == "__main__":
    main()

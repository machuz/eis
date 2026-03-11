---
title: Measuring Engineering Impact from Git History Alone
published: true
description: A 7-axis scoring model that quantifies engineer impact using nothing but git log and git blame. Code survival, debt cleanup, bus factor — all from data you already have.
tags: engineering, productivity, git, management
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/engineering-impact-framework-diagram-fixed.png
---

*Why commit counts, PR counts, and lines of code fail to capture real engineering strength*

As engineering teams grow, one question becomes unavoidable:

**How do you evaluate engineering strength using observable signals instead of intuition?**

Most organizations fall back on weak proxies.

- Commit count
- Lines of code
- Number of PRs

All of these are easy to measure — and easy to misinterpret.

A typo fix and a system-wide architectural change both count as "one PR".
A generated lockfile can add thousands of lines.
Commit habits vary wildly between engineers.

Yet inside every team, people still have a sense of who the strongest engineers are.

> "This person writes code that lasts."
> "That person touches everything but somehow nothing improves."

Those intuitions exist, but they are rarely measurable.

I wanted a model that uses **only git history** to approximate real technical influence in a codebase.

Internally, I sometimes jokingly call it an engineer's **"combat power."**

But what it actually measures is something more precise:

> **observable technical impact recorded in the codebase itself**

This article describes that model.

---

## The Core Idea: Code That Survives Matters Most

The strongest engineers don't just write code.

They write code that **continues to exist months later** without needing to be rewritten.

So the most important signal in this model is:

**Code Survival**

But even survival must be handled carefully.

Raw git blame favors early contributors.
Someone who wrote a lot of code three years ago may dominate blame history even if they haven't contributed since.

To fix this, the model applies **time-decayed survival**.

Recent code counts far more than ancient code.

Example decay weights:

| Age | Weight |
|---|---|
| 7 days | 0.96 |
| 30 days | 0.85 |
| 90 days | 0.61 |
| 180 days | 0.37 |
| 1 year | 0.13 |
| 2 years | 0.02 |

![Time-decayed Survival Weight curve showing exponential decay over 730 days](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/survival-decay-curve.png)
*Figure 1. Time-decayed survival gives much more weight to recently written code than legacy code that simply remains untouched.*

This allows the model to approximate:

> **who is currently writing durable code**

rather than who wrote the most code historically.

---

## The 7 Axes of Engineering Impact

![Framework overview: Git History flows through 7 signals into Engineering Impact Score](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/engineering-impact-framework-diagram-fixed.png)
*Figure 2. The Engineering Impact Score aggregates seven observable signals derived from git history.*

The model evaluates engineers across seven signals.

| Axis | Weight | What it captures |
|---|---|---|
| Production | 15% | Volume of code changes |
| Quality | 10% | Low rate of fix/revert commits |
| Survival | **25%** | Code that still exists today (time-decayed) |
| Design | 20% | Contributions to architecture |
| Breadth | 10% | Number of repositories touched |
| Debt Cleanup | 15% | Fixing issues created by others |
| Indispensability | 5% | Bus-factor risk |

The highest weight goes to **Survival**, because it reflects **design durability**.

---

## Production: Measure Changes, Not Commits

Commit counts are unreliable.

Some engineers make large commits.
Others split work into many small commits.

Instead we measure:

```
insertions + deletions
```

Certain files must be excluded:

- generated code
- dependency lockfiles
- Swagger docs
- mock files

Otherwise automated changes distort the numbers.

Production measures **execution throughput**, but it should never be used alone.

---

## Quality: The Fix Ratio

A rough proxy for first-pass correctness.

```
quality = 100 - fix_ratio
fix_ratio = fix_commits / total_commits
```

Fix commits are detected using patterns like:

- `fix`, `revert`, `hotfix`
- Language-specific keywords (e.g. `修正` in Japanese teams)

This metric is imperfect.

Large refactors and proactive improvements sometimes appear as "fixes", so **Quality is intentionally weighted lower**.

---

## Survival: The Most Important Metric

To measure survival, we examine git blame lines and apply exponential time decay.

```python
import math
from collections import defaultdict

tau = 180  # days — weight ≈ 0.37 at 6 months

weighted_survival = defaultdict(float)
for line in blame_lines:
    days_alive = (now - line.committer_time).days
    weight = math.exp(-days_alive / tau)
    weighted_survival[line.author] += weight
```

Each surviving line contributes a decayed weight to its author.

This captures something powerful:

> Engineers whose code remains stable over time accumulate high survival scores.

Engineers whose code is constantly rewritten do not.

---

## Design Influence

Architectural work tends to appear in specific areas of a codebase.

**Backend:**
- repository interfaces, domain services
- routers, middleware
- dependency injection

**Frontend:**
- core logic, shared stores
- hooks, type systems

Frequent commits in these areas signal **architectural involvement**.

Design influence doesn't measure whether a design decision was correct.

It measures **who participates in shaping the system structure**.

---

## Breadth: Operating Across the System

Breadth counts how many repositories an engineer contributes to.

This distinguishes engineers who only operate in one silo vs engineers who understand the system holistically.

---

## Debt Cleanup: The Quiet Heroes

One of the most revealing metrics is **technical debt cleanup**.

When a fix commit modifies code, we check **who originally wrote those lines**.

```python
for fix_commit in fix_commits:
    fixer = fix_commit.author
    for changed_line in fix_commit.changed_lines:
        original_author = git_blame(file, at=parent_commit)
        if original_author != fixer:
            debt_generated[original_author] += 1
            debt_cleaned[fixer] += 1

debt_ratio = debt_cleaned / max(debt_generated, 1)
# > 1 = Cleaner  |  < 1 = Debt creator
```

If an engineer frequently fixes other people's mistakes, their cleanup score rises.

This surfaces engineers who:

- stabilize the system
- maintain reliability
- quietly keep things working

These engineers are often undervalued by naive metrics.

---

## Indispensability: Bus Factor Risk

This metric measures ownership concentration.

```python
for module in all_modules:
    top_share = max(blame_distribution[module].values()) / total
    if top_share >= 0.8:
        critical_modules[top_author].append(module)
    elif top_share >= 0.6:
        high_risk_modules[top_author].append(module)

indispensability = critical_count * 1.0 + high_count * 0.5
```

If one engineer owns more than 80% of the lines in a module, that module becomes a **bus factor risk**.

Indispensability reflects both:

- expertise
- organizational fragility

For that reason, it has a low weight.

---

## Scoring

Each metric is normalized within its domain (backend / frontend / infra).

```
norm(value) = min(value / max_value * 100, 100)
```

Final score:

```
score =
  production * 0.15 +
  quality * 0.10 +
  survival * 0.25 +
  design * 0.20 +
  breadth * 0.10 +
  debt_cleanup * 0.15 +
  indispensability * 0.05
```

The scale is intentionally strict.

| Score | Assessment | Approx. Hourly Rate (JPY) |
|---|---|---|
| 80+ | Irreplaceable core member | ¥12,000–20,000 |
| 60–79 | Near-core. Strong | ¥9,000–15,000 |
| **40–59** | **Senior-level (40+ is genuinely strong)** | **¥7,000–11,000** |
| 30–39 | Mid-level | ¥6,000–9,000 |
| 20–29 | Junior–Mid | ¥5,000–8,000 |
| <20 | Junior | ¥3,500–6,000 |

**40 = Senior.** With relative scoring across 7 axes, just putting up decent numbers across the board requires serious ability.

---

## Patterns That Emerge

Once scores are calculated, recognizable patterns appear.

![Engineering Archetypes plotted in Production vs Survival space](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/engineering-archetypes-paper-figure.png)
*Figure 3. Different engineer archetypes emerge naturally when production and time-decayed survival are plotted together.*

### Architect
High production, survival, design influence, and debt cleanup.

These engineers build the structural backbone of the system.

### Cleaner
Moderate production but extremely high durability and cleanup.

Often undervalued, but essential for system stability.

### High-Output / Low-Durability
Large code volume but low survival and frequent fixes.

These engineers generate technical churn. They *look* productive, which makes them dangerous to evaluate on output alone.

### Wide Presence / Low Impact
High repository breadth but limited design or durable code.

Visible across the system but with shallow influence.

---

## Why Revenue Metrics Miss Engineering Health

Business KPIs measure **product success**, not **engineering health**.

A company can grow rapidly while its codebase quietly deteriorates.

Engineering health depends on signals such as:

- **Code durability** — are you rewriting the same features every quarter?
- **Technical debt accumulation** — does adding 1 feature generate 2 bug fixes?
- **Bus factor** — how many modules die if one person leaves?
- **Design concentration** — is all architecture decided by one person?

Git history contains all of these signals.

**Even with revenue growing, if Survival decline + Debt increase + Bus Factor concentration are progressing simultaneously, the organization will collapse at scale.**

---

## Accuracy Scales with Design Quality

This metric has a property where **higher codebase design quality yields higher accuracy**.

In well-structured codebases (Clean Architecture, DDD), "touching design files = making design decisions" holds true, and high Survival means "the design withstood change pressure."

In chaotic codebases, high Survival might just mean "dead code nobody touches."

**The metric's low accuracy is itself a signal of poor design.** If it doesn't match gut feeling, the codebase structure may not withstand measurement. Investing in design is itself infrastructure for improving evaluation accuracy.

---

## Limitations

This model is not perfect.

- Commit messages affect fix detection accuracy
- Blame analysis can be expensive on large repos
- Frontend refactors reduce survival scores (→ separate BE/FE scoring)
- Copy-paste inflates production (→ offset by survival and design)
- Debt cleanup depends on sample size (→ threshold: <10 events = reference value)

However, even with imperfections, these signals are far more informative than raw commit counts.

---

## This Is Not a Measure of Human Worth

This model does **not** measure a person's value.

It estimates **technical influence observable in a codebase**.

Engineers contribute in many ways that git cannot capture:

- mentoring
- domain expertise
- documentation
- team culture

Those contributions matter. But observable signals are still useful.

---

## Final Thought

Git quietly records the real history of a system.

If we analyze that history carefully, we can understand:

- who builds durable systems
- who stabilizes them
- where organizational risk exists

Numbers will never tell the whole story. But they are far better than guessing.

As AI-driven development grows, this model could also measure **the rigidity of AI-generated code**. Is AI a debt factory or a cleaner? Git records humans and AI equally. Quantitatively tracking AI-written code survival and debt ratio — that feels like the next phase of engineering evaluation.

If you run this model on your own codebase, the results may surprise you.

---

**GitHub:** [machuz/engineering-impact-score](https://github.com/machuz/engineering-impact-score) — Full formulas, methodology, and blog posts in English and Japanese.

**Support:** If this helped you see your team differently — [GitHub Sponsors](https://github.com/sponsors/machuz) / PayPay: `w_machu7`

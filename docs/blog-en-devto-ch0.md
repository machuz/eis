---
title: "Git Archaeology #0 — What If Git History Could Tell You Who Your Strongest Engineers Are?"
published: true
description: "A 3-minute intro to Engineering Impact Score — an OSS CLI that quantifies engineering impact from git log and git blame alone. No API keys, no AI tokens."
tags: opensource, productivity, git, career
cover_image: https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/blog/png/cover-ch0.png?v=1
---

*git log + git blame. That's all it takes.*

---

## What Is This?

**Engineering Impact Score (EIS, pronounced "ace")** is an open-source CLI tool that quantifies engineering impact from Git history alone.

No external APIs. No AI tokens. Just `git log` and `git blame`.

```bash
brew tap machuz/tap && brew install eis
cd your-repo
eis
```

That's it. Here's what you get:

![Terminal Output](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/terminal-output.svg)

---

## Why I Built This

Commit counts. PR counts. Lines of code. Easy to measure — and all meaningless.

A typo fix and a system-wide architecture change both count as "one PR." A generated lockfile adds thousands of lines. Commit frequency varies wildly between engineers.

Yet inside every team, people know who the strongest engineers are.

> "That person writes code that lasts."
> "That person touches everything but nothing improves."

Those intuitions exist, but they're not measurable. I wanted to **turn gut feeling into numbers** — and I wanted those numbers to come from a source that can't be gamed by politics: **the git history itself.**

---

## The Telescope

Over the years, I've recruited several strong engineers by reaching out personally. I'm grateful that many of them said yes.

Why did they come? I don't think it was just the tech stack or compensation.

**"This person actually sees my work."** — I believe that's what they felt.

For engineers, having your technical contributions properly observed matters deeply. Not commit counts or PR counts, but **whether your code survives, whether you contribute to architecture, whether you clean up debt** — having someone who sees that substance.

I had that eye. At least, that's my self-assessment.

EIS is that **observer's eye, turned into an open-source telescope.**

Anyone can use it. Point it at any team. Through the lens of git history — a lens that cannot lie.

---

## Making Architecture a Science

Being strong at math. Being strong at algorithms. Being strong at language specifications.

These all have theory forged in academia over decades. Computational complexity, type theory, formal verification — they can guarantee correctness through mathematical proof. They stand on scientific foundations.

Academic attempts at software architecture have existed for over 30 years. Architecture Description Languages and evaluation methods have been proposed. But **they haven't converged into a unified theory.** The efforts remain fragmented and haven't made it into practice.

"What is good design?" "Is this team's structure healthy?" — the industry has best practices and rules of thumb for these questions, but very little quantitative language.

And now AI writes enormous amounts of code.

**The value of writing code is declining in relative terms. What matters most is structure.** What structure do you place code on top of? Does that structure withstand change? Where does the team's knowledge accumulate?

If EIS can become a **tool for making architecture a science**, I'd be glad.

---

## 7 Axes of Impact

EIS scores engineers across 7 axes:

| Axis | Weight | What it measures |
|---|---|---|
| Production | 15% | Volume of changes |
| Quality | 10% | First-time quality (low fix/revert rate) |
| **Survival** | **25%** | **Does your code still exist today? (time-decayed)** |
| Design | 20% | Contributions to architecture files |
| Breadth | 10% | Cross-repository activity |
| Debt Cleanup | 15% | Cleaning up other people's debt |
| Indispensability | 5% | Module ownership (bus factor) |

The most important axis is **Survival**. Is the code you wrote still there after 6 months? After a year?

Writing a lot means nothing if it gets rewritten next month. **Engineers who write code that lasts are the strong ones.**

---

## 3-Axis Archetypes

Beyond scores, EIS classifies engineers along three independent axes:

**Role** — What you contribute
- Architect: designs the structure
- Anchor: guards quality
- Cleaner: pays down debt
- Producer: generates volume
- Specialist: deep in one area

**Style** — How you contribute
- Builder: builds and designs simultaneously
- Resilient: rebuilds after destruction
- Rescue: pays down others' debt
- Churn: high volume, low survival
- Mass: mass production with low survival
- Balanced: even across all axes
- Spread: touches everything but lacks depth

**State** — Lifecycle phase
- Former: left the team, but their code remains
- Silent: low activity, low survival (detected only for experienced engineers)
- Fragile: code survives only because nobody touches it
- Growing: low volume, but high quality
- Active: currently contributing

From these classifications, **team structure becomes visible.**

---

## What This Reveals

Here's what EIS has surfaced in real teams:

- **A departed Architect's code still makes up 30% of the codebase** (Former detection)
- **Code that survives only because nobody touches it** — not because it's good (Fragile detection)
- **No Producers on the team** — the layer that generates volume on top of structure is empty (Producer Vacuum)
- **Architect Bus Factor = 1** — all design knowledge concentrated in one person

Cold git history tells **team stories** you didn't know you had.

---

## The Series

This blog series — **Git Archaeology** — applies EIS to real teams and explores what the numbers reveal.

1. **[Measuring Engineering Impact from Git History Alone](https://dev.to/machuz/git-archaeology-1-measuring-engineering-impact-from-git-history-alone-55fh)** — Full 7-axis scoring design
2. **Team Topology** — How scores reveal team structure
3. **FE Architects Diverge** — Frontend evolution model
4. **BE Architects Converge** — Backend evolution model
5. **Infra Goes Silent** — The visibility problem of infrastructure engineers
6. **Gravity Map** — Visualizing code gravity fields
7. **Risk Detection** — Quantifying technical risk
8. **Normalization Design** — Why hybrid scoring
9. **Domain Separation** — Why mixing BE/FE/Infra pollutes scores
10. **Time Decay Design** — The math behind Survival
11. **Archetype Design** — Full classification logic
12. **Team Metrics** — Team health diagnostics
13. **Robust Survival** — Tested code survival
14. **Change Pressure** — Quantifying change pressure
15. **Multi-Repo Analysis** — Cross-organization scoring
16. **The Future** — What's next for EIS

---

## Install

```bash
# Homebrew
brew tap machuz/tap && brew install eis

# Go
go install github.com/machuz/engineering-impact-score/cmd/eis@latest
```

**GitHub**: [engineering-impact-score](https://github.com/machuz/engineering-impact-score)

---

![EIS — the Git Telescope](https://raw.githubusercontent.com/machuz/engineering-impact-score/main/docs/images/logo-full-zenn.png)

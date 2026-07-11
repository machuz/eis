# MCP Effect Measurement

Research on **how git-history structural context changes an AI coding agent's decisions** — and whether that change is real, valuable, and defensible.

The OrbitLens MCP server exposes EIS-derived signals (code survival, change pressure, ownership concentration / bus factor, co-change / blast radius, fragility) to AI coding agents. The obvious question — *"does this make the AI write better code?"* — turns out to be the wrong one, and unanswerable: code quality confounds engineer skill, task difficulty, review quality, and requirement churn.

So we measured the layer underneath it: **not whether the code got better, but whether the agent's decisions changed** — and isolated that change causally.

## Contents

- [`findings.md`](./findings.md) — the pilot: method, results (N=8 + controls), the mechanism, an honest value verdict, and limitations.

## One-paragraph result

The signals do **not** help an agent find bugs — a capable model reading the diff already finds them, and that overlap only grows as base models improve. What they *do*, measurably, is shift an **implementing** agent's verification posture: work it would defer ("I'll skip this test") becomes work it does before merge, when the code sits on a high-churn, single-owner module — pre-empting, at authoring time, gaps that human review otherwise catches later. The value rides entirely on signals that are **not derivable from the code** (who can maintain it, how often it churns, what co-changes with it). The same signals can *hurt* a weaker consuming model, which cargo-cults the numbers and over-engineers trivial changes — a failure that is fixable by framing (deliver the guard *next to* the signal, not just once at connect time). Net: a real, narrow, defensibly-moated effect whose ROI is conditioned on the consuming agent's judgment and on how the context is framed.

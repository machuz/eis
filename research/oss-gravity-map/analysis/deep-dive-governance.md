# Deep Dive: Three Governance Models — Rails, Laravel, esbuild

*Same relational metric. Three different things for a system to lean on.*

---

## Overview

These three projects sit at three extremes of where structural influence settles in open-source software. Gravity here is relational — a structural shape passed through two gates: did the code survive *under others' pressure* (survGate), and did others build on the surviving foundation (catGate). A system leans on someone only when both gates stay open.

| Project | Category | Engineers | Gravity Conc. | Top10 Design | Governance |
|---|---|---|---|---|---|
| **Rails** | Framework (Ruby) | 6,056 | **9.8%** | **57.0** | Multi-architect civilization |
| **Laravel** | Framework (PHP) | 4,149 | 38.0% | 12.0 | Creator's kingdom |
| **esbuild** | Bundler (Go) | 124 | **77.0%** | 10.0 | One-person universe |

Gravity Concentration runs from **9.8% (Rails) to 77.0% (esbuild)** — roughly an 8x gap (and swc, at 92.0%, is denser still). The same instrument, the same two gates, and three civilizations come apart. The question underneath each is the same: once the noise of activity is filtered out, whom does the structure actually lean on?

---

## 1. Rails — The Multi-Architect Civilization

![Rails Gravity Concentration](chart-gravity-concentration.svg)

### Key Numbers

| Metric | Value |
|---|---|
| Engineers | 6,056 |
| Gravity Concentration | **9.8%** (lowest of all 29 repos) |
| Top10 Design | **57.0** (higher than most single-architect projects) |
| Top10 Gravity (avg) | 2.4 |
| Top10 Survival (avg) | 2.5 |
| Authors with Design > 40 | **9 people** |

### Top 10 Gravity Ranking

| # | Engineer | Gravity | Design | Indisp. | Catalysis | RobustSurv | State |
|---|---|---|---|---|---|---|---|
| 1 | David Heinemeier Hansson | 4.0 | 100 | 100 | 19.5 | 0 | Silent |
| 2 | Jean Boussier | 3.8 | 49.1 | 0 | 62.6 | 0 | Growing |
| 3 | Rafael Mendonça França | 3.6 | 89.5 | 21.4 | 22.2 | 0 | Silent |
| 4 | zzak | 3.5 | 1.1 | 28.6 | 25.8 | 25.3 | Silent |
| 5 | Matthew Draper | 2.7 | 24.3 | 0 | 100 | 0 | Growing |
| 6 | Gannon McGibbon | 1.9 | 13.0 | 0 | 43.9 | 0 | Silent |
| 7 | Xavier Noria | 1.5 | 48.6 | 0 | 26.9 | 0 | Silent |
| 8 | Aaron Patterson | 1.4 | 96.8 | 28.6 | 2.4 | 0 | Silent |
| 9 | Jeremy Kemper | 1.1 | 89.2 | 0 | 0 | 0 | Silent |
| 10 | Ryuta Kamizono | 0.9 | 58.8 | 0 | 7.1 | 0 | Silent |

### Design Authority Distribution

```
DHH             ████████████████████████████████████████████████████████  100.0
Aaron Patterson ██████████████████████████████████████████████████████     96.8
Rafael França   ██████████████████████████████████████████████████         89.5
Jeremy Kemper   ██████████████████████████████████████████████████         89.2
R. Kamizono     █████████████████████████████████                          58.8
Jon Leighton    █████████████████████████████████                          58.7
José Valim      █████████████████████████████                              51.7
Jean Boussier   ███████████████████████████                                49.1
Xavier Noria    ███████████████████████████                                48.6
Joshua Peek     ██████████████████████                                     39.0
                ─────────────────────────── Design > 35 threshold ────────
```

**10 people exceed Design 35; 9 exceed 40.** That is unusual for a framework whose user-facing API has a single signature.

### Analysis

Rails reads less like a creator's kingdom and more like a **multi-architect civilization**.

DHH set the gravitational center and still holds Design 100 and Indispensability 100. Yet almost none of the top names carry surviving code where *others* keep pressing — nine of the top ten read RobustSurvival 0 (zzak's 25.3 is the lone exception), so survGate stays near its floor for nearly all of them. What separates them is whose foundation others still build on: DHH leads at Gravity 4.0 not because his decades-old design persists under live load, but because his catGate (Catalysis 19.5) and his shape together edge past the field. Behind him the scores compress — Jean Boussier 3.8, Rafael França 3.6, zzak 3.5 — and the gap to the leader is small.

Top10 Design averaging **57.0** sits above most projects that are supposed to concentrate design authority. The design history is spread thin: Aaron Patterson (96.8), Rafael França (89.5), Jeremy Kemper (89.2), José Valim (51.7), Xavier Noria (48.6) all carry deep architectural reach, and none of them lead.

The Gravity Concentration of **9.8%** is the lowest of all 29 repos — below Kubernetes (27.5%) and the Rust compiler (27.6%). The instrument reads no single person the system leans on more than the rest. The structure does not tilt toward anyone, not even the creator.

What that looks like, if one wanted a name for it, is architect succession that already happened — the gravity never settled on one chair.

---

## 2. Laravel — The Creator's Kingdom

![Rails vs Laravel](chart-rails-vs-laravel.svg)

### Key Numbers

| Metric | Value |
|---|---|
| Engineers | 4,149 |
| Gravity Concentration | 38.0% |
| Top10 Design | **12.0** (far below Rails 57.0) |
| Top10 Gravity (avg) | 9.3 |
| Otwell's lead margin | **65.1 vs 6.5** (10x the #2) |
| Authors with Design > 40 | **2 people** |

### Top 10 Gravity Ranking

| # | Engineer | Gravity | Design | Indisp. | Catalysis | RobustSurv | State |
|---|---|---|---|---|---|---|---|
| 1 | Taylor Otwell | 65.1 | 100 | 100 | 100 | 78.4 | Growing |
| 2 | Luke Kuzmish | 6.5 | 1.5 | 50 | 25.4 | 53.9 | — |
| 3 | Mior Muhammad Zaki | 4.7 | 4.2 | 0 | 60.1 | 84.3 | Fragile |
| 4 | Caleb White | 4.4 | 0.9 | 50 | 45.6 | 20.9 | — |
| 5 | Nuno Maduro | 2.8 | 4.9 | 50 | 29.3 | 15.7 | Silent |
| 6 | Patrick Carlo-Hickman | 2.6 | 0.1 | 0 | 19.3 | 100 | — |
| 7 | Andrew Brown | 2.0 | 3.7 | 0 | 23.0 | 46.2 | Silent |
| 8 | Lucas Michot | 1.8 | 4.5 | 0 | 11.8 | 12.7 | Silent |
| 9 | Günther Debrauwer | 1.5 | 0.5 | 0 | 24.3 | 19.6 | — |
| 10 | Ahmed Alaa | 1.3 | 0.1 | 0 | 26.1 | 15.4 | — |

### Design Monopoly

```
Taylor Otwell      ████████████████████████████████████████████████████████  100.0
Graham Campbell    ██████████████████████████                                  46.4
                   ─────────────────────── gap ──────────────────────────────
Dries Vints        ████                                                         6.3
Nuno Maduro        ███                                                          4.9
Lucas Michot       ███                                                          4.5
Mohamed Said       ██                                                           4.3
Mior M. Zaki       ██                                                           4.2
Andrew Brown       ██                                                           3.7
Others             ▏                                                            0.0
```

**Taylor Otwell holds nearly all the Design.** Only Graham Campbell (46.4) has any comparable architectural reach. Everyone else falls below Design 7.

### Analysis

Laravel reads as a **creator's kingdom** — and here both gates open for one person.

Taylor Otwell is the only author with Catalysis 100 *and* RobustSurvival 78.4: others build on his foundation, and his code persists in modules where others keep committing. Both gates stay wide open, his shape (Design 100, Indispensability 100) is saturated, and the product is a Gravity of **65.1** — about ten times the #2, Luke Kuzmish at 6.5. The system leans on him along every axis the instrument can read.

This is not, in itself, a fault. A single architect's consistent reading means less churn — coherence through one hand.

But the structure rests on one chair. Compare Rails: same framework category, yet Rails spreads design history across its top contributors (Top10 Design avg **57.0**, 9 authors above 40) while Laravel's averages **12.0** with only 2 — and no one near the leader. Where Rails reads no one the system tilts toward, Laravel reads exactly one.

Same framework pattern. Opposite governance physics.

The open question is not which is better but what produced the difference. Rails is twenty-some years deep, and its gravity never settled on a single name. Laravel is younger, and its creator is still the one the structure leans on. Whether the field lines stay this concentrated, or spread the way Rails did, is something only the next decade can read.

---

## 3. esbuild — The One-Person Universe

### Key Numbers

| Metric | Value |
|---|---|
| Engineers | 124 |
| Gravity Concentration | **77.0%** (swc 92.0% is higher) |
| Top10 Design | 10.0 |
| Evan Wallace | Design/Breadth/Indisp **100**, RobustSurvival **0**, Gravity **15.0** |
| #2 contributor — all shape axes | Design 0, Indispensability 0, RobustSurvival 0 |

### The Singularity

```
Evan Wallace:    Gravity 15.0 | Design 100 | Indisp 100 | RobustSurv 0
                 ════════════════════════════════════════════════════════════════

Everyone else:   Gravity ~0.2 | Design 0 | Indisp 0 | RobustSurv 0
                 (no other contributor has any Design or Indispensability)
```

Every contributor other than Evan Wallace reads **zero Design, zero Indispensability, zero RobustSurvival** — their gravity rounds to ~0.2. None of their code became structural.

### Where esbuild's 77.0% Sits

| Project | Gravity Concentration |
|---|---|
| swc | 92.0% |
| **esbuild** | **77.0%** |
| express | 60.3% |
| vite | 59.9% |
| react | 31.4% |
| rust | 27.6% |
| kubernetes | 27.5% |
| rails | 9.8% |

esbuild's concentration is **~3x Kubernetes** and **~8x Rails**. It is not the single densest project (swc edges it out), but it is the purest one-person universe.

### Analysis

esbuild is what a single architect's universe reads like when no one else is inside it.

Evan Wallace wrote esbuild as a demonstration that JavaScript bundling could run an order of magnitude faster, across 4,202 commits, with Design, Breadth and Indispensability all at 100. And yet the gravity stops at **15.0** — not because the shape is small, but because the survGate sits at its floor. Robust survival counts only code that lasts in modules where *others* commit, and in a solo project there are no others pressing on the code. RobustSurvival reads **0**, survGate collapses to 0.15, and the relational gates decline to mint the influence that the raw shape alone might suggest. The instrument is doing exactly what it should: it cannot see a structure leaning on someone when no one else is standing on it.

This is the **far end of the single-architect model** — a singularity. In physics, singularities are where the ordinary laws stop applying, and the reading is similar here: esbuild does not distribute, does not succession-plan, and has no reason to. It is a finished artifact, not an evolving ecosystem.

The other contributors sent patches and fixes, but none of it became structural. Design 0 and Indispensability 0 mean none of it shaped the architecture or became leaned on.

A singularity is complete in itself — and that completeness is exactly why no one else can inherit it.

---

## Three Models Compared

![Three Modes of Structural Authority](chart-three-modes.svg)

| Dimension | Rails | Laravel | esbuild |
|---|---|---|---|
| **Governance** | Multi-architect civilization | Creator's kingdom | Singularity |
| **Design Distribution** | 9 authors > 40 | 1 architect (+ 1 deputy) | 1 architect, period |
| **Who the system leans on** | No one in particular | Otwell, every axis | Wallace, but unleant-on |
| **Both gates open for the leader** | No (RobustSurv 0) | Yes (Cat 100, RobustSurv 78.4) | No (RobustSurv 0) |
| **Leader's Gravity** | 4.0 (margin to #2: 0.2) | 65.1 (10x the #2) | 15.0 (held by survGate floor) |
| **Gravity Concentration** | 9.8% | 38.0% | 77.0% |
| **Top10 Design** | 57.0 | 12.0 | 10.0 |
| **Engineers observed** | 6,056 | 4,149 | 124 |

### Key Insight

The same metric, applied to three projects, reads three relationships between a system and the people inside it. Gravity is relational: it only registers where the structure observably leans on someone — code that survives under others' pressure, a foundation others build on.

- **Rails** tilts toward no one. The top scores compress (4.0, 3.8, 3.6), no leader carries surviving code under live load, and the lowest concentration of all 29 repos (9.8%) reads a system that would not collapse with any single departure.
- **Laravel** leans on Taylor Otwell along every axis at once — the rare case where both gates open for one person, putting his gravity ten times above the next name. Coherent, and resting on one chair.
- **esbuild** is shaped entirely by one person, and the instrument still declines to call it gravity, because no one else stands on the code. Robust survival is zero, the survGate sits at its floor, and the singularity stays uninheritable by construction.

The question worth holding is not which model is best, but a quieter one the map keeps open: once activity is filtered out, what does a system actually lean on — and would anyone notice if that quietly left?

---

*Generated by [EIS (Engineering Impact Score)](https://github.com/machuz/eis) — OSS Gravity Map Project*
*Gravity = catGate(Catalysis) × survGate(RobustSurvival) × shape(0.45·Design + 0.25·Breadth + 0.30·Indispensability)*

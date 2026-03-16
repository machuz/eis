# The Laws of Physics Are Not Uniform — Cross-Language Analysis

*OSS Gravity Map: Do type systems and frameworks shape gravity?*

## Hypothesis

Different programming languages create different "physical laws" in their code universes. Specifically:

1. **Type system expressiveness** may correlate with how Design influence distributes across a team
2. **Gravity concentration** — the degree to which structural influence is held by a few people — may vary systematically by language family
3. **Framework-driven** ecosystems may absorb gravity into the framework itself, lowering individual engineers' structural influence

This is an empirical investigation using EIS (Engineering Impact Score) data from 29 major OSS repositories.

## Methodology

- **29 repositories** analyzed with `eis analyze`, covering **55,343 engineers** total
- Repositories classified into **5 categories** by type system and structural culture
- For each repo, the **top 10 contributors by Gravity** are examined (averages of the broader population are diluted by thousands of low-activity contributors)
- **Gravity Concentration** = share of total gravity held by the top 10 contributors

### Categories

| Category | Characteristics | Languages | Repos |
|---|---|---|---|
| **Expressive** | Rich type system, ADTs, pattern matching, traits | Rust, Scala | 5 |
| **Go (Self-structured)** | Static, nominal typing, anti-framework culture, explicit interfaces | Go | 7 |
| **Framework-driven** | Structure delegated to framework; conventions over code | Ruby (Rails), PHP (Laravel), Java (Spring), Python (FastAPI), TS (NestJS), Elixir (Phoenix) | 6 |
| **Systems (C/C++)** | Static, manual memory, templates | C, C++ | 5 |
| **Dynamic / Structural** | Dynamic or structural typing, self-structured | JavaScript, TypeScript, Python | 6 |

**Why split Go and Java?** Go and Java share nominal type systems, but their **structural culture is opposite**. Go eschews frameworks — engineers build structure from scratch using standard library and explicit interfaces. Java (Spring) delegates structure to framework annotations, dependency injection, and convention. This produces fundamentally different gravity distributions.

**Why split Framework-driven and Dynamic?** NestJS, FastAPI, and Phoenix are framework-driven even though their languages are dynamic or structural. Express, React, and Prettier are dynamic-language projects where engineers build their own structure.

---

## Results

### Summary by Category

| Category | Repos | Avg Size | Top10 Design | Top10 Survival | Top10 Gravity | Gravity Concentration |
|---|---|---|---|---|---|---|
| **Go (Self-structured)** | 7 | 2,135 | 31.3 | 19.9 | 52.8 | **16.4%** |
| **Dynamic / Structural** | 6 | 1,184 | 24.8 | 17.1 | 51.2 | 10.4% |
| **Expressive** | 5 | 2,151 | 26.3 | 13.7 | 51.2 | 6.7% |
| **Framework-driven** | 6 | 2,578 | 23.7 | 15.4 | 47.2 | 5.1% |
| **Systems (C/C++)** | 5 | 1,414 | 25.4 | 14.4 | 52.4 | 3.4% |

### Key Observations

#### 1. Go Concentrates Gravity — Expressive Types Distribute It

| Category | Gravity Concentration |
|---|---|
| **Go (Self-structured)** | **16.4%** |
| **Dynamic / Structural** | 10.4% |
| **Expressive (Rust/Scala)** | 6.7% |
| **Framework-driven** | 5.1% |
| **Systems (C/C++)** | 3.4% |

**Go concentrates gravity at 2.4x the rate of Expressive type systems, and 3.2x the rate of Framework-driven projects.** Go's anti-framework culture means someone *must* build the structure — and that structural authority concentrates in a few architects. The explicit interface pattern creates a small number of contract-defining files that dominate `git blame`.

Expressive type systems distribute gravity more evenly. When the type system itself can encode design decisions (through ADTs, traits, type-level programming), **structure becomes embedded in the type signatures rather than in a few architects' code**. The type system acts as a distributed architect.

#### 2. Frameworks Absorb Gravity

The **Framework-driven** category shows the lowest Top10 Gravity (**47.2**) and the lowest Top10 Design (**23.7**). This is the framework absorbing structural influence that would otherwise belong to engineers.

When Rails, Laravel, Spring, or Phoenix defines the routing, DI, middleware, and lifecycle, those design decisions don't appear in any engineer's `git blame`. **The framework is the invisible architect.** EIS can only observe gravity that lives in code — gravity that lives in framework conventions is Dark Matter.

#### 3. Design Influence Tells the Story

| Category | Top10 Design |
|---|---|
| **Go (Self-structured)** | **31.3** |
| **Expressive** | 26.3 |
| **Systems (C/C++)** | 25.4 |
| **Dynamic / Structural** | 24.8 |
| **Framework-driven** | **23.7** |

Go leads Design because its architects write the interfaces, the routing structure, the middleware patterns — all from scratch. In framework-driven ecosystems, these same design decisions are made by the framework, not by engineers.

---

## Deep Dive: Rails vs Laravel

Both are iconic framework-driven projects with legendary creator-architects. Both creators are still active (or recently active). Yet the gravity physics are strikingly different.

| Metric | Rails (Ruby) | Laravel (PHP) |
|---|---|---|
| Engineers | 6,512 | 4,449 |
| **Top10 Design** | **51.2** | **17.2** |
| Top10 Survival | 7.3 | **15.3** |
| Top10 Gravity | 54.0 | 56.2 |
| Gravity Concentration | **0.9%** | 1.5% |

### Rails: Distributed Design

| # | Engineer | Gravity | Design | Indispensability |
|---|---|---|---|---|
| 1 | David Heinemeier Hansson | 100.0 | 100.0 | 100.0 |
| 2 | Aaron Patterson | 64.0 | 54.0 | 44.4 |
| 3 | Jeremy Kemper | 58.9 | 89.0 | 5.6 |
| 4 | Rafael Mendonça França | 55.0 | 68.4 | 11.1 |
| 5 | Pratik Naik | 50.4 | 31.1 | 27.8 |
| 6 | Xavier Noria | 48.2 | 45.8 | 11.1 |
| 7 | José Valim | 41.8 | 39.4 | 0.0 |
| 8 | Joshua Peek | 41.0 | 36.6 | 0.0 |

**Design is distributed across many architects.** Aaron Patterson (54.0), Jeremy Kemper (89.0), Rafael França (68.4), Xavier Noria (45.8), José Valim (39.4) — five people with Design > 35. Rails has **multiple design authorities**.

### Laravel: Concentrated Design

| # | Engineer | Gravity | Design | Indispensability |
|---|---|---|---|---|
| 1 | Taylor Otwell | 100.0 | 100.0 | 100.0 |
| 2 | Kay W. | 70.0 | 0.1 | 100.0 |
| 3 | Lucas Michot | 51.1 | 3.5 | 50.0 |
| 4 | Tim MacDonald | 50.8 | 2.6 | 50.0 |
| 9 | Graham Campbell | 48.9 | 63.1 | 0.0 |

**Taylor Otwell holds all the Design.** Only Graham Campbell (63.1) has significant Design influence — everyone else is below 4. Laravel is a **one-architect universe**.

### What This Means

Rails and Laravel are both "framework-driven" — but:

- **Rails** has evolved into a **multi-architect civilization**. DHH created the structure, but many have since reshaped it. Top10 Design of 51.2 is higher than most Go projects. Rails is "framework-driven" in its user-facing API, but internally it functions more like a self-structured project.
- **Laravel** remains a **creator-centric kingdom**. Taylor Otwell's Indispensability = 100 and Design = 100 mean the framework's architecture still flows through one person. This is efficient (Survival 15.3 > Rails' 7.3 — less rewriting) but creates a single point of failure.

**Same framework-driven category. Completely different governance physics.**

This suggests that the Framework-driven category itself has a spectrum: from distributed governance (Rails) to concentrated governance (Laravel). The framework absorbs gravity from users, but the question is whether it also distributes gravity among its own contributors.

---

### Per-Repository Detail

| Repository | Category | Language | Engineers | Top10 Design | Top10 Survival | Top10 Gravity | Grav Conc |
|---|---|---|---|---|---|---|---|
| polars | Expressive | Rust | 694 | 17.4 | 29.8 | 41.3 | 7.0% |
| rust | Expressive | Rust | 7,914 | 53.0 | 5.7 | 68.9 | 0.6% |
| scala | Expressive | Scala | 794 | 32.3 | 5.9 | 56.7 | 6.8% |
| scala3 | Expressive | Scala 3 | 995 | 12.6 | 12.9 | 49.2 | 3.3% |
| swc | Expressive | Rust | 357 | 16.4 | 14.0 | 40.1 | 15.7% |
| argo-cd | Go (Self-structured) | Go | 1,889 | 41.4 | 20.5 | 52.0 | 4.0% |
| esbuild | Go (Self-structured) | Go | 125 | 10.0 | 10.0 | 37.0 | 92.5% |
| grafana | Go (Self-structured) | Go/TS | 2,893 | 16.3 | 24.6 | 50.1 | 1.9% |
| kubernetes | Go (Self-structured) | Go | 5,217 | 42.5 | 19.9 | 56.1 | 0.8% |
| loki | Go (Self-structured) | Go | 1,319 | 34.0 | 19.3 | 65.5 | 7.8% |
| prometheus | Go (Self-structured) | Go | 1,281 | 44.5 | 11.9 | 62.1 | 5.4% |
| terraform | Go (Self-structured) | Go | 2,223 | 30.3 | 32.7 | 46.5 | 2.4% |
| rails | Framework-driven | Ruby | 6,512 | 51.2 | 7.3 | 54.0 | 0.9% |
| laravel | Framework-driven | PHP | 4,449 | 17.2 | 15.3 | 56.2 | 1.5% |
| spring-boot | Framework-driven | Java | 1,504 | 30.2 | 34.1 | 52.8 | 7.3% |
| nest | Framework-driven | TypeScript | 684 | 13.3 | 10.2 | 39.6 | 7.0% |
| fastapi | Framework-driven | Python | 897 | 11.7 | 11.2 | 37.7 | 9.4% |
| phoenix | Framework-driven | Elixir | 1,424 | 18.9 | 14.4 | 42.9 | 4.5% |
| ClickHouse | Systems (C/C++) | C++ | 2,563 | 17.3 | 30.0 | 54.8 | 1.4% |
| arrow | Systems (C/C++) | C++/Multi | 1,466 | 28.3 | 12.4 | 50.9 | 3.8% |
| duckdb | Systems (C/C++) | C++ | 690 | 20.7 | 4.9 | 49.3 | 4.5% |
| envoy | Systems (C/C++) | C++ | 1,444 | 46.4 | 15.7 | 61.2 | 3.5% |
| redis | Systems (C/C++) | C | 905 | 14.3 | 8.9 | 45.6 | 4.0% |
| eslint | Dynamic / Structural | JavaScript | 1,180 | 18.6 | 12.8 | 58.6 | 8.1% |
| express | Dynamic / Structural | JavaScript | 391 | 13.5 | 3.2 | 38.9 | 23.9% |
| prettier | Dynamic / Structural | JavaScript | 799 | 42.9 | 10.0 | 48.2 | 13.4% |
| react | Dynamic / Structural | JavaScript | 2,010 | 17.5 | 31.1 | 48.0 | 3.4% |
| superset | Dynamic / Structural | Python/TS | 1,477 | 23.6 | 29.5 | 57.7 | 4.5% |
| vite | Dynamic / Structural | TypeScript | 1,247 | 32.6 | 16.0 | 55.8 | 9.3% |

---

## Interpretation

### Three Modes of Structural Authority

The data reveals three distinct ways that code universes distribute structural authority:

1. **Architect-centric** (Go) — Structure is built by people. A small number of architects hold disproportionate design authority. Gravity concentration: high.

2. **Type-distributed** (Rust, Scala) — Structure is encoded in the type system. Design authority is distributed across anyone who writes type signatures. Gravity concentration: low.

3. **Framework-absorbed** (Rails, Laravel, Spring, NestJS, FastAPI, Phoenix) — Structure lives in the framework. But within this mode, there's a spectrum from distributed governance (Rails) to concentrated governance (Laravel).

These three modes produce similar-looking code, but the **physics of who holds structural authority** is fundamentally different. And this has consequences for team scaling, architect succession, and entropy resistance.

### This Is Not About Superiority

These numbers do not say one language is better than another. They show that **each language family creates a different gravitational physics**.

- In a **small universe** where complexity is manageable, framework-absorbed gravity works well. Bootstrap quickly, let the framework be your architect.
- In a **large universe** with many contributors: does the language help distribute design authority (Expressive), or does it force centralization (Go)?
- In **self-structured** universes (Go, Dynamic): the architect becomes a single point of failure. The question is whether you can sustain architect succession.

**Knowing which physics your universe operates under helps you make better structural decisions.**

### Limitations

- **29 repositories** is a meaningful sample but not exhaustive
- Repository maturity, governance model, and community culture are confounding variables
- EIS observes `git blame` and commit patterns — design decisions expressed outside of code (RFCs, ADRs, discussions) are invisible (Dark Matter)
- Some repositories span multiple languages (grafana: Go+TS, superset: Python+TS)
- Category boundaries are judgment calls — reasonable people may classify differently
- The Rails vs Laravel comparison shows that **within-category variance can be as large as between-category variance** — categories are useful but not deterministic

### Future Directions

- **Entropy resistance by language**: Do expressive type systems show higher Robust Survival as universe size grows?
- **Architect succession patterns**: Do certain language families produce smoother generational transitions?
- **Framework effect isolation**: Compare framework-driven vs bare projects in the same language
- **Scale threshold analysis**: At what universe size does gravity concentration become a risk factor?
- **Governance spectrum within Framework-driven**: What makes Rails distribute design authority while Laravel concentrates it?

---

## Conclusion

The hypothesis that "the laws of physics are not uniform across code universes" is supported by the data.

**Gravity concentration varies by 4.8x between Go (Self-structured) and Systems (C/C++).** Between the three structural modes — architect-centric, type-distributed, framework-absorbed — the difference in how design authority flows through a team is stark.

The Rails vs Laravel deep dive reveals that even within the same category, **governance physics can diverge dramatically**. Rails distributes design across many architects (Top10 Design: 51.2); Laravel concentrates it in one (Top10 Design: 17.2). Same framework pattern, opposite authority structures.

For years, debates about technology choices have been aerial battles — fought with experience and intuition. This data is a first step toward **making design debates scientific**.

---

*Generated by [EIS (Engineering Impact Score)](https://github.com/machuz/engineering-impact-score) — OSS Gravity Map Project*
*29 repositories, 55,343 engineers, observed through commit light.*

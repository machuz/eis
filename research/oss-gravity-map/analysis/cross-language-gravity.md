# The Laws of Physics Are Not Uniform — Cross-Language Analysis

*OSS Gravity Map: do type systems and frameworks decide who a system leans on?*

## Hypothesis

It is tempting to read a code universe by its language. Each language carries a folklore — Rust spreads structure into the type system, Go forces a few architects to hand-build it, frameworks hide it from everyone. So a first guess might be:

1. **Type-system expressiveness** lets Design influence spread across a team rather than pool in one place
2. **Gravity concentration** — the share of structural influence a few people hold — might fall into clean bands by language family
3. **Framework-driven** ecosystems might absorb gravity into the framework, leaving individual engineers with less structural load to carry

This is an empirical look at whether language is really the variable that decides concentration, using EIS (Engineering Impact Score) data from 29 major OSS repositories. The short version, before any numbers: language turns out to be the weaker signal. Two repositories in the *same* language can sit at opposite ends of the concentration range, and what separates them is governance, not syntax.

## What Gravity Counts

Gravity here is not commit volume. It is a structural *shape* passed through two honest gates, so neither surviving code alone nor catalysis alone can mint it:

```
shape    = 0.45·Design + 0.25·Breadth + 0.30·Indispensability
catGate  = 0.15 + 0.85·(Catalysis / 100)
survGate = 0.15 + 0.85·(RobustSurvival / 100)
Gravity  = catGate × survGate × shape
```

- **survGate (RobustSurvival)** asks whether the code lasts *under others' pressure* — lines surviving in modules where authors other than you keep committing. Dead-corner or self-churned code does not qualify.
- **catGate (Catalysis)** asks whether others build on that surviving foundation. It is the one axis necessarily zero for a solo author, so it witnesses that the structure observably leans on someone.

This matters for what follows. **Gravity concentration is a relational reading, not a popularity reading.** A founder whose modules no one else contests reads as quiet — the telescope cannot see influence that no one leans on. So when a category concentrates gravity, it is saying: in these universes, a few people hold the load-bearing role *and* the rest of the system still rests on them while everyone edits around it.

## Methodology

- **29 repositories** analyzed with `eis analyze`, covering **51,321 engineers** total
- Repositories grouped into **5 families** by type system and structural culture
- For each repo we read the **top contributors by Gravity** — population averages are diluted by thousands of low-activity contributors, so they tell us little
- **Gravity Concentration** = share of total gravity held by the top 3 contributors (consistent with the per-universe figures in RESULTS.md)

### Families

| Family | Characteristics | Languages | Repos |
|---|---|---|---|
| **Expressive** | Rich type system, ADTs, pattern matching, traits | Rust, Scala | 5 |
| **Go (Self-structured)** | Static, nominal typing, anti-framework culture, explicit interfaces | Go | 7 |
| **Framework-driven** | Structure delegated to framework; conventions over code | Ruby (Rails), PHP (Laravel), Java (Spring), Python (FastAPI), TS (NestJS), Elixir (Phoenix) | 6 |
| **Systems (C/C++)** | Static, manual memory, templates | C, C++ | 5 |
| **Dynamic / Structural** | Dynamic or structural typing, self-structured | JavaScript, TypeScript, Python | 6 |

These boundaries are a convenience for reading the data, not a claim that language sets the physics. As the per-repo table below shows, the variance *inside* a family routinely swamps the gap between families.

---

## Results

### Summary by Family

| Family | Repos | Avg Size | Top10 Design | Top10 Survival | Top10 Gravity | Gravity Concentration (top-3) |
|---|---|---|---|---|---|---|
| **Framework-driven** | 6 | 2,413 | 23.4 | 16.5 | 6.21 | **61.9%** |
| **Dynamic / Structural** | 6 | 1,146 | 28.5 | 14.3 | 6.48 | **60.3%** |
| **Expressive** | 5 | 1,967 | 27.5 | 15.1 | 5.74 | **54.9%** |
| **Go (Self-structured)** | 7 | 1,962 | 32.2 | 15.3 | 4.94 | **50.1%** |
| **Systems (C/C++)** | 5 | 1,280 | 24.1 | 9.7 | 3.35 | **30.0%** |

### Key Observations

#### 1. Concentration tracks governance, not language family

![Gravity Concentration by Family](chart-gravity-concentration.svg)

| Family | Gravity Concentration (top-3) |
|---|---|
| **Framework-driven** | 61.9% |
| **Dynamic / Structural** | 60.3% |
| **Expressive** | 54.9% |
| **Go (Self-structured)** | 50.1% |
| **Systems (C/C++)** | 30.0% |

The families do separate — Framework-driven concentrates the most (61.9%), Systems / C++ the least (30.0%), a ~2.1x spread — but the ordering is not the clean type-system story the hypothesis predicted. Dynamic, Expressive, and Go all land within a dozen points of each other in the middle. The single clean signal is at the bottom: Systems / C++ stands apart, where long-lived maintainers each hold real but non-dominant gravity and no small group monopolizes the load.

More telling is what happens *inside* each family. The same language sits at both ends of the range:

- **Rust** spans polars (78.4%, near-solo) to the rust compiler (27.6%, broadly shared)
- **Go** spans esbuild (77.0%, one author) to kubernetes (27.5%, deeply distributed)
- **Framework-driven** spans spring-boot (97.2%) and fastapi (94.5%) down to rails (9.8%)
- **JavaScript** spans prettier (95.3%) down to react (31.4%)

If language set the physics, polars and the rust compiler could not differ by 50 points while sharing a type system. What separates them is whether the surviving, contested structure runs through one or two people or through a standing bench. That is a governance fact, not a syntax fact. The honest reading is: **the gate measures who is still load-bearing, and being load-bearing is something a community arranges, not something a compiler decides.**

The old additive model once ranked Go most-concentrated; under the relational gate that reading does not survive. Go sits mid-pack (50.1%) because its many explicit-interface authors each hold genuine, non-dominant gravity.

#### 2. Where structure lives outside code, the telescope reads dark

The Framework-driven family carries low Top10 Design (23.4) for how much structure these projects clearly have. When Rails, Laravel, Spring, or Phoenix defines the routing, DI, middleware, and lifecycle, those decisions do not appear in any single engineer's `git blame`. The framework holds structure that no contributor's gravity can witness — EIS reads only gravity that lives in code, and gravity that lives in convention reads as Dark Matter.

But notice this does *not* mean frameworks flatten concentration. The opposite shows up: with the framework's own structure removed from view, what remains is often a single creator-architect's load, and that reads as high concentration (spring-boot 97.2, fastapi 94.5). The framework hides its own gravity; it does not hide the founder's.

#### 3. Design influence does not predict concentration

![Top 10 Design Score by Family](chart-top10-design.svg)

| Family | Top10 Design |
|---|---|
| **Go (Self-structured)** | 32.2 |
| **Dynamic / Structural** | 28.5 |
| **Expressive** | 27.5 |
| **Systems (C/C++)** | 24.1 |
| **Framework-driven** | 23.4 |

Go leads on Design because its architects write the interfaces, the routing, the middleware from scratch, all in code the telescope can see. Yet Go is only mid-pack on concentration — high Design that is *spread across many authors* does not pool into a few hands. Design measures how much structure people write; concentration measures whether the surviving structure leans on a few. The two come apart, and that gap is the whole point: a family can be design-heavy and still distributed.

---

## Deep Dive: Rails vs Laravel

![Rails vs Laravel](chart-rails-vs-laravel.svg)

Both are iconic framework-driven projects with legendary creator-architects, both creators still active. Same language folklore, same era, same pattern. Yet under the relational gate the physics diverge sharply — and in a direction language alone could never predict.

| Metric | Rails (Ruby) | Laravel (PHP) |
|---|---|---|
| Engineers | 6,056 | 4,149 |
| **Top10 Design** | **57.0** | **12.0** |
| Top10 Survival | 2.5 | 44.7 |
| Top10 Gravity (avg) | 2.44 | 9.27 |
| Gravity Concentration (top-3) | **9.8%** | **38.0%** |

### Rails: a standing bench

| # | Engineer | Gravity | Design | Indispensability |
|---|---|---|---|---|
| 1 | David Heinemeier Hansson | 4.0 | 100 | 100 |
| 2 | Jean Boussier | 3.8 | 49 | 0 |
| 3 | Rafael Mendonça França | 3.6 | 90 | 21 |
| 4 | zzak | 3.5 | 1 | 29 |
| 5 | Matthew Draper | 2.7 | 24 | 0 |
| 6 | Gannon McGibbon | 1.9 | 13 | 0 |
| 7 | Xavier Noria | 1.5 | 49 | 0 |
| 8 | Aaron Patterson | 1.4 | 97 | 29 |

DHH still holds Design 100 and Indispensability 100 — the original structure is unmistakably his — yet his gravity (4.0) sits barely ahead of the next several names. The live load has fanned out: Jean Boussier, Rafael França, zzak all carry gravity within a point of the founder, on code that survives where others keep working. Deep design history is held by many (Aaron Patterson 97, Jeremy Kemper 89, Rafael França 90, José Valim 52, Xavier Noria 49). Rails reads as a project with several design authorities and a real succession in progress.

### Laravel: a one-architect universe

| # | Engineer | Gravity | Design | Indispensability |
|---|---|---|---|---|
| 1 | Taylor Otwell | 65.1 | 100 | 100 |
| 2 | Luke Kuzmish | 6.5 | 2 | 50 |
| 3 | Mior Muhammad Zaki | 4.7 | 4 | 0 |
| 4 | Caleb White | 4.4 | 1 | 50 |
| 5 | Nuno Maduro | 2.8 | 5 | 50 |
| 6 | Patrick Carlo-Hickman | 2.6 | 0 | 0 |
| 7 | Andrew Brown | 2.0 | 4 | 0 |
| 8 | Lucas Michot | 1.8 | 5 | 0 |

Taylor Otwell holds essentially all of it: gravity 65.1, an order of magnitude above the next name, with Design 100 and Indispensability 100. No one else clears Design 5. The framework's surviving architecture runs through one person, and the gate confirms it rather than softening it.

### What this means

Both sit in the Framework-driven family, yet:

- **Rails** behaves like a self-structured project from the inside. Top10 Design 57.0 is higher than most Go projects; concentration 9.8% is the lowest in the whole sample. DHH created the structure, but the gate shows the load no longer pools in him.
- **Laravel** behaves like a creator-centric kingdom. Concentration 38.0%, Otwell at 65.1 — efficient, coherent, and a single point of failure if that one person ever steps away.

**Same framework pattern. Opposite governance physics.** The within-family gap here (9.8% vs 38.0%) is wider than most between-family gaps. Whatever decides concentration, it is not the framework label and it is not the language.

---

### Per-Repository Detail

![Gravity Concentration vs Project Size](chart-per-repo-scatter.svg)

| Repository | Family | Language | Engineers | Top10 Design | Top10 Survival | Top10 Gravity | Grav Conc (top-3) |
|---|---|---|---|---|---|---|---|
| polars | Expressive | Rust | 665 | 38.4 | 11.2 | 9.51 | 78.4% |
| swc | Expressive | Rust | 349 | 17.8 | 16.9 | 9.76 | 92.0% |
| rust | Expressive | Rust | 7,215 | 33.5 | 11.5 | 3.96 | 27.6% |
| scala | Expressive | Scala | 709 | 31.4 | 10.0 | 2.74 | 37.3% |
| scala3 | Expressive | Scala 3 | 895 | 16.6 | 26.1 | 2.75 | 38.9% |
| argo-cd | Go (Self-structured) | Go | 1,832 | 35.4 | 19.1 | 3.88 | 50.9% |
| esbuild | Go (Self-structured) | Go | 124 | 10.0 | 0.0 | 1.64 | 77.0% |
| grafana | Go (Self-structured) | Go/TS | 2,715 | 25.8 | 20.5 | 5.98 | 45.1% |
| kubernetes | Go (Self-structured) | Go | 4,510 | 15.4 | 16.2 | 4.23 | 27.5% |
| loki | Go (Self-structured) | Go | 1,272 | 48.6 | 10.0 | 4.84 | 29.8% |
| prometheus | Go (Self-structured) | Go | 1,159 | 43.1 | 14.0 | 6.84 | 64.4% |
| terraform | Go (Self-structured) | Go | 2,121 | 47.4 | 27.2 | 7.16 | 56.2% |
| rails | Framework-driven | Ruby | 6,056 | 57.0 | 2.5 | 2.44 | 9.8% |
| laravel | Framework-driven | PHP | 4,149 | 12.0 | 44.7 | 9.27 | 38.0% |
| spring-boot | Framework-driven | Java | 1,430 | 29.9 | 21.5 | 11.81 | 97.2% |
| nest | Framework-driven | TypeScript | 638 | 10.6 | 20.3 | 2.27 | 68.1% |
| fastapi | Framework-driven | Python | 861 | 11.5 | 0.0 | 1.62 | 94.5% |
| phoenix | Framework-driven | Elixir | 1,343 | 19.2 | 10.0 | 9.85 | 63.5% |
| ClickHouse | Systems (C/C++) | C++ | 2,189 | 29.0 | 11.1 | 3.41 | 33.9% |
| arrow | Systems (C/C++) | C++/Multi | 1,369 | 18.5 | 10.2 | 1.44 | 14.8% |
| duckdb | Systems (C/C++) | C++ | 618 | 16.5 | 17.3 | 8.97 | 83.5% |
| envoy | Systems (C/C++) | C++ | 1,377 | 41.0 | 10.0 | 1.66 | 7.9% |
| redis | Systems (C/C++) | C | 848 | 15.5 | 0.0 | 1.25 | 9.6% |
| eslint | Dynamic / Structural | JavaScript | 1,153 | 24.1 | 18.6 | 9.53 | 66.8% |
| express | Dynamic / Structural | JavaScript | 378 | 14.2 | 10.0 | 1.27 | 60.3% |
| prettier | Dynamic / Structural | JavaScript | 782 | 16.4 | 19.1 | 9.20 | 95.3% |
| react | Dynamic / Structural | JavaScript | 1,927 | 33.3 | 10.0 | 2.57 | 31.4% |
| superset | Dynamic / Structural | Python/TS | 1,433 | 52.5 | 13.4 | 6.09 | 47.7% |
| vite | Dynamic / Structural | TypeScript | 1,204 | 30.6 | 15.0 | 10.19 | 59.9% |
---

## Interpretation

### Concentration is a governance reading, not a language reading

![Three Modes of Structural Authority](chart-three-modes.svg)

Reading the data straight, concentration does not sort cleanly by type system. It sorts by how a community arranges its load-bearing role. Three patterns recur, and they cut *across* language families:

1. **Single-founder gravity** — one creator's surviving structure is what the system rests on, and no one else has come to contest it yet. swc (kdy1), prettier (fisker Cheung), vite (sapphi-red), spring-boot, fastapi, polars, laravel. These span Rust, JavaScript, TypeScript, Java, Python, PHP. The common thread is governance — a small core that no one else has built around — not the language.

2. **Standing bench** — several authors hold contested, survived structure at once, with succession visibly underway. The rust compiler, kubernetes, rails, react, arrow. Again multiple languages; again the common thread is community shape.

3. **Framework-as-architect** — structure lives in convention the telescope cannot read. This lowers visible Design across the family, but it does not by itself flatten concentration: with the framework's own gravity invisible, what remains is often the founder's, which reads high (spring-boot, fastapi) or, where a bench has formed inside the framework's own contributors, low (rails).

The same code can come out of any of these. What differs is the relational fact of who is still leaned on — and that is something a project grows into, not something its language hands it.

### This is not about superiority

These numbers do not rank one language over another. They suggest a quieter question worth asking of any universe: when the noise of activity is filtered out, does the surviving structure rest on one person or on several — and would the project notice if that one quietly left?

- In a **small universe** where complexity stays manageable, single-founder gravity may serve well. The cost arrives later, with scale.
- In a **large universe** with many contributors, the question is whether a standing bench has formed, or whether one person remains the only thing the structure leans on.
- The single-founder pattern is efficient until succession is needed. Whether a project can grow a bench before then is a governance choice, not a property of its syntax.

Knowing which pattern a universe operates under may help reason about scaling, succession, and where entropy is held back — or where it is one departure away from drifting.

### Limitations

- **29 repositories** is a meaningful sample but not exhaustive
- Repository maturity, governance model, and community culture are confounding — and, on this reading, they may be the *main* variables rather than confounders
- EIS observes `git blame` and commit patterns — structure expressed outside code (RFCs, ADRs, framework conventions) is invisible (Dark Matter)
- Some repositories span multiple languages (grafana: Go+TS, superset: Python+TS)
- Family boundaries are judgment calls; reasonable people may classify differently
- The Rails vs Laravel comparison shows within-family variance can exceed between-family variance — families are useful for reading, not deterministic

### Future Directions

- **Entropy resistance by family**: does any pattern hold higher Robust Survival as universe size grows?
- **Succession patterns**: which communities grow a standing bench before the founder steps away?
- **Framework effect isolation**: compare framework-driven vs bare projects in the same language
- **Scale threshold analysis**: at what universe size does single-founder gravity become a risk?
- **Governance spectrum within families**: what lets Rails grow a bench while Laravel stays single-founder?

---

## Conclusion

The hypothesis that "the laws of physics are not uniform across code universes" holds — but the reason is not the one the language folklore would suggest.

**Gravity concentration spans ~2.1x between Framework-driven projects (61.9%, most concentrated) and Systems / C++ (30.0%, most distributed).** Yet the family ordering is muddled in the middle, and the same language appears at both extremes: polars 78.4% beside the rust compiler 27.6%, esbuild 77.0% beside kubernetes 27.5%, prettier 95.3% beside react 31.4%. Language does not decide who a system leans on. **Governance does** — whether a single founder still holds the surviving, contested structure, or a standing bench has come to share it.

The Rails vs Laravel deep dive states it plainly. Same framework pattern, same era: Rails distributes (top-3 concentration 9.8%, DHH no longer leading its gravity), Laravel concentrates (38.0%, Otwell at 65.1, an order of magnitude clear). The gap between two projects in the same family is wider than the gap between most families.

For years, debates about technology choices have been fought with experience and intuition. This data is a step toward reading the structural question — who a system actually rests on — from the light the commits leave behind.

---

*Generated by [EIS (Engineering Impact Score)](https://github.com/machuz/eis) — OSS Gravity Map Project*
*29 repositories, 51,321 engineers, observed through commit light.*

### Introduction

Hi. I'm an infrastructure and backend lead at an entertainment tech company in Japan. We build poker room management software.

As the team grew, something kept bugging me: **how do you quantitatively evaluate an engineer's real ability?**

Commit count? Lines of code? PR count? All too one-dimensional. Commit granularity varies wildly between people. Line count gets inflated by auto-generated files and package-lock.json. PR count treats a typo fix and an architecture redesign as the same "1."

"This person is strong." "That person is good at politics but can't write code." — We all have these gut feelings. But gut feelings are subjective and can't survive a salary negotiation.

After a late-night brainstorming session with Claude Code (and some drinks), I built **a combat power score using nothing but git history data that matches gut feeling with surprising accuracy**. Here's how.

*Disagreements welcome. But on my team, this matched intuition so well it was almost eerie.*

### Measuring on 7 Axes

Cutting straight to the conclusion. I measure engineer combat power on these 7 axes:

| Axis | Weight | What it measures |
|---|---|---|
| Production | 15% | Lines changed. Raw output volume |
| First-pass Quality | 10% | Low fix/revert commit ratio. Getting it right the first time |
| Code Survival | **25%** | Does code you wrote still exist today? Measured with **time decay**. Design durability |
| Design | 20% | Commits to architecture files. Involvement in design decisions |
| Breadth | 10% | Number of repositories contributed to. Cross-cutting mobility |
| Debt Cleanup | 15% | How much of others' tech debt you clean up. Team contribution |
| Indispensability | 5% | Number of modules only you can maintain. Impact and risk |

I started with 5 axes, but real measurements showed I couldn't adequately capture "the person who quietly fixes others' bugs" or "dangerous knowledge silos." So I added Debt Cleanup and Indispensability.

**Code Survival has the highest weight (25%)** — this is the core thesis of the model. With time decay (detailed below), it answers: "Are you still writing durable designs *right now*?"

Quality is weighted low (10%) because commit-message-based detection is a rough proxy. Indispensability is low (5%) because "indispensable because they're good" and "indispensable because nobody else learned it" are mixed signals. The weight design itself carries meaning.

All 7 axes are **calculated separately for BE / FE / Infra**. Without this, backend code volume contaminates frontend rankings and vice versa.

### How Each Metric Is Calculated

#### Production — Lines, not commits

I initially measured commit count but switched to lines changed (insertions + deletions).

The reason is simple: **a person who changes 100 lines in 1 commit and a person who makes 100 single-line commits shouldn't score the same**.

These files are excluded (otherwise the numbers break):

- `package-lock.json`, `yarn.lock` — library updates move tens of thousands of lines
- `docs/swagger*` — auto-generated
- `mock_*`, `*.gen.*` — code generation

```bash
git log --all --no-merges --format="COMMIT:%an||%s" --numstat
```

#### First-pass Quality — What "fix" tells you

```
quality = 100 - fix_ratio
fix_ratio = fix commits / total commits × 100
```

Fix commit detection:

```python
is_fix = re.match(r"^(fix|revert|hotfix)", subject.lower())
```

A high fix ratio usually means **your code frequently needs corrections**. It's a good enough proxy for first-pass quality. However, proactive improvements and major design changes can also show up as fix commits, so this should be interpreted alongside the Design score. That's also why Quality is weighted at only 10%.

#### Code Survival — Time decay measures "current ability"

This is the heart of the metric.

Naive git blame gives high scores to "someone who wrote tons of code 3 years ago and does nothing now." That's wrong. What I want to know is: **are you still actively writing good designs?**

The solution: **exponential time decay**.

```python
import math

tau = 180  # days. Weight ≈ 0.37 at 6 months

weighted_survival = defaultdict(float)
for line in blame_lines:
    days_alive = (now - line.committer_time).days
    weight = math.exp(-days_alive / tau)
    weighted_survival[line.author] += weight
```

| Days elapsed | Weight | Meaning |
|---|---|---|
| 7 | 0.96 | Nearly full. Recent code |
| 30 | 0.85 | Still high |
| 90 | 0.61 | Just over half |
| 180 | 0.37 | About 1/3. Six months ago |
| 365 | 0.13 | Quite low. One year ago |
| 730 | 0.02 | Near zero. Two years ago |

This causes departed members' scores to naturally decline, giving a more accurate picture of **the team's current strength**. It solves the problem of someone who wrote tons of code during the founding phase but is long gone still sitting at the top of the rankings.

Raw blame line count (Raw Survival) is kept separately to visualize "who built the foundation of the codebase." The combat power score uses only the time-decayed version.

#### Design — The files you touch reveal everything

Define "architecture files" and count commits to them.

For backend:

- `repository/*interface*` — repository interfaces
- `domainservice/` — domain services
- `partprocess/` — shared business logic
- `delegateprocess/` — external system integration
- `router.go`, `middleware/` — API routing
- `di/*.go` — dependency injection

For frontend:

- `core/`, `stores/`, `types/`, `hooks/`, `lib/`, `ui/`

People who frequently commit to these files are **the people building the skeleton of the system**.

To be precise, this measures "design involvement" rather than "design skill" directly. But someone who never touches design files is unlikely to be producing good designs. It works well enough as a proxy.

#### Breadth — How many repos can you work across?

Simply count how many repositories out of all repos have commits from each person. Simple, but it cleanly separates **people who only look at their own territory from people who work across the entire team**.

#### Debt Cleanup — Who's doing the cleaning?

For fix commits, examine the "blame before the change" of modified lines. In other words, **track who wrote the code that someone else had to fix**.

```python
for fix_commit in fix_commits:
    fixer = fix_commit.author
    for changed_line in fix_commit.changed_lines:
        original_author = git_blame(file, at=parent_commit)
        if original_author != fixer:
            debt_generated[original_author] += 1  # created debt
            debt_cleaned[fixer] += 1              # cleaned debt

debt_ratio = debt_cleaned / max(debt_generated, 1)
# ratio > 1 → Cleaner (fixes others' debt)
# ratio < 1 → Creator (others fix your debt)
```

The moment I added this, **the person quietly fixing others' bugs became visible**. And the reverse — "writes a lot but causes fix work for everyone around them" — became equally obvious.

**Note**: Members with fewer than 10 total debt events (generated + cleaned) have too small a sample for reliable ratios, so they're treated as reference values.

#### Indispensability — The flip side of Bus Factor

Examine blame line distribution per module (domain directory) and **count modules where one person owns 80%+ of the code**.

```python
for module in all_modules:
    top_share = max(blame_distribution[module].values()) / total
    if top_share >= 0.8:
        critical_modules[top_author].append(module)
    elif top_share >= 0.6:
        high_risk_modules[top_author].append(module)

indispensability = critical_count * 1.0 + high_count * 0.5
```

High indispensability = modules that collapse if this person leaves. This is both a strength and a risk. The weight is kept at 5% because it's hard to distinguish "indispensable because they're great" from "indispensable because nobody else bothered to learn it." Indispensability is both an evaluation metric and **an alert for knowledge transfer priorities**.

### Normalization and Score Calculation

Each metric is **normalized by the maximum value within the same domain**.

```
norm(value, max) = min(value / max × 100, 100)
```

The person with the highest production in BE gets 100; everyone else is relative. Same within FE. **Never mix BE and FE for comparison**.

The total score is a weighted sum:

```
total = production × 0.15
      + quality    × 0.10
      + survival   × 0.25
      + design     × 0.20
      + breadth    × 0.10
      + debt_cleanup × 0.15
      + indispensability × 0.05
```

Out of 100 points.

One important note: **this metric is deliberately harsh**. When you hear "out of 100," you might picture a school test. This is nothing like that. Scoring high on all 7 axes is structurally difficult. Pushing production tends to lower quality. Raising design requires sustained involvement in architecture files. Debt cleanup only goes up if you write high-quality code yourself *while also* fixing others' bugs.

Here's a rough guide:

| Total Score | Assessment | Approx. Hourly Rate (JPY, pre-tax) |
|---|---|---|
| 80–100 | Irreplaceable core member. 1–2 people per team reach this | ¥12,000–20,000 |
| 60–79 | Near-core. Strong | ¥9,000–15,000 |
| **40–59** | **Senior engineer equivalent. Scoring 40+ on this metric means genuinely skilled** | **¥7,000–11,000** |
| 30–39 | Mid-level | ¥6,000–9,000 |
| 20–29 | Junior to mid | ¥5,000–8,000 |
| –19 | Junior | ¥3,500–6,000 |

**40 = Senior**. That's how strict this metric is. With relative scoring across 7 axes, just putting up decent numbers across the board requires serious ability. If your team's seniors are scoring 40 and you're worried — don't be. An engineer scoring in the 40s on this metric is more than competitive in the market.

### Distribution Patterns Reveal Engineer "Archetypes"

When you look at the scores, something interesting emerges. **The distribution pattern across 7 axes reveals each engineer's "archetype."**

#### Architect: Production↑ Survival↑ Design↑ Debt Cleanup↑

Writes a lot, it all survives, designs the architecture, and cleans up others' debt. The team's core. When this type leaves, the product stalls.

Quality score sometimes appears low, but this is often "fix rate goes up because they're aggressively making design changes." If Design is high, low Quality is evidence of proactive refactoring — and that's healthy.

#### Mass Producer: Production↑ Quality↓ Survival↓ Debt Cleanup↓

Writes a lot, but high fix rate and low survival. **A cycle of writing and breaking**. Debt cleanup is low too — meaning they're generating work for others.

This type *looks* productive, which makes them dangerous. They may actually be **the team's debt factory**. Evaluating on production alone makes them look great, which is the worst part.

#### Solid Cleaner: Production→ Quality↑ Survival↑ Debt Cleanup↑

Not the top producer, but low fix rate, code that survives, and quietly cleans up others' debt. **Writes durable code and takes out the team's trash**.

Unglamorous but incredibly valuable. This type's fair hourly rate should be higher than their visibility suggests.

#### Political: Breadth↑ Production↓ Survival↓ Design↓

Shows up in many repos, but production, survival, and design are all low. **Broad presence, no code contribution**.

From my own experience, I've seen this pattern in high-cost team members. Breadth high, everything else low, total score in the 20s. Numbers are brutally honest.

#### Specialist / Growing

Specialists deliver overwhelming results in narrow domains but lack cross-cutting breadth. Knowledge-silo risk exists, but they're valuable. Growing types don't produce volume yet but have low fix rates — they write carefully. If production and design grow, they'll level up.

#### Archetype Quick Reference

| Type | Prod | Qual | Surv | Design | Breadth | Debt | Indisp | Risk |
|---|---|---|---|---|---|---|---|---|
| Architect | ◎ | △–○ | ◎ | ◎ | ○ | ◎ | ◎ | — |
| Mass Producer | ◎ | ✕ | ✕ | △ | △ | ✕ | △ | **High** |
| Solid Cleaner | ○ | ◎ | ◎ | ○ | ○ | ◎ | △ | — |
| Political | ✕ | △ | ✕ | ✕ | ◎ | △ | ✕ | **High** |
| Specialist | ◎ | ◎ | ◎ | ○ | ✕ | ○ | ◎ | △ Silo |
| Growing | △ | ◎ | ○ | ✕ | △ | ○ | ✕ | — |

**Mass Producer and Political types can look like strong contributors when you only look at individual metrics**. Organizations that evaluate on production alone or breadth alone will rate these types highly. Only by combining 7 axes can you smoke them out.

### Real-World Measurement

I measured my own team (14 repos, 10+ members including departed) and am sharing a partial result.

#### BE Combat Power Rankings (Excerpt)

| # | Member | Prod | Qual | Surv RW | Design | Breadth | Debt | Indisp | Total | Type |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | machu | 100 | 57 | 100 | 100 | 74 | 100 | 43 | **90.3** | Architect |
| 2 | Member A (departed) | 69 | 73 | 12 | 67 | 81 | 11 | 100 | **52.8** | Former Architect |
| 3 | Member B | 17 | 69 | 50 | 14 | 48 | 88 | 35 | **44.5** | Solid Cleaner |
| 4 | Member C | 27 | 84 | 30 | 28 | 52 | 71 | 8 | **41.8** | Solid |
| — | Member X (departed) | 6 | 79 | ≈0 | 4 | 78 | — † | 0 | **24.9** | Political |

† Insufficient sample (fewer than 10 fix commit involvements). Neutral value of 50 used for calculation

Member X was highly paid. Total: 24.9. **Breadth high at 78, but production 6, design 4, survival near zero. The political pattern was unmistakable.** This metric would have detected it beforehand.

Member A built the initial architecture — design 67 and breadth 81 are strong numbers. Indispensability 100 is the team's highest — meaning **they're departed but still own the most modules**. After leaving, time decay dropped their RW survival to 12, but the codebase is still largely built on their foundation. The numbers clearly show they were an Architect type while active.

Member B's production of 17 isn't flashy, but RW survival 50 (2nd place) = recently written code that keeps surviving. Debt cleanup 88 = quietly fixing others' bugs. **Adding debt cleanup to the model is what finally gives this type of person proper recognition.**

My own quality score (57) is low, but this reflects aggressive architecture changes (introducing DelegateProcess layer, designing PartProcess layer) and a style of iterating on abstract domain concepts in code. Combined with Design 100, it reads as proactive improvement rather than careless coding.

#### FE Combat Power Rankings (Excerpt)

| # | Member | Prod | Qual | Surv RW | Design | Breadth | Debt | Indisp | Total | Type |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | Member D | 100 | 84 | 100 | 100 | 62 | 39 | 100 | **85.4** | Architect |
| 2 | Member E | 45 | 58 | 23 | 23 | 100 | 69 | 39 | **45.1** | Breadth |
| — | Member Y (departed) | 24 | 18 | ≈0 | 17 | 38 | 0 | 0 | **12.6** | Mass Producer |

Member D's indispensability of 100 mirrors Member A on the BE side — **they own nearly all of the FE core library (180,000+ lines)**. The biggest risk of the Architect type: the moment they leave, nobody can make FE design decisions. Their debt cleanup of 39 is mid-range within FE, but with 129 self-fixes, they show a self-contained style of writing and correcting their own code.

Member Y's quality of 18 is a brutal number. **82% of commits are fixes or corrections**. RW survival ≈ 0, debt cleanup 0. Wrote a lot, fixed a lot, none of it survived. Never cleaned anyone else's debt either.

### Why Revenue KPIs Alone Can't Show Engineering Org Health

This section is for executives and managers.

"Revenue is growing, so the engineering org is fine" — this is a dangerous misconception. Revenue KPIs measure **product-market fit**, not **engineering organization health**.

An analogy: revenue is "vehicle speed," engineering org state is "engine condition." A failing engine can still go fast downhill. Speed doesn't mean the engine is healthy.

To assess engineering health, you need **4 metrics invisible from revenue**:

#### 1. Code Durability (Survival)

Is code still running 6 months later without design changes?

Low-survival organizations are **rewriting the same features over and over**. It doesn't directly affect revenue, but development velocity gradually decays. A team that spends every quarter "rebuilding what we built last quarter" will never scale.

#### 2. Debt Accumulation Rate

Does every new feature increase the volume of fixes to existing code?

Organizations with low debt cleanup ratios reach a state where **adding 1 feature generates 2 bug fixes** in existing code. If revenue is growing but "development feels slow," look here.

#### 3. Bus Factor Risk (Indispensability)

How many modules stop functioning if a specific engineer leaves?

On my team, a departed member owns 97% of blame lines in certain modules. Without someone who understands that person's design philosophy, **even bug fixes become too scary to attempt**.

#### 4. Design Decision Concentration

Who are architectural decisions concentrated in?

If everything concentrates on a single architect, their mistakes propagate across the entire product. Conversely, a team with distributed design scores has **evidence that design review is functioning**.

---

**Even with revenue growing, if Survival decline + Debt increase + Bus Factor concentration are progressing simultaneously, the organization will collapse at scale.**

The combat power score quantifies these and simultaneously visualizes "engineers worth investing in" and "organizational risks." Organizations that have nothing but "gut feeling" for engineer evaluation — just producing these numbers will change the landscape.

### Limitations and How to Read This Metric

No metric is perfect. This one has limitations too:

- **Fix rate depends on commit messages**. Inconsistent messaging reduces accuracy. That's why Quality is weighted at 10%
- **Blame is slow on large repos**. Sampling may be needed
- **FE has higher refactor frequency**, so survival tends to be lower → that's why BE/FE are separated
- **Line count grows with copy-paste too** → offset by survival and design scores
- **Debt cleanup depends on sample size**. Members with few fix commits get extreme ratios, so a threshold (fewer than 10 = reference value) is applied
- **Indispensability mixes "strength" and "organizational dysfunction"**. That's why it's weighted at 5%

Not omnipotent, but **it functions more than well enough as "backing for intuition."** At the very least, it's 100x better than "gut feeling."

#### Accuracy Improves with Better Design

This metric has a property where **higher codebase design quality yields higher accuracy**.

In codebases with clear layer separation (Clean Architecture, DDD), "touching design files = making design decisions" holds true, and high Survival can be interpreted as "the design withstood change pressure." Conversely, in codebases with no design philosophy and chaotic file structures, high Survival might just mean "dead code nobody touches," and low Design scores might just mean "architecture files don't clearly exist."

In other words, **the metric's low accuracy is itself a signal of poor design**. If you measure and find "this doesn't match gut feeling at all," it may not be the metric that's wrong — it may be that the codebase structure can't withstand measurement. Investing in design is itself infrastructure for improving team evaluation accuracy.

#### Notes on Reading Survival

Raw git blame has a bias toward people who did massive initial implementation. Code that nobody touches still counts as their lines as long as it exists.

This article avoids this problem by using **time-decayed Survival** instead of raw blame. Code from 2 years ago contributes only 0.02, emphasizing "designs that still stand today."

Even so, Survival shouldn't be interpreted alone. **Combine it with Design and Bus Factor** to distinguish "surviving because it's good design" from "surviving because nobody touches it." If the Survival leader and Design leader are the same person, that's healthy. If they're different, there may be "untouchable debt" lurking in the codebase.

#### What's your engineering leader's combat power?

If you're reading this and thinking "let's try it" — I have a question.

**What rank is the person making engineering decisions at your company, within the team?**

If you're serious about building a strong in-house engineering org, **the technical decision-maker's combat power must be the highest (or at least near the top) on the team**. Because the correctness of design decisions shows up in code. The person deciding architecture writes code themselves, and that code survives — this is the only way to guarantee design quality.

The reverse: when someone with low combat power sits in a decision-making position, **high-combat-power members' design judgments get overruled by a low-scoring superior**. This is lethal for the organization.

**An organization where someone with no trace in the codebase is making design decisions** may be structurally dysfunctional. If you measure and find your boss's combat power is near the bottom of the team — that's not the metric's fault. It might be a structural problem with the organization.

### Measure Every 3 Months

This metric's value comes from regular measurement.

If survival has improved compared to 3 months ago, that member's **design ability is growing**. If fix rate has dropped, **first-pass quality is improving**. If debt cleanup ratio has risen, **team contribution is increasing**.

Conversely, if only production is up while everything else is flat, **volume increased but quality didn't change**.

Numbers don't lie.

### Computation Cost

This metric **consumes a fair amount of tokens when computed by AI**. At 14-repo, 10-person scale, running git log/blame/fix commit tracking takes **30–60 minutes**, consuming 500K–1M tokens. Including trial and error, it can exceed 1.5M tokens.

Honestly, you need a flat-rate plan like Claude Max to run this casually. API-based pricing would cost thousands to tens of thousands of yen per measurement. But once the formulas are finalized, you can automate with git commands and scripts, bringing ongoing AI cost to zero.

### Closing

Evaluating engineers is hard. But "hard" doesn't mean "impossible to quantify."

Using nothing but git history — data every team already has — I built a score that matches intuition to a surprising degree. Of course, this metric alone shouldn't determine everything. Code review quality, documentation, team contributions — there's enormous value that doesn't show up in numbers.

But **first, quantify what can be quantified**. Then supplement with qualitative assessment for what can't. That order matters.

An organization where "that person is great" and "that person is meh" exists only as unspoken atmosphere, without quantitative evaluation, is not healthy. **Evaluation requires resolution on the evaluator's side.** This combat power score is a tool for increasing that resolution.

Even as AI-driven development increases, this metric can tell you whether you're building on good foundations, having productive conversations with AI, and getting it to write good code. Whether you can build good foundations through AI dialogue matters too. There's still room for human judgment there. Though general-purpose AI might eliminate that room too, lol.

And one more thing. When the era of general-purpose AI fully arrives, **this could work as a metric for measuring the "rigidity" of AI-written code**. In an age where AI generates massive amounts of code — what's the survival rate? Is it still there in blame 6 months later? Or did humans rewrite all of it? When viewed through debt cleanup ratios, is the AI a debt factory or a cleaner? — Git records humans and AI equally. Instead of qualitative reports like "development speed improved with AI," **quantitatively tracking AI-written code's survival rate and debt ratio** feels like the next phase of evaluation.

One last thing. **This score doesn't reduce human value to a number — it's an approximation of technical influence as it appears in the codebase.** The ability to mentor juniors through code review, depth of domain knowledge, creating psychological safety on a team — there's an enormous amount of contribution that doesn't appear in numbers. This metric only quantifies what can be read from "traces left in git." It's not omnipotent. But it's infinitely better than zero.

To build the strongest team in the world, start by knowing where you actually stand. Measure your team's combat power. Numbers are more honest than you think.

---
title: "Every engineering metric gets gamed. One of them structurally can't."
series: "OrbitLens Ace"
published: false
description: "DORA, velocity, commit counts, lines, even churn — every metric that measures activity gets gamed, because activity is producible at will. Survival is the one signal in git you cannot inflate without actually leaving code that lasts. Here is the mechanism, and the data."
tags: opensource, ai, git, metrics
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-ace-mark.png?v=1
---

<a href="https://ace.orbitlens.io"><img src="https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-ace-mark.png?v=1" alt="OrbitLens Ace" width="72"></a>

**[OrbitLens Ace → ace.orbitlens.io](https://ace.orbitlens.io)**

*You can fake a busy quarter. You cannot fake what is still there two years later.*

---

Every engineering metric that has ever been used to judge people has been gamed.

Lines of code rewarded whoever typed the most, so people typed more. Commit counts rewarded whoever committed the most, so people split commits. Velocity rewarded whoever closed the most points, so points inflated. DORA measured deployment frequency, so teams deployed trivia. Even code churn — the metric the "code health" tools lean on — rewards low numbers, and a number you can lower on purpose is a number you can manage instead of the thing underneath it.

This is not a story about dishonest engineers. It is Goodhart's law, and it is structural. **Every one of those metrics measures *activity*. And activity is producible at will.** The moment a measure of activity becomes a target, the cheapest way to move it is to produce more activity — not more of whatever the activity was supposed to stand for.

So the honest question is not "which activity metric is best?" It is: **is there anything in a git history that you cannot move by being busier?**

There is exactly one thing. Not because we were clever, but because of what it is made of.

## What survives is not something you do

Take every line a person wrote. Wait. Come back months later and ask a narrower question than "did they work hard?" Ask: *is that specific line still there?* Not rewritten. Not reverted. Not quietly deleted in someone else's refactor. Still load-bearing at HEAD.

That is survival. In EIS we measure it with time-decayed `git blame`: a line's weight falls off as the months pass unless the line keeps existing, and it counts for more when *other* people have built on top of it rather than leaving it as a private island. Survival weighted by others building on it is what we call **gravity** — structural pull that lasts.

Now watch what happens to each way of gaming it.

- **Split your commits into a hundred.** Survival counts surviving *lines*, never commits. A hundred commits that touch the same ten lines leave ten lines. The denominator does not see your commit graph at all.
- **Write busywork — reformatting, churn, motion.** Code that gets rewritten is, by definition, code that did not survive. Churn *removes* itself from the number. The harder you thrash a file, the less of your thrashing is there later to count.
- **Write an enormous volume.** Only the fraction that lasts survives. Volume with a short half-life decays to nothing on the same schedule as anyone else's.
- **Build a private empire nobody touches.** The gravity weighting asks whether *others* built on your code. You cannot supply that yourself. It is contributed by other people, over time, and it is the one input to the score that is not in your hands.

Every gaming vector routes through activity. Survival is what is *left when the activity is subtracted*. You cannot inflate the residue by adding more of the thing that gets subtracted.

## The data says volume barely predicts it

If survival were just a fancier way of counting output, the people who commit the most would be the people whose code survives the most. They are not.

Across seven open-source repositories — 547 real contributors — the rank correlation between a person's commit count and the mass of their surviving code is **ρ = 0.28**. Commit volume explains a single-digit share of who is still standing in the codebase. In three of the seven, the top committer is *not* the top survivor. Being the busiest person in the git log tells you surprisingly little about whose work the codebase actually kept.

We saw the same thing, sharper, inside a small production team. At one point in time, one engineer held roughly fifteen times the surviving-code mass of a peer — by the activity story, a landslide. Two years later, most of that lead had been overwritten by the normal churn of a living codebase, and the peer's code had quietly become the part everyone else was building on. The volume did not survive. The structure did. No performance review reversed that ranking — the git history did, on its own, by simply letting time pass over both.

(The names do not matter and are not the point. We do not publish them — a measure that would rank a colleague in public is exactly the kind of measure we are arguing against. Survival is read at the level of the code, never the person's worth.)

## Why this matters now, and not five years ago

For most of software's history, activity was a *decent proxy*. If you wanted a lot of surviving code, one workable route really was to write a lot of code. The proxy held because writing was expensive, so volume correlated with intent.

AI breaks the proxy. When a model can emit a thousand lines a minute, activity stops being scarce, and a metric that measures a thing which is no longer scarce measures noise. Commit counts, diff sizes, PR throughput — in an AI-heavy repo they inflate for free and mean nothing.

Survival is the layer left standing. A model can generate code; it cannot decide whether that code lasts. That is decided later, by whether the code holds under change and whether other people build on it — the two inputs no author, human or machine, controls. Which is why, as everything else collapses into noise, the surviving layer is the only signal in the repository that still carries information.

## What survival is *not* — because the honesty is the product

Here is where we part company with the rest of the category, which sells fear and certainty. Survival is not virtue. Code can survive because it is excellent, or because it sits in a corner nobody dares touch — persistence and value are not the same thing, and we spend real effort separating "lasted because others built on it" from "lasted because it was abandoned." The gate that does that separation needs a *crowd* to work; on a three-person team there are not enough independent hands for it to fully fire, and the measure leans back toward raw persistence.

And survival is not a forecast. It tells you what *did* last. We have tested, hard, whether it predicts what *will* last, and across many repositories it does not do that cleanly — a result we will write up in full, including the parts that did not go our way, because an observatory that only reports its wins is not an observatory.

That is the whole posture. We are not claiming a number that tells you who to promote or which file will break. We are claiming one honest thing: **of everything git records, survival is the only measurement you cannot move by being busier.** In a year when busyness became free, that turns out to be the only measurement left worth reading.

---

The telescope does not bend. It reads what is there — the code that survived, and the pull it still exerts — and it refuses to turn that reading into a scoreboard over people. If that is the instrument you want pointed at your own repository, it runs locally, on git alone, and sends nothing outward.

**[Read your own surviving layer → ace.orbitlens.io](https://ace.orbitlens.io)** · **[EIS is open source →](https://github.com/machuz/eis)**

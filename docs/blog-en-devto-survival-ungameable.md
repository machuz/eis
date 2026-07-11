---
title: "Every engineering metric gets gamed. One of them structurally can't."
series: "OrbitLens Ace"
published: false
description: "Lines, commits, velocity, DORA, even churn — every metric that measures activity ends up gamed, because activity is cheap to produce. There is one number in a git history you can't move by being busier. Here's why, and the data."
tags: opensource, ai, git, metrics
cover_image: https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-ace-mark.png?v=1
---

<a href="https://ace.orbitlens.io"><img src="https://raw.githubusercontent.com/machuz/eis/main/docs/images/logo-ace-mark.png?v=1" alt="OrbitLens Ace" width="72"></a>

**[OrbitLens Ace → ace.orbitlens.io](https://ace.orbitlens.io)**

*A busy quarter is easy to stage. Code that's still there in two years isn't.*

---

Pick any metric a team has ever used to judge people, and someone has quietly figured out how to move it without doing the underlying thing.

Lines of code rewarded typing, so people typed. Commit counts rewarded committing, so commits got smaller and more frequent. Velocity rewarded closed points, and points drifted upward until a "3" meant nothing. DORA measured how often you deploy, so trivial changes started shipping on their own. Even churn — the number the "code health" tools lean on — is something you can lower on purpose, which means you can manage the number instead of the mess underneath it.

None of that requires dishonest engineers. It's Goodhart's law doing what it always does. Every one of those numbers is a measure of *activity*, and activity is cheap to produce. Once you're paid for activity, the fastest way to get paid more is to produce more of it — not more of whatever the activity was supposed to be a sign of.

So the question worth asking isn't which activity metric is least bad. It's whether a git history contains anything at all that you can't move just by being busier.

It turns out there's one. And it's not because we were clever — it's because of what the thing is actually made of.

## What lasts isn't something you do

Take everything a person wrote, wait a while, and ask a smaller question than "did they work hard." Ask whether the specific lines are still there. Not reverted, not rewritten, not quietly swallowed by someone else's refactor. Still holding weight at HEAD.

That's survival. We read it with time-decayed `git blame`: a line's weight fades month by month unless the line keeps existing, and it counts for more once *other* people have built on top of it instead of leaving it as a private island. Survival that others have built on is what we call gravity — the structural pull that outlives the person who created it.

Try to game it and watch where each trick goes.

Split one commit into a hundred, and survival still counts the surviving lines, never the commits — a hundred commits over the same ten lines leave you ten lines. Write reformatting and busywork, and it deletes itself from the score, because code that gets rewritten is, by definition, code that didn't survive; the harder you thrash a file, the less of the thrashing is left to count later. Write an enormous volume, and only the part that lasts registers; a large body of code with a short half-life decays on the same clock as anyone else's.

The last trick is the one that gives it away. Build a private empire nobody else touches, and the gravity weighting asks whether *others* built on your code — and that's the one input you can't supply yourself. Other people contribute it, over months, by choosing to build on your work or choosing to tear it out. It's the only term in the score that never passes through your own hands.

Every gaming move runs through activity. Survival is what's left after you subtract the activity. You can't inflate the remainder by adding more of the thing that gets subtracted.

## Volume barely predicts it

If survival were just output with better makeup, the people who commit the most would be the ones whose code lasts the most. They aren't.

Across seven open-source repositories and 547 real contributors, a person's commit count and the surviving mass of their code correlate at ρ = 0.28. Being busy in the log explains a single-digit slice of who's still standing in the codebase. In three of the seven, the person with the most commits isn't the person with the most surviving code at all.

I saw a sharper version of this inside a small production team. At one point one engineer held roughly fifteen times the surviving-code mass of a colleague — a landslide, if you were reading the activity. Two years on, most of that lead had been overwritten by the ordinary churn of a codebase that was still alive, and the colleague's work had quietly become the part everyone else was building against. Nobody ran a review to correct the ranking. Time did, by passing over both of them the same way.

(The names aren't the point, and we don't publish them. A number that would rank a coworker in public is exactly the thing we're arguing against. Survival is read at the level of the code, never the level of a person's worth.)

## Why this lands now

For most of software's history, activity was a decent stand-in. If you wanted a lot of surviving code, writing a lot of code really was one way to get there — writing was expensive, so volume tracked intent well enough.

Then a model started writing a thousand lines a minute, and the stand-in came apart. Commit counts, diff sizes, PR throughput — in a repo where half the code is generated, they inflate for free. A measure of something that's no longer scarce is a measure of nothing.

Survival is the layer that's still standing. A model can produce code; whether the code lasts gets decided afterward, by whether it holds under change and whether anyone builds on it — the two things no author controls, human or otherwise. So as the rest of the dashboard fills with noise, the surviving layer is the last part of the repository that still carries a signal.

## What survival isn't

Here's where we walk away from the rest of the category, which mostly sells fear. Survival isn't virtue. Code can last because it's good, and code can last because it sits in a corner everyone's afraid to touch — persistence and value aren't the same thing, and separating "lasted because others built on it" from "lasted because it was abandoned" is real work, work that needs a crowd to do at all. On a three-person team there aren't enough independent hands for that separation to fire, and the measure leans back toward raw persistence. We say so.

And survival doesn't predict. It tells you what lasted, not what will. We spent real effort testing whether it forecasts durability, and across a lot of repositories it doesn't do that cleanly — a result we'll write up in full, the inconvenient parts included, because an observatory that only reports its wins isn't one.

That's the whole claim, and it's a small one. Not a number that says who to promote or which file breaks next. Just this: of everything git records, survival is the one measurement you can't move by being busier. In a year where being busy got free, that turned out to be the only measurement left worth reading.

---

The telescope doesn't bend. It reads what's on the ground — the code that survived, and the pull it still has — and it won't turn that into a scoreboard over people. If that's the instrument you want pointed at your own repo, it runs locally, on git alone, and sends nothing out.

**[Read your own surviving layer → ace.orbitlens.io](https://ace.orbitlens.io)** · **[EIS is open source →](https://github.com/machuz/eis)**

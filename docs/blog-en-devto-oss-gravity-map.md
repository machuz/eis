---
title: "The gravity moved: 10 years of React's git history, and who actually held it together"
published: true
description: "I read 29 open-source repos and 55,343 engineers with nothing but git log and git blame. React's center of gravity shifted across five generations in ten years; esbuild's structure rests on a single mind; 440 engineers hold real weight while staying invisible. A telescope for structural influence, not activity."
tags: opensource, git, productivity, career
---

Dan Abramov held React together in 2020. By 2024, he didn't — and the person who did, most people outside the project can't name.

I didn't interview anyone to learn this. I didn't read a single release note. I only read what the repository had already been writing about itself, every day, for ten years: its git history.

There's a quiet assumption in how we talk about engineers — that the people who matter are the ones we can see. The loud reviewer. The name on the conference talk. The top of the commit-count leaderboard. But a codebase records a different story, and it doesn't know how to flatter anyone.

So I read it. 29 open-source repositories. 55,343 engineers. Nothing but `git log` and `git blame` — no surveys, no AI tokens, no self-reports.

A few things surfaced that I haven't stopped thinking about:

- **React's center of gravity shifted, on schedule.** Across 2016–2025, influence didn't stay with its famous faces. It migrated — Abramov rising to an Architect peak around 2020, then receding; Sebastian Markbåge settling into the highest Anchor signal in the dataset by 2024. Five generational hand-offs, written plainly into the history. Nobody announced them. The code did.
- **esbuild's entire structure rests on a single mind.** 92.5% of its gravity concentrates in one engineer. That isn't fragility — it can be the cleanest coherence there is: an entire design held in one head, kept consistent in a way no committee reliably manages. Gravity doesn't have to be distributed to be real.
- **440 engineers hold real structural weight while staying invisible.** No maintainer badge, no public profile to speak of. They surface only when you stop counting activity and start measuring what survives.
- **Language families have their own physics.** Go repos concentrated gravity around 16%; Rust and Scala around 7%. Roughly a 4–5× gap in how much one person can come to hold.

I want to be careful about what this is. It isn't a scoreboard, and it isn't a verdict on anyone's worth. The signal is deliberately narrow: not how much you committed, but how much of what you committed is still standing months later — and whether it's standing because the design holds under change, or only because nobody dares to touch it. The first I'd call gravity. The second is something quieter, and easy to mistake for strength.

Maybe that's the part worth sitting with. Activity is the easiest thing in the world to see, and it says almost nothing about who is holding the structure together. The engineers who shape gravity tend to be the ones a team only notices the moment they leave.

The lens is open source — a small CLI called EIS that reads only git metadata, never the code itself. The full map of all 29 repositories, React's decade included, is here: https://machuz.github.io/eis/research/oss-gravity-map/analysis/oss-gravity-map-en.html

I'll leave you with the question I can't shake. If you ran this on your own repository tonight — not the commit counts, the gravity — would the names that came up be the names you'd expect?

— machuz

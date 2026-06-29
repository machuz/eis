# Canonical brand & logo assets

This file is the source of truth for which logo marks are current. Use only the
assets listed below. If you need a logo somewhere, point at one of these — do not
introduce a new variant or revive a retired one.

## Current assets

| Asset | What it is | Notes |
|---|---|---|
| `logo-ace-mark.svg` | **OrbitLens Ace product mark** — the gold orbit-ring (a hollow-centre lens, brushstroke concentric rings). | The product mark for Ace. Copy of the canonical `lp/logo-ace.svg` in the orbit repo. |
| `orbitlens-royal-purple-shadow.png` | **OrbitLens brand icon** — the purple orbit-ring. | Used as favicon and in nav/breadcrumb lockups across the library. |
| `logo-full.svg` / `logo-full.png` | **EIS lockup** — the gold orbit-ring mark + the wordmark "EIS — the Git Telescope — Engineering Impact Signal". | Used as the sign-off footer in the blog articles. **Referenced by published blog posts via absolute `raw.githubusercontent.com/machuz/eis/main/docs/images/logo-full.png` URLs — do NOT move or rename this path**, or the live posts will 404. Re-render the PNG in place via `rsvg-convert` if the SVG changes. |

## Retired — do not reintroduce

- The old dark **"radar" aperture mark** (a dark circle with tick/blade marks and
  a starfield interior — formerly `logo-icon.svg` / `logo-icon.png` /
  `logo-icon-transparent.*`). Fully removed from the repo.
- Any **"Engineering Impact Score"** wordmark. The metric is **Signal**, not Score.

If a mark looks like a dark radar dial, or a wordmark says "Score", it is wrong —
replace it with one of the current assets above.

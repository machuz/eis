package scorer

import "math"

// minTrustedPool is the smallest contributor pool for which max-normalization
// yields a trustworthy 0..100 spread. Max-normalization always pins the largest
// author at 100, so in a one- or two-person domain "you are the max" means only
// "you are the largest of one or two people" — not a real percentile against a
// population. Below this size we damp the whole pool's normalized scores by a
// confidence factor, so a tiny domain cannot mint a confident 100 on any axis.
const minTrustedPool = 3

// poolConfidence is the trust we place in a pool's relative ranking. With fewer
// than minTrustedPool contributors, max-normalization is not a real percentile —
// the largest of one or two people is pinned at 100 either way — so we halve the
// pool's normalized scores: untrusted, but not erased. We deliberately do NOT
// zero a lone author's individual axes (survival, design, breadth are real,
// observable facts even with no one to rank against — erasing them would be
// evaluation, not observation). The relational collapse a solo author should see
// is carried by gravity's catGate, where Catalysis is genuinely zero with no one
// building on you — not by deleting the observation here.
func poolConfidence(n int) float64 {
	if n >= minTrustedPool {
		return 1.0
	}
	return 0.5
}

// Normalize max-normalizes values to 0..100, then damps by poolConfidence(pool)
// so small domains do not produce untrustworthy maxed-out scores. `pool` is the
// domain's contributor count (NOT len(values): an axis map omits zero-valued
// authors, so it would understate the true pool).
func Normalize(values map[string]float64, pool int) map[string]float64 {
	maxVal := 0.0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	conf := poolConfidence(pool)
	result := make(map[string]float64)
	for author, v := range values {
		if maxVal == 0 {
			result[author] = 0
		} else {
			result[author] = math.Min(v/maxVal*100, 100) * conf
		}
	}
	return result
}

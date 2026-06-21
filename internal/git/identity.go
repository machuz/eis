package git

import "strings"

// BuildIdentityMap derives a raw-name → canonical-name mapping from commit
// authorship, collapsing the common case where ONE person commits under several
// display names sharing a single email (e.g. "sebmarkbage" / "Sebastian
// Markbåge", or "Joseph Savona" / "Joe Savona"). Author email is the stable
// identity; display names drift, so name-only aggregation splits a single
// engineer's survival, catalysis, and gravity across phantom identities.
//
// Per email, the canonical name is the one with the most commits. A name that
// appears under several emails is attached to whichever email it commits under
// most. The returned map contains only entries that actually change a name
// (name != canonical); callers apply it to commit and blame authors BEFORE any
// config alias resolution, so config aliases still handle the residue (one
// person across several emails whose names never coincide).
//
// Ordering is deterministic: ties break toward the higher commit count, then the
// lexically smaller name, so the same repo always yields the same canonical set.
func BuildIdentityMap(commits []Commit) map[string]string {
	// email -> name -> commit count
	emailNames := map[string]map[string]int{}
	// name -> email -> commit count
	nameEmails := map[string]map[string]int{}

	for _, c := range commits {
		email := strings.ToLower(strings.TrimSpace(c.Email))
		name := c.Author
		if email == "" || name == "" {
			continue
		}
		if emailNames[email] == nil {
			emailNames[email] = map[string]int{}
		}
		emailNames[email][name]++
		if nameEmails[name] == nil {
			nameEmails[name] = map[string]int{}
		}
		nameEmails[name][email]++
	}

	// canonical name per email = most-committed name (deterministic tiebreak)
	emailCanon := make(map[string]string, len(emailNames))
	for email, names := range emailNames {
		emailCanon[email] = pickTop(names)
	}

	out := map[string]string{}
	for name, emails := range nameEmails {
		// attach this name to the email it commits under most, then take that
		// email's canonical name.
		canon := emailCanon[pickTop(emails)]
		if canon != "" && canon != name {
			out[name] = canon
		}
	}
	return out
}

// pickTop returns the key with the highest count; ties break toward the
// lexically smaller key so the result is deterministic across runs.
func pickTop(counts map[string]int) string {
	best := ""
	bestN := -1
	for k, n := range counts {
		if n > bestN || (n == bestN && k < best) {
			best, bestN = k, n
		}
	}
	return best
}

// CanonicalizeAuthors rewrites Commit and BlameLine author names in place using
// the email-derived identity map from BuildIdentityMap. Blame lines carry no
// email, so they are remapped by name — which is sound because the map keys on
// the same display names git records in both streams.
func CanonicalizeAuthors(commits []Commit, blame []BlameLine, idmap map[string]string) {
	if len(idmap) == 0 {
		return
	}
	for i := range commits {
		if c, ok := idmap[commits[i].Author]; ok {
			commits[i].Author = c
		}
	}
	for i := range blame {
		if c, ok := idmap[blame[i].Author]; ok {
			blame[i].Author = c
		}
	}
}

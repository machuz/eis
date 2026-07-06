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
	acc := NewIdentityAccumulator()
	for i := range commits {
		acc.Add(commits[i].Author, commits[i].Email)
	}
	return acc.Build()
}

// IdentityAccumulator builds the raw-name → canonical-name identity map
// incrementally, one (name, email) observation at a time, so a streaming log
// walk can derive the map without materializing []Commit. Feed every commit's
// (Author, Email) via Add, then call Build. Build is a pure function of the
// accumulated counts, so the result is identical to BuildIdentityMap over the
// same commits regardless of Add order (counts + deterministic pickTop).
type IdentityAccumulator struct {
	// email -> name -> commit count
	emailNames map[string]map[string]int
	// name -> email -> commit count
	nameEmails map[string]map[string]int
}

// NewIdentityAccumulator returns an empty identity accumulator.
func NewIdentityAccumulator() *IdentityAccumulator {
	return &IdentityAccumulator{
		emailNames: map[string]map[string]int{},
		nameEmails: map[string]map[string]int{},
	}
}

// Add records one commit's (author name, email) observation.
func (a *IdentityAccumulator) Add(name, email string) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || name == "" || isSharedEmail(email) {
		return
	}
	if a.emailNames[email] == nil {
		a.emailNames[email] = map[string]int{}
	}
	a.emailNames[email][name]++
	if a.nameEmails[name] == nil {
		a.nameEmails[name] = map[string]int{}
	}
	a.nameEmails[name][email]++
}

// Build finalizes the identity map (only entries where name != canonical).
func (a *IdentityAccumulator) Build() map[string]string {
	// canonical name per email = most-committed name (deterministic tiebreak)
	emailCanon := make(map[string]string, len(a.emailNames))
	for email, names := range a.emailNames {
		emailCanon[email] = pickTop(names)
	}

	out := map[string]string{}
	for name, emails := range a.nameEmails {
		// attach this name to the email it commits under most, then take that
		// email's canonical name.
		canon := emailCanon[pickTop(emails)]
		if canon != "" && canon != name {
			out[name] = canon
		}
	}
	return out
}

// sharedEmails are addresses used by MANY distinct people, so grouping names
// under them would merge unrelated engineers. Note: a per-user GitHub noreply
// (e.g. "user@users.noreply.github.com" or "1234+user@users.noreply.github.com")
// is NOT shared — it identifies one account, and collapsing its name variants is
// exactly the intent — so only the generic, account-less addresses are excluded.
var sharedEmails = map[string]bool{
	"web-flow@github.com":                                   true, // GitHub web-UI commits, authored by anyone
	"noreply@github.com":                                    true,
	"actions@github.com":                                    true,
	"github-actions@github.com":                             true,
	"githubactions@github.com":                              true,
	"41898282+github-actions[bot]@users.noreply.github.com": true,
}

// isSharedEmail reports whether an email is a shared/automation address that must
// not anchor identity merges. A guard against the per-name-count heuristic's
// failure mode: real prolific contributors legitimately commit under many name
// spellings from one personal email (e.g. swc's kdy1 uses 5), so a low
// name-count cutoff would wrongly split them; an explicit address denylist
// targets only the genuinely-shared addresses instead.
func isSharedEmail(email string) bool {
	if sharedEmails[email] {
		return true
	}
	// "noreply@" with no user part, or no "@" at all, can't identify a person.
	if !strings.Contains(email, "@") {
		return true
	}
	return false
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
		for j, ca := range commits[i].CoAuthors {
			if c, ok := idmap[ca]; ok {
				commits[i].CoAuthors[j] = c
			}
		}
	}
	for i := range blame {
		if c, ok := idmap[blame[i].Author]; ok {
			blame[i].Author = c
		}
	}
}

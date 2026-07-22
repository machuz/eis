package git

import "testing"

// PrimaryEmails must key by the CANONICAL name — the name every downstream
// author string has already been rewritten to — not by each raw spelling.
func TestPrimaryEmails_KeysByCanonicalName(t *testing.T) {
	a := NewIdentityAccumulator()
	// One person, two name spellings, one personal email. "Alice Smith"
	// commits more, so it is canonical.
	a.Add("Alice Smith", "alice@example.com")
	a.Add("Alice Smith", "alice@example.com")
	a.Add("alice", "alice@example.com")

	got := a.PrimaryEmails()
	if got["Alice Smith"] != "alice@example.com" {
		t.Fatalf("canonical name missing or wrong: %#v", got)
	}
}

// A GitHub noreply address is the payload that makes this worth emitting: it
// carries the numeric user id, so an account can be resolved from the clone
// with no API call.
func TestPrimaryEmails_KeepsNoreplyAddress(t *testing.T) {
	a := NewIdentityAccumulator()
	a.Add("Dev One", "12345+devone@users.noreply.github.com")

	if got := a.PrimaryEmails()["Dev One"]; got != "12345+devone@users.noreply.github.com" {
		t.Fatalf("noreply address dropped: %q", got)
	}
}

// Shared/automation addresses identify nobody, so they must not be emitted as
// somebody's primary email — the same rule Build() applies when merging.
func TestPrimaryEmails_ExcludesSharedAddresses(t *testing.T) {
	a := NewIdentityAccumulator()
	a.Add("Some Contributor", "web-flow@github.com")
	a.Add("Bot Runner", "github-actions@github.com")

	got := a.PrimaryEmails()
	if _, ok := got["Some Contributor"]; ok {
		t.Fatalf("web-flow@github.com emitted as a personal email: %#v", got)
	}
	if _, ok := got["Bot Runner"]; ok {
		t.Fatalf("github-actions@github.com emitted as a personal email: %#v", got)
	}
}

// Determinism (W-02): the map is a pure function of the accumulated counts, so
// Add order must not change it.
func TestPrimaryEmails_OrderIndependent(t *testing.T) {
	build := func(order []([2]string)) map[string]string {
		a := NewIdentityAccumulator()
		for _, p := range order {
			a.Add(p[0], p[1])
		}
		return a.PrimaryEmails()
	}
	fwd := build([][2]string{
		{"Alice", "a@example.com"}, {"Alice", "a@example.com"},
		{"Bob", "b@example.com"}, {"alice", "a@example.com"},
	})
	rev := build([][2]string{
		{"alice", "a@example.com"}, {"Bob", "b@example.com"},
		{"Alice", "a@example.com"}, {"Alice", "a@example.com"},
	})
	if len(fwd) != len(rev) {
		t.Fatalf("size differs by insertion order: %d vs %d", len(fwd), len(rev))
	}
	for k, v := range fwd {
		if rev[k] != v {
			t.Fatalf("value for %q differs by insertion order: %q vs %q", k, v, rev[k])
		}
	}
}

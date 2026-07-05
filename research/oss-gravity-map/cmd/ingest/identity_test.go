package main

import (
	"context"
	"testing"
)

func TestResolveInlineNoreply(t *testing.T) {
	cases := []struct {
		name    string
		emails  []string
		wantID  int64
		wantLog string
		wantOK  bool
	}{
		{
			name:    "id+login noreply resolves to numeric id",
			emails:  []string{"1234567+octocat@users.noreply.github.com"},
			wantID:  1234567,
			wantLog: "octocat",
			wantOK:  true,
		},
		{
			name:   "login-only noreply carries no id (needs API)",
			emails: []string{"octocat@users.noreply.github.com"},
			wantOK: false,
		},
		{
			name:   "real email is not an inline id source",
			emails: []string{"alice@example.com"},
			wantOK: false,
		},
		{
			name:    "dominant email wins: first id-bearing in order is taken",
			emails:  []string{"alice@example.com", "42+alice@users.noreply.github.com"},
			wantID:  42,
			wantLog: "alice",
			wantOK:  true,
		},
		{
			name:   "a spoofed-looking subdomain is rejected",
			emails: []string{"42+x@users.noreply.github.com.evil.com"},
			wantOK: false,
		},
		{
			name:   "zero id is rejected",
			emails: []string{"0+x@users.noreply.github.com"},
			wantOK: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := resolveInlineNoreply(tc.emails)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if got.ID != tc.wantID || got.Login != tc.wantLog {
				t.Errorf("got {%d,%q}, want {%d,%q}", got.ID, got.Login, tc.wantID, tc.wantLog)
			}
		})
	}
}

func TestEmailsByCount_DeterministicDominance(t *testing.T) {
	byEmail := map[string]int{
		"b@x.com": 3,
		"a@x.com": 3, // tie with b → lexical: a before b
		"c@x.com": 5, // highest count → first
	}
	got := emailsByCount(byEmail)
	want := []string{"c@x.com", "a@x.com", "b@x.com"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

// resolve() must prefer the free, all-time inline-noreply id and never need the
// API when every author commits under an id-bearing noreply address.
func TestResolve_PrefersInlineNoreply_NoAPI(t *testing.T) {
	r := newResolver("", 0) // empty github token: any API attempt would no-op
	emails := authorEmails{
		"Alice": {"99+alice@users.noreply.github.com": 10},
		"Bob":   {"bob@personal.com": 4}, // unresolved (no token, no inline id)
	}
	got := r.resolve(context.Background(), "owner/repo", emails)
	if a, ok := got["Alice"]; !ok || a.ID != 99 || a.Login != "alice" {
		t.Errorf("Alice = %+v (ok=%v), want id=99 login=alice", got["Alice"], ok)
	}
	if _, ok := got["Bob"]; ok {
		t.Errorf("Bob must stay unresolved without a token/inline id, got %+v", got["Bob"])
	}
}

// TestResolveViaAPI_CarriesCommitsAPILogin locks the fix: an author resolved
// through the Commits-API email map (a normal, non-noreply email) must come back
// NAMED — both id and login — not id-only. Before the fix this path returned
// {ID} with a blank login, so the whole non-noreply long tail (many top OSS
// committers, e.g. antirez) stayed a scoreless Member-N even under dev reveal.
func TestResolveViaAPI_CarriesCommitsAPILogin(t *testing.T) {
	r := newResolver("", 0)
	repoMail := map[string]resolvedAuthor{
		"salvatore@example.com": {ID: 65632, Login: "antirez"},
	}
	got, ok := r.resolveViaAPI(context.Background(), []string{"salvatore@example.com"}, repoMail)
	if !ok || got.ID != 65632 || got.Login != "antirez" {
		t.Errorf("resolveViaAPI = %+v (ok=%v), want id=65632 login=antirez", got, ok)
	}
}

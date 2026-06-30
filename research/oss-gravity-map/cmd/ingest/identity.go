package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/git"
)

// noreplyWithID matches the GitHub per-user noreply address that embeds the
// account's NUMERIC id and login, e.g. "1234567+octocat@users.noreply.github.com".
// This is the cheapest, most reliable id source: it comes straight from the
// commit's author email in the full git history (no API call, no recency window),
// and the id is unforgeable by anyone but the account owner.
var noreplyWithID = regexp.MustCompile(`^([0-9]+)\+([A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?)@users\.noreply\.github\.com$`)

// noreplyLoginOnly matches the older login-only noreply form, e.g.
// "octocat@users.noreply.github.com" — carries the login but not the id, so it
// needs one /users/{login} lookup to resolve.
var noreplyLoginOnly = regexp.MustCompile(`^([A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?)@users\.noreply\.github\.com$`)

// resolvedAuthor is one canonical author's GitHub identity. ID == 0 means
// unresolved (the author never committed under a resolvable email within the
// pipeline's reach); such authors are dropped before the payload is built.
type resolvedAuthor struct {
	ID    int64
	Login string
}

// authorEmails holds, per canonical author NAME (matching the eis analysis JSON's
// `member`), the count of commits seen under each lowercased email. The dominant
// email — the one an author commits under most — is the identity signal we trust.
type authorEmails map[string]map[string]int

// collectAuthorEmails reproduces eis's exact author canonicalization so the names
// it keys by match analyze.go's `member` field one-for-one:
//
//	BuildIdentityMap + CanonicalizeAuthors  (collapse split names under one email)
//	cfg.ResolveAuthor                        (config aliases)
//	cfg.IsExcludedAuthor                     (drop bots/excluded, same as filterCommits)
//
// It returns name -> email -> commit-count. Emails stay in memory only.
func collectAuthorEmails(ctx context.Context, repoPath string, cfg *config.Config, workers int) (authorEmails, error) {
	commits, err := git.ParseLogParallel(ctx, repoPath, workers, cfg.CommentFilterEnabled())
	if err != nil {
		return nil, fmt.Errorf("parse log: %w", err)
	}
	// Same order as analyze.go: identity collapse THEN alias resolution.
	idmap := git.BuildIdentityMap(commits)
	git.CanonicalizeAuthors(commits, nil, idmap)

	out := authorEmails{}
	for _, c := range commits {
		// filterCommits applies ResolveAuthor before the exclusion test, so the
		// excluded-author check must see the resolved name too.
		name := cfg.ResolveAuthor(c.Author)
		if cfg.IsExcludedAuthor(c.Author) {
			continue
		}
		email := strings.ToLower(strings.TrimSpace(c.Email))
		if email == "" {
			continue
		}
		if out[name] == nil {
			out[name] = map[string]int{}
		}
		out[name][email]++
	}
	return out, nil
}

// emailsByCount returns an author's emails ordered by descending commit count,
// ties broken lexically — deterministic so the same history always resolves the
// same way (W-02).
func emailsByCount(byEmail map[string]int) []string {
	emails := make([]string, 0, len(byEmail))
	for e := range byEmail {
		emails = append(emails, e)
	}
	sort.Slice(emails, func(i, j int) bool {
		if byEmail[emails[i]] != byEmail[emails[j]] {
			return byEmail[emails[i]] > byEmail[emails[j]]
		}
		return emails[i] < emails[j]
	})
	return emails
}

// resolver turns canonical author names into GitHub numeric ids. It layers three
// email-keyed sources, cheapest first:
//
//	1. inline noreply id   — free, all-time, from the full git log
//	2. Commits API         — email -> id over the repo's recent commits (bounded)
//	3. /users/{login}      — for login-only noreply emails
//
// A display-name match is deliberately NOT used: it could attribute one
// engineer's gravity to another's account. Email is the stable join key.
type resolver struct {
	httpc       *http.Client
	githubToken string
	maxPages    int
	// caches, keyed to avoid re-hitting the API
	loginID  map[string]int64            // login -> id (/users)
	repoMail map[string]map[string]int64 // "owner/repo" -> email -> id (Commits API)
}

func newResolver(githubToken string, maxPages int) *resolver {
	return &resolver{
		httpc:       &http.Client{Timeout: 30 * time.Second},
		githubToken: githubToken,
		maxPages:    maxPages,
		loginID:     map[string]int64{},
		repoMail:    map[string]map[string]int64{},
	}
}

// resolve maps each author name in `emails` to a GitHub identity. Unresolved
// authors are simply absent from the result (callers drop them).
func (r *resolver) resolve(ctx context.Context, fullName string, emails authorEmails) map[string]resolvedAuthor {
	out := make(map[string]resolvedAuthor, len(emails))
	var needAPI bool
	for name, byEmail := range emails {
		ordered := emailsByCount(byEmail)
		if ra, ok := resolveInlineNoreply(ordered); ok {
			out[name] = ra
			continue
		}
		needAPI = true
	}
	if !needAPI || r.githubToken == "" {
		return out
	}
	// One bounded Commits API sweep per repo, shared across the unresolved authors.
	mail := r.repoEmailIDs(ctx, fullName)
	for name, byEmail := range emails {
		if _, done := out[name]; done {
			continue
		}
		ra, ok := r.resolveViaAPI(ctx, emailsByCount(byEmail), mail)
		if ok {
			out[name] = ra
		}
	}
	return out
}

// resolveInlineNoreply pulls an id straight out of an "<id>+<login>@users.noreply"
// email. Tries the author's emails in dominance order and takes the first that
// carries an id.
func resolveInlineNoreply(orderedEmails []string) (resolvedAuthor, bool) {
	for _, e := range orderedEmails {
		if m := noreplyWithID.FindStringSubmatch(e); m != nil {
			id, err := strconv.ParseInt(m[1], 10, 64)
			if err == nil && id > 0 {
				return resolvedAuthor{ID: id, Login: m[2]}, true
			}
		}
	}
	return resolvedAuthor{}, false
}

// resolveViaAPI resolves through the Commits-API email map first, then falls back
// to a /users/{login} lookup for login-only noreply emails.
func (r *resolver) resolveViaAPI(ctx context.Context, orderedEmails []string, repoMail map[string]int64) (resolvedAuthor, bool) {
	for _, e := range orderedEmails {
		if id := repoMail[e]; id > 0 {
			return resolvedAuthor{ID: id}, true
		}
	}
	for _, e := range orderedEmails {
		if m := noreplyLoginOnly.FindStringSubmatch(e); m != nil {
			if id := r.loginToID(ctx, m[1]); id > 0 {
				return resolvedAuthor{ID: id, Login: m[1]}, true
			}
		}
	}
	return resolvedAuthor{}, false
}

// repoEmailIDs builds (and caches) email -> github id from the repo's recent
// commits via the GitHub Commits API. Bounded to maxPages * 100 commits — this is
// a best-effort enrichment for authors whose dominant email isn't a noreply, not
// an exhaustive sweep. Cached per repo for the run.
func (r *resolver) repoEmailIDs(ctx context.Context, fullName string) map[string]int64 {
	if m, ok := r.repoMail[fullName]; ok {
		return m
	}
	m := map[string]int64{}
	r.repoMail[fullName] = m
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 {
		return m
	}
	type commitEnvelope struct {
		Commit struct {
			Author struct {
				Email string `json:"email"`
			} `json:"author"`
		} `json:"commit"`
		Author *struct {
			ID int64 `json:"id"`
		} `json:"author"`
	}
	for page := 1; page <= r.maxPages; page++ {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?per_page=100&page=%d", parts[0], parts[1], page)
		body, status, err := r.get(ctx, url)
		if err != nil || status == http.StatusNotFound || status == http.StatusForbidden {
			break
		}
		if status >= 400 {
			break
		}
		var batch []commitEnvelope
		if json.Unmarshal(body, &batch) != nil || len(batch) == 0 {
			break
		}
		for _, e := range batch {
			email := strings.ToLower(strings.TrimSpace(e.Commit.Author.Email))
			if email == "" || e.Author == nil || e.Author.ID == 0 {
				continue
			}
			if _, seen := m[email]; !seen {
				m[email] = e.Author.ID
			}
		}
		if len(batch) < 100 {
			break
		}
	}
	return m
}

// loginToID resolves a login to its numeric id via /users/{login}, cached.
func (r *resolver) loginToID(ctx context.Context, login string) int64 {
	if id, ok := r.loginID[login]; ok {
		return id
	}
	r.loginID[login] = 0 // negative-cache by default
	body, status, err := r.get(ctx, "https://api.github.com/users/"+login)
	if err != nil || status != http.StatusOK {
		return 0
	}
	var u struct {
		ID int64 `json:"id"`
	}
	if json.Unmarshal(body, &u) == nil && u.ID > 0 {
		r.loginID[login] = u.ID
	}
	return r.loginID[login]
}

func (r *resolver) get(ctx context.Context, url string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+r.githubToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := r.httpc.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

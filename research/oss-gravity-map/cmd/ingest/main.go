// Command ingest publishes an EIS research run to the OrbitLens research store.
//
// It reads the per-repo eis analysis JSON produced by scripts/analyze-repos.sh,
// resolves each author's git identity to a GitHub numeric user id (the join key
// the research map + Claim flow rely on), assembles one research run, and POSTs
// it to the privileged ingest endpoint:
//
//	POST {api-base}/oss/research/ingest   Authorization: Bearer $RESEARCH_INGEST_TOKEN
//
// This is an OBSERVATION-class step (W-01): it reads git history and the already
// computed eis signals and writes them through; it never recomputes or invents.
// Author emails are resolved to ids in memory and never written to disk or the
// payload — only the numeric id leaves the process.
//
// Usage:
//
//	ingest --run-id <id> [--api-base https://api.stg.orbitlens.io] [--dry-run]
//
// Env: RESEARCH_INGEST_TOKEN (required unless --dry-run), GITHUB_TOKEN (optional,
// enables the Commits-API identity fallback for non-noreply emails).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/machuz/eis/v2/internal/config"
)

// --- ingest payload (mirrors services/public-api usecase.IngestPayload) ---

type ingestObservation struct {
	ExGitHubID int64 `json:"ex_github_id"`
	// Period "" / "cumulative" = the HEAD standing; "YYYY" = a timeline window
	// (oss-public-claim.md S4). LifetimeGravity is carried on the cumulative row.
	Period           string  `json:"period,omitempty"`
	Gravity          float64 `json:"gravity"`
	LifetimeGravity  float64 `json:"lifetime_gravity,omitempty"`
	Survival         float64 `json:"survival"`
	Production       float64 `json:"production"`
	Catalysis        float64 `json:"catalysis"`
	Design           float64 `json:"design"`
	Breadth          float64 `json:"breadth"`
	Indispensability float64 `json:"indispensability"`
	DebtCleanup      float64 `json:"debt_cleanup"`
	RoleArchetype    string  `json:"role_archetype"`
	StyleArchetype   string  `json:"style_archetype"`
	StateArchetype   string  `json:"state_archetype"`
}

type ingestRepo struct {
	FullName       string              `json:"full_name"`
	LanguageFamily string              `json:"language_family"`
	Observations   []ingestObservation `json:"observations"`
}

type ingestPayload struct {
	RunID      string       `json:"run_id"`
	ObservedAt time.Time    `json:"observed_at"`
	Repos      []ingestRepo `json:"repos"`
}

// --- inputs ---

type manifestEntry struct {
	Dir            string `yaml:"dir"`
	FullName       string `yaml:"full_name"`
	LanguageFamily string `yaml:"language_family"`
}

type repoManifest struct {
	Repos []manifestEntry `yaml:"repos"`
}

// analysisMember is the subset of an eis analysis member entry the ingest needs.
type analysisMember struct {
	Member           string  `json:"member"`
	Production       float64 `json:"production"`
	Catalysis        float64 `json:"catalysis"`
	Survival         float64 `json:"survival"`
	Design           float64 `json:"design"`
	Breadth          float64 `json:"breadth"`
	DebtCleanup      float64 `json:"debt_cleanup"`
	Indispensability float64 `json:"indispensability"`
	Gravity          float64 `json:"gravity"`
	LifetimeGravity  float64 `json:"lifetime_gravity"`
	Role             string  `json:"role"`
	Style            string  `json:"style"`
	State            string  `json:"state"`
}

type analysisJSON struct {
	Domains []struct {
		Members []analysisMember `json:"members"`
	} `json:"domains"`
}

// timelineJSON is the subset of `eis timeline --format json` the ingest needs:
// per author, their signals in each period window. The author name matches the
// analysis `member` (eis canonicalizes identically), so both join on the same id.
type timelinePeriod struct {
	Label            string  `json:"label"`
	Production       float64 `json:"production"`
	Catalysis        float64 `json:"catalysis"`
	Survival         float64 `json:"survival"`
	Design           float64 `json:"design"`
	Breadth          float64 `json:"breadth"`
	DebtCleanup      float64 `json:"debt_cleanup"`
	Indispensability float64 `json:"indispensability"`
	Gravity          float64 `json:"gravity"`
	Role             string  `json:"role"`
	Style            string  `json:"style"`
	State            string  `json:"state"`
}

type timelineJSON struct {
	Authors []struct {
		Author  string           `json:"author"`
		Periods []timelinePeriod `json:"periods"`
	} `json:"authors"`
}

func main() {
	var (
		manifestPath = flag.String("manifest", "ingest-repos.yaml", "repo manifest (dir/full_name/language_family)")
		resultsDir   = flag.String("results-dir", "data/results", "dir of eis analysis JSON (<dir>.json)")
		timelineDir  = flag.String("timeline-dir", "data/results", "dir of eis timeline JSON (<dir>-timeline.json); missing = cumulative only")
		reposDir     = flag.String("repos-dir", "data/repos", "dir of cloned repos (<dir>/.git)")
		configsDir   = flag.String("configs-dir", "configs", "dir of per-repo eis configs (<dir>.yaml)")
		apiBase      = flag.String("api-base", "https://api.stg.orbitlens.io", "ingest API base URL")
		runID        = flag.String("run-id", "", "research run id (lineage); required")
		maxPages     = flag.Int("max-commit-pages", 10, "GitHub Commits API pages for the email→id fallback")
		only         = flag.String("only", "", "comma-separated manifest dirs to ingest (default: all). Lets a CI loop clone+analyze+ingest one repo at a time, keeping disk bounded — re-POSTs share the run-id and upsert idempotently.")
		dryRun       = flag.Bool("dry-run", false, "resolve and assemble but do not POST; print a summary")
	)
	flag.Parse()

	if *runID == "" {
		fatalf("--run-id is required (pass a deterministic id, e.g. the CI run id or a UTC date)")
	}
	token := os.Getenv("RESEARCH_INGEST_TOKEN")
	if token == "" && !*dryRun {
		fatalf("RESEARCH_INGEST_TOKEN is required (set --dry-run to skip the POST)")
	}

	manifest, err := loadManifest(*manifestPath)
	if err != nil {
		fatalf("manifest: %v", err)
	}
	if *only != "" {
		manifest.filter(splitCSV(*only))
		if len(manifest.Repos) == 0 {
			fatalf("--only %q matched no manifest dirs", *only)
		}
	}

	ctx := context.Background()
	res := newResolver(os.Getenv("GITHUB_TOKEN"), *maxPages)
	workers := runtime.NumCPU()

	payload := ingestPayload{RunID: *runID, ObservedAt: time.Now().UTC()}
	var totalObs, resolved, dropped int

	for _, m := range manifest.Repos {
		resultPath := filepath.Join(*resultsDir, m.Dir+".json")
		members, err := loadAnalysis(resultPath)
		if err != nil {
			warnf("%s: skipping (analysis: %v)", m.Dir, err)
			continue
		}
		repoPath := filepath.Join(*reposDir, m.Dir)
		if !isGitRepo(repoPath) {
			warnf("%s: skipping (no clone at %s)", m.Dir, repoPath)
			continue
		}

		cfg := loadRepoConfig(*configsDir, m.Dir)
		emails, err := collectAuthorEmails(ctx, repoPath, cfg, workers)
		if err != nil {
			warnf("%s: skipping (identity: %v)", m.Dir, err)
			continue
		}
		ids := res.resolve(ctx, m.FullName, emails)

		obs := make([]ingestObservation, 0, len(members))
		for _, mem := range members {
			ra, ok := ids[mem.Member]
			if !ok || ra.ID == 0 {
				dropped++
				continue
			}
			// The cumulative (HEAD) standing carries lifetime_gravity.
			obs = append(obs, ingestObservation{
				ExGitHubID:       ra.ID,
				Period:           "cumulative",
				Gravity:          mem.Gravity,
				LifetimeGravity:  mem.LifetimeGravity,
				Survival:         mem.Survival,
				Production:       mem.Production,
				Catalysis:        mem.Catalysis,
				Design:           mem.Design,
				Breadth:          mem.Breadth,
				Indispensability: mem.Indispensability,
				DebtCleanup:      mem.DebtCleanup,
				RoleArchetype:    mem.Role,
				StyleArchetype:   mem.Style,
				StateArchetype:   mem.State,
			})
		}
		resolved += len(obs)
		totalObs += len(members)

		// Timeline (軌跡): per-period rows, joined to the same ids. Optional — a
		// missing/failed timeline just leaves the cumulative standing.
		tl := loadTimeline(filepath.Join(*timelineDir, m.Dir+"-timeline.json"))
		var periodRows int
		for name, periods := range tl {
			ra, ok := ids[name]
			if !ok || ra.ID == 0 {
				continue
			}
			for _, p := range periods {
				if p.Label == "" {
					continue
				}
				// Skip fully-inactive periods (no gravity, no archetype): they carry
				// no signal and would bloat the store with one 0-row per author per
				// idle window. Any gravity OR any archetype keeps the row.
				if p.Gravity == 0 && isDash(p.Role) && isDash(p.Style) && isDash(p.State) {
					continue
				}
				obs = append(obs, ingestObservation{
					ExGitHubID: ra.ID, Period: p.Label,
					Gravity: p.Gravity, Survival: p.Survival, Production: p.Production,
					Catalysis: p.Catalysis, Design: p.Design, Breadth: p.Breadth,
					Indispensability: p.Indispensability, DebtCleanup: p.DebtCleanup,
					RoleArchetype: p.Role, StyleArchetype: p.Style, StateArchetype: p.State,
				})
				periodRows++
			}
		}
		// Deterministic order so a re-run produces an identical payload (W-02).
		sort.Slice(obs, func(i, j int) bool { return obs[i].ExGitHubID < obs[j].ExGitHubID })
		payload.Repos = append(payload.Repos, ingestRepo{
			FullName:       m.FullName,
			LanguageFamily: m.LanguageFamily,
			Observations:   obs,
		})
		fmt.Fprintf(os.Stderr, "  %-28s %3d/%-3d authors resolved, %d timeline rows\n", m.FullName, len(obs)-periodRows, len(members), periodRows)
	}

	fmt.Fprintf(os.Stderr, "\nrun %s: %d repos, %d observations resolved (%d dropped, unresolved identity)\n",
		payload.RunID, len(payload.Repos), resolved, dropped)

	if len(payload.Repos) == 0 {
		fatalf("no repos resolved — nothing to ingest (did analyze + clone run?)")
	}

	if *dryRun {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(payload)
		return
	}

	if err := post(ctx, *apiBase, token, payload); err != nil {
		fatalf("ingest POST: %v", err)
	}
	fmt.Fprintf(os.Stderr, "ingested run %s to %s\n", payload.RunID, *apiBase)
}

func loadManifest(path string) (*repoManifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m repoManifest
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if len(m.Repos) == 0 {
		return nil, fmt.Errorf("manifest has no repos")
	}
	return &m, nil
}

// filter keeps only the manifest entries whose dir is in `dirs`.
func (m *repoManifest) filter(dirs []string) {
	want := make(map[string]bool, len(dirs))
	for _, d := range dirs {
		want[d] = true
	}
	kept := m.Repos[:0]
	for _, r := range m.Repos {
		if want[r.Dir] {
			kept = append(kept, r)
		}
	}
	m.Repos = kept
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// loadAnalysis reads one eis analysis JSON and flattens its domains' members.
// An author can surface in more than one domain of a repo; we keep the entry
// with the highest gravity (their dominant contribution) so each (repo, author)
// yields exactly one observation — deterministically.
func loadAnalysis(path string) ([]analysisMember, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var a analysisJSON
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	best := map[string]analysisMember{}
	for _, d := range a.Domains {
		for _, mem := range d.Members {
			if mem.Member == "" {
				continue
			}
			if cur, ok := best[mem.Member]; !ok || mem.Gravity > cur.Gravity {
				best[mem.Member] = mem
			}
		}
	}
	if len(best) == 0 {
		return nil, fmt.Errorf("no members")
	}
	out := make([]analysisMember, 0, len(best))
	for _, mem := range best {
		out = append(out, mem)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Member < out[j].Member })
	return out, nil
}

// isDash reports whether an archetype value is absent (empty or the em-dash eis
// writes when a period has no settled archetype).
func isDash(s string) bool { return s == "" || s == "—" }

// loadTimeline reads an eis timeline JSON (軌跡) and returns author name → their
// per-period signals. Missing/unreadable file → nil (timeline is optional; the
// cumulative standing still ingests). Author names match the analysis `member`.
func loadTimeline(path string) map[string][]timelinePeriod {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	// `eis timeline --format json` emits TWO concatenated JSON objects: the
	// per-author timeline, then a {team_timelines:…} object. Decode reads only the
	// first (the per-author one, which is what we need) and ignores the rest.
	var t timelineJSON
	if err := json.NewDecoder(f).Decode(&t); err != nil {
		warnf("timeline %s: %v", path, err)
		return nil
	}
	out := make(map[string][]timelinePeriod, len(t.Authors))
	for _, a := range t.Authors {
		if a.Author == "" || len(a.Periods) == 0 {
			continue
		}
		out[a.Author] = a.Periods
	}
	return out
}

// loadRepoConfig mirrors analyze-repos.sh: use configs/<dir>.yaml when present,
// else eis's default config. Matching the analysis config keeps author
// canonicalization identical to the run that produced the signals.
func loadRepoConfig(configsDir, dir string) *config.Config {
	path := filepath.Join(configsDir, dir+".yaml")
	if _, err := os.Stat(path); err == nil {
		if cfg, err := config.Load(path, true); err == nil {
			return cfg
		}
	}
	return config.Default()
}

func isGitRepo(path string) bool {
	fi, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && fi.IsDir()
}

func post(ctx context.Context, apiBase, token string, payload ingestPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := apiBase + "/oss/research/ingest"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(rb))
	}
	fmt.Fprintf(os.Stderr, "ingest result: %s\n", string(rb))
	return nil
}

func warnf(format string, a ...any) { fmt.Fprintf(os.Stderr, "WARN "+format+"\n", a...) }

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "ERROR "+format+"\n", a...)
	os.Exit(1)
}

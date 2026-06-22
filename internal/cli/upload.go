package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/machuz/eis/v2/internal/git"
	"github.com/machuz/eis/v2/internal/scorer"
)

// defaultUploadURL is the observatory the CLI uploads signals to. The URL is
// fixed (users never type it); ACE_API_URL overrides it for staging only. This
// is the dedicated per-service public API host (ALB-direct), named as a sibling
// of the frontend (ace-api.orbitlens.io ↔ ace.orbitlens.io) rather than a
// generic api. host — true/ideal get true-api./ideal-api. The contract stays
// off the frontend's ace.orbitlens.io/api/* path and stable across FE changes.
const defaultUploadURL = "https://ace-api.orbitlens.io"

// uploadMemberSignal mirrors the SaaS UploadMemberSignal contract
// (services/ace/backend/app/usecase/signal_upload.go). The json tags are the
// wire contract — keep them byte-for-byte in sync with the server.
type uploadMemberSignal struct {
	Author             string  `json:"author"`
	Production         float64 `json:"production"`
	Catalysis          float64 `json:"catalysis"`
	Survival           float64 `json:"survival"`
	RawSurvival        float64 `json:"raw_survival"`
	RobustSurvival     float64 `json:"robust_survival"`
	DormantSurvival    float64 `json:"dormant_survival"`
	RawRobustSurvival  float64 `json:"raw_robust_survival"`
	RawDormantSurvival float64 `json:"raw_dormant_survival"`
	Design             float64 `json:"design"`
	Breadth            float64 `json:"breadth"`
	DebtCleanup        float64 `json:"debt_cleanup"`
	Indispensability   float64 `json:"indispensability"`
	Gravity            float64 `json:"gravity"`
	Total              float64 `json:"total"`
	TotalCommits       int     `json:"total_commits"`
	LinesAdded         int     `json:"lines_added"`
	LinesDeleted       int     `json:"lines_deleted"`
	RecentlyActive     bool    `json:"recently_active"`
	RoleArchetype      string  `json:"role_archetype"`
	StyleArchetype     string  `json:"style_archetype"`
	StateArchetype     string  `json:"state_archetype"`
	RoleConfidence     float64 `json:"role_confidence"`
	StyleConfidence    float64 `json:"style_confidence"`
	StateConfidence    float64 `json:"state_confidence"`
}

// uploadRequest is the POST body for /v1/signals/upload. The git_sha /
// sha_author_date / analysis_time fields carry the observation lineage the
// observatory needs (and cannot derive, since it never sees the repo): they key
// observations_member_signals as (galaxy, git_sha, kind, user) at sha_author_date.
type uploadRequest struct {
	GitSHA        string               `json:"git_sha"`
	SHAAuthorDate time.Time            `json:"sha_author_date"`
	AnalysisTime  time.Time            `json:"analysis_time"`
	Domain        string               `json:"domain"`
	Signals       []uploadMemberSignal `json:"signals"`
}

// uploadResults sends the analysis to the observatory. One request per domain,
// each stamped with the analyzed repo's HEAD sha + author date. This is the
// "code never leaves your machine — only light reaches the observatory" path
// for GitHub Enterprise / no-clone customers: the SaaS stores these as
// upload-sourced observations (not git-reproducible server-side; credibility is
// gauged by the Trust Signal).
func uploadResults(ctx context.Context, domainResults []DomainResults, repoPath, token string, analysisTime time.Time) error {
	if token == "" {
		return fmt.Errorf("upload requires a token: pass --token or set EIS_TOKEN (create one in Settings → API tokens)")
	}
	baseURL := os.Getenv("ACE_API_URL")
	if baseURL == "" {
		baseURL = defaultUploadURL
	}

	sha, err := git.HeadHash(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("resolve HEAD sha (%s): %w", repoPath, err)
	}
	authorDate, err := git.HeadAuthorDate(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("resolve HEAD author date (%s): %w", repoPath, err)
	}
	// analysisTime is the W-02 envelope clock the scores were actually computed
	// against (threaded in from RunAnalyzePipeline), NOT a fresh time.Now():
	// the recorded analysis_time MUST equal the decay/classification reference,
	// or the observation drifts run-to-run for the same git_sha.
	if analysisTime.IsZero() {
		analysisTime = time.Now().UTC()
	}

	for _, dr := range domainResults {
		req := uploadRequest{
			GitSHA:        sha,
			SHAAuthorDate: authorDate,
			AnalysisTime:  analysisTime,
			Domain:        string(dr.Domain),
			Signals:       mapMemberSignals(dr.Results),
		}
		if err := postUpload(ctx, baseURL, token, req); err != nil {
			return fmt.Errorf("upload domain %q: %w", dr.Domain, err)
		}
	}
	return nil
}

func mapMemberSignals(results []scorer.Result) []uploadMemberSignal {
	out := make([]uploadMemberSignal, 0, len(results))
	for _, r := range results {
		out = append(out, uploadMemberSignal{
			Author:             r.Author,
			Production:         r.Production,
			Catalysis:          r.Catalysis,
			Survival:           r.Survival,
			RawSurvival:        r.RawSurvival,
			RobustSurvival:     r.RobustSurvival,
			DormantSurvival:    r.DormantSurvival,
			RawRobustSurvival:  r.RawRobustSurv,
			RawDormantSurvival: r.RawDormantSurv,
			Design:             r.Design,
			Breadth:            r.Breadth,
			DebtCleanup:        r.DebtCleanup,
			Indispensability:   r.Indispensability,
			Gravity:            r.Gravity,
			Total:              r.Impact,
			TotalCommits:       r.TotalCommits,
			LinesAdded:         r.LinesAdded,
			LinesDeleted:       r.LinesDeleted,
			RecentlyActive:     r.RecentlyActive,
			RoleArchetype:      r.Role,
			StyleArchetype:     r.Style,
			StateArchetype:     r.State,
			RoleConfidence:     r.RoleConf,
			StyleConfidence:    r.StyleConf,
			StateConfidence:    r.StateConf,
		})
	}
	return out
}

func postUpload(ctx context.Context, baseURL, token string, req uploadRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	url := baseURL + "/v1/signals/upload"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("post to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("observatory rejected upload (%d): %s", resp.StatusCode, bytes.TrimSpace(snippet))
	}
	return nil
}

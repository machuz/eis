package output

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/machuz/engineering-impact-score/internal/metric"
	"github.com/machuz/engineering-impact-score/internal/scorer"
)

type jsonOutput struct {
	Domains []jsonDomain `json:"domains"`
}

type jsonDomain struct {
	Name      string           `json:"name"`
	Repos     int              `json:"repos"`
	Members   []jsonMember     `json:"members"`
	BusFactor []jsonBusFactor  `json:"bus_factor,omitempty"`
	PerRepo   []jsonRepoResult `json:"per_repo,omitempty"`
}

type jsonRepoResult struct {
	RepoName string       `json:"repo_name"`
	Members  []jsonMember `json:"members"`
}

type jsonMember struct {
	Rank             int     `json:"rank"`
	Member           string  `json:"member"`
	Active           bool    `json:"active"`
	Commits          int     `json:"commits"`
	Production       float64 `json:"production"`
	Quality          float64 `json:"quality"`
	Survival         float64 `json:"survival"`
	RobustSurvival   float64 `json:"robust_survival"`
	DormantSurvival  float64 `json:"dormant_survival"`
	Design           float64 `json:"design"`
	Breadth          float64 `json:"breadth"`
	DebtCleanup      float64 `json:"debt_cleanup"`
	Indispensability float64 `json:"indispensability"`
	Gravity          float64 `json:"gravity"`
	Total            float64 `json:"total"`
	Role             string  `json:"role"`
	RoleConf         float64 `json:"role_confidence"`
	Style            string  `json:"style"`
	StyleConf        float64 `json:"style_confidence"`
	State            string  `json:"state"`
	StateConf        float64 `json:"state_confidence"`
}

type jsonBusFactor struct {
	Repo   string  `json:"repo"`
	Level  string  `json:"level"`
	Module string  `json:"module"`
	Owner  string  `json:"owner"`
	Share  float64 `json:"share"`
}

// JSONWriter accumulates domain data for a single JSON output at the end.
type JSONWriter struct {
	output jsonOutput
}

func NewJSONWriter() *JSONWriter {
	return &JSONWriter{}
}

func (w *JSONWriter) AddDomain(domainName string, repoCount int, results []scorer.Result, risks []metric.ModuleRisk) {
	d := jsonDomain{
		Name:  domainName,
		Repos: repoCount,
	}

	for i, r := range results {
		m := jsonMember{
			Rank:             i + 1,
			Member:           r.Author,
			Active:           r.RecentlyActive,
			Commits:          r.TotalCommits,
			Production:       round1(r.Production),
			Quality:          round1(r.Quality),
			Survival:         round1(r.Survival),
			RobustSurvival:   round1(r.RobustSurvival),
			DormantSurvival:  round1(r.DormantSurvival),
			Design:           round1(r.Design),
			Breadth:          round1(r.Breadth),
			DebtCleanup:      round1(r.DebtCleanup),
			Indispensability: round1(r.Indispensability),
			Gravity:          round1(r.Gravity),
			Total:            round1(r.Total),
			Role:             r.Role,
			RoleConf:         r.RoleConf,
			Style:            r.Style,
			StyleConf:        r.StyleConf,
			State:            r.State,
			StateConf:        r.StateConf,
		}
		d.Members = append(d.Members, m)
	}

	for _, r := range risks {
		d.BusFactor = append(d.BusFactor, jsonBusFactor{
			Level:  r.Level,
			Module: r.Module,
			Owner:  r.TopAuthor,
			Share:  round1(r.Share * 100),
		})
	}

	w.output.Domains = append(w.output.Domains, d)
}

// AddPerRepo appends per-repo results to the last added domain (or matching domain).
func (w *JSONWriter) AddPerRepo(domainName, repoName string, results []scorer.Result) {
	// Find matching domain
	for i := len(w.output.Domains) - 1; i >= 0; i-- {
		if w.output.Domains[i].Name == domainName {
			rr := jsonRepoResult{RepoName: repoName}
			for j, r := range results {
				rr.Members = append(rr.Members, jsonMember{
					Rank:             j + 1,
					Member:           r.Author,
					Active:           r.RecentlyActive,
					Commits:          r.TotalCommits,
					Production:       round1(r.Production),
					Quality:          round1(r.Quality),
					Survival:         round1(r.Survival),
					RobustSurvival:   round1(r.RobustSurvival),
					DormantSurvival:  round1(r.DormantSurvival),
					Design:           round1(r.Design),
					Breadth:          round1(r.Breadth),
					DebtCleanup:      round1(r.DebtCleanup),
					Indispensability: round1(r.Indispensability),
					Gravity:          round1(r.Gravity),
					Total:            round1(r.Total),
					Role:             r.Role,
					RoleConf:         r.RoleConf,
					Style:            r.Style,
					StyleConf:        r.StyleConf,
					State:            r.State,
					StateConf:        r.StateConf,
				})
			}
			w.output.Domains[i].PerRepo = append(w.output.Domains[i].PerRepo, rr)
			return
		}
	}
}

func (w *JSONWriter) Flush() error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(w.output)
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// PrintRankingsJSON is a convenience for single-domain output (not used in multi-domain flow).
func PrintRankingsJSON(domain string, repoCount int, results []scorer.Result, risks []metric.ModuleRisk) {
	w := NewJSONWriter()
	w.AddDomain(domain, repoCount, results, risks)
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "json encode error: %v\n", err)
	}
}

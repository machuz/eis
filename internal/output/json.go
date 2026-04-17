package output

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/machuz/eis/internal/metric"
	"github.com/machuz/eis/internal/scorer"
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

	// Test coverage footprint aggregated across repos in this domain.
	TotalFiles     int     `json:"total_files,omitempty"`
	TotalTestFiles int     `json:"total_test_files,omitempty"`
	TestFileRatio  float64 `json:"test_file_ratio,omitempty"`

	// Module Science Phase 1
	Cochange  []jsonCochangeRepo   `json:"cochange,omitempty"`
	Ownership []jsonModuleOwnership `json:"ownership,omitempty"`

	// Module Science Phase 2
	ModuleScores []jsonModuleScore `json:"module_scores,omitempty"`
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
	LinesAdded       int     `json:"lines_added"`
	LinesDeleted     int     `json:"lines_deleted"`
	Production       float64 `json:"production"`
	Quality          float64 `json:"quality"`
	Survival         float64 `json:"survival"`
	RobustSurvival   float64 `json:"robust_survival"`
	DormantSurvival  float64 `json:"dormant_survival"`
	TestedSurvival   float64 `json:"tested_survival"`
	UntestedSurvival float64 `json:"untested_survival"`
	Design           float64 `json:"design"`
	Breadth          float64 `json:"breadth"`
	DebtCleanup      float64 `json:"debt_cleanup"`
	Indispensability float64 `json:"indispensability"`
	Gravity          float64 `json:"gravity"`
	Impact           float64 `json:"impact"`
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

type jsonCochangeRepo struct {
	Pairs []jsonCochangePair `json:"pairs"`
}

type jsonCochangePair struct {
	ModuleA       string  `json:"module_a"`
	ModuleB       string  `json:"module_b"`
	CochangeCount int     `json:"cochange_count"`
	Coupling      float64 `json:"coupling"`
}

type jsonModuleScore struct {
	Module                string  `json:"module"`
	BoundaryIntegrity     float64 `json:"boundary_integrity"`
	ChangeAbsorption      float64 `json:"change_absorption"`
	KnowledgeDistribution float64 `json:"knowledge_distribution"`
	Stability             float64 `json:"stability"`
	ChangePressure        float64 `json:"change_pressure"`
	AvgCoupling           float64 `json:"avg_coupling"`
	MaxCoupling           float64 `json:"max_coupling"`
	CouplingPairCount     int     `json:"coupling_pair_count"`
	AuthorCount           int     `json:"author_count"`
	BlameLines            int     `json:"blame_lines"`
	ModuleCommits         int     `json:"module_commits"`
	OwnershipLevel        string  `json:"ownership_level"`
	TopAuthorShare        float64 `json:"top_author_share"`
	OwnerActive           bool    `json:"owner_active"`
	Coupling              string  `json:"coupling"`
	CouplingConf          float64 `json:"coupling_confidence"`
	Vitality              string  `json:"vitality"`
	VitalityConf          float64 `json:"vitality_confidence"`
	Ownership             string  `json:"ownership"`
	OwnershipConf         float64 `json:"ownership_confidence"`
}

type jsonModuleOwnership struct {
	Module      string  `json:"module"`
	TotalLines  int     `json:"total_lines"`
	AuthorCount int     `json:"author_count"`
	TopAuthor   string  `json:"top_author"`
	TopShare    float64 `json:"top_share"`
	Entropy     float64 `json:"entropy"`
	Level       string  `json:"level"`
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
			LinesAdded:       r.LinesAdded,
			LinesDeleted:     r.LinesDeleted,
			Production:       round1(r.Production),
			Quality:          round1(r.Quality),
			Survival:         round1(r.Survival),
			RobustSurvival:   round1(r.RobustSurvival),
			DormantSurvival:  round1(r.DormantSurvival),
			TestedSurvival:   round1(r.TestedSurvival),
			UntestedSurvival: round1(r.UntestedSurvival),
			Design:           round1(r.Design),
			Breadth:          round1(r.Breadth),
			DebtCleanup:      round1(r.DebtCleanup),
			Indispensability: round1(r.Indispensability),
			Gravity:          round1(r.Gravity),
			Impact:           round1(r.Impact),
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

// AddTestCoverage annotates the most recently added domain with test-file
// footprint so SaaS can reason about test culture at a glance.
func (w *JSONWriter) AddTestCoverage(domainName string, totalFiles, totalTestFiles int, ratio float64) {
	for i := len(w.output.Domains) - 1; i >= 0; i-- {
		if w.output.Domains[i].Name == domainName {
			w.output.Domains[i].TotalFiles = totalFiles
			w.output.Domains[i].TotalTestFiles = totalTestFiles
			w.output.Domains[i].TestFileRatio = round2(ratio)
			return
		}
	}
}

// AddModuleScience appends co-change and ownership data to the last added domain.
func (w *JSONWriter) AddModuleScience(domainName string, cochangeResults []metric.CochangeResult, ownership []metric.ModuleOwnership) {
	for i := len(w.output.Domains) - 1; i >= 0; i-- {
		if w.output.Domains[i].Name == domainName {
			// Co-change: per-repo results, only significant pairs
			for _, cr := range cochangeResults {
				repo := jsonCochangeRepo{}
				for _, p := range cr.Pairs {
					if p.CochangeCount >= 5 && p.Coupling >= 0.10 {
						repo.Pairs = append(repo.Pairs, jsonCochangePair{
							ModuleA:       p.ModuleA,
							ModuleB:       p.ModuleB,
							CochangeCount: p.CochangeCount,
							Coupling:      round2(p.Coupling),
						})
					}
				}
				if len(repo.Pairs) > 0 {
					w.output.Domains[i].Cochange = append(w.output.Domains[i].Cochange, repo)
				}
			}

			// Ownership: all modules
			for _, o := range ownership {
				w.output.Domains[i].Ownership = append(w.output.Domains[i].Ownership, jsonModuleOwnership{
					Module:      o.Module,
					TotalLines:  o.TotalLines,
					AuthorCount: o.AuthorCount,
					TopAuthor:   o.TopAuthor,
					TopShare:    round2(o.TopShare),
					Entropy:     round2(o.Entropy),
					Level:       o.Level,
				})
			}
			return
		}
	}
}

// AddModuleScores appends module topology scores to the matching domain.
func (w *JSONWriter) AddModuleScores(domainName string, modules []scorer.ModuleScore) {
	for i := len(w.output.Domains) - 1; i >= 0; i-- {
		if w.output.Domains[i].Name == domainName {
			for _, ms := range modules {
				w.output.Domains[i].ModuleScores = append(w.output.Domains[i].ModuleScores, jsonModuleScore{
					Module:                ms.Module,
					BoundaryIntegrity:     round1(ms.BoundaryIntegrity),
					ChangeAbsorption:      round1(ms.ChangeAbsorption),
					KnowledgeDistribution: round1(ms.KnowledgeDistribution),
					Stability:             round1(ms.Stability),
					ChangePressure:        round2(ms.ChangePressure),
					AvgCoupling:           round2(ms.AvgCoupling),
					MaxCoupling:           round2(ms.MaxCoupling),
					CouplingPairCount:     ms.CouplingPairCount,
					AuthorCount:           ms.AuthorCount,
					BlameLines:            ms.BlameLines,
					ModuleCommits:         ms.ModuleCommits,
					OwnershipLevel:        ms.OwnershipLevel,
					TopAuthorShare:        round2(ms.TopAuthorShare),
					OwnerActive:           ms.OwnerActive,
					Coupling:              ms.Coupling,
					CouplingConf:          round2(ms.CouplingConf),
					Vitality:              ms.Vitality,
					VitalityConf:          round2(ms.VitalityConf),
					Ownership:             ms.Ownership,
					OwnershipConf:         round2(ms.OwnershipConf),
				})
			}
			return
		}
	}
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
					LinesAdded:       r.LinesAdded,
					LinesDeleted:     r.LinesDeleted,
					Production:       round1(r.Production),
					Quality:          round1(r.Quality),
					Survival:         round1(r.Survival),
					RobustSurvival:   round1(r.RobustSurvival),
					DormantSurvival:  round1(r.DormantSurvival),
					TestedSurvival:   round1(r.TestedSurvival),
					UntestedSurvival: round1(r.UntestedSurvival),
					Design:           round1(r.Design),
					Breadth:          round1(r.Breadth),
					DebtCleanup:      round1(r.DebtCleanup),
					Indispensability: round1(r.Indispensability),
					Gravity:          round1(r.Gravity),
					Impact:           round1(r.Impact),
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

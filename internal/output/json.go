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
	Name     string           `json:"name"`
	Repos    int              `json:"repos"`
	Members  []jsonMember     `json:"members"`
	BusFactor []jsonBusFactor `json:"bus_factor,omitempty"`
}

type jsonMember struct {
	Rank             int            `json:"rank"`
	Member           string         `json:"member"`
	Active           bool           `json:"active"`
	Commits          int            `json:"commits"`
	Production       float64        `json:"production"`
	Quality          float64        `json:"quality"`
	Survival         float64        `json:"survival"`
	Design           float64        `json:"design"`
	Breadth          float64        `json:"breadth"`
	DebtCleanup      float64        `json:"debt_cleanup"`
	Indispensability float64        `json:"indispensability"`
	Total            float64        `json:"total"`
	Type             string         `json:"type"`
	TypeConf         float64        `json:"type_confidence"`
	Secondary        *jsonArchetype `json:"secondary,omitempty"`
}

type jsonArchetype struct {
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
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
			Design:           round1(r.Design),
			Breadth:          round1(r.Breadth),
			DebtCleanup:      round1(r.DebtCleanup),
			Indispensability: round1(r.Indispensability),
			Total:            round1(r.Total),
			Type:             r.Archetype,
			TypeConf:         r.ArchetypeConf,
		}
		if r.Secondary.Name != "" && r.Secondary.Confidence > 0 {
			m.Secondary = &jsonArchetype{
				Type:       r.Secondary.Name,
				Confidence: r.Secondary.Confidence,
			}
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

func (w *JSONWriter) Flush() error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(w.output)
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

// PrintRankingsJSON is a convenience for single-domain output (not used in multi-domain flow).
func PrintRankingsJSON(domain string, repoCount int, results []scorer.Result, risks []metric.ModuleRisk) {
	w := NewJSONWriter()
	w.AddDomain(domain, repoCount, results, risks)
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "json encode error: %v\n", err)
	}
}

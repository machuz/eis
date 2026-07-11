package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/machuz/eis/v2/internal/timeline"
)

type timelineJSONOutput struct {
	Domain  string               `json:"domain"`
	Span    string               `json:"span"`
	Periods []timelineJSONPeriod `json:"periods"`
	Authors []timelineJSONAuthor `json:"authors"`
}

type timelineJSONPeriod struct {
	Label string `json:"label"`
	Start string `json:"start"`
	End   string `json:"end"`
	// ModuleSurvivalByAuthor: module → author → time-decayed surviving blame
	// mass as of this period's End. Emitted so research/SaaS callers can
	// reconstruct per-(module,author) survival history (Orphaned Gravity,
	// post-departure module collapse). nil (omitted) on windows with no
	// module survival. Map keys are marshalled in sorted order by
	// encoding/json, so output is deterministic. Values are emitted exactly as
	// computed — the surviving-mass sum is a pure function of the blame line
	// set and the window's decay reference (W-02), so there is no run-to-run
	// noise to mask.
	ModuleSurvivalByAuthor map[string]map[string]float64 `json:"module_survival_by_author,omitempty"`
}

type timelineJSONAuthor struct {
	Author      string                     `json:"author"`
	Periods     []timelineJSONAuthorPeriod `json:"periods"`
	Transitions []timelineJSONTransition   `json:"transitions,omitempty"`
}

type timelineJSONAuthorPeriod struct {
	Label            string  `json:"label"`
	Impact           float64 `json:"impact"`
	Production       float64 `json:"production"`
	Catalysis        float64 `json:"catalysis"`
	Survival         float64 `json:"survival"`
	RobustSurvival   float64 `json:"robust_survival"`
	DormantSurvival  float64 `json:"dormant_survival"`
	Design           float64 `json:"design"`
	Breadth          float64 `json:"breadth"`
	DebtCleanup      float64 `json:"debt_cleanup"`
	Indispensability float64 `json:"indispensability"`
	Gravity          float64 `json:"gravity"`
	Commits          int     `json:"commits"`
	LinesAdded       int     `json:"lines_added"`
	LinesDeleted     int     `json:"lines_deleted"`
	Role             string  `json:"role"`
	RoleConf         float64 `json:"role_confidence"`
	Style            string  `json:"style"`
	StyleConf        float64 `json:"style_confidence"`
	State            string  `json:"state"`
	StateConf        float64 `json:"state_confidence"`
}

type timelineJSONTransition struct {
	Axis     string `json:"axis"`
	From     string `json:"from"`
	To       string `json:"to"`
	AtPeriod string `json:"at_period"`
}

// nonEmptyMSBA returns m with empty inner maps dropped, or nil if nothing
// remains, so the JSON field is omitted (omitempty) on windows with no module
// survival. Values pass through unchanged — the surviving-mass sums are exact
// per W-02, so there is nothing to round or mask.
func nonEmptyMSBA(m map[string]map[string]float64) map[string]map[string]float64 {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]map[string]float64, len(m))
	for mod, authors := range m {
		if len(authors) == 0 {
			continue
		}
		out[mod] = authors
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// PrintTimelineJSON outputs timeline data as JSON.
func PrintTimelineJSON(domainName, span string, periods []timeline.PeriodResult, timelines []timeline.AuthorTimeline) {
	out := timelineJSONOutput{
		Domain: domainName,
		Span:   span,
	}

	for _, p := range periods {
		out.Periods = append(out.Periods, timelineJSONPeriod{
			Label:                  p.Label,
			Start:                  p.Start,
			End:                    p.End,
			ModuleSurvivalByAuthor: nonEmptyMSBA(p.ModuleSurvivalByAuthor),
		})
	}

	for _, tl := range timelines {
		author := timelineJSONAuthor{
			Author: tl.Author,
		}

		for _, p := range tl.Periods {
			author.Periods = append(author.Periods, timelineJSONAuthorPeriod{
				Label:            p.Label,
				Impact:           round1(p.Impact),
				Production:       round1(p.Production),
				Catalysis:        round1(p.Catalysis),
				Survival:         round1(p.Survival),
				RobustSurvival:   round1(p.RobustSurvival),
				DormantSurvival:  round1(p.DormantSurvival),
				Design:           round1(p.Design),
				Breadth:          round1(p.Breadth),
				DebtCleanup:      round1(p.DebtCleanup),
				Indispensability: round1(p.Indispensability),
				Gravity:          round1(p.Gravity),
				Commits:          p.TotalCommits,
				LinesAdded:       p.LinesAdded,
				LinesDeleted:     p.LinesDeleted,
				Role:             p.Role,
				RoleConf:         round2(p.RoleConf),
				Style:            p.Style,
				StyleConf:        round2(p.StyleConf),
				State:            p.State,
				StateConf:        round2(p.StateConf),
			})
		}

		for _, tr := range tl.Transitions {
			author.Transitions = append(author.Transitions, timelineJSONTransition{
				Axis:     tr.Axis,
				From:     tr.From,
				To:       tr.To,
				AtPeriod: tr.AtPeriod,
			})
		}

		out.Authors = append(out.Authors, author)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "json encode error: %v\n", err)
	}
}

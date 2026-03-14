package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/machuz/engineering-impact-score/internal/timeline"
)

type timelineJSONOutput struct {
	Domain  string                  `json:"domain"`
	Span    string                  `json:"span"`
	Periods []timelineJSONPeriod    `json:"periods"`
	Authors []timelineJSONAuthor    `json:"authors"`
}

type timelineJSONPeriod struct {
	Label string `json:"label"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type timelineJSONAuthor struct {
	Author      string                    `json:"author"`
	Periods     []timelineJSONAuthorPeriod `json:"periods"`
	Transitions []timelineJSONTransition   `json:"transitions,omitempty"`
}

type timelineJSONAuthorPeriod struct {
	Label            string  `json:"label"`
	Total            float64 `json:"total"`
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
	Commits          int     `json:"commits"`
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

// PrintTimelineJSON outputs timeline data as JSON.
func PrintTimelineJSON(domainName, span string, periods []timeline.PeriodResult, timelines []timeline.AuthorTimeline) {
	out := timelineJSONOutput{
		Domain: domainName,
		Span:   span,
	}

	for _, p := range periods {
		out.Periods = append(out.Periods, timelineJSONPeriod{
			Label: p.Label,
			Start: p.Start,
			End:   p.End,
		})
	}

	for _, tl := range timelines {
		author := timelineJSONAuthor{
			Author: tl.Author,
		}

		for _, p := range tl.Periods {
			author.Periods = append(author.Periods, timelineJSONAuthorPeriod{
				Label:            p.Label,
				Total:            round1(p.Total),
				Production:       round1(p.Production),
				Quality:          round1(p.Quality),
				Survival:         round1(p.Survival),
				RobustSurvival:   round1(p.RobustSurvival),
				DormantSurvival:  round1(p.DormantSurvival),
				Design:           round1(p.Design),
				Breadth:          round1(p.Breadth),
				DebtCleanup:      round1(p.DebtCleanup),
				Indispensability: round1(p.Indispensability),
				Gravity:          round1(p.Gravity),
				Commits:          p.TotalCommits,
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

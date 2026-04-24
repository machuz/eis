package output

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/machuz/eis/v2/internal/timeline"
)

// PrintTimelineCSV outputs timeline data as CSV.
func PrintTimelineCSV(domainName string, timelines []timeline.AuthorTimeline) {
	w := csv.NewWriter(os.Stdout)

	// Header
	w.Write([]string{
		"domain", "author", "period",
		"impact", "production", "quality", "survival", "robust_survival", "dormant_survival",
		"design", "breadth", "debt_cleanup", "indispensability", "gravity",
		"commits", "lines_added", "lines_deleted", "role", "role_conf", "style", "style_conf", "state", "state_conf",
	})

	for _, tl := range timelines {
		for _, p := range tl.Periods {
			w.Write([]string{
				domainName,
				tl.Author,
				p.Label,
				fmt.Sprintf("%.1f", p.Impact),
				fmt.Sprintf("%.1f", p.Production),
				fmt.Sprintf("%.1f", p.Quality),
				fmt.Sprintf("%.1f", p.Survival),
				fmt.Sprintf("%.1f", p.RobustSurvival),
				fmt.Sprintf("%.1f", p.DormantSurvival),
				fmt.Sprintf("%.1f", p.Design),
				fmt.Sprintf("%.1f", p.Breadth),
				fmt.Sprintf("%.1f", p.DebtCleanup),
				fmt.Sprintf("%.1f", p.Indispensability),
				fmt.Sprintf("%.1f", p.Gravity),
				fmt.Sprintf("%d", p.TotalCommits),
				fmt.Sprintf("%d", p.LinesAdded),
				fmt.Sprintf("%d", p.LinesDeleted),
				p.Role,
				fmt.Sprintf("%.2f", p.RoleConf),
				p.Style,
				fmt.Sprintf("%.2f", p.StyleConf),
				p.State,
				fmt.Sprintf("%.2f", p.StateConf),
			})
		}
	}

	w.Flush()
}

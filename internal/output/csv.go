package output

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/machuz/engineering-impact-score/internal/scorer"
)

func PrintRankingsCSV(domain string, results []scorer.Result, writeHeader bool) {
	w := csv.NewWriter(os.Stdout)

	if writeHeader {
		w.Write([]string{"domain", "rank", "member", "active", "commits", "production", "quality", "survival", "design", "breadth", "debt_cleanup", "indispensability", "total", "type", "type_conf", "secondary", "secondary_conf"})
	}

	for i, r := range results {
		active := "false"
		if r.RecentlyActive {
			active = "true"
		}
		w.Write([]string{
			domain,
			fmt.Sprintf("%d", i+1),
			r.Author,
			active,
			fmt.Sprintf("%d", r.TotalCommits),
			fmt.Sprintf("%.1f", r.Production),
			fmt.Sprintf("%.1f", r.Quality),
			fmt.Sprintf("%.1f", r.Survival),
			fmt.Sprintf("%.1f", r.Design),
			fmt.Sprintf("%.1f", r.Breadth),
			fmt.Sprintf("%.1f", r.DebtCleanup),
			fmt.Sprintf("%.1f", r.Indispensability),
			fmt.Sprintf("%.1f", r.Total),
			r.Archetype,
			fmt.Sprintf("%.2f", r.ArchetypeConf),
			r.Secondary.Name,
			fmt.Sprintf("%.2f", r.Secondary.Confidence),
		})
	}

	w.Flush()
}

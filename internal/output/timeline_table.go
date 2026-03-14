package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/machuz/engineering-impact-score/internal/timeline"
)

// PrintTimelineTable outputs timeline data as a colored terminal table.
func PrintTimelineTable(domainName, span string, timelines []timeline.AuthorTimeline) {
	if len(timelines) == 0 {
		return
	}

	fmt.Println()
	color.New(color.FgHiCyan, color.Bold).Printf("═══ %s Timeline (%s spans) ═══\n", domainName, span)

	headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
	nameFmt := color.New(color.FgHiYellow, color.Bold).SprintfFunc()
	dimFmt := color.New(color.FgHiBlack).SprintfFunc()
	labelFmt := color.New(color.FgHiBlue).SprintfFunc()

	for _, tl := range timelines {
		fmt.Println()
		fmt.Printf("--- %s ---\n", nameFmt("%s", tl.Author))

		// Header
		fmt.Printf("%-18s %6s %5s %5s %5s %7s  %-12s %-12s %-12s\n",
			headerFmt("Period"),
			headerFmt("Total"),
			headerFmt("Prod"),
			headerFmt("Qual"),
			headerFmt("Surv"),
			headerFmt("Design"),
			headerFmt("Role"),
			headerFmt("Style"),
			headerFmt("State"),
		)

		for _, p := range tl.Periods {
			if p.Total == 0 && p.TotalCommits == 0 {
				fmt.Printf("%-18s %6s %5s %5s %5s %7s  %-12s %-12s %-12s\n",
					p.Label, dimFmt("—"), dimFmt("—"), dimFmt("—"), dimFmt("—"), dimFmt("—"),
					dimFmt("—"), dimFmt("—"), dimFmt("—"),
				)
				continue
			}

			totalStr := formatTotal(p.Total)

			survStr := fmt.Sprintf("%.0f", p.Survival)
			if p.RobustSurvival > 0 || p.DormantSurvival > 0 {
				survStr = fmt.Sprintf("%.0f", p.RobustSurvival)
			}

			roleStr := dimFmt("—")
			if p.Role != "" && p.Role != "—" {
				roleStr = labelFmt("%s", p.Role)
			}
			styleStr := dimFmt("—")
			if p.Style != "" && p.Style != "—" {
				styleStr = labelFmt("%s", p.Style)
			}
			stateStr := dimFmt("—")
			if p.State != "" && p.State != "—" {
				stateStr = labelFmt("%s", p.State)
			}

			fmt.Printf("%-18s %6s %5.0f %5.0f %5s %7.0f  %-12s %-12s %-12s\n",
				p.Label,
				totalStr,
				p.Production,
				p.Quality,
				survStr,
				p.Design,
				roleStr,
				styleStr,
				stateStr,
			)
		}

		// Print transitions
		if len(tl.Transitions) > 0 {
			transitionFmt := color.New(color.FgHiGreen, color.Bold).SprintfFunc()
			var parts []string
			for _, tr := range tl.Transitions {
				parts = append(parts, transitionFmt("↑ %s: %s→%s", tr.Axis, tr.From, tr.To))
			}
			fmt.Printf("%s %s\n", strings.Repeat(" ", 18), strings.Join(parts, "  "))
		}
	}

	fmt.Println()

	// Print summary of notable transitions across all timelines
	var notableTransitions []string
	for _, tl := range timelines {
		for _, tr := range tl.Transitions {
			notableTransitions = append(notableTransitions,
				fmt.Sprintf("%s: %s %s→%s (%s)", tl.Author, tr.Axis, tr.From, tr.To, tr.AtPeriod))
		}
	}

	if len(notableTransitions) > 0 {
		color.New(color.FgHiGreen, color.Bold).Println("Notable transitions:")
		for _, t := range notableTransitions {
			fmt.Fprintf(os.Stdout, "  • %s\n", t)
		}
		fmt.Println()
	}
}

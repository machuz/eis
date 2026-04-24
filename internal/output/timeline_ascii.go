package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/guptarohit/asciigraph"
	"github.com/machuz/eis/v2/internal/timeline"
)

// PrintTimelineASCII renders author timelines as ASCII line charts.
func PrintTimelineASCII(domainName, span string, timelines []timeline.AuthorTimeline) {
	if len(timelines) == 0 {
		return
	}

	fmt.Println()
	color.New(color.FgHiCyan, color.Bold).Printf("=== %s Timeline (%s spans) ===\n", domainName, span)

	for _, tl := range timelines {
		fmt.Println()
		color.New(color.FgHiYellow, color.Bold).Printf("--- %s ---\n", tl.Author)

		// Build series data
		impactData := make([]float64, 0, len(tl.Periods))
		prodData := make([]float64, 0, len(tl.Periods))
		qualData := make([]float64, 0, len(tl.Periods))
		survData := make([]float64, 0, len(tl.Periods))
		designData := make([]float64, 0, len(tl.Periods))
		var labels []string

		for _, p := range tl.Periods {
			impactData = append(impactData, p.Impact)
			prodData = append(prodData, p.Production)
			qualData = append(qualData, p.Quality)
			survData = append(survData, p.Survival)
			designData = append(designData, p.Design)
			labels = append(labels, p.Label)
		}

		data := [][]float64{impactData, prodData, qualData, survData, designData}

		// Calculate chart width to match label axis spacing
		chartWidth := 60
		if len(labels) > 1 {
			spacing := chartWidth / (len(labels) - 1)
			if spacing < len(labels[0])+2 {
				spacing = len(labels[0]) + 2
				chartWidth = spacing * (len(labels) - 1)
			}
		}

		chart := asciigraph.PlotMany(data,
			asciigraph.Height(15),
			asciigraph.Width(chartWidth),
			asciigraph.SeriesColors(
				asciigraph.Blue,
				asciigraph.Green,
				asciigraph.Yellow,
				asciigraph.Red,
				asciigraph.Cyan,
			),
			asciigraph.Caption(buildLabelAxis(labels)),
		)
		fmt.Println(chart)

		// Legend
		fmt.Printf("  %s  %s  %s  %s  %s\n",
			color.BlueString("Impact"),
			color.GreenString("Production"),
			color.YellowString("Quality"),
			color.RedString("Survival"),
			color.CyanString("Design"),
		)

		// Transitions
		if len(tl.Transitions) > 0 {
			fmt.Println()
			for _, tr := range tl.Transitions {
				fmt.Printf("  %s %s: %s -> %s (%s)\n",
					color.HiGreenString("*"),
					tr.Axis, tr.From, tr.To, tr.AtPeriod,
				)
			}
		}
	}

	fmt.Println()
}

// buildLabelAxis creates a caption string showing period labels spaced across the chart width.
func buildLabelAxis(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	if len(labels) == 1 {
		return labels[0]
	}

	// Approximate spacing: chart data area is roughly the number of data points spread across
	// We build a simple label line with even spacing
	totalWidth := 60
	if len(labels) > 1 {
		spacing := totalWidth / (len(labels) - 1)
		if spacing < len(labels[0])+2 {
			spacing = len(labels[0]) + 2
		}

		var b strings.Builder
		for i, l := range labels {
			if i == len(labels)-1 {
				b.WriteString(l)
			} else {
				padded := l
				if len(padded) < spacing {
					padded += strings.Repeat(" ", spacing-len(padded))
				}
				b.WriteString(padded)
			}
		}
		return b.String()
	}

	return strings.Join(labels, "  ")
}

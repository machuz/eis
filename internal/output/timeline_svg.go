package output

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/machuz/eis/v2/internal/timeline"
	chart "github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

var sanitizeRe = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func sanitizeFilename(s string) string {
	s = strings.TrimSpace(s)
	s = sanitizeRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "unknown"
	}
	return s
}

// seriesConfig defines the 5 score dimensions to chart.
type seriesConfig struct {
	Name  string
	Color drawing.Color
}

// Gruvbox Dark palette
var (
	gruvboxBg     = drawing.Color{R: 40, G: 40, B: 40, A: 255}     // #282828
	gruvboxBgSoft = drawing.Color{R: 50, G: 48, B: 47, A: 255}     // #32302f
	gruvboxFg     = drawing.Color{R: 235, G: 219, B: 178, A: 255}  // #ebdbb2
	gruvboxFgDim  = drawing.Color{R: 168, G: 153, B: 132, A: 255}  // #a89984
	gruvboxGrid   = drawing.Color{R: 80, G: 73, B: 69, A: 255}     // #504945
)

var scoreSeries = []seriesConfig{
	{"Impact", drawing.Color{R: 69, G: 133, B: 136, A: 255}},      // #458588 gruvbox blue
	{"Production", drawing.Color{R: 152, G: 151, B: 26, A: 255}},  // #98971a gruvbox green
	{"Quality", drawing.Color{R: 250, G: 189, B: 47, A: 255}},     // #fabd2f gruvbox yellow
	{"Survival", drawing.Color{R: 211, G: 134, B: 155, A: 255}},   // #d3869b gruvbox purple
	{"Design", drawing.Color{R: 142, G: 192, B: 124, A: 255}},     // #8ec07c gruvbox aqua
}

// WriteTimelineSVG generates SVG chart files for each author and team timeline.
// It writes files into dir and returns the list of generated file paths.
func WriteTimelineSVG(dir string, domainTimelines []DomainTimelineData, teamTimelines []timeline.TeamTimeline) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	var generated []string

	// Author charts: one SVG per author per domain
	for _, dt := range domainTimelines {
		for _, tl := range dt.Timelines {
			if len(tl.Periods) == 0 {
				continue
			}

			fname := fmt.Sprintf("%s-%s.svg", sanitizeFilename(dt.DomainName), sanitizeFilename(tl.Author))
			fpath := filepath.Join(dir, fname)

			title := fmt.Sprintf("%s - %s", tl.Author, dt.DomainName)
			if err := writeAuthorSVG(fpath, title, tl.Periods); err != nil {
				return generated, fmt.Errorf("write SVG %s: %w", fname, err)
			}
			generated = append(generated, fpath)
		}
	}

	// Team charts: one SVG per team per domain
	for _, tl := range teamTimelines {
		if len(tl.Periods) == 0 {
			continue
		}

		fname := fmt.Sprintf("team-%s-%s.svg", sanitizeFilename(tl.TeamName), sanitizeFilename(tl.Domain))
		fpath := filepath.Join(dir, fname)

		title := fmt.Sprintf("Team %s - %s", tl.TeamName, tl.Domain)
		if err := writeTeamSVG(fpath, title, tl.Periods); err != nil {
			return generated, fmt.Errorf("write SVG %s: %w", fname, err)
		}
		generated = append(generated, fpath)
	}

	return generated, nil
}

func writeAuthorSVG(path, title string, periods []timeline.AuthorPeriod) error {
	labels := make([]string, len(periods))
	values := make([][]float64, len(scoreSeries))
	for i := range scoreSeries {
		values[i] = make([]float64, len(periods))
	}

	for i, p := range periods {
		labels[i] = p.Label
		values[0][i] = p.Impact
		values[1][i] = p.Production
		values[2][i] = p.Quality
		values[3][i] = p.Survival
		values[4][i] = p.Design
	}

	return writeSVGChart(path, title, labels, values)
}

func writeTeamSVG(path, title string, periods []timeline.TeamPeriodSnapshot) error {
	labels := make([]string, len(periods))
	values := make([][]float64, len(scoreSeries))
	for i := range scoreSeries {
		values[i] = make([]float64, len(periods))
	}

	for i, p := range periods {
		labels[i] = p.Label
		values[0][i] = p.AvgImpact
		values[1][i] = p.AvgProduction
		values[2][i] = p.AvgQuality
		values[3][i] = p.AvgSurvival
		values[4][i] = p.AvgDesign
	}

	return writeSVGChart(path, title, labels, values)
}

func writeSVGChart(path, title string, labels []string, values [][]float64) error {
	n := len(labels)
	if n == 0 {
		return nil
	}

	// Build X values as indices
	xValues := make([]float64, n)
	for i := range xValues {
		xValues[i] = float64(i)
	}

	// Build custom ticks for X axis
	ticks := make([]chart.Tick, n)
	for i, label := range labels {
		ticks[i] = chart.Tick{
			Value: float64(i),
			Label: label,
		}
	}

	// Build series
	var series []chart.Series
	for si, sc := range scoreSeries {
		series = append(series, chart.ContinuousSeries{
			Name: sc.Name,
			Style: chart.Style{
				StrokeColor: sc.Color,
				StrokeWidth: 2.5,
				DotColor:    sc.Color,
				DotWidth:    5,
			},
			XValues: xValues,
			YValues: values[si],
		})
	}

	graph := chart.Chart{
		Title: title,
		TitleStyle: chart.Style{
			FontSize:  16,
			FontColor: gruvboxFg,
		},
		Width:  1000,
		Height: 500,
		Background: chart.Style{
			FillColor: gruvboxBg,
			Padding: chart.Box{
				Top:    50,
				Left:   20,
				Right:  20,
				Bottom: 20,
			},
		},
		Canvas: chart.Style{
			FillColor: gruvboxBgSoft,
		},
		XAxis: chart.XAxis{
			Name: "Period",
			NameStyle: chart.Style{
				FontSize:  12,
				FontColor: gruvboxFgDim,
			},
			Style: chart.Style{
				FontSize:  11,
				FontColor: gruvboxFgDim,
				StrokeColor: gruvboxGrid,
			},
			Ticks: ticks,
			GridMajorStyle: chart.Style{
				StrokeColor: gruvboxGrid,
				StrokeWidth: 1,
			},
		},
		YAxis: chart.YAxis{
			Name: "Score",
			NameStyle: chart.Style{
				FontSize:  12,
				FontColor: gruvboxFgDim,
			},
			Style: chart.Style{
				FontSize:  11,
				FontColor: gruvboxFgDim,
				StrokeColor: gruvboxGrid,
			},
			Range: &chart.ContinuousRange{Min: 0, Max: 100},
			GridMajorStyle: chart.Style{
				StrokeColor: gruvboxGrid,
				StrokeWidth: 1,
			},
		},
		Series: series,
	}

	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph, chart.Style{
			FontSize:  11,
			FontColor: gruvboxFg,
			FillColor: gruvboxBgSoft,
			StrokeColor: gruvboxGrid,
		}),
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return graph.Render(chart.SVG, f)
}

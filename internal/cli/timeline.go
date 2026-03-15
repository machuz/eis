package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/machuz/engineering-impact-score/internal/config"
	"github.com/machuz/engineering-impact-score/internal/output"
	"github.com/machuz/engineering-impact-score/internal/timeline"
	pkgtimeline "github.com/machuz/engineering-impact-score/pkg/timeline"
)

func runTimeline(args []string) error {
	fs := flag.NewFlagSet("timeline", flag.ExitOnError)
	configPath := fs.String("config", "", "Config file path")
	spanFlag := fs.String("span", "3m", "Period span: 1w, 1m, 3m, 6m, 1y")
	periodsFlag := fs.Int("periods", 4, "Number of periods to show (0=all)")
	sinceFlag := fs.String("since", "", "Start date (e.g. 2024-01-01, overrides --periods)")
	formatFlag := fs.String("format", "table", "Output format: table, csv, json, ascii, html, svg")
	outputFlag := fs.String("output", "", "Output file/directory path (html: file path, svg: directory; defaults: eis-timeline.html / current dir)")
	recursive := fs.Bool("recursive", false, "Recursively find git repos under given paths")
	maxDepth := fs.Int("depth", 2, "Max directory depth for recursive search")
	domainFilter := fs.String("domain", "", "Only analyze specific domain")
	authorFilter := fs.String("author", "", "Filter to specific author(s), comma-separated")
	workers := fs.Int("workers", 4, "Number of concurrent blame workers")
	sampleSize := fs.Int("sample", 0, "Max files to blame per repo (overrides config)")
	tau := fs.Float64("tau", 0, "Survival decay parameter (overrides config)")
	activeDays := fs.Int("active-days", 0, "Days to consider author active (overrides config)")
	pressureMode := fs.String("pressure-mode", "include", "Change pressure mode: include or ignore")
	verbose := fs.Bool("verbose", false, "Show detailed debug output")
	_ = fs.Bool("no-cache", false, "Skip disk cache (currently unused with library mode)")

	flagArgs, pathArgs := separateArgs(args, fs)
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	repoPaths := pathArgs
	if len(repoPaths) == 0 {
		repoPaths = []string{"."}
	}

	// Resolve to absolute paths
	for i, p := range repoPaths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("resolve path %s: %w", p, err)
		}
		repoPaths[i] = abs
	}

	// Recursive: find git repos
	if *recursive {
		var discovered []string
		for _, root := range repoPaths {
			repos, err := findGitRepos(root, *maxDepth)
			if err != nil {
				return fmt.Errorf("scan %s: %w", root, err)
			}
			discovered = append(discovered, repos...)
		}
		if len(discovered) == 0 {
			return fmt.Errorf("no git repos found under %v (depth=%d)", repoPaths, *maxDepth)
		}
		repoPaths = discovered
		fmt.Fprintf(os.Stderr, "Found %d git repos\n\n", len(repoPaths))
	}

	explicitConfig := *configPath != ""
	if !explicitConfig {
		*configPath = "eis.yaml"
	}

	// Load config
	cfg, err := config.Load(*configPath, explicitConfig)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	quiet := *formatFlag == "json" || *formatFlag == "csv" || *formatFlag == "html" || *formatFlag == "svg"
	spinnerQuiet = quiet

	if !quiet && len(cfg.Aliases) > 0 {
		fmt.Fprintf(os.Stderr, "Loaded %d author aliases from config\n", len(cfg.Aliases))
	}

	// Parse author filter and adjust exclude list
	var authorList []string
	if *authorFilter != "" {
		for _, a := range strings.Split(*authorFilter, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				authorList = append(authorList, a)
			}
		}
	}

	if len(authorList) > 0 {
		includeCanonical := make(map[string]bool)
		for _, a := range authorList {
			includeCanonical[strings.ToLower(cfg.ResolveAuthor(a))] = true
		}

		var newExclude []string
		for _, exc := range cfg.ExcludeAuthors {
			canonical := strings.ToLower(cfg.ResolveAuthor(exc))
			if !includeCanonical[strings.ToLower(exc)] && !includeCanonical[canonical] {
				newExclude = append(newExclude, exc)
			}
		}
		cfg.ExcludeAuthors = newExclude
	}

	// Print period info before running analysis
	spanMonths, spanDays, err := pkgtimeline.ParseSpan(*spanFlag)
	if err != nil {
		return err
	}

	sinceDate := parseTimelineSince(*sinceFlag)
	windows := pkgtimeline.BuildPeriods(spanMonths, spanDays, *periodsFlag, sinceDate, time.Now())
	if !quiet {
		fmt.Fprintf(os.Stderr, "Timeline: %d periods (%s span)\n", len(windows), *spanFlag)
		for _, w := range windows {
			fmt.Fprintf(os.Stderr, "  %s: %s → %s\n", w.Label, w.Start.Format("2006-01-02"), w.End.Format("2006-01-02"))
		}
		fmt.Fprintln(os.Stderr)
	}

	// Build callbacks for progress UI
	cb := &pkgtimeline.Callbacks{}
	if !quiet {
		cb.OnRepoStart = func(repoName string, d string) {
			domainLabel := color.New(color.FgCyan).Sprintf("[%s]", d)
			color.New(color.Bold).Printf("Loading: %s %s\n", repoName, domainLabel)
		}
		cb.OnPeriodStart = func(label string, index, total int) {
			fmt.Fprintf(os.Stderr, "\n[Period %d/%d] %s\n", index+1, total, label)
		}
	}
	if *verbose {
		cb.OnVerbose = func(msg string) {
			fmt.Fprintf(os.Stderr, "\n%s", msg)
		}
	}

	// Run the library analysis
	opts := pkgtimeline.Options{
		Span:         *spanFlag,
		Periods:      *periodsFlag,
		Since:        *sinceFlag,
		Workers:      *workers,
		DomainFilter: *domainFilter,
		PressureMode: *pressureMode,
		Tau:          *tau,
		SampleSize:   *sampleSize,
		ActiveDays:   *activeDays,
	}

	domainTimelines, err := pkgtimeline.Run(opts, repoPaths, cfg, cb)
	if err != nil {
		return err
	}

	// Build per-domain author timelines
	type builtDomainTimeline struct {
		domain    string
		span      string
		periods   []timeline.PeriodResult
		timelines []timeline.AuthorTimeline
	}

	var builtTimelines []builtDomainTimeline
	for _, dt := range domainTimelines {
		timelines := timeline.BuildTimeline(dt.Periods)

		// Apply author filter
		if len(authorList) > 0 {
			var filtered []timeline.AuthorTimeline
			for _, tl := range timelines {
				for _, a := range authorList {
					if strings.EqualFold(tl.Author, a) {
						filtered = append(filtered, tl)
						break
					}
				}
			}
			timelines = filtered
		}

		builtTimelines = append(builtTimelines, builtDomainTimeline{
			domain:    dt.Domain,
			span:      *spanFlag,
			periods:   dt.Periods,
			timelines: timelines,
		})
	}

	// Build team timelines
	var teamTimelines []timeline.TeamTimeline
	for _, dt := range domainTimelines {
		teamPeriodResults := pkgtimeline.BuildTeamPeriodResults(dt.Domain, dt.Periods, cfg)

		for teamName, periods := range teamPeriodResults {
			tl := timeline.BuildTeamTimeline(teamName, dt.Domain, periods)
			teamTimelines = append(teamTimelines, tl)
		}
	}

	// Output
	if *formatFlag == "html" || *formatFlag == "svg" {
		var htmlDomains []output.DomainTimelineData
		for _, bt := range builtTimelines {
			htmlDomains = append(htmlDomains, output.DomainTimelineData{
				DomainName: bt.domain,
				Span:       bt.span,
				Timelines:  bt.timelines,
			})
		}

		if *formatFlag == "html" {
			outPath := *outputFlag
			if outPath == "" {
				outPath = "eis-timeline.html"
			}

			f, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("create output file: %w", err)
			}
			defer f.Close()

			if err := output.WriteTimelineHTML(f, htmlDomains, teamTimelines); err != nil {
				return fmt.Errorf("write HTML: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Timeline HTML written to %s\n", outPath)
		} else {
			// SVG
			outDir := *outputFlag
			if outDir == "" {
				outDir = "."
			}

			files, err := output.WriteTimelineSVG(outDir, htmlDomains, teamTimelines)
			if err != nil {
				return fmt.Errorf("write SVG: %w", err)
			}

			for _, f := range files {
				fmt.Fprintf(os.Stderr, "SVG written: %s\n", f)
			}
			fmt.Fprintf(os.Stderr, "Generated %d SVG files\n", len(files))
		}
	} else {
		for _, bt := range builtTimelines {
			switch *formatFlag {
			case "json":
				output.PrintTimelineJSON(bt.domain, *spanFlag, bt.periods, bt.timelines)
			case "csv":
				output.PrintTimelineCSV(bt.domain, bt.timelines)
			case "ascii":
				output.PrintTimelineASCII(bt.domain, *spanFlag, bt.timelines)
			default:
				output.PrintTimelineTable(bt.domain, *spanFlag, bt.timelines)
			}
		}

		if len(teamTimelines) > 0 {
			switch *formatFlag {
			case "json":
				output.PrintTeamTimelineJSON(teamTimelines)
			case "csv":
				output.PrintTeamTimelineCSV(teamTimelines)
			case "ascii":
				for _, tl := range teamTimelines {
					output.PrintTeamTimelineASCII(tl)
				}
			default:
				for _, tl := range teamTimelines {
					output.PrintTeamTimelineTable(tl)
				}
			}
		}
	}

	if !quiet {
		color.New(color.FgHiBlack).Fprintf(os.Stderr, "\nTimeline analysis complete\n")
	}

	return nil
}

// parseTimelineSince parses a date string, returning zero time on empty/invalid input.
func parseTimelineSince(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse("2006-01-02", s)
	return t
}

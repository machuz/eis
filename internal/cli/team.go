package cli

import (
	"flag"
	"fmt"

	"github.com/machuz/engineering-impact-score/internal/config"
	"github.com/machuz/engineering-impact-score/internal/output"
	"github.com/machuz/engineering-impact-score/internal/team"
)

func runTeam(args []string) error {
	fs := flag.NewFlagSet("team", flag.ExitOnError)
	configPath := fs.String("config", "", "Config file path")
	tau := fs.Float64("tau", 0, "Survival decay parameter (overrides config)")
	sampleSize := fs.Int("sample", 0, "Max files to blame per repo (overrides config)")
	workers := fs.Int("workers", 4, "Number of concurrent blame workers")
	recursive := fs.Bool("recursive", false, "Recursively find git repos under given paths")
	maxDepth := fs.Int("depth", 2, "Max directory depth for recursive search")
	formatFlag := fs.String("format", "table", "Output format: table, csv, json")
	pressureMode := fs.String("pressure-mode", "include", "Change pressure mode: include or ignore")
	activeDays := fs.Int("active-days", 0, "Days to consider author active (overrides config)")
	domainFilter := fs.String("domain", "", "Only analyze repos in this domain")
	verbose := fs.Bool("verbose", false, "Show detailed debug output (file-level timing)")
	noCache := fs.Bool("no-cache", false, "Skip disk cache")

	flagArgs, pathArgs := separateArgs(args, fs)
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	explicitConfig := *configPath != ""
	if !explicitConfig {
		*configPath = "eis.yaml"
	}

	opts := AnalyzeOptions{
		ConfigPath:     *configPath,
		ExplicitConfig: explicitConfig,
		Tau:            *tau,
		SampleSize:     *sampleSize,
		Workers:        *workers,
		Recursive:      *recursive,
		MaxDepth:       *maxDepth,
		Format:         *formatFlag,
		PressureMode:   *pressureMode,
		ActiveDays:     *activeDays,
		DomainFilter:   *domainFilter,
		Verbose:        *verbose,
		NoCache:        *noCache,
	}

	// Run the shared analysis pipeline
	domainResults, cfg, err := RunAnalyzePipeline(opts, pathArgs)
	if err != nil {
		return err
	}

	// Aggregate into teams
	teamResults := aggregateTeams(domainResults, cfg)

	if len(teamResults) == 0 {
		return fmt.Errorf("no team results to display")
	}

	// Output
	switch opts.Format {
	case "json":
		return output.PrintTeamJSON(teamResults)
	case "csv":
		output.PrintTeamCSV(teamResults)
	default:
		output.PrintTeamTable(teamResults)
	}

	return nil
}

// aggregateTeams builds TeamResults from domain results.
// If cfg.Teams is configured, uses those groupings.
// Otherwise, treats each domain as a single team.
func aggregateTeams(domainResults []DomainResults, cfg *config.Config) []team.TeamResult {
	var results []team.TeamResult

	if len(cfg.Teams) > 0 {
		// Build a lookup: domain results by domain name (case-insensitive)
		domainMap := make(map[string]DomainResults)
		for _, dr := range domainResults {
			domainMap[string(dr.Domain)] = dr
		}

		for teamName, entry := range cfg.Teams {
			// Find matching domain results
			dr, ok := domainMap[entry.Domain]
			if !ok {
				// Try case-insensitive match
				for key, val := range domainMap {
					if key == entry.Domain {
						dr = val
						ok = true
						break
					}
				}
			}
			if !ok {
				continue
			}

			tr := team.Aggregate(teamName, entry.Domain, dr.RepoCount, dr.Results, entry.Members)
			if tr.MemberCount > 0 {
				results = append(results, tr)
			}
		}
	} else {
		// No teams configured: each domain = one team
		for _, dr := range domainResults {
			tr := team.Aggregate(string(dr.Domain), string(dr.Domain), dr.RepoCount, dr.Results, nil)
			if tr.MemberCount > 0 {
				results = append(results, tr)
			}
		}
	}

	return results
}

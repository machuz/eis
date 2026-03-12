package cli

import "fmt"

var version = "dev"

func Run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "analyze":
		return runAnalyze(args[1:])
	case "version":
		fmt.Printf("eis v%s\n", version)
		return nil
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printUsage() {
	fmt.Println(`eis - Engineering Impact Score

Usage:
  eis analyze [path...]       Analyze git repos and output rankings
  eis version                 Print version
  eis help                    Show this help

Examples:
  eis analyze .                                  Analyze current repo
  eis analyze /path/to/repo                      Analyze a single repo
  eis analyze /path/to/repo1 /path/to/repo2      Analyze multiple repos
  eis analyze --recursive /path/to/workspace     Auto-detect repos under directory
  eis analyze --recursive --depth 3 ~/projects   Search up to 3 levels deep
  eis analyze --format json --recursive ~/work   Output as JSON
  eis analyze --format csv --recursive ~/work    Output as CSV

Options for analyze:
  --config <path>             Config file (default: eis.yaml in CWD)
  --recursive                 Recursively find git repos under given paths
  --depth <n>                 Max directory depth for recursive search (default: 2)
  --tau <days>                Survival decay parameter (default: 180)
  --sample <n>                Max files to blame per repo (default: 500)
  --workers <n>               Number of concurrent blame workers (default: 4)
  --format <fmt>              Output format: table, csv, json (default: table)

Config file (eis.yaml):
  aliases:                    Map git author names to canonical names
  exclude_authors:            Authors to exclude from analysis
  exclude_file_patterns:      File patterns to exclude from production
  architecture_patterns:      Patterns for design score
  blame_extensions:           File extensions for blame analysis
  weights:                    Axis weights (must sum to 1.0)
  tau:                        Survival decay (default: 180 days)
  debt_threshold:             Min events for debt score (default: 10)`)
}

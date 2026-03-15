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
	case "team":
		return runTeam(args[1:])
	case "timeline":
		return runTimeline(args[1:])
	case "cache":
		return runCache(args[1:])
	case "version":
		fmt.Printf("\n"+
			"        \033[33m✦\033[0m       \033[36m*\033[0m        \033[35m✧\033[0m\n"+
			"\n"+
			"         ╭────────╮\n"+
			"        │  \033[33m✦\033[0m  \033[36m.\033[0m  \033[35m✦\033[0m │\n"+
			"         ╰────┬───╯\n"+
			"     \033[36m.\033[0m        │\n"+
			"              │\n"+
			"           ___│___\n"+
			"          /_______\\\n"+
			"\n"+
			"     \033[35m✧\033[0m     the Git Telescope  v%s     \033[33m✦\033[0m\n\n", version)
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
	fmt.Println(`eis - Engineering Impact Score (pronounced "ace") — the Git Telescope

Usage:
  eis analyze [path...]       Analyze git repos and output individual rankings
  eis team [path...]          Analyze and aggregate into team-level metrics
  eis timeline [path...]      Track score evolution over time periods
  eis cache clear              Clear cached data
  eis cache status             Show cache size
  eis version                 Print version
  eis help                    Show this help

Examples:
  eis analyze .                                  Analyze current repo
  eis analyze --recursive /path/to/workspace     Auto-detect repos under directory
  eis analyze --format json --recursive ~/work   Output as JSON
  eis analyze --recursive --per-repo ~/work      Per-repository breakdown
  eis team --recursive /path/to/workspace        Team analysis (auto-group by domain)
  eis team --config eis.yaml --recursive ~/work  Team analysis with team config
  eis timeline --recursive ~/work                Timeline (default: 4 periods, 3m span)
  eis timeline --span 6m --periods 0 ~/work      Full history in 6-month spans
  eis timeline --since 2024-01-01 ~/work         From specific date
  eis timeline --author machuz,ponsaaan ~/work    Filter to specific authors

Options (shared by analyze, team, and timeline):
  --no-cache                  Skip disk cache (re-run all git operations)
  --config <path>             Config file (default: eis.yaml in CWD)
  --recursive                 Recursively find git repos under given paths
  --depth <n>                 Max directory depth for recursive search (default: 2)
  --tau <days>                Survival decay parameter (default: 180)
  --sample <n>                Max files to blame per repo (default: 500)
  --workers <n>               Number of concurrent blame workers (default: 4)
  --format <fmt>              Output format: table, csv, json (default: table)
  --active-days <n>           Days to consider author active (default: 30)
  --pressure-mode <mode>      Change pressure mode: include or ignore (default: include)
  --domain <name>             Filter to single domain
  --verbose                   Show detailed debug output (file-level timing, slow ops)
  --per-repo                  Show per-repository breakdown (with --recursive)

Timeline-specific options:
  --span <period>             Period span: 3m, 6m, 1y (default: 3m)
  --periods <n>               Number of periods to show (default: 4, 0=all)
  --since <date>              Start date (e.g. 2024-01-01, overrides --periods)
  --author <names>            Filter to specific author(s), comma-separated

Config file (eis.yaml):
  aliases:                    Map git author names to canonical names
  exclude_authors:            Authors to exclude from analysis
  exclude_file_patterns:      File patterns to exclude from production
  architecture_patterns:      Patterns for design score
  blame_extensions:           File extensions for blame analysis
  weights:                    Axis weights (must sum to 1.0)
  tau:                        Survival decay (default: 180 days)
  active_days:                Days to consider author active (default: 30)
  debt_threshold:             Min events for debt score (default: 10)
  blame_timeout:              Per-file blame timeout in seconds (default: 120)
  teams:                      Team definitions (optional, see config.example.yaml)`)
}

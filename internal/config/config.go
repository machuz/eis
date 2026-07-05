package config

import (
	"fmt"
	"math"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/machuz/eis/v2/internal/domain"
)

type Config struct {
	Tau                  float64  `yaml:"tau"`
	SampleSize           int      `yaml:"sample_size"`
	DebtThreshold        int      `yaml:"debt_threshold"`
	ExcludeFilePatterns  []string `yaml:"exclude_file_patterns"`
	ArchitecturePatterns []string `yaml:"architecture_patterns"`
	BlameExtensions      []string `yaml:"blame_extensions"`
	ExcludeAuthors       []string `yaml:"exclude_authors"`
	// BotCoAuthorPatterns are extra substrings (matched in a Co-authored-by name or
	// email) that mark a co-author as a bot/AI tool, on top of the built-in list
	// ([bot], Claude, Copilot, Cursor, …). Excluded co-authors don't dilute the
	// human author's gravity. Add your own AI tools here.
	BotCoAuthorPatterns []string          `yaml:"bot_coauthor_patterns"`
	Aliases             map[string]string `yaml:"aliases"`
	// AuthorGitHubIDs pins a canonical author name (post-alias) to a GitHub numeric
	// user id. It exists for the research ingest's identity resolver: a founder whose
	// dominant commit email predates GitHub (e.g. Linus Torvalds' 2005 osdl.org
	// address) cannot be resolved email→id and would otherwise be dropped. Pinning
	// the name lets such a contributor appear under their real id. Ignored by the
	// analysis pipeline (which keys by name); consumed only by the ingest tool.
	AuthorGitHubIDs    map[string]int64     `yaml:"author_github_ids"`
	Weights            Weights              `yaml:"weights"`
	BusFactor          BusFactor            `yaml:"bus_factor"`
	DefaultDomains     *bool                `yaml:"default_domains"`
	Domains            DomainsConfig        `yaml:"domains"`
	Teams              map[string]TeamEntry `yaml:"teams"`
	ProductionDailyRef float64              `yaml:"production_daily_ref"`
	ExcludeRepos       []string             `yaml:"exclude_repos"`
	ActiveDays         int                  `yaml:"active_days"`
	BlameTimeout       int                  `yaml:"blame_timeout"`
	// BlameMoveDetection controls git blame move/copy detection: off | file (-M,
	// default) | commit (-M -C) | full (-M -C -C). Higher levels recover the
	// original author of lines moved across files at higher cost on large repos.
	BlameMoveDetection string `yaml:"blame_move_detection"`
	// AttributeSubtreeSquash controls whether lines introduced by a
	// `git subtree ... --squash` import are credited to the integrator who ran
	// the command. Default false = suppress: such lines collapse the true
	// authors into the squasher (blame has no one else to point at, and --squash
	// discarded the originals), so crediting them is always wrong. Set true only
	// to restore the pre-fix behavior. See git.SubtreeSquashCommits.
	AttributeSubtreeSquash bool `yaml:"attribute_subtree_squash"`
	// ModulePatterns is the set of glob patterns used by module resolution.
	// Each pattern's components are matched against a file path; `*` matches
	// a SINGLE path component (no `/`). On match, the module identifier is
	// the substituted prefix of length len(pattern.components).
	//
	// Examples:
	//   "services/*"     services/ace/backend/foo.go -> services/ace
	//   "apps/*/lib"     apps/api/lib/utils/foo.go  -> apps/api/lib
	//
	// When multiple patterns match the same path, the LONGEST resulting
	// module identifier wins; ties are broken by first-listed.
	// When ModulePatterns is empty (and no per-repo override applies),
	// DefaultModulePatterns is used. Paths matching nothing fall back to
	// the conservative 2-component default (e.g. "a/b/c/d.go" -> "a/b").
	ModulePatterns []string `yaml:"module_patterns"`
	// ExcludeModules drops resolved modules from ALL module-topology metrics
	// (module signals, Breadth, cochange, ownership fragmentation,
	// indispensability, per-module test ratio). Each pattern's components are
	// matched component-wise (filepath.Match, so `*`/`?`/`[...]` work) against
	// the FRONT of a resolved module id — a 1-component pattern matches any
	// module whose first component matches.
	//
	// This is the symmetric counterpart to ModulePatterns ("what IS a module"
	// vs "what is NOT"). It exists because the 2-component fallback turns every
	// unmatched top-level directory into a module, so date-stamped migration /
	// generated directories (e.g. "20240403", "20250417/update_store_plans")
	// pollute the topology and inflate Breadth (the Hill number over per-module
	// survival shares). Matching is deterministic (W-02) and config-driven —
	// no date/heuristic inference (W-01).
	//
	// Examples:
	//   "20*"        excludes "20240403" and "20250417/update_store_plans"
	//   "migrations" excludes "migrations" and "migrations/<anything>"
	//
	// Default empty — opt-in, so existing observations are unchanged until set.
	ExcludeModules []string `yaml:"exclude_modules"`
	// RepoOverrides keys per-repo module pattern lists by the repo's
	// full identifier (matching whatever the caller threads down — for
	// the CLI today this is filepath.Base(repoPath); for SaaS callers
	// it's typically "owner/name"). A non-empty override REPLACES
	// ModulePatterns for that repo. An empty / missing override uses
	// the org-level ModulePatterns.
	RepoOverrides map[string]RepoConfig `yaml:"repo_overrides"`
	// MaxBlameFileBytes is the upper bound (in bytes) for files passed to
	// git blame. Files whose blob size exceeds this are skipped before the
	// blame call, which protects against pathological huge single-line
	// dumps (e.g. SQL bulk inserts) that would otherwise stall git blame
	// for minutes per file. <= 0 disables the filter.
	MaxBlameFileBytes int64 `yaml:"max_blame_file_bytes"`
	// UntestedSurvivalWeight multiplies the survival contribution of blame lines
	// whose source file is not guarded by a test. 1.0 disables the weighting and
	// matches pre-v2 behaviour; 0.5 (default) treats untested code as half-value.
	UntestedSurvivalWeight float64 `yaml:"untested_survival_weight"`
	// CommentFilter controls whether ParseLog requests full patches (`git log -p`)
	// to exclude comment-only / blank diff lines from Production / Design / Debt
	// insertion counts (gaming protection). Default true. Setting it false makes
	// ParseLog use `--numstat` only — it skips generating and parsing the full
	// patch stream, which is the dominant cost on large repos (it is ~24x less
	// output and roughly halves git's own work), at the price of counts that
	// include comment/blank lines. Gravity is essentially unaffected (its shape
	// comes from Design/Breadth/Indispensability and its gates from blame-derived
	// survival + catalysis, not from raw insertion counts), so the fast path is
	// well suited to timeline / large-repo runs where speed matters most.
	CommentFilter *bool `yaml:"comment_filter"`
	// RespectGitattributes controls whether paths marked linguist-generated or
	// linguist-vendored in .gitattributes are excluded from blame + change-volume
	// metrics (so generated / vendored files can't inflate gravity). Default true.
	// Set false to consider every tracked file regardless of attributes.
	RespectGitattributes *bool `yaml:"respect_gitattributes"`
	// ModuleLivenessMinMonths is the module liveness gate threshold (ADR step
	// 2). A FALLBACK-derived module (the conservative 2-component default —
	// where date-dir / DML / dump pollution enters) earns module-hood only by
	// being touched in >= this many DISTINCT calendar months; otherwise it
	// folds into metric.PeripheralModule (observation preserved, not dropped —
	// D-07). Pattern-DECLARED modules never fold. <= 1 disables the gate.
	// Deterministic: the month set comes from commit.Date, not wall-clock
	// (W-02). Default 2.
	//
	// Global for now; a per-repo override (RepoConfig) is a future step.
	ModuleLivenessMinMonths int `yaml:"module_liveness_min_months"`
}

// TeamEntry defines a named team with its members and optional domain scope.
type TeamEntry struct {
	Domain  string   `yaml:"domain"`
	Members []string `yaml:"members"`
}

// RepoConfig holds per-repo overrides. Only fields relevant for repo-level
// overrides live here — anything missing falls back to the org-level Config.
type RepoConfig struct {
	// ModulePatterns, when non-empty, fully REPLACES Config.ModulePatterns
	// for this repo. An empty/nil list means "no override" — use the
	// org-level patterns (or DefaultModulePatterns if those are empty too).
	ModulePatterns []string `yaml:"module_patterns"`
	// ExcludeModules is ADDED to (not replacing) Config.ExcludeModules for
	// this repo — org-level excludes always apply; the repo adds its own.
	ExcludeModules []string `yaml:"exclude_modules"`
}

// DefaultModulePatterns is the built-in set of monorepo-convention glob
// patterns. Equivalent to the historical "convention dirs" hard-coded list,
// but expressed as glob patterns so users can extend / refine it.
//
// The two trailing anchors `**/domain/*` and `**/usecase/*` capture
// sub-domain modules at ANY depth: a `domain` or `usecase` directory marks a
// business-knowledge boundary, so the module identifier extends down to the
// immediate child of that directory wherever the monorepo nests it (e.g.
// "services/ace/backend/app/domain/star"). Infrastructure and presentation
// directories are deliberately NOT anchored — they carry framework/plumbing,
// not business knowledge. See docs/eis/module-recognition.md (in the orbit
// repo) for the rationale.
var DefaultModulePatterns = []string{
	"services/*",
	"packages/*",
	"apps/*",
	"modules/*",
	"libs/*",
	"**/domain/*",
	"**/usecase/*",
}

// PatternsForRepo returns the effective module-pattern list for a repo,
// honoring RepoOverrides first, then Config.ModulePatterns, then
// DefaultModulePatterns. The lookup key is whatever the caller threads
// in (CLI: repo base name; SaaS: typically "owner/name"). To explicitly
// suppress conventions for one repo, set its override ModulePatterns to
// a sentinel like ["-"] — the simple len==0 → defaults semantics means
// an empty list reverts to defaults, not to "no patterns at all".
func PatternsForRepo(cfg *Config, repoName string) []string {
	if cfg != nil {
		if ov, ok := cfg.RepoOverrides[repoName]; ok && len(ov.ModulePatterns) > 0 {
			return ov.ModulePatterns
		}
		if len(cfg.ModulePatterns) > 0 {
			return cfg.ModulePatterns
		}
	}
	return DefaultModulePatterns
}

// ExcludesForRepo returns the effective module-exclude list for a repo:
// org-level Config.ExcludeModules plus any per-repo additions (additive, not
// replacing — unlike module patterns). The lookup key matches PatternsForRepo.
func ExcludesForRepo(cfg *Config, repoName string) []string {
	if cfg == nil {
		return nil
	}
	out := append([]string(nil), cfg.ExcludeModules...)
	if ov, ok := cfg.RepoOverrides[repoName]; ok && len(ov.ExcludeModules) > 0 {
		out = append(out, ov.ExcludeModules...)
	}
	return out
}

// DomainsConfig maps domain names to their configuration.
// Supports both legacy format (list of repo patterns) and new format (object with repos + extensions).
//
//	legacy:  backend: [api, worker]
//	new:     mobile: { repos: [ios-app], extensions: [.swift, .kt] }
type DomainsConfig map[string]DomainEntry

// DomainEntry defines repo patterns and file extensions for a domain.
type DomainEntry struct {
	Repos      []string `yaml:"repos"`
	Extensions []string `yaml:"extensions"`
	// Tau overrides the global survival decay half-life (days) for this domain, so
	// a fast-churn domain (e.g. Frontend) can decay faster than a slow one (e.g.
	// Infra) — "survival" then measures durability relative to the domain's natural
	// turnover, not raw calendar time. nil → use the global Tau.
	Tau *float64 `yaml:"tau"`
}

// TauForDomain returns the survival decay half-life for a domain: the domain's
// override if set, else the global Tau. Both the config key and the queried label
// are run through domain.NormalizeName, so a config key "backend" (or "Backend")
// matches the resolved label "BE" — otherwise an override would silently never
// apply, since the pipeline queries with the normalized label.
func (c *Config) TauForDomain(domainName string) float64 {
	want := domain.NormalizeName(domainName)
	for name, entry := range c.Domains {
		if entry.Tau != nil && domain.NormalizeName(name) == want {
			return *entry.Tau
		}
	}
	return c.Tau
}

// UnmarshalYAML supports both legacy format (list of strings) and new format (object).
func (e *DomainEntry) UnmarshalYAML(value *yaml.Node) error {
	// Try list of strings first (legacy format: repo patterns only)
	var list []string
	if err := value.Decode(&list); err == nil {
		e.Repos = list
		return nil
	}
	// Object format
	type raw DomainEntry
	var r raw
	if err := value.Decode(&r); err != nil {
		return err
	}
	*e = DomainEntry(r)
	return nil
}

type Weights struct {
	Production       float64 `yaml:"production"`
	Catalysis        float64 `yaml:"catalysis"`
	Survival         float64 `yaml:"survival"`
	Design           float64 `yaml:"design"`
	Breadth          float64 `yaml:"breadth"`
	DebtCleanup      float64 `yaml:"debt_cleanup"`
	Indispensability float64 `yaml:"indispensability"`
}

type BusFactor struct {
	Critical float64 `yaml:"critical"`
	High     float64 `yaml:"high"`
}

// CommentFilterEnabled reports whether ParseLog should request full patches to
// strip comment/blank lines. It defaults to true; only an explicit
// `comment_filter: false` (or --no-comment-filter) turns on the fast numstat path.
// RespectGitattributesEnabled reports whether linguist-generated/vendored paths
// should be excluded. Defaults to true (nil → on), like CommentFilter.
func (c *Config) RespectGitattributesEnabled() bool {
	return c.RespectGitattributes == nil || *c.RespectGitattributes
}

func (c *Config) CommentFilterEnabled() bool {
	return c.CommentFilter == nil || *c.CommentFilter
}

func Default() *Config {
	return &Config{
		Tau:                     180,
		SampleSize:              500,
		DebtThreshold:           10,
		ActiveDays:              30,
		BlameTimeout:            120,
		BlameMoveDetection:      "file",
		MaxBlameFileBytes:       5 * 1024 * 1024,
		ProductionDailyRef:      1000,
		UntestedSurvivalWeight:  0.5,
		ModuleLivenessMinMonths: 2,
		ExcludeFilePatterns: []string{
			"package-lock.json",
			"yarn.lock",
			"pnpm-lock.yaml",
			"go.sum",
			"docs/swagger*",
			"docs/doc.go",
			"docs/openapi*",
			"*generated*",
			"mock_*",
			"*.gen.*",
		},
		// Architecture patterns use the segment-precise glob convention shared
		// with the module resolver: `*` matches exactly one path segment,
		// `**` matches any depth. Specific architectural roles (repository,
		// domainservice, router, middleware, di, core) use `**/` so they match
		// at any depth. Generic framework-vocabulary directories prone to route
		// collisions (stores, hooks, types) stay single-level `*/` so a state
		// store at "app/stores/" counts but a route like "app/(site)/stores/"
		// does not.
		ArchitecturePatterns: []string{
			"**/repository/*interface*",
			"**/domainservice/",
			"**/router.go",
			"**/middleware/",
			"**/di/*.go",
			"**/core/",
			"*/stores/",
			"*/hooks/",
			"*/types/",
		},
		BlameExtensions: []string{
			"*.go", "*.ts", "*.tsx", "*.py", "*.rs", "*.java", "*.rb",
			"*.c", "*.h", "*.cpp", "*.hpp", "*.cc", "*.S",
			"*.scala", "*.hs", "*.ml", "*.mli",
		},
		ExcludeAuthors: []string{"github-actions[bot]", "renovate[bot]", "dependabot[bot]"},
		Aliases:        map[string]string{},
		Weights: Weights{
			Production:       0.10,
			Catalysis:        0.15,
			Survival:         0.25,
			Design:           0.20,
			Breadth:          0.10,
			DebtCleanup:      0.15,
			Indispensability: 0.05,
		},
		BusFactor: BusFactor{
			Critical: 0.80,
			High:     0.60,
		},
	}
}

// Load reads a config file. If explicit is true (user passed --config),
// a missing file is an error. Otherwise, a missing file returns defaults.
func Load(path string, explicit bool) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && !explicit {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Tau <= 0 {
		return fmt.Errorf("tau must be positive, got %f", c.Tau)
	}
	if c.SampleSize <= 0 {
		return fmt.Errorf("sample_size must be positive, got %d", c.SampleSize)
	}
	if c.DebtThreshold < 0 {
		return fmt.Errorf("debt_threshold must be non-negative, got %d", c.DebtThreshold)
	}
	if c.UntestedSurvivalWeight < 0 || c.UntestedSurvivalWeight > 1.0 {
		return fmt.Errorf("untested_survival_weight must be within [0.0, 1.0], got %f", c.UntestedSurvivalWeight)
	}
	switch c.BlameMoveDetection {
	case "", "off", "file", "commit", "full":
	default:
		return fmt.Errorf("blame_move_detection must be one of off|file|commit|full, got %q", c.BlameMoveDetection)
	}
	for name, entry := range c.Domains {
		if entry.Tau != nil && *entry.Tau <= 0 {
			return fmt.Errorf("domains.%s.tau must be positive, got %f", name, *entry.Tau)
		}
	}

	w := c.Weights
	sum := w.Production + w.Catalysis + w.Survival + w.Design + w.Breadth + w.DebtCleanup + w.Indispensability
	if math.Abs(sum-1.0) > 0.01 {
		return fmt.Errorf("weights must sum to 1.0, got %f", sum)
	}

	if c.BusFactor.Critical <= c.BusFactor.High {
		return fmt.Errorf("bus_factor.critical (%f) must be greater than bus_factor.high (%f)", c.BusFactor.Critical, c.BusFactor.High)
	}

	return nil
}

// UseDefaultDomains returns whether built-in domain extension mappings should be used.
// Defaults to true when default_domains is not set in config.
func (c *Config) UseDefaultDomains() bool {
	if c.DefaultDomains == nil {
		return true
	}
	return *c.DefaultDomains
}

// CustomExtensions returns a map of domain names to their custom extension lists.
// Used with domain.BuildExtMap to create the merged extension-to-domain map.
func (c *Config) CustomExtensions() map[string][]string {
	m := make(map[string][]string)
	for name, entry := range c.Domains {
		if len(entry.Extensions) > 0 {
			m[name] = entry.Extensions
		}
	}
	return m
}

func (c *Config) ResolveAuthor(name string) string {
	if canonical, ok := c.Aliases[name]; ok {
		return canonical
	}
	return name
}

func (c *Config) IsExcludedRepo(repoName string) bool {
	for _, excluded := range c.ExcludeRepos {
		if repoName == excluded {
			return true
		}
	}
	return false
}

func (c *Config) IsExcludedAuthor(name string) bool {
	resolved := c.ResolveAuthor(name)
	for _, excluded := range c.ExcludeAuthors {
		if resolved == excluded {
			return true
		}
	}
	return false
}

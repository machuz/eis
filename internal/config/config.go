package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Tau                  float64           `yaml:"tau"`
	SampleSize           int               `yaml:"sample_size"`
	DebtThreshold        int               `yaml:"debt_threshold"`
	ExcludeFilePatterns  []string          `yaml:"exclude_file_patterns"`
	ArchitecturePatterns []string          `yaml:"architecture_patterns"`
	BlameExtensions      []string          `yaml:"blame_extensions"`
	ExcludeAuthors       []string          `yaml:"exclude_authors"`
	Aliases              map[string]string `yaml:"aliases"`
	Weights              Weights           `yaml:"weights"`
	BusFactor            BusFactor         `yaml:"bus_factor"`
}

type Weights struct {
	Production       float64 `yaml:"production"`
	Quality          float64 `yaml:"quality"`
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

func Default() *Config {
	return &Config{
		Tau:          180,
		SampleSize:   500,
		DebtThreshold: 10,
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
		ArchitecturePatterns: []string{
			"*/repository/*interface*",
			"*/domainservice/",
			"*/router.go",
			"*/middleware/",
			"di/*.go",
			"*/core/",
			"*/stores/",
			"*/hooks/",
			"*/types/",
		},
		BlameExtensions: []string{
			"*.go", "*.ts", "*.tsx", "*.py", "*.rs", "*.java", "*.rb",
			"*.c", "*.h", "*.cpp", "*.hpp", "*.cc", "*.S",
			"*.scala", "*.hs", "*.ml", "*.mli",
		},
		ExcludeAuthors:  []string{"github-actions[bot]", "renovate[bot]", "dependabot[bot]"},
		Aliases:         map[string]string{},
		Weights: Weights{
			Production:       0.15,
			Quality:          0.10,
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

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) ResolveAuthor(name string) string {
	if canonical, ok := c.Aliases[name]; ok {
		return canonical
	}
	return name
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

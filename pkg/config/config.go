package config

import "github.com/machuz/eis/v2/internal/config"

type Config = config.Config
type Weights = config.Weights
type BusFactor = config.BusFactor
type TeamEntry = config.TeamEntry
type DomainsConfig = config.DomainsConfig
type DomainEntry = config.DomainEntry
type RepoConfig = config.RepoConfig

var (
	Load                  = config.Load
	Default               = config.Default
	PatternsForRepo       = config.PatternsForRepo
	DefaultModulePatterns = config.DefaultModulePatterns
)

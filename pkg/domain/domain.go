package domain

import "github.com/machuz/eis/v2/internal/domain"

type Domain = domain.Domain

const (
	Backend  = domain.Backend
	Frontend = domain.Frontend
	Infra    = domain.Infra
	Firmware = domain.Firmware
	Unknown  = domain.Unknown
)

var (
	DetectFromFiles  = domain.DetectFromFiles
	MatchRepoPattern = domain.MatchRepoPattern
	NormalizeName    = domain.NormalizeName
	BuildExtMap      = domain.BuildExtMap
	SortDomains      = domain.SortDomains
)

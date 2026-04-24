package metric

import "github.com/machuz/eis/v2/internal/metric"

type RawScores = metric.RawScores
type SurvivalResult = metric.SurvivalResult
type ChangePressure = metric.ChangePressure
type DebtData = metric.DebtData
type ModuleRisk = metric.ModuleRisk
type VerboseFunc = metric.VerboseFunc

// Module Topology types
type ModulePair = metric.ModulePair
type CochangeResult = metric.CochangeResult
type ModuleOwnership = metric.ModuleOwnership

var (
	CalcProduction           = metric.CalcProduction
	CalcLines                = metric.CalcLines
	CalcQuality              = metric.CalcQuality
	CalcSurvival             = metric.CalcSurvival
	CalcSurvivalWithPressure = metric.CalcSurvivalWithPressure
	CalcDesign               = metric.CalcDesign
	CalcDebt                 = metric.CalcDebt
	CalcIndispensability     = metric.CalcIndispensability
	CalcChangePressure       = metric.CalcChangePressure
	GetFixCommits            = metric.GetFixCommits
	IsExcluded               = metric.IsExcluded
	ModuleOf                 = metric.ModuleOf
	NewRawScores             = metric.NewRawScores

	// Module Topology
	CalcCochange               = metric.CalcCochange
	CalcModuleSurvival         = metric.CalcModuleSurvival
	CalcOwnershipFragmentation = metric.CalcOwnershipFragmentation
)

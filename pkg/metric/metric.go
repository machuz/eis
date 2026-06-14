package metric

import "github.com/machuz/eis/v2/internal/metric"

type RawScores = metric.RawScores
type SurvivalResult = metric.SurvivalResult
type ChangePressure = metric.ChangePressure
type DebtData = metric.DebtData
type ModuleRisk = metric.ModuleRisk
type VerboseFunc = metric.VerboseFunc

// Module topology / resolution
type ModulePair = metric.ModulePair
type CochangeResult = metric.CochangeResult
type ModuleOwnership = metric.ModuleOwnership
type ModuleResolver = metric.ModuleResolver

var (
	CalcProduction           = metric.CalcProduction
	CalcLines                = metric.CalcLines
	CalcCatalysis            = metric.CalcCatalysis
	CalcSurvival             = metric.CalcSurvival
	CalcSurvivalWithPressure = metric.CalcSurvivalWithPressure
	CalcDesign               = metric.CalcDesign
	CalcDebt                 = metric.CalcDebt
	CalcIndispensability     = metric.CalcIndispensability
	CalcChangePressure       = metric.CalcChangePressure
	GetFixCommits            = metric.GetFixCommits
	IsExcluded               = metric.IsExcluded
	NewRawScores             = metric.NewRawScores

	// Module resolution (glob-pattern based)
	NewModuleResolver     = metric.NewModuleResolver
	DefaultModulePatterns = metric.DefaultModulePatterns

	// Breadth
	ComputeBreadth = metric.ComputeBreadth

	// Module Topology
	CalcCochange               = metric.CalcCochange
	CalcModuleSurvival         = metric.CalcModuleSurvival
	CalcModuleSurvivalByAuthor = metric.CalcModuleSurvivalByAuthor
	CalcOwnershipFragmentation = metric.CalcOwnershipFragmentation
)

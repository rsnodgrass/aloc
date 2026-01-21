package effort

import (
	"math"

	"github.com/modern-tooling/aloc/internal/model"
)

// COCOMOOptions contains parameters for COCOMO estimation
type COCOMOOptions struct {
	// CostPerMonth is the average monthly cost per engineer (default: 15000)
	CostPerMonth float64
	// Model type: "organic", "semi-detached", "embedded" (default: organic)
	Model string
}

// DefaultCOCOMOOptions returns default COCOMO parameters
func DefaultCOCOMOOptions() COCOMOOptions {
	return COCOMOOptions{
		CostPerMonth: 15000,
		Model:        "organic",
	}
}

// HumanEffortEstimate contains COCOMO calculation results
type HumanEffortEstimate struct {
	EffortPersonMonths float64 `json:"effort_person_months"`
	ScheduleMonths     float64 `json:"schedule_months"`
	TeamSize           float64 `json:"team_size"`
	EstimatedCost      float64 `json:"estimated_cost"`
	Model              string  `json:"model"`
	KLOC               float64 `json:"kloc"`
}

// COCOMO coefficients by model type
type cocomoCoeffs struct {
	a, b, c, d float64
}

var cocomoModels = map[string]cocomoCoeffs{
	"organic":       {a: 2.4, b: 1.05, c: 2.5, d: 0.38},
	"semi-detached": {a: 3.0, b: 1.12, c: 2.5, d: 0.35},
	"embedded":      {a: 3.6, b: 1.20, c: 2.5, d: 0.32},
}

// CalculateHumanEffort estimates development effort using COCOMO Basic model
// COCOMO Basic Organic:
//
//	Effort (PM) = a * (KLOC)^b
//	Time (months) = c * (Effort)^d
//	Team = Effort / Time
//	Cost = Effort * cost_per_month
func CalculateHumanEffort(loc int, opts COCOMOOptions) HumanEffortEstimate {
	if opts.CostPerMonth <= 0 {
		opts.CostPerMonth = 15000
	}
	if opts.Model == "" {
		opts.Model = "organic"
	}

	coeffs, ok := cocomoModels[opts.Model]
	if !ok {
		coeffs = cocomoModels["organic"]
		opts.Model = "organic"
	}

	kloc := float64(loc) / 1000.0
	if kloc < 0.001 {
		return HumanEffortEstimate{Model: "COCOMO Basic " + opts.Model, KLOC: kloc}
	}

	// Effort in person-months
	effort := coeffs.a * math.Pow(kloc, coeffs.b)

	// Schedule in months
	schedule := coeffs.c * math.Pow(effort, coeffs.d)

	// Team size
	team := effort / schedule

	// Cost
	cost := effort * opts.CostPerMonth

	return HumanEffortEstimate{
		EffortPersonMonths: effort,
		ScheduleMonths:     schedule,
		TeamSize:           team,
		EstimatedCost:      cost,
		Model:              "COCOMO Basic " + opts.Model,
		KLOC:               kloc,
	}
}

// CalculateConventionalTeam estimates Market Replacement cost for a conventional team.
// Returns a range (low-high) based on productivity and coordination variance.
func CalculateConventionalTeam(loc int, opts COCOMOOptions) *model.TeamEstimate {
	if opts.CostPerMonth <= 0 {
		opts.CostPerMonth = 15000
	}

	kloc := float64(loc) / 1000.0
	if kloc < 0.001 {
		return nil
	}

	// Use organic COCOMO as baseline
	coeffs := cocomoModels["organic"]

	// Optimistic: higher productivity, lower overhead (0.85x effort)
	effortLow := coeffs.a * math.Pow(kloc, coeffs.b) * 0.85
	scheduleLow := coeffs.c * math.Pow(effortLow, coeffs.d)
	teamLow := effortLow / scheduleLow
	costLow := effortLow * opts.CostPerMonth

	// Pessimistic: lower productivity, higher rework (1.30x effort)
	effortHigh := coeffs.a * math.Pow(kloc, coeffs.b) * 1.30
	scheduleHigh := coeffs.c * math.Pow(effortHigh, coeffs.d)
	teamHigh := effortHigh / scheduleHigh
	costHigh := effortHigh * opts.CostPerMonth

	return &model.TeamEstimate{
		Cost:       model.EstimateRange{Low: costLow, High: costHigh},
		ScheduleMo: model.EstimateRange{Low: scheduleLow, High: scheduleHigh},
		TeamSize:   model.EstimateRange{Low: teamLow, High: teamHigh},
		Model:      "Conventional Team (COCOMO-based)",
	}
}

// CalculateAgenticTeam estimates effort for an AI-native, agentic delivery model.
// Assumes parallel AI agent execution with human supervision/integration.
//
// Conservative assumptions (defensible to senior engineers):
//   - Humans still do final design & merge
//   - AI output requires review and correction
//   - Task decomposition is imperfect
//   - Not all work parallelizes
//   - Coordination overhead remains, just reduced
//
// What changes: parallelism (via agents), idle time, drafting cost, iteration cycles
// What does NOT: architecture decisions, final correctness, integration, human review
func CalculateAgenticTeam(loc int, opts COCOMOOptions) *model.TeamEstimate {
	if opts.CostPerMonth <= 0 {
		opts.CostPerMonth = 15000
	}

	kloc := float64(loc) / 1000.0
	if kloc < 0.001 {
		return nil
	}

	conv := CalculateConventionalTeam(loc, opts)
	if conv == nil {
		return nil
	}

	// Schedule: 40-60% of conventional time
	// Parallel execution helps early/mid phases; late-stage integration still serializes
	scheduleLow := conv.ScheduleMo.Low * 0.40
	scheduleHigh := conv.ScheduleMo.High * 0.60

	// Team: 33-50% of conventional size
	// Someone still owns infra, QA, release; human review is a real bottleneck
	// Assumes strong engineers, but not unicorns
	teamLow := conv.TeamSize.Low * 0.33
	teamHigh := conv.TeamSize.High * 0.50
	if teamLow < 2 {
		teamLow = 2 // minimum viable team
	}

	// Cost falls out from: fewer people × shorter time + AI tooling
	humanCostLow := teamLow * scheduleLow * opts.CostPerMonth
	humanCostHigh := teamHigh * scheduleHigh * opts.CostPerMonth

	// AI tooling: ~$2K-$10K/month for typical projects, higher for large codebases
	aiToolingLow := 2000.0
	aiToolingHigh := 10000.0
	if kloc > 100 {
		aiToolingLow = 5000.0
		aiToolingHigh = 20000.0
	}

	// Total cost = human labor + AI tooling over schedule duration
	costLow := humanCostLow + (aiToolingLow * scheduleLow)
	costHigh := humanCostHigh + (aiToolingHigh * scheduleHigh)

	return &model.TeamEstimate{
		Cost:        model.EstimateRange{Low: costLow, High: costHigh},
		ScheduleMo:  model.EstimateRange{Low: scheduleLow, High: scheduleHigh},
		TeamSize:    model.EstimateRange{Low: teamLow, High: teamHigh},
		AIToolingMo: model.EstimateRange{Low: aiToolingLow, High: aiToolingHigh},
		Model:       "AI-Native Team (Agentic/Parallel)",
	}
}

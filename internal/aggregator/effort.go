package aggregator

import (
	"fmt"

	"github.com/modern-tooling/aloc/internal/effort"
	"github.com/modern-tooling/aloc/internal/model"
)

// EffortOptions controls effort estimation
type EffortOptions struct {
	IncludeHuman      bool
	IncludeAI         bool
	AIModel           string  // "sonnet", "opus", "haiku"
	HumanCostPerMonth float64 // default 15000
}

// DefaultEffortOptions returns sensible defaults
func DefaultEffortOptions() EffortOptions {
	return EffortOptions{
		IncludeHuman:      true,
		IncludeAI:         true,
		AIModel:           "sonnet",
		HumanCostPerMonth: 15000,
	}
}

// ComputeEffortEstimates calculates human and AI effort estimates (legacy)
func ComputeEffortEstimates(totalLOC int, opts EffortOptions) *model.EffortEstimates {
	return ComputeEffortEstimatesWithLines(totalLOC, model.LineMetrics{Code: totalLOC}, opts)
}

// HybridReductionRates defines AI assistance savings per role
var HybridReductionRates = map[model.Role]struct {
	Reduction   float64
	Description string
}{
	model.RoleTest:  {0.30, "AI writes stubs → human reviews"},
	model.RoleDocs:  {0.50, "AI drafts → human edits"},
	model.RoleCore:  {0.20, "AI suggests → human implements"},
	model.RoleInfra: {0.20, "AI generates boilerplate"},
}

// ComputeEffortEstimatesWithLines calculates human and AI effort estimates using line metrics
func ComputeEffortEstimatesWithLines(totalLOC int, lines model.LineMetrics, opts EffortOptions) *model.EffortEstimates {
	if !opts.IncludeHuman && !opts.IncludeAI {
		return nil
	}

	estimates := &model.EffortEstimates{}

	// Human effort (COCOMO)
	if opts.IncludeHuman {
		cocomoOpts := effort.COCOMOOptions{
			CostPerMonth: opts.HumanCostPerMonth,
			Model:        "organic",
		}
		if cocomoOpts.CostPerMonth <= 0 {
			cocomoOpts.CostPerMonth = 15000
		}
		human := effort.CalculateHumanEffort(totalLOC, cocomoOpts)
		estimates.Human = &model.HumanEffort{
			EstimatedCost:      human.EstimatedCost,
			EffortPersonMonths: human.EffortPersonMonths,
			ScheduleMonths:     human.ScheduleMonths,
			TeamSize:           human.TeamSize,
			Model:              human.Model,
		}
	}

	// AI effort (implementation model - accounts for iterations)
	if opts.IncludeAI {
		aiModel := opts.AIModel
		if aiModel == "" {
			aiModel = "sonnet"
		}
		aiCost := effort.CalculateImplementationCost(totalLOC, aiModel)
		estimates.AI = &model.AIEffort{
			InputTokens:  aiCost.InputTokens,
			OutputTokens: aiCost.OutputTokens,
			InputCost:    aiCost.InputCost,
			OutputCost:   aiCost.OutputCost,
			TotalCost:    aiCost.TotalCost,
			Model:        aiCost.Model,
			Assumptions:  aiCost.Assumptions,
		}
	}

	// Compute comparison and hybrid savings if both estimates exist
	if estimates.Human != nil && estimates.AI != nil {
		humanCost := estimates.Human.EstimatedCost
		aiCost := estimates.AI.TotalCost

		// Calculate hybrid cost (weighted average savings)
		hybridSavings := 0.30 // default 30% savings with AI assistance
		hybridCost := humanCost * (1 - hybridSavings)

		// Calculate ratio
		ratio := 0.0
		if aiCost > 0 {
			ratio = humanCost / aiCost
		}

		estimates.Comparison = &model.CostComparison{
			AIOnly:        aiCost,
			HumanOnly:     humanCost,
			HybridCost:    hybridCost,
			HybridSavings: hybridSavings,
			Ratio:         ratio,
		}
	}

	return estimates
}

// ComputeEffortWithResponsibilities computes effort with per-role hybrid breakdown
func ComputeEffortWithResponsibilities(totalLOC int, lines model.LineMetrics, responsibilities []model.Responsibility, ratios model.Ratios, opts EffortOptions) *model.EffortEstimates {
	estimates := ComputeEffortEstimatesWithLines(totalLOC, lines, opts)
	if estimates == nil {
		return nil
	}

	// Compute delivery model estimates (ranges, not point estimates)
	cocomoOpts := effort.COCOMOOptions{
		CostPerMonth: opts.HumanCostPerMonth,
		Model:        "organic",
	}
	if cocomoOpts.CostPerMonth <= 0 {
		cocomoOpts.CostPerMonth = 15000
	}

	// Market Replacement Estimate (Conventional Team)
	estimates.Conventional = effort.CalculateConventionalTeam(totalLOC, cocomoOpts)

	// AI-Native Team Estimate (Agentic/Parallel)
	estimates.Agentic = effort.CalculateAgenticTeam(totalLOC, cocomoOpts)

	// Compute per-role hybrid savings
	if estimates.Human != nil {
		estimates.HybridBreakdown = ComputeHybridBreakdown(responsibilities, estimates.Human.EstimatedCost)
	}

	// Compute quick actions
	estimates.QuickActions = ComputeQuickActions(responsibilities, ratios, estimates.Human)

	// Compute elite operator reference (skill-band cost model)
	if estimates.Human != nil {
		eliteRef := effort.CalculateEliteOperatorReference(estimates.Human.EstimatedCost)
		estimates.EliteReference = &model.EliteOperatorReference{
			HybridCostLow:  eliteRef.HybridCostLow,
			HybridCostHigh: eliteRef.HybridCostHigh,
			VsMarketLow:    eliteRef.VsMarketLow,
			VsMarketHigh:   eliteRef.VsMarketHigh,
			Description:    eliteRef.Description,
		}
	}

	return estimates
}

// ComputeHybridBreakdown calculates per-role savings from AI assistance
func ComputeHybridBreakdown(responsibilities []model.Responsibility, totalHumanCost float64) []model.HybridSavings {
	var breakdown []model.HybridSavings
	var totalLOC int

	for _, r := range responsibilities {
		totalLOC += r.LOC
	}

	if totalLOC == 0 {
		return breakdown
	}

	for _, r := range responsibilities {
		rate, ok := HybridReductionRates[r.Role]
		if !ok || r.LOC == 0 {
			continue
		}

		// Proportion of total cost for this role
		roleProportion := float64(r.LOC) / float64(totalLOC)
		roleCost := totalHumanCost * roleProportion
		dollarsSaved := roleCost * rate.Reduction

		breakdown = append(breakdown, model.HybridSavings{
			Role:         r.Role,
			Reduction:    rate.Reduction,
			DollarsSaved: dollarsSaved,
			Description:  rate.Description,
		})
	}

	return breakdown
}

// ComputeQuickActions generates actionable recommendations based on ratios
func ComputeQuickActions(responsibilities []model.Responsibility, ratios model.Ratios, human *model.HumanEffort) []model.QuickAction {
	var actions []model.QuickAction

	// Find core LOC for gap calculations
	var coreLOC int
	for _, r := range responsibilities {
		if r.Role == model.RoleCore {
			coreLOC = r.LOC
			break
		}
	}

	if coreLOC == 0 {
		return actions
	}

	// Check test coverage
	if ratios.TestToCore < 0.5 {
		targetRatio := 0.5
		currentTestLOC := int(ratios.TestToCore * float32(coreLOC))
		targetTestLOC := int(targetRatio * float64(coreLOC))
		locGap := targetTestLOC - currentTestLOC

		// Estimate savings from AI-assisted test generation
		savings := 0.0
		if human != nil {
			// Cost per LOC * gap * 30% reduction
			costPerLOC := human.EstimatedCost / float64(coreLOC)
			savings = costPerLOC * float64(locGap) * 0.30
		}

		actions = append(actions, model.QuickAction{
			Priority:    1,
			Description: fmt.Sprintf("Add ~%d LOC of tests to reach 0.5 test/core ratio", locGap),
			Savings:     savings,
			LOCGap:      locGap,
		})
	}

	// Check documentation
	if ratios.DocsToCore < 0.2 {
		targetRatio := 0.2
		currentDocsLOC := int(ratios.DocsToCore * float32(coreLOC))
		targetDocsLOC := int(targetRatio * float64(coreLOC))
		locGap := targetDocsLOC - currentDocsLOC

		savings := 0.0
		if human != nil {
			costPerLOC := human.EstimatedCost / float64(coreLOC)
			savings = costPerLOC * float64(locGap) * 0.50 // 50% AI reduction for docs
		}

		actions = append(actions, model.QuickAction{
			Priority:    2,
			Description: fmt.Sprintf("AI-draft missing docs: saves ~$%.0fK at 50%% effort reduction", savings/1000),
			Savings:     savings,
			LOCGap:      locGap,
		})
	}

	return actions
}

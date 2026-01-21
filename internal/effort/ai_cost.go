package effort

import (
	"fmt"

	"github.com/modern-tooling/aloc/internal/model"
)

// AICostEstimate contains calculated AI costs
type AICostEstimate struct {
	InputCost    float64  `json:"input_cost"`
	OutputCost   float64  `json:"output_cost"`
	TotalCost    float64  `json:"total_cost"`
	Model        string   `json:"model"`
	InputTokens  int64    `json:"input_tokens"`
	OutputTokens int64    `json:"output_tokens"`
	Assumptions  []string `json:"assumptions"`
}

// CalculateAICost computes AI costs from token estimates and pricing
func CalculateAICost(tokens TokenEstimate, pricing model.ModelPricing) AICostEstimate {
	// Cost = tokens / 1_000_000 * cost_per_million
	inputCost := float64(tokens.InputTokens) / 1_000_000.0 * pricing.InputCostPerMTok
	outputCost := float64(tokens.OutputTokens) / 1_000_000.0 * pricing.OutputCostPerMTok
	totalCost := inputCost + outputCost

	assumptions := append([]string{}, tokens.Assumptions...)
	assumptions = append(assumptions,
		fmt.Sprintf("Input: $%.2f per million tokens", pricing.InputCostPerMTok),
		fmt.Sprintf("Output: $%.2f per million tokens", pricing.OutputCostPerMTok),
	)

	return AICostEstimate{
		InputCost:    inputCost,
		OutputCost:   outputCost,
		TotalCost:    totalCost,
		Model:        pricing.Name,
		InputTokens:  tokens.InputTokens,
		OutputTokens: tokens.OutputTokens,
		Assumptions:  assumptions,
	}
}

// CalculateAICostForLOC is a convenience function that estimates tokens and calculates cost
func CalculateAICostForLOC(loc int, modelName string) AICostEstimate {
	tokens := EstimateTokens(loc, DefaultTokenOptions())
	pricing := model.GetModelPricing(modelName)
	return CalculateAICost(tokens, pricing)
}

// CalculateAICostForLines estimates AI cost using line metrics (more accurate)
func CalculateAICostForLines(lines model.LineMetrics, modelName string) AICostEstimate {
	tokens := EstimateTokensFromLines(lines)
	pricing := model.GetModelPricing(modelName)
	return CalculateAICost(tokens, pricing)
}

// FormatCost returns a human-readable cost string
func FormatCost(cost float64) string {
	if cost >= 1000000 {
		return fmt.Sprintf("$%.2fM", cost/1000000)
	}
	if cost >= 1000 {
		return fmt.Sprintf("$%.2fK", cost/1000)
	}
	return fmt.Sprintf("$%.2f", cost)
}

// FormatTokens returns a human-readable token count
func FormatTokens(tokens int64) string {
	if tokens >= 1000000000 {
		return fmt.Sprintf("%.1fB", float64(tokens)/1000000000)
	}
	if tokens >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1000000)
	}
	if tokens >= 1000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1000)
	}
	return fmt.Sprintf("%d", tokens)
}

// CalculateImplementationCost computes realistic AI implementation costs
// This uses the corrected model that accounts for multiple iterations
func CalculateImplementationCost(loc int, modelName string) AICostEstimate {
	tokens := EstimateImplementationTokens(loc)
	pricing := model.GetModelPricing(modelName)

	inputCost := float64(tokens.InputTokens) / 1_000_000.0 * pricing.InputCostPerMTok
	outputCost := float64(tokens.OutputTokens) / 1_000_000.0 * pricing.OutputCostPerMTok
	totalCost := inputCost + outputCost

	assumptions := append([]string{}, tokens.Assumptions...)
	assumptions = append(assumptions,
		fmt.Sprintf("Model: %s", pricing.Name),
		fmt.Sprintf("Input: $%.2f/MTok, Output: $%.2f/MTok", pricing.InputCostPerMTok, pricing.OutputCostPerMTok),
	)

	return AICostEstimate{
		InputCost:    inputCost,
		OutputCost:   outputCost,
		TotalCost:    totalCost,
		Model:        pricing.Name,
		InputTokens:  tokens.InputTokens,
		OutputTokens: tokens.OutputTokens,
		Assumptions:  assumptions,
	}
}

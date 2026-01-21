package effort

import (
	"testing"

	"github.com/modern-tooling/aloc/internal/model"
)

func TestCalculateAICost(t *testing.T) {
	tokens := TokenEstimate{
		InputTokens:  4000000, // 4M
		OutputTokens: 1000000, // 1M
		TotalTokens:  5000000,
		Assumptions:  []string{"test assumption"},
	}

	pricing := model.ClaudeSonnet
	result := CalculateAICost(tokens, pricing)

	// Input: 4M tokens @ $3/MTok = $12
	expectedInput := 12.0
	if result.InputCost != expectedInput {
		t.Errorf("InputCost = %v, want %v", result.InputCost, expectedInput)
	}

	// Output: 1M tokens @ $15/MTok = $15
	expectedOutput := 15.0
	if result.OutputCost != expectedOutput {
		t.Errorf("OutputCost = %v, want %v", result.OutputCost, expectedOutput)
	}

	// Total: $27
	expectedTotal := 27.0
	if result.TotalCost != expectedTotal {
		t.Errorf("TotalCost = %v, want %v", result.TotalCost, expectedTotal)
	}

	if result.Model != pricing.Name {
		t.Errorf("Model = %v, want %v", result.Model, pricing.Name)
	}

	if result.InputTokens != tokens.InputTokens {
		t.Errorf("InputTokens = %v, want %v", result.InputTokens, tokens.InputTokens)
	}

	if result.OutputTokens != tokens.OutputTokens {
		t.Errorf("OutputTokens = %v, want %v", result.OutputTokens, tokens.OutputTokens)
	}

	// assumptions should include original plus pricing info
	if len(result.Assumptions) < 2 {
		t.Error("Assumptions should include original assumptions plus pricing info")
	}
}

func TestCalculateAICostDifferentModels(t *testing.T) {
	tokens := TokenEstimate{
		InputTokens:  1000000,
		OutputTokens: 500000,
		TotalTokens:  1500000,
		Assumptions:  []string{},
	}

	sonnetCost := CalculateAICost(tokens, model.ClaudeSonnet)
	opusCost := CalculateAICost(tokens, model.ClaudeOpus)
	haikuCost := CalculateAICost(tokens, model.ClaudeHaiku)

	// Opus should be most expensive, Haiku least
	if haikuCost.TotalCost >= sonnetCost.TotalCost {
		t.Error("Haiku should be cheaper than Sonnet")
	}
	if sonnetCost.TotalCost >= opusCost.TotalCost {
		t.Error("Sonnet should be cheaper than Opus")
	}
}

func TestCalculateAICostForLOC(t *testing.T) {
	result := CalculateAICostForLOC(10000, "sonnet")

	if result.TotalCost <= 0 {
		t.Error("TotalCost should be positive")
	}
	if result.Model != "Claude Sonnet" {
		t.Errorf("Model = %v, want 'Claude Sonnet'", result.Model)
	}
}

func TestFormatCost(t *testing.T) {
	tests := []struct {
		cost float64
		want string
	}{
		{0.50, "$0.50"},
		{5.00, "$5.00"},
		{999.99, "$999.99"},
		{1500.00, "$1.50K"},
		{10000.00, "$10.00K"},
		{2500000.00, "$2.50M"},
	}

	for _, tt := range tests {
		got := FormatCost(tt.cost)
		if got != tt.want {
			t.Errorf("FormatCost(%v) = %v, want %v", tt.cost, got, tt.want)
		}
	}
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		tokens int64
		want   string
	}{
		{500, "500"},
		{999, "999"},
		{1500, "1.5K"},
		{10000, "10.0K"},
		{2500000, "2.5M"},
		{1000000000, "1.0B"},
		{3500000000, "3.5B"},
	}

	for _, tt := range tests {
		got := FormatTokens(tt.tokens)
		if got != tt.want {
			t.Errorf("FormatTokens(%v) = %v, want %v", tt.tokens, got, tt.want)
		}
	}
}

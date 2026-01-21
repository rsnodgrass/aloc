package effort

import (
	"testing"

	"github.com/modern-tooling/aloc/internal/model"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name  string
		lines model.LineMetrics
	}{
		{"1k code lines", model.LineMetrics{Code: 1000, Comments: 200}},
		{"10k code lines", model.LineMetrics{Code: 10000, Comments: 2000}},
		{"heavy comments", model.LineMetrics{Code: 5000, Comments: 3000}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultTokenOptions()
			opts.Lines = tt.lines
			result := EstimateTokens(tt.lines.Code, opts)

			// Input tokens should be based on chars
			if result.InputTokens <= 0 {
				t.Error("InputTokens should be positive")
			}

			// Output tokens should be summary-dominated (capped)
			if result.OutputTokens < 500 || result.OutputTokens > 5000 {
				t.Errorf("OutputTokens = %d, should be between 500-5000", result.OutputTokens)
			}

			if result.TotalTokens != result.InputTokens+result.OutputTokens {
				t.Error("TotalTokens should equal InputTokens + OutputTokens")
			}

			if len(result.Assumptions) == 0 {
				t.Error("Assumptions should not be empty")
			}
		})
	}
}

func TestDefaultTokenOptions(t *testing.T) {
	opts := DefaultTokenOptions()

	if opts.CharsPerToken != 4.0 {
		t.Errorf("CharsPerToken = %v, want 4.0", opts.CharsPerToken)
	}
}

func TestEstimateTokensFromLines(t *testing.T) {
	lines := model.LineMetrics{
		Code:     1000,
		Comments: 200,
	}
	result := EstimateTokensFromLines(lines)

	// Expected: (1000*25 + 200*40) / 4 = (25000 + 8000) / 4 = 8250 input tokens
	expectedInput := int64((1000*25 + 200*40) / 4)
	if result.InputTokens != expectedInput {
		t.Errorf("InputTokens = %d, want %d", result.InputTokens, expectedInput)
	}
}

func TestEstimateOutputTokens(t *testing.T) {
	tests := []struct {
		loc  int
		want int64
	}{
		{0, 500},    // minimum
		{100, 500},  // minimum
		{1000, 500}, // minimum
		{50000, 5000}, // maximum
	}

	for _, tt := range tests {
		got := EstimateOutputTokens(tt.loc)
		if got != tt.want {
			t.Errorf("EstimateOutputTokens(%d) = %d, want %d", tt.loc, got, tt.want)
		}
	}
}

func TestEstimateInputTokens(t *testing.T) {
	tests := []struct {
		outputTokens int64
		ratio        float64
		want         int64
	}{
		{1000, 4.0, 4000},
		{1000, 3.0, 3000},
		{1000, 0, 4000},    // zero ratio should default to 4.0
		{1000, -1.0, 4000}, // negative ratio should default to 4.0
		{0, 4.0, 0},
	}

	for _, tt := range tests {
		got := EstimateInputTokens(tt.outputTokens, tt.ratio)
		if got != tt.want {
			t.Errorf("EstimateInputTokens(%d, %v) = %d, want %d",
				tt.outputTokens, tt.ratio, got, tt.want)
		}
	}
}

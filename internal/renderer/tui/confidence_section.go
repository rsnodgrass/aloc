package tui

import (
	"fmt"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderConfidenceLine renders classification confidence as a single line
func RenderConfidenceLine(confidence model.ConfidenceInfo, theme *renderer.Theme) string {
	level := "high"
	if confidence.AutoClassified < 0.5 {
		if confidence.Override > 0.3 {
			level = "medium (manual overrides)"
		} else {
			level = "low (heuristic-only)"
		}
	} else if confidence.AutoClassified < 0.8 {
		level = "medium"
	}

	return fmt.Sprintf("Classification confidence: %s\n", level)
}

// RenderConfidenceSection renders detailed classification confidence (for --explain mode)
func RenderConfidenceSection(confidence model.ConfidenceInfo, effort *model.EffortEstimates, theme *renderer.Theme) string {
	// Detailed version moved to --explain mode
	return RenderConfidenceLine(confidence, theme)
}

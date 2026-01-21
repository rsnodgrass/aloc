package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderScaleAndEffort renders the unified scale section (top of report)
// Note: Effort estimates moved to end of report for clarity
func RenderScaleAndEffort(report *model.Report, theme *renderer.Theme) string {
	var b strings.Builder

	b.WriteString(theme.PrimaryBold.Render("Codebase Scale") + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

	// Compact single-line format: "166.7K LOC (122.3K code, 17.6K comments) · 620 files · 12 languages"
	lines := report.Summary.Lines
	if lines.Total > 0 {
		fmt.Fprintf(&b, "%s LOC %s · %s files · %d languages\n",
			formatMagnitude(lines.Total),
			theme.Dim.Render(fmt.Sprintf("(%s code, %s comments)",
				formatMagnitude(lines.Code),
				formatMagnitude(lines.Comments))),
			formatNumber(report.Summary.Files),
			report.Summary.Languages)
	} else {
		fmt.Fprintf(&b, "%s LOC · %s files · %d languages\n",
			formatMagnitude(report.Summary.LOCTotal),
			formatNumber(report.Summary.Files),
			report.Summary.Languages)
	}

	return b.String()
}

// formatCurrencyCompact formats currency compactly
// Never shows cents, rounds up to nearest dollar, max 1 decimal place
func formatCurrencyCompact(amount float64) string {
	// Round up to nearest dollar
	amount = float64(int(amount + 0.5))

	if amount >= 1000000 {
		return fmt.Sprintf("$%.1fM", amount/1000000)
	}
	if amount >= 1000 {
		return fmt.Sprintf("$%.0fK", amount/1000)
	}
	return fmt.Sprintf("$%.0f", amount)
}

package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderExecutiveSignals renders the executive summary section with colors
func RenderExecutiveSignals(report *model.Report, theme *renderer.Theme) string {
	var b strings.Builder

	// Section header
	b.WriteString(theme.PrimaryBold.Render("Executive Signals") + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

	// Key metrics with accent color for numbers
	b.WriteString(fmt.Sprintf("%-30s %s\n", "Total Lines of Code",
		theme.Accent.Render(formatMagnitude(report.Summary.LOCTotal))))
	b.WriteString(fmt.Sprintf("%-30s %s\n", "Total Files",
		theme.Accent.Render(formatNumber(report.Summary.Files))))
	b.WriteString(fmt.Sprintf("%-30s %s\n", "Languages Detected",
		theme.Accent.Render(fmt.Sprintf("%d", report.Summary.Languages))))

	// Find dominant responsibility
	dominant := findDominantRole(report.Responsibilities)
	b.WriteString(fmt.Sprintf("%-30s %s\n", "Responsibility Dominance", dominant))

	b.WriteString("\n")

	// Line breakdown section
	if report.Summary.Lines.Total > 0 {
		b.WriteString(theme.PrimaryBold.Render("Line Breakdown") + "\n")
		b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

		lines := report.Summary.Lines
		total := float64(lines.Total)

		b.WriteString(fmt.Sprintf("%-30s %s\n", "Total Lines",
			theme.Accent.Render(formatNumber(lines.Total))))
		b.WriteString(fmt.Sprintf("  %-28s %s  %s\n", "Code",
			theme.Info.Render(formatNumber(lines.Code)),
			theme.Dim.Render(fmt.Sprintf("(%.1f%%)", float64(lines.Code)/total*100))))
		b.WriteString(fmt.Sprintf("  %-28s %s  %s\n", "Comments",
			theme.Secondary.Render(formatNumber(lines.Comments)),
			theme.Dim.Render(fmt.Sprintf("(%.1f%%)", float64(lines.Comments)/total*100))))
		b.WriteString(fmt.Sprintf("  %-28s %s  %s\n", "Blanks",
			theme.Dim.Render(formatNumber(lines.Blanks)),
			theme.Dim.Render(fmt.Sprintf("(%.1f%%)", float64(lines.Blanks)/total*100))))

		b.WriteString("\n")
	}

	// Structural warnings with warning color
	warnings := detectWarnings(report)
	if len(warnings) > 0 {
		b.WriteString(theme.WarningBold.Render("Structural Imbalances Detected") + "\n")
		for _, w := range warnings {
			b.WriteString(fmt.Sprintf("  %s %s\n",
				theme.Warning.Render("⚠"),
				w))
		}
		b.WriteString("\n")
	}

	// Classification confidence
	confLevel := "high"
	confStyle := theme.Success
	if report.Confidence.AutoClassified < 0.5 {
		confLevel = "low"
		confStyle = theme.Warning
	}
	b.WriteString(fmt.Sprintf("Classification Confidence: %s %s\n",
		confStyle.Render(confLevel),
		theme.Dim.Render(fmt.Sprintf("(%.0f%% auto, %.0f%% heuristic)",
			report.Confidence.AutoClassified*100,
			report.Confidence.Heuristic*100))))

	return b.String()
}

func findDominantRole(responsibilities []model.Responsibility) string {
	if len(responsibilities) == 0 {
		return "unknown"
	}

	var total int
	for _, r := range responsibilities {
		total += r.LOC
	}

	if total == 0 {
		return "unknown"
	}

	top := responsibilities[0]
	pct := float64(top.LOC) / float64(total) * 100

	if pct > 90 {
		return fmt.Sprintf("%s overwhelmingly dominant (%.1f%%)", top.Role, pct)
	}
	if pct > 70 {
		return fmt.Sprintf("%s strongly dominant (%.1f%%)", top.Role, pct)
	}
	if pct > 50 {
		return fmt.Sprintf("%s dominant (%.1f%%)", top.Role, pct)
	}
	return fmt.Sprintf("%s leads (%.1f%%)", top.Role, pct)
}

func detectWarnings(report *model.Report) []string {
	var warnings []string

	// Test coverage warning
	if report.Ratios.TestToCore < 0.1 {
		warnings = append(warnings, "Test coverage extremely low")
	} else if report.Ratios.TestToCore < 0.3 {
		warnings = append(warnings, "Test coverage below healthy baseline")
	}

	// Documentation warning
	if report.Ratios.DocsToCore < 0.05 {
		warnings = append(warnings, "Documentation surface very light")
	}

	// Generated code warning
	if report.Ratios.GeneratedToCore > 0.5 {
		warnings = append(warnings, "High proportion of generated code")
	}

	return warnings
}

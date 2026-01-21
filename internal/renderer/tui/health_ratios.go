package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RatioHealth represents the health assessment of a ratio
type RatioHealth struct {
	Symbol      string
	Description string
	IsGood      bool
	IsWarning   bool
}

// RenderHealthRatios renders ratios with interpretive health symbols
func RenderHealthRatios(ratios model.Ratios, theme *renderer.Theme) string {
	var b strings.Builder

	b.WriteString(theme.PrimaryBold.Render("Key Ratios & Health Signals") + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

	// Test/Core ratio
	testHealth := assessTestRatio(ratios.TestToCore)
	b.WriteString(renderRatioLine("Test / Core", ratios.TestToCore, testHealth, theme))

	// Docs/Core ratio
	docsHealth := assessDocsRatio(ratios.DocsToCore)
	b.WriteString(renderRatioLine("Docs / Core", ratios.DocsToCore, docsHealth, theme))

	// Infra/Core ratio
	infraHealth := assessInfraRatio(ratios.InfraToCore)
	b.WriteString(renderRatioLine("Infra / Core", ratios.InfraToCore, infraHealth, theme))

	// Config/Core ratio
	configHealth := assessConfigRatio(ratios.ConfigToCore)
	b.WriteString(renderRatioLine("Config / Core", ratios.ConfigToCore, configHealth, theme))

	// Generated/Core ratio
	genHealth := assessGeneratedRatio(ratios.GeneratedToCore)
	b.WriteString(renderRatioLine("Generated / Core", ratios.GeneratedToCore, genHealth, theme))

	return b.String()
}

// RenderHealthRatiosWithComments renders ratios including comment/code ratio
func RenderHealthRatiosWithComments(ratios model.Ratios, lines model.LineMetrics, theme *renderer.Theme) string {
	var b strings.Builder

	b.WriteString(theme.PrimaryBold.Render("Key Ratios & Health Signals") + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

	// Comment/Code ratio (first - institutional knowledge signal)
	if lines.Code > 0 {
		commentRatio := float32(lines.Comments) / float32(lines.Code)
		commentHealth := assessCommentRatio(commentRatio)
		b.WriteString(renderRatioLine("Comment / Code", commentRatio, commentHealth, theme))
	}

	// Test/Core ratio
	testHealth := assessTestRatio(ratios.TestToCore)
	b.WriteString(renderRatioLine("Test / Core", ratios.TestToCore, testHealth, theme))

	// Docs/Core ratio
	docsHealth := assessDocsRatio(ratios.DocsToCore)
	b.WriteString(renderRatioLine("Docs / Core", ratios.DocsToCore, docsHealth, theme))

	// Infra/Core ratio
	infraHealth := assessInfraRatio(ratios.InfraToCore)
	b.WriteString(renderRatioLine("Infra / Core", ratios.InfraToCore, infraHealth, theme))

	// Config/Core ratio
	configHealth := assessConfigRatio(ratios.ConfigToCore)
	b.WriteString(renderRatioLine("Config / Core", ratios.ConfigToCore, configHealth, theme))

	// Generated/Core ratio
	genHealth := assessGeneratedRatio(ratios.GeneratedToCore)
	b.WriteString(renderRatioLine("Generated / Core", ratios.GeneratedToCore, genHealth, theme))

	return b.String()
}

func renderRatioLine(label string, value float32, health RatioHealth, theme *renderer.Theme) string {
	var symbolStyle = theme.Dim
	if health.IsGood {
		symbolStyle = theme.Success
	} else if health.IsWarning {
		symbolStyle = theme.Warning
	}

	// Numbers not colored (Tufte) - only symbol gets color
	return fmt.Sprintf("%-24s %8.2f    %s %s\n",
		label,
		value,
		symbolStyle.Render(health.Symbol),
		health.Description)
}

// RenderHealthRatiosWithGauges renders ratios with visual range indicators
func RenderHealthRatiosWithGauges(ratios model.Ratios, lines model.LineMetrics, theme *renderer.Theme) string {
	var b strings.Builder

	b.WriteString(theme.PrimaryBold.Render("Health Ratios") + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

	// Test/Core with gauge
	testHealth := assessTestRatio(ratios.TestToCore)
	b.WriteString(renderRatioWithGauge("Test / Core", float64(ratios.TestToCore), 0.5, 0.8, testHealth, theme))

	// Comment/Code ratio with gauge
	if lines.Code > 0 {
		commentRatio := float32(lines.Comments) / float32(lines.Code)
		commentHealth := assessCommentRatio(commentRatio)
		b.WriteString(renderRatioWithGauge("Comment / Code", float64(commentRatio), 0.15, 0.35, commentHealth, theme))
	}

	// Docs/Core with gauge
	docsHealth := assessDocsRatio(ratios.DocsToCore)
	b.WriteString(renderRatioWithGauge("Docs / Core", float64(ratios.DocsToCore), 0.2, 0.5, docsHealth, theme))

	// Infra/Core with gauge (lower is better)
	infraHealth := assessInfraRatio(ratios.InfraToCore)
	b.WriteString(renderRatioWithGauge("Infra / Core", float64(ratios.InfraToCore), 0.0, 0.1, infraHealth, theme))

	// Config/Core with gauge (lower is better)
	configHealth := assessConfigRatio(ratios.ConfigToCore)
	b.WriteString(renderRatioWithGauge("Config / Core", float64(ratios.ConfigToCore), 0.0, 0.05, configHealth, theme))

	return b.String()
}

// renderRatioWithGauge renders a ratio with a visual range indicator
func renderRatioWithGauge(label string, value, targetMin, targetMax float64, health RatioHealth, theme *renderer.Theme) string {
	var b strings.Builder

	var symbolStyle = theme.Dim
	if health.IsGood {
		symbolStyle = theme.Success
	} else if health.IsWarning {
		symbolStyle = theme.Warning
	}

	// Render gauge bar (shorter for visual restraint)
	gauge := renderGauge(value, targetMin, targetMax, 18, theme)

	b.WriteString(fmt.Sprintf("  %-14s %5.2f  %s  %s %s\n",
		label,
		value,
		gauge,
		symbolStyle.Render(health.Symbol),
		health.Description))

	return b.String()
}

// renderGauge creates a visual gauge showing value position relative to target range
func renderGauge(value, targetMin, targetMax float64, width int, theme *renderer.Theme) string {
	// Scale: 0 to 2x targetMax (to show over-healthy values too)
	maxScale := targetMax * 2
	if maxScale < 1.0 {
		maxScale = 1.0
	}

	// Calculate positions
	minPos := int((targetMin / maxScale) * float64(width))
	maxPos := int((targetMax / maxScale) * float64(width))
	valuePos := int((value / maxScale) * float64(width))

	if valuePos > width-1 {
		valuePos = width - 1
	}
	if valuePos < 0 {
		valuePos = 0
	}

	// Build gauge string
	bar := make([]rune, width)
	for i := range bar {
		if i >= minPos && i <= maxPos {
			// Healthy range
			bar[i] = '▓'
		} else {
			bar[i] = '·'
		}
	}

	// Mark current value position
	if valuePos >= 0 && valuePos < width {
		bar[valuePos] = '│'
	}

	// Style the parts
	var result strings.Builder
	for i, c := range bar {
		if i >= minPos && i <= maxPos {
			result.WriteString(theme.Success.Render(string(c)))
		} else if i == valuePos {
			if value >= targetMin && value <= targetMax {
				result.WriteString(theme.Success.Render(string(c)))
			} else {
				result.WriteString(theme.Warning.Render(string(c)))
			}
		} else {
			result.WriteString(theme.Dim.Render(string(c)))
		}
	}

	return result.String()
}

func assessCommentRatio(ratio float32) RatioHealth {
	switch {
	case ratio >= 0.15 && ratio <= 0.35:
		return RatioHealth{"✓", "healthy explanation density", true, false}
	case ratio > 0.35:
		return RatioHealth{"◦", "heavily commented", false, false}
	case ratio >= 0.08:
		return RatioHealth{"◦", "moderate commentary", false, false}
	case ratio >= 0.03:
		return RatioHealth{"⚠", "sparse commentary", false, true}
	default:
		return RatioHealth{"⚠", "minimal institutional knowledge", false, true}
	}
}

func assessTestRatio(ratio float32) RatioHealth {
	switch {
	case ratio >= 0.5 && ratio <= 0.8:
		return RatioHealth{"✓", "healthy baseline", true, false}
	case ratio > 0.8:
		return RatioHealth{"✓", "comprehensive coverage", true, false}
	case ratio >= 0.3:
		return RatioHealth{"◦", "moderate coverage", false, false}
	case ratio >= 0.1:
		return RatioHealth{"⚠", "below healthy baseline", false, true}
	default:
		return RatioHealth{"⚠", "extremely low", false, true}
	}
}

func assessDocsRatio(ratio float32) RatioHealth {
	switch {
	case ratio >= 0.2:
		return RatioHealth{"✓", "well documented", true, false}
	case ratio >= 0.1:
		return RatioHealth{"◦", "adequate documentation", false, false}
	case ratio >= 0.05:
		return RatioHealth{"⚠", "sparse documentation", false, true}
	default:
		return RatioHealth{"⚠", "minimal documentation", false, true}
	}
}

func assessInfraRatio(ratio float32) RatioHealth {
	switch {
	case ratio > 0.2:
		return RatioHealth{"⚠", "high operational complexity", false, true}
	case ratio > 0.1:
		return RatioHealth{"◦", "moderate operational footprint", false, false}
	default:
		return RatioHealth{"✓", "low operational complexity", true, false}
	}
}

func assessGeneratedRatio(ratio float32) RatioHealth {
	switch {
	case ratio > 0.5:
		return RatioHealth{"⚠", "high automation reliance", false, true}
	case ratio > 0.2:
		return RatioHealth{"◦", "moderate generated code", false, false}
	default:
		return RatioHealth{"✓", "minimal automation footprint", true, false}
	}
}

func assessConfigRatio(ratio float32) RatioHealth {
	switch {
	case ratio > 0.15:
		return RatioHealth{"⚠", "large configuration surface", false, true}
	case ratio > 0.05:
		return RatioHealth{"◦", "moderate config surface", false, false}
	default:
		return RatioHealth{"✓", "manageable config surface", true, false}
	}
}

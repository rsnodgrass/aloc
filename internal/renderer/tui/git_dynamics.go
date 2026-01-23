package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/git"
	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// Sparkline layout constants (non-adaptive, keep consistent)
const (
	labelWidth   = 8  // "prod    "
	gapWidth     = 2  // spacing between label and sparkline
	minSparkline = 16 // anything smaller is useless
	maxSparkline = 60 // beyond this, returns diminish
)

// sparklineWidth computes the available sparkline width from terminal width
func sparklineWidth(termWidth int) int {
	available := termWidth - labelWidth - gapWidth

	if available < minSparkline {
		return minSparkline
	}
	if available > maxSparkline {
		return maxSparkline
	}
	return available
}

// RenderGitDynamics renders the complete git dynamics section in Tufte order:
// sparklines (evidence) → signals (interpretation) → effort adjustment (impact)
func RenderGitDynamics(gitMetrics *model.GitMetrics, theme *renderer.Theme, width int) string {
	if gitMetrics == nil {
		return ""
	}

	var sb strings.Builder

	// section header (consistent 80-char rule like other sections)
	sb.WriteString("\n")
	sb.WriteString(theme.PrimaryBold.Render("Codebase Dynamics (git)"))
	sb.WriteString("\n")
	sb.WriteString(theme.Dim.Render(strings.Repeat("─", 80)))
	sb.WriteString("\n")

	// 1. SPARKLINES FIRST (evidence before interpretation)
	sb.WriteString(renderSparklines(gitMetrics, theme, width))

	// 2. SIGNALS (compact interpretation)
	sb.WriteString(renderSignals(gitMetrics, theme))

	// 3. INTERPRETATION (bridges dynamics → cost)
	if gitMetrics.AnalysisNote != "" {
		sb.WriteString(theme.Secondary.Render("Interpretation"))
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "  %s\n\n", theme.Dim.Render(gitMetrics.AnalysisNote))
	}

	return sb.String()
}

// renderSparklines renders the churn sparklines with role-based coloring
func renderSparklines(gitMetrics *model.GitMetrics, theme *renderer.Theme, termWidth int) string {
	if len(gitMetrics.ChurnSeries) == 0 {
		return ""
	}

	var sb strings.Builder

	// sparkline header with clear x-axis orientation
	sb.WriteString(theme.Dim.Render(fmt.Sprintf("Churn over time (last %d months, weekly)", gitMetrics.WindowMonths)))
	sb.WriteString("\n")

	// compute target sparkline width
	targetWidth := sparklineWidth(termWidth)

	// render sparklines for each role (in consistent order)
	roles := []model.Role{model.RoleCore, model.RoleTest, model.RoleInfra}
	roleLabels := map[model.Role]string{
		model.RoleCore:  "core",
		model.RoleTest:  "test",
		model.RoleInfra: "infra",
	}

	for _, role := range roles {
		if sparkline, ok := gitMetrics.ChurnSeries[role]; ok {
			// pad label BEFORE styling (ANSI codes break width calculation)
			paddedLabel := fmt.Sprintf("%-*s", labelWidth, roleLabels[role])

			// render adaptive sparkline from raw values
			var glyphs string
			if len(sparkline.Values) > 0 {
				glyphs = git.RenderAdaptiveSparkline(sparkline.Values, targetWidth)
			} else {
				// fallback to pre-rendered glyphs (pad/truncate to fit)
				glyphs = padOrTruncSparkline(sparkline.Glyphs, targetWidth)
			}

			// color the sparkline by role
			fmt.Fprintf(&sb, "%s%s%s\n",
				theme.ForRole(role).Render(paddedLabel),
				strings.Repeat(" ", gapWidth),
				theme.ForRole(role).Render(glyphs),
			)
		}
	}
	sb.WriteString("\n")

	return sb.String()
}

// effortDirection represents the impact on effort estimation
type effortDirection int

const (
	effortNeutral  effortDirection = iota // ● no impact
	effortIncrease                        // ▲ increases effort
	effortDecrease                        // ▼ decreases effort
)

func (d effortDirection) icon() string {
	switch d {
	case effortIncrease:
		return "▼" // down = negative (increases effort/risk)
	case effortDecrease:
		return "▲" // up = positive (reduces effort/risk)
	default:
		return "●"
	}
}

// signalHint contains hint text and effort direction
type signalHint struct {
	text      string
	direction effortDirection
}

// renderSignals renders the grouped signals block with categories
func renderSignals(gitMetrics *model.GitMetrics, theme *renderer.Theme) string {
	var sb strings.Builder

	// Signals header
	sb.WriteString(theme.Secondary.Render("Signals"))
	sb.WriteString("\n")

	// Change Distribution
	sb.WriteString(theme.Dim.Render("Change distribution"))
	sb.WriteString("\n")
	renderSignal(&sb, theme, "Churn concentration",
		fmt.Sprintf("%.0f%% files → %.0f%% edits", gitMetrics.ChurnConcentration.FilePercent, gitMetrics.ChurnConcentration.EditPercent),
		churnConcentrationHint(gitMetrics.ChurnConcentration))
	renderSignal(&sb, theme, "Rewrite pressure",
		fmt.Sprintf("%.0f%% replace/delete heavy", gitMetrics.RewritePressure*100),
		rewritePressureHint(gitMetrics.RewritePressure))
	sb.WriteString("\n")

	// Stability
	sb.WriteString(theme.Dim.Render("Stability"))
	sb.WriteString("\n")
	renderSignal(&sb, theme, "Stable core",
		fmt.Sprintf("%.0f%% untouched (18+ mo)", gitMetrics.StableCore*100),
		stableCoreHint(gitMetrics.StableCore))
	renderSignal(&sb, theme, "Volatile surface",
		fmt.Sprintf("%.0f%% LOC changed ≥5×", gitMetrics.VolatileSurface*100),
		volatileSurfaceHint(gitMetrics.VolatileSurface))
	sb.WriteString("\n")

	// Coordination
	sb.WriteString(theme.Dim.Render("Coordination"))
	sb.WriteString("\n")
	renderSignal(&sb, theme, "Ownership",
		fmt.Sprintf("%.0f%% prod single-author", gitMetrics.OwnershipConcentration*100),
		ownershipHint(gitMetrics.OwnershipConcentration))
	renderSignal(&sb, theme, "Parallelism",
		gitMetrics.ParallelismSignal,
		parallelismHint(gitMetrics.ParallelismSignal))
	sb.WriteString("\n")

	return sb.String()
}

// renderSignal renders a single signal line with label, value, icon, and hint
func renderSignal(sb *strings.Builder, theme *renderer.Theme, label, value string, hint signalHint) {
	// pad raw strings BEFORE styling (ANSI codes break width calculation)
	paddedLabel := fmt.Sprintf("%-20s", label)
	paddedValue := fmt.Sprintf("%-32s", value)
	paddedHint := fmt.Sprintf("%-14s", hint.text)

	fmt.Fprintf(sb, "  %s %s %s %s\n",
		theme.Dim.Render(paddedLabel),
		paddedValue,
		hint.direction.icon(),
		theme.Dim.Render(paddedHint))
}

// Hint functions provide intuitive labels with effort direction icons

func churnConcentrationHint(stat model.GitChurnStat) signalHint {
	if stat.FilePercent < 20 && stat.EditPercent > 60 {
		return signalHint{"hotspots", effortIncrease}
	}
	return signalHint{"distributed", effortNeutral}
}

func rewritePressureHint(v float64) signalHint {
	if v > 0.40 {
		return signalHint{"heavy rework", effortIncrease}
	}
	if v > 0.25 {
		return signalHint{"rework", effortIncrease}
	}
	return signalHint{"incremental", effortNeutral}
}

func stableCoreHint(v float64) signalHint {
	if v > 0.50 {
		return signalHint{"mature", effortDecrease}
	}
	if v > 0.20 {
		return signalHint{"settling", effortNeutral}
	}
	return signalHint{"evolving", effortNeutral}
}

func volatileSurfaceHint(v float64) signalHint {
	if v > 0.20 {
		return signalHint{"turbulent", effortIncrease}
	}
	if v > 0.10 {
		return signalHint{"active", effortNeutral}
	}
	return signalHint{"stable", effortDecrease}
}

func ownershipHint(v float64) signalHint {
	if v > 0.50 {
		return signalHint{"bus factor", effortIncrease}
	}
	if v > 0.30 {
		return signalHint{"concentrated", effortIncrease}
	}
	return signalHint{"distributed", effortNeutral}
}

func parallelismHint(signal string) signalHint {
	switch signal {
	case "high":
		return signalHint{"team velocity", effortDecrease}
	case "moderate":
		return signalHint{"mixed", effortNeutral}
	default:
		return signalHint{"sequential", effortNeutral}
	}
}

// padOrTruncSparkline adjusts sparkline to target width (fallback for pre-rendered)
func padOrTruncSparkline(glyphs string, targetWidth int) string {
	runes := []rune(glyphs)
	if len(runes) >= targetWidth {
		return string(runes[len(runes)-targetWidth:]) // keep most recent
	}
	// pad with baseline glyph on left (older time)
	padding := strings.Repeat("▁", targetWidth-len(runes))
	return padding + glyphs
}

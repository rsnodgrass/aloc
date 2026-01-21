package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderDevelopmentCost renders the effort comparison section with delivery model estimates
func RenderDevelopmentCost(effort *model.EffortEstimates, theme *renderer.Theme) string {
	if effort == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(theme.PrimaryBold.Render("Development Effort Models") + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

	// Market Replacement Estimate (Conventional Team)
	if effort.Conventional != nil {
		conv := effort.Conventional
		b.WriteString(theme.Secondary.Render("Market Replacement Estimate (Conventional Team)") + "\n")
		fmt.Fprintf(&b, "  %s – %s · %.0f–%.0f months · ~%.0f–%.0f engineers\n",
			formatCurrencyCompact(conv.Cost.Low),
			formatCurrencyCompact(conv.Cost.High),
			conv.ScheduleMo.Low, conv.ScheduleMo.High,
			conv.TeamSize.Low, conv.TeamSize.High)
		b.WriteString("\n")
	}

	// AI-Native Team Estimate (Agentic/Parallel)
	if effort.Agentic != nil {
		ag := effort.Agentic
		b.WriteString(theme.Secondary.Render("AI-Native Team Estimate (Agentic/Parallel)") + "\n")
		fmt.Fprintf(&b, "  %s – %s · %.0f–%.0f months · ~%.0f–%.0f engineers\n",
			formatCurrencyCompact(ag.Cost.Low),
			formatCurrencyCompact(ag.Cost.High),
			ag.ScheduleMo.Low, ag.ScheduleMo.High,
			ag.TeamSize.Low, ag.TeamSize.High)
		// AI tooling: monthly rate and total cost on separate line
		if ag.AIToolingMo.High > 0 {
			aiTotalLow := ag.AIToolingMo.Low * ag.ScheduleMo.Low
			aiTotalHigh := ag.AIToolingMo.High * ag.ScheduleMo.High
			fmt.Fprintf(&b, "  %s\n",
				theme.Dim.Render(fmt.Sprintf("(AI tooling: %s–%s/mo, total %s–%s)",
					formatCurrencyCompact(ag.AIToolingMo.Low),
					formatCurrencyCompact(ag.AIToolingMo.High),
					formatCurrencyCompact(aiTotalLow),
					formatCurrencyCompact(aiTotalHigh))))
		}
	}

	// Disclaimer
	b.WriteString("\n")
	b.WriteString(theme.Dim.Render("* Rough estimates only, +/- depending on the effectiveness and experience of engineers") + "\n")

	return b.String()
}

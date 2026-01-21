package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderEffortModels renders the human and AI effort estimates section (scc-style)
func RenderEffortModels(effort *model.EffortEstimates, theme *renderer.Theme) string {
	if effort == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(theme.PrimaryBold.Render("Effort Models — Human & AI Cost Estimates") + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

	// Human effort section (COCOMO style, matching scc format)
	if effort.Human != nil {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%-42s %s\n",
			"Estimated Cost to Develop (organic)",
			theme.Accent.Render(formatCurrency(effort.Human.EstimatedCost))))
		b.WriteString(fmt.Sprintf("%-42s %s\n",
			"Estimated Schedule Effort (organic)",
			theme.Accent.Render(fmt.Sprintf("%.2f months", effort.Human.ScheduleMonths))))
		b.WriteString(fmt.Sprintf("%-42s %s\n",
			"Estimated People Required (organic)",
			theme.Accent.Render(fmt.Sprintf("%.2f", effort.Human.TeamSize))))
	}

	// AI effort section
	if effort.AI != nil {
		b.WriteString("\n")
		b.WriteString(theme.Info.Render(fmt.Sprintf("AI Effort (%s)", effort.AI.Model)) + "\n")
		b.WriteString(fmt.Sprintf("  %-28s %s → %s\n",
			"Input Tokens",
			theme.Dim.Render(formatTokenCount(effort.AI.InputTokens)),
			theme.Accent.Render(formatCurrency(effort.AI.InputCost))))
		b.WriteString(fmt.Sprintf("  %-28s %s → %s\n",
			"Output Tokens",
			theme.Dim.Render(formatTokenCount(effort.AI.OutputTokens)),
			theme.Accent.Render(formatCurrency(effort.AI.OutputCost))))
		b.WriteString(fmt.Sprintf("  %-28s        %s\n",
			"Total AI Cost",
			theme.AccentBold.Render(formatCurrency(effort.AI.TotalCost))))

		// Comparison ratio if both exist
		if effort.Human != nil && effort.Human.EstimatedCost > 0 {
			ratio := effort.AI.TotalCost / effort.Human.EstimatedCost * 100
			b.WriteString(fmt.Sprintf("\n  %s\n",
				theme.Dim.Render(fmt.Sprintf("(%.4f%% of human cost)", ratio))))
		}
	}

	return b.String()
}

func formatCurrency(amount float64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("$%.2fM", amount/1000000)
	}
	if amount >= 1000 {
		return fmt.Sprintf("$%.2fK", amount/1000)
	}
	return fmt.Sprintf("$%.2f", amount)
}

func formatTokenCount(tokens int64) string {
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

package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderQuickActions renders actionable recommendations
func RenderQuickActions(actions []model.QuickAction, theme *renderer.Theme) string {
	if len(actions) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString(theme.PrimaryBold.Render("Quick Actions") + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n\n")

	for _, action := range actions {
		arrow := theme.Success.Render("→")
		desc := action.Description

		// Add savings info if available
		var savingsStr string
		if action.Savings > 0 {
			if action.Savings >= 1000 {
				savingsStr = fmt.Sprintf(" (saves ~$%.0fK)", action.Savings/1000)
			} else {
				savingsStr = fmt.Sprintf(" (saves ~$%.0f)", action.Savings)
			}
			savingsStr = theme.Success.Render(savingsStr)
		}

		b.WriteString(fmt.Sprintf("  %s %s%s\n", arrow, desc, savingsStr))
	}

	b.WriteString("\n")
	return b.String()
}

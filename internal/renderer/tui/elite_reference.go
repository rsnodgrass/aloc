package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderEliteReference renders the elite operator reference (tasteful, no hype)
func RenderEliteReference(ref *model.EliteOperatorReference, theme *renderer.Theme) string {
	if ref == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(theme.Dim.Render("Elite Operator Reference (observed):") + "\n")
	b.WriteString(fmt.Sprintf("  ~%s – %s  %s\n",
		formatCurrencyCompact(ref.HybridCostLow),
		formatCurrencyCompact(ref.HybridCostHigh),
		theme.Dim.Render("("+ref.Description+")")))

	return b.String()
}

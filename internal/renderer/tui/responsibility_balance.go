package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderResponsibilityBalance renders role distribution as a clean inline summary
func RenderResponsibilityBalance(responsibilities []model.Responsibility, totalLOC int, theme *renderer.Theme) string {
	var b strings.Builder

	b.WriteString(theme.PrimaryBold.Render("Responsibility Balance") + "\n")

	if totalLOC == 0 || len(responsibilities) == 0 {
		b.WriteString("No data\n")
		return b.String()
	}

	// Sort by LOC descending
	sorted := make([]model.Responsibility, len(responsibilities))
	copy(sorted, responsibilities)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LOC > sorted[j].LOC
	})

	// Collect significant roles (>=1%) and sum tiny ones into "other"
	var parts []string
	var otherPct float64

	for _, r := range sorted {
		if r.LOC == 0 {
			continue
		}
		pct := float64(r.LOC) / float64(totalLOC) * 100
		if pct >= 1.0 {
			roleStyle := theme.ForRole(r.Role)
			parts = append(parts, roleStyle.Render(fmt.Sprintf("%s %.0f%%", r.Role, pct)))
		} else {
			otherPct += pct
		}
	}

	if otherPct >= 0.5 {
		parts = append(parts, theme.Dim.Render(fmt.Sprintf("other %.0f%%", otherPct)))
	}

	b.WriteString(strings.Join(parts, theme.Dim.Render(" · ")) + "\n")

	return b.String()
}

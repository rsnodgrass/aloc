package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderLanguageMatrix renders a heatmap of language × responsibility density
func RenderLanguageMatrix(languages []model.LanguageComp, theme *renderer.Theme) string {
	var b strings.Builder

	b.WriteString(theme.PrimaryBold.Render("Language × Responsibility Density (LOC)") + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

	if len(languages) == 0 {
		b.WriteString("No data\n")
		return b.String()
	}

	// Define columns (roles to show)
	roles := []model.Role{model.RoleCore, model.RoleTest, model.RoleInfra, model.RoleDocs, model.RoleConfig}

	// Header
	b.WriteString(fmt.Sprintf("%-14s", ""))
	for _, role := range roles {
		b.WriteString(theme.ForRole(role).Render(fmt.Sprintf("%-10s", role)))
	}
	b.WriteString("\n")

	// Find global max for density scaling
	globalMax := 0
	for _, lang := range languages {
		for _, loc := range lang.Responsibilities {
			if loc > globalMax {
				globalMax = loc
			}
		}
	}

	// Sort languages by total LOC
	sorted := make([]model.LanguageComp, len(languages))
	copy(sorted, languages)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LOCTotal > sorted[j].LOCTotal
	})

	// Limit to top 10 languages
	if len(sorted) > 10 {
		sorted = sorted[:10]
	}

	// Render each language row
	for _, lang := range sorted {
		b.WriteString(fmt.Sprintf("%-14s", truncate(lang.Language, 13)))
		for _, role := range roles {
			loc := lang.Responsibilities[role]
			density := renderDensityColored(loc, globalMax, role, theme)
			b.WriteString(fmt.Sprintf("%-10s", density))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n" + theme.Dim.Render("Legend: █ high  ▓ mid-high  ▒ mid  ░ low  · none") + "\n")

	return b.String()
}

// renderDensityColored returns a 5-character density indicator with role colors
func renderDensityColored(value, maxValue int, role model.Role, theme *renderer.Theme) string {
	if value == 0 {
		return theme.Dim.Render("·····")
	}
	if maxValue == 0 {
		return theme.Dim.Render("·····")
	}

	roleStyle := theme.ForRole(role)
	ratio := float64(value) / float64(maxValue)

	// 5 density levels
	switch {
	case ratio >= 0.8:
		return roleStyle.Render("█████")
	case ratio >= 0.5:
		return roleStyle.Render("███") + theme.Dim.Render("▓▓")
	case ratio >= 0.25:
		return roleStyle.Render("██") + theme.Dim.Render("▒▒▒")
	case ratio >= 0.1:
		return roleStyle.Render("█") + theme.Dim.Render("░░░░")
	default:
		return theme.Dim.Render("░····")
	}
}

// renderDensity returns a 5-character density indicator (legacy, non-colored)
func renderDensity(value, maxValue int) string {
	if value == 0 {
		return "·····"
	}
	if maxValue == 0 {
		return "·····"
	}

	ratio := float64(value) / float64(maxValue)

	// 5 density levels
	switch {
	case ratio >= 0.8:
		return "█████"
	case ratio >= 0.5:
		return "███▓▓"
	case ratio >= 0.25:
		return "██▒▒▒"
	case ratio >= 0.1:
		return "█░░░░"
	default:
		return "░····"
	}
}

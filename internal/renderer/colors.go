package renderer

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/modern-tooling/aloc/internal/model"
)

// AdaptiveColor provides light/dark mode aware colors
type AdaptiveColor struct {
	Light string
	Dark  string
}

// Color returns the appropriate color for the current terminal background
func (ac AdaptiveColor) Color() lipgloss.TerminalColor {
	// lipgloss auto-detects terminal background
	return lipgloss.AdaptiveColor{Light: ac.Light, Dark: ac.Dark}
}

// Semantic color palette (ox-cli style)
var (
	// Success - green tones for positive indicators
	ColorSuccess = AdaptiveColor{Light: "#4F6A48", Dark: "#7A8F78"}

	// Warning - amber tones for caution
	ColorWarning = AdaptiveColor{Light: "#B86830", Dark: "#C47A4A"}

	// Error - red tones for problems
	ColorError = AdaptiveColor{Light: "#A33636", Dark: "#E07070"}

	// Info - blue tones for informational content
	ColorInfo = AdaptiveColor{Light: "#4A7A9C", Dark: "#8FA8C8"}

	// Dim - gray for secondary/muted content
	ColorDim = AdaptiveColor{Light: "#666666", Dark: "#888888"}

	// Primary - sage green for main content
	ColorPrimary = AdaptiveColor{Light: "#4F6A48", Dark: "#7A8F78"}

	// Accent - copper for highlighting numbers/stats
	ColorAccent = AdaptiveColor{Light: "#B86830", Dark: "#C47A4A"}

	// Secondary - muted for less important items
	ColorSecondary = AdaptiveColor{Light: "#5A6A7A", Dark: "#8A9AAA"}
)

// Role-based colors (kept for backward compatibility)
var (
	// Core code - primary blue
	ColorCoreCode = lipgloss.AdaptiveColor{Light: "#4A7A9C", Dark: "#8FA8C8"}

	// Test code - safety green
	ColorTestCode = lipgloss.AdaptiveColor{Light: "#4F6A48", Dark: "#7A8F78"}

	// Infrastructure - operational orange
	ColorInfraCode = lipgloss.AdaptiveColor{Light: "#B86830", Dark: "#C47A4A"}

	// Documentation - knowledge purple
	ColorDocsCode = lipgloss.AdaptiveColor{Light: "#7A5A9C", Dark: "#A88AC8"}

	// Configuration - caution yellow
	ColorConfigCode = lipgloss.AdaptiveColor{Light: "#8A7830", Dark: "#B8A858"}

	// Generated - low emphasis gray
	ColorGeneratedCode = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#888888"}

	// Vendor - external muted
	ColorVendorCode = lipgloss.AdaptiveColor{Light: "#5A5A5A", Dark: "#7A7A7A"}

	// Deprecated - warning red
	ColorDeprecatedCode = lipgloss.AdaptiveColor{Light: "#A33636", Dark: "#E07070"}
)

// Theme holds styled text renderers for each semantic role
type Theme struct {
	// Semantic styles
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Error     lipgloss.Style
	Info      lipgloss.Style
	Dim       lipgloss.Style
	Primary   lipgloss.Style
	Accent    lipgloss.Style
	Secondary lipgloss.Style

	// Bold variants
	PrimaryBold lipgloss.Style
	AccentBold  lipgloss.Style
	SuccessBold lipgloss.Style
	WarningBold lipgloss.Style
	ErrorBold   lipgloss.Style

	// Magnitude styles (for LOC numbers) - gradient from dim to bold
	MagnitudeBytes lipgloss.Style // raw numbers, dimmest
	MagnitudeK     lipgloss.Style // thousands, normal
	MagnitudeM     lipgloss.Style // millions, bold
	MagnitudeG     lipgloss.Style // billions, bold + accent

	// Role-based styles (backward compat)
	Core        lipgloss.Style
	Test        lipgloss.Style
	Infra       lipgloss.Style
	Docs        lipgloss.Style
	Config      lipgloss.Style
	Generated   lipgloss.Style
	Vendor      lipgloss.Style
	Deprecated  lipgloss.Style

	// Legacy names for compatibility
	Safety      lipgloss.Style
	Operational lipgloss.Style
	Knowledge   lipgloss.Style
	Fragility   lipgloss.Style
	LowEmphasis lipgloss.Style
	External    lipgloss.Style

	HealthGood    lipgloss.Style
	HealthWarning lipgloss.Style
	HealthBad     lipgloss.Style

	noColor bool
}

// NewDefaultTheme creates a theme with semantic colors
func NewDefaultTheme() *Theme {
	t := &Theme{
		// Semantic styles
		Success:   lipgloss.NewStyle().Foreground(ColorSuccess.Color()),
		Warning:   lipgloss.NewStyle().Foreground(ColorWarning.Color()),
		Error:     lipgloss.NewStyle().Foreground(ColorError.Color()),
		Info:      lipgloss.NewStyle().Foreground(ColorInfo.Color()),
		Dim:       lipgloss.NewStyle().Foreground(ColorDim.Color()),
		Primary:   lipgloss.NewStyle().Foreground(ColorPrimary.Color()),
		Accent:    lipgloss.NewStyle().Foreground(ColorAccent.Color()),
		Secondary: lipgloss.NewStyle().Foreground(ColorSecondary.Color()),

		// Bold variants
		PrimaryBold: lipgloss.NewStyle().Foreground(ColorPrimary.Color()).Bold(true),
		AccentBold:  lipgloss.NewStyle().Foreground(ColorAccent.Color()).Bold(true),
		SuccessBold: lipgloss.NewStyle().Foreground(ColorSuccess.Color()).Bold(true),
		WarningBold: lipgloss.NewStyle().Foreground(ColorWarning.Color()).Bold(true),
		ErrorBold:   lipgloss.NewStyle().Foreground(ColorError.Color()).Bold(true),

		// Magnitude styles - gradient from dim to bold+accent
		MagnitudeBytes: lipgloss.NewStyle().Foreground(ColorDim.Color()),
		MagnitudeK:     lipgloss.NewStyle().Foreground(ColorSecondary.Color()),
		MagnitudeM:     lipgloss.NewStyle().Foreground(ColorPrimary.Color()).Bold(true),
		MagnitudeG:     lipgloss.NewStyle().Foreground(ColorAccent.Color()).Bold(true),

		// Role-based styles
		Core:       lipgloss.NewStyle().Foreground(ColorCoreCode),
		Test:       lipgloss.NewStyle().Foreground(ColorTestCode),
		Infra:      lipgloss.NewStyle().Foreground(ColorInfraCode),
		Docs:       lipgloss.NewStyle().Foreground(ColorDocsCode),
		Config:     lipgloss.NewStyle().Foreground(ColorConfigCode),
		Generated:  lipgloss.NewStyle().Foreground(ColorGeneratedCode),
		Vendor:     lipgloss.NewStyle().Foreground(ColorVendorCode),
		Deprecated: lipgloss.NewStyle().Foreground(ColorDeprecatedCode),

		// Health indicators
		HealthGood:    lipgloss.NewStyle().Foreground(ColorSuccess.Color()),
		HealthWarning: lipgloss.NewStyle().Foreground(ColorWarning.Color()),
		HealthBad:     lipgloss.NewStyle().Foreground(ColorError.Color()),

		noColor: false,
	}

	// Alias legacy names
	t.Safety = t.Test
	t.Operational = t.Infra
	t.Knowledge = t.Docs
	t.Fragility = t.Config
	t.LowEmphasis = t.Generated
	t.External = t.Vendor

	return t
}

// NewNoColorTheme creates a theme without colors
func NewNoColorTheme() *Theme {
	plain := lipgloss.NewStyle()
	bold := lipgloss.NewStyle().Bold(true)

	return &Theme{
		// Semantic styles
		Success:   plain,
		Warning:   plain,
		Error:     plain,
		Info:      plain,
		Dim:       plain,
		Primary:   plain,
		Accent:    plain,
		Secondary: plain,

		// Bold variants (keep bold even without color)
		PrimaryBold: bold,
		AccentBold:  bold,
		SuccessBold: bold,
		WarningBold: bold,
		ErrorBold:   bold,

		// Magnitude styles (M and G get bold even without color)
		MagnitudeBytes: plain,
		MagnitudeK:     plain,
		MagnitudeM:     bold,
		MagnitudeG:     bold,

		// Role-based styles
		Core:       plain,
		Test:       plain,
		Infra:      plain,
		Docs:       plain,
		Config:     plain,
		Generated:  plain,
		Vendor:     plain,
		Deprecated: plain,

		// Legacy names
		Safety:      plain,
		Operational: plain,
		Knowledge:   plain,
		Fragility:   plain,
		LowEmphasis: plain,
		External:    plain,

		HealthGood:    plain,
		HealthWarning: plain,
		HealthBad:     plain,

		noColor: true,
	}
}

// ForRole returns the style for a semantic role
func (t *Theme) ForRole(role model.Role) lipgloss.Style {
	switch role {
	case model.RoleCore:
		return t.Core
	case model.RoleTest:
		return t.Test
	case model.RoleInfra:
		return t.Infra
	case model.RoleDocs, model.RoleExamples:
		return t.Docs
	case model.RoleConfig:
		return t.Config
	case model.RoleGenerated:
		return t.Generated
	case model.RoleVendor:
		return t.Vendor
	case model.RoleDeprecated:
		return t.Deprecated
	case model.RoleScripts:
		return t.Core
	default:
		return t.Primary
	}
}

// IsNoColor returns true if colors are disabled
func (t *Theme) IsNoColor() bool {
	return t.noColor
}

// ShouldDisableColor checks environment for color support
func ShouldDisableColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return true
	}
	if os.Getenv("TERM") == "dumb" {
		return true
	}
	return false
}

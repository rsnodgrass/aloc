package tui

import (
	"io"
	"os"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
	"golang.org/x/term"
)

type TUIRenderer struct {
	theme      *renderer.Theme
	width      int
	writer     io.Writer
	noColor    bool
	noEmbedded bool
}

func NewTUIRenderer(opts renderer.Options) *TUIRenderer {
	width := opts.Width
	if width == 0 {
		width = detectWidth()
	}

	var theme *renderer.Theme
	if opts.NoColor {
		theme = renderer.NewNoColorTheme()
	} else {
		theme = renderer.NewDefaultTheme()
	}

	return &TUIRenderer{
		theme:      theme,
		width:      width,
		writer:     opts.Writer,
		noColor:    opts.NoColor,
		noEmbedded: opts.NoEmbedded,
	}
}

func (r *TUIRenderer) Render(report *model.Report) error {
	var sections []string

	// 1. Scale (the answer - facts)
	sections = append(sections, RenderScaleAndEffort(report, r.theme))

	// 2. Responsibility Balance (role distribution)
	sections = append(sections, RenderResponsibilityBalance(report.Responsibilities, report.Summary.LOCTotal, r.theme))

	// 3. Language Breakdown (supporting evidence)
	if len(report.Languages) > 0 {
		sections = append(sections, RenderLanguageLedger(report.Languages, r.theme, r.noEmbedded))
	}

	// 4. Health Ratios (interpretive layer - ratios comparing roles)
	sections = append(sections, RenderHealthRatiosWithGauges(report.Ratios, report.Summary.Lines, r.theme))

	// 5. Effort Comparison (economics - last)
	if report.Effort != nil && report.Effort.Comparison != nil {
		sections = append(sections, RenderDevelopmentCost(report.Effort, r.theme))
	}

	output := strings.Join(sections, "\n")
	_, err := r.writer.Write([]byte(output + "\n"))
	return err
}

func detectWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width == 0 {
		return 80
	}
	if width > 100 {
		return 100
	}
	return width
}

package renderer

import (
	"io"
	"os"

	"github.com/modern-tooling/aloc/internal/model"
)

type Renderer interface {
	Render(report *model.Report) error
}

type Options struct {
	Writer     io.Writer
	NoColor    bool
	Compact    bool
	Width      int
	Pretty     bool
	NoEmbedded bool // hide embedded code blocks in Markdown
}

func DefaultOptions() Options {
	return Options{
		Writer:  os.Stdout,
		NoColor: ShouldDisableColor(),
		Compact: false,
		Width:   0,
		Pretty:  false,
	}
}

package scanner

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/modern-tooling/aloc/internal/model"
)

type Scanner struct {
	walker *Walker
}

type Options struct {
	NumWorkers int
	Exclude    []string
	DeepMode   bool
}

func NewScanner(root string, opts Options) (*Scanner, error) {
	walker, err := NewWalker(root, WalkOptions{
		NumWorkers: opts.NumWorkers,
		Exclude:    opts.Exclude,
		DeepMode:   opts.DeepMode,
	})
	if err != nil {
		return nil, err
	}
	return &Scanner{walker: walker}, nil
}

func (s *Scanner) Scan(ctx context.Context) (<-chan *model.RawFile, <-chan error) {
	// large buffers for streaming performance
	results := make(chan *model.RawFile, 8192)
	errs := make(chan error, 256)

	paths, walkErrs := s.walker.Walk(ctx)

	var wg sync.WaitGroup
	for i := 0; i < s.walker.numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				select {
				case <-ctx.Done():
					return
				default:
				}

				info, statErr := os.Stat(path)
				if statErr != nil {
					errs <- statErr
					continue
				}

				lang := DetectLanguage(path)

				// Use embedded-aware counting for Markdown/MDX
				var lines model.LineMetrics
				var embedded map[string]model.LineMetrics
				var countErr error

				if lang == "Markdown" || lang == "MDX" {
					lines, embedded, countErr = CountLinesWithEmbedded(path)
				} else {
					lines, countErr = CountLines(path)
				}
				if countErr != nil {
					errs <- countErr
					continue
				}

				// use relative path for inference rules to work correctly
				relPath, relErr := filepath.Rel(s.walker.root, path)
				if relErr != nil {
					relPath = path
				}

				results <- &model.RawFile{
					Path:         relPath,
					Bytes:        info.Size(),
					LOC:          lines.Code,
					Lines:        lines,
					LanguageHint: lang,
					Embedded:     embedded,
				}
			}
		}()
	}

	go func() {
		for err := range walkErrs {
			errs <- err
		}
	}()

	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	return results, errs
}

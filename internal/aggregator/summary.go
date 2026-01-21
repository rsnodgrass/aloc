package aggregator

import "github.com/modern-tooling/aloc/internal/model"

func ComputeSummary(records []*model.FileRecord) model.Summary {
	langs := make(map[string]bool)
	var totalLOC int
	var lines model.LineMetrics

	for _, r := range records {
		totalLOC += r.LOC
		lines.Total += r.Lines.Total
		lines.Blanks += r.Lines.Blanks
		lines.Comments += r.Lines.Comments
		lines.Code += r.Lines.Code
		if r.Language != "" && r.Language != "unknown" {
			langs[r.Language] = true
		}
	}

	return model.Summary{
		Files:     len(records),
		LOCTotal:  totalLOC,
		Lines:     lines,
		Languages: len(langs),
	}
}

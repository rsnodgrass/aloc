package aggregator

import (
	"sort"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/scanner"
)

type langAccum struct {
	Language string
	LOCTotal int
	Files    int
	Code     int // code lines (line type)
	Comments int // comment lines (line type)
	Blanks   int // blank lines (line type)
	Tests    int // test LOC (role-based, subset of Code)
	Config   int // config LOC
	ByRole   map[model.Role]int
	Embedded map[string]model.LineMetrics // embedded code blocks (for Markdown, etc.)
}

func ComputeLanguageBreakdown(records []*model.FileRecord) []model.LanguageComp {
	byLang := make(map[string]*langAccum)

	for _, r := range records {
		if r.Language == "" || r.Language == "unknown" {
			continue
		}

		acc, ok := byLang[r.Language]
		if !ok {
			acc = &langAccum{
				Language: r.Language,
				ByRole:   make(map[model.Role]int),
				Embedded: make(map[string]model.LineMetrics),
			}
			byLang[r.Language] = acc
		}

		// Track line types (Code + Comments + Blanks = Total)
		acc.Code += r.Lines.Code
		acc.Comments += r.Lines.Comments
		acc.Blanks += r.Lines.Blanks
		acc.LOCTotal += r.Lines.Code + r.Lines.Comments + r.Lines.Blanks
		acc.Files++
		acc.ByRole[r.Role] += r.LOC

		// Track tests by role (subset of Code, shown separately)
		switch r.Role {
		case model.RoleTest:
			acc.Tests += r.Lines.Code // test code lines only
		case model.RoleConfig:
			acc.Config += r.LOC
		}

		// Accumulate embedded code block stats (for Markdown, etc.)
		for lang, metrics := range r.Embedded {
			existing := acc.Embedded[lang]
			existing.Total += metrics.Total
			existing.Code += metrics.Code
			existing.Comments += metrics.Comments
			existing.Blanks += metrics.Blanks
			acc.Embedded[lang] = existing
		}
	}

	var result []model.LanguageComp
	for _, acc := range byLang {
		// Filter to significant roles (>=10% of language total)
		significant := make(map[model.Role]int)
		for role, loc := range acc.ByRole {
			if float32(loc)/float32(acc.LOCTotal) >= 0.10 {
				significant[role] = loc
			}
		}

		// Only include embedded if non-empty
		var embedded map[string]model.LineMetrics
		if len(acc.Embedded) > 0 {
			embedded = acc.Embedded
		}

		result = append(result, model.LanguageComp{
			Language:         acc.Language,
			Category:         string(scanner.GetLanguageCategory(acc.Language)),
			LOCTotal:         acc.LOCTotal,
			Files:            acc.Files,
			Code:             acc.Code,
			Comments:         acc.Comments,
			Blanks:           acc.Blanks,
			Tests:            acc.Tests,
			Config:           acc.Config,
			Responsibilities: significant,
			Embedded:         embedded,
		})
	}

	// Sort by category order, then by LOC descending within category
	sort.Slice(result, func(i, j int) bool {
		catI := scanner.GetCategoryDisplayOrder(scanner.LanguageCategory(result[i].Category))
		catJ := scanner.GetCategoryDisplayOrder(scanner.LanguageCategory(result[j].Category))
		if catI != catJ {
			return catI < catJ
		}
		return result[i].LOCTotal > result[j].LOCTotal
	})

	return result
}

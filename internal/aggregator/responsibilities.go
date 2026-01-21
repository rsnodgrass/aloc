package aggregator

import (
	"sort"

	"github.com/modern-tooling/aloc/internal/model"
)

type roleAccum struct {
	Role          model.Role
	LOC           int
	Files         int
	ConfidenceSum float64
	SubRoleCounts map[model.TestKind]int
}

func ComputeResponsibilities(records []*model.FileRecord) []model.Responsibility {
	byRole := make(map[model.Role]*roleAccum)

	for _, r := range records {
		acc, ok := byRole[r.Role]
		if !ok {
			acc = &roleAccum{
				Role:          r.Role,
				SubRoleCounts: make(map[model.TestKind]int),
			}
			byRole[r.Role] = acc
		}
		acc.LOC += r.LOC
		acc.Files++
		acc.ConfidenceSum += float64(r.Confidence) * float64(r.LOC)

		if r.Role == model.RoleTest && r.SubRole != "" {
			acc.SubRoleCounts[r.SubRole] += r.LOC
		}
	}

	var result []model.Responsibility
	for _, acc := range byRole {
		var confidence float32
		if acc.LOC > 0 {
			confidence = float32(acc.ConfidenceSum / float64(acc.LOC))
		}

		resp := model.Responsibility{
			Role:       acc.Role,
			LOC:        acc.LOC,
			Files:      acc.Files,
			Confidence: confidence,
		}

		// Test breakdown
		if acc.Role == model.RoleTest && len(acc.SubRoleCounts) > 0 {
			resp.Breakdown = computeTestBreakdown(acc.SubRoleCounts, acc.LOC)
		}

		result = append(result, resp)
	}

	// Sort by LOC descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].LOC > result[j].LOC
	})

	return result
}

func computeTestBreakdown(counts map[model.TestKind]int, total int) map[model.TestKind]float32 {
	if total == 0 {
		return nil
	}
	breakdown := make(map[model.TestKind]float32)
	for kind, loc := range counts {
		breakdown[kind] = float32(loc) / float32(total)
	}
	return breakdown
}

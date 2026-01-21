package inference

import (
	"sort"

	"github.com/modern-tooling/aloc/internal/model"
)

type RoleScore struct {
	Weights  map[model.Role]float32
	Signals  map[model.Role][]model.Signal
	SubRoles map[model.Role]model.TestKind
}

func NewRoleScore() *RoleScore {
	return &RoleScore{
		Weights:  make(map[model.Role]float32),
		Signals:  make(map[model.Role][]model.Signal),
		SubRoles: make(map[model.Role]model.TestKind),
	}
}

func (s *RoleScore) Add(role model.Role, weight float32, signal model.Signal) {
	s.Weights[role] += weight
	s.Signals[role] = append(s.Signals[role], signal)
}

func (s *RoleScore) AddWithSubRole(role model.Role, subRole model.TestKind, weight float32, signal model.Signal) {
	s.Add(role, weight, signal)
	if subRole != "" {
		s.SubRoles[role] = subRole
	}
}

func (s *RoleScore) MaxWeight() float32 {
	var max float32
	for _, w := range s.Weights {
		if w > max {
			max = w
		}
	}
	return max
}

type rankedRole struct {
	Role   model.Role
	Weight float32
}

func (s *RoleScore) Resolve() (model.Role, model.TestKind, float32, []model.Signal) {
	if len(s.Weights) == 0 {
		return model.RoleCore, "", 0.30, nil
	}

	// Sort by weight descending
	ranked := make([]rankedRole, 0, len(s.Weights))
	for role, weight := range s.Weights {
		ranked = append(ranked, rankedRole{role, weight})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Weight == ranked[j].Weight {
			return rolePriority(ranked[i].Role) < rolePriority(ranked[j].Role)
		}
		return ranked[i].Weight > ranked[j].Weight
	})

	topRole := ranked[0].Role
	topWeight := ranked[0].Weight
	confidence := topWeight

	// Ambiguity penalty
	if len(ranked) > 1 {
		secondWeight := ranked[1].Weight
		if topWeight-secondWeight < 0.15 {
			confidence *= 0.8 // 20% penalty
		}
	}

	// Agreement bonus
	signals := s.Signals[topRole]
	agreementFactor := float32(len(signals)) * 0.25
	if agreementFactor > 1.0 {
		agreementFactor = 1.0
	}
	confidence *= agreementFactor
	if confidence > 1.0 {
		confidence = 1.0
	}

	// Get sub-role for test
	var subRole model.TestKind
	if topRole == model.RoleTest {
		subRole = s.SubRoles[topRole]
		if subRole == "" {
			subRole = model.TestUnit
		}
	}

	return topRole, subRole, confidence, signals
}

// rolePriority returns the tie-break priority (lower is higher priority)
func rolePriority(role model.Role) int {
	priorities := map[model.Role]int{
		model.RoleVendor:     1,
		model.RoleGenerated:  2,
		model.RoleTest:       3,
		model.RoleInfra:      4,
		model.RoleCore:       5,
		model.RoleDocs:       6,
		model.RoleConfig:     7,
		model.RoleScripts:    8,
		model.RoleExamples:   9,
		model.RoleDeprecated: 10,
	}
	if p, ok := priorities[role]; ok {
		return p
	}
	return 100
}

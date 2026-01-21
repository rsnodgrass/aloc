package inference

import (
	"testing"

	"github.com/modern-tooling/aloc/internal/model"
)

func TestRoleScoreAdd(t *testing.T) {
	score := NewRoleScore()
	score.Add(model.RoleTest, 0.75, model.SignalFilename)
	score.Add(model.RoleTest, 0.60, model.SignalPath)

	if score.Weights[model.RoleTest] != 1.35 {
		t.Errorf("Weight = %v, want 1.35", score.Weights[model.RoleTest])
	}

	if len(score.Signals[model.RoleTest]) != 2 {
		t.Errorf("Signals count = %v, want 2", len(score.Signals[model.RoleTest]))
	}
}

func TestRoleScoreAddWithSubRole(t *testing.T) {
	score := NewRoleScore()
	score.AddWithSubRole(model.RoleTest, model.TestIntegration, 0.80, model.SignalFilename)

	if score.Weights[model.RoleTest] != 0.80 {
		t.Errorf("Weight = %v, want 0.80", score.Weights[model.RoleTest])
	}

	if score.SubRoles[model.RoleTest] != model.TestIntegration {
		t.Errorf("SubRole = %v, want integration", score.SubRoles[model.RoleTest])
	}
}

func TestRoleScoreAddWithSubRole_EmptySubRole(t *testing.T) {
	score := NewRoleScore()
	score.AddWithSubRole(model.RoleInfra, "", 0.80, model.SignalFilename)

	if _, exists := score.SubRoles[model.RoleInfra]; exists {
		t.Error("SubRole should not be set for empty SubRole")
	}
}

func TestRoleScoreResolve_SingleRole(t *testing.T) {
	score := NewRoleScore()
	score.Add(model.RoleTest, 0.75, model.SignalFilename)
	score.Add(model.RoleTest, 0.60, model.SignalPath)

	role, _, confidence, signals := score.Resolve()

	if role != model.RoleTest {
		t.Errorf("Role = %v, want test", role)
	}
	if len(signals) != 2 {
		t.Errorf("Signals count = %v, want 2", len(signals))
	}
	if confidence <= 0 {
		t.Error("Confidence should be positive")
	}
}

func TestRoleScoreResolve_AmbiguityPenalty(t *testing.T) {
	score := NewRoleScore()
	score.Add(model.RoleTest, 0.75, model.SignalFilename)
	score.Add(model.RoleCore, 0.70, model.SignalPath)

	_, _, confidence, _ := score.Resolve()

	// Ambiguity penalty: top (0.75) - second (0.70) = 0.05 < 0.15 threshold
	// confidence = 0.75 * 0.8 (penalty) * 0.25 (1 signal agreement) = 0.15
	if confidence >= 0.75 {
		t.Errorf("Confidence = %v, should be reduced due to ambiguity", confidence)
	}
}

func TestRoleScoreResolve_EmptyScore(t *testing.T) {
	score := NewRoleScore()
	role, _, confidence, signals := score.Resolve()

	if role != model.RoleCore {
		t.Errorf("Role = %v, want core (default)", role)
	}
	if confidence != 0.30 {
		t.Errorf("Confidence = %v, want 0.30", confidence)
	}
	if signals != nil {
		t.Errorf("Signals = %v, want nil", signals)
	}
}

func TestRoleScoreResolve_TieBreak(t *testing.T) {
	score := NewRoleScore()
	score.Add(model.RoleVendor, 0.50, model.SignalPath)
	score.Add(model.RoleCore, 0.50, model.SignalPath)

	role, _, _, _ := score.Resolve()

	// Vendor has priority 1, Prod has priority 5
	if role != model.RoleVendor {
		t.Errorf("Role = %v, want vendor (tie-break winner)", role)
	}
}

func TestRoleScoreResolve_TieBreak_GeneratedVsTest(t *testing.T) {
	score := NewRoleScore()
	score.Add(model.RoleGenerated, 0.50, model.SignalPath)
	score.Add(model.RoleTest, 0.50, model.SignalPath)

	role, _, _, _ := score.Resolve()

	// Generated has priority 2, Test has priority 3
	if role != model.RoleGenerated {
		t.Errorf("Role = %v, want generated (tie-break winner)", role)
	}
}

func TestRoleScoreResolve_TestSubRole(t *testing.T) {
	score := NewRoleScore()
	score.AddWithSubRole(model.RoleTest, model.TestE2E, 0.80, model.SignalFilename)

	role, subRole, _, _ := score.Resolve()

	if role != model.RoleTest {
		t.Errorf("Role = %v, want test", role)
	}
	if subRole != model.TestE2E {
		t.Errorf("SubRole = %v, want e2e", subRole)
	}
}

func TestRoleScoreResolve_TestDefaultSubRole(t *testing.T) {
	score := NewRoleScore()
	score.Add(model.RoleTest, 0.80, model.SignalFilename)

	role, subRole, _, _ := score.Resolve()

	if role != model.RoleTest {
		t.Errorf("Role = %v, want test", role)
	}
	if subRole != model.TestUnit {
		t.Errorf("SubRole = %v, want unit (default)", subRole)
	}
}

func TestRoleScoreResolve_NonTestNoSubRole(t *testing.T) {
	score := NewRoleScore()
	score.Add(model.RoleInfra, 0.80, model.SignalFilename)

	role, subRole, _, _ := score.Resolve()

	if role != model.RoleInfra {
		t.Errorf("Role = %v, want infra", role)
	}
	if subRole != "" {
		t.Errorf("SubRole = %v, want empty for non-test", subRole)
	}
}

func TestRoleScoreMaxWeight(t *testing.T) {
	score := NewRoleScore()
	score.Add(model.RoleTest, 0.75, model.SignalFilename)
	score.Add(model.RoleCore, 0.30, model.SignalPath)

	if score.MaxWeight() != 0.75 {
		t.Errorf("MaxWeight = %v, want 0.75", score.MaxWeight())
	}
}

func TestRoleScoreMaxWeight_Empty(t *testing.T) {
	score := NewRoleScore()

	if score.MaxWeight() != 0 {
		t.Errorf("MaxWeight = %v, want 0 for empty", score.MaxWeight())
	}
}

func TestRoleScoreMaxWeight_MultipleSignals(t *testing.T) {
	score := NewRoleScore()
	score.Add(model.RoleTest, 0.30, model.SignalFilename)
	score.Add(model.RoleTest, 0.30, model.SignalPath)
	score.Add(model.RoleCore, 0.50, model.SignalExtension)

	// Test has 0.60 total, Prod has 0.50
	if score.MaxWeight() != 0.60 {
		t.Errorf("MaxWeight = %v, want 0.60", score.MaxWeight())
	}
}

func TestRoleScoreResolve_AgreementBonus(t *testing.T) {
	// With more signals agreeing, confidence increases
	score1 := NewRoleScore()
	score1.Add(model.RoleTest, 0.80, model.SignalFilename)

	score2 := NewRoleScore()
	score2.Add(model.RoleTest, 0.40, model.SignalFilename)
	score2.Add(model.RoleTest, 0.40, model.SignalPath)

	_, _, conf1, _ := score1.Resolve()
	_, _, conf2, _ := score2.Resolve()

	// Both have same total weight (0.80), but score2 has more signals
	// score1: 0.80 * 0.25 = 0.20 (1 signal)
	// score2: 0.80 * 0.50 = 0.40 (2 signals)
	if conf2 <= conf1 {
		t.Errorf("conf2 (%v) should be > conf1 (%v) due to agreement bonus", conf2, conf1)
	}
}

func TestRoleScoreResolve_ConfidenceCapped(t *testing.T) {
	score := NewRoleScore()
	// Add many signals to potentially exceed 1.0
	score.Add(model.RoleTest, 0.90, model.SignalFilename)
	score.Add(model.RoleTest, 0.80, model.SignalPath)
	score.Add(model.RoleTest, 0.70, model.SignalExtension)
	score.Add(model.RoleTest, 0.60, model.SignalHeader)
	score.Add(model.RoleTest, 0.50, model.SignalOverride)

	_, _, confidence, _ := score.Resolve()

	if confidence > 1.0 {
		t.Errorf("Confidence = %v, should be capped at 1.0", confidence)
	}
}

func TestRolePriority(t *testing.T) {
	tests := []struct {
		role     model.Role
		priority int
	}{
		{model.RoleVendor, 1},
		{model.RoleGenerated, 2},
		{model.RoleTest, 3},
		{model.RoleInfra, 4},
		{model.RoleCore, 5},
		{model.RoleDocs, 6},
		{model.RoleConfig, 7},
		{model.RoleScripts, 8},
		{model.RoleExamples, 9},
		{model.RoleDeprecated, 10},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			got := rolePriority(tt.role)
			if got != tt.priority {
				t.Errorf("rolePriority(%v) = %v, want %v", tt.role, got, tt.priority)
			}
		})
	}
}

func TestRolePriority_Unknown(t *testing.T) {
	got := rolePriority("unknown")
	if got != 100 {
		t.Errorf("rolePriority(unknown) = %v, want 100", got)
	}
}

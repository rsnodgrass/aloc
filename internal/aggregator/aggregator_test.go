package aggregator

import (
	"testing"

	"github.com/modern-tooling/aloc/internal/model"
)

func TestComputeSummary(t *testing.T) {
	records := []*model.FileRecord{
		{Path: "a.go", LOC: 100, Language: "Go"},
		{Path: "b.go", LOC: 200, Language: "Go"},
		{Path: "c.ts", LOC: 150, Language: "TypeScript"},
	}

	summary := ComputeSummary(records)

	if summary.Files != 3 {
		t.Errorf("Files = %v, want 3", summary.Files)
	}
	if summary.LOCTotal != 450 {
		t.Errorf("LOCTotal = %v, want 450", summary.LOCTotal)
	}
	if summary.Languages != 2 {
		t.Errorf("Languages = %v, want 2", summary.Languages)
	}
}

func TestComputeSummary_ExcludesUnknown(t *testing.T) {
	records := []*model.FileRecord{
		{Path: "a.go", LOC: 100, Language: "Go"},
		{Path: "b.xyz", LOC: 50, Language: "unknown"},
	}

	summary := ComputeSummary(records)

	if summary.Languages != 1 {
		t.Errorf("Languages = %v, want 1 (excluding unknown)", summary.Languages)
	}
}

func TestComputeResponsibilities(t *testing.T) {
	records := []*model.FileRecord{
		{Path: "a.go", LOC: 100, Role: model.RoleCore, Confidence: 0.9},
		{Path: "b.go", LOC: 200, Role: model.RoleCore, Confidence: 0.95},
		{Path: "a_test.go", LOC: 150, Role: model.RoleTest, SubRole: model.TestUnit, Confidence: 0.85},
	}

	resp := ComputeResponsibilities(records)

	if len(resp) != 2 {
		t.Errorf("Responsibilities count = %v, want 2", len(resp))
	}

	// Should be sorted by LOC descending
	if resp[0].Role != model.RoleCore {
		t.Errorf("First role = %v, want core", resp[0].Role)
	}
	if resp[0].LOC != 300 {
		t.Errorf("Core LOC = %v, want 300", resp[0].LOC)
	}
	if resp[0].Files != 2 {
		t.Errorf("Prod Files = %v, want 2", resp[0].Files)
	}
}

func TestComputeResponsibilities_TestBreakdown(t *testing.T) {
	records := []*model.FileRecord{
		{Path: "a_test.go", LOC: 100, Role: model.RoleTest, SubRole: model.TestUnit, Confidence: 0.9},
		{Path: "b_test.go", LOC: 50, Role: model.RoleTest, SubRole: model.TestIntegration, Confidence: 0.9},
	}

	resp := ComputeResponsibilities(records)

	if resp[0].Breakdown == nil {
		t.Fatal("Test breakdown should not be nil")
	}

	unitPct := resp[0].Breakdown[model.TestUnit]
	if unitPct < 0.6 || unitPct > 0.7 {
		t.Errorf("Unit percentage = %v, want ~0.67", unitPct)
	}
}

func TestComputeRatios(t *testing.T) {
	resp := []model.Responsibility{
		{Role: model.RoleCore, LOC: 1000},
		{Role: model.RoleTest, LOC: 500},
		{Role: model.RoleInfra, LOC: 200},
	}

	ratios := ComputeRatios(resp)

	if ratios.TestToCore != 0.5 {
		t.Errorf("TestToCore = %v, want 0.5", ratios.TestToCore)
	}
	if ratios.InfraToCore != 0.2 {
		t.Errorf("InfraToCore = %v, want 0.2", ratios.InfraToCore)
	}
}

func TestComputeRatios_ZeroProd(t *testing.T) {
	resp := []model.Responsibility{
		{Role: model.RoleTest, LOC: 500},
	}

	ratios := ComputeRatios(resp)

	// Should not panic with division by zero
	if ratios.TestToCore != 500 {
		t.Errorf("TestToCore = %v, want 500 (core=1)", ratios.TestToCore)
	}
}

func TestComputeLanguageBreakdown(t *testing.T) {
	records := []*model.FileRecord{
		{Path: "a.go", LOC: 100, Lines: model.LineMetrics{Code: 100}, Language: "Go", Role: model.RoleCore},
		{Path: "b.go", LOC: 200, Lines: model.LineMetrics{Code: 200}, Language: "Go", Role: model.RoleCore},
		{Path: "a_test.go", LOC: 100, Lines: model.LineMetrics{Code: 100}, Language: "Go", Role: model.RoleTest},
		{Path: "c.ts", LOC: 150, Lines: model.LineMetrics{Code: 150}, Language: "TypeScript", Role: model.RoleCore},
	}

	langs := ComputeLanguageBreakdown(records)

	if len(langs) != 2 {
		t.Errorf("Languages count = %v, want 2", len(langs))
	}

	// Both Go and TypeScript are in "Primary" category, Go should be first (more LOC)
	if langs[0].Language != "Go" {
		t.Errorf("First language = %v, want Go", langs[0].Language)
	}
	if langs[0].LOCTotal != 400 {
		t.Errorf("Go LOCTotal = %v, want 400", langs[0].LOCTotal)
	}
	if langs[0].Category != "Primary" {
		t.Errorf("Go Category = %v, want Primary", langs[0].Category)
	}
}

func TestComputeLanguageBreakdown_SignificantOnly(t *testing.T) {
	records := []*model.FileRecord{
		{Path: "a.go", LOC: 1000, Lines: model.LineMetrics{Code: 1000}, Language: "Go", Role: model.RoleCore},
		{Path: "b.go", LOC: 10, Lines: model.LineMetrics{Code: 10}, Language: "Go", Role: model.RoleConfig}, // <10%
	}

	langs := ComputeLanguageBreakdown(records)

	// Config should be filtered out (only ~1% of Go total)
	if _, exists := langs[0].Responsibilities[model.RoleConfig]; exists {
		t.Error("Config should be filtered out (< 10%)")
	}
}

func TestCompute(t *testing.T) {
	records := []*model.FileRecord{
		{Path: "a.go", LOC: 100, Language: "Go", Role: model.RoleCore, Confidence: 0.9},
		{Path: "a_test.go", LOC: 50, Language: "Go", Role: model.RoleTest, Confidence: 0.95},
	}

	report := Compute(records, Options{IncludeFiles: true})

	if report.Meta.Generator != "aloc" {
		t.Errorf("Generator = %v, want aloc", report.Meta.Generator)
	}
	if report.Summary.Files != 2 {
		t.Errorf("Summary.Files = %v, want 2", report.Summary.Files)
	}
	if len(report.Files) != 2 {
		t.Errorf("Files count = %v, want 2", len(report.Files))
	}
}

func TestCompute_ExcludeFiles(t *testing.T) {
	records := []*model.FileRecord{
		{Path: "a.go", LOC: 100, Role: model.RoleCore},
	}

	report := Compute(records, Options{IncludeFiles: false})

	if report.Files != nil {
		t.Error("Files should be nil when IncludeFiles is false")
	}
}

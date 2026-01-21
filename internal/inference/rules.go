package inference

import "github.com/modern-tooling/aloc/internal/model"

// PathRule defines a path-based heuristic for role classification
type PathRule struct {
	Fragment string
	Role     model.Role
	Weight   float32
}

// PathRules contains all path-based classification rules
var PathRules = []PathRule{
	// Test paths
	{"/test/", model.RoleTest, 0.60},
	{"/tests/", model.RoleTest, 0.60},
	{"/__tests__/", model.RoleTest, 0.60},
	{"/spec/", model.RoleTest, 0.55},
	{"/unit/", model.RoleTest, 0.60},
	{"/integration/", model.RoleTest, 0.65},
	{"/e2e/", model.RoleTest, 0.70},

	// Infra paths
	{"/infra/", model.RoleInfra, 0.65},
	{"/terraform/", model.RoleInfra, 0.65},
	{"/pulumi/", model.RoleInfra, 0.65},
	{"/helm/", model.RoleInfra, 0.65},
	{"/.github/workflows/", model.RoleInfra, 0.70},
	{"/.gitlab-ci/", model.RoleInfra, 0.70},
	{"/ci/", model.RoleInfra, 0.60},
	{"/deploy/", model.RoleInfra, 0.65},

	// Docs paths
	{"/docs/", model.RoleDocs, 0.65},
	{"/doc/", model.RoleDocs, 0.65},
	{"/site/", model.RoleDocs, 0.60},

	// Config paths
	{"/config/", model.RoleConfig, 0.55},
	{"/configs/", model.RoleConfig, 0.55},

	// Scripts paths
	{"/scripts/", model.RoleScripts, 0.55},
	{"/tools/", model.RoleScripts, 0.55},
	{"/bin/", model.RoleScripts, 0.55},
	{"/hack/", model.RoleScripts, 0.50},

	// Examples paths
	{"/examples/", model.RoleExamples, 0.55},
	{"/samples/", model.RoleExamples, 0.55},
	{"/demo/", model.RoleExamples, 0.55},

	// Vendor paths (high confidence)
	{"/vendor/", model.RoleVendor, 0.90},
	{"/third_party/", model.RoleVendor, 0.90},
	{"/node_modules/", model.RoleVendor, 0.95},

	// Generated paths
	{"/dist/", model.RoleGenerated, 0.70},
	{"/build/", model.RoleGenerated, 0.70},
	{"/gen/", model.RoleGenerated, 0.80},
	{"/generated/", model.RoleGenerated, 0.80},
	{"/pb/", model.RoleGenerated, 0.80},
}

// FilenameRule defines a filename pattern-based heuristic
type FilenameRule struct {
	Pattern   string
	MatchType string // "suffix", "prefix", "contains"
	Role      model.Role
	SubRole   model.TestKind
	Weight    float32
}

// FilenameRules contains all filename-based classification rules
var FilenameRules = []FilenameRule{
	// Test patterns
	{"_test.", "contains", model.RoleTest, model.TestUnit, 0.75},
	{".spec.", "contains", model.RoleTest, model.TestUnit, 0.70},
	{".test.", "contains", model.RoleTest, model.TestUnit, 0.70},
	{"_spec.", "contains", model.RoleTest, model.TestUnit, 0.70},
	{".e2e.", "contains", model.RoleTest, model.TestE2E, 0.80},
	{"_e2e.", "contains", model.RoleTest, model.TestE2E, 0.80},
	{".integration.", "contains", model.RoleTest, model.TestIntegration, 0.80},
	{"_integration.", "contains", model.RoleTest, model.TestIntegration, 0.80},
	{"_fixture.", "contains", model.RoleTest, model.TestFixture, 0.60},
	{"_mock.", "contains", model.RoleTest, model.TestFixture, 0.55},
	{"_stub.", "contains", model.RoleTest, model.TestFixture, 0.55},
	{"_fake.", "contains", model.RoleTest, model.TestFixture, 0.55},

	// Infra patterns
	{"dockerfile", "prefix", model.RoleInfra, "", 0.85},
	{"docker-compose", "prefix", model.RoleInfra, "", 0.80},
	{"makefile", "prefix", model.RoleInfra, "", 0.65},
	{"taskfile", "prefix", model.RoleInfra, "", 0.65},
	{"justfile", "prefix", model.RoleInfra, "", 0.65},
	{".tf", "suffix", model.RoleInfra, "", 0.90},
	{".tfvars", "suffix", model.RoleInfra, "", 0.90},
	{"helmfile", "prefix", model.RoleInfra, "", 0.85},
	{".hcl", "suffix", model.RoleInfra, "", 0.80},

	// Config patterns
	{".env", "prefix", model.RoleConfig, "", 0.85},
	{"config.", "prefix", model.RoleConfig, "", 0.60},
	{"settings.", "prefix", model.RoleConfig, "", 0.60},
	{".config.", "contains", model.RoleConfig, "", 0.55},
	{".conf", "suffix", model.RoleConfig, "", 0.55},

	// Docs patterns
	{"readme", "prefix", model.RoleDocs, "", 0.80},
	{"changelog", "prefix", model.RoleDocs, "", 0.75},
	{"contributing", "prefix", model.RoleDocs, "", 0.75},
	{"license", "prefix", model.RoleDocs, "", 0.70},
}

// ExtensionRule defines an extension-based heuristic
type ExtensionRule struct {
	Ext    string
	Role   model.Role
	Weight float32
}

// ExtensionRules contains all extension-based classification rules
var ExtensionRules = []ExtensionRule{
	// Docs extensions
	{".md", model.RoleDocs, 0.20},
	{".mdx", model.RoleDocs, 0.20},
	{".rst", model.RoleDocs, 0.20},
	{".adoc", model.RoleDocs, 0.20},

	// Config extensions (weak signal)
	{".yaml", model.RoleConfig, 0.15},
	{".yml", model.RoleConfig, 0.15},
	{".toml", model.RoleConfig, 0.15},
	{".json", model.RoleConfig, 0.10},
	{".ini", model.RoleConfig, 0.15},

	// Generated extensions (strong signal)
	{".lock", model.RoleGenerated, 0.90},
	{".sum", model.RoleGenerated, 0.85},
	{".pb.go", model.RoleGenerated, 0.90},
	{".pb.ts", model.RoleGenerated, 0.90},
	{".gen.go", model.RoleGenerated, 0.85},

	// Interface/contract
	{".proto", model.RoleDocs, 0.40},
}

// HeaderRule defines a header content-based heuristic
type HeaderRule struct {
	Pattern string
	Role    model.Role
	Weight  float32
}

// HeaderRules contains all header-based classification rules
var HeaderRules = []HeaderRule{
	// Generated code markers
	{"Code generated by", model.RoleGenerated, 0.95},
	{"DO NOT EDIT", model.RoleGenerated, 0.90},
	{"@generated", model.RoleGenerated, 0.90},
	{"AUTO-GENERATED", model.RoleGenerated, 0.85},
	{"This file was automatically generated", model.RoleGenerated, 0.90},
	{"This file is auto-generated", model.RoleGenerated, 0.90},

	// Infra markers
	{"terraform {", model.RoleInfra, 0.80},
	{"provider \"", model.RoleInfra, 0.75},

	// Test markers (language-specific)
	{"describe(", model.RoleTest, 0.60},
	{"test(", model.RoleTest, 0.60},
	{"func Test", model.RoleTest, 0.70},
	{"@Test", model.RoleTest, 0.65},
	{"#[test]", model.RoleTest, 0.70},
	{"def test_", model.RoleTest, 0.65},

	// Deprecated markers
	{"// Deprecated:", model.RoleDeprecated, 0.70},
	{"// DEPRECATED", model.RoleDeprecated, 0.65},
	{"@deprecated", model.RoleDeprecated, 0.70},
}

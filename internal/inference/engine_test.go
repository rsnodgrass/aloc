package inference

import (
	"testing"

	"github.com/modern-tooling/aloc/internal/model"
)

func TestEngineInfer_TestFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/internal/auth/login_test.go",
		LOC:          100,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test", record.Role)
	}
	if record.Confidence < 0.1 {
		t.Errorf("Confidence = %v, should be positive", record.Confidence)
	}
	if record.Path != file.Path {
		t.Errorf("Path = %v, want %v", record.Path, file.Path)
	}
	if record.LOC != file.LOC {
		t.Errorf("LOC = %v, want %v", record.LOC, file.LOC)
	}
	if record.Language != file.LanguageHint {
		t.Errorf("Language = %v, want %v", record.Language, file.LanguageHint)
	}
}

func TestEngineInfer_TestFileInTestDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/test/auth/login_test.go",
		LOC:          100,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test", record.Role)
	}
	// Should have higher confidence with both path and filename signals
	if record.Confidence < 0.3 {
		t.Errorf("Confidence = %v, should be >= 0.3 with multiple signals", record.Confidence)
	}
}

func TestEngineInfer_InfraFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/infra/terraform/main.tf",
		LOC:          50,
		LanguageHint: "Terraform",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra", record.Role)
	}
}

func TestEngineInfer_DocsFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/docs/README.md",
		LOC:          30,
		LanguageHint: "Markdown",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleDocs {
		t.Errorf("Role = %v, want docs", record.Role)
	}
}

func TestEngineInfer_ProdFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/internal/service/handler.go",
		LOC:          200,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleCore {
		t.Errorf("Role = %v, want core", record.Role)
	}
}

func TestEngineInfer_VendorFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/vendor/github.com/pkg/errors/errors.go",
		LOC:          500,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleVendor {
		t.Errorf("Role = %v, want vendor", record.Role)
	}
}

func TestEngineInfer_NodeModulesVendor(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/node_modules/react/index.js",
		LOC:          1000,
		LanguageHint: "JavaScript",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleVendor {
		t.Errorf("Role = %v, want vendor", record.Role)
	}
}

func TestEngineInfer_GeneratedFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/gen/proto/api.pb.go",
		LOC:          500,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleGenerated {
		t.Errorf("Role = %v, want generated", record.Role)
	}
}

func TestEngineInfer_ConfigFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/config/settings.yaml",
		LOC:          50,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleConfig {
		t.Errorf("Role = %v, want config", record.Role)
	}
}

func TestEngineInfer_ScriptsFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/scripts/deploy.sh",
		LOC:          50,
		LanguageHint: "Shell",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleScripts {
		t.Errorf("Role = %v, want scripts", record.Role)
	}
}

func TestEngineInfer_ExamplesFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/examples/basic/main.go",
		LOC:          50,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleExamples {
		t.Errorf("Role = %v, want examples", record.Role)
	}
}

func TestEngineInfer_WithOverrides(t *testing.T) {
	engine := NewEngine(Options{
		Overrides: map[model.Role][]string{
			model.RoleInfra: {"ops/**"},
		},
	})

	file := &model.RawFile{
		Path:         "ops/scripts/deploy.sh",
		LOC:          50,
		LanguageHint: "Shell",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (from override)", record.Role)
	}
	// Confidence is adjusted by agreement factor: 1.0 * 0.25 (1 signal) = 0.25
	if record.Confidence != 0.25 {
		t.Errorf("Confidence = %v, want 0.25 for override with agreement factor", record.Confidence)
	}
	// Should have override signal
	if len(record.Signals) != 1 || record.Signals[0] != model.SignalOverride {
		t.Errorf("Signals = %v, want [override]", record.Signals)
	}
}

func TestEngineInfer_OverrideTakesPrecedence(t *testing.T) {
	engine := NewEngine(Options{
		Overrides: map[model.Role][]string{
			model.RoleCore: {"/project/test/**"},
		},
	})

	file := &model.RawFile{
		Path:         "/project/test/helper_test.go",
		LOC:          50,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	// Override should take precedence over path/filename rules
	if record.Role != model.RoleCore {
		t.Errorf("Role = %v, want core (from override)", record.Role)
	}
}

func TestEngineInfer_Dockerfile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/Dockerfile",
		LOC:          30,
		LanguageHint: "Dockerfile",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra", record.Role)
	}
}

func TestEngineInfer_DockerCompose(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/docker-compose.yml",
		LOC:          100,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra", record.Role)
	}
}

func TestEngineInfer_Makefile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/Makefile",
		LOC:          100,
		LanguageHint: "Makefile",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra", record.Role)
	}
}

func TestEngineInfer_SpecFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/src/components/Button.spec.tsx",
		LOC:          50,
		LanguageHint: "TypeScript",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test", record.Role)
	}
}

func TestEngineInfer_E2EFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/e2e/login.e2e.ts",
		LOC:          100,
		LanguageHint: "TypeScript",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test", record.Role)
	}
	if record.SubRole != model.TestE2E {
		t.Errorf("SubRole = %v, want e2e", record.SubRole)
	}
}

func TestEngineInfer_IntegrationFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/tests/api_integration.go",
		LOC:          200,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test", record.Role)
	}
	if record.SubRole != model.TestIntegration {
		t.Errorf("SubRole = %v, want integration", record.SubRole)
	}
}

func TestEngineInfer_GoSumFile(t *testing.T) {
	engine := NewEngine(Options{})

	// go.sum is a generated file that ends with .sum extension
	file := &model.RawFile{
		Path:         "/project/go.sum",
		LOC:          500,
		LanguageHint: "Go Checksums",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleGenerated {
		t.Errorf("Role = %v, want generated", record.Role)
	}
}

func TestEngineInfer_PbGoFile(t *testing.T) {
	engine := NewEngine(Options{})

	// .pb.go files are generated protobuf files
	file := &model.RawFile{
		Path:         "/project/api/service.pb.go",
		LOC:          1000,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleGenerated {
		t.Errorf("Role = %v, want generated", record.Role)
	}
}

func TestEngineInferBatch(t *testing.T) {
	engine := NewEngine(Options{})

	files := []*model.RawFile{
		{Path: "/project/src/main.go", LOC: 100, LanguageHint: "Go"},
		{Path: "/project/test/main_test.go", LOC: 50, LanguageHint: "Go"},
		{Path: "/project/docs/README.md", LOC: 30, LanguageHint: "Markdown"},
	}

	records := engine.InferBatch(files)

	if len(records) != 3 {
		t.Fatalf("Records count = %v, want 3", len(records))
	}

	if records[0].Role != model.RoleCore {
		t.Errorf("records[0].Role = %v, want core", records[0].Role)
	}
	if records[1].Role != model.RoleTest {
		t.Errorf("records[1].Role = %v, want test", records[1].Role)
	}
	if records[2].Role != model.RoleDocs {
		t.Errorf("records[2].Role = %v, want docs", records[2].Role)
	}
}

func TestEngineInferBatch_WithNeighborhood(t *testing.T) {
	engine := NewEngine(Options{Neighborhood: true})

	files := []*model.RawFile{
		{Path: "/project/test/a_test.go", LOC: 100, LanguageHint: "Go"},
		{Path: "/project/test/b_test.go", LOC: 100, LanguageHint: "Go"},
		{Path: "/project/test/c_test.go", LOC: 100, LanguageHint: "Go"},
		{Path: "/project/test/helper.go", LOC: 50, LanguageHint: "Go"},
	}

	records := engine.InferBatch(files)

	if len(records) != 4 {
		t.Fatalf("Records count = %v, want 4", len(records))
	}

	// helper.go should get test role from neighborhood
	helper := records[3]
	if helper.Role != model.RoleTest {
		t.Errorf("helper.go Role = %v, want test (from neighborhood)", helper.Role)
	}
}

func TestEngineInferBatch_NeighborhoodDisabled(t *testing.T) {
	engine := NewEngine(Options{Neighborhood: false})

	files := []*model.RawFile{
		{Path: "/project/test/a_test.go", LOC: 100, LanguageHint: "Go"},
		{Path: "/project/test/b_test.go", LOC: 100, LanguageHint: "Go"},
		{Path: "/project/test/c_test.go", LOC: 100, LanguageHint: "Go"},
		{Path: "/project/test/helper.go", LOC: 50, LanguageHint: "Go"},
	}

	records := engine.InferBatch(files)

	if len(records) != 4 {
		t.Fatalf("Records count = %v, want 4", len(records))
	}

	// Without neighborhood, helper.go gets prod (default) since it only has path signal
	helper := records[3]
	// With /test/ path signal, it should still get test role even without neighborhood
	if helper.Role != model.RoleTest {
		t.Errorf("helper.go Role = %v, want test (from path signal)", helper.Role)
	}
}

func TestEngineInferBatch_Empty(t *testing.T) {
	engine := NewEngine(Options{})

	records := engine.InferBatch([]*model.RawFile{})

	if len(records) != 0 {
		t.Errorf("Records count = %v, want 0", len(records))
	}
}

func TestEngineInfer_GithubWorkflow(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/.github/workflows/ci.yml",
		LOC:          50,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra", record.Role)
	}
}

func TestEngineInfer_EnvFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/.env.example",
		LOC:          20,
		LanguageHint: "Dotenv",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleConfig {
		t.Errorf("Role = %v, want config", record.Role)
	}
}

func TestEngineInfer_ProtoFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/api/proto/service.proto",
		LOC:          100,
		LanguageHint: "Protocol Buffers",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleDocs {
		t.Errorf("Role = %v, want docs (interface definition)", record.Role)
	}
}

func TestEngineInfer_MockFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/internal/service/client_mock.go",
		LOC:          100,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test", record.Role)
	}
	if record.SubRole != model.TestFixture {
		t.Errorf("SubRole = %v, want fixture", record.SubRole)
	}
}

func TestEngineInfer_FixtureFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/test/data_fixture.json",
		LOC:          50,
		LanguageHint: "JSON",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test", record.Role)
	}
}

func TestNewEngine_NilOverrides(t *testing.T) {
	engine := NewEngine(Options{})

	if engine.overrides != nil {
		t.Error("overrides should be nil when not provided")
	}
}

func TestNewEngine_WithOverrides(t *testing.T) {
	engine := NewEngine(Options{
		Overrides: map[model.Role][]string{
			model.RoleInfra: {"deploy/**"},
		},
	})

	if engine.overrides == nil {
		t.Error("overrides should not be nil when provided")
	}
}

func TestNewEngine_OptionsFlags(t *testing.T) {
	engine := NewEngine(Options{
		HeaderProbe:  true,
		Neighborhood: true,
	})

	if !engine.enableHeaderProbe {
		t.Error("enableHeaderProbe should be true")
	}
	if !engine.enableNeighborhood {
		t.Error("enableNeighborhood should be true")
	}
}

func TestEngineInfer_GenGoFile(t *testing.T) {
	engine := NewEngine(Options{})

	// .gen.go files are generated files
	file := &model.RawFile{
		Path:         "/project/internal/models/user.gen.go",
		LOC:          200,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleGenerated {
		t.Errorf("Role = %v, want generated", record.Role)
	}
}

func TestEngineInfer_DistDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/dist/bundle.js",
		LOC:          10000,
		LanguageHint: "JavaScript",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleGenerated {
		t.Errorf("Role = %v, want generated (dist directory)", record.Role)
	}
}

func TestEngineInfer_BuildDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/build/output.js",
		LOC:          5000,
		LanguageHint: "JavaScript",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleGenerated {
		t.Errorf("Role = %v, want generated (build directory)", record.Role)
	}
}

func TestEngineInfer_ThirdPartyVendor(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/third_party/some-lib/code.go",
		LOC:          300,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleVendor {
		t.Errorf("Role = %v, want vendor (third_party)", record.Role)
	}
}

func TestEngineInfer_CIDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/ci/pipeline.yml",
		LOC:          100,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (ci directory)", record.Role)
	}
}

func TestEngineInfer_DeployDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/deploy/kubernetes/deployment.yaml",
		LOC:          50,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (deploy directory)", record.Role)
	}
}

func TestEngineInfer_SamplesDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/samples/hello-world/main.go",
		LOC:          20,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleExamples {
		t.Errorf("Role = %v, want examples (samples directory)", record.Role)
	}
}

func TestEngineInfer_DemoDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/demo/app/main.go",
		LOC:          100,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleExamples {
		t.Errorf("Role = %v, want examples (demo directory)", record.Role)
	}
}

func TestEngineInfer_HackDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/hack/update-deps.sh",
		LOC:          50,
		LanguageHint: "Shell",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleScripts {
		t.Errorf("Role = %v, want scripts (hack directory)", record.Role)
	}
}

func TestEngineInfer_BinDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/bin/run.sh",
		LOC:          20,
		LanguageHint: "Shell",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleScripts {
		t.Errorf("Role = %v, want scripts (bin directory)", record.Role)
	}
}

func TestEngineInfer_ConfigsDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/configs/app.yaml",
		LOC:          30,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleConfig {
		t.Errorf("Role = %v, want config (configs directory)", record.Role)
	}
}

func TestEngineInfer_JestTestsDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/__tests__/component.test.js",
		LOC:          100,
		LanguageHint: "JavaScript",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test (__tests__ directory)", record.Role)
	}
}

func TestEngineInfer_SpecDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/spec/models/user_spec.rb",
		LOC:          50,
		LanguageHint: "Ruby",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test (spec directory)", record.Role)
	}
}

func TestEngineInfer_UnitTestDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/test/unit/service_test.go",
		LOC:          100,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test (unit test directory)", record.Role)
	}
}

func TestEngineInfer_IntegrationTestDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/test/integration/api_test.go",
		LOC:          200,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test (integration test directory)", record.Role)
	}
}

func TestEngineInfer_E2ETestDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/test/e2e/flow_test.go",
		LOC:          300,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test (e2e test directory)", record.Role)
	}
}

func TestEngineInfer_PulumiDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/pulumi/index.ts",
		LOC:          100,
		LanguageHint: "TypeScript",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (pulumi directory)", record.Role)
	}
}

func TestEngineInfer_HelmDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/helm/chart/values.yaml",
		LOC:          50,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (helm directory)", record.Role)
	}
}

func TestEngineInfer_TerraformFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/main.tf",
		LOC:          100,
		LanguageHint: "Terraform",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (.tf file)", record.Role)
	}
}

func TestEngineInfer_TfvarsFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/terraform/prod.tfvars",
		LOC:          30,
		LanguageHint: "Terraform",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (.tfvars file)", record.Role)
	}
}

func TestEngineInfer_HclFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/terragrunt.hcl",
		LOC:          50,
		LanguageHint: "HCL",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (.hcl file)", record.Role)
	}
}

func TestEngineInfer_DocDirSingular(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/doc/guide.md",
		LOC:          200,
		LanguageHint: "Markdown",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleDocs {
		t.Errorf("Role = %v, want docs (doc directory)", record.Role)
	}
}

func TestEngineInfer_SiteDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/site/index.html",
		LOC:          100,
		LanguageHint: "HTML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleDocs {
		t.Errorf("Role = %v, want docs (site directory)", record.Role)
	}
}

func TestEngineInfer_ChangelogFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/CHANGELOG.md",
		LOC:          500,
		LanguageHint: "Markdown",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleDocs {
		t.Errorf("Role = %v, want docs (CHANGELOG file)", record.Role)
	}
}

func TestEngineInfer_ContributingFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/CONTRIBUTING.md",
		LOC:          100,
		LanguageHint: "Markdown",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleDocs {
		t.Errorf("Role = %v, want docs (CONTRIBUTING file)", record.Role)
	}
}

func TestEngineInfer_LicenseFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/LICENSE",
		LOC:          20,
		LanguageHint: "Text",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleDocs {
		t.Errorf("Role = %v, want docs (LICENSE file)", record.Role)
	}
}

func TestEngineInfer_ConfigPrefix(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/config.yaml",
		LOC:          50,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleConfig {
		t.Errorf("Role = %v, want config (config. prefix)", record.Role)
	}
}

func TestEngineInfer_SettingsPrefix(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/settings.json",
		LOC:          30,
		LanguageHint: "JSON",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleConfig {
		t.Errorf("Role = %v, want config (settings. prefix)", record.Role)
	}
}

func TestEngineInfer_ConfSuffix(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/nginx.conf",
		LOC:          100,
		LanguageHint: "Nginx",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleConfig {
		t.Errorf("Role = %v, want config (.conf suffix)", record.Role)
	}
}

func TestEngineInfer_Taskfile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/Taskfile.yml",
		LOC:          50,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (Taskfile)", record.Role)
	}
}

func TestEngineInfer_Justfile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/Justfile",
		LOC:          100,
		LanguageHint: "Just",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (Justfile)", record.Role)
	}
}

func TestEngineInfer_Helmfile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/helmfile.yaml",
		LOC:          100,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (helmfile)", record.Role)
	}
}

func TestEngineInfer_StubFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/internal/service/client_stub.go",
		LOC:          50,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test", record.Role)
	}
	if record.SubRole != model.TestFixture {
		t.Errorf("SubRole = %v, want fixture", record.SubRole)
	}
}

func TestEngineInfer_FakeFile(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/internal/db/repository_fake.go",
		LOC:          100,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleTest {
		t.Errorf("Role = %v, want test", record.Role)
	}
	if record.SubRole != model.TestFixture {
		t.Errorf("SubRole = %v, want fixture", record.SubRole)
	}
}

func TestEngineInfer_GitlabCIDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/.gitlab-ci/jobs/build.yml",
		LOC:          50,
		LanguageHint: "YAML",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleInfra {
		t.Errorf("Role = %v, want infra (.gitlab-ci directory)", record.Role)
	}
}

func TestEngineInfer_GenDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/gen/models/user.go",
		LOC:          200,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleGenerated {
		t.Errorf("Role = %v, want generated (gen directory)", record.Role)
	}
}

func TestEngineInfer_GeneratedDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/generated/api/client.go",
		LOC:          500,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleGenerated {
		t.Errorf("Role = %v, want generated (generated directory)", record.Role)
	}
}

func TestEngineInfer_PbDir(t *testing.T) {
	engine := NewEngine(Options{})

	file := &model.RawFile{
		Path:         "/project/pb/service.go",
		LOC:          300,
		LanguageHint: "Go",
	}

	record := engine.Infer(file)

	if record.Role != model.RoleGenerated {
		t.Errorf("Role = %v, want generated (pb directory)", record.Role)
	}
}

package aggregator

import "github.com/modern-tooling/aloc/internal/model"

func ComputeRatios(responsibilities []model.Responsibility) model.Ratios {
	byRole := make(map[model.Role]int)
	for _, r := range responsibilities {
		byRole[r.Role] = r.LOC
	}

	coreLOC := byRole[model.RoleCore]
	if coreLOC == 0 {
		coreLOC = 1 // avoid division by zero
	}

	return model.Ratios{
		TestToCore:      float32(byRole[model.RoleTest]) / float32(coreLOC),
		InfraToCore:     float32(byRole[model.RoleInfra]) / float32(coreLOC),
		DocsToCore:      float32(byRole[model.RoleDocs]) / float32(coreLOC),
		GeneratedToCore: float32(byRole[model.RoleGenerated]) / float32(coreLOC),
		ConfigToCore:    float32(byRole[model.RoleConfig]) / float32(coreLOC),
	}
}

package aggregator

import (
	"time"

	"github.com/modern-tooling/aloc/internal/model"
)

const Version = "0.1.0"

type Options struct {
	IncludeFiles  bool
	RepoInfo      *model.RepoInfo
	IncludeEffort bool
	EffortOpts    EffortOptions
}

func Compute(records []*model.FileRecord, opts Options) *model.Report {
	responsibilities := ComputeResponsibilities(records)

	report := &model.Report{
		Meta: model.Meta{
			SchemaVersion:    "1.0",
			GeneratedAt:      time.Now().UTC(),
			Generator:        "aloc",
			GeneratorVersion: Version,
			Repo:             opts.RepoInfo,
		},
		Summary:          ComputeSummary(records),
		Responsibilities: responsibilities,
		Ratios:           ComputeRatios(responsibilities),
		Languages:        ComputeLanguageBreakdown(records),
		Confidence:       computeConfidenceInfo(records),
	}

	if opts.IncludeEffort {
		report.Effort = ComputeEffortWithResponsibilities(
			report.Summary.LOCTotal,
			report.Summary.Lines,
			responsibilities,
			report.Ratios,
			opts.EffortOpts,
		)
	}

	if opts.IncludeFiles {
		report.Files = records
	}

	return report
}

func computeConfidenceInfo(records []*model.FileRecord) model.ConfidenceInfo {
	var totalLOC int
	var highConfLOC int
	var overrideLOC int

	for _, r := range records {
		totalLOC += r.LOC
		if r.Confidence >= 0.80 {
			highConfLOC += r.LOC
		}
		for _, s := range r.Signals {
			if s == model.SignalOverride {
				overrideLOC += r.LOC
				break
			}
		}
	}

	if totalLOC == 0 {
		return model.ConfidenceInfo{}
	}

	return model.ConfidenceInfo{
		AutoClassified: float32(highConfLOC) / float32(totalLOC),
		Heuristic:      float32(totalLOC-highConfLOC-overrideLOC) / float32(totalLOC),
		Override:       float32(overrideLOC) / float32(totalLOC),
	}
}

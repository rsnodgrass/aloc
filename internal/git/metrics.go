package git

import (
	"fmt"
	"time"

	"github.com/modern-tooling/aloc/internal/model"
)

// Options controls git analysis behavior
type Options struct {
	SparklineMonths int  // months of history for sparklines (default 6)
	StabilityMonths int  // months threshold for stable code (default 18)
	Smooth          bool // use bi-weekly buckets instead of weekly
}

// DefaultOptions returns sensible defaults
func DefaultOptions() Options {
	return Options{
		SparklineMonths: 6,
		StabilityMonths: 18,
		Smooth:          false,
	}
}

// Analyze performs full git history analysis
func Analyze(root string, records []*model.FileRecord, opts Options) (*GitMetrics, error) {
	if opts.SparklineMonths == 0 {
		opts.SparklineMonths = 6
	}
	if opts.StabilityMonths == 0 {
		opts.StabilityMonths = 18
	}

	// use longer window for stability analysis
	historyMonths := max(opts.StabilityMonths, opts.SparklineMonths)

	// parse git history
	events, err := ParseHistory(ParseOptions{
		SinceMonths: historyMonths,
		Root:        root,
	})
	if err != nil {
		return nil, fmt.Errorf("parse git history: %w", err)
	}

	if len(events) == 0 {
		return &GitMetrics{
			WindowMonths:      opts.SparklineMonths,
			ParallelismSignal: "low",
			AnalysisNote:      "No commits found in analysis window",
		}, nil
	}

	// map roles from current file records
	MapRoles(events, records)
	fileLOC := BuildFileLOCMap(records)

	now := time.Now()

	// compute metrics
	churnStat := CalculateChurnConcentration(events)
	stableCore, volatileSurface := CalculateStability(events, fileLOC, opts.StabilityMonths)
	rewritePressure := CalculateRewritePressure(events)
	ownershipConc := CalculateOwnershipConcentration(events, fileLOC)
	parallelism := CalculateParallelismSignal(events)

	// build sparklines
	churnSeries := BuildChurnSeries(events, now, opts.SparklineMonths, opts.Smooth)

	// build AI timeline (shared across all roles)
	aiTimeline := buildAITimeline(events, now, opts.SparklineMonths, opts.Smooth)
	hasAnyAI := HasAnyAIAssisted(events)

	// calculate effort adjustments
	adjustments, net := CalculateEffortAdjustments(
		churnStat, stableCore, volatileSurface, rewritePressure, ownershipConc,
		churnSeries,
	)

	// generate analysis note
	note := generateAnalysisNote(churnSeries, parallelism)

	bucketCount := 0
	if len(churnSeries) > 0 {
		for _, s := range churnSeries {
			bucketCount = len(s.Buckets)
			break
		}
	}

	return &GitMetrics{
		ChurnConcentration:     churnStat,
		StableCore:             stableCore,
		VolatileSurface:        volatileSurface,
		RewritePressure:        rewritePressure,
		OwnershipConcentration: ownershipConc,
		ParallelismSignal:      parallelism,
		ChurnSeries:            churnSeries,
		AITimeline:             aiTimeline,
		HasAnyAI:               hasAnyAI,
		Adjustments:            adjustments,
		NetAdjustment:          net,
		WindowMonths:           opts.SparklineMonths,
		BucketCount:            bucketCount,
		CommitCount:            len(events),
		AnalysisNote:           note,
	}, nil
}

// buildAITimeline creates the AI marker timeline aligned with sparkline buckets
func buildAITimeline(events []ChangeEvent, now time.Time, months int, smooth bool) []bool {
	var buckets []Bucket
	if smooth {
		buckets = BuildBiweeklyBuckets(now, months)
	} else {
		buckets = BuildWeeklyBuckets(now, months)
	}

	AssignAIMarkers(buckets, events)
	return BuildAITimeline(buckets)
}

// generateAnalysisNote creates a single interpretive sentence
func generateAnalysisNote(churnSeries map[model.Role]*Sparkline, parallelism string) string {
	// check for late infra volatility pattern
	if infraSparkline, ok := churnSeries[model.RoleInfra]; ok {
		if lateVolatility(infraSparkline.Buckets) {
			return "Late infra volatility coincides with release hardening."
		}
	}

	// check for sustained prod activity
	if prodSparkline, ok := churnSeries[model.RoleCore]; ok {
		if sustainedHighChurn(prodSparkline.Buckets) {
			return "Sustained prod activity suggests active development phase."
		}
	}

	// default based on parallelism
	switch parallelism {
	case "high":
		return "High parallelism indicates distributed team activity."
	case "moderate":
		return "Moderate parallelism with overlapping commit windows."
	default:
		return "Low parallelism suggests focused, sequential work."
	}
}

package git

import (
	"time"

	"github.com/modern-tooling/aloc/internal/model"
)

// Bucket represents a time window for churn aggregation
type Bucket struct {
	Start time.Time
	End   time.Time
	Churn int  // lines added + lines deleted
	HasAI bool // at least one AI-assisted commit in this bucket
}

// ChangeEvent represents a single file change from git history
type ChangeEvent struct {
	When       time.Time
	Path       string
	Added      int
	Deleted    int
	Role       model.Role // mapped from file classification
	Author     string     // hashed for privacy, used only for ownership calc
	AIAssisted bool       // commit had explicit AI assistance marker
}

// Sparkline is the rendered output for a responsibility
type Sparkline struct {
	Role    model.Role
	Buckets []Bucket
	Glyphs  string // pre-rendered at default resolution
	Values  []int  // raw weekly churn values for adaptive rendering
}

// ChurnStat represents churn concentration metrics
type ChurnStat struct {
	FilePercent float64 // top X% of files
	EditPercent float64 // account for Y% of edits
}

// EffortAdjustment represents a single adjustment factor
type EffortAdjustment struct {
	Reason     string
	Adjustment float64 // +0.12 = +12%
}

// GitMetrics contains all git-derived metrics
type GitMetrics struct {
	// Summary metrics
	ChurnConcentration     ChurnStat // "14% of files → 67% of edits"
	StableCore             float64   // % LOC untouched in N months
	VolatileSurface        float64   // % LOC changed ≥5× recently
	RewritePressure        float64   // delete/add ratio
	OwnershipConcentration float64   // % LOC owned by single author
	ParallelismSignal      string    // "low" | "moderate" | "high"

	// Sparkline data (normalized 0-1 per weekly bucket)
	ChurnSeries map[model.Role]*Sparkline

	// AI assistance timeline (binary per bucket, shared across roles)
	AITimeline []bool // true if bucket had any AI-assisted commit
	HasAnyAI   bool   // true if any commit in window was AI-assisted

	// Effort adjustments
	Adjustments   []EffortAdjustment
	NetAdjustment float64 // multiplicative factor (e.g., 0.25 = +25%)

	// Metadata
	WindowMonths int
	BucketCount  int
	CommitCount  int
	AnalysisNote string // single interpretation sentence
}

// RepoHint contains lightweight git detection info (no full analysis)
type RepoHint struct {
	HasGit     bool
	RepoAge    time.Duration // time since first commit
	LastCommit time.Time     // most recent commit
	IsActive   bool          // commit in last 7 days
}

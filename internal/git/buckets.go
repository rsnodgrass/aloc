package git

import (
	"time"

	"github.com/modern-tooling/aloc/internal/model"
)

// BuildWeeklyBuckets creates time buckets for sparkline aggregation
func BuildWeeklyBuckets(now time.Time, months int) []Bucket {
	// 6 months ≈ 26 weeks
	weeks := months*4 + 2

	buckets := make([]Bucket, weeks)

	end := now.Truncate(24 * time.Hour)
	for i := weeks - 1; i >= 0; i-- {
		start := end.AddDate(0, 0, -7)
		buckets[i] = Bucket{
			Start: start,
			End:   end,
			Churn: 0,
		}
		end = start
	}

	return buckets
}

// BuildDailyBuckets creates fine-grained time buckets for full-width sparklines
func BuildDailyBuckets(now time.Time, months int) []Bucket {
	// 6 months ≈ 180 days
	days := months * 30

	buckets := make([]Bucket, days)

	end := now.Truncate(24 * time.Hour)
	for i := days - 1; i >= 0; i-- {
		start := end.AddDate(0, 0, -1)
		buckets[i] = Bucket{
			Start: start,
			End:   end,
			Churn: 0,
		}
		end = start
	}

	return buckets
}

// BuildFixedBuckets creates a specific number of buckets spanning the given months
func BuildFixedBuckets(now time.Time, months int, count int) []Bucket {
	if count <= 0 {
		count = 60 // default to ~60 chars width
	}

	totalDays := months * 30
	daysPerBucket := totalDays / count
	if daysPerBucket < 1 {
		daysPerBucket = 1
	}

	buckets := make([]Bucket, count)

	end := now.Truncate(24 * time.Hour)
	for i := count - 1; i >= 0; i-- {
		start := end.AddDate(0, 0, -daysPerBucket)
		buckets[i] = Bucket{
			Start: start,
			End:   end,
			Churn: 0,
		}
		end = start
	}

	return buckets
}

// BuildBiweeklyBuckets creates larger time buckets for smoother sparklines
func BuildBiweeklyBuckets(now time.Time, months int) []Bucket {
	// 6 months ≈ 13 bi-weekly periods
	periods := months*2 + 1

	buckets := make([]Bucket, periods)

	end := now.Truncate(24 * time.Hour)
	for i := periods - 1; i >= 0; i-- {
		start := end.AddDate(0, 0, -14)
		buckets[i] = Bucket{
			Start: start,
			End:   end,
			Churn: 0,
		}
		end = start
	}

	return buckets
}

// AssignChurn distributes change events into time buckets
func AssignChurn(buckets []Bucket, events []ChangeEvent, roleFilter model.Role) {
	for _, ev := range events {
		if roleFilter != "" && ev.Role != roleFilter {
			continue
		}

		churn := ev.Added + ev.Deleted

		for i := range buckets {
			if ev.When.After(buckets[i].Start) && !ev.When.After(buckets[i].End) {
				buckets[i].Churn += churn
				break
			}
		}
	}
}

// NormalizeBuckets converts churn counts to 0-1 range
func NormalizeBuckets(buckets []Bucket) []float64 {
	max := 0
	for _, b := range buckets {
		if b.Churn > max {
			max = b.Churn
		}
	}

	normalized := make([]float64, len(buckets))
	if max == 0 {
		return normalized // all zeros → baseline glyphs
	}

	for i, b := range buckets {
		normalized[i] = float64(b.Churn) / float64(max)
	}

	return normalized
}

// AssignAIMarkers marks buckets that contain AI-assisted commits
// This is a binary signal per bucket (not a count or percentage)
func AssignAIMarkers(buckets []Bucket, events []ChangeEvent) {
	for _, ev := range events {
		if !ev.AIAssisted {
			continue
		}

		for i := range buckets {
			if ev.When.After(buckets[i].Start) && !ev.When.After(buckets[i].End) {
				buckets[i].HasAI = true
				break
			}
		}
	}
}

// BuildAITimeline extracts the HasAI signal from buckets
func BuildAITimeline(buckets []Bucket) []bool {
	timeline := make([]bool, len(buckets))
	for i, b := range buckets {
		timeline[i] = b.HasAI
	}
	return timeline
}

// HasAnyAIAssisted checks if any events are AI-assisted
func HasAnyAIAssisted(events []ChangeEvent) bool {
	for _, ev := range events {
		if ev.AIAssisted {
			return true
		}
	}
	return false
}

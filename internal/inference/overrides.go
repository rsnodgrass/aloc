package inference

import (
	"path/filepath"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
)

type Overrides struct {
	rules []overrideRule
}

type overrideRule struct {
	pattern string
	role    model.Role
}

type OverrideResult struct {
	Role    model.Role
	Pattern string
}

func NewOverrides(config map[model.Role][]string) *Overrides {
	var rules []overrideRule
	for role, patterns := range config {
		for _, pattern := range patterns {
			rules = append(rules, overrideRule{
				pattern: pattern,
				role:    role,
			})
		}
	}
	return &Overrides{rules: rules}
}

func (o *Overrides) Match(filePath string) *OverrideResult {
	for _, rule := range o.rules {
		if matchGlob(rule.pattern, filePath) {
			return &OverrideResult{
				Role:    rule.role,
				Pattern: rule.pattern,
			}
		}
	}
	return nil
}

func matchGlob(pattern, path string) bool {
	// Normalize paths
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	// Handle ** wildcard
	if strings.Contains(pattern, "**") {
		return matchDoublestar(pattern, path)
	}

	// Use standard filepath.Match for simple patterns
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	if matched {
		return true
	}

	// Try matching against full path
	matched, _ = filepath.Match(pattern, path)
	return matched
}

func matchDoublestar(pattern, path string) bool {
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return false
	}

	prefix := parts[0]
	suffix := parts[1]

	// Check prefix
	if prefix != "" && !strings.HasPrefix(path, strings.TrimSuffix(prefix, "/")) {
		return false
	}

	// Check suffix
	if suffix != "" {
		suffix = strings.TrimPrefix(suffix, "/")
		if !strings.HasSuffix(path, suffix) && !matchAnySuffix(path, suffix) {
			return false
		}
	}

	return true
}

func matchAnySuffix(path, pattern string) bool {
	// Handle patterns like *.go
	if strings.HasPrefix(pattern, "*") {
		ext := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, ext)
	}

	// Check if pattern matches any path component
	components := strings.Split(path, "/")
	for i := range components {
		subpath := strings.Join(components[i:], "/")
		if matched, _ := filepath.Match(pattern, subpath); matched {
			return true
		}
	}
	return false
}

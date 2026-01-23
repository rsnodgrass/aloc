package git

import "testing"

func TestDetectAIMarker(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected bool
	}{
		{
			name:     "claude code marker",
			body:     "Fix bug in parser\n\nCo-Authored-By: Claude <noreply@anthropic.com>",
			expected: true,
		},
		{
			name:     "claude code with model name",
			body:     "Add feature\n\nCo-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>",
			expected: true,
		},
		{
			name:     "aider marker",
			body:     "Refactor tests\n\nCo-authored-by: aider (gpt-4) <noreply@aider.chat>",
			expected: true,
		},
		{
			name:     "generic ai-assisted marker",
			body:     "Update docs\n\nAI-Assisted: true",
			expected: true,
		},
		{
			name:     "generic ai-assisted-by marker",
			body:     "Fix typo\n\nAI-Assisted-By: claude",
			expected: true,
		},
		{
			name:     "case insensitive",
			body:     "Update\n\nCO-AUTHORED-BY: CLAUDE <noreply@anthropic.com>",
			expected: true,
		},
		{
			name:     "no marker - regular commit",
			body:     "Fix bug in parser",
			expected: false,
		},
		{
			name:     "no marker - human co-author",
			body:     "Fix bug\n\nCo-Authored-By: John Doe <john@example.com>",
			expected: false,
		},
		{
			name:     "no marker - similar but not matching",
			body:     "Claude helped me understand this issue",
			expected: false,
		},
		{
			name:     "empty body",
			body:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectAIMarker(tt.body)
			if result != tt.expected {
				t.Errorf("detectAIMarker(%q) = %v, want %v", tt.body, result, tt.expected)
			}
		})
	}
}

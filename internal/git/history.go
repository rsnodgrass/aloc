package git

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/modern-tooling/aloc/internal/model"
)

// ParseOptions controls git log parsing
type ParseOptions struct {
	SinceMonths int    // how far back to look
	Root        string // repository root
}

// ParseHistory runs git log and returns change events
func ParseHistory(opts ParseOptions) ([]ChangeEvent, error) {
	since := time.Now().AddDate(0, -opts.SinceMonths, 0).Format("2006-01-02")

	// single efficient git command
	// format: hash|email|timestamp followed by commit body (for AI marker detection)
	// %x00 separates header from body, %x01 marks end of body
	cmd := exec.Command("git", "-C", opts.Root,
		"log",
		"--numstat",
		"--format=%H|%ae|%aI%x00%b%x01",
		"--since="+since,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseGitLog(string(out)), nil
}

// parseGitLog parses the git log output into change events
func parseGitLog(output string) []ChangeEvent {
	var events []ChangeEvent
	var currentAuthor string
	var currentTime time.Time
	var currentAIAssisted bool

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// commit header line: hash|email|timestamp followed by \x00 body \x01
		// the body may span multiple lines between \x00 and \x01
		if strings.Contains(line, "|") && strings.Count(line, "|") == 2 {
			// extract header and potential body start
			header, bodyPart, _ := strings.Cut(line, "\x00")

			parts := strings.Split(header, "|")
			if len(parts) == 3 {
				currentAuthor = hashAuthor(parts[1])
				t, err := time.Parse(time.RFC3339, parts[2])
				if err == nil {
					currentTime = t
				}

				// collect full commit body (may span multiple lines)
				var body strings.Builder
				body.WriteString(bodyPart)

				// body ends at \x01 marker
				for !strings.Contains(body.String(), "\x01") && scanner.Scan() {
					body.WriteString("\n")
					body.WriteString(scanner.Text())
				}

				currentAIAssisted = detectAIMarker(body.String())
			}
			continue
		}

		// numstat line: added\tdeleted\tpath
		fields := strings.Split(line, "\t")
		if len(fields) == 3 {
			// handle binary files (- - path)
			if fields[0] == "-" || fields[1] == "-" {
				continue
			}

			added, err1 := strconv.Atoi(fields[0])
			deleted, err2 := strconv.Atoi(fields[1])
			if err1 != nil || err2 != nil {
				continue
			}

			events = append(events, ChangeEvent{
				When:       currentTime,
				Path:       fields[2],
				Added:      added,
				Deleted:    deleted,
				Author:     currentAuthor,
				AIAssisted: currentAIAssisted,
			})
		}
	}

	return events
}

// detectAIMarker checks if commit message contains explicit AI assistance markers
// Only detects explicit markers, never infers from style or timing
func detectAIMarker(body string) bool {
	lower := strings.ToLower(body)

	// supported markers (case-insensitive)
	// only includes tools verified to add commit markers
	markers := []string{
		// claude code: "Co-Authored-By: Claude <noreply@anthropic.com>"
		"co-authored-by: claude",
		// aider: "Co-authored-by: aider (model) <noreply@aider.chat>"
		"co-authored-by: aider",
		// generic markers teams may add manually
		"ai-assisted:",
		"ai-assisted-by:",
	}

	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// hashAuthor creates a privacy-preserving hash of an email
func hashAuthor(email string) string {
	h := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return hex.EncodeToString(h[:8]) // 16 char hex, anonymous
}

// MapRoles assigns roles to events based on file records
func MapRoles(events []ChangeEvent, records []*model.FileRecord) {
	roleMap := make(map[string]model.Role)
	for _, r := range records {
		roleMap[r.Path] = r.Role
	}

	for i := range events {
		if role, ok := roleMap[events[i].Path]; ok {
			events[i].Role = role
		}
	}
}

// BuildFileLOCMap creates a map of file paths to their current LOC
func BuildFileLOCMap(records []*model.FileRecord) map[string]int {
	m := make(map[string]int)
	for _, r := range records {
		m[r.Path] = r.LOC
	}
	return m
}

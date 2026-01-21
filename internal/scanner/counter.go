package scanner

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modern-tooling/aloc/internal/model"
)

// bufferPool reuses 256KB buffers - the documented sweet spot for SSD I/O
var bufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 256*1024)
		return &buf
	},
}

// CountLOC counts lines of code, returning 0 if file is binary.
// This combines binary detection with LOC counting to avoid opening the file twice.
func CountLOC(path string) (int, error) {
	metrics, err := CountLines(path)
	if err != nil {
		return 0, err
	}
	return metrics.Code, nil
}

// CountLines counts all line types (total, blanks, comments, code).
// Returns zero metrics if file is binary.
func CountLines(path string) (model.LineMetrics, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.LineMetrics{}, err
	}
	defer f.Close()

	// Get pooled buffer for binary check
	bufPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufPtr)
	buf := *bufPtr

	// Read first chunk and check for binary
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return model.LineMetrics{}, err
	}

	// Binary check: look for NUL byte in first 512 bytes
	checkLen := min(n, 512)
	for i := 0; i < checkLen; i++ {
		if buf[i] == 0 {
			return model.LineMetrics{}, nil // binary file, no metrics
		}
	}

	// Seek back to start for line counting
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return model.LineMetrics{}, err
	}

	lang := detectLangFromPath(path)
	return countLinesFromReader(f, lang, bufPtr), nil
}

func countLinesFromReader(f *os.File, lang string, bufPtr *[]byte) model.LineMetrics {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(*bufPtr, 256*1024)

	var metrics model.LineMetrics
	inBlockComment := false
	blockStart, blockEnd := getBlockCommentMarkers(lang)
	lineComment := getLineCommentMarker(lang)

	for scanner.Scan() {
		metrics.Total++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Empty line
		if trimmed == "" {
			metrics.Blanks++
			continue
		}

		// Track if this line contributes to code
		isCode := false
		isComment := false

		// Handle block comments
		if blockStart != "" {
			if inBlockComment {
				// Inside block comment
				if idx := strings.Index(trimmed, blockEnd); idx >= 0 {
					inBlockComment = false
					remainder := strings.TrimSpace(trimmed[idx+len(blockEnd):])
					if remainder != "" && !strings.HasPrefix(remainder, lineComment) {
						isCode = true
					} else if remainder != "" {
						isComment = true
					} else {
						isComment = true
					}
				} else {
					isComment = true
				}
			} else if strings.Contains(trimmed, blockStart) {
				// Check for block comment start
				startIdx := strings.Index(trimmed, blockStart)
				beforeComment := strings.TrimSpace(trimmed[:startIdx])
				endIdx := strings.Index(trimmed[startIdx+len(blockStart):], blockEnd)

				if endIdx >= 0 {
					// Block comment starts and ends on same line
					afterComment := strings.TrimSpace(trimmed[startIdx+len(blockStart)+endIdx+len(blockEnd):])
					if beforeComment != "" || afterComment != "" {
						isCode = true
					} else {
						isComment = true
					}
				} else {
					// Block comment starts but doesn't end
					inBlockComment = true
					if beforeComment != "" {
						isCode = true
					} else {
						isComment = true
					}
				}
			} else if lineComment != "" && strings.HasPrefix(trimmed, lineComment) {
				isComment = true
			} else {
				isCode = true
			}
		} else {
			// No block comment support for this language
			if lineComment != "" && strings.HasPrefix(trimmed, lineComment) {
				isComment = true
			} else {
				isCode = true
			}
		}

		if isCode {
			metrics.Code++
		} else if isComment {
			metrics.Comments++
		}
	}

	return metrics
}

// countLOCFromReader is kept for backward compatibility
func countLOCFromReader(f *os.File, lang string, bufPtr *[]byte) int {
	return countLinesFromReader(f, lang, bufPtr).Code
}

func getLineCommentMarker(lang string) string {
	switch lang {
	case "Go", "TypeScript", "JavaScript", "Java", "C", "C++", "Rust", "Swift", "Kotlin", "Scala", "C#", "PHP", "Dart", "Groovy":
		return "//"
	case "Python", "Shell", "YAML", "Ruby", "Perl", "R", "Makefile", "Dockerfile", "Just", "TOML":
		return "#"
	case "SQL", "Lua", "Haskell":
		return "--"
	case "Lisp", "Clojure", "Scheme":
		return ";"
	case "Plain Text":
		return "" // no comments in plain text
	default:
		return "//"
	}
}

func getBlockCommentMarkers(lang string) (string, string) {
	switch lang {
	case "Go", "TypeScript", "JavaScript", "Java", "C", "C++", "Rust", "Swift", "Kotlin", "Scala", "CSS", "SQL", "C#", "PHP", "Dart", "Groovy", "SCSS", "SASS", "LESS":
		return "/*", "*/"
	case "HTML", "XML", "Markdown", "MDX", "Vue", "Svelte":
		return "<!--", "-->"
	case "Python":
		return `"""`, `"""`
	case "Shell", "Makefile", "Dockerfile", "Just", "TOML", "YAML", "Ruby", "Perl", "R", "Plain Text":
		return "", "" // no block comments
	case "Lua":
		return "--[[", "]]"
	case "Haskell":
		return "{-", "-}"
	default:
		return "/*", "*/"
	}
}

func detectLangFromPath(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	return extToLanguage(ext)
}

// CountLinesWithEmbedded counts lines and extracts embedded code blocks (for Markdown/MDX)
func CountLinesWithEmbedded(path string) (model.LineMetrics, map[string]model.LineMetrics, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.LineMetrics{}, nil, err
	}
	defer f.Close()

	// Get pooled buffer for binary check
	bufPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufPtr)
	buf := *bufPtr

	// Read first chunk and check for binary
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return model.LineMetrics{}, nil, err
	}

	// Binary check: look for NUL byte in first 512 bytes
	checkLen := min(n, 512)
	for i := 0; i < checkLen; i++ {
		if buf[i] == 0 {
			return model.LineMetrics{}, nil, nil // binary file
		}
	}

	// Seek back to start
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return model.LineMetrics{}, nil, err
	}

	lang := detectLangFromPath(path)
	if lang == "Markdown" || lang == "MDX" {
		return countMarkdownWithEmbedded(f, bufPtr)
	}

	// Non-Markdown: use regular counting
	metrics := countLinesFromReader(f, lang, bufPtr)
	return metrics, nil, nil
}

// countMarkdownWithEmbedded parses Markdown and extracts fenced code blocks
func countMarkdownWithEmbedded(f *os.File, bufPtr *[]byte) (model.LineMetrics, map[string]model.LineMetrics, error) {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(*bufPtr, 256*1024)

	var metrics model.LineMetrics
	embedded := make(map[string]model.LineMetrics)

	inCodeBlock := false
	codeBlockLang := ""
	var codeBlockLines []string

	for scanner.Scan() {
		metrics.Total++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for fenced code block start/end
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				// Starting a code block
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(trimmed, "```")
				codeBlockLang = strings.TrimSpace(codeBlockLang)
				// Normalize language name
				codeBlockLang = normalizeCodeBlockLang(codeBlockLang)
				codeBlockLines = nil
				metrics.Code++ // the ``` line itself is "code" in Markdown
			} else {
				// Ending a code block
				inCodeBlock = false
				metrics.Code++ // the closing ``` line

				// Process accumulated code block
				if codeBlockLang != "" && len(codeBlockLines) > 0 {
					blockMetrics := countCodeBlockLines(codeBlockLines, codeBlockLang)
					existing := embedded[codeBlockLang]
					existing.Total += blockMetrics.Total
					existing.Code += blockMetrics.Code
					existing.Comments += blockMetrics.Comments
					existing.Blanks += blockMetrics.Blanks
					embedded[codeBlockLang] = existing
				}
				codeBlockLang = ""
			}
			continue
		}

		if inCodeBlock {
			// Inside code block - accumulate for later processing
			codeBlockLines = append(codeBlockLines, line)
			metrics.Code++ // code blocks count as code in Markdown
		} else if trimmed == "" {
			metrics.Blanks++
		} else if strings.HasPrefix(trimmed, "<!--") {
			metrics.Comments++
		} else {
			metrics.Code++ // prose is "code" in Markdown
		}
	}

	if len(embedded) == 0 {
		return metrics, nil, nil
	}
	return metrics, embedded, nil
}

// normalizeCodeBlockLang converts code fence language hints to canonical names
func normalizeCodeBlockLang(hint string) string {
	// Remove common suffixes/annotations
	hint = strings.Split(hint, " ")[0] // "typescript jsx" -> "typescript"
	hint = strings.Trim(hint, "`")     // remove any stray backticks
	hint = strings.ToLower(hint)

	switch hint {
	case "ts", "typescript":
		return "TypeScript"
	case "tsx":
		return "TypeScript" // TSX grouped with TypeScript
	case "js", "javascript":
		return "JavaScript"
	case "jsx":
		return "JavaScript"
	case "py", "python", "python3":
		return "Python"
	case "go", "golang":
		return "Go"
	case "rb", "ruby":
		return "Ruby"
	case "sh", "bash", "shell", "zsh":
		return "Shell"
	case "yml", "yaml":
		return "YAML"
	case "json", "jsonc":
		return "JSON"
	case "sql":
		return "SQL"
	case "html":
		return "HTML"
	case "css":
		return "CSS"
	case "scss":
		return "SCSS"
	case "dockerfile":
		return "Dockerfile"
	case "tf", "terraform", "hcl":
		return "Terraform"
	case "rs", "rust":
		return "Rust"
	case "java":
		return "Java"
	case "c":
		return "C"
	case "cpp", "c++", "cxx":
		return "C++"
	case "cs", "csharp":
		return "C#"
	case "md", "markdown":
		return "Markdown"
	case "xml":
		return "XML"
	case "toml":
		return "TOML"
	case "makefile", "make":
		return "Makefile"
	case "":
		return "" // no language specified
	default:
		// Capitalize first letter for unknown languages
		if len(hint) > 0 {
			return strings.ToUpper(hint[:1]) + hint[1:]
		}
		return hint
	}
}

// countCodeBlockLines counts lines within a code block using language-specific rules
func countCodeBlockLines(lines []string, lang string) model.LineMetrics {
	var metrics model.LineMetrics
	lineComment := getLineCommentMarker(lang)
	blockStart, blockEnd := getBlockCommentMarkers(lang)
	inBlockComment := false

	for _, line := range lines {
		metrics.Total++
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			metrics.Blanks++
			continue
		}

		isCode := false
		isComment := false

		// Handle block comments
		if blockStart != "" {
			if inBlockComment {
				if idx := strings.Index(trimmed, blockEnd); idx >= 0 {
					inBlockComment = false
					remainder := strings.TrimSpace(trimmed[idx+len(blockEnd):])
					if remainder != "" && !strings.HasPrefix(remainder, lineComment) {
						isCode = true
					} else {
						isComment = true
					}
				} else {
					isComment = true
				}
			} else if strings.Contains(trimmed, blockStart) {
				startIdx := strings.Index(trimmed, blockStart)
				beforeComment := strings.TrimSpace(trimmed[:startIdx])
				endIdx := strings.Index(trimmed[startIdx+len(blockStart):], blockEnd)

				if endIdx >= 0 {
					afterComment := strings.TrimSpace(trimmed[startIdx+len(blockStart)+endIdx+len(blockEnd):])
					if beforeComment != "" || afterComment != "" {
						isCode = true
					} else {
						isComment = true
					}
				} else {
					inBlockComment = true
					if beforeComment != "" {
						isCode = true
					} else {
						isComment = true
					}
				}
			} else if lineComment != "" && strings.HasPrefix(trimmed, lineComment) {
				isComment = true
			} else {
				isCode = true
			}
		} else {
			if lineComment != "" && strings.HasPrefix(trimmed, lineComment) {
				isComment = true
			} else {
				isCode = true
			}
		}

		if isCode {
			metrics.Code++
		} else if isComment {
			metrics.Comments++
		}
	}

	return metrics
}

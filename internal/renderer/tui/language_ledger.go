package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// tableCell represents a single cell with its plain text and optional styling
type tableCell struct {
	text  string                                      // plain text (no ANSI)
	style func(string, *renderer.Theme) string       // optional styling function
}

// tableRow represents a row in the table
type tableRow struct {
	cells      []tableCell
	isHeader   bool   // category header row (spans all columns)
	headerText string // text for header rows
	isBlank    bool   // blank separator row
}

// alignColumn specifies column alignment
type alignColumn int

const (
	alignLeft alignColumn = iota
	alignRight
)

// tableSpec defines the table structure
type tableSpec struct {
	alignments []alignColumn
	colWidths  []int
}

// computeColumnWidths calculates max width for each column across all rows
func computeColumnWidths(rows []tableRow, numCols int) []int {
	widths := make([]int, numCols)
	for _, row := range rows {
		if row.isHeader || row.isBlank {
			continue
		}
		for i, cell := range row.cells {
			if i < numCols && len(cell.text) > widths[i] {
				widths[i] = len(cell.text)
			}
		}
	}
	return widths
}

// renderAlignedTable renders rows with computed column widths
func renderAlignedTable(b *strings.Builder, rows []tableRow, spec tableSpec, theme *renderer.Theme) {
	for _, row := range rows {
		if row.isBlank {
			b.WriteString("\n")
			continue
		}
		if row.isHeader {
			b.WriteString(theme.Dim.Render(row.headerText) + "\n")
			continue
		}

		for i, cell := range row.cells {
			width := spec.colWidths[i]
			align := alignLeft
			if i < len(spec.alignments) {
				align = spec.alignments[i]
			}

			// Pad the plain text to computed width
			var padded string
			if align == alignRight {
				padded = fmt.Sprintf("%*s", width, cell.text)
			} else {
				padded = fmt.Sprintf("%-*s", width, cell.text)
			}

			// Apply styling if present
			if cell.style != nil {
				b.WriteString(cell.style(padded, theme))
			} else {
				b.WriteString(padded)
			}

			// Add spacing between columns (except after last)
			if i < len(row.cells)-1 {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n")
	}
}

// RenderLanguageLedger renders a numeric language breakdown grouped by category
// Shows languages that cumulatively account for ≥95% of total LOC
// If noEmbedded is true, embedded code blocks (e.g., in Markdown) are not shown
func RenderLanguageLedger(languages []model.LanguageComp, theme *renderer.Theme, noEmbedded bool) string {
	if len(languages) == 0 {
		return ""
	}

	// Calculate total LOC
	var totalLOC int
	for _, lang := range languages {
		totalLOC += lang.LOCTotal
	}

	// Filter to 95% coverage only if >8 languages, otherwise show all
	var filtered []model.LanguageComp
	var cumulativeLOC int

	if len(languages) > 8 {
		threshold := float64(totalLOC) * 0.95
		for _, lang := range languages {
			filtered = append(filtered, lang)
			cumulativeLOC += lang.LOCTotal
			if float64(cumulativeLOC) >= threshold {
				break
			}
		}
	} else {
		filtered = languages
		cumulativeLOC = totalLOC
	}

	var b strings.Builder

	// Header - only show coverage note if we filtered (i.e., >8 languages)
	if len(languages) > 8 && totalLOC > 0 {
		coveragePct := float64(cumulativeLOC) / float64(totalLOC) * 100
		b.WriteString(theme.PrimaryBold.Render("Language Breakdown") +
			theme.Dim.Render(fmt.Sprintf(" (covers %.0f%% of LOC)", coveragePct)) + "\n")
	} else {
		b.WriteString(theme.PrimaryBold.Render("Language Breakdown") + "\n")
	}
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

	// PHASE 1: Format all cells (no printing)
	// Columns: [indent+language, code, comments, blanks, total, tests]
	const numCols = 6
	var rows []tableRow

	// Column header row
	rows = append(rows, tableRow{
		cells: []tableCell{
			{text: ""},
			{text: "Code"},
			{text: "Comments"},
			{text: "Blanks"},
			{text: "Total"},
			{text: "(Tests)", style: styleDim},
		},
	})

	// Collect embedded rows separately to include in width calculation
	type embeddedRowData struct {
		parentIdx int
		rows      []tableRow
	}
	var embeddedData []embeddedRowData

	currentCategory := ""
	for _, lang := range filtered {
		// Category header when category changes
		if lang.Category != currentCategory {
			if currentCategory != "" {
				rows = append(rows, tableRow{isBlank: true})
			}
			currentCategory = lang.Category
			rows = append(rows, tableRow{isHeader: true, headerText: currentCategory})
		}

		// Format test cell
		testText, testPct := formatTestParts(lang.Tests, lang.Code)

		// Language data row
		rows = append(rows, tableRow{
			cells: []tableCell{
				{text: "  " + lang.Language},
				{text: formatLOCPlain(lang.Code), style: styleMagnitude(lang.Code)},
				{text: formatLOCPlain(lang.Comments), style: styleMagnitude(lang.Comments)},
				{text: formatLOCPlain(lang.Blanks), style: styleDim},
				{text: formatLOCPlain(lang.LOCTotal), style: styleMagnitude(lang.LOCTotal)},
				{text: testText + " " + testPct, style: styleDim},
			},
		})

		// Collect embedded language rows
		if !noEmbedded && len(lang.Embedded) > 0 {
			embRows := formatEmbeddedRows(lang.Embedded, theme)
			embeddedData = append(embeddedData, embeddedRowData{
				parentIdx: len(rows) - 1,
				rows:      embRows,
			})
			// Include embedded rows in width calculation
			rows = append(rows, embRows...)
		}
	}

	// PHASE 2: Compute column widths
	colWidths := computeColumnWidths(rows, numCols)

	// Ensure minimum widths for readability
	minWidths := []int{22, 6, 8, 6, 6, 11}
	for i, min := range minWidths {
		if i < len(colWidths) && colWidths[i] < min {
			colWidths[i] = min
		}
	}

	// PHASE 3: Render with computed widths
	spec := tableSpec{
		alignments: []alignColumn{alignLeft, alignRight, alignRight, alignRight, alignRight, alignRight},
		colWidths:  colWidths,
	}

	// Render rows (embedded rows were already appended in correct positions)
	renderAlignedTable(&b, rows, spec, theme)

	return b.String()
}

// formatLOCPlain formats LOC as plain string without ANSI codes
// Returns consistent width formatting with magnitude suffix
func formatLOCPlain(n int) string {
	if n == 0 {
		return "—"
	}
	if n >= 1000000000 {
		return fmt.Sprintf("%.1fG", float64(n)/1000000000)
	}
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// formatTestParts returns the test LOC and percentage as separate strings
func formatTestParts(tests, code int) (string, string) {
	if tests == 0 {
		return "—", ""
	}
	pct := 0
	if code > 0 {
		pct = int(float64(tests) / float64(code) * 100)
	}
	return formatLOCPlain(tests), fmt.Sprintf("%3d%%", pct)
}

// styleDim applies dim styling
func styleDim(text string, theme *renderer.Theme) string {
	return theme.Dim.Render(text)
}

// styleMagnitude returns a styling function based on value magnitude
func styleMagnitude(n int) func(string, *renderer.Theme) string {
	return func(text string, theme *renderer.Theme) string {
		if n == 0 {
			return theme.Dim.Render(text)
		}
		if n >= 1000000000 {
			return theme.MagnitudeG.Render(text)
		}
		if n >= 1000000 {
			return theme.MagnitudeM.Render(text)
		}
		if n >= 1000 {
			return theme.MagnitudeK.Render(text)
		}
		return theme.MagnitudeBytes.Render(text)
	}
}

// embeddedEntry holds a language and its metrics for sorting
type embeddedEntry struct {
	Language string
	Metrics  model.LineMetrics
}

// formatEmbeddedRows formats embedded language data as table rows
func formatEmbeddedRows(embedded map[string]model.LineMetrics, _ *renderer.Theme) []tableRow {
	// Sort embedded languages by total lines (descending)
	entries := make([]embeddedEntry, 0, len(embedded))
	for lang, metrics := range embedded {
		entries = append(entries, embeddedEntry{Language: lang, Metrics: metrics})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Metrics.Total > entries[j].Metrics.Total
	})

	// Separate into significant (>=100 lines) and other
	var significant []embeddedEntry
	var other model.LineMetrics
	for _, entry := range entries {
		total := entry.Metrics.Code + entry.Metrics.Comments + entry.Metrics.Blanks
		if total >= 100 {
			significant = append(significant, entry)
		} else {
			other.Code += entry.Metrics.Code
			other.Comments += entry.Metrics.Comments
			other.Blanks += entry.Metrics.Blanks
			other.Total += total
		}
	}

	var rows []tableRow

	// Format significant languages with tree prefix
	for i, entry := range significant {
		m := entry.Metrics
		total := m.Code + m.Comments + m.Blanks
		prefix := "├─"
		if i == len(significant)-1 && other.Total == 0 {
			prefix = "└─"
		}

		rows = append(rows, tableRow{
			cells: []tableCell{
				{text: fmt.Sprintf("  %s %s", prefix, entry.Language), style: styleDim},
				{text: formatLOCPlain(m.Code), style: styleDim},
				{text: formatLOCPlain(m.Comments), style: styleDim},
				{text: formatLOCPlain(m.Blanks), style: styleDim},
				{text: formatLOCPlain(total), style: styleDim},
				{text: ""}, // no tests for embedded
			},
		})
	}

	// Format "(other)" if there are rolled-up languages
	if other.Total > 0 {
		rows = append(rows, tableRow{
			cells: []tableCell{
				{text: "  └─ (other)", style: styleDim},
				{text: formatLOCPlain(other.Code), style: styleDim},
				{text: formatLOCPlain(other.Comments), style: styleDim},
				{text: formatLOCPlain(other.Blanks), style: styleDim},
				{text: formatLOCPlain(other.Total), style: styleDim},
				{text: ""}, // no tests for embedded
			},
		})
	}

	return rows
}


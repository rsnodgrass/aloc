package inference

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
)

type Engine struct {
	overrides         *Overrides
	enableHeaderProbe bool
	enableNeighborhood bool
}

type Options struct {
	HeaderProbe   bool
	Neighborhood  bool
	Overrides     map[model.Role][]string
}

func NewEngine(opts Options) *Engine {
	var overrides *Overrides
	if opts.Overrides != nil {
		overrides = NewOverrides(opts.Overrides)
	}
	return &Engine{
		overrides:          overrides,
		enableHeaderProbe:  opts.HeaderProbe,
		enableNeighborhood: opts.Neighborhood,
	}
}

func (e *Engine) Infer(file *model.RawFile) *model.FileRecord {
	score := NewRoleScore()

	// 1. Check overrides first (weight 1.0)
	if e.overrides != nil {
		if override := e.overrides.Match(file.Path); override != nil {
			score.Add(override.Role, 1.0, model.SignalOverride)
			return e.buildRecord(file, score)
		}
	}

	// 2. Apply path rules
	applyPathRules(file.Path, score)

	// 3. Apply filename rules
	applyFilenameRules(file.Path, score)

	// 4. Apply extension bias (only if not already decisive)
	if score.MaxWeight() < 0.50 {
		applyExtensionRules(file.Path, score)
	}

	// 5. Apply header probe (optional)
	if e.enableHeaderProbe && score.MaxWeight() < 0.80 {
		applyHeaderRules(file.Path, score)
	}

	return e.buildRecord(file, score)
}

func (e *Engine) InferBatch(files []*model.RawFile) []*model.FileRecord {
	records := make([]*model.FileRecord, len(files))
	for i, f := range files {
		records[i] = e.Infer(f)
	}

	// Second pass: neighborhood inference
	if e.enableNeighborhood {
		applyNeighborhoodInference(records)
	}

	return records
}

func (e *Engine) buildRecord(file *model.RawFile, score *RoleScore) *model.FileRecord {
	role, subRole, confidence, signals := score.Resolve()
	return &model.FileRecord{
		Path:       file.Path,
		LOC:        file.LOC,
		Lines:      file.Lines,
		Language:   file.LanguageHint,
		Role:       role,
		SubRole:    subRole,
		Confidence: confidence,
		Signals:    signals,
		Embedded:   file.Embedded,
	}
}

func applyPathRules(path string, score *RoleScore) {
	lowerPath := strings.ToLower(path)
	for _, rule := range PathRules {
		if strings.Contains(lowerPath, rule.Fragment) {
			score.Add(rule.Role, rule.Weight, model.SignalPath)
		}
	}
}

func applyFilenameRules(path string, score *RoleScore) {
	filename := strings.ToLower(filepath.Base(path))
	for _, rule := range FilenameRules {
		var matched bool
		switch rule.MatchType {
		case "suffix":
			matched = strings.HasSuffix(filename, strings.ToLower(rule.Pattern))
		case "prefix":
			matched = strings.HasPrefix(filename, strings.ToLower(rule.Pattern))
		case "contains":
			matched = strings.Contains(filename, strings.ToLower(rule.Pattern))
		}
		if matched {
			score.AddWithSubRole(rule.Role, rule.SubRole, rule.Weight, model.SignalFilename)
		}
	}
}

func applyExtensionRules(path string, score *RoleScore) {
	ext := strings.ToLower(filepath.Ext(path))
	// Check compound extensions like .pb.go
	base := filepath.Base(path)
	for _, rule := range ExtensionRules {
		if strings.HasSuffix(strings.ToLower(base), rule.Ext) || ext == rule.Ext {
			score.Add(rule.Role, rule.Weight, model.SignalExtension)
		}
	}
}

func applyHeaderRules(path string, score *RoleScore) {
	header, err := readHeader(path, 2048)
	if err != nil {
		return
	}
	content := string(header)
	for _, rule := range HeaderRules {
		if strings.Contains(content, rule.Pattern) {
			score.Add(rule.Role, rule.Weight, model.SignalHeader)
		}
	}
}

func readHeader(path string, maxBytes int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, maxBytes)
	n, _ := f.Read(buf)
	return buf[:n], nil
}

func applyNeighborhoodInference(records []*model.FileRecord) {
	// Group by directory
	byDir := make(map[string][]*model.FileRecord)
	for _, r := range records {
		dir := filepath.Dir(r.Path)
		byDir[dir] = append(byDir[dir], r)
	}

	for _, dirRecords := range byDir {
		if len(dirRecords) < 3 {
			continue
		}

		// Count role distribution
		roleCounts := make(map[model.Role]int)
		var total int
		for _, r := range dirRecords {
			if r.Confidence >= 0.70 {
				roleCounts[r.Role]++
				total++
			}
		}

		if total < 2 {
			continue
		}

		// Find dominant role
		var dominantRole model.Role
		var dominantCount int
		for role, count := range roleCounts {
			if count > dominantCount {
				dominantCount = count
				dominantRole = role
			}
		}

		// Apply to low-confidence files
		dominantRatio := float32(dominantCount) / float32(total)
		if dominantRatio >= 0.70 {
			for _, r := range dirRecords {
				if r.Confidence < 0.60 && r.Role != dominantRole {
					r.Role = dominantRole
					r.Confidence = r.Confidence + 0.40
					if r.Confidence > 1.0 {
						r.Confidence = 1.0
					}
					r.Signals = append(r.Signals, model.SignalNeighborhood)
				}
			}
		}
	}
}

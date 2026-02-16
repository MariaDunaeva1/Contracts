package validator

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DatasetStats struct {
	NumExamples int            `json:"num_examples"`
	AvgLength   float64        `json:"avg_length"`
	ClassDist   map[string]int `json:"class_distribution"`
}

type ValidationResult struct {
	Valid    bool            `json:"valid"`
	Checks   map[string]bool `json:"checks"`
	Warnings []string        `json:"warnings"`
	Errors   []string        `json:"errors"`
	Stats    DatasetStats    `json:"stats"`
}

type DatasetExample struct {
	Text  string `json:"text"`
	Label string `json:"label"`
}

func ValidateDataset(content []byte, format string) ValidationResult {
	result := ValidationResult{
		Checks:   make(map[string]bool),
		Warnings: []string{},
		Errors:   []string{},
		Stats: DatasetStats{
			ClassDist: make(map[string]int),
		},
	}

	var examples []DatasetExample

	// 1. Parsing
	err := json.Unmarshal(content, &examples)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid JSON structure: %v", err))
		result.Valid = false
		return result
	}

	if len(examples) == 0 {
		result.Errors = append(result.Errors, "Dataset is empty")
		result.Valid = false
		return result
	}

	// 2. Quality Checks
	totalLength := 0
	uniqueTexts := make(map[string]bool)
	duplicates := 0

	for i, ex := range examples {
		// Check fields
		if strings.TrimSpace(ex.Text) == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Example %d: missing 'text'", i))
		}
		if strings.TrimSpace(ex.Label) == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Example %d: missing 'label'", i))
		}

		// Length check
		if len(ex.Text) > 2000 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Example %d: text too long (%d > 2000 chars)", i, len(ex.Text)))
		}
		totalLength += len(ex.Text)

		// Duplicates
		if uniqueTexts[ex.Text] {
			duplicates++
		}
		uniqueTexts[ex.Text] = true

		// Stats
		result.Stats.ClassDist[ex.Label]++
	}

	// Stats Calculation
	result.Stats.NumExamples = len(examples)
	if len(examples) > 0 {
		result.Stats.AvgLength = float64(totalLength) / float64(len(examples))
	}

	// 3. Minimum Examples Rule
	if len(examples) < 50 {
		result.Errors = append(result.Errors, fmt.Sprintf("Insufficient examples: %d < 50", len(examples)))
	}

	// 4. Duplicate Limit Rule (e.g. warning if > 10% duplicates)
	duplicateRate := float64(duplicates) / float64(len(examples))
	if duplicateRate > 0.1 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("High duplicate rate: %.2f%%", duplicateRate*100))
	}

	// Final Validity Check
	result.Checks["format_valid"] = len(result.Errors) == 0
	result.Checks["min_examples"] = len(examples) >= 50
	result.Valid = len(result.Errors) == 0

	return result
}

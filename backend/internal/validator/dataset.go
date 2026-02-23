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

type DatasetMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DatasetChat struct {
	Messages []DatasetMessage `json:"messages"`
}

type DatasetExample struct {
	Text  string `json:"text"`
	Label string `json:"label"`
}

type DatasetInstruction struct {
	Instruction string `json:"instruction"`
	Input       string `json:"input"`
	Output      string `json:"output"`
}

// CUAD/SQuAD format structures
type CUADDataset struct {
	Version string     `json:"version"`
	Data    []CUADData `json:"data"`
}

type CUADData struct {
	Title      string          `json:"title"`
	Paragraphs []CUADParagraph `json:"paragraphs"`
}

type CUADParagraph struct {
	Context string   `json:"context"`
	Qas     []CUADQa `json:"qas"`
}

type CUADQa struct {
	Question     string       `json:"question"`
	Id           string       `json:"id"`
	Answers      []CUADAnswer `json:"answers"`
	IsImpossible bool         `json:"is_impossible"`
}

type CUADAnswer struct {
	Text        string `json:"text"`
	AnswerStart int    `json:"answer_start"`
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
	var chatExamples []DatasetChat

	// 1. Parsing
	isChat := false

	firstChar := byte(0)
	for _, b := range content {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			firstChar = b
			break
		}
	}

	if firstChar == '[' {
		// Try parsing as JSON array
		var rawList []json.RawMessage
		if err := json.Unmarshal(content, &rawList); err == nil && len(rawList) > 0 {
			// Probe the first element
			var first map[string]interface{}
			if json.Unmarshal(rawList[0], &first) == nil {
				if _, ok := first["messages"]; ok {
					isChat = true
					chatExamples = make([]DatasetChat, 0, len(rawList))
					for _, raw := range rawList {
						var chatEx DatasetChat
						if json.Unmarshal(raw, &chatEx) == nil && len(chatEx.Messages) > 0 {
							chatExamples = append(chatExamples, chatEx)
						}
					}
				} else if _, ok := first["instruction"]; ok {
					examples = make([]DatasetExample, 0, len(rawList))
					for _, raw := range rawList {
						var ie DatasetInstruction
						if json.Unmarshal(raw, &ie) == nil {
							text := ie.Instruction
							if ie.Input != "" {
								text += "\n\nInput: " + ie.Input
							}
							examples = append(examples, DatasetExample{Text: text, Label: ie.Output})
						}
					}
				} else if _, ok := first["text"]; ok {
					examples = make([]DatasetExample, 0, len(rawList))
					for _, raw := range rawList {
						var ex DatasetExample
						if json.Unmarshal(raw, &ex) == nil && ex.Text != "" {
							examples = append(examples, ex)
						}
					}
				} else {
					result.Errors = append(result.Errors, "Unrecognized JSON array format based on first element")
				}
			} else {
				result.Errors = append(result.Errors, "Invalid JSON array elements")
			}
		} else {
			result.Errors = append(result.Errors, "Dataset looks like a JSON array but failed to parse")
		}
	} else if firstChar == '{' {
		// Try parsing as dictionary (CUAD)
		var cuadDataset CUADDataset
		if err := json.Unmarshal(content, &cuadDataset); err == nil && len(cuadDataset.Data) > 0 {
			// Convert CUAD format to flat examples
			for _, data := range cuadDataset.Data {
				for _, para := range data.Paragraphs {
					for _, qa := range para.Qas {
						label := "Unknown"
						if len(qa.Answers) > 0 {
							label = qa.Answers[0].Text
						} else if qa.IsImpossible {
							label = "None"
						}
						examples = append(examples, DatasetExample{
							Text:  "Question: " + qa.Question + "\n\nContext: " + para.Context,
							Label: label,
						})
					}
				}
			}
		} else {
			result.Errors = append(result.Errors, "Unrecognized JSON object format (expected CUAD)")
		}
	} else {
		// Try JSONL
		lines := strings.Split(string(content), "\n")
		examples = make([]DatasetExample, 0, len(lines)/2+1)
		chatExamples = make([]DatasetChat, 0, len(lines)/2+1)

		schemaDetected := ""
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}

			if schemaDetected == "" {
				var first map[string]interface{}
				if json.Unmarshal([]byte(line), &first) == nil {
					if _, ok := first["messages"]; ok {
						schemaDetected = "chat"
						isChat = true
					} else if _, ok := first["instruction"]; ok {
						schemaDetected = "instruction"
					} else if _, ok := first["text"]; ok {
						schemaDetected = "example"
					} else {
						result.Errors = append(result.Errors, fmt.Sprintf("Line %d: unrecognized schema", i+1))
						continue
					}
				} else {
					result.Errors = append(result.Errors, fmt.Sprintf("Line %d: invalid JSON line", i+1))
					continue
				}
			}

			if schemaDetected == "chat" {
				var chatEx DatasetChat
				if err := json.Unmarshal([]byte(line), &chatEx); err == nil && len(chatEx.Messages) > 0 {
					chatExamples = append(chatExamples, chatEx)
				} else {
					result.Errors = append(result.Errors, fmt.Sprintf("Line %d: invalid chat formatting", i+1))
				}
			} else if schemaDetected == "instruction" {
				var instEx DatasetInstruction
				if err := json.Unmarshal([]byte(line), &instEx); err == nil && instEx.Instruction != "" {
					text := instEx.Instruction
					if instEx.Input != "" {
						text += "\n\nInput: " + instEx.Input
					}
					examples = append(examples, DatasetExample{Text: text, Label: instEx.Output})
				} else {
					result.Errors = append(result.Errors, fmt.Sprintf("Line %d: invalid instruction formatting", i+1))
				}
			} else if schemaDetected == "example" {
				var ex DatasetExample
				if err := json.Unmarshal([]byte(line), &ex); err == nil && ex.Text != "" {
					examples = append(examples, ex)
				} else {
					result.Errors = append(result.Errors, fmt.Sprintf("Line %d: invalid text label formatting", i+1))
				}
			}
		}
	}

	totalEx := len(examples)
	if isChat {
		totalEx = len(chatExamples)
	}

	if totalEx == 0 {
		result.Errors = append(result.Errors, "Dataset is empty or format not recognized")
		result.Valid = false
		return result
	}

	// 2. Quality Checks
	totalLength := 0
	uniqueTexts := make(map[string]bool)
	duplicates := 0

	if isChat {
		for i, ex := range chatExamples {
			// Extract full text for quality checks
			fullText := ""
			for _, msg := range ex.Messages {
				fullText += msg.Content + " "
			}

			if strings.TrimSpace(fullText) == "" {
				result.Errors = append(result.Errors, fmt.Sprintf("Example %d: missing message content", i))
			}

			totalLength += len(fullText)
			if uniqueTexts[fullText] {
				duplicates++
			}
			uniqueTexts[fullText] = true

			// For chat, we don't necessarily have a "label" unless we define it
			result.Stats.ClassDist["chat"]++
		}
	} else {
		for i, ex := range examples {
			if strings.TrimSpace(ex.Text) == "" {
				result.Errors = append(result.Errors, fmt.Sprintf("Example %d: missing 'text'", i))
			}
			if strings.TrimSpace(ex.Label) == "" {
				result.Errors = append(result.Errors, fmt.Sprintf("Example %d: missing 'label'", i))
			}

			if len(ex.Text) > 5000 { // Increased limit for legal docs
				result.Warnings = append(result.Warnings, fmt.Sprintf("Example %d: text very long (%d chars)", i, len(ex.Text)))
			}
			totalLength += len(ex.Text)

			if uniqueTexts[ex.Text] {
				duplicates++
			}
			uniqueTexts[ex.Text] = true
			result.Stats.ClassDist[ex.Label]++
		}
	}

	// Stats Calculation
	result.Stats.NumExamples = totalEx
	if totalEx > 0 {
		result.Stats.AvgLength = float64(totalLength) / float64(totalEx)
	}

	// 3. Minimum Examples Rule (Reduced to 10 for dev/testing)
	if totalEx < 10 {
		result.Errors = append(result.Errors, fmt.Sprintf("Insufficient examples: %d < 10", totalEx))
	}

	// 4. Duplicate Limit Rule
	duplicateRate := float64(duplicates) / float64(totalEx)
	if duplicateRate > 0.3 { // Relaxed for legal docs which might have boilerplate
		result.Warnings = append(result.Warnings, fmt.Sprintf("High duplicate rate: %.2f%%", duplicateRate*100))
	}

	// Final Validity Check
	result.Checks["format_valid"] = len(result.Errors) == 0
	result.Checks["min_examples"] = totalEx >= 10
	result.Valid = len(result.Errors) == 0

	return result
}

// ValidateTextDataset validates non-JSON text files (txt, csv, md, pdf, docx)
func ValidateTextDataset(content []byte, formatType string) ValidationResult {
	result := ValidationResult{
		Checks:   make(map[string]bool),
		Warnings: []string{},
		Errors:   []string{},
		Stats: DatasetStats{
			ClassDist: make(map[string]int),
		},
	}

	// For binary formats (pdf, docx), we can't parse content here
	// Just validate that the file is non-empty and let Python extract text
	if formatType == "pdf" || formatType == "docx" {
		if len(content) == 0 {
			result.Errors = append(result.Errors, "File is empty")
			result.Valid = false
			return result
		}

		result.Stats.NumExamples = 1 // Treat the whole document as 1 example
		result.Stats.AvgLength = float64(len(content))
		result.Stats.ClassDist[formatType] = 1
		result.Checks["format_valid"] = true
		result.Checks["min_examples"] = true
		result.Valid = true
		return result
	}

	// For text-based formats (txt, csv, md), count lines
	text := string(content)
	lines := strings.Split(text, "\n")

	// Filter empty lines
	nonEmptyLines := 0
	totalLength := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			nonEmptyLines++
			totalLength += len(trimmed)
		}
	}

	if nonEmptyLines == 0 {
		result.Errors = append(result.Errors, "File is empty or contains only blank lines")
		result.Valid = false
		return result
	}

	result.Stats.NumExamples = nonEmptyLines
	result.Stats.AvgLength = float64(totalLength) / float64(nonEmptyLines)
	result.Stats.ClassDist[formatType] = nonEmptyLines

	if nonEmptyLines < 1 {
		result.Errors = append(result.Errors, fmt.Sprintf("File has no content: %d lines", nonEmptyLines))
	}

	result.Checks["format_valid"] = len(result.Errors) == 0
	result.Checks["min_examples"] = true // No minimum for text docs
	result.Valid = len(result.Errors) == 0

	return result
}

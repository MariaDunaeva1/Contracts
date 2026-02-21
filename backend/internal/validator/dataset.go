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
	// Try parsing as simple text/label first
	err := json.Unmarshal(content, &examples)
	isChat := false

	if err != nil {
		// Try parsing as chat/messages
		err = json.Unmarshal(content, &chatExamples)
		if err != nil {
			// Try parsing as instruction/input/output
			var instructExamples []DatasetInstruction
			err = json.Unmarshal(content, &instructExamples)
			if err == nil && len(instructExamples) > 0 && instructExamples[0].Instruction != "" {
				// Convert to standard examples for internal use
				examples = make([]DatasetExample, len(instructExamples))
				for i, ie := range instructExamples {
					text := ie.Instruction
					if ie.Input != "" {
						text += "\n\nInput: " + ie.Input
					}
					examples[i] = DatasetExample{
						Text:  text,
						Label: ie.Output,
					}
				}
			} else {
				// Try checking for SQuAD/CUAD dictionary format ("data": [...])
				var cuadDataset CUADDataset
				if err = json.Unmarshal(content, &cuadDataset); err == nil && len(cuadDataset.Data) > 0 {
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

								// We map Question + Context -> Extracted Answer
								examples = append(examples, DatasetExample{
									Text:  "Question: " + qa.Question + "\n\nContext: " + para.Context,
									Label: label,
								})
							}
						}
					}
				} else {
					// Try JSONL if it fails as a single array or dict
					lines := strings.Split(string(content), "\n")
					examples = []DatasetExample{}
					chatExamples = []DatasetChat{}

					for i, line := range lines {
						if strings.TrimSpace(line) == "" {
							continue
						}
						var ex DatasetExample
						var chatEx DatasetChat
						var instEx DatasetInstruction

						if json.Unmarshal([]byte(line), &ex) == nil && ex.Text != "" {
							examples = append(examples, ex)
						} else if json.Unmarshal([]byte(line), &chatEx) == nil && len(chatEx.Messages) > 0 {
							chatExamples = append(chatExamples, chatEx)
							isChat = true
						} else if json.Unmarshal([]byte(line), &instEx) == nil && instEx.Instruction != "" {
							text := instEx.Instruction
							if instEx.Input != "" {
								text += "\n\nInput: " + instEx.Input
							}
							examples = append(examples, DatasetExample{
								Text:  text,
								Label: instEx.Output,
							})
						} else {
							result.Errors = append(result.Errors, fmt.Sprintf("Line %d: invalid JSON structure", i+1))
						}
					}
				}
			}
		} else {
			isChat = true
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

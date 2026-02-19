package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Model struct {
	gorm.Model
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	BaseModel        string         `json:"base_model"`
	Type             string         `json:"type"`
	JobID            *uint          `json:"job_id"`
	Job              *Job           `json:"job,omitempty"`
	StoragePath      string         `json:"storage_path"`
	LoRAAdaptersPath string         `json:"lora_adapters_path"`
	GGUFPath         string         `json:"gguf_path"`
	Files            datatypes.JSON `json:"files"`
	TrainingMetrics  datatypes.JSON `json:"training_metrics"`
	EvalResults      datatypes.JSON `json:"eval_results"`
	Status           string         `json:"status"`
	TotalSize        int64          `json:"total_size"`
}

type Evaluation struct {
	gorm.Model
	ModelID       uint           `json:"model_id"`
	JobID         *uint          `json:"job_id"`
	Status        string         `json:"status"`
	TestSetPath   string         `json:"test_set_path"`
	BaseModelName string         `json:"base_model_name"`
	FineTunedName string         `json:"fine_tuned_name"`
	Results       datatypes.JSON `json:"results"`
	Examples      datatypes.JSON `json:"examples"`
	StartedAt     *time.Time     `json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	ErrorMessage  string         `json:"error_message"`
}

type LogEntry struct {
	gorm.Model
	JobID     uint      `json:"job_id" gorm:"index"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
}

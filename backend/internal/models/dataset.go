package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Dataset struct {
	gorm.Model
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	FilePath          string         `json:"file_path"` // Path in MinIO
	Type              string         `json:"type"`      // csv, json
	NumExamples       int            `json:"num_examples"`
	AvgLength         float64        `json:"avg_length"`
	ValidationStatus  string         `json:"validation_status"` // valid, warning, error
	ValidationDetails datatypes.JSON `json:"validation_details"`
}

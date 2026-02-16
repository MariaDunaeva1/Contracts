package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	DatasetID      uint           `json:"dataset_id"`
	Dataset        Dataset        `json:"dataset"`
	Status         string         `json:"status"` // pending, starting, running, completed, failed, cancelled
	Configuration  datatypes.JSON `json:"configuration"`
	Metrics        datatypes.JSON `json:"metrics"`
	KaggleKernelID string         `json:"kaggle_kernel_id"`
}

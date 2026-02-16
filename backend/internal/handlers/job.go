package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"finetune-studio/internal/database"
	"finetune-studio/internal/models"
	"finetune-studio/internal/worker"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type CreateJobRequest struct {
	DatasetID     uint                   `json:"dataset_id" binding:"required"`
	Configuration map[string]interface{} `json:"configuration"`
}

// CreateJob handles POST /api/v1/jobs
func CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate Dataset exists
	var dataset models.Dataset
	if err := database.DB.First(&dataset, req.DatasetID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dataset not found"})
		return
	}

	// Create Job in DB
	configJSON, _ := json.Marshal(req.Configuration)
	job := models.Job{
		DatasetID:     req.DatasetID,
		Status:        "pending",
		Configuration: datatypes.JSON(configJSON),
		Metrics:       datatypes.JSON([]byte("{}")),
	}

	if err := database.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		return
	}

	// Enqueue Job
	select {
	case worker.Pool.JobQueue <- job.ID:
		c.JSON(http.StatusCreated, job)
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Job queue is full"})
	}
}

// ListJobs handles GET /api/v1/jobs
func ListJobs(c *gin.Context) {
	var jobs []models.Job
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	query := database.DB.Model(&models.Job{}).Preload("Dataset")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  jobs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetJob handles GET /api/v1/jobs/:id
func GetJob(c *gin.Context) {
	id := c.Param("id")
	var job models.Job

	if err := database.DB.Preload("Dataset").First(&job, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// CancelJob handles DELETE /api/v1/jobs/:id
func CancelJob(c *gin.Context) {
	id := c.Param("id")
	var job models.Job

	if err := database.DB.First(&job, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	if job.Status == "completed" || job.Status == "failed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel a completed or failed job"})
		return
	}

	job.Status = "cancelled"
	job.Metrics = datatypes.JSON([]byte(fmt.Sprintf(`{"cancelled_at": "%s"}`, time.Now().Format(time.RFC3339))))
	database.DB.Save(&job)

	c.JSON(http.StatusOK, gin.H{"message": "Job marked as cancelled"})
}

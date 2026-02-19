package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"finetune-studio/internal/database"
	"finetune-studio/internal/models"
	"finetune-studio/internal/storage"

	"github.com/gin-gonic/gin"
)

type ModelHandler struct {
	modelStorage *storage.ModelStorage
}

func NewModelHandler(modelStorage *storage.ModelStorage) *ModelHandler {
	return &ModelHandler{
		modelStorage: modelStorage,
	}
}

// ListModels handles GET /api/v1/models
func (h *ModelHandler) ListModels(c *gin.Context) {
	var modelsList []models.Model
	
	query := database.DB.Model(&models.Model{})
	
	// Try to preload relationships, but don't fail if they don't exist
	query = query.Preload("Job").Preload("Job.Dataset")

	// Filters
	if baseModel := c.Query("base_model"); baseModel != "" {
		query = query.Where("base_model = ?", baseModel)
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		query = query.Where("created_at >= ?", dateFrom)
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		query = query.Where("created_at <= ?", dateTo)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	if err := query.Count(&total).Error; err != nil {
		// If table doesn't exist or other error, return empty list
		c.JSON(http.StatusOK, gin.H{
			"data":  []models.Model{},
			"total": 0,
			"page":  page,
			"limit": limit,
		})
		return
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&modelsList).Error; err != nil {
		// Return empty list instead of error
		c.JSON(http.StatusOK, gin.H{
			"data":  []models.Model{},
			"total": 0,
			"page":  page,
			"limit": limit,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  modelsList,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetModel handles GET /api/v1/models/:id
func (h *ModelHandler) GetModel(c *gin.Context) {
	id := c.Param("id")
	var model models.Model

	if err := database.DB.Preload("Job").Preload("Job.Dataset").First(&model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	ctx := context.Background()

	// Generate presigned URLs for download (24h expiry)
	downloadLinks := make(map[string]string)
	
	if model.LoRAAdaptersPath != "" && h.modelStorage.FileExists(ctx, "models", model.LoRAAdaptersPath) {
		url, err := h.modelStorage.GetPresignedURL(ctx, "models", model.LoRAAdaptersPath, 24*time.Hour)
		if err == nil {
			downloadLinks["lora_adapters"] = url
		}
	}

	if model.GGUFPath != "" && h.modelStorage.FileExists(ctx, "models", model.GGUFPath) {
		url, err := h.modelStorage.GetPresignedURL(ctx, "models", model.GGUFPath, 24*time.Hour)
		if err == nil {
			downloadLinks["gguf"] = url
		}
	}

	// Parse files JSON
	var files map[string]interface{}
	if model.Files != nil {
		json.Unmarshal(model.Files, &files)
	}

	response := gin.H{
		"model":          model,
		"download_links": downloadLinks,
		"files":          files,
	}

	c.JSON(http.StatusOK, response)
}

// DownloadModel handles GET /api/v1/models/:id/download
func (h *ModelHandler) DownloadModel(c *gin.Context) {
	id := c.Param("id")
	var model models.Model

	if err := database.DB.First(&model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	if model.Status != "ready" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model is not ready for download"})
		return
	}

	// Set headers for ZIP download
	filename := fmt.Sprintf("model-%d-%s.zip", model.ID, time.Now().Format("20060102"))
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Transfer-Encoding", "chunked")

	// Stream ZIP directly to response
	ctx := context.Background()
	if err := h.modelStorage.StreamModelZIP(ctx, model.StoragePath, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create ZIP: %v", err)})
		return
	}
}

// CreateModel handles POST /api/v1/models
func (h *ModelHandler) CreateModel(c *gin.Context) {
	var model models.Model
	if err := c.ShouldBindJSON(&model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default status
	if model.Status == "" {
		model.Status = "uploading"
	}

	// Calculate total size if storage path is provided
	if model.StoragePath != "" {
		ctx := context.Background()
		size, err := h.modelStorage.CalculateTotalSize(ctx, model.StoragePath)
		if err == nil {
			model.TotalSize = size
		}
	}

	if err := database.DB.Create(&model).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create model"})
		return
	}

	c.JSON(http.StatusCreated, model)
}

// UpdateModel handles PUT /api/v1/models/:id
func (h *ModelHandler) UpdateModel(c *gin.Context) {
	id := c.Param("id")
	var model models.Model

	if err := database.DB.First(&model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	var updates models.Model
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update allowed fields
	if updates.Name != "" {
		model.Name = updates.Name
	}
	if updates.Description != "" {
		model.Description = updates.Description
	}
	if updates.Status != "" {
		model.Status = updates.Status
	}
	if updates.TrainingMetrics != nil {
		model.TrainingMetrics = updates.TrainingMetrics
	}
	if updates.EvalResults != nil {
		model.EvalResults = updates.EvalResults
	}

	if err := database.DB.Save(&model).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update model"})
		return
	}

	c.JSON(http.StatusOK, model)
}

// DeleteModel handles DELETE /api/v1/models/:id
func (h *ModelHandler) DeleteModel(c *gin.Context) {
	id := c.Param("id")
	var model models.Model

	if err := database.DB.First(&model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	// TODO: Delete files from MinIO

	if err := database.DB.Delete(&model).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete model"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Model deleted successfully"})
}

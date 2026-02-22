package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"finetune-studio/internal/database"
	"finetune-studio/internal/logger"
	"finetune-studio/internal/models"
	"finetune-studio/internal/storage"
	"finetune-studio/internal/validator"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// UploadDataset handles POST /api/v1/datasets
func UploadDataset(c *gin.Context) {
	// 1. Parse Multipart Form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	name := c.PostForm("name")
	if name == "" {
		name = header.Filename
	}
	description := c.PostForm("description")

	// 2. Read Content
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// 3. Validate extension
	ext := filepath.Ext(header.Filename)
	allowedExts := map[string]string{
		".json":  "application/json",
		".jsonl": "application/json",
		".txt":   "text/plain",
		".csv":   "text/csv",
		".md":    "text/markdown",
		".pdf":   "application/pdf",
		".docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}

	contentType, extAllowed := allowedExts[ext]
	if !extAllowed {
		logger.Warn("Unsupported file extension", zap.String("ext", ext))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Supported formats: .json, .jsonl, .txt, .csv, .md, .pdf, .docx"})
		return
	}

	logger.Info("Validating dataset", zap.String("filename", header.Filename), zap.Int("size", len(content)))

	// Determine dataset type and validate accordingly
	isJSON := ext == ".json" || ext == ".jsonl"
	var validationResult validator.ValidationResult
	var datasetType string

	if isJSON {
		datasetType = "json"
		validationResult = validator.ValidateDataset(content, "json")
	} else {
		datasetType = ext[1:] // Remove the leading dot
		validationResult = validator.ValidateTextDataset(content, datasetType)
	}

	if !validationResult.Valid {
		logger.Warn("Dataset validation failed", zap.Any("errors", validationResult.Errors))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": validationResult,
		})
		return
	}

	// 4. Upload to MinIO
	objectName := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	ctx := c.Request.Context()

	logger.Info("Uploading to MinIO", zap.String("object", objectName))
	_, err = storage.Client.PutObject(ctx, "datasets", objectName, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logger.Error("Failed to upload to storage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to storage", "details": err.Error()})
		return
	}

	// 5. Save to DB
	logger.Info("Saving dataset metadata to DB", zap.String("name", name))
	validationJSON, _ := json.Marshal(validationResult)
	dataset := models.Dataset{
		Name:              name,
		Description:       description,
		FilePath:          objectName,
		Type:              datasetType,
		NumExamples:       validationResult.Stats.NumExamples,
		AvgLength:         validationResult.Stats.AvgLength,
		ValidationStatus:  "valid",
		ValidationDetails: datatypes.JSON(validationJSON),
	}

	if len(validationResult.Warnings) > 0 {
		dataset.ValidationStatus = "warning"
	}

	if err := database.DB.Create(&dataset).Error; err != nil {
		logger.Error("Failed to save dataset to DB", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
		return
	}

	logger.Info("Dataset uploaded successfully", zap.Uint("id", dataset.ID))
	c.JSON(http.StatusCreated, dataset)
}

// ListDatasets handles GET /api/v1/datasets
func ListDatasets(c *gin.Context) {
	var datasets []models.Dataset
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	query := database.DB.Model(&models.Dataset{})

	if name := c.Query("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	var total int64
	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&datasets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch datasets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  datasets,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetDataset handles GET /api/v1/datasets/:id
func GetDataset(c *gin.Context) {
	id := c.Param("id")
	var dataset models.Dataset

	if err := database.DB.First(&dataset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dataset not found"})
		return
	}

	// Preview from MinIO
	obj, err := storage.Client.GetObject(context.Background(), "datasets", dataset.FilePath, minio.GetObjectOptions{})
	if err == nil {
		defer obj.Close()
		// Read first 1KB for preview
		buf := make([]byte, 1024)
		n, _ := obj.Read(buf)
		// Try to parse partial json or just return string
		c.JSON(http.StatusOK, gin.H{
			"dataset": dataset,
			"preview": string(buf[:n]),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dataset": dataset, "preview": "Not available"})
}

// DeleteDataset handles DELETE /api/v1/datasets/:id
func DeleteDataset(c *gin.Context) {
	id := c.Param("id")
	var dataset models.Dataset

	if err := database.DB.First(&dataset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dataset not found"})
		return
	}

	// Soft delete in DB
	if err := database.DB.Delete(&dataset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dataset"})
		return
	}

	// Async delete from MinIO
	go func(filePath string) {
		err := storage.Client.RemoveObject(context.Background(), "datasets", filePath, minio.RemoveObjectOptions{})
		if err != nil {
			fmt.Printf("Failed to remove object %s: %v\n", filePath, err)
		}
	}(dataset.FilePath)

	c.JSON(http.StatusOK, gin.H{"message": "Dataset deleted"})
}

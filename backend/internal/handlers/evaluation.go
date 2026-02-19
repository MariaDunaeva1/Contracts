package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"finetune-studio/internal/database"
	"finetune-studio/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type EvaluationHandler struct{}

func NewEvaluationHandler() *EvaluationHandler {
	return &EvaluationHandler{}
}

type CreateEvaluationRequest struct {
	TestSetPath   string `json:"test_set_path"`
	BaseModelName string `json:"base_model_name"`
}

// CreateEvaluation handles POST /api/v1/models/:id/evaluate
func (h *EvaluationHandler) CreateEvaluation(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := strconv.ParseUint(modelIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID"})
		return
	}

	// Verify model exists
	var model models.Model
	if err := database.DB.First(&model, modelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	var req CreateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.BaseModelName == "" {
		req.BaseModelName = model.BaseModel
	}

	// Create evaluation record
	now := time.Now()
	evaluation := models.Evaluation{
		ModelID:       uint(modelID),
		JobID:         model.JobID,
		Status:        "pending",
		TestSetPath:   req.TestSetPath,
		BaseModelName: req.BaseModelName,
		FineTunedName: model.Name,
		StartedAt:     &now,
		Results:       datatypes.JSON([]byte("{}")),
		Examples:      datatypes.JSON([]byte("[]")),
	}

	if err := database.DB.Create(&evaluation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create evaluation"})
		return
	}

	// TODO: Trigger evaluation job asynchronously
	// For now, return the evaluation ID
	c.JSON(http.StatusCreated, gin.H{
		"evaluation_id": evaluation.ID,
		"status":        evaluation.Status,
		"message":       "Evaluation job created",
	})
}

// GetEvaluation handles GET /api/v1/evaluations/:id
func (h *EvaluationHandler) GetEvaluation(c *gin.Context) {
	id := c.Param("id")
	var evaluation models.Evaluation

	if err := database.DB.Preload("Model").First(&evaluation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evaluation not found"})
		return
	}

	// Parse JSON fields
	var results map[string]interface{}
	var examples []map[string]interface{}

	if evaluation.Results != nil {
		json.Unmarshal(evaluation.Results, &results)
	}
	if evaluation.Examples != nil {
		json.Unmarshal(evaluation.Examples, &examples)
	}

	response := gin.H{
		"id":              evaluation.ID,
		"model_id":        evaluation.ModelID,
		"status":          evaluation.Status,
		"test_set_path":   evaluation.TestSetPath,
		"base_model_name": evaluation.BaseModelName,
		"fine_tuned_name": evaluation.FineTunedName,
		"results":         results,
		"examples":        examples,
		"started_at":      evaluation.StartedAt,
		"completed_at":    evaluation.CompletedAt,
		"error_message":   evaluation.ErrorMessage,
	}

	c.JSON(http.StatusOK, response)
}

// ListEvaluations handles GET /api/v1/evaluations
func (h *EvaluationHandler) ListEvaluations(c *gin.Context) {
	var evaluations []models.Evaluation

	query := database.DB.Model(&models.Evaluation{}).Preload("Model")

	// Filter by model_id
	if modelID := c.Query("model_id"); modelID != "" {
		query = query.Where("model_id = ?", modelID)
	}

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&evaluations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch evaluations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  evaluations,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// UpdateEvaluation handles PUT /api/v1/evaluations/:id
func (h *EvaluationHandler) UpdateEvaluation(c *gin.Context) {
	id := c.Param("id")
	var evaluation models.Evaluation

	if err := database.DB.First(&evaluation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evaluation not found"})
		return
	}

	var updates struct {
		Status       string                 `json:"status"`
		Results      map[string]interface{} `json:"results"`
		Examples     []map[string]interface{} `json:"examples"`
		ErrorMessage string                 `json:"error_message"`
	}

	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if updates.Status != "" {
		evaluation.Status = updates.Status
		
		if updates.Status == "completed" || updates.Status == "failed" {
			now := time.Now()
			evaluation.CompletedAt = &now
		}
	}

	if updates.Results != nil {
		resultsJSON, _ := json.Marshal(updates.Results)
		evaluation.Results = datatypes.JSON(resultsJSON)
	}

	if updates.Examples != nil {
		examplesJSON, _ := json.Marshal(updates.Examples)
		evaluation.Examples = datatypes.JSON(examplesJSON)
	}

	if updates.ErrorMessage != "" {
		evaluation.ErrorMessage = updates.ErrorMessage
	}

	if err := database.DB.Save(&evaluation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update evaluation"})
		return
	}

	c.JSON(http.StatusOK, evaluation)
}

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ContractHandler struct {
	pythonServiceURL string
	httpClient       *http.Client
}

func NewContractHandler(pythonURL string) *ContractHandler {
	return &ContractHandler{
		pythonServiceURL: pythonURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Long timeout for LLM processing
		},
	}
}

// AnalyzeContractRequest represents the request to analyze a contract
type AnalyzeContractRequest struct {
	ContractText string `json:"contract_text" binding:"required"`
	ContractName string `json:"contract_name"`
	ContractID   string `json:"contract_id"`
}

// SearchClausesRequest represents a semantic search request
type SearchClausesRequest struct {
	Query   string            `json:"query" binding:"required"`
	TopK    int               `json:"top_k"`
	Filters map[string]string `json:"filters"`
}

// POST /api/v1/contracts/analyze
// Analyze contract with RAG - extract clauses, compare with historical, assess risk
func (h *ContractHandler) AnalyzeContract(c *gin.Context) {
	var req AnalyzeContractRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default contract name if not provided
	if req.ContractName == "" {
		req.ContractName = "New Contract"
	}

	// Call Python RAG service
	result, err := h.callPythonService("/analyze", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to analyze contract",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// POST /api/v1/clauses/search
// Semantic search for similar clauses
func (h *ContractHandler) SearchClauses(c *gin.Context) {
	var req SearchClausesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default top_k if not provided
	if req.TopK == 0 {
		req.TopK = 5
	}

	// Call Python RAG service
	result, err := h.callPythonService("/search", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search clauses",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /api/v1/contracts/:id/similar
// Find similar contracts based on contract ID
func (h *ContractHandler) FindSimilarContracts(c *gin.Context) {
	contractID := c.Param("id")
	topK := c.DefaultQuery("top_k", "5")

	// Get clauses for this contract
	clausesResult, err := h.callPythonService(
		fmt.Sprintf("/contracts/%s/clauses", contractID),
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get contract clauses",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contract_id": contractID,
		"clauses":     clausesResult,
		"top_k":       topK,
	})
}

// GET /api/v1/contracts/:id/clauses
// Get all clauses for a contract
func (h *ContractHandler) GetContractClauses(c *gin.Context) {
	contractID := c.Param("id")

	result, err := h.callPythonService(
		fmt.Sprintf("/contracts/%s/clauses", contractID),
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get clauses",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DELETE /api/v1/contracts/:id/index
// Remove contract from vector database
func (h *ContractHandler) DeleteContractIndex(c *gin.Context) {
	contractID := c.Param("id")

	result, err := h.callPythonService(
		fmt.Sprintf("/contracts/%s", contractID),
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete contract",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /api/v1/rag/health
// Check RAG service health
func (h *ContractHandler) RAGHealth(c *gin.Context) {
	result, err := h.callPythonService("/health", nil)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /api/v1/rag/stats
// Get RAG service statistics
func (h *ContractHandler) RAGStats(c *gin.Context) {
	result, err := h.callPythonService("/stats", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Helper function to call Python service
func (h *ContractHandler) callPythonService(endpoint string, payload interface{}) (map[string]interface{}, error) {
	url := h.pythonServiceURL + endpoint

	var req *http.Request
	var err error

	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Python service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Python service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

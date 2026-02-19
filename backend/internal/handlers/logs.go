package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"finetune-studio/internal/models"
	"finetune-studio/internal/services/logs"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService *logs.LogService
}

func NewLogHandler(logService *logs.LogService) *LogHandler {
	return &LogHandler{
		logService: logService,
	}
}

// StreamLogs handles GET /api/v1/jobs/:id/logs (SSE)
func (h *LogHandler) StreamLogs(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Create ticker for periodic updates
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Track last sent timestamp to avoid duplicates
	lastTimestamp := time.Now().Add(-1 * time.Hour)

	// Send initial logs
	initialLogs, err := h.logService.GetLogsFromDB(uint(jobID), 50)
	if err == nil && len(initialLogs) > 0 {
		formatted := h.logService.FormatLogsForSSE(initialLogs)
		fmt.Fprintf(c.Writer, "data: %s\n\n", formatted)
		c.Writer.Flush()
		
		if len(initialLogs) > 0 {
			lastTimestamp = initialLogs[0].Timestamp
		}
	}

	// Stream updates
	for {
		select {
		case <-ticker.C:
			// Aggregate new logs from MinIO
			if err := h.logService.AggregateLogsFromMinIO(c.Request.Context(), uint(jobID)); err != nil {
				fmt.Fprintf(c.Writer, "event: error\ndata: Failed to fetch logs\n\n")
				c.Writer.Flush()
				continue
			}

			// Fetch new logs since last timestamp
			newLogs, err := h.logService.FetchLatestLogs(c.Request.Context(), uint(jobID), lastTimestamp)
			if err != nil {
				fmt.Fprintf(c.Writer, "event: error\ndata: %v\n\n", err)
				c.Writer.Flush()
				continue
			}

			if len(newLogs) > 0 {
				formatted := h.logService.FormatLogsForSSE(newLogs)
				fmt.Fprintf(c.Writer, "data: %s\n\n", formatted)
				c.Writer.Flush()
				
				// Update last timestamp
				for _, log := range newLogs {
					if log.Timestamp.After(lastTimestamp) {
						lastTimestamp = log.Timestamp
					}
				}
			} else {
				// Send heartbeat
				fmt.Fprintf(c.Writer, ": heartbeat\n\n")
				c.Writer.Flush()
			}

		case <-c.Request.Context().Done():
			return
		}
	}
}

// GetLogs handles GET /api/v1/jobs/:id/logs (regular JSON)
func (h *LogHandler) GetLogs(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	
	// Aggregate latest logs from MinIO
	h.logService.AggregateLogsFromMinIO(c.Request.Context(), uint(jobID))

	// Get logs from DB
	logEntries, err := h.logService.GetLogsFromDB(uint(jobID), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id": jobID,
		"logs":   logEntries,
		"count":  len(logEntries),
	})
}

// CreateLogEntry handles POST /api/v1/jobs/:id/logs (for kernel to push logs)
func (h *LogHandler) CreateLogEntry(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var entry models.LogEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry.JobID = uint(jobID)
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	if err := h.logService.SaveLogToDB(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save log"})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

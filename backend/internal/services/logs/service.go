package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"finetune-studio/internal/database"
	"finetune-studio/internal/models"

	"github.com/minio/minio-go/v7"
)

type LogService struct {
	minioClient *minio.Client
}

func NewLogService(minioClient *minio.Client) *LogService {
	return &LogService{
		minioClient: minioClient,
	}
}

// FetchLatestLogs retrieves the latest logs for a job from MinIO
func (s *LogService) FetchLatestLogs(ctx context.Context, jobID uint, since time.Time) ([]models.LogEntry, error) {
	// Path pattern: logs/{jobID}/log_{timestamp}.json
	prefix := fmt.Sprintf("%d/", jobID)
	
	objectCh := s.minioClient.ListObjects(ctx, "logs", minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var logs []models.LogEntry
	
	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Error listing logs: %v", object.Err)
			continue
		}

		// Skip if object is older than 'since'
		if object.LastModified.Before(since) {
			continue
		}

		// Download and parse log file
		obj, err := s.minioClient.GetObject(ctx, "logs", object.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Printf("Error getting log object %s: %v", object.Key, err)
			continue
		}

		data, err := io.ReadAll(obj)
		obj.Close()
		if err != nil {
			log.Printf("Error reading log object %s: %v", object.Key, err)
			continue
		}

		// Parse JSON logs (array or single object)
		var entries []models.LogEntry
		if err := json.Unmarshal(data, &entries); err != nil {
			// Try single object
			var entry models.LogEntry
			if err := json.Unmarshal(data, &entry); err != nil {
				log.Printf("Error parsing log object %s: %v", object.Key, err)
				continue
			}
			entries = []models.LogEntry{entry}
		}

		logs = append(logs, entries...)
	}

	return logs, nil
}

// GetLogsFromDB retrieves logs from database
func (s *LogService) GetLogsFromDB(jobID uint, limit int) ([]models.LogEntry, error) {
	var logs []models.LogEntry
	err := database.DB.Where("job_id = ?", jobID).
		Order("timestamp desc").
		Limit(limit).
		Find(&logs).Error
	
	return logs, err
}

// SaveLogToDB saves a log entry to database
func (s *LogService) SaveLogToDB(entry models.LogEntry) error {
	return database.DB.Create(&entry).Error
}

// AggregateLogsFromMinIO reads logs from MinIO and saves to DB
func (s *LogService) AggregateLogsFromMinIO(ctx context.Context, jobID uint) error {
	// Get last processed timestamp from DB
	var lastLog models.LogEntry
	database.DB.Where("job_id = ?", jobID).Order("timestamp desc").First(&lastLog)
	
	since := time.Time{}
	if lastLog.ID > 0 {
		since = lastLog.Timestamp
	}

	logs, err := s.FetchLatestLogs(ctx, jobID, since)
	if err != nil {
		return err
	}

	// Save new logs to DB
	for _, entry := range logs {
		entry.JobID = jobID
		if err := s.SaveLogToDB(entry); err != nil {
			log.Printf("Error saving log to DB: %v", err)
		}
	}

	return nil
}

// FormatLogsForSSE formats logs for Server-Sent Events
func (s *LogService) FormatLogsForSSE(logs []models.LogEntry) string {
	var lines []string
	for _, entry := range logs {
		timestamp := entry.Timestamp.Format("15:04:05")
		line := fmt.Sprintf("[%s] [%s] %s", timestamp, strings.ToUpper(entry.Level), entry.Message)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

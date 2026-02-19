package storage

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
)

type ModelStorage struct {
	client *minio.Client
}

func NewModelStorage(client *minio.Client) *ModelStorage {
	return &ModelStorage{client: client}
}

// GetPresignedURL generates a presigned URL for downloading a file
func (s *ModelStorage) GetPresignedURL(ctx context.Context, bucket, objectPath string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, bucket, objectPath, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// ListModelFiles lists all files for a model
func (s *ModelStorage) ListModelFiles(ctx context.Context, modelPath string) ([]minio.ObjectInfo, error) {
	var files []minio.ObjectInfo
	
	objectCh := s.client.ListObjects(ctx, "models", minio.ListObjectsOptions{
		Prefix:    modelPath,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		files = append(files, object)
	}

	return files, nil
}

// GetFileSize gets the size of a file in MinIO
func (s *ModelStorage) GetFileSize(ctx context.Context, bucket, objectPath string) (int64, error) {
	info, err := s.client.StatObject(ctx, bucket, objectPath, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

// GetJSON retrieves and parses a JSON file from MinIO
func (s *ModelStorage) GetJSON(ctx context.Context, bucket, objectPath string, target interface{}) error {
	obj, err := s.client.GetObject(ctx, bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return fmt.Errorf("failed to read object: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}

// StreamModelZIP creates a ZIP archive of model files and streams it
func (s *ModelStorage) StreamModelZIP(ctx context.Context, modelPath string, writer io.Writer) error {
	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	// List all files for the model
	files, err := s.ListModelFiles(ctx, modelPath)
	if err != nil {
		return fmt.Errorf("failed to list model files: %w", err)
	}

	for _, file := range files {
		if file.Size == 0 {
			continue // Skip directories
		}

		// Get file from MinIO
		obj, err := s.client.GetObject(ctx, "models", file.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Printf("Error getting file %s: %v", file.Key, err)
			continue
		}

		// Create file in ZIP
		zipFile, err := zipWriter.Create(file.Key)
		if err != nil {
			obj.Close()
			log.Printf("Error creating zip entry %s: %v", file.Key, err)
			continue
		}

		// Copy content
		if _, err := io.Copy(zipFile, obj); err != nil {
			obj.Close()
			log.Printf("Error copying file %s to zip: %v", file.Key, err)
			continue
		}

		obj.Close()
	}

	return nil
}

// CalculateTotalSize calculates the total size of all files in a model
func (s *ModelStorage) CalculateTotalSize(ctx context.Context, modelPath string) (int64, error) {
	files, err := s.ListModelFiles(ctx, modelPath)
	if err != nil {
		return 0, err
	}

	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}

	return totalSize, nil
}

// FileExists checks if a file exists in MinIO
func (s *ModelStorage) FileExists(ctx context.Context, bucket, objectPath string) bool {
	_, err := s.client.StatObject(ctx, bucket, objectPath, minio.StatObjectOptions{})
	return err == nil
}

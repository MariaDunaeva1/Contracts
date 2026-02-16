package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"finetune-studio/internal/database"
	"finetune-studio/internal/models"
	"finetune-studio/internal/services/kaggle"
	"finetune-studio/internal/storage"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"gorm.io/datatypes"
)

type WorkerPool struct {
	Workers       int
	JobQueue      chan uint
	Quit          chan bool
	KaggleService *kaggle.Service
}

// Global instance
var Pool *WorkerPool

func NewWorkerPool(workers int, kaggleSvc *kaggle.Service) *WorkerPool {
	return &WorkerPool{
		Workers:       workers,
		JobQueue:      make(chan uint, 100),
		Quit:          make(chan bool),
		KaggleService: kaggleSvc,
	}
}

func (w *WorkerPool) Start() {
	for i := 0; i < w.Workers; i++ {
		go w.worker(i)
	}
	log.Printf("🚀 Worker Pool started with %d workers", w.Workers)
}

func (w *WorkerPool) worker(id int) {
	for jobID := range w.JobQueue {
		w.processJob(id, jobID)
	}
}

func (w *WorkerPool) processJob(workerID int, jobID uint) {
	log.Printf("[Worker %d] Processing Job %d", workerID, jobID)

	var job models.Job
	if err := database.DB.Preload("Dataset").First(&job, jobID).Error; err != nil {
		log.Printf("[Worker %d] Job %d not found: %v", workerID, jobID, err)
		return
	}

	// 1. Mark as Starting
	updateJobStatus(&job, "starting")

	// Check if Kaggle is configured
	if os.Getenv("KAGGLE_USERNAME") == "" || os.Getenv("KAGGLE_KEY") == "" {
		log.Printf("[Worker %d] Kaggle not configured, running in SIMULATION mode", workerID)
		w.processJobSimulated(workerID, &job)
		return
	}

	// === KAGGLE MODE ===
	log.Printf("[Worker %d] Running Job %d via Kaggle", workerID, jobID)
	w.processJobKaggle(workerID, &job)
}

func (w *WorkerPool) processJobKaggle(workerID int, job *models.Job) {
	// 1. Download dataset from MinIO to temp file
	tmpDir := fmt.Sprintf("/tmp/job_%d", job.ID)
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	datasetFile := fmt.Sprintf("%s/dataset.json", tmpDir)
	obj, err := storage.Client.GetObject(context.Background(), "datasets", job.Dataset.FilePath, minio.GetObjectOptions{})
	if err != nil {
		updateJobFailed(job, fmt.Sprintf("Failed to download dataset from MinIO: %v", err))
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(obj)
	obj.Close()
	if err := os.WriteFile(datasetFile, buf.Bytes(), 0644); err != nil {
		updateJobFailed(job, fmt.Sprintf("Failed to write dataset: %v", err))
		return
	}
	log.Printf("[Worker %d] Dataset downloaded to %s (%d bytes)", workerID, datasetFile, buf.Len())

	// 2. Upload dataset to Kaggle
	updateJobMetrics(job, map[string]interface{}{"stage": "uploading_dataset"})
	datasetName := fmt.Sprintf("job-%d-dataset", job.ID)
	datasetRef, err := w.KaggleService.CreateDataset(datasetName, datasetFile)
	if err != nil {
		updateJobFailed(job, fmt.Sprintf("Failed to upload dataset to Kaggle: %v", err))
		return
	}
	log.Printf("[Worker %d] Dataset uploaded to Kaggle: %s", workerID, datasetRef)

	// 3. Read notebook template and push kernel
	updateJobStatus(job, "running")
	updateJobMetrics(job, map[string]interface{}{"stage": "pushing_kernel", "kaggle_dataset": datasetRef})

	notebookBytes, err := os.ReadFile("/app/templates/finetune-kernel.ipynb")
	if err != nil {
		updateJobFailed(job, fmt.Sprintf("Failed to read notebook template: %v", err))
		return
	}

	kernelSlug := fmt.Sprintf("finetune-job-%d", job.ID)
	kernelRef, err := w.KaggleService.PushKernel(kernelSlug, notebookBytes, []string{datasetRef})
	if err != nil {
		updateJobFailed(job, fmt.Sprintf("Failed to push kernel: %v", err))
		return
	}
	job.KaggleKernelID = kernelRef
	database.DB.Save(job)
	log.Printf("[Worker %d] Kernel pushed: %s", workerID, kernelRef)

	// 4. Poll kernel status
	updateJobMetrics(job, map[string]interface{}{"stage": "training", "kernel_ref": kernelRef})
	for {
		// Check cancellation
		database.DB.First(job, job.ID)
		if job.Status == "cancelled" {
			log.Printf("[Worker %d] Job %d cancelled", workerID, job.ID)
			return
		}

		status, err := w.KaggleService.GetKernelStatus(kernelRef)
		if err != nil {
			log.Printf("[Worker %d] Error polling status: %v", workerID, err)
		}

		updateJobMetrics(job, map[string]interface{}{
			"stage":         "training",
			"kernel_status": status,
			"kernel_ref":    kernelRef,
		})

		if status == "completed" {
			updateJobStatus(job, "completed")
			log.Printf("[Worker %d] Job %d completed on Kaggle!", workerID, job.ID)
			return
		} else if status == "failed" {
			updateJobFailed(job, "Kaggle kernel execution failed")
			return
		}

		time.Sleep(30 * time.Second) // Poll every 30s
	}
}

func (w *WorkerPool) processJobSimulated(workerID int, job *models.Job) {
	updateJobStatus(job, "running")

	config := make(map[string]interface{})
	_ = json.Unmarshal([]byte(job.Configuration), &config)
	epochs := 5
	if e, ok := config["epochs"].(float64); ok {
		epochs = int(e)
	}

	for i := 1; i <= epochs; i++ {
		database.DB.First(job, job.ID)
		if job.Status == "cancelled" {
			log.Printf("[Worker %d] Job %d cancelled", workerID, job.ID)
			return
		}

		time.Sleep(5 * time.Second)

		metrics := map[string]interface{}{
			"mode":         "simulation",
			"epoch":        i,
			"total_epochs": epochs,
			"loss":         0.5 - (float64(i) * 0.05),
			"accuracy":     0.5 + (float64(i) * 0.08),
			"time_elapsed": fmt.Sprintf("%ds", i*5),
		}
		metricsJSON, _ := json.Marshal(metrics)
		job.Metrics = datatypes.JSON(metricsJSON)
		database.DB.Save(job)

		log.Printf("[Worker %d] Job %d - Epoch %d/%d (simulated)", workerID, job.ID, i, epochs)
	}

	updateJobStatus(job, "completed")
	log.Printf("[Worker %d] Job %d completed (simulated)", workerID, job.ID)
}

func updateJobStatus(job *models.Job, status string) {
	job.Status = status
	database.DB.Save(job)
}

func updateJobMetrics(job *models.Job, metrics map[string]interface{}) {
	metricsJSON, _ := json.Marshal(metrics)
	job.Metrics = datatypes.JSON(metricsJSON)
	database.DB.Save(job)
}

func updateJobFailed(job *models.Job, reason string) {
	job.Status = "failed"
	metricsJSON, _ := json.Marshal(map[string]interface{}{"error": reason})
	job.Metrics = datatypes.JSON(metricsJSON)
	database.DB.Save(job)
	log.Printf("[Job %d] FAILED: %s", job.ID, reason)
}

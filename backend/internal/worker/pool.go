package worker

import (
	"encoding/json"
	"finetune-studio/internal/database"
	"finetune-studio/internal/models"
	"fmt"
	"log"
	"time"

	"gorm.io/datatypes"
)

type WorkerPool struct {
	Workers  int
	JobQueue chan uint
	Quit     chan bool
}

// Global instance to push jobs
var Pool *WorkerPool

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		Workers:  workers,
		JobQueue: make(chan uint, 100), // Buffer for 100 jobs
		Quit:     make(chan bool),
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
	if err := database.DB.First(&job, jobID).Error; err != nil {
		log.Printf("[Worker %d] Job %d not found: %v", workerID, jobID, err)
		return
	}

	// 1. Mark as Starting
	updateStatus(&job, "starting")

	// Simulate Initialization
	time.Sleep(2 * time.Second)

	// 2. Mark as Running
	updateStatus(&job, "running")

	// Simulate Training Loop (Mock)
	config := make(map[string]interface{})
	_ = json.Unmarshal([]byte(job.Configuration), &config)
	epochs := 5 // default
	if e, ok := config["epochs"].(float64); ok {
		epochs = int(e)
	}

	for i := 1; i <= epochs; i++ {
		// Checks for cancellation
		database.DB.First(&job, jobID) // Re-fetch to check status
		if job.Status == "cancelled" {
			log.Printf("[Worker %d] Job %d cancelled", workerID, jobID)
			return
		}

		// Simulate Epoch Work
		time.Sleep(5 * time.Second) // 5 seconds per epoch

		// Update Metrics
		metrics := map[string]interface{}{
			"epoch":        i,
			"total_epochs": epochs,
			"loss":         0.5 - (float64(i) * 0.05), // Fake decreasing loss
			"accuracy":     0.5 + (float64(i) * 0.08), // Fake increasing accuracy
			"time_elapsed": fmt.Sprintf("%ds", i*5),
		}
		metricsJSON, _ := json.Marshal(metrics)
		job.Metrics = datatypes.JSON(metricsJSON)
		database.DB.Save(&job)

		log.Printf("[Worker %d] Job %d - Epoch %d/%d completed", workerID, jobID, i, epochs)
	}

	// 3. Mark as Completed
	updateStatus(&job, "completed")
	log.Printf("[Worker %d] Job %d completed successfully", workerID, jobID)
}

func updateStatus(job *models.Job, status string) {
	job.Status = status
	database.DB.Save(job)
}

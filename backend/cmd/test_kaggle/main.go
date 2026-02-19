package main

import (
	"finetune-studio/internal/config"
	"finetune-studio/internal/database"
	"finetune-studio/internal/services/kaggle"
	"finetune-studio/internal/storage"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	// Connect to DB
	database.Connect(cfg.DatabaseURL)

	// Connect to MinIO
	storage.Connect(cfg.MinioEndpoint, cfg.MinioUser, cfg.MinioPassword, cfg.MinioUseSSL)

	// Initialize Kaggle
	if cfg.KaggleUsername == "" || cfg.KaggleKey == "" {
		log.Fatal("❌ KAGGLE_USERNAME or KAGGLE_KEY not set. Check .env file.")
	}

	svc := kaggle.NewService("/tmp/kaggle_test")
	log.Printf("✅ Kaggle Service initialized for user: %s", cfg.KaggleUsername)

	// Test 1: Create a dummy dataset
	log.Println("--- Test 1: Creating test dataset on Kaggle ---")
	tmpFile := "/tmp/kaggle_test/test_dataset.json"
	os.MkdirAll("/tmp/kaggle_test", 0755)
	os.WriteFile(tmpFile, []byte(`[{"text":"hello world", "label":"positive"},{"text":"this is bad","label":"negative"}]`), 0644)

	ref, err := svc.CreateDataset("test-finetune-dataset", tmpFile)
	if err != nil {
		log.Fatalf("❌ CreateDataset failed: %v", err)
	}
	fmt.Printf("✅ Dataset created: %s\n", ref)

	// Test 2: Push a test kernel
	log.Println("--- Test 2: Pushing test kernel to Kaggle ---")
	notebookBytes, err := os.ReadFile("/app/templates/finetune-kernel.ipynb")
	if err != nil {
		log.Fatalf("❌ Failed to read notebook: %v", err)
	}

	kernelRef, err := svc.PushKernel("test-finetune-kernel", notebookBytes, []string{ref})
	if err != nil {
		log.Fatalf("❌ PushKernel failed: %v", err)
	}
	fmt.Printf("✅ Kernel pushed: %s\n", kernelRef)

	// Test 3: Poll status
	log.Println("--- Test 3: Polling kernel status ---")
	for i := 0; i < 60; i++ { // Poll for up to 30 minutes
		status, err := svc.GetKernelStatus(kernelRef)
		if err != nil {
			log.Printf("⚠️  Status poll error: %v", err)
		} else {
			fmt.Printf("  [%s] Status: %s\n", time.Now().Format("15:04:05"), status)
			if status == "completed" {
				fmt.Println("🎉 Kernel completed successfully!")
				return
			} else if status == "failed" {
				fmt.Println("❌ Kernel failed!")
				return
			}
		}
		time.Sleep(30 * time.Second)
	}

	fmt.Println("⏰ Timed out waiting for kernel completion")
}

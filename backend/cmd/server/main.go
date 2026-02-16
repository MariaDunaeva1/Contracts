package main

import (
	"finetune-studio/internal/config"
	"finetune-studio/internal/database"
	"finetune-studio/internal/handlers"
	"finetune-studio/internal/storage"
	"finetune-studio/internal/worker"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Cargar configuración
	cfg := config.LoadConfig()

	// 2. Conectar a Base de Datos
	database.Connect(cfg.DatabaseURL)

	// 3. Conectar a MinIO
	storage.Connect(cfg.MinioEndpoint, cfg.MinioUser, cfg.MinioPassword, cfg.MinioUseSSL)

	// 4. Inicializar Router
	r := gin.Default()

	// Middleware CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health Check
	r.GET("/api/v1/health", func(c *gin.Context) {
		dbStatus := "up"
		sqlDB, err := database.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "down"
		}

		storageStatus := "up"
		// Simple check: list buckets
		_, err = storage.Client.ListBuckets(c)
		if err != nil {
			storageStatus = "down"
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"services": gin.H{
				"db":      dbStatus,
				"storage": storageStatus,
			},
		})
	})

	// Dataset Routes
	v1 := r.Group("/api/v1")
	{
		v1.POST("/datasets", handlers.UploadDataset)
		v1.GET("/datasets", handlers.ListDatasets)
		v1.GET("/datasets/:id", handlers.GetDataset)
		v1.DELETE("/datasets/:id", handlers.DeleteDataset)
	}

	// Job Routes
	{
		v1.POST("/jobs", handlers.CreateJob)
		v1.GET("/jobs", handlers.ListJobs)
		v1.GET("/jobs/:id", handlers.GetJob)
		v1.DELETE("/jobs/:id", handlers.CancelJob)
	}

	// Initialize Worker Pool
	worker.Pool = worker.NewWorkerPool(5)
	worker.Pool.Start()

	log.Println("🚀 Server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

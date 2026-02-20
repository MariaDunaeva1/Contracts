package main

import (
	"context"
	"finetune-studio/internal/config"
	"finetune-studio/internal/database"
	"finetune-studio/internal/handlers"
	"finetune-studio/internal/logger"
	"finetune-studio/internal/metrics"
	"finetune-studio/internal/middleware"
	"finetune-studio/internal/services/kaggle"
	"finetune-studio/internal/services/logs"
	"finetune-studio/internal/storage"
	"finetune-studio/internal/worker"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Version = "1.0.0"

func main() {
	// 1. Initialize logger
	if err := logger.Initialize(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting Finetune Studio",
		zap.String("version", Version),
		zap.String("env", getEnv("APP_ENV", "development")),
	)

	// 2. Load configuration
	cfg := config.LoadConfig()

	// 3. Connect to Database
	database.Connect(cfg.DatabaseURL)
	
	// Configure connection pool
	sqlDB, err := database.DB.DB()
	if err != nil {
		logger.Fatal("Failed to get database connection", zap.Error(err))
	}
	
	maxConns := getEnvInt("DB_MAX_CONNECTIONS", 25)
	maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNECTIONS", 5)
	connLifetime := getEnvDuration("DB_CONNECTION_MAX_LIFETIME", 5*time.Minute)
	
	sqlDB.SetMaxOpenConns(maxConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connLifetime)
	
	logger.Info("Database connection pool configured",
		zap.Int("max_connections", maxConns),
		zap.Int("max_idle", maxIdleConns),
		zap.Duration("max_lifetime", connLifetime),
	)

	// 4. Connect to MinIO
	storage.Connect(cfg.MinioEndpoint, cfg.MinioUser, cfg.MinioPassword, cfg.MinioUseSSL)

	// 5. Set Gin mode
	if getEnv("APP_ENV", "development") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 6. Initialize Router (without default middleware)
	r := gin.New()
	r.Use(gin.Recovery())

	// 7. Apply middleware
	// Compression
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	
	// Request size limit
	maxSizeMB := getEnvInt("MAX_REQUEST_SIZE_MB", 10)
	r.Use(middleware.RequestSizeLimit(maxSizeMB))
	
	// CORS
	allowedOrigins := strings.Split(getEnv("ALLOWED_ORIGINS", "*"), ",")
	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))
	
	// Logging
	r.Use(middleware.RequestLogger(logger.Log))
	
	// Metrics (if enabled)
	if getEnv("METRICS_ENABLED", "true") == "true" {
		r.Use(metrics.PrometheusMiddleware())
	}
	
	// Rate limiting
	rateLimitRPM := getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100)
	r.Use(middleware.RateLimitMiddleware(rateLimitRPM))

	// 8. Health Check (enhanced)
	startTime := time.Now()
	r.GET("/api/v1/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		overallStatus := "healthy"
		statusCode := http.StatusOK

		// Check database
		dbStatus := "up"
		dbResponseTime := time.Now()
		sqlDB, err := database.DB.DB()
		if err != nil || sqlDB.PingContext(ctx) != nil {
			dbStatus = "down"
			overallStatus = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}
		dbLatency := time.Since(dbResponseTime)

		// Check storage
		storageStatus := "up"
		storageResponseTime := time.Now()
		_, err = storage.Client.ListBuckets(ctx)
		if err != nil {
			storageStatus = "down"
			overallStatus = "degraded"
			if statusCode == http.StatusOK {
				statusCode = http.StatusOK // Storage is not critical
			}
		}
		storageLatency := time.Since(storageResponseTime)

		// Worker pool status
		workerStatus := "up"
		if worker.Pool == nil {
			workerStatus = "down"
		}

		// Update DB metrics
		if sqlDB != nil {
			stats := sqlDB.Stats()
			metrics.SetDBConnectionsActive(stats.InUse)
			metrics.SetDBConnectionsIdle(stats.Idle)
		}

		c.JSON(statusCode, gin.H{
			"status":  overallStatus,
			"version": Version,
			"uptime":  time.Since(startTime).String(),
			"services": gin.H{
				"database": gin.H{
					"status":        dbStatus,
					"response_time": dbLatency.String(),
				},
				"storage": gin.H{
					"status":        storageStatus,
					"response_time": storageLatency.String(),
				},
				"workers": gin.H{
					"status": workerStatus,
				},
			},
		})
	})

	// Metrics endpoint
	if getEnv("METRICS_ENABLED", "true") == "true" {
		r.GET("/api/v1/metrics", metrics.Handler())
	}

	// 9. Initialize services
	kaggleSvc := kaggle.NewService("/tmp/kaggle_workdir")
	
	workerPoolSize := getEnvInt("WORKER_POOL_SIZE", 5)
	worker.Pool = worker.NewWorkerPool(workerPoolSize, kaggleSvc)
	worker.Pool.Start()
	
	logService := logs.NewLogService(storage.Client)
	logHandler := handlers.NewLogHandler(logService)
	
	modelStorage := storage.NewModelStorage(storage.Client)
	modelHandler := handlers.NewModelHandler(modelStorage)
	
	evaluationHandler := handlers.NewEvaluationHandler()

	logger.Info("Services initialized",
		zap.Int("worker_pool_size", workerPoolSize),
	)

	// 10. Setup routes
	v1 := r.Group("/api/v1")
	
	// Rate limit for expensive endpoints
	expensiveRateLimit := getEnvInt("RATE_LIMIT_EXPENSIVE_ENDPOINTS", 10)
	expensiveLimiter := middleware.ExpensiveEndpointRateLimit(expensiveRateLimit)

	// Dataset Routes
	{
		v1.POST("/datasets", expensiveLimiter, handlers.UploadDataset)
		v1.GET("/datasets", handlers.ListDatasets)
		v1.GET("/datasets/:id", handlers.GetDataset)
		v1.DELETE("/datasets/:id", handlers.DeleteDataset)
	}

	// Job Routes
	{
		v1.POST("/jobs", expensiveLimiter, handlers.CreateJob)
		v1.GET("/jobs", handlers.ListJobs)
		v1.GET("/jobs/:id", handlers.GetJob)
		v1.DELETE("/jobs/:id", handlers.CancelJob)
	}

	// Log Routes
	{
		v1.GET("/jobs/:id/logs", logHandler.StreamLogs)
		v1.POST("/jobs/:id/logs", logHandler.CreateLogEntry)
	}

	// Model Routes
	{
		v1.GET("/models", modelHandler.ListModels)
		v1.GET("/models/:id", modelHandler.GetModel)
		v1.GET("/models/:id/download", expensiveLimiter, modelHandler.DownloadModel)
		v1.POST("/models", modelHandler.CreateModel)
		v1.PUT("/models/:id", modelHandler.UpdateModel)
		v1.DELETE("/models/:id", modelHandler.DeleteModel)
	}

	// Evaluation Routes
	{
		v1.POST("/models/:id/evaluate", expensiveLimiter, evaluationHandler.CreateEvaluation)
		v1.GET("/evaluations", evaluationHandler.ListEvaluations)
		v1.GET("/evaluations/:id", evaluationHandler.GetEvaluation)
		v1.PUT("/evaluations/:id", evaluationHandler.UpdateEvaluation)
	}

	// Contract Analysis Routes (RAG)
	contractHandler := handlers.NewContractHandler("http://localhost:8001")
	{
		v1.POST("/contracts/analyze", expensiveLimiter, contractHandler.AnalyzeContract)
		v1.POST("/clauses/search", contractHandler.SearchClauses)
		v1.GET("/contracts/:id/similar", contractHandler.FindSimilarContracts)
		v1.GET("/contracts/:id/clauses", contractHandler.GetContractClauses)
		v1.DELETE("/contracts/:id/index", contractHandler.DeleteContractIndex)
		v1.GET("/rag/health", contractHandler.RAGHealth)
		v1.GET("/rag/stats", contractHandler.RAGStats)
	}

	// 11. Setup HTTP server with graceful shutdown
	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting",
			zap.String("port", port),
			zap.Strings("allowed_origins", allowedOrigins),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// 12. Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Stop accepting new jobs
	if worker.Pool != nil {
		worker.Pool.Stop()
		logger.Info("Worker pool stopped")
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

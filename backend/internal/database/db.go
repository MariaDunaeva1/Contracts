package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"finetune-studio/internal/models"
)

var DB *gorm.DB

func Connect(databaseURL string) {
	var err error
	
	// Reintentar conexión (útil cuando docker-compose levanta db y api simultáneamente)
	for i := 0; i < 5; i++ {
		DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database. Retrying in 2s... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL successfully")

	// AutoMigrate models
	err = DB.AutoMigrate(&models.Dataset{}, &models.Job{})
	if err != nil {
		log.Printf("❌ AutoMigrate failed: %v", err)
	} else {
		log.Println("✅ Database schema migrated successfully")
	}
}

package storage

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client

func Connect(endpoint, accessKey, secretKey string, useSSL bool) {
	var err error
	Client, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	log.Println("✅ MinIO client initialized")
	initBuckets()
}

func initBuckets() {
	buckets := []string{"datasets", "models", "logs"}
	ctx := context.Background()

	for _, bucket := range buckets {
		exists, err := Client.BucketExists(ctx, bucket)
		if err != nil {
			log.Printf("⚠️ Check bucket %s failed: %v", bucket, err)
			continue
		}
		if !exists {
			err = Client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				log.Printf("❌ Failed to create bucket %s: %v", bucket, err)
			} else {
				log.Printf("📂 Created bucket: %s", bucket)
			}
		} else {
			log.Printf("📂 Bucket exists: %s", bucket)
		}
	}
}

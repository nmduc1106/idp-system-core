package main

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🚀 Dang khoi dong kiem tra ket noi he thong...")

	// 1. Kiem tra PostgreSQL
	dsn := "host=localhost user=idp_user password=secret_password dbname=idp_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Postgres LOI: %v", err)
	}
	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Postgres Ping LOI: %v", err)
	}
	fmt.Println("✅ PostgreSQL: Ket noi thanh cong!")

	// 2. Kiem tra RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("❌ RabbitMQ LOI: %v", err)
	}
	defer conn.Close()
	fmt.Println("✅ RabbitMQ: Ket noi thanh cong!")

	// 3. Kiem tra MinIO
	minioClient, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minio_admin", "minio_secret_key", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("❌ MinIO Client Setup LOI: %v", err)
	}
	// Kiem tra bucket 'documents' co ton tai khong
	exists, err := minioClient.BucketExists(context.Background(), "documents")
	if err != nil {
		log.Fatalf("❌ MinIO Check Bucket LOI: %v", err)
	}
	if exists {
		fmt.Println("✅ MinIO: Ket noi thanh cong & Tim thay bucket 'documents'!")
	} else {
		fmt.Printf("⚠️ MinIO: Ket noi duoc nhung KHONG THAY bucket 'documents'. Ban da tao no chua?\n")
	}

	fmt.Println("------------------------------------------------")
	fmt.Println("🎉 CHUC MUNG! HA TANG DA SAN SANG DE CODE.")
}
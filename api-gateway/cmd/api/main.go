package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "idp-api-gateway/docs"

	"idp-api-gateway/internal/adapters/handlers"
	"idp-api-gateway/internal/adapters/middlewares"
	"idp-api-gateway/internal/adapters/queue"
	
	// Import package repositories (chứa cả UserRepo và DocRepo)
	"idp-api-gateway/internal/adapters/repositories" 
	
	"idp-api-gateway/internal/adapters/storage"
	"idp-api-gateway/internal/core/pkg/tracing"
	"idp-api-gateway/internal/core/services"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// @title           IDP API Gateway
// @version         1.0
// @description     Hệ thống xử lý tài liệu thông minh (Intelligent Document Processing).
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// --- 0. Setup Tracing ---
	tp, err := tracing.InitTracer()
	if err != nil {
		log.Fatal("❌ Failed to initialize tracer:", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("❌ Error shutting down tracer provider: %v", err)
		}
	}()

	// --- CẤU HÌNH ---
	// Key này dùng để ký và giải mã Token
	jwtSecret := "my_super_secret_key_change_me"

	// 1. Setup Database Connection
	dsn := "host=localhost user=idp_user password=secret_password dbname=idp_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to DB:", err)
	}

	// 2. Setup MinIO Connection
	minioClient, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minio_admin", "minio_secret_key", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal("❌ Failed to connect to MinIO:", err)
	}

	// 3. Setup RabbitMQ (Self-Healing Producer)
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}
	queueProducer := queue.NewRabbitMQProducer(amqpURL)
	defer queueProducer.Close()

	// --- 4. KHỞI TẠO MODULE AUTHENTICATION (Mới) ---
	// Lấy *sql.DB từ Gorm để truyền vào UserRepo
	sqlDB, _ := db.DB()
	
	// [ĐÃ SỬA] Dùng 'repositories' thay vì 'repoPostgres'
	userRepo := repositories.NewUserRepository(sqlDB) 
	
	authService := services.NewAuthService(userRepo, jwtSecret) // Service xử lý đăng nhập/đăng ký
	authHandler := handlers.NewAuthHandler(authService)         // API Handler cho Auth

	// --- 5. KHỞI TẠO MODULE DOCUMENT (Cũ) ---
	docRepo := repositories.NewPostgresRepository(db)
	fileStorage := storage.NewMinIOStorage(minioClient)


	idpService := services.NewIDPService(docRepo, fileStorage, queueProducer)
	httpHandler := handlers.NewHTTPHandler(idpService)

	// --- 6. SETUP ROUTER ---
	r := gin.Default()

	// Add OpenTelemetry Gin Middleware
	r.Use(otelgin.Middleware("api-gateway"))

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API Routes Group
	v1 := r.Group("/api/v1")
	{
		// A. Public Routes (Ai cũng gọi được)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// B. Protected Routes (Phải có Token mới gọi được)
		// Gắn "Người bảo vệ" (Middleware) vào đây
		protected := v1.Group("/", middlewares.JWTMiddleware(jwtSecret))
		{
			protected.POST("/upload", httpHandler.Upload) // Upload an toàn
			protected.GET("/jobs/:id", httpHandler.GetJob)

			users := protected.Group("/users")
			{
				users.GET("/me", authHandler.GetMe)
			}
		}
	}

	fmt.Println("---------------------------------------------------------")
	fmt.Println("🚀 API Gateway running on :8080 (With JWT Auth)")
	fmt.Println("📄 Swagger Docs: http://localhost:8080/swagger/index.html")
	fmt.Println("---------------------------------------------------------")
	
	if err := r.Run(":8080"); err != nil {
		log.Fatal("❌ Server shutdown:", err)
	}
}
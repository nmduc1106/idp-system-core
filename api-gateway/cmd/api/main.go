package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "idp-api-gateway/docs"

	"idp-api-gateway/internal/adapters/cache"
	"idp-api-gateway/internal/adapters/handlers"
	"idp-api-gateway/internal/adapters/middlewares"
	"idp-api-gateway/internal/adapters/queue"
	"idp-api-gateway/internal/adapters/repositories"
	"idp-api-gateway/internal/adapters/storage"
	"idp-api-gateway/internal/core/domain"
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
	if err := godotenv.Load("../.env"); err != nil {
        log.Println("⚠️ Không tìm thấy file .env (Sẽ sử dụng biến môi trường của OS/Docker)")
    }
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
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "my_super_secret_key_change_me" // Fallback cho dev
	}

	// 1. Setup Database Connection (Lấy từ Env do Docker cung cấp)
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Dùng cho chạy localhost ngoài Docker
		dsn = "host=localhost user=idp_user password=secret_password dbname=idp_db port=5432 sslmode=disable"
	}
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to DB:", err)
	}

	// Auto-migrate: Đăng ký đầy đủ 3 model chính
	if err := db.AutoMigrate(&domain.User{}, &domain.Document{}, &domain.Job{}); err != nil {
		log.Fatal("❌ Failed to auto-migrate database schema:", err)
	}
	log.Println("✅ Database schema migrated successfully")

	// 2. Setup MinIO Connection
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}
	
	// Thêm Fallback cho User & Pass khi chạy local
	minioUser := os.Getenv("MINIO_ACCESS_KEY")
	if minioUser == "" {
		minioUser = "minio_admin"
	}
	
	minioPass := os.Getenv("MINIO_SECRET_KEY")
	if minioPass == "" {
		minioPass = "minio_secret_key"
	}

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioUser, minioPass, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal("❌ Failed to connect to MinIO:", err)
	}

	// 3. Setup RabbitMQ
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}
	queueProducer := queue.NewRabbitMQProducer(amqpURL)
	defer queueProducer.Close()

	// Setup Redis Cache
	redisClient := cache.NewRedisClient()
	defer redisClient.Close()

	// --- 4. KHỞI TẠO MODULE AUTHENTICATION ---
	sqlDB, _ := db.DB()
	userRepo := repositories.NewUserRepository(sqlDB)
	authService := services.NewAuthService(userRepo, jwtSecret, redisClient)
	authHandler := handlers.NewAuthHandler(authService)

	// --- 5. KHỞI TẠO MODULE DOCUMENT ---
	docRepo := repositories.NewPostgresRepository(db)
	fileStorage := storage.NewMinIOStorage(minioClient)
	idpService := services.NewIDPService(docRepo, fileStorage, queueProducer, redisClient, redisClient)
	httpHandler := handlers.NewHTTPHandler(idpService)

	// --- 6. KHỞI TẠO MODULE ADMIN ---
	adminRepo := repositories.NewAdminRepository(db)
	adminService := services.NewAdminService(adminRepo, redisClient)
	adminHandler := handlers.NewAdminHandler(adminService)

	// --- 7. KHỞI TẠO WEBHOOK HANDLER ---
	webhookHandler := handlers.NewWebhookHandler(idpService)

	// --- 8. SETUP ROUTER ---
	r := gin.Default()

	// CORS Middleware - must be before any route definitions
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With", "Accept-Version", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(otelgin.Middleware("api-gateway"))

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Internal routes (no auth - called by Python worker within Docker network)
	internal := r.Group("/internal")
	{
		internal.POST("/webhook/job-completed", webhookHandler.JobCompleted)
	}

	// API Routes
	v1 := r.Group("/api/v1")
	{
		// Public
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// Protected
		protected := v1.Group("/", middlewares.JWTMiddleware(jwtSecret))
		{
			protected.POST("/upload", middlewares.RateLimitMiddleware(redisClient, 10, time.Minute), httpHandler.Upload)
			protected.GET("/jobs", httpHandler.GetUserJobs)
			protected.GET("/jobs/export", httpHandler.ExportJobsExcel)
			protected.GET("/jobs/:id", httpHandler.GetJob)
			protected.GET("/jobs/:id/stream", httpHandler.StreamJob)
			protected.POST("/auth/logout", authHandler.Logout)

			users := protected.Group("/users")
			{
				users.GET("/me", authHandler.GetMe)
			}

			// Admin-only routes
			adminGroup := protected.Group("/admin", middlewares.RequireRole("ADMIN"))
			{
				adminGroup.GET("/stats", adminHandler.GetStats)
				adminGroup.GET("/jobs", adminHandler.GetJobs)
				adminGroup.GET("/users", adminHandler.GetUsers)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("---------------------------------------------------------")
	fmt.Println("🚀 API Gateway running on :" + port + " (With JWT Auth)")
	fmt.Println("📄 Swagger Docs: http://localhost:" + port + "/swagger/index.html")
	fmt.Println("---------------------------------------------------------")

	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Server shutdown:", err)
	}
}
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/whatsapp-promo-poc/api/internal/config"
	"github.com/whatsapp-promo-poc/api/internal/database"
	"github.com/whatsapp-promo-poc/api/internal/handler"
	"github.com/whatsapp-promo-poc/api/internal/middleware"
	"github.com/whatsapp-promo-poc/api/internal/service"
	"github.com/whatsapp-promo-poc/pkg/i18n"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewConnection(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed test data
	if err := database.SeedTestData(db); err != nil {
		log.Printf("Warning: Failed to seed test data (may already exist): %v", err)
	}

	// Initialize i18n
	i18nManager, err := i18n.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}

	// Initialize services
	promoService := service.NewPromoService(
		db,
		i18nManager,
		cfg.MockTelecomDelayMs,
		cfg.MockAntiFraudEnabled,
		cfg.RateLimitPerMinute,
		cfg.RateLimitPerHour,
	)

	// Initialize handlers
	redeemHandler := handler.NewRedeemHandler(promoService)
	healthHandler := handler.NewHealthHandler(db)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	// Routes
	v1 := r.Group("/api/v1")
	{
		v1.POST("/redeem", redeemHandler.Handle)
		v1.GET("/health", healthHandler.Handle)
	}

	// Start server
	addr := cfg.Host + ":" + cfg.Port
	log.Printf("Starting API server on %s", addr)
	log.Printf("Database path: %s", cfg.DBPath)
	log.Printf("Rate limits: %d/min, %d/hour", cfg.RateLimitPerMinute, cfg.RateLimitPerHour)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

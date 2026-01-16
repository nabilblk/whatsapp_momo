package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"

	"github.com/whatsapp-promo-poc/bridge/internal/client"
	"github.com/whatsapp-promo-poc/bridge/internal/config"
	"github.com/whatsapp-promo-poc/bridge/internal/handler"
	"github.com/whatsapp-promo-poc/bridge/internal/whatsapp"
	"github.com/whatsapp-promo-poc/pkg/i18n"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	log.Printf("Starting WhatsApp Bridge")
	log.Printf("API URL: %s", cfg.APIUrl)
	log.Printf("Session path: %s", cfg.SessionPath)
	log.Printf("Default language: %s", cfg.DefaultLanguage)

	// Initialize i18n
	i18nManager, err := i18n.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}

	// Initialize API client
	apiClient := client.NewAPIClient(cfg.APIUrl)

	// Wait for API to be ready
	waitForAPI(apiClient, cfg.APIUrl)

	// Initialize WhatsApp client
	ctx := context.Background()
	waClient, err := whatsapp.NewClient(ctx, &whatsapp.Config{
		SessionDBPath: cfg.SessionPath,
		LogLevel:      cfg.LogLevel,
	})
	if err != nil {
		log.Fatalf("Failed to create WhatsApp client: %v", err)
	}

	// Initialize message handler
	msgHandler := handler.NewMessageHandler(
		apiClient,
		i18nManager,
		cfg.DefaultLanguage,
		waClient.SendText,
	)

	// Set message handler
	waClient.SetMessageHandler(msgHandler.Handle)

	// Connect to WhatsApp
	if err := waClient.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to WhatsApp: %v", err)
	}

	log.Println("========================================")
	log.Println("WhatsApp Bridge is running!")
	log.Println("Send a promo code to test the system")
	log.Println("========================================")

	// Wait for shutdown signal
	waClient.WaitForShutdown()
}

func waitForAPI(apiClient *client.APIClient, apiURL string) {
	ctx := context.Background()
	maxRetries := 60 // Wait up to 2 minutes

	log.Printf("Waiting for API to be ready at %s...", apiURL)

	for i := 0; i < maxRetries; i++ {
		if err := apiClient.HealthCheck(ctx); err == nil {
			log.Println("API is ready!")
			return
		}
		if i > 0 && i%10 == 0 {
			log.Printf("Still waiting for API... (%d/%d)", i, maxRetries)
		}
		time.Sleep(2 * time.Second)
	}

	log.Fatal("API did not become ready in time. Please ensure the API service is running.")
}

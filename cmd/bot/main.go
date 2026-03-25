package main

import (
	"log"

	"purrtopia/internal/ai"
	"purrtopia/internal/auth"
	"purrtopia/internal/bot"
	"purrtopia/internal/config"
	"purrtopia/internal/database"
)

func main() {
	// Load configuration (自動驗證 required 欄位)
	cfg := config.Load()

	// 初始化權限模組
	auth.Init(cfg)

	// 初始化 Gemini AI（可選）
	if cfg.Gemini.APIKey != "" {
		if err := ai.Init(&cfg.Gemini); err != nil {
			log.Printf("Warning: Failed to initialize Gemini AI: %v", err)
		} else {
			log.Println("Gemini AI initialized successfully")
		}
	} else {
		log.Println("Warning: GEMINI_API_KEY not set, AI features disabled")
	}

	// db連線
	if err := database.Connect(&cfg.DB); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create bot instance
	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Start the bot
	err = b.Start()
	if err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}

	// Wait for interrupt signal
	b.Wait()

	// Graceful shutdown
	err = b.Stop()
	if err != nil {
		log.Fatalf("Failed to stop bot: %v", err)
	}

	log.Println("Bot has been shut down gracefully")
}

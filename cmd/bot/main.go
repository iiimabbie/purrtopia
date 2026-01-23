package main

import (
	"log"

	"discord-bot-template/internal/bot"
	"discord-bot-template/internal/config"
	"discord-bot-template/internal/database"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate token
	if cfg.Token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
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

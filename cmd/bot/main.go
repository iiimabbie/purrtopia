package main

import (
	"log"

	"discord-bot-template/internal/bot"
	"discord-bot-template/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate token
	if cfg.Token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

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

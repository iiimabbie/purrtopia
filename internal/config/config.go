package config

import (
	"os"
)

// Config holds all configuration for the bot
type Config struct {
	Token   string
	GuildID string // Optional: for testing commands in specific guild
}

// Load returns configuration from environment variables
func Load() *Config {
	return &Config{
		Token:   getEnv("DISCORD_TOKEN", ""),
		GuildID: getEnv("GUILD_ID", ""), // Leave empty to register global commands
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

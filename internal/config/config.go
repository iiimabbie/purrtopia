package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the bot
type Config struct {
	Token   string
	GuildID string // Optional: for testing commands in specific guild
	DB      DBConfig
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// DSN returns the MySQL connection string
func (c *DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

// Load returns configuration from environment variables
func Load() *Config {
	return &Config{
		Token:   getEnv("DISCORD_TOKEN", ""),
		GuildID: getEnv("GUILD_ID", ""), // Leave empty to register global commands
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "heartopia"),
			Password: getEnv("DB_PASSWORD", "heartopia_secret"),
			Name:     getEnv("DB_NAME", "heartopia"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

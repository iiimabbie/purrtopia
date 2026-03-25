package config

import (
	"context"
	"fmt"
	"log"

	"github.com/sethvargo/go-envconfig"
)

// Config holds all configuration for the bot
type Config struct {
	Token    string   `env:"DISCORD_TOKEN,required"`
	GuildIDs []string `env:"GUILD_IDS"` // Comma-separated guild IDs, empty = global commands
	OwnerIDs []string `env:"BOT_OWNER_IDS"`
	AdminIDs []string `env:"BOT_ADMIN_IDS"`
	DB     DBConfig
	Gemini GeminiConfig
}

// GeminiConfig holds Gemini AI configuration
type GeminiConfig struct {
	APIKey           string `env:"GEMINI_API_KEY"`
	Model            string `env:"GEMINI_MODEL,default=gemini-2.0-flash"`
	FallbackEndpoint string `env:"GEMINI_FALLBACK_ENDPOINT"` // OpenAI-compatible endpoint (e.g., https://mingyuuu.zeabur.app/v1)
	FallbackAPIKey   string `env:"GEMINI_FALLBACK_API_KEY"`
	FallbackModel    string `env:"GEMINI_FALLBACK_MODEL,default=gemini-2.5-flash"`
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string `env:"DB_HOST,default=localhost"`
	Port     string `env:"DB_PORT,default=5432"`
	User     string `env:"DB_USER,default=heartopia"`
	Password string `env:"DB_PASSWORD,default=heartopia"`
	Name     string `env:"DB_NAME,default=heartopia"`
}

// DSN returns the PostgreSQL connection string
func (c *DBConfig) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Taipei",
		c.Host, c.User, c.Password, c.Name, c.Port)
}

// Load returns configuration from environment variables
func Load() *Config {
	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return &cfg
}

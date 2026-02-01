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
	DB       DBConfig
	Gemini   GeminiConfig
	Keywords KeywordConfig
}

// GeminiConfig holds Gemini AI configuration
type GeminiConfig struct {
	APIKey           string `env:"GEMINI_API_KEY"`
	Model            string `env:"GEMINI_MODEL,default=gemini-2.0-flash"`
	FallbackEndpoint string `env:"GEMINI_FALLBACK_ENDPOINT"` // OpenAI-compatible endpoint (e.g., https://mingyuuu.zeabur.app/v1)
	FallbackAPIKey   string `env:"GEMINI_FALLBACK_API_KEY"`
	FallbackModel    string `env:"GEMINI_FALLBACK_MODEL,default=gemini-2.5-flash"`
	EnableMention    bool   `env:"AI_ENABLE_MENTION,default=false"`  // Enable @bot mention chat feature
	EnableClassify   bool   `env:"AI_ENABLE_CLASSIFY,default=false"` // Enable keyword classification feature
}

// KeywordConfig holds keyword trigger configuration
type KeywordConfig struct {
	Triggers        []string `env:"KEYWORD_TRIGGERS,default=粉紅,泡泡,粉泡,粉紅泡泡"`
	BubbleLink      string   `env:"CHANNEL_BUBBLE"`
	DailyTownLink   string   `env:"CHANNEL_DAILY_TOWN"`
	WeatherLink     string   `env:"CHANNEL_WEATHER"`
	HelperLink      string   `env:"CHANNEL_HELPER"`
	EnabledChannels []string `env:"KEYWORD_ENABLED_CHANNELS"`
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string `env:"DB_HOST,default=localhost"`
	Port     string `env:"DB_PORT,default=3306"`
	User     string `env:"DB_USER,default=purrtopia"`
	Password string `env:"DB_PASSWORD,default=changeme"`
	Name     string `env:"DB_NAME,default=purrtopia"`
}

// DSN returns the MySQL connection string
func (c *DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

// Load returns configuration from environment variables
func Load() *Config {
	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return &cfg
}

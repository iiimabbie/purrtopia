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
	GuildID  string   `env:"GUILD_ID"`
	OwnerIDs []string `env:"BOT_OWNER_IDS"`
	AdminIDs []string `env:"BOT_ADMIN_IDS"`
	DB       DBConfig
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

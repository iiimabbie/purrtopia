package database

import (
	"log"
)

// InitSchema creates all necessary tables
func InitSchema() error {
	schemaAvatars := `
	CREATE TABLE IF NOT EXISTS user_avatars (
		id INT AUTO_INCREMENT PRIMARY KEY,
		discord_id VARCHAR(32) NOT NULL,
		discord_username VARCHAR(128) NOT NULL,
		game_uid VARCHAR(64) NOT NULL,
		image_url TEXT NOT NULL,
		emoji_id VARCHAR(32) DEFAULT NULL,
		emoji_name VARCHAR(64) DEFAULT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_discord_id (discord_id),
		INDEX idx_game_uid (game_uid),
		INDEX idx_emoji_id (emoji_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	_, err := DB.Exec(schemaAvatars)
	if err != nil {
		return err
	}

	// Table for tracking drawn heads
	schemaDrawnHeads := `
	CREATE TABLE IF NOT EXISTS drawn_heads (
		id INT AUTO_INCREMENT PRIMARY KEY,
		drawer_discord_id VARCHAR(32) NOT NULL,
		avatar_id INT NOT NULL,
		drawn_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_drawer_discord_id (drawer_discord_id),
		INDEX idx_avatar_id (avatar_id),
		FOREIGN KEY (avatar_id) REFERENCES user_avatars(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	_, err = DB.Exec(schemaDrawnHeads)
	if err != nil {
		return err
	}

	// Migration: add emoji columns if they don't exist
	// Check if emoji_id column exists
	var colExists int
	err = DB.QueryRow(`
		SELECT COUNT(*) FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = 'user_avatars'
		AND COLUMN_NAME = 'emoji_id'
	`).Scan(&colExists)

	if err == nil && colExists == 0 {
		DB.Exec("ALTER TABLE user_avatars ADD COLUMN emoji_id VARCHAR(32) DEFAULT NULL")
		DB.Exec("ALTER TABLE user_avatars ADD COLUMN emoji_name VARCHAR(64) DEFAULT NULL")
		DB.Exec("CREATE INDEX idx_emoji_id ON user_avatars(emoji_id)")
		log.Println("Added emoji columns to user_avatars table")
	}

	log.Println("Database schema initialized")
	return nil
}

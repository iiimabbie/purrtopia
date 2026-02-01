package database

import (
	"log"
)

// InitSchema creates all necessary tables
func InitSchema() error {
	schemaAvatars := `
	CREATE TABLE IF NOT EXISTS user_avatars (
		id INT AUTO_INCREMENT PRIMARY KEY,
		type VARCHAR(10) DEFAULT 'USER',            -- 類型
		discord_id VARCHAR(32) DEFAULT NULL,        -- 允許 NULL
		discord_username VARCHAR(128) DEFAULT NULL, -- 允許 NULL
		game_uid VARCHAR(64) NOT NULL,              -- NPC 的名字或 ID
		image_url TEXT NOT NULL,
		emoji_id VARCHAR(32) DEFAULT NULL,
		emoji_name VARCHAR(64) DEFAULT NULL,
		weight INT DEFAULT 100,                     -- 機率權重
		rarity VARCHAR(20) DEFAULT 'N',             -- 稀有度顯示
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_discord_id (discord_id),
		INDEX idx_type (type)
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

	// Snow season proxy buy table (with server region support)
	schemaProxyBuy := `
	CREATE TABLE IF NOT EXISTS snow_proxy_buy (
		id INT AUTO_INCREMENT PRIMARY KEY,
		discord_id VARCHAR(32) NOT NULL,
		server_region VARCHAR(10) NOT NULL DEFAULT 'asia',
		items TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE KEY unique_user_server (discord_id, server_region),
		INDEX idx_discord_id (discord_id),
		INDEX idx_server_region (server_region),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err = DB.Exec(schemaProxyBuy)
	if err != nil {
		return err
	}

	// Snow season proxy sell table (with server region support)
	schemaProxySell := `
	CREATE TABLE IF NOT EXISTS snow_proxy_sell (
		id INT AUTO_INCREMENT PRIMARY KEY,
		discord_id VARCHAR(32) NOT NULL,
		server_region VARCHAR(10) NOT NULL DEFAULT 'asia',
		items TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE KEY unique_user_server (discord_id, server_region),
		INDEX idx_discord_id (discord_id),
		INDEX idx_server_region (server_region),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err = DB.Exec(schemaProxySell)
	if err != nil {
		return err
	}

	// Migration: add server_region column if tables already exist without it
	migrateServerRegion()

	log.Println("Database schema initialized")
	return nil
}

// migrateServerRegion adds server_region column to existing tables
func migrateServerRegion() {
	tables := []string{"snow_proxy_buy", "snow_proxy_sell"}

	for _, table := range tables {
		var colExists int
		err := DB.QueryRow(`
			SELECT COUNT(*) FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND COLUMN_NAME = 'server_region'
		`, table).Scan(&colExists)

		if err == nil && colExists == 0 {
			// Add server_region column
			_, err := DB.Exec("ALTER TABLE " + table + " ADD COLUMN server_region VARCHAR(10) NOT NULL DEFAULT 'asia' AFTER discord_id")
			if err != nil {
				log.Printf("Failed to add server_region to %s: %v", table, err)
				continue
			}

			// Drop old unique constraint and add new one
			DB.Exec("ALTER TABLE " + table + " DROP INDEX discord_id")
			DB.Exec("ALTER TABLE " + table + " ADD UNIQUE KEY unique_user_server (discord_id, server_region)")
			DB.Exec("ALTER TABLE " + table + " ADD INDEX idx_server_region (server_region)")

			log.Printf("Added server_region column to %s table", table)
		}
	}
}

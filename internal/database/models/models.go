package models

import (
	"time"
)

// UserAvatar 用戶頭像 (類似 JPA Entity)
type UserAvatar struct {
	ID              int        `gorm:"primaryKey;autoIncrement"`
	Type            string     `gorm:"size:10;default:'USER'"`
	DiscordID       *string    `gorm:"size:32;index"`
	DiscordUsername *string    `gorm:"size:128"`
	GameUID         string     `gorm:"column:game_uid;size:64;not null"`
	ImageURL        string     `gorm:"column:image_url;type:text;not null"`
	EmojiID         *string    `gorm:"size:32"`
	EmojiName       *string    `gorm:"size:64"`
	Weight          int        `gorm:"default:100"`
	Rarity          string     `gorm:"size:20;default:'N'"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (UserAvatar) TableName() string {
	return "user_avatars"
}

// DrawnHead 抽頭記錄
type DrawnHead struct {
	ID              int       `gorm:"primaryKey;autoIncrement"`
	DrawerDiscordID string    `gorm:"size:32;not null;index"`
	AvatarID        int       `gorm:"not null;index"`
	DrawnAt         time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (DrawnHead) TableName() string {
	return "drawn_heads"
}

// SnowProxyBuy 冰雪市場代購
type SnowProxyBuy struct {
	ID           int       `gorm:"primaryKey;autoIncrement"`
	DiscordID    string    `gorm:"size:32;not null;index;uniqueIndex:unique_user_server"`
	ServerRegion string    `gorm:"size:10;not null;default:'asia';index;uniqueIndex:unique_user_server"`
	Items        string    `gorm:"type:text;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index"`
}

// TableName 指定表名
func (SnowProxyBuy) TableName() string {
	return "snow_proxy_buy"
}

// SnowProxySell 冰雪市場代售
type SnowProxySell struct {
	ID           int       `gorm:"primaryKey;autoIncrement"`
	DiscordID    string    `gorm:"size:32;not null;index;uniqueIndex:unique_user_server"`
	ServerRegion string    `gorm:"size:10;not null;default:'asia';index;uniqueIndex:unique_user_server"`
	Items        string    `gorm:"type:text;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index"`
}

// TableName 指定表名
func (SnowProxySell) TableName() string {
	return "snow_proxy_sell"
}

package models

import (
	"time"
)

// SeasonProxySell 潮流季代售
type SeasonProxySell struct {
	ID           int       `gorm:"primaryKey;autoIncrement"`
	DiscordID    string    `gorm:"size:32;not null;index;uniqueIndex:unique_user_server"`
	ServerRegion string    `gorm:"size:10;not null;default:'asia';index;uniqueIndex:unique_user_server"`
	Items        string    `gorm:"type:text;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;index"`
}

// TableName 指定表名
func (SeasonProxySell) TableName() string {
	return "season_proxy_sell"
}

// SeasonItem 潮流季代售物品目錄
type SeasonItem struct {
	ID       int    `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"size:100;not null"`
	Category int    `gorm:"not null;default:1"` // 1=第一組選單, 2=第二組選單
}

// TableName 指定表名
func (SeasonItem) TableName() string {
	return "season_items"
}

// GroupActivity 揪團活動
type GroupActivity struct {
	ID           int       `gorm:"primaryKey;autoIncrement"`
	OrganizerID  string    `gorm:"size:32;not null"`
	ActivityType string    `gorm:"size:20;not null"`            // "rainbow", "fishing"
	ActivityTime string    `gorm:"size:200;not null"`           // 使用者填入的揪團時間
	Requirements string    `gorm:"type:text;not null"`          // 詳細要求（可空）
	MessageID    string    `gorm:"size:32;not null;default:''"` // 面板訊息 ID
	ChannelID    string    `gorm:"size:32;not null;default:''"`
	ThreadID     string    `gorm:"size:32;not null;default:''"` // 空 = 尚未開討論串
	ThreadName   string    `gorm:"size:100;not null;default:''"` // 隨機團名
	MaxMembers   int       `gorm:"not null;default:12"`         // 人數上限（含揪團人）
	Members      string    `gorm:"type:text;not null"`          // 逗號分隔的報名者 Discord ID
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (GroupActivity) TableName() string {
	return "group_activities"
}

// BotConfig 機器人設定（key-value）
type BotConfig struct {
	ID    int    `gorm:"primaryKey;autoIncrement"`
	Key   string `gorm:"size:100;not null;uniqueIndex"`
	Value string `gorm:"type:text;not null"`
}

// TableName 指定表名
func (BotConfig) TableName() string {
	return "bot_config"
}

// GroupThreadName 揪團隨機團名目錄
type GroupThreadName struct {
	ID           int    `gorm:"primaryKey;autoIncrement"`
	ActivityType string `gorm:"size:20;not null;index"` // "rainbow" or "fishing"
	Name         string `gorm:"size:100;not null"`
}

// TableName 指定表名
func (GroupThreadName) TableName() string {
	return "group_thread_names"
}

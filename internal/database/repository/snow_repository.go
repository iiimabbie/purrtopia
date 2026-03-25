package repository

import (
	"time"

	"purrtopia/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProxyEntry 代售清單項目
type ProxyEntry struct {
	DiscordID string
	Items     string
}

// SnowProxySellRepository 冰雪市場代售 Repository
type SnowProxySellRepository struct {
	db *gorm.DB
}

// NewSnowProxySellRepository 建立 SnowProxySellRepository
func NewSnowProxySellRepository(db *gorm.DB) *SnowProxySellRepository {
	return &SnowProxySellRepository{db: db}
}

// GetItemsByDiscordIDAndServerAfterTime 取得用戶在指定伺服器的代售物品字串
func (r *SnowProxySellRepository) GetItemsByDiscordIDAndServerAfterTime(discordID, serverRegion string, after time.Time) (string, error) {
	var entry models.SnowProxySell
	err := r.db.Where("discord_id = ? AND server_region = ? AND created_at >= ?", discordID, serverRegion, after).First(&entry).Error
	if err != nil {
		return "", err
	}
	return entry.Items, nil
}

// FindAllByServerAfterTime 取得指定伺服器、時間後的所有代售項目
func (r *SnowProxySellRepository) FindAllByServerAfterTime(serverRegion string, after time.Time) ([]ProxyEntry, error) {
	var entries []ProxyEntry
	err := r.db.Model(&models.SnowProxySell{}).
		Select("discord_id, items").
		Where("server_region = ? AND created_at >= ?", serverRegion, after).
		Order("created_at DESC").
		Scan(&entries).Error
	return entries, err
}

// Upsert 新增或更新代售項目
func (r *SnowProxySellRepository) Upsert(entry *models.SnowProxySell) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "discord_id"}, {Name: "server_region"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"items": entry.Items, "created_at": time.Now()}),
	}).Create(entry).Error
}

// DeleteByDiscordIDAndServerAfterTime 刪除指定用戶、伺服器、時間後的代售項目
func (r *SnowProxySellRepository) DeleteByDiscordIDAndServerAfterTime(discordID, serverRegion string, after time.Time) error {
	return r.db.Where("discord_id = ? AND server_region = ? AND created_at >= ?", discordID, serverRegion, after).
		Delete(&models.SnowProxySell{}).Error
}

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

// SeasonProxySellRepository 潮流季代售 Repository
type SeasonProxySellRepository struct {
	db *gorm.DB
}

// NewSeasonProxySellRepository 建立 SeasonProxySellRepository
func NewSeasonProxySellRepository(db *gorm.DB) *SeasonProxySellRepository {
	return &SeasonProxySellRepository{db: db}
}

// GetItemsByDiscordIDAndServerAfterTime 取得用戶在指定伺服器的代售物品字串
func (r *SeasonProxySellRepository) GetItemsByDiscordIDAndServerAfterTime(discordID, serverRegion string, after time.Time) (string, error) {
	var entry models.SeasonProxySell
	err := r.db.Where("discord_id = ? AND server_region = ? AND created_at >= ?", discordID, serverRegion, after).First(&entry).Error
	if err != nil {
		return "", err
	}
	return entry.Items, nil
}

// FindAllByServerAfterTime 取得指定伺服器、時間後的所有代售項目
func (r *SeasonProxySellRepository) FindAllByServerAfterTime(serverRegion string, after time.Time) ([]ProxyEntry, error) {
	var entries []ProxyEntry
	err := r.db.Model(&models.SeasonProxySell{}).
		Select("discord_id, items").
		Where("server_region = ? AND created_at >= ?", serverRegion, after).
		Order("created_at DESC").
		Scan(&entries).Error
	return entries, err
}

// Upsert 新增或更新代售項目
func (r *SeasonProxySellRepository) Upsert(entry *models.SeasonProxySell) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "discord_id"}, {Name: "server_region"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"items": entry.Items, "created_at": time.Now()}),
	}).Create(entry).Error
}

// DeleteByDiscordIDAndServerAfterTime 刪除指定用戶、伺服器、時間後的代售項目
func (r *SeasonProxySellRepository) DeleteByDiscordIDAndServerAfterTime(discordID, serverRegion string, after time.Time) error {
	return r.db.Where("discord_id = ? AND server_region = ? AND created_at >= ?", discordID, serverRegion, after).
		Delete(&models.SeasonProxySell{}).Error
}

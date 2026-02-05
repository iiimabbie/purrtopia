package repository

import (
	"purrtopia/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AvatarRepository 頭像 Repository (類似 JPA Repository)
type AvatarRepository struct {
	db *gorm.DB
}

// NewAvatarRepository 建立 AvatarRepository
func NewAvatarRepository(db *gorm.DB) *AvatarRepository {
	return &AvatarRepository{db: db}
}

// ExistsByDiscordID 檢查用戶是否已有頭像
func (r *AvatarRepository) ExistsByDiscordID(discordID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserAvatar{}).Where("discord_id = ?", discordID).Count(&count).Error
	return count > 0, err
}

// GetEmojiIDByDiscordID 取得用戶的 emoji ID
func (r *AvatarRepository) GetEmojiIDByDiscordID(discordID string) (*string, error) {
	var avatar models.UserAvatar
	err := r.db.Select("emoji_id").Where("discord_id = ?", discordID).First(&avatar).Error
	if err != nil {
		return nil, err
	}
	return avatar.EmojiID, nil
}

// Save 儲存或更新頭像 (類似 JPA save)
func (r *AvatarRepository) Save(avatar *models.UserAvatar) error {
	return r.db.Save(avatar).Error
}

// UpdateByDiscordID 根據 Discord ID 更新頭像
func (r *AvatarRepository) UpdateByDiscordID(discordID string, updates map[string]interface{}) error {
	return r.db.Model(&models.UserAvatar{}).Where("discord_id = ?", discordID).Updates(updates).Error
}

// Create 建立新頭像
func (r *AvatarRepository) Create(avatar *models.UserAvatar) error {
	return r.db.Create(avatar).Error
}

// FindAll 取得所有頭像
func (r *AvatarRepository) FindAll() ([]models.UserAvatar, error) {
	var avatars []models.UserAvatar
	err := r.db.Find(&avatars).Error
	return avatars, err
}

// FindByDiscordID 根據 Discord ID 查詢頭像
func (r *AvatarRepository) FindByDiscordID(discordID string) (*models.UserAvatar, error) {
	var avatar models.UserAvatar
	err := r.db.Where("discord_id = ?", discordID).First(&avatar).Error
	if err != nil {
		return nil, err
	}
	return &avatar, nil
}

// Upsert 新增或更新頭像 (如果存在則更新)
func (r *AvatarRepository) Upsert(avatar *models.UserAvatar) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "discord_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"discord_username", "game_uid", "image_url", "emoji_id", "emoji_name"}),
	}).Create(avatar).Error
}

// DrawnHeadRepository 抽頭記錄 Repository
type DrawnHeadRepository struct {
	db *gorm.DB
}

// NewDrawnHeadRepository 建立 DrawnHeadRepository
func NewDrawnHeadRepository(db *gorm.DB) *DrawnHeadRepository {
	return &DrawnHeadRepository{db: db}
}

// Create 建立抽頭記錄
func (r *DrawnHeadRepository) Create(drawnHead *models.DrawnHead) error {
	return r.db.Create(drawnHead).Error
}

// DrawnEmojiResult 抽過的 emoji 結果
type DrawnEmojiResult struct {
	EmojiID   string
	EmojiName string
}

// FindDrawnEmojisByDiscordID 取得用戶抽過的所有 emoji
func (r *DrawnHeadRepository) FindDrawnEmojisByDiscordID(discordID string) ([]DrawnEmojiResult, error) {
	var results []DrawnEmojiResult
	err := r.db.Table("drawn_heads dh").
		Select("ua.emoji_id, ua.emoji_name").
		Joins("JOIN user_avatars ua ON dh.avatar_id = ua.id").
		Where("dh.drawer_discord_id = ? AND ua.emoji_id IS NOT NULL AND ua.emoji_name IS NOT NULL", discordID).
		Group("ua.emoji_id, ua.emoji_name").
		Order("MAX(dh.drawn_at) DESC").
		Scan(&results).Error
	return results, err
}

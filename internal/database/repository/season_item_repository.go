package repository

import (
	"purrtopia/internal/database/models"

	"gorm.io/gorm"
)

// SeasonItemRepository 潮流季物品 Repository
type SeasonItemRepository struct {
	db *gorm.DB
}

// NewSeasonItemRepository 建立 SeasonItemRepository
func NewSeasonItemRepository(db *gorm.DB) *SeasonItemRepository {
	return &SeasonItemRepository{db: db}
}

// FindAll 取得所有物品（按 ID 排序）
func (r *SeasonItemRepository) FindAll() ([]models.SeasonItem, error) {
	var items []models.SeasonItem
	err := r.db.Order("id ASC").Find(&items).Error
	return items, err
}

// FindByCategory 取得指定分類的物品
func (r *SeasonItemRepository) FindByCategory(category int) ([]models.SeasonItem, error) {
	var items []models.SeasonItem
	err := r.db.Where("category = ?", category).Order("id ASC").Find(&items).Error
	return items, err
}

// FindByID 根據 ID 取得物品
func (r *SeasonItemRepository) FindByID(id int) (*models.SeasonItem, error) {
	var item models.SeasonItem
	err := r.db.First(&item, id).Error
	return &item, err
}

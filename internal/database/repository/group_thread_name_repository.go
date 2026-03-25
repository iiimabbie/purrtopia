package repository

import (
	"math/rand"

	"purrtopia/internal/database/models"

	"gorm.io/gorm"
)

// GroupThreadNameRepository 揪團隨機團名 repository
type GroupThreadNameRepository struct {
	db *gorm.DB
}

func NewGroupThreadNameRepository(db *gorm.DB) *GroupThreadNameRepository {
	return &GroupThreadNameRepository{db: db}
}

// FindRandomUnused 從指定類型中隨機取一個未被使用的團名。
// usedNames 為目前仍在進行中的活動已使用的團名。
// 若所有名字都已使用，則從全部名字中隨機取一個（fallback）。
func (r *GroupThreadNameRepository) FindRandomUnused(actType string, usedNames []string) (string, error) {
	var all []string
	if err := r.db.Model(&models.GroupThreadName{}).
		Where("activity_type = ?", actType).
		Pluck("name", &all).Error; err != nil {
		return "", err
	}
	if len(all) == 0 {
		return "", gorm.ErrRecordNotFound
	}

	// 建立已使用名稱的 set
	usedSet := make(map[string]struct{}, len(usedNames))
	for _, n := range usedNames {
		usedSet[n] = struct{}{}
	}

	// 篩選未使用的
	var available []string
	for _, n := range all {
		if _, used := usedSet[n]; !used {
			available = append(available, n)
		}
	}

	// 若有未使用的，從中隨機選
	if len(available) > 0 {
		return available[rand.Intn(len(available))], nil
	}

	// fallback：所有名字都被使用中，從全部中隨機取一個
	return all[rand.Intn(len(all))], nil
}

// CountByType 取得指定類型的團名數量（用於判斷是否需要 seed）
func (r *GroupThreadNameRepository) CountByType(actType string) (int64, error) {
	var count int64
	err := r.db.Model(&models.GroupThreadName{}).
		Where("activity_type = ?", actType).
		Count(&count).Error
	return count, err
}

// CreateBatch 批次新增團名
func (r *GroupThreadNameRepository) CreateBatch(names []models.GroupThreadName) error {
	return r.db.Create(&names).Error
}

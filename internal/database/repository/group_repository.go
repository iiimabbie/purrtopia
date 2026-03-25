package repository

import (
	"purrtopia/internal/database/models"

	"gorm.io/gorm"
)

// GroupActivityRepository 揪團活動 repository
type GroupActivityRepository struct {
	db *gorm.DB
}

func NewGroupActivityRepository(db *gorm.DB) *GroupActivityRepository {
	return &GroupActivityRepository{db: db}
}

// Create 新增揪團活動
func (r *GroupActivityRepository) Create(activity *models.GroupActivity) error {
	return r.db.Create(activity).Error
}

// FindByID 依 PK 取得活動
func (r *GroupActivityRepository) FindByID(id int) (*models.GroupActivity, error) {
	var activity models.GroupActivity
	err := r.db.First(&activity, id).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// UpdateMessageIDs 發布後更新 message_id 和 channel_id
func (r *GroupActivityRepository) UpdateMessageIDs(id int, msgID, chID string) error {
	return r.db.Model(&models.GroupActivity{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"message_id": msgID,
			"channel_id": chID,
		}).Error
}

// UpdateThread 更新討論串 ID 和報名成員列表
func (r *GroupActivityRepository) UpdateThread(id int, threadID, members string) error {
	return r.db.Model(&models.GroupActivity{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"thread_id": threadID,
			"members":   members,
		}).Error
}

// Delete 刪除指定活動記錄
func (r *GroupActivityRepository) Delete(id int) error {
	return r.db.Delete(&models.GroupActivity{}, id).Error
}

// FindActiveThreadNames 取得目前仍活躍的活動所使用的團名（避免重複選取）
func (r *GroupActivityRepository) FindActiveThreadNames(actType string) ([]string, error) {
	var names []string
	err := r.db.Model(&models.GroupActivity{}).
		Where("activity_type = ? AND thread_name != ''", actType).
		Pluck("thread_name", &names).Error
	return names, err
}

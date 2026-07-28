package repository

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"purrtopia/internal/database/models"

	"gorm.io/gorm"
)

// BotConfigRepository 機器人設定 Repository
type BotConfigRepository struct {
	db    *gorm.DB
	cache map[string]string
	mu    sync.RWMutex
}

// NewBotConfigRepository 建立 BotConfigRepository
func NewBotConfigRepository(db *gorm.DB) *BotConfigRepository {
	return &BotConfigRepository{
		db:    db,
		cache: make(map[string]string),
	}
}

// LoadAll 從 DB 載入所有設定到 cache
func (r *BotConfigRepository) LoadAll() error {
	var configs []models.BotConfig
	if err := r.db.Find(&configs).Error; err != nil {
		return fmt.Errorf("載入 bot_config 失敗: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]string, len(configs))
	for _, c := range configs {
		r.cache[c.Key] = c.Value
	}

	log.Printf("[bot_config] 已載入 %d 筆設定", len(configs))
	return nil
}

// Get 取得設定值（從 cache）
func (r *BotConfigRepository) Get(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cache[key]
}

// GetList 取得逗號分隔的設定值，回傳 slice
func (r *BotConfigRepository) GetList(key string) []string {
	val := r.Get(key)
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// GetMap 取得逗號分隔的設定值，回傳 set（map[string]struct{}）
func (r *BotConfigRepository) GetMap(key string) map[string]struct{} {
	list := r.GetList(key)
	m := make(map[string]struct{}, len(list))
	for _, v := range list {
		m[v] = struct{}{}
	}
	return m
}

// Set 設定值（寫入 DB + 更新 cache）
func (r *BotConfigRepository) Set(key, value string) error {
	config := models.BotConfig{Key: key, Value: value}
	err := r.db.Where("key = ?", key).Assign(config).FirstOrCreate(&config).Error
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache[key] = value
	return nil
}

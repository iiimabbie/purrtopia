package keywords

import (
	"fmt"
	"strings"

	"purrtopia/internal/config"
)

var cfg *config.KeywordConfig

// Init initializes the keywords module with configuration
func Init(keywordCfg *config.KeywordConfig) {
	cfg = keywordCfg
}

// MatchKeyword checks if the message contains any trigger keywords
func MatchKeyword(message string) bool {
	if cfg == nil || len(cfg.Triggers) == 0 {
		return false
	}

	messageLower := strings.ToLower(message)
	for _, keyword := range cfg.Triggers {
		if strings.Contains(messageLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// IsChannelEnabled checks if the keyword trigger is enabled for the given channel
func IsChannelEnabled(channelID string) bool {
	if cfg == nil {
		return false
	}

	// If no channels specified, all channels are enabled
	if len(cfg.EnabledChannels) == 0 {
		return true
	}

	for _, id := range cfg.EnabledChannels {
		if id == channelID {
			return true
		}
	}
	return false
}

// AI disclaimer footer
const aiDisclaimer = "\n\n-# 本訊息由 AI 自動判讀並分類，僅供參考。若有誤判情形，建議自主查閱左側頻道列表以獲取資訊。"

// GetResponseByType returns the appropriate response based on question type
func GetResponseByType(questionType string) string {
	if cfg == nil {
		return ""
	}

	var response string

	switch questionType {
	case "bubble":
		if cfg.BubbleLink != "" {
			response = fmt.Sprintf("關於泡泡的問題，請參考這個頻道：\n%s", cfg.BubbleLink)
		} else {
			response = "關於泡泡的問題，請查看相關頻道公告！"
		}

	case "daily_town":
		if cfg.DailyTownLink != "" {
			response = fmt.Sprintf("關於每日任務/小鎮報的問題，請參考這個頻道：\n%s", cfg.DailyTownLink)
		} else {
			response = "關於每日任務的問題，請查看每日小鎮報頻道！"
		}

	case "location":
		if cfg.HelperLink != "" {
			response = fmt.Sprintf("關於位置的問題，可以使用心動小助手查詢：\n%s", cfg.HelperLink)
		} else {
			response = "關於位置的問題，請使用心動小助手查詢！"
		}

	case "weather":
		if cfg.WeatherLink != "" {
			response = fmt.Sprintf("關於天氣/隕石/朵朵的問題，請參考這個頻道：\n%s", cfg.WeatherLink)
		} else if cfg.HelperLink != "" {
			response = fmt.Sprintf("關於天氣/隕石/朵朵的問題，可以使用心動小助手查詢：\n%s", cfg.HelperLink)
		} else {
			response = "關於天氣的問題，請使用心動小助手查詢！"
		}

	case "general":
		// general 類型不回覆，無法明確分類的問題不處理
		return ""

	default:
		return ""
	}

	return response + aiDisclaimer
}

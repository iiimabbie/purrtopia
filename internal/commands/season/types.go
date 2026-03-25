package season

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// 伺服器區域常數
const (
	ServerAsia   = "asia"   // 亞服
	ServerTWHKMO = "twhkmo" // 台港澳服
)

// 所有支援的伺服器區域
var AllServers = []string{ServerAsia, ServerTWHKMO}

// 潮流季主題顏色
const ColorSeason = 0x87CEEB

// 分頁常數
const proxySellPageSize = 10

// proxySellItem 代售物品
type proxySellItem struct {
	ID   int
	Name string
}

// 代售物品列表（每季更新）
var proxySellItems = []proxySellItem{
	// TODO: 新一季物品在這裡填入
}

// proxySellEntry 代表資料庫中的代售項目
type proxySellEntry struct {
	DiscordID string
	Items     string
}

// getItemNameByID 根據 ID 取得物品名稱
func getItemNameByID(id int) string {
	for _, item := range proxySellItems {
		if item.ID == id {
			return item.Name
		}
	}
	return ""
}

// parseItemIDs 解析逗號分隔的物品 ID 字串
func parseItemIDs(s string) []int {
	var ids []int
	for _, part := range strings.Split(s, ",") {
		var id int
		if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// itemIDsToNames 將物品 ID 轉換為名稱（按 ID 排序）
func itemIDsToNames(ids []int) []string {
	sortedIDs := make([]int, len(ids))
	copy(sortedIDs, ids)
	sort.Ints(sortedIDs)

	var names []string
	for _, id := range sortedIDs {
		if name := getItemNameByID(id); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// containsItemID 檢查物品 ID 是否在逗號分隔的字串中
func containsItemID(itemsStr string, targetID int) bool {
	ids := parseItemIDs(itemsStr)
	for _, id := range ids {
		if id == targetID {
			return true
		}
	}
	return false
}

// getLastSaturdayReset 取得上週六早上 6 點（台灣時間 UTC+8）
func getLastSaturdayReset() time.Time {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		loc = time.FixedZone("Asia/Taipei", 8*60*60)
	}

	now := time.Now().In(loc)
	daysSinceSaturday := int(now.Weekday()+1) % 7

	lastSaturday := now.AddDate(0, 0, -daysSinceSaturday)
	resetTime := time.Date(lastSaturday.Year(), lastSaturday.Month(), lastSaturday.Day(), 6, 0, 0, 0, loc)

	if now.Before(resetTime) {
		resetTime = resetTime.AddDate(0, 0, -7)
	}

	return resetTime.UTC()
}

// getServerDisplayName 取得伺服器區域的顯示名稱
func getServerDisplayName(server string) string {
	switch server {
	case ServerAsia:
		return "亞服"
	case ServerTWHKMO:
		return "台港澳服"
	default:
		return server
	}
}

// extractServerFromCustomID 從 customID 中提取伺服器區域
func extractServerFromCustomID(customID string) string {
	for _, server := range AllServers {
		if strings.HasSuffix(customID, "_"+server) {
			return server
		}
	}
	return ""
}

// extractPageFromCustomID 從 customID 中提取頁碼
func extractPageFromCustomID(customID string) int {
	for _, server := range AllServers {
		suffix := "_" + server
		if strings.HasSuffix(customID, suffix) {
			prefix := "season_sell_page_"
			trimmed := strings.TrimSuffix(customID, suffix)
			if strings.HasPrefix(trimmed, prefix) {
				var page int
				fmt.Sscanf(strings.TrimPrefix(trimmed, prefix), "%d", &page)
				return page
			}
		}
	}
	return 0
}

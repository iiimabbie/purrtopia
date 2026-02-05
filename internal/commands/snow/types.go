package snow

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

// 連結常數
const (
	snowBubbleLink     = "https://discord.com/channels/1438429535975641120/1467392127712628900/1467392390045372616"
	menuPriceLink      = "https://discord.com/channels/1438429535975641120/1452258561156710462/1462660483457876077"
	proxyBuyItemsImage = "https://duk.tw/vdkUPU.jpg"
)

// 冰雪主題顏色（冰藍色）
const ColorSnow = 0x87CEEB

// 分頁常數
const proxySellPageSize = 10

// 代購直排列表
var proxyBuyColumns = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M"}

// proxySellItem 代售物品
type proxySellItem struct {
	ID   int
	Name string
}

// 代售物品列表（按 ID 排序）
// ID 3 保留但不顯示在選單中
var proxySellItems = []proxySellItem{
	{1, "冰晶帝王蟹"},
	{2, "冰晶翻車魚"},
	// ID 3 保留（尚未推出）
	{4, "冰晶河魨"},
	{5, "冰晶海馬"},
	{6, "冰杯咖啡"},
	{7, "冰杯拿鐵"},
	{8, "白蘿蔔"},
	{9, "白蘿蔔泥肉"},
	{10, "白蘿蔔奶油濃湯"},
	{11, "原味糖霜鬆餅"},
	{12, "藍莓糖霜鬆餅"},
	{13, "樹莓糖霜鬆餅"},
	{14, "蘋果糖霜鬆餅"},
	{15, "橘子糖霜鬆餅"},
	{16, "極光晚宴"},
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
	// 先排序 ID 以保持顯示順序一致
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
// 如果現在還沒到週六早上 6 點，則返回上上週六
func getLastSaturdayReset() time.Time {
	// 台灣時區 (UTC+8)
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		// 找不到時區則使用固定偏移
		loc = time.FixedZone("Asia/Taipei", 8*60*60)
	}

	now := time.Now().In(loc)

	// 找到最近的週六
	daysSinceSaturday := int(now.Weekday()+1) % 7 // 週六 = 6，所以 (weekday + 1) % 7 得到距離週六的天數

	// 取得上週六早上 6 點
	lastSaturday := now.AddDate(0, 0, -daysSinceSaturday)
	resetTime := time.Date(lastSaturday.Year(), lastSaturday.Month(), lastSaturday.Day(), 6, 0, 0, 0, loc)

	// 如果還沒到這週六早上 6 點，則往回推一週
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
// 例如: "snow_proxy_buy_add_asia" -> "asia"
// 例如: "snow_sell_page_0_twhkmo" -> "twhkmo"
func extractServerFromCustomID(customID string) string {
	for _, server := range AllServers {
		if strings.HasSuffix(customID, "_"+server) {
			return server
		}
	}
	return ""
}

// extractPageFromCustomID 從 customID 中提取頁碼
// 例如: "snow_sell_page_2_asia" -> 2
func extractPageFromCustomID(customID string) int {
	for _, server := range AllServers {
		suffix := "_" + server
		if strings.HasSuffix(customID, suffix) {
			prefix := "snow_sell_page_"
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

// extractColumnFromCustomID 從 customID 中提取直排
// 例如: "snow_proxy_buy_row_select_asia_A" -> "A"
func extractColumnFromCustomID(customID string) string {
	parts := strings.Split(customID, "_")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// extractServerAndColumnFromRowSelectID 從列選擇 customID 中提取伺服器和直排
// 例如: "snow_proxy_buy_row_select_asia_A" -> ("asia", "A")
func extractServerAndColumnFromRowSelectID(customID string) (string, string) {
	// 格式: snow_proxy_buy_row_select_{server}_{column}
	prefix := "snow_proxy_buy_row_select_"
	if !strings.HasPrefix(customID, prefix) {
		return "", ""
	}
	rest := strings.TrimPrefix(customID, prefix)
	parts := strings.Split(rest, "_")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

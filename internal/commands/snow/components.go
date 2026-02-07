package snow

import (
	"fmt"
	"strings"

	"purrtopia/internal/component"
	"purrtopia/internal/database"

	"github.com/bwmarrin/discordgo"
)

// ============================================
// 按鈕建構器
// ============================================

// buildMainButtons 建構主選單按鈕
func buildMainButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_bubble").Label("⛄ 雪人泡泡位置").Secondary().Build()).
			AddButton(component.NewButton().CustomID("snow_server_twhkmo").Label("🇹🇼 代購代售【台港澳】").Primary().Build()).
			AddButton(component.NewButton().CustomID("snow_server_asia").Label("🌏 代購代售【亞服】").Primary().Build()).
			Build(),
	}
}

// buildOtherButtons 建構其他功能按鈕（排除指定按鈕）
func buildOtherButtons(exclude string) []discordgo.MessageComponent {
	row := component.NewActionRow()

	if exclude != "snow_bubble" {
		row.AddButton(component.NewButton().CustomID("snow_bubble").Label("⛄ 雪人泡泡位置").Primary().Build())
	}
	if exclude != "snow_server_twhkmo" {
		row.AddButton(component.NewButton().CustomID("snow_server_twhkmo").Label("🇹🇼 代購代售【台港澳】").Secondary().Build())
	}
	if exclude != "snow_server_asia" {
		row.AddButton(component.NewButton().CustomID("snow_server_asia").Label("🌏 代購代售【亞服】").Secondary().Build())
	}

	return []discordgo.MessageComponent{row.Build()}
}

// buildServerActionMenuButtons 建構伺服器動作選單按鈕（買/賣選擇）
func buildServerActionMenuButtons(server string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_action_buy_" + server).Label("🛒 代購功能").Primary().Build()).
			AddButton(component.NewButton().CustomID("snow_action_sell_" + server).Label("💰 代售功能").Primary().Build()).
			AddButton(component.NewButton().CustomID("snow_back_main").Label("⬅️ 返回").Secondary().Build()).
			Build(),
	}
}

// buildProxyBuyButtons 建構代購功能按鈕
func buildProxyBuyButtons(server string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_proxy_buy_add_"+server).Label("📝 登記可幫代購物品").Success().Build()).
			AddButton(component.NewButton().CustomID("snow_proxy_buy_search_"+server).Label("🔍 找代購物品").Primary().Build()).
			AddButton(component.NewButton().CustomID("snow_back_server_"+server).Label("⬅️ 返回").Secondary().Build()).
			Build(),
	}
}

// buildProxySellButtons 建構代售功能按鈕
func buildProxySellButtons(server string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_proxy_sell_add_"+server).Label("📝 登記").Success().Build()).
			AddButton(component.NewButton().CustomID("snow_proxy_sell_search_"+server).Label("🔍 找物品").Primary().Build()).
			AddButton(component.NewButton().CustomID("snow_proxy_sell_list_"+server).Label("📋 總覽").Primary().Build()).
			AddButton(component.NewButton().CustomID("snow_back_server_"+server).Label("⬅️ 返回").Secondary().Build()).
			Build(),
	}
}

// buildClearConfirmButtons 建構清除確認按鈕
func buildClearConfirmButtons(actionType, server string) []discordgo.MessageComponent {
	var yesID, noID string
	if actionType == "buy" {
		yesID = "snow_proxy_buy_clear_yes_" + server
		noID = "snow_proxy_buy_clear_no_" + server
	} else {
		yesID = "snow_proxy_sell_clear_yes_" + server
		noID = "snow_proxy_sell_clear_no_" + server
	}

	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID(yesID).Label("是").Danger().Build()).
			AddButton(component.NewButton().CustomID(noID).Label("否").Secondary().Build()).
			Build(),
	}
}

// ============================================
// 代售選單建構器
// ============================================

// proxySellSplitID 是代售選單分割點，ID < 此值放第一個選單，>= 放第二個
const proxySellSplitID = 18

// buildProxySellSelectMenu 建構代售登記下拉選單（分兩個選單避免 Discord 25 上限）
func buildProxySellSelectMenu(server, userID string) []discordgo.MessageComponent {
	// 取得用戶已登記的物品
	selectedItems := make(map[int]bool)
	existingItems, err := database.ProxySellRepo.GetItemsByDiscordIDAndServerAfterTime(userID, server, getLastSaturdayReset())
	if err == nil && existingItems != "" {
		for _, idStr := range strings.Split(existingItems, ",") {
			var id int
			if _, err := fmt.Sscanf(strings.TrimSpace(idStr), "%d", &id); err == nil {
				selectedItems[id] = true
			}
		}
	}

	// 分割物品
	var items1, items2 []proxySellItem
	for _, item := range proxySellItems {
		if item.ID < proxySellSplitID {
			items1 = append(items1, item)
		} else {
			items2 = append(items2, item)
		}
	}

	// 第一個選單
	select1 := component.NewSelect().
		CustomID("snow_proxy_sell_select_" + server).
		Placeholder("食物・農產品・魚（多選）").
		MinValues(0).
		MaxValues(len(items1))
	for _, item := range items1 {
		if selectedItems[item.ID] {
			select1.AddOptionDefault(item.Name, fmt.Sprintf("%d", item.ID), "")
		} else {
			select1.AddOption(item.Name, fmt.Sprintf("%d", item.ID), "")
		}
	}

	// 第二個選單
	select2 := component.NewSelect().
		CustomID("snow_proxy_sell_select2_" + server).
		Placeholder("蟲・花・鳥（多選）").
		MinValues(0).
		MaxValues(len(items2))
	for _, item := range items2 {
		if selectedItems[item.ID] {
			select2.AddOptionDefault(item.Name, fmt.Sprintf("%d", item.ID), "")
		} else {
			select2.AddOption(item.Name, fmt.Sprintf("%d", item.ID), "")
		}
	}

	return []discordgo.MessageComponent{
		component.NewActionRow().AddSelect(select1.Build()).Build(),
		component.NewActionRow().AddSelect(select2.Build()).Build(),
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_back_sell_" + server).Label("⬅️ 返回").Secondary().Build()).
			AddButton(component.NewButton().CustomID("snow_proxy_sell_clear_" + server).Label("🗑️ 清除登記").Danger().Build()).
			Build(),
	}
}

// buildProxySellSearchSelectMenu 建構代售搜尋下拉選單（分兩個）
func buildProxySellSearchSelectMenu(server string) []discordgo.MessageComponent {
	select1 := component.NewSelect().
		CustomID("snow_proxy_sell_search_select_" + server).
		Placeholder("搜尋：食物・飲品・活動").
		MinValues(1).
		MaxValues(1)

	select2 := component.NewSelect().
		CustomID("snow_proxy_sell_search_select2_" + server).
		Placeholder("搜尋：蝴蝶・植物・冬裝動物").
		MinValues(1).
		MaxValues(1)

	for _, item := range proxySellItems {
		if item.ID < proxySellSplitID {
			select1.AddOption(item.Name, fmt.Sprintf("%d", item.ID), "")
		} else {
			select2.AddOption(item.Name, fmt.Sprintf("%d", item.ID), "")
		}
	}

	return []discordgo.MessageComponent{
		component.NewActionRow().AddSelect(select1.Build()).Build(),
		component.NewActionRow().AddSelect(select2.Build()).Build(),
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_back_sell_" + server).Label("⬅️ 返回").Secondary().Build()).
			Build(),
	}
}

// buildProxySellListButtons 建構代售總覽分頁按鈕
func buildProxySellListButtons(page, totalPages int, server string) []discordgo.MessageComponent {
	row := component.NewActionRow()

	// 上一頁按鈕
	if page > 0 {
		row.AddButton(component.NewButton().
			CustomID(fmt.Sprintf("snow_sell_page_%d_%s", page-1, server)).
			Label("◀️ 上一頁").
			Secondary().
			Build())
	}

	// 下一頁按鈕
	if page < totalPages-1 {
		row.AddButton(component.NewButton().
			CustomID(fmt.Sprintf("snow_sell_page_%d_%s", page+1, server)).
			Label("下一頁 ▶️").
			Secondary().
			Build())
	}

	// 返回按鈕
	row.AddButton(component.NewButton().
		CustomID("snow_back_sell_" + server).
		Label("⬅️ 返回").
		Secondary().
		Build())

	return []discordgo.MessageComponent{row.Build()}
}

// ============================================
// 代購選單建構器
// ============================================

// buildProxyBuySelectMenus 建構代購登記下拉選單（直排 + 橫排）
func buildProxyBuySelectMenus(server, column string, selectedRows map[int]bool) []discordgo.MessageComponent {
	// 第一列：直排選擇下拉（A-M）- 單選
	colSelect := component.NewSelect().
		CustomID("snow_proxy_buy_col_select_" + server).
		Placeholder("選擇直排 A-M").
		MinValues(1).
		MaxValues(1)

	for _, col := range proxyBuyColumns {
		if col == column {
			colSelect.AddOptionDefault(col, col, "")
		} else {
			colSelect.AddOption(col, col, "")
		}
	}

	// 第二列：橫排選擇下拉（1-10）- 多選，使用 Default 顯示已選狀態
	rowSelect := component.NewSelect().
		CustomID(fmt.Sprintf("snow_proxy_buy_row_select_%s_%s", server, column)).
		Placeholder("選擇橫排 1-10（可多選，點擊切換）").
		MinValues(0).
		MaxValues(10)

	for r := 1; r <= 10; r++ {
		label := fmt.Sprintf("%d", r)
		value := fmt.Sprintf("%d", r)
		if selectedRows[r] {
			// 使用 AddOptionDefault 讓 Discord 原生顯示已選狀態
			rowSelect.AddOptionDefault(label, value, "")
		} else {
			rowSelect.AddOption(label, value, "")
		}
	}

	// 第三列：返回按鈕和清除按鈕
	row3 := component.NewActionRow().
		AddButton(component.NewButton().CustomID("snow_back_server_" + server).Label("⬅️ 返回選單").Secondary().Build()).
		AddButton(component.NewButton().CustomID("snow_proxy_buy_clear_" + server).Label("🗑️ 清除登記").Danger().Build())

	return []discordgo.MessageComponent{
		component.NewActionRow().AddSelect(colSelect.Build()).Build(),
		component.NewActionRow().AddSelect(rowSelect.Build()).Build(),
		row3.Build(),
	}
}

// getUserSelectedRowsForColumn 取得用戶在指定直排已選擇的橫排
func getUserSelectedRowsForColumn(userID, server, column string) map[int]bool {
	selectedRows := make(map[int]bool)

	// 使用 GORM Repository
	existingItems, err := database.ProxyBuyRepo.GetItemsByDiscordIDAndServerAfterTime(userID, server, getLastSaturdayReset())

	if err == nil && existingItems != "" {
		for _, item := range strings.Split(existingItems, ",") {
			item = strings.TrimSpace(item)
			if len(item) >= 2 && strings.HasPrefix(item, column) {
				var row int
				if _, err := fmt.Sscanf(item[1:], "%d", &row); err == nil {
					selectedRows[row] = true
				}
			}
		}
	}
	return selectedRows
}

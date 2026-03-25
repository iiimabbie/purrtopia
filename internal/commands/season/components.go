package season

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
			AddButton(component.NewButton().CustomID("season_action_sell_twhkmo").Label("🇹🇼 代售【台港澳】").Primary().Build()).
			AddButton(component.NewButton().CustomID("season_action_sell_asia").Label("🌏 代售【亞服】").Primary().Build()).
			Build(),
	}
}

// buildBackToMainButton 建構返回主選單按鈕
func buildBackToMainButton() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("season_back_main").Label("⬅️ 返回").Secondary().Build()).
			Build(),
	}
}

// buildProxySellButtons 建構代售功能按鈕
func buildProxySellButtons(server string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("season_proxy_sell_add_"+server).Label("📝 登記").Success().Build()).
			AddButton(component.NewButton().CustomID("season_proxy_sell_search_"+server).Label("🔍 找物品").Primary().Build()).
			AddButton(component.NewButton().CustomID("season_proxy_sell_list_"+server).Label("📋 總覽").Primary().Build()).
			AddButton(component.NewButton().CustomID("season_back_main").Label("⬅️ 返回").Secondary().Build()).
			Build(),
	}
}

// buildClearConfirmButtons 建構清除確認按鈕
func buildClearConfirmButtons(server string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("season_proxy_sell_clear_yes_"+server).Label("是").Danger().Build()).
			AddButton(component.NewButton().CustomID("season_proxy_sell_clear_no_"+server).Label("否").Secondary().Build()).
			Build(),
	}
}

// ============================================
// 代售選單建構器
// ============================================

// buildProxySellSelectMenu 建構代售登記下拉選單（按 category 分組）
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

	items1 := getSeasonItemsByCategory(1)
	items2 := getSeasonItemsByCategory(2)

	var rows []discordgo.MessageComponent

	if len(items1) > 0 {
		select1 := component.NewSelect().
			CustomID("season_proxy_sell_select_" + server).
			Placeholder("第一組物品（多選）").
			MinValues(0).
			MaxValues(len(items1))
		for _, item := range items1 {
			if selectedItems[item.ID] {
				select1.AddOptionDefault(item.Name, fmt.Sprintf("%d", item.ID), "")
			} else {
				select1.AddOption(item.Name, fmt.Sprintf("%d", item.ID), "")
			}
		}
		rows = append(rows, component.NewActionRow().AddSelect(select1.Build()).Build())
	}

	if len(items2) > 0 {
		select2 := component.NewSelect().
			CustomID("season_proxy_sell_select2_" + server).
			Placeholder("第二組物品（多選）").
			MinValues(0).
			MaxValues(len(items2))
		for _, item := range items2 {
			if selectedItems[item.ID] {
				select2.AddOptionDefault(item.Name, fmt.Sprintf("%d", item.ID), "")
			} else {
				select2.AddOption(item.Name, fmt.Sprintf("%d", item.ID), "")
			}
		}
		rows = append(rows, component.NewActionRow().AddSelect(select2.Build()).Build())
	}

	rows = append(rows, component.NewActionRow().
		AddButton(component.NewButton().CustomID("season_back_sell_"+server).Label("⬅️ 返回代售").Secondary().Build()).
		AddButton(component.NewButton().CustomID("season_proxy_sell_clear_"+server).Label("🗑️ 清除登記").Danger().Build()).
		Build())

	return rows
}

// buildProxySellSearchSelectMenu 建構代售搜尋下拉選單
func buildProxySellSearchSelectMenu(server string) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent

	items1 := getSeasonItemsByCategory(1)
	items2 := getSeasonItemsByCategory(2)

	if len(items1) > 0 {
		select1 := component.NewSelect().
			CustomID("season_proxy_sell_search_select_" + server).
			Placeholder("搜尋：第一組物品").
			MinValues(1).
			MaxValues(1)
		for _, item := range items1 {
			select1.AddOption(item.Name, fmt.Sprintf("%d", item.ID), "")
		}
		rows = append(rows, component.NewActionRow().AddSelect(select1.Build()).Build())
	}

	if len(items2) > 0 {
		select2 := component.NewSelect().
			CustomID("season_proxy_sell_search_select2_" + server).
			Placeholder("搜尋：第二組物品").
			MinValues(1).
			MaxValues(1)
		for _, item := range items2 {
			select2.AddOption(item.Name, fmt.Sprintf("%d", item.ID), "")
		}
		rows = append(rows, component.NewActionRow().AddSelect(select2.Build()).Build())
	}

	rows = append(rows, component.NewActionRow().
		AddButton(component.NewButton().CustomID("season_back_sell_"+server).Label("⬅️ 返回代售").Secondary().Build()).
		Build())

	return rows
}

// buildProxySellListButtons 建構代售總覽分頁按鈕
func buildProxySellListButtons(page, totalPages int, server string) []discordgo.MessageComponent {
	row := component.NewActionRow()

	if page > 0 {
		row.AddButton(component.NewButton().
			CustomID(fmt.Sprintf("season_sell_page_%d_%s", page-1, server)).
			Label("◀️ 上一頁").
			Secondary().
			Build())
	}

	if page < totalPages-1 {
		row.AddButton(component.NewButton().
			CustomID(fmt.Sprintf("season_sell_page_%d_%s", page+1, server)).
			Label("下一頁 ▶️").
			Secondary().
			Build())
	}

	row.AddButton(component.NewButton().
		CustomID("season_back_sell_" + server).
		Label("⬅️ 返回").
		Secondary().
		Build())

	return []discordgo.MessageComponent{row.Build()}
}

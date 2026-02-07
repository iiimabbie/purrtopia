package snow

import (
	"fmt"
	"log"
	"strings"

	"purrtopia/internal/commands"
	"purrtopia/internal/database"
	"purrtopia/internal/database/models"
	"purrtopia/internal/embed"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

// ============================================
// 代售處理器
// ============================================

// getProxySellEntries 取得指定伺服器自上次重置以來的所有代售資料
func getProxySellEntries(serverRegion string) ([]proxySellEntry, error) {
	resetTime := getLastSaturdayReset()

	// 使用 GORM Repository
	rows, err := database.ProxySellRepo.FindAllByServerAfterTime(serverRegion, resetTime)
	if err != nil {
		return nil, err
	}

	var entries []proxySellEntry
	for _, row := range rows {
		entries = append(entries, proxySellEntry{
			DiscordID: row.DiscordID,
			Items:     row.Items,
		})
	}
	return entries, nil
}

// handleProxySellAdd 處理代售登記按鈕
func handleProxySellAdd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	userID := getUserID(i)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildProxySellSelectEmbed(server)},
			Components: buildProxySellSelectMenu(server, userID),
		},
	})
}

// handleProxySellSelect 處理代售物品選擇（第一個選單，ID < proxySellSplitID）
func handleProxySellSelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	proxySellSave(s, i, true)
}

// handleProxySellSelect2 處理代售物品選擇（第二個選單，ID >= proxySellSplitID）
func handleProxySellSelect2(s *discordgo.Session, i *discordgo.InteractionCreate) {
	proxySellSave(s, i, false)
}

// proxySellSave 處理代售登記儲存，isFirstMenu 表示是否為第一個選單
func proxySellSave(s *discordgo.Session, i *discordgo.InteractionCreate, isFirstMenu bool) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	data := i.MessageComponentData()
	selectedValues := data.Values

	userID := getUserID(i)
	if userID == "" {
		respondWithError(s, i, "無法取得用戶資訊")
		return
	}

	// 取得現有登記，保留另一個選單的選擇
	var keepIDs []string
	existingItems, err := database.ProxySellRepo.GetItemsByDiscordIDAndServerAfterTime(userID, server, getLastSaturdayReset())
	if err == nil && existingItems != "" {
		for _, idStr := range strings.Split(existingItems, ",") {
			idStr = strings.TrimSpace(idStr)
			var id int
			if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil {
				// 保留另一個選單範圍的 ID
				if isFirstMenu && id >= proxySellSplitID {
					keepIDs = append(keepIDs, idStr)
				} else if !isFirstMenu && id < proxySellSplitID {
					keepIDs = append(keepIDs, idStr)
				}
			}
		}
	}

	// 合併：保留的 + 這次選的
	allIDs := append(keepIDs, selectedValues...)
	itemIDs := strings.Join(allIDs, ",")

	// 轉換為物品名稱用於顯示（按 ID 排序）
	var intIDs []int
	for _, v := range allIDs {
		var id int
		fmt.Sscanf(v, "%d", &id)
		intIDs = append(intIDs, id)
	}
	selectedNames := itemIDsToNames(intIDs)

	// 使用 GORM Repository 儲存
	err = database.ProxySellRepo.Upsert(&models.SnowProxySell{
		DiscordID:    userID,
		ServerRegion: server,
		Items:        itemIDs,
	})
	if err != nil {
		log.Printf("儲存代售資料失敗: %v", err)
		respondWithError(s, i, "儲存失敗，請稍後再試")
		return
	}

	log.Printf("[代售登記] 用戶=%s(%s) 伺服器=%s 物品=%s", getUserDisplayName(i), userID, server, strings.Join(selectedNames, ","))

	e := embed.New().
		Title("✅ 已儲存，可繼續選擇另一個選單").
		Description(fmt.Sprintf("【%s】目前登記的物品：\n**%s**\n\n選完後按「返回」即可。", getServerDisplayName(server), strings.Join(selectedNames, "、"))).
		Color(embed.ColorSuccess).
		Build()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{e},
			Components: buildProxySellSelectMenu(server, userID),
		},
	})
}

// handleProxySellSearch 處理代售搜尋按鈕
func handleProxySellSearch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	// 顯示選擇選單讓用戶選擇要搜尋的物品
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildProxySellSearchEmbed(server)},
			Components: buildProxySellSearchSelectMenu(server),
		},
	})
}

// handleProxySellSearchSelect 處理代售搜尋選擇
func handleProxySellSearchSelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	data := i.MessageComponentData()
	if len(data.Values) == 0 {
		respondWithError(s, i, "請選擇一個物品")
		return
	}

	var searchItemID int
	fmt.Sscanf(data.Values[0], "%d", &searchItemID)
	searchItemName := getItemNameByID(searchItemID)

	// userID := getUserID(i)
	// log.Printf("[代售查詢] 用戶=%s(%s) 伺服器=%s 搜尋物品=%s", getUserDisplayName(i), userID, server, searchItemName)

	entries, err := getProxySellEntries(server)
	if err != nil {
		log.Printf("查詢代售資料失敗: %v", err)
		respondWithError(s, i, "查詢失敗，請稍後再試")
		return
	}

	var result string
	count := 0
	for _, entry := range entries {
		// 檢查此用戶是否有搜尋的物品
		if containsItemID(entry.Items, searchItemID) {
			result += fmt.Sprintf("• <@%s>\n", entry.DiscordID)
			count++
		}
	}

	var e *discordgo.MessageEmbed
	if count == 0 {
		e = embed.New().
			Title(fmt.Sprintf("🔍【%s】找「%s」的代售者", getServerDisplayName(server), searchItemName)).
			Description("目前沒有人登記可以代售這個物品🥹").
			Color(ColorSnow).
			Build()
	} else {
		e = embed.New().
			Title(fmt.Sprintf("🔍【%s】找「%s」的代售者", getServerDisplayName(server), searchItemName)).
			Description(fmt.Sprintf("以下玩家可以幫你代售：\n\n%s", result)).
			Color(ColorSnow).
			FooterText(fmt.Sprintf("共 %d 人 • 本週資料", count)).
			Build()
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{e},
			Components: buildProxySellButtons(server),
		},
	})
}

// handleProxySellList 處理代售總覽按鈕
func handleProxySellList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	showProxySellListPage(s, i, 0, server)
}

// handleProxySellPage 處理代售總覽分頁
func handleProxySellPage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	server := extractServerFromCustomID(customID)
	page := extractPageFromCustomID(customID)

	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	showProxySellListPage(s, i, page, server)
}

// showProxySellListPage 顯示代售總覽分頁
func showProxySellListPage(s *discordgo.Session, i *discordgo.InteractionCreate, page int, server string) {
	entries, err := getProxySellEntries(server)
	if err != nil {
		log.Printf("查詢代售資料失敗: %v", err)
		respondWithError(s, i, "查詢失敗，請稍後再試")
		return
	}

	totalCount := len(entries)
	totalPages := (totalCount + proxySellPageSize - 1) / proxySellPageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// 限制頁碼範圍
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	var e *discordgo.MessageEmbed
	if totalCount == 0 {
		e = embed.New().
			Title(fmt.Sprintf("📋 代售總覽【%s】", getServerDisplayName(server))).
			Description("目前還沒有人登記代售物品~\n你可以成為第一個！").
			Color(ColorSnow).
			Build()
	} else {
		// 取得當前頁的資料
		start := page * proxySellPageSize
		end := start + proxySellPageSize
		if end > totalCount {
			end = totalCount
		}

		var result string
		for idx, entry := range entries[start:end] {
			// 將物品 ID 轉換為名稱
			ids := parseItemIDs(entry.Items)
			names := itemIDsToNames(ids)
			itemsDisplay := strings.Join(names, "、")
			if itemsDisplay == "" {
				itemsDisplay = entry.Items
			}
			result += fmt.Sprintf("%d. <@%s>: %s\n", start+idx+1, entry.DiscordID, itemsDisplay)
		}

		e = embed.New().
			Title(fmt.Sprintf("📋 代售總覽【%s】", getServerDisplayName(server))).
			Description(result).
			Color(ColorSnow).
			FooterText(fmt.Sprintf("第 %d/%d 頁 • 共 %d 筆 • 本週資料", page+1, totalPages, totalCount)).
			Build()
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{e},
			Components: buildProxySellListButtons(page, totalPages, server),
		},
	})
}

// handleProxySellClear 顯示代售清除確認
func handleProxySellClear(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildClearConfirmEmbed(server)},
			Components: buildClearConfirmButtons(server),
		},
	})
}

// handleProxySellClearConfirm 確認清除代售 - 刪除資料並返回代售選單
func handleProxySellClearConfirm(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	userID := getUserID(i)
	if userID == "" {
		respondWithError(s, i, "無法取得用戶資訊")
		return
	}

	// 使用 GORM Repository 刪除
	err := database.ProxySellRepo.DeleteByDiscordIDAndServerAfterTime(userID, server, getLastSaturdayReset())
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("刪除代售資料失敗: %v", err)
	}

	log.Printf("[代售清除] 用戶=%s(%s) 伺服器=%s", getUserDisplayName(i), userID, server)

	// 返回代售選單
	showProxySellMenu(s, i, server)
}

// handleProxySellClearCancel 取消清除代售 - 返回登記頁面
func handleProxySellClearCancel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	userID := getUserID(i)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildProxySellSelectEmbed(server)},
			Components: buildProxySellSelectMenu(server, userID),
		},
	})
}

// registerProxySellHandlers 註冊代售相關處理器
func registerProxySellHandlers() {
	for _, server := range AllServers {
		// 代售子按鈕
		commands.RegisterComponent("snow_proxy_sell_add_"+server, handleProxySellAdd)
		commands.RegisterComponent("snow_proxy_sell_search_"+server, handleProxySellSearch)
		commands.RegisterComponent("snow_proxy_sell_list_"+server, handleProxySellList)
		commands.RegisterComponent("snow_proxy_sell_clear_"+server, handleProxySellClear)
		commands.RegisterComponent("snow_proxy_sell_clear_yes_"+server, handleProxySellClearConfirm)
		commands.RegisterComponent("snow_proxy_sell_clear_no_"+server, handleProxySellClearCancel)

		// 選擇選單處理器
		commands.RegisterComponent("snow_proxy_sell_select_"+server, handleProxySellSelect)
		commands.RegisterComponent("snow_proxy_sell_select2_"+server, handleProxySellSelect2)
		commands.RegisterComponent("snow_proxy_sell_search_select_"+server, handleProxySellSearchSelect)
		commands.RegisterComponent("snow_proxy_sell_search_select2_"+server, handleProxySellSearchSelect)

		// 分頁處理器（0-9 頁）
		for page := 0; page < 10; page++ {
			commands.RegisterComponent(fmt.Sprintf("snow_sell_page_%d_%s", page, server), handleProxySellPage)
		}
	}
}

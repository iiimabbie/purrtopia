package snow

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"purrtopia/internal/commands"
	"purrtopia/internal/database"
	"purrtopia/internal/database/models"
	"purrtopia/internal/embed"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

// ============================================
// 代購處理器
// ============================================

// handleProxyBuyAdd 處理代購登記按鈕
func handleProxyBuyAdd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	userID := getUserID(i)
	defaultColumn := "A"
	selectedRows := getUserSelectedRowsForColumn(userID, server, defaultColumn)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{buildProxyBuySelectEmbed(server, defaultColumn)},
			Components:  buildProxyBuySelectMenus(server, defaultColumn, selectedRows),
			Attachments: &[]*discordgo.MessageAttachment{},
		},
	})
}

// handleProxyBuyColumnSelect 處理代購直排選擇
func handleProxyBuyColumnSelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	data := i.MessageComponentData()
	if len(data.Values) == 0 {
		respondWithError(s, i, "請選擇一個直排")
		return
	}
	selectedColumn := data.Values[0]
	userID := getUserID(i)

	// 取得用戶在此直排的已選橫排
	selectedRows := getUserSelectedRowsForColumn(userID, server, selectedColumn)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildProxyBuySelectEmbed(server, selectedColumn)},
			Components: buildProxyBuySelectMenus(server, selectedColumn, selectedRows),
		},
	})
}

// handleProxyBuyRowSelect 處理代購橫排選擇（多選 1-10）
// 選擇後自動儲存
func handleProxyBuyRowSelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	server, column := extractServerAndColumnFromRowSelectID(customID)
	if server == "" || column == "" {
		respondWithError(s, i, "無法取得伺服器或直排資訊")
		return
	}

	data := i.MessageComponentData()
	userID := getUserID(i)
	if userID == "" {
		respondWithError(s, i, "無法取得用戶資訊")
		return
	}

	// 從選擇中建立已選橫排 map
	selectedRows := make(map[int]bool)
	for _, v := range data.Values {
		var row int
		fmt.Sscanf(v, "%d", &row)
		selectedRows[row] = true
	}

	// 取得所有直排的現有物品（不只是這個直排）
	existingItems, err := database.ProxyBuyRepo.GetItemsByDiscordIDAndServerAfterTime(userID, server, getLastSaturdayReset())

	itemSet := make(map[string]bool)
	if err == nil && existingItems != "" {
		for _, item := range strings.Split(existingItems, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				// 保留其他直排的物品
				if len(item) >= 1 && !strings.HasPrefix(item, column) {
					itemSet[item] = true
				}
			}
		}
	}

	// 加入目前直排的選擇
	for row := 1; row <= 10; row++ {
		if selectedRows[row] {
			itemSet[fmt.Sprintf("%s%d", column, row)] = true
		}
	}

	// 轉換為排序後的字串
	var allItems []string
	for item := range itemSet {
		allItems = append(allItems, item)
	}
	sort.Strings(allItems)
	finalItems := strings.Join(allItems, ",")

	// 使用 GORM Repository 儲存
	err = database.ProxyBuyRepo.Upsert(&models.SnowProxyBuy{
		DiscordID:    userID,
		ServerRegion: server,
		Items:        finalItems,
	})
	if err != nil {
		log.Printf("儲存代購資料失敗: %v", err)
	} else {
		log.Printf("[代購登記] 用戶=%s(%s) 伺服器=%s 直排=%s 物品=%s", getUserDisplayName(i), userID, server, column, finalItems)
	}

	// 更新 UI
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildProxyBuySelectEmbed(server, column)},
			Components: buildProxyBuySelectMenus(server, column, selectedRows),
		},
	})
}

// handleProxyBuySearch 處理代購搜尋按鈕
func handleProxyBuySearch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	// 顯示 Modal 讓用戶輸入要搜尋的座標
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "snow_proxy_buy_search_modal_" + server,
			Title:    "🔍 找代購物品",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "coordinate",
							Label:       "輸入物品座標（僅單個如A10）",
							Style:       discordgo.TextInputShort,
							Placeholder: "A1",
							Required:    true,
							MinLength:   2,
							MaxLength:   3,
						},
					},
				},
			},
		},
	})
}

// handleProxyBuySearchModal 處理代購搜尋 Modal 提交
func handleProxyBuySearchModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.ModalSubmitData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	data := i.ModalSubmitData()
	coordinate := strings.ToUpper(strings.TrimSpace(data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value))

	// 驗證座標格式（A-M 後接 1-10）
	if len(coordinate) < 2 || len(coordinate) > 3 {
		respondWithError(s, i, "座標格式錯誤，請輸入像 A1, M9 這樣的格式")
		return
	}

	col := string(coordinate[0])
	rowStr := coordinate[1:]
	var row int
	fmt.Sscanf(rowStr, "%d", &row)

	validCols := "ABCDEFGHIJKLM"
	if !strings.Contains(validCols, col) || row < 1 || row > 10 {
		respondWithError(s, i, "座標格式錯誤，直排 A-M，橫排 1-10")
		return
	}

	userID := getUserID(i)
	log.Printf("[代購查詢] 用戶=%s(%s) 伺服器=%s 搜尋座標=%s", getUserDisplayName(i), userID, server, coordinate)

	// 使用 GORM Repository 搜尋
	rows, err := database.ProxyBuyRepo.FindAllByServerAfterTime(server, getLastSaturdayReset())
	if err != nil {
		log.Printf("查詢代購資料失敗: %v", err)
		respondWithError(s, i, "查詢失敗，請稍後再試")
		return
	}

	var result string
	count := 0
	for _, row := range rows {
		// 檢查此用戶是否有搜尋的座標
		for _, item := range strings.Split(row.Items, ",") {
			if strings.TrimSpace(item) == coordinate {
				result += fmt.Sprintf("• <@%s>\n", row.DiscordID)
				count++
				break
			}
		}
	}

	var e *discordgo.MessageEmbed
	if count == 0 {
		e = embed.New().
			Title(fmt.Sprintf("🔍【%s】找「%s」的代購者", getServerDisplayName(server), coordinate)).
			Description("目前沒有人登記可以代購這個物品🥹").
			Image(proxyBuyItemsImage).
			Color(ColorSnow).
			Build()
	} else {
		e = embed.New().
			Title(fmt.Sprintf("🔍【%s】找「%s」的代購者", getServerDisplayName(server), coordinate)).
			Description(fmt.Sprintf("以下玩家可以幫你代購：\n\n%s", result)).
			Image(proxyBuyItemsImage).
			Color(ColorSnow).
			FooterText(fmt.Sprintf("共 %d 人 • 本週資料", count)).
			Build()
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{e},
			Components:  buildProxyBuyButtons(server),
			Attachments: &[]*discordgo.MessageAttachment{},
		},
	})
}

// handleProxyBuyClear 顯示代購清除確認
func handleProxyBuyClear(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{buildClearConfirmEmbed("buy", server)},
			Components:  buildClearConfirmButtons("buy", server),
			Attachments: &[]*discordgo.MessageAttachment{},
		},
	})
}

// handleProxyBuyClearConfirm 確認清除代購 - 刪除資料並返回代購選單
func handleProxyBuyClearConfirm(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	err := database.ProxyBuyRepo.DeleteByDiscordIDAndServerAfterTime(userID, server, getLastSaturdayReset())
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("刪除代購資料失敗: %v", err)
	}

	log.Printf("[代購清除] 用戶=%s(%s) 伺服器=%s", getUserDisplayName(i), userID, server)

	// 返回代購選單
	showProxyBuyMenu(s, i, server)
}

// handleProxyBuyClearCancel 取消清除代購 - 返回登記頁面
func handleProxyBuyClearCancel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	userID := getUserID(i)
	defaultColumn := "A"
	selectedRows := getUserSelectedRowsForColumn(userID, server, defaultColumn)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{buildProxyBuySelectEmbed(server, defaultColumn)},
			Components:  buildProxyBuySelectMenus(server, defaultColumn, selectedRows),
			Attachments: &[]*discordgo.MessageAttachment{},
		},
	})
}

// registerProxyBuyHandlers 註冊代購相關處理器
func registerProxyBuyHandlers() {
	// 代購子按鈕（通用處理器）
	for _, server := range AllServers {
		commands.RegisterComponent("snow_proxy_buy_add_"+server, handleProxyBuyAdd)
		commands.RegisterComponent("snow_proxy_buy_search_"+server, handleProxyBuySearch)
		commands.RegisterComponent("snow_proxy_buy_col_select_"+server, handleProxyBuyColumnSelect)
		commands.RegisterComponent("snow_proxy_buy_clear_"+server, handleProxyBuyClear)
		commands.RegisterComponent("snow_proxy_buy_clear_yes_"+server, handleProxyBuyClearConfirm)
		commands.RegisterComponent("snow_proxy_buy_clear_no_"+server, handleProxyBuyClearCancel)

		// 代購橫排選擇處理器
		for _, col := range proxyBuyColumns {
			commands.RegisterComponent(fmt.Sprintf("snow_proxy_buy_row_select_%s_%s", server, col), handleProxyBuyRowSelect)
		}

		// Modal 處理器
		commands.RegisterModal("snow_proxy_buy_search_modal_"+server, handleProxyBuySearchModal)
	}
}

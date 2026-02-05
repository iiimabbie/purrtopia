package snow

import (
	"purrtopia/internal/commands"

	"github.com/bwmarrin/discordgo"
)

func init() {
	// 註冊指令
	commands.RegisterCommand(snowSeasonCommand, SnowSeasonHandler)

	// 註冊按鈕處理器
	commands.RegisterComponent("snow_bubble", handleSnowBubble)

	// 伺服器選擇處理器（主選單 -> 伺服器動作選單）
	for _, server := range AllServers {
		commands.RegisterComponent("snow_server_"+server, handleServerSelect)
	}

	// 動作選擇處理器（伺服器動作選單 -> 買/賣選單）
	for _, server := range AllServers {
		commands.RegisterComponent("snow_action_buy_"+server, handleActionBuy)
		commands.RegisterComponent("snow_action_sell_"+server, handleActionSell)
	}

	// 返回按鈕
	commands.RegisterComponent("snow_back_main", handleBackToMain)
	for _, server := range AllServers {
		commands.RegisterComponent("snow_back_server_"+server, handleBackToServerActionMenu)
		commands.RegisterComponent("snow_back_buy_"+server, handleBackToBuy)
		commands.RegisterComponent("snow_back_sell_"+server, handleBackToSell)
	}

	// 註冊代購和代售處理器
	registerProxyBuyHandlers()
	registerProxySellHandlers()
}

var snowSeasonCommand = &discordgo.ApplicationCommand{
	Name:        "冰雪季",
	Description: "冰雪季相關資訊與功能",
}

// ============================================
// 主指令處理器
// ============================================

// SnowSeasonHandler 冰雪季指令主處理器
func SnowSeasonHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildMainEmbed()},
			Components: buildMainButtons(),
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// ============================================
// 導航處理器
// ============================================

// handleSnowBubble 處理雪人泡泡按鈕
func handleSnowBubble(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildSnowBubbleEmbed()},
			Components: buildOtherButtons("snow_bubble"),
		},
	})
}

// handleServerSelect 處理伺服器選擇按鈕
func handleServerSelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildServerActionMenuEmbed(server)},
			Components: buildServerActionMenuButtons(server),
		},
	})
}

// handleActionBuy 處理代購功能按鈕
func handleActionBuy(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}
	showProxyBuyMenu(s, i, server)
}

// handleActionSell 處理代售功能按鈕
func handleActionSell(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}
	showProxySellMenu(s, i, server)
}

// showProxyBuyMenu 顯示代購選單
func showProxyBuyMenu(s *discordgo.Session, i *discordgo.InteractionCreate, server string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{buildProxyBuyMenuEmbed(server)},
			Components:  buildProxyBuyButtons(server),
			Attachments: &[]*discordgo.MessageAttachment{},
		},
	})
}

// showProxySellMenu 顯示代售選單
func showProxySellMenu(s *discordgo.Session, i *discordgo.InteractionCreate, server string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildProxySellEmbed(server)},
			Components: buildProxySellButtons(server),
		},
	})
}

// handleBackToMain 處理返回主選單按鈕
func handleBackToMain(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{buildMainEmbed()},
			Components:  buildMainButtons(),
			Attachments: &[]*discordgo.MessageAttachment{}, // 清除附件
		},
	})
}

// handleBackToServerActionMenu 處理返回伺服器動作選單按鈕
func handleBackToServerActionMenu(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{buildServerActionMenuEmbed(server)},
			Components:  buildServerActionMenuButtons(server),
			Attachments: &[]*discordgo.MessageAttachment{}, // 清除附件
		},
	})
}

// handleBackToBuy 處理返回代購選單按鈕
func handleBackToBuy(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}
	showProxyBuyMenu(s, i, server)
}

// handleBackToSell 處理返回代售選單按鈕
func handleBackToSell(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}
	showProxySellMenu(s, i, server)
}

// ============================================
// 輔助函式
// ============================================

// getUserID 取得用戶 ID
func getUserID(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

// getUserDisplayName 取得用戶顯示名稱（優先使用伺服器暱稱）
func getUserDisplayName(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		// 優先使用伺服器暱稱
		if i.Member.Nick != "" {
			return i.Member.Nick
		}
		// 其次使用全局顯示名稱
		if i.Member.User.GlobalName != "" {
			return i.Member.User.GlobalName
		}
		// 最後使用用戶名
		return i.Member.User.Username
	}
	if i.User != nil {
		if i.User.GlobalName != "" {
			return i.User.GlobalName
		}
		return i.User.Username
	}
	return ""
}

// respondWithError 回應錯誤訊息
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{buildErrorEmbed(message)},
			Components:  buildMainButtons(),
			Attachments: &[]*discordgo.MessageAttachment{}, // 清除附件
		},
	})
}

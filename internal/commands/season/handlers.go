package season

import (
	"purrtopia/internal/commands"

	"github.com/bwmarrin/discordgo"
)

func init() {
	// 註冊指令
	commands.RegisterCommand(seasonCommand, SeasonHandler)

	// 代售功能按鈕（主選單直接進入代售）
	for _, server := range AllServers {
		commands.RegisterComponent("season_action_sell_"+server, handleActionSell)
	}

	// 返回按鈕
	commands.RegisterComponent("season_back_main", handleBackToMain)
	for _, server := range AllServers {
		commands.RegisterComponent("season_back_sell_"+server, handleBackToSell)
	}

	// 註冊代售處理器
	registerProxySellHandlers()
}

var seasonCommand = &discordgo.ApplicationCommand{
	Name:        "潮流季",
	Description: "潮流季相關資訊與功能",
}

// ============================================
// 主指令處理器
// ============================================

// SeasonHandler 潮流季指令主處理器
func SeasonHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

// handleActionSell 處理代售功能按鈕
func handleActionSell(s *discordgo.Session, i *discordgo.InteractionCreate) {
	server := extractServerFromCustomID(i.MessageComponentData().CustomID)
	if server == "" {
		respondWithError(s, i, "無法取得伺服器資訊")
		return
	}
	showProxySellMenu(s, i, server)
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
			Attachments: &[]*discordgo.MessageAttachment{},
		},
	})
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
		if i.Member.Nick != "" {
			return i.Member.Nick
		}
		if i.Member.User.GlobalName != "" {
			return i.Member.User.GlobalName
		}
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
			Attachments: &[]*discordgo.MessageAttachment{},
		},
	})
}

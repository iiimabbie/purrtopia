package snow

import (
	"fmt"

	"purrtopia/internal/embed"

	"github.com/bwmarrin/discordgo"
)

// ============================================
// Embed 建構器
// ============================================

// buildMainEmbed 建構主選單 Embed
func buildMainEmbed() *discordgo.MessageEmbed {
	return embed.New().
		Title("❄️ 冰雪季功能選單").
		Description("### 冰雪季懶人包 ➡️ https://discord.com/channels/1438429535975641120/1438749728446873654/1466829733957341185\n ### 冰雪市場使用方式 ➡️ https://discord.com/channels/1438429535975641120/1467541407378964716/1467551966996467956").
		FooterText("代購功能因不實用而移除。").
		Color(ColorSnow).
		Build()
}

// buildSnowBubbleEmbed 建構雪人泡泡位置 Embed
func buildSnowBubbleEmbed() *discordgo.MessageEmbed {
	return embed.New().
		Title("⛄ 雪人泡泡位置").
		Description(fmt.Sprintf("在這個頻道 ⬇️\n%s", snowBubbleLink)).
		Color(ColorSnow).
		Build()
}

// buildProxySellEmbed 建構代售功能 Embed
func buildProxySellEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("💰 代售功能【%s】", getServerDisplayName(server))).
		Description("### 你可以登記你本週可以代售的物品，或是搜尋有誰可以幫你代售。\n*登記資料會在**每週六早上 6 點清空**，因應遊戲內每週更新一次。*").
		Color(ColorSnow).
		FooterText("登記代表你接受在discord被標記，會有人找你代售哦🌺。\n另外請鄰居們記得**餵食**幫助你代售的發展家。 ").
		Build()
}

// buildProxySellSelectEmbed 建構代售登記選擇 Embed
func buildProxySellSelectEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("💰 登記代售物品【%s】", getServerDisplayName(server))).
		Description("請選擇你本週可以代售的物品：").
		Color(ColorSnow).
		Build()
}

// buildProxySellSearchEmbed 建構代售搜尋 Embed
func buildProxySellSearchEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("🔍 找可代售物品【%s】", getServerDisplayName(server))).
		Description("請選擇你想找的物品：").
		Color(ColorSnow).
		Build()
}

// buildClearConfirmEmbed 建構清除確認 Embed
func buildClearConfirmEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("🗑️ 清除代售登記【%s】", getServerDisplayName(server))).
		Description("資料真的都會不見唷🥺").
		Color(embed.ColorWarning).
		Build()
}

// buildErrorEmbed 建構錯誤 Embed
func buildErrorEmbed(message string) *discordgo.MessageEmbed {
	return embed.New().
		Title("❌ 錯誤").
		Description(message).
		Color(embed.ColorError).
		Build()
}

// buildSuccessEmbed 建構成功 Embed
func buildSuccessEmbed(title, description string) *discordgo.MessageEmbed {
	return embed.New().
		Title(title).
		Description(description).
		Color(embed.ColorSuccess).
		Build()
}

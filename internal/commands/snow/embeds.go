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
		Description("冰雪季懶人包 ➡️ https://discord.com/channels/1438429535975641120/1438749728446873654/1466829733957341185\n 冰雪市場使用方式 ➡️ https://discord.com/channels/1438429535975641120/1467541407378964716/1467551966996467956").
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

// buildServerActionMenuEmbed 建構伺服器動作選單 Embed（選擇買/賣）
func buildServerActionMenuEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("❄️ 冰雪季【%s】", getServerDisplayName(server))).
		Description("⚠️代購功能較為繁瑣為DC機制導致，若有不理解都歡迎於頻道詢問。\n\n因為操作有一定複雜度，不一定要使用bot的代購\n也可於頻道直接詢問有誰有什麼物品。\n————————————————————————————\n\n⚠️代售功能就較為簡單，歡迎大家多多使用。\n————————————————————————————\n\n").
		FooterText("辛苦大家腦力激盪一下橫排數字記憶🙇🏻‍♀️").
		Color(ColorSnow).
		Build()
}

// buildProxyBuyEmbed 建構代購功能 Embed
func buildProxyBuyEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("🛒 代購功能【%s】", getServerDisplayName(server))).
		Description("這裡可以**登記你能幫別人代購**的物品，或是搜尋有誰可以幫你代購。").
		Color(ColorSnow).
		Build()
}

// buildProxyBuyMenuEmbed 建構代購選單 Embed（含圖片）
func buildProxyBuyMenuEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("🛒 代購功能【%s】", getServerDisplayName(server))).
		Description("這裡可以登記你能幫別人代購的物品，或是搜尋有誰可以幫你代購。\n從左方直列英文開始再來是橫向數字，例如傢俱床就是`M9`").
		Image(proxyBuyItemsImage).
		FooterText("圖源來自小紅書@露娜，已取得引用同意。").
		Color(ColorSnow).
		Build()
}

// buildProxySellEmbed 建構代售功能 Embed
func buildProxySellEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("💰 代售功能【%s】", getServerDisplayName(server))).
		Description("你可以登記你本週可以代售的物品，或是搜尋有誰可以幫你代售。\n*登記資料會在**每週六早上 6 點清空**，因應遊戲內每週更新一次。*").
		Color(ColorSnow).
		FooterText("登記代表你接受在discord被標記，會有人找你代售哦🌺。\n另外請鄰居們記得餵食幫助你代售的發展家。 ").
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

// buildProxyBuySelectEmbed 建構代購登記選擇 Embed
func buildProxyBuySelectEmbed(server, column string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("📝 登記代購物品【%s】", getServerDisplayName(server))).
		Description(fmt.Sprintf("1️⃣ 選擇直排（A-M）\n2️⃣ 選擇橫排（1-10，**可多選**，選完自動儲存）\n\n目前直排：**%s**", column)).
		Image(proxyBuyItemsImage).
		FooterText("圖源來自小紅書@露娜，已取得引用同意。").
		Color(ColorSnow).
		Build()
}

// buildClearConfirmEmbed 建構清除確認 Embed
func buildClearConfirmEmbed(actionType, server string) *discordgo.MessageEmbed {
	var title string
	if actionType == "buy" {
		title = fmt.Sprintf("🗑️ 清除代購登記【%s】", getServerDisplayName(server))
	} else {
		title = fmt.Sprintf("🗑️ 清除代售登記【%s】", getServerDisplayName(server))
	}

	return embed.New().
		Title(title).
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

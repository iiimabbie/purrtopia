package snow

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"purrtopia/internal/commands"
	"purrtopia/internal/component"
	"purrtopia/internal/database"
	"purrtopia/internal/embed"

	"github.com/bwmarrin/discordgo"
)

// Server region constants
const (
	ServerAsia   = "asia"   // 亞服
	ServerTWHKMO = "twhkmo" // 台港澳服
)

func init() {
	// Register command
	commands.RegisterCommand(snowSeasonCommand, SnowSeasonHandler)

	// Register button handlers
	commands.RegisterComponent("snow_bubble", handleSnowBubble)

	// Server selection handlers (main menu -> server-specific action menu)
	commands.RegisterComponent("snow_server_asia", makeServerActionMenuHandler(ServerAsia))
	commands.RegisterComponent("snow_server_twhkmo", makeServerActionMenuHandler(ServerTWHKMO))

	// Action selection handlers (server action menu -> buy/sell menu)
	commands.RegisterComponent("snow_action_buy_asia", makeActionHandler(ServerAsia, "buy"))
	commands.RegisterComponent("snow_action_buy_twhkmo", makeActionHandler(ServerTWHKMO, "buy"))
	commands.RegisterComponent("snow_action_sell_asia", makeActionHandler(ServerAsia, "sell"))
	commands.RegisterComponent("snow_action_sell_twhkmo", makeActionHandler(ServerTWHKMO, "sell"))

	// Proxy buy sub-buttons (with server region)
	commands.RegisterComponent("snow_proxy_buy_add_asia", makeProxyBuyAddHandler(ServerAsia))
	commands.RegisterComponent("snow_proxy_buy_add_twhkmo", makeProxyBuyAddHandler(ServerTWHKMO))
	commands.RegisterComponent("snow_proxy_buy_search_asia", makeProxyBuySearchHandler(ServerAsia))
	commands.RegisterComponent("snow_proxy_buy_search_twhkmo", makeProxyBuySearchHandler(ServerTWHKMO))

	// Proxy sell sub-buttons (with server region)
	commands.RegisterComponent("snow_proxy_sell_add_asia", makeProxySellAddHandler(ServerAsia))
	commands.RegisterComponent("snow_proxy_sell_add_twhkmo", makeProxySellAddHandler(ServerTWHKMO))
	commands.RegisterComponent("snow_proxy_sell_search_asia", makeProxySellSearchHandler(ServerAsia))
	commands.RegisterComponent("snow_proxy_sell_search_twhkmo", makeProxySellSearchHandler(ServerTWHKMO))
	commands.RegisterComponent("snow_proxy_sell_list_asia", makeProxySellListHandler(ServerAsia))
	commands.RegisterComponent("snow_proxy_sell_list_twhkmo", makeProxySellListHandler(ServerTWHKMO))

	// Pagination handlers (with server region, pages 0-9)
	for i := 0; i < 10; i++ {
		commands.RegisterComponent(fmt.Sprintf("snow_sell_page_%d_asia", i), makeProxySellPageHandlerWithServer(i, ServerAsia))
		commands.RegisterComponent(fmt.Sprintf("snow_sell_page_%d_twhkmo", i), makeProxySellPageHandlerWithServer(i, ServerTWHKMO))
	}

	// Select menu handlers (with server region)
	commands.RegisterComponent("snow_proxy_sell_select_asia", makeProxySellSelectHandler(ServerAsia))
	commands.RegisterComponent("snow_proxy_sell_select_twhkmo", makeProxySellSelectHandler(ServerTWHKMO))
	commands.RegisterComponent("snow_proxy_sell_search_select_asia", makeProxySellSearchSelectHandler(ServerAsia))
	commands.RegisterComponent("snow_proxy_sell_search_select_twhkmo", makeProxySellSearchSelectHandler(ServerTWHKMO))

	// Modal handlers for proxy buy search
	commands.RegisterModal("snow_proxy_buy_search_modal_asia", makeProxyBuySearchModalHandler(ServerAsia))
	commands.RegisterModal("snow_proxy_buy_search_modal_twhkmo", makeProxyBuySearchModalHandler(ServerTWHKMO))

	// Back buttons
	commands.RegisterComponent("snow_back_main", handleBackToMain)
	commands.RegisterComponent("snow_back_server_asia", makeBackToServerActionMenuHandler(ServerAsia))
	commands.RegisterComponent("snow_back_server_twhkmo", makeBackToServerActionMenuHandler(ServerTWHKMO))
	commands.RegisterComponent("snow_back_buy_asia", makeBackToBuyHandler(ServerAsia))
	commands.RegisterComponent("snow_back_buy_twhkmo", makeBackToBuyHandler(ServerTWHKMO))
	commands.RegisterComponent("snow_back_sell_asia", makeBackToSellHandler(ServerAsia))
	commands.RegisterComponent("snow_back_sell_twhkmo", makeBackToSellHandler(ServerTWHKMO))

	// Proxy buy column select handlers
	commands.RegisterComponent("snow_proxy_buy_col_select_asia", makeProxyBuyColumnSelectHandler(ServerAsia))
	commands.RegisterComponent("snow_proxy_buy_col_select_twhkmo", makeProxyBuyColumnSelectHandler(ServerTWHKMO))

	// Proxy buy row select handlers (multi-select dropdown, auto-saves on select)
	// Format: snow_proxy_buy_row_select_{server}_{column}
	for _, col := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M"} {
		commands.RegisterComponent(fmt.Sprintf("snow_proxy_buy_row_select_%s_%s", ServerAsia, col), makeProxyBuyRowSelectHandler(ServerAsia, col))
		commands.RegisterComponent(fmt.Sprintf("snow_proxy_buy_row_select_%s_%s", ServerTWHKMO, col), makeProxyBuyRowSelectHandler(ServerTWHKMO, col))
	}
}

var snowSeasonCommand = &discordgo.ApplicationCommand{
	Name:        "冰雪季",
	Description: "冰雪季相關資訊與功能",
}

// Link constants
const (
	snowBubbleLink     = "https://discord.com/channels/1438429535975641120/1467392127712628900/1467392390045372616"
	menuPriceLink      = "https://discord.com/channels/1438429535975641120/1452258561156710462/1462660483457876077"
	proxyBuyItemsImage = "https://duk.tw/vdkUPU.jpg"
)

// Snow theme color (ice blue)
const ColorSnow = 0x87CEEB

// Pagination constants
const proxySellPageSize = 10

// ============================================
// Proxy sell items (ordered by ID)
// ID 3 is reserved but not shown in menu
// ============================================

type proxySellItem struct {
	ID   int
	Name string
}

var proxySellItems = []proxySellItem{
	{1, "冰晶帝王蟹"},
	{2, "冰晶翻車魚"},
	// ID 3 is reserved (not yet released)
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

// proxySellEntry represents a database entry
type proxySellEntry struct {
	DiscordID string
	Items     string
}

// getItemNameByID returns item name by ID
func getItemNameByID(id int) string {
	for _, item := range proxySellItems {
		if item.ID == id {
			return item.Name
		}
	}
	return ""
}

// parseItemIDs parses comma-separated item IDs string to slice
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

// itemIDsToNames converts item IDs to names (sorted by ID)
func itemIDsToNames(ids []int) []string {
	// Sort IDs first for consistent display order
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

// containsItemID checks if the item ID is in the comma-separated string
func containsItemID(itemsStr string, targetID int) bool {
	ids := parseItemIDs(itemsStr)
	for _, id := range ids {
		if id == targetID {
			return true
		}
	}
	return false
}

// getLastSaturdayReset returns the last Saturday 6:00 AM Taiwan time (UTC+8)
// If current time is before Saturday 6:00 AM, returns last week's Saturday
func getLastSaturdayReset() time.Time {
	// Taiwan timezone (UTC+8)
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		// Fallback to fixed offset if timezone not found
		loc = time.FixedZone("Asia/Taipei", 8*60*60)
	}

	now := time.Now().In(loc)

	// Find the most recent Saturday
	daysSinceSaturday := int(now.Weekday()+1) % 7 // Saturday = 6, so (weekday + 1) % 7 gives days since Saturday

	// Get last Saturday at 6:00 AM
	lastSaturday := now.AddDate(0, 0, -daysSinceSaturday)
	resetTime := time.Date(lastSaturday.Year(), lastSaturday.Month(), lastSaturday.Day(), 6, 0, 0, 0, loc)

	// If we haven't reached this Saturday's 6 AM yet, go back another week
	if now.Before(resetTime) {
		resetTime = resetTime.AddDate(0, 0, -7)
	}

	return resetTime.UTC()
}

// getServerDisplayName returns the display name for a server region
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

// getProxySellEntries fetches all proxy sell entries since last reset for a specific server
func getProxySellEntries(serverRegion string) ([]proxySellEntry, error) {
	resetTime := getLastSaturdayReset()

	rows, err := database.DB.Query(`
		SELECT discord_id, items
		FROM snow_proxy_sell
		WHERE created_at >= ? AND server_region = ?
		ORDER BY created_at DESC
	`, resetTime, serverRegion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []proxySellEntry
	for rows.Next() {
		var entry proxySellEntry
		if err := rows.Scan(&entry.DiscordID, &entry.Items); err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// ============================================
// Embed builders
// ============================================

func buildMainEmbed() *discordgo.MessageEmbed {
	return embed.New().
		Title("❄️ 冰雪季功能選單").
		Description("冰雪季懶人包 ➡️ https://discord.com/channels/1438429535975641120/1438749728446873654/1466829733957341185\n 冰雪市場使用方式 ➡️ https://discord.com/channels/1438429535975641120/1467541407378964716/1467551966996467956").
		Color(ColorSnow).
		// FooterText("\n⚠️目前線上開發中，若交互失敗5秒後再點擊一次。").
		Build()
}

func buildSnowBubbleEmbed() *discordgo.MessageEmbed {
	return embed.New().
		Title("⛄ 雪人泡泡位置").
		Description(fmt.Sprintf("在這個頻道 ⬇️\n%s", snowBubbleLink)).
		Color(ColorSnow).
		Build()
}

// Server action menu embed (buy/sell selection after choosing server)
func buildServerActionMenuEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("❄️ 冰雪季【%s】", getServerDisplayName(server))).
		Description("⚠️代購功能較為繁瑣為DC機制導致，若有不理解都歡迎於頻道詢問。\n\n因為操作有一定複雜度，不一定要使用bot的代購\n也可於頻道直接詢問有誰有什麼物品。\n————————————————————————————\n\n⚠️代售功能就較為簡單，歡迎大家多多使用。\n————————————————————————————\n\n").
		FooterText("辛苦大家腦力激盪一下橫排數字記憶🙇🏻‍♀️").
		Color(ColorSnow).
		Build()
}

func buildProxyBuyEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("🛒 代購功能【%s】", getServerDisplayName(server))).
		Description("這裡可以**登記你能幫別人代購**的物品，或是搜尋有誰可以幫你代購。").
		Color(ColorSnow).
		Build()
}

func buildProxySellEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("💰 代售功能【%s】", getServerDisplayName(server))).
		Description("你可以登記你本週可以代售的物品，或是搜尋有誰可以幫你代售。\n*登記資料會在**每週六早上 6 點清空**，因應遊戲內每週更新一次。*").
		Color(ColorSnow).
		FooterText("登記代表你接受在discord被標記，會有人找你代售哦🌺。\n另外請鄰居們記得餵食幫助你代售的發展家。 ").
		Build()
}

func buildProxySellSelectEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("💰 登記代售物品【%s】", getServerDisplayName(server))).
		Description("請選擇你本週可以代售的物品：").
		Color(ColorSnow).
		Build()
}

func buildProxySellSearchEmbed(server string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("🔍 找可代售物品【%s】", getServerDisplayName(server))).
		Description("請選擇你想找的物品：").
		Color(ColorSnow).
		Build()
}

// ============================================
// Main command handler
// ============================================

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
// Button builders
// ============================================

func buildMainButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_bubble").Label("⛄ 雪人泡泡位置").Secondary().Build()).
			AddButton(component.NewButton().CustomID("snow_server_twhkmo").Label("🇹🇼 代購代售【台港澳】").Primary().Build()).
			AddButton(component.NewButton().CustomID("snow_server_asia").Label("🌏 代購代售【亞服】").Primary().Build()).
			Build(),
	}
}

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

// Server action menu buttons (buy/sell selection)
func buildServerActionMenuButtons(server string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_action_buy_" + server).Label("🛒 代購功能").Primary().Build()).
			AddButton(component.NewButton().CustomID("snow_action_sell_" + server).Label("💰 代售功能").Primary().Build()).
			AddButton(component.NewButton().CustomID("snow_back_main").Label("⬅️ 返回").Secondary().Build()).
			Build(),
	}
}

func buildProxyBuyButtons(server string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_proxy_buy_add_"+server).Label("📝 登記可幫代購物品").Success().Build()).
			AddButton(component.NewButton().CustomID("snow_proxy_buy_search_"+server).Label("🔍 找代購物品").Primary().Build()).
			AddButton(component.NewButton().CustomID("snow_back_server_"+server).Label("⬅️ 返回").Secondary().Build()).
			Build(),
	}
}

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

func buildProxySellSelectMenu(server string) []discordgo.MessageComponent {
	selectMenu := component.NewSelect().
		CustomID("snow_proxy_sell_select_" + server).
		Placeholder("選擇你本週可以代售的物品（多選，重複登記=覆蓋）").
		MinValues(1).
		MaxValues(len(proxySellItems))

	for _, item := range proxySellItems {
		selectMenu.AddOption(item.Name, fmt.Sprintf("%d", item.ID), "")
	}

	return []discordgo.MessageComponent{
		component.NewActionRow().AddSelect(selectMenu.Build()).Build(),
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_back_sell_" + server).Label("⬅️ 返回").Secondary().Build()).
			Build(),
	}
}

func buildProxySellSearchSelectMenu(server string) []discordgo.MessageComponent {
	selectMenu := component.NewSelect().
		CustomID("snow_proxy_sell_search_select_" + server).
		Placeholder("選擇你想找的物品").
		MinValues(1).
		MaxValues(1)

	for _, item := range proxySellItems {
		selectMenu.AddOption(item.Name, fmt.Sprintf("%d", item.ID), "")
	}

	return []discordgo.MessageComponent{
		component.NewActionRow().AddSelect(selectMenu.Build()).Build(),
		component.NewActionRow().
			AddButton(component.NewButton().CustomID("snow_back_sell_" + server).Label("⬅️ 返回").Secondary().Build()).
			Build(),
	}
}

func buildProxySellListButtons(page, totalPages int, server string) []discordgo.MessageComponent {
	row := component.NewActionRow()

	// Previous button
	if page > 0 {
		row.AddButton(component.NewButton().
			CustomID(fmt.Sprintf("snow_sell_page_%d_%s", page-1, server)).
			Label("◀️ 上一頁").
			Secondary().
			Build())
	}

	// Next button
	if page < totalPages-1 {
		row.AddButton(component.NewButton().
			CustomID(fmt.Sprintf("snow_sell_page_%d_%s", page+1, server)).
			Label("下一頁 ▶️").
			Secondary().
			Build())
	}

	// Back button
	row.AddButton(component.NewButton().
		CustomID("snow_back_sell_" + server).
		Label("⬅️ 返回").
		Secondary().
		Build())

	return []discordgo.MessageComponent{row.Build()}
}

// ============================================
// Button handlers
// ============================================

func handleSnowBubble(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildSnowBubbleEmbed()},
			Components: buildOtherButtons("snow_bubble"),
		},
	})
}

// Server action menu handler factory (main menu -> server action menu)
func makeServerActionMenuHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{buildServerActionMenuEmbed(server)},
				Components: buildServerActionMenuButtons(server),
			},
		})
	}
}

// Action handler factory (server action menu -> buy/sell menu)
func makeActionHandler(server, action string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if action == "buy" {
			showProxyBuyMenu(s, i, server)
		} else {
			showProxySellMenu(s, i, server)
		}
	}
}

func showProxyBuyMenu(s *discordgo.Session, i *discordgo.InteractionCreate, server string) {
	e := embed.New().
		Title(fmt.Sprintf("🛒 代購功能【%s】", getServerDisplayName(server))).
		Description("這裡可以登記你能幫別人代購的物品，或是搜尋有誰可以幫你代購。\n從左方直列英文開始再來是橫向數字，例如傢俱床就是`M9`").
		Image(proxyBuyItemsImage).
		FooterText("圖源來自小紅書@露娜，已取得引用同意。").
		Color(ColorSnow).
		Build()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{e},
			Components:  buildProxyBuyButtons(server),
			Attachments: &[]*discordgo.MessageAttachment{},
		},
	})
}

func showProxySellMenu(s *discordgo.Session, i *discordgo.InteractionCreate, server string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{buildProxySellEmbed(server)},
			Components: buildProxySellButtons(server),
		},
	})
}

func handleBackToMain(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{buildMainEmbed()},
			Components:  buildMainButtons(),
			Attachments: &[]*discordgo.MessageAttachment{}, // Clear attachments
		},
	})
}

// Back to server action menu handler factory
func makeBackToServerActionMenuHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:      []*discordgo.MessageEmbed{buildServerActionMenuEmbed(server)},
				Components:  buildServerActionMenuButtons(server),
				Attachments: &[]*discordgo.MessageAttachment{}, // Clear attachments
			},
		})
	}
}

func makeBackToBuyHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		showProxyBuyMenu(s, i, server)
	}
}

func makeBackToSellHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		showProxySellMenu(s, i, server)
	}
}

// ============================================
// Proxy buy handlers
// ============================================

func makeProxyBuyAddHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
}

func makeProxyBuyColumnSelectHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.MessageComponentData()
		if len(data.Values) == 0 {
			respondWithError(s, i, "請選擇一個直排")
			return
		}
		selectedColumn := data.Values[0]
		userID := getUserID(i)

		// Get user's current selections for this column
		selectedRows := getUserSelectedRowsForColumn(userID, server, selectedColumn)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{buildProxyBuySelectEmbed(server, selectedColumn)},
				Components: buildProxyBuySelectMenus(server, selectedColumn, selectedRows),
			},
		})
	}
}

// getUserSelectedRowsForColumn returns which rows the user has selected for a specific column
func getUserSelectedRowsForColumn(userID, server, column string) map[int]bool {
	selectedRows := make(map[int]bool)

	var existingItems string
	err := database.DB.QueryRow(`
		SELECT items FROM snow_proxy_buy
		WHERE discord_id = ? AND server_region = ? AND created_at >= ?
	`, userID, server, getLastSaturdayReset()).Scan(&existingItems)

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

func buildProxyBuySelectEmbed(server, column string) *discordgo.MessageEmbed {
	return embed.New().
		Title(fmt.Sprintf("📝 登記代購物品【%s】", getServerDisplayName(server))).
		Description(fmt.Sprintf("1️⃣ 選擇直排（A-M）\n2️⃣ 選擇橫排（1-10，**可多選**，選完自動儲存）\n\n目前直排：**%s**", column)).
		Image(proxyBuyItemsImage).
		FooterText("圖源來自小紅書@露娜，已取得引用同意。").
		Color(ColorSnow).
		Build()
}

func buildProxyBuySelectMenus(server, column string, selectedRows map[int]bool) []discordgo.MessageComponent {
	// Row 1: Column select dropdown (A-M) - single select
	colSelect := component.NewSelect().
		CustomID("snow_proxy_buy_col_select_" + server).
		Placeholder("選擇直排 A-M").
		MinValues(1).
		MaxValues(1)

	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M"}
	for _, col := range columns {
		if col == column {
			colSelect.AddOptionDefault(col, col, "")
		} else {
			colSelect.AddOption(col, col, "")
		}
	}

	// Row 2: Row select dropdown (1-10) - multi select with Default for toggle behavior
	rowSelect := component.NewSelect().
		CustomID(fmt.Sprintf("snow_proxy_buy_row_select_%s_%s", server, column)).
		Placeholder("選擇橫排 1-10（可多選，點擊切換）").
		MinValues(0).
		MaxValues(10)

	for r := 1; r <= 10; r++ {
		label := fmt.Sprintf("%d", r)
		value := fmt.Sprintf("%d", r)
		if selectedRows[r] {
			// Use AddOptionDefault so Discord shows it as selected natively
			rowSelect.AddOptionDefault(label, value, "")
		} else {
			rowSelect.AddOption(label, value, "")
		}
	}

	// Row 3: Back button only (save happens automatically when selecting rows)
	row3 := component.NewActionRow().
		AddButton(component.NewButton().CustomID("snow_back_server_" + server).Label("⬅️ 返回選單").Secondary().Build())

	return []discordgo.MessageComponent{
		component.NewActionRow().AddSelect(colSelect.Build()).Build(),
		component.NewActionRow().AddSelect(rowSelect.Build()).Build(),
		row3.Build(),
	}
}

// makeProxyBuyRowSelectHandler handles row select menu (multi-select 1-10)
// This saves the selection automatically when user closes the dropdown
func makeProxyBuyRowSelectHandler(server, column string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.MessageComponentData()
		userID := getUserID(i)
		if userID == "" {
			respondWithError(s, i, "無法取得用戶資訊")
			return
		}

		// Build selected rows map from selection
		selectedRows := make(map[int]bool)
		for _, v := range data.Values {
			var row int
			fmt.Sscanf(v, "%d", &row)
			selectedRows[row] = true
		}

		// Get existing items for all columns (not just this column)
		var existingItems string
		err := database.DB.QueryRow(`
			SELECT items FROM snow_proxy_buy
			WHERE discord_id = ? AND server_region = ? AND created_at >= ?
		`, userID, server, getLastSaturdayReset()).Scan(&existingItems)

		itemSet := make(map[string]bool)
		if err == nil && existingItems != "" {
			for _, item := range strings.Split(existingItems, ",") {
				item = strings.TrimSpace(item)
				if item != "" {
					// Keep items from other columns
					if len(item) >= 1 && !strings.HasPrefix(item, column) {
						itemSet[item] = true
					}
				}
			}
		}

		// Add current column's selections
		for row := 1; row <= 10; row++ {
			if selectedRows[row] {
				itemSet[fmt.Sprintf("%s%d", column, row)] = true
			}
		}

		// Convert to sorted string
		var allItems []string
		for item := range itemSet {
			allItems = append(allItems, item)
		}
		sort.Strings(allItems)
		finalItems := strings.Join(allItems, ",")

		// Save to database (always upsert, never delete to avoid ID gaps)
		_, err = database.DB.Exec(`
			INSERT INTO snow_proxy_buy (discord_id, server_region, items)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE items = ?, created_at = NOW()
		`, userID, server, finalItems, finalItems)
		if err != nil {
			log.Printf("Failed to save proxy buy: %v", err)
		}

		// Update UI
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{buildProxyBuySelectEmbed(server, column)},
				Components: buildProxyBuySelectMenus(server, column, selectedRows),
			},
		})
	}
}

func makeProxyBuySearchHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Show modal for user to input the coordinate they want to search
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
}

func makeProxyBuySearchModalHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.ModalSubmitData()
		coordinate := strings.ToUpper(strings.TrimSpace(data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value))

		// Validate coordinate format (A-M followed by 1-10)
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

		// Search for users who have this coordinate registered
		rows, err := database.DB.Query(`
			SELECT discord_id, items
			FROM snow_proxy_buy
			WHERE server_region = ? AND created_at >= ?
		`, server, getLastSaturdayReset())
		if err != nil {
			log.Printf("Failed to query proxy buy: %v", err)
			respondWithError(s, i, "查詢失敗，請稍後再試")
			return
		}
		defer rows.Close()

		var result string
		count := 0
		for rows.Next() {
			var discordID, items string
			if err := rows.Scan(&discordID, &items); err != nil {
				continue
			}
			// Check if this user has the searched coordinate
			for _, item := range strings.Split(items, ",") {
				if strings.TrimSpace(item) == coordinate {
					result += fmt.Sprintf("• <@%s>\n", discordID)
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
}

// ============================================
// Proxy sell handlers
// ============================================

func makeProxySellAddHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{buildProxySellSelectEmbed(server)},
				Components: buildProxySellSelectMenu(server),
			},
		})
	}
}

func makeProxySellSelectHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.MessageComponentData()
		selectedValues := data.Values

		userID := getUserID(i)
		if userID == "" {
			respondWithError(s, i, "無法取得用戶資訊")
			return
		}

		// Convert selected values to item IDs string (comma-separated)
		itemIDs := strings.Join(selectedValues, ",")

		// Convert to item names for display (sorted by ID)
		var intIDs []int
		for _, v := range selectedValues {
			var id int
			fmt.Sscanf(v, "%d", &id)
			intIDs = append(intIDs, id)
		}
		selectedNames := itemIDsToNames(intIDs)

		// Upsert into database with server_region
		_, err := database.DB.Exec(`
			INSERT INTO snow_proxy_sell (discord_id, server_region, items)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE items = ?, created_at = NOW()
		`, userID, server, itemIDs, itemIDs)
		if err != nil {
			log.Printf("Failed to save proxy sell: %v", err)
			respondWithError(s, i, "儲存失敗，請稍後再試")
			return
		}

		e := embed.New().
			Title("✅ 登記成功").
			Description(fmt.Sprintf("【%s】已登記你可以代售的物品：\n**%s**\n\n其他人現在可以找到你了！", getServerDisplayName(server), strings.Join(selectedNames, "、"))).
			Color(embed.ColorSuccess).
			Build()

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{e},
				Components: buildProxySellButtons(server),
			},
		})
	}
}

func makeProxySellSearchHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Show select menu to choose which item to search
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{buildProxySellSearchEmbed(server)},
				Components: buildProxySellSearchSelectMenu(server),
			},
		})
	}
}

func makeProxySellSearchSelectHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.MessageComponentData()
		if len(data.Values) == 0 {
			respondWithError(s, i, "請選擇一個物品")
			return
		}

		var searchItemID int
		fmt.Sscanf(data.Values[0], "%d", &searchItemID)
		searchItemName := getItemNameByID(searchItemID)

		entries, err := getProxySellEntries(server)
		if err != nil {
			log.Printf("Failed to query proxy sell: %v", err)
			respondWithError(s, i, "查詢失敗，請稍後再試")
			return
		}

		var result string
		count := 0
		for _, entry := range entries {
			// Check if this user has the searched item
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
}

func makeProxySellListHandler(server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		showProxySellListPageWithServer(s, i, 0, server)
	}
}

func makeProxySellPageHandlerWithServer(page int, server string) commands.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		showProxySellListPageWithServer(s, i, page, server)
	}
}

func showProxySellListPageWithServer(s *discordgo.Session, i *discordgo.InteractionCreate, page int, server string) {
	entries, err := getProxySellEntries(server)
	if err != nil {
		log.Printf("Failed to query proxy sell: %v", err)
		respondWithError(s, i, "查詢失敗，請稍後再試")
		return
	}

	totalCount := len(entries)
	totalPages := (totalCount + proxySellPageSize - 1) / proxySellPageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// Clamp page number
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
		// Get entries for current page
		start := page * proxySellPageSize
		end := start + proxySellPageSize
		if end > totalCount {
			end = totalCount
		}

		var result string
		for idx, entry := range entries[start:end] {
			// Convert item IDs to names
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

// ============================================
// Helper functions
// ============================================

func getUserID(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	e := embed.New().
		Title("❌ 錯誤").
		Description(message).
		Color(embed.ColorError).
		Build()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:      []*discordgo.MessageEmbed{e},
			Components:  buildMainButtons(),
			Attachments: &[]*discordgo.MessageAttachment{}, // Clear attachments
		},
	})
}

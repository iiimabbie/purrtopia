package group

import (
	"fmt"
	"strings"

	"purrtopia/internal/database"

	"github.com/bwmarrin/discordgo"
)

// Config keys（對應 bot_config 表的 key）
const (
	ConfigGroupAllowedChannels = "group_allowed_channels"
	ConfigToolboxRoleID        = "toolbox_role_id"
	ConfigRainbowRoleID        = "rainbow_role_id"
	ConfigFishingRoleID        = "fishing_role_id"
)

// ActivityType 揪團類型
type ActivityType string

const (
	ActivityRainbow ActivityType = "rainbow"
	ActivityFishing ActivityType = "fishing"
)

// activityNames 各類型顯示名稱
var activityNames = map[ActivityType]string{
	ActivityRainbow: "彩虹花束",
	ActivityFishing: "出海",
}

// activityRoleKeys 各類型對應的 bot_config key
var activityRoleKeys = map[ActivityType]string{
	ActivityRainbow: ConfigRainbowRoleID,
	ActivityFishing: ConfigFishingRoleID,
}

// activityColors 各類型主題色
var activityColors = map[ActivityType]int{
	ActivityRainbow: 0xFF69B4, // 粉紅
	ActivityFishing: 0x00BFFF, // 天藍
}

// activityEmojis 各類型 emoji
var activityEmojis = map[ActivityType]string{
	ActivityRainbow: "🌈",
	ActivityFishing: "🎣",
}

// activityAnnouncements 各類型發布時的公告文字
var activityAnnouncements = map[ActivityType]string{
	ActivityRainbow: "灑彩虹花束的時間到囉🌸✨",
	ActivityFishing: "來吃魚🪝🐟",
}

func (a ActivityType) Name() string         { return activityNames[a] }
func (a ActivityType) RoleID() string       { return database.BotConfigRepo.Get(activityRoleKeys[a]) }
func (a ActivityType) Color() int           { return activityColors[a] }
func (a ActivityType) Emoji() string        { return activityEmojis[a] }
func (a ActivityType) Label() string        { return a.Emoji() + " " + a.Name() }
func (a ActivityType) Announcement() string { return activityAnnouncements[a] }

// memberCount 計算報名人數
func memberCount(members string) int {
	if members == "" {
		return 0
	}
	count := 0
	for _, m := range strings.Split(members, ",") {
		if strings.TrimSpace(m) != "" {
			count++
		}
	}
	return count
}

// addMember 加入報名者 ID
func addMember(members, discordID string) string {
	discordID = strings.TrimSpace(discordID)
	if members == "" {
		return discordID
	}
	return members + "," + discordID
}

// hasMember 檢查是否已報名
func hasMember(members, discordID string) bool {
	if members == "" {
		return false
	}
	for _, m := range strings.Split(members, ",") {
		if strings.TrimSpace(m) == discordID {
			return true
		}
	}
	return false
}

// removeMember 從成員列表移除指定 Discord ID
func removeMember(members, discordID string) string {
	if members == "" {
		return ""
	}
	var kept []string
	for _, m := range strings.Split(members, ",") {
		m = strings.TrimSpace(m)
		if m != "" && m != discordID {
			kept = append(kept, m)
		}
	}
	return strings.Join(kept, ",")
}

// extractActivityID 從 customID 取出活動 ID
func extractActivityID(customID, prefix string) int {
	trimmed := strings.TrimPrefix(customID, prefix)
	var id int
	fmt.Sscanf(trimmed, "%d", &id)
	return id
}

// getDisplayName 取得互動者的顯示名稱（暱稱 > 全域顯示名稱 > 帳號名稱）
func getDisplayName(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		if i.Member.Nick != "" {
			return i.Member.Nick
		}
		if i.Member.User != nil {
			if i.Member.User.GlobalName != "" {
				return i.Member.User.GlobalName
			}
			return i.Member.User.Username
		}
	}
	if i.User != nil {
		if i.User.GlobalName != "" {
			return i.User.GlobalName
		}
		return i.User.Username
	}
	return "某人"
}

// getUserID 取得互動的 Discord 使用者 ID
func getUserID(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

// channelCheck 檢查是否在允許的頻道，否則回 ephemeral 擋住
func channelCheck(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	allowed := database.BotConfigRepo.GetMap(ConfigGroupAllowedChannels)
	if _, ok := allowed[i.ChannelID]; !ok {
		respondWithEphemeral(s, i, "此頻道不支援揪團功能哦！")
		return false
	}
	return true
}

// respondWithEphemeral 回覆私人可見訊息
func respondWithEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// errorMsg 統一錯誤訊息內容
func errorMsg() string {
	roleID := database.BotConfigRepo.Get(ConfigToolboxRoleID)
	if roleID == "" {
		return "對不起，我好像有點錯誤 🥺"
	}
	return fmt.Sprintf("對不起，我好像有點錯誤，可以幫我 <@&%s> 嗎? 🥺", roleID)
}

// respondWithError 回覆統一友善錯誤訊息（用於尚未 acknowledge 的 interaction）
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondWithEphemeral(s, i, errorMsg())
}

// editWithError 編輯已 deferred 的 interaction 為統一錯誤訊息
func editWithError(s *discordgo.Session, i *discordgo.InteractionCreate) {
	msg := errorMsg()
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
}

// sendFollowupEphemeral 在已 acknowledge 的 interaction 後補一則只有按鈕者看到的訊息
func sendFollowupEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

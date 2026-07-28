package bot

import (
	"log"

	"purrtopia/internal/database"

	"github.com/bwmarrin/discordgo"
)

// Config keys（對應 bot_config 表的 key）
const (
	ConfigReactionRoleMessageID = "reaction_role_message_id"
	ConfigReactionRoleIDs       = "reaction_role_ids"
)

// onReactionAdd 監聽反應新增事件：🛠️ → 發放所有身份組
func (b *Bot) onReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	messageID := database.BotConfigRepo.Get(ConfigReactionRoleMessageID)
	if messageID == "" || r.MessageID != messageID {
		return
	}
	if s.State.User != nil && r.UserID == s.State.User.ID {
		return
	}
	log.Printf("[reactionrole] 收到反應 msgID=%s userID=%s emoji.Name=%q emoji.ID=%q", r.MessageID, r.UserID, r.Emoji.Name, r.Emoji.ID)

	if r.Emoji.Name != "🛠️" {
		return
	}

	roleIDs := database.BotConfigRepo.GetList(ConfigReactionRoleIDs)
	for _, roleID := range roleIDs {
		if err := s.GuildMemberRoleAdd(r.GuildID, r.UserID, roleID); err != nil {
			log.Printf("[reactionrole] 新增身份組失敗 userID=%s roleID=%s: %v", r.UserID, roleID, err)
		}
	}
	log.Printf("[reactionrole] 發放身份組 userID=%s", r.UserID)
}

// onReactionRemove 監聯反應移除事件：取消 🛠️ → 移除所有身份組
func (b *Bot) onReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	messageID := database.BotConfigRepo.Get(ConfigReactionRoleMessageID)
	if messageID == "" || r.MessageID != messageID {
		return
	}
	if s.State.User != nil && r.UserID == s.State.User.ID {
		return
	}
	if r.Emoji.Name != "🛠️" {
		return
	}

	roleIDs := database.BotConfigRepo.GetList(ConfigReactionRoleIDs)
	for _, roleID := range roleIDs {
		if err := s.GuildMemberRoleRemove(r.GuildID, r.UserID, roleID); err != nil {
			log.Printf("[reactionrole] 移除身份組失敗 userID=%s roleID=%s: %v", r.UserID, roleID, err)
		}
	}
	log.Printf("[reactionrole] 移除身份組 userID=%s", r.UserID)
}

package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// reactionRoleMessageID 監看的訊息 ID
const reactionRoleMessageID = "1476903139050520668"

// reactionRoleIDs 要發放／移除的身份組 ID 清單
var reactionRoleIDs = []string{
	"1476601339151650826", // 出海
	"1476601591774839007", // 灑花
}

// onReactionAdd 監聽反應新增事件：🛠️ → 發放所有身份組
func (b *Bot) onReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.MessageID != reactionRoleMessageID {
		return
	}
	if s.State.User != nil && r.UserID == s.State.User.ID {
		return
	}
	log.Printf("[reactionrole] 收到反應 msgID=%s userID=%s emoji.Name=%q emoji.ID=%q", r.MessageID, r.UserID, r.Emoji.Name, r.Emoji.ID)

	if r.Emoji.Name != "🛠️" {
		return
	}

	for _, roleID := range reactionRoleIDs {
		if err := s.GuildMemberRoleAdd(r.GuildID, r.UserID, roleID); err != nil {
			log.Printf("[reactionrole] 新增身份組失敗 userID=%s roleID=%s: %v", r.UserID, roleID, err)
		}
	}
	log.Printf("[reactionrole] 發放身份組 userID=%s", r.UserID)
}

// onReactionRemove 監聽反應移除事件：取消 🛠️ → 移除所有身份組
func (b *Bot) onReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	if r.MessageID != reactionRoleMessageID {
		return
	}
	if s.State.User != nil && r.UserID == s.State.User.ID {
		return
	}
	if r.Emoji.Name != "🛠️" {
		return
	}

	for _, roleID := range reactionRoleIDs {
		if err := s.GuildMemberRoleRemove(r.GuildID, r.UserID, roleID); err != nil {
			log.Printf("[reactionrole] 移除身份組失敗 userID=%s roleID=%s: %v", r.UserID, roleID, err)
		}
	}
	log.Printf("[reactionrole] 移除身份組 userID=%s", r.UserID)
}

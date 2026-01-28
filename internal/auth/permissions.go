package auth

import (
	"github.com/bwmarrin/discordgo"
	"purrtopia/internal/config"
)

// Permission 定義權限等級
type Permission int

const (
	PermissionNone Permission = iota
	PermissionServerAdmin
	PermissionBotAdmin
	PermissionBotOwner
)

var cfg *config.Config

// Init 初始化權限模組（需在 main 中呼叫）
func Init(c *config.Config) {
	cfg = c
}

// isBotOwner 檢查用戶是否為 Bot 擁有者
func isBotOwner(userID string) bool {
	if cfg == nil {
		return false
	}
	for _, ownerID := range cfg.OwnerIDs {
		if ownerID == userID {
			return true
		}
	}
	return false
}

// isBotAdmin 檢查用戶是否為 Bot 管理員
func isBotAdmin(userID string) bool {
	if cfg == nil {
		return false
	}
	for _, adminID := range cfg.AdminIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

// isServerAdmin 檢查用戶是否為伺服器管理員（Discord Administrator 權限）
func isServerAdmin(s *discordgo.Session, guildID, userID string) bool {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return false
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return false
	}

	// 伺服器擁有者自動擁有管理員權限
	if guild.OwnerID == userID {
		return true
	}

	// 檢查成員的角色是否有 Administrator 權限
	for _, roleID := range member.Roles {
		for _, role := range guild.Roles {
			if role.ID == roleID && role.Permissions&discordgo.PermissionAdministrator != 0 {
				return true
			}
		}
	}
	return false
}

// CheckPermission 檢查用戶的最高權限等級
func CheckPermission(s *discordgo.Session, guildID, userID string) Permission {
	if isBotOwner(userID) {
		return PermissionBotOwner
	}
	if isBotAdmin(userID) {
		return PermissionBotAdmin
	}
	if isServerAdmin(s, guildID, userID) {
		return PermissionServerAdmin
	}
	return PermissionNone
}

// HasPermission 檢查用戶是否有指定的權限等級（或更高）
func HasPermission(s *discordgo.Session, guildID, userID string, required Permission) bool {
	return CheckPermission(s, guildID, userID) >= required
}

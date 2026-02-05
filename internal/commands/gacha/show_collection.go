package gacha

import (
	"fmt"
	"log"
	"strings"

	"purrtopia/internal/commands"
	"purrtopia/internal/database"
	"purrtopia/internal/embed"

	"github.com/bwmarrin/discordgo"
)

// func init() {
// 	commands.RegisterCommand(showCollectionCommand, ShowCollectionHandler)
// }

var showCollectionCommand = &discordgo.ApplicationCommand{
	Name:        "展示",
	Description: "展示你抽過的所有頭顱",
}

// /展示 command
func ShowCollectionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string
	if i.Member != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	if userID == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "無法取得用戶資訊",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// 取得用戶抽過的頭
	emojis, err := getDrawnEmojis(userID)
	if err != nil {
		log.Printf("Failed to get drawn emojis: %v", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "查詢時發生錯誤，請稍後再試",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if len(emojis) == 0 {
		e := embed.New().
			Title("🎭 頭顱收藏").
			Description("你還沒有抽過任何頭顱！\n快去使用 `/抽頭` 收集頭顱吧～").
			Color(embed.ColorBlurple).
			Build()

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{e},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// response 含 emoji
	emojiStr := strings.Join(emojis, "")

	e := embed.New().
		Title("🎭 頭顱收藏").
		Description(fmt.Sprintf("噹啷啷～! 這裡是你抽過的頭唷:\n\n%s", emojiStr)).
		Color(embed.ColorBlurple).
		Footer(fmt.Sprintf("共收集了 %d 顆頭顱", len(emojis)), "").
		Timestamp().
		Build()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{e},
		},
	})
}

// 取抽過的頭
func getDrawnEmojis(discordID string) ([]string, error) {
	// 使用 GORM Repository
	rows, err := database.DrawnHeadRepo.FindDrawnEmojisByDiscordID(discordID)
	if err != nil {
		return nil, err
	}

	var emojis []string
	for _, row := range rows {
		if row.EmojiID != "" && row.EmojiName != "" {
			emojis = append(emojis, fmt.Sprintf("<:%s:%s>", row.EmojiName, row.EmojiID))
		}
	}

	return emojis, nil
}

// Ensure commands package is imported (for future init registration)
var _ = commands.RegisterCommand

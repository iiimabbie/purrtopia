package commands

import (
	"database/sql"
	"fmt"
	"log"

	"discord-bot-template/internal/database"
	"discord-bot-template/internal/embed"

	"github.com/bwmarrin/discordgo"
)

func init() {
	RegisterCommand(drawHeadCommand, DrawHeadHandler)
}

var drawHeadCommand = &discordgo.ApplicationCommand{
	Name:        "抽頭",
	Description: "隨機抽取一顆伺服器成員頭顱",
}

type avatarResult struct {
	ID              int
	DiscordID       string
	DiscordUsername string
	GameUID         string
	ImageURL        string
	EmojiID         sql.NullString
	EmojiName       sql.NullString
}

// /抽頭 command
func DrawHeadHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// 隨機從DB抽仇
	avatar, err := getRandomAvatar()
	if err != nil {
		log.Printf("Failed to get random avatar: %v", err)

		var description string
		if err == sql.ErrNoRows {
			description = "目前還沒有人上傳頭顱！\n使用 `/上傳頭顱` 成為第一個捐獻者吧～"
		} else {
			description = "抽取頭顱時發生錯誤，請稍後再試"
		}

		e := embed.New().
			Title("抽頭失敗").
			Description(description).
			Color(embed.ColorRed).
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

	// 表符
	var emojiStr string
	if avatar.EmojiID.Valid && avatar.EmojiName.Valid {
		emojiStr = fmt.Sprintf("<:%s:%s>", avatar.EmojiName.String, avatar.EmojiID.String)
	}

	// 組 response
	eb := embed.New().
		Title("🎲 抽到了一顆頭！").
		Color(embed.ColorBlurple).
		Thumbnail(avatar.ImageURL).
		InlineField("Heartopia UID", avatar.GameUID).
		InlineField("Discord", fmt.Sprintf("<@%s>", avatar.DiscordID))

	if emojiStr != "" {
		eb.Description(fmt.Sprintf("恭喜你抽到了 <@%s> 的頭顱！%s", avatar.DiscordID, emojiStr))
	} else {
		eb.Description(fmt.Sprintf("恭喜你抽到了 <@%s> 的頭顱！", avatar.DiscordID))
	}

	e := eb.Timestamp().Build()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{e},
		},
	})

	// Record the drawn head
	var drawerID string
	if i.Member != nil {
		drawerID = i.Member.User.ID
	} else if i.User != nil {
		drawerID = i.User.ID
	}
	if drawerID != "" {
		if err := recordDrawnHead(drawerID, avatar.ID); err != nil {
			log.Printf("Failed to record drawn head: %v", err)
		}
	}
}

// getRandomAvatar 從DB隨機抽頭
func getRandomAvatar() (*avatarResult, error) {
	var avatar avatarResult
	err := database.DB.QueryRow(`
		SELECT id, discord_id, discord_username, game_uid, image_url, emoji_id, emoji_name
		FROM user_avatars
		ORDER BY RAND()
		LIMIT 1
	`).Scan(
		&avatar.ID,
		&avatar.DiscordID,
		&avatar.DiscordUsername,
		&avatar.GameUID,
		&avatar.ImageURL,
		&avatar.EmojiID,
		&avatar.EmojiName,
	)
	if err != nil {
		return nil, err
	}
	return &avatar, nil
}

// 抽到的頭寫入DB
func recordDrawnHead(drawerDiscordID string, avatarID int) error {
	_, err := database.DB.Exec(
		"INSERT INTO drawn_heads (drawer_discord_id, avatar_id) VALUES (?, ?)",
		drawerDiscordID, avatarID,
	)
	return err
}

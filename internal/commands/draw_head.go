package commands

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"purrtopia/internal/database"
	"purrtopia/internal/embed"

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
	Type            string
	DiscordID       sql.NullString
	DiscordUsername sql.NullString
	GameUID         string
	ImageURL        string
	EmojiID         sql.NullString
	EmojiName       sql.NullString
	Weight          int
	Rarity          string
}

// /抽頭 command
func DrawHeadHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// 隨機從DB抽仇(權重)
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
    InlineField("稀有度", avatar.Rarity). // 顯示稀有度
    InlineField("名稱/UID", avatar.GameUID)

	var descriptionText string

	// 判斷是玩家還是 NPC 來決定顯示方式
	if avatar.Type == "NPC" || !avatar.DiscordID.Valid {
		eb.Description(fmt.Sprintf("恭喜你抽到了 **%s** 的頭顱！(NPC)", avatar.GameUID))
	} else {
		// 是玩家，可以 Ping 他
		eb.InlineField("Discord", fmt.Sprintf("<@%s>", avatar.DiscordID.String))
		eb.Description(fmt.Sprintf("恭喜你抽到了 <@%s> 的頭顱！", avatar.DiscordID.String))
	}

	// 將 emojiStr 加回描述中
	if emojiStr != "" {
		descriptionText += " " + emojiStr
	}
	eb.Description(descriptionText)

	e := eb.Timestamp().Build()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{e},
		},
	})

	// 記錄抽到的頭 (如果是 NPC，drawer_discord_id 依然是觸發指令的人)
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

// getRandomAvatar 從DB隨機抽頭 (加權隨機)
func getRandomAvatar() (*avatarResult, error) {
    // 1. 撈出所有可以被抽的頭 (包含 USER 和 NPC)
    rows, err := database.DB.Query(`
        SELECT id, type, discord_id, discord_username, game_uid, image_url, emoji_id, emoji_name, weight, rarity
        FROM user_avatars
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var candidates []*avatarResult
    var totalWeight int

    // 2. 讀取資料並計算總權重
    for rows.Next() {
        var av avatarResult
        err := rows.Scan(
            &av.ID, &av.Type, &av.DiscordID, &av.DiscordUsername, 
            &av.GameUID, &av.ImageURL, &av.EmojiID, &av.EmojiName, 
            &av.Weight, &av.Rarity,
        )
        if err != nil {
            continue
        }
        // 權重 <= 0 就不給抽
        if av.Weight > 0 {
            candidates = append(candidates, &av)
            totalWeight += av.Weight
        }
    }

    if len(candidates) == 0 {
        return nil, sql.ErrNoRows
    }

    // 3. 隨機擲骰子 (Weighted Random Selection)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomValue := r.Intn(totalWeight) // 0 到 totalWeight-1

	currentSum := 0
	for _, av := range candidates {
		currentSum += av.Weight
		if randomValue < currentSum {
			return av, nil // 命中這個頭
		}
	}

    // 理論上不會跑到這裡，回傳最後一個當備案
    return candidates[len(candidates)-1], nil
}

// 抽到的頭寫入DB
func recordDrawnHead(drawerDiscordID string, avatarID int) error {
	_, err := database.DB.Exec(
		"INSERT INTO drawn_heads (drawer_discord_id, avatar_id) VALUES (?, ?)",
		drawerDiscordID, avatarID,
	)
	return err
}

package gacha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"regexp"
	"strings"

	"purrtopia/internal/commands"
	"purrtopia/internal/component"
	"purrtopia/internal/database"
	"purrtopia/internal/embed"

	"github.com/bwmarrin/discordgo"
)

// func init() {
// 	commands.RegisterCommand(uploadAvatarCommand, UploadAvatarHandler)
// 	commands.RegisterComponent("upload_avatar_btn", UploadAvatarOpenModalHandler)
// 	commands.RegisterModal("upload_avatar_modal", UploadAvatarModalSubmitHandler)
// }

var uploadAvatarCommand = &discordgo.ApplicationCommand{
	Name:        "上傳頭顱",
	Description: "上傳你的遊戲頭像圖片",
}

// /上傳頭顱 command
// TODO: 要做僅管理員? 常駐msg? 反正用戶只要點表單按鈕? 那就要一個專門的頻道
func UploadAvatarHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	e := embed.New().
		Title("上傳頭顱").
		Description(
			"請點擊下方按鈕填寫資料\n\n"+
				"**需要準備：**\n"+
				"• 你的遊戲 UID\n"+
				"• 頭像圖片連結\n\n"+
				"**推薦免費圖床：**\n"+
				"• [Imgur](https://imgur.com/) - 最常用，免註冊\n"+
				"• [ImgBB](https://imgbb.com/) - 簡單快速\n"+
				"• [Postimages](https://postimages.org/) - 無需註冊\n\n"+
				"上傳圖片後複製「直接連結」(Direct Link) 即可\n\n"+
				"⚠️ **注意：** 若已上傳過頭顱，再次上傳將會覆蓋原有頭顱").
		Color(embed.ColorBlurple).
		Footer("使用本功能即代表您同意 Bot 儲存並使用您上傳的資訊", "").
		Build()

	openBtn := component.NewButton().
		CustomID("upload_avatar_btn").
		Label("填寫資料").
		Primary().
		Emoji("📝").
		Build()
	btnRow := component.SingleButtonRow(openBtn)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{e},
			Components: []discordgo.MessageComponent{btnRow},
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// 表單
func UploadAvatarOpenModalHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	uidInput := component.NewTextInput().
		CustomID("avatar_game_uid").
		Label("遊戲 UID").
		Placeholder("請輸入你的遊戲 UID").
		Short().
		Required().
		MinLength(1).
		MaxLength(64).
		Build()

	imageInput := component.NewTextInput().
		CustomID("avatar_image_url").
		Label("頭像圖片連結").
		Placeholder("https://i.imgur.com/xxxxx.png").
		Short().
		Required().
		MinLength(10).
		MaxLength(500).
		Build()

	modal := component.NewModal().
		CustomID("upload_avatar_modal").
		Title("上傳頭顱").
		AddTextInput(uidInput).
		AddTextInput(imageInput).
		Build()

	s.InteractionRespond(i.Interaction, modal)
}

// 提交表單
func UploadAvatarModalSubmitHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	gameUID := strings.TrimSpace(component.GetModalValue(data, "avatar_game_uid"))
	imageURL := strings.TrimSpace(component.GetModalValue(data, "avatar_image_url"))

	// 入參檢核
	if !isValidImageURL(imageURL) {
		e := embed.New().
			Title("上傳失敗").
			Description("請提供有效的圖片連結（需以 http:// 或 https:// 開頭）").
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

	user := i.Member.User

	// 把圖片弄成emoji
	emoji, err := createApplicationEmoji(s, gameUID, imageURL)
	if err != nil {
		log.Printf("Failed to create emoji: %v", err)
		e := embed.New().
			Title("上傳失敗").
			Description("創建表情符號時發生錯誤：" + err.Error()).
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

	// 寫入DB
	err = saveUserAvatar(user.ID, user.Username, gameUID, imageURL, emoji.ID, emoji.Name)
	if err != nil {
		log.Printf("Failed to save avatar: %v", err)
		// 如果DB寫入失敗的話 要把bot的emiji刪掉 當作從一開始就不成功
		s.ApplicationEmojiDelete(s.State.User.ID, emoji.ID)

		e := embed.New().
			Title("上傳失敗").
			Description("儲存資料時發生錯誤，請稍後再試").
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

	// Success response
	emojiStr := fmt.Sprintf("<:%s:%s>", emoji.Name, emoji.ID)
	e := embed.New().
		Title("上傳成功！").
		Description(fmt.Sprintf("你的頭顱已成功上傳 %s\n現在可以使用 `/抽頭` 來抽其他人的頭哦～", emojiStr)).
		Color(embed.ColorSuccess).
		Thumbnail(imageURL).
		InlineField("Heartopia UID", gameUID).
		InlineField("Discord", fmt.Sprintf("<@%s>", user.ID)+emojiStr).
		Footer("感謝捐獻你的頭！", "").
		Timestamp().
		Build()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{e},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// 頭像img入參檢核
func isValidImageURL(url string) bool {
	url = strings.ToLower(url)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}
	return true
}

// 把圖片變成 bot emoji
func createApplicationEmoji(s *discordgo.Session, gameUID, imageURL string) (*discordgo.Emoji, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("無法下載圖片: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("圖片下載失敗 (HTTP %d)", resp.StatusCode)
	}

	if resp.ContentLength > 256*1024 {
		return nil, fmt.Errorf("圖片太大，請使用小於 256KB 的圖片")
	}

	imgData, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024+1))
	if err != nil {
		return nil, fmt.Errorf("讀取圖片失敗: %w", err)
	}

	if len(imgData) > 256*1024 {
		return nil, fmt.Errorf("圖片太大，請使用小於 256KB 的圖片")
	}

	contentType := http.DetectContentType(imgData)
	var img image.Image
	var decodeErr error

	switch contentType {
	case "image/png":
		img, decodeErr = png.Decode(bytes.NewReader(imgData))
	case "image/jpeg":
		img, decodeErr = jpeg.Decode(bytes.NewReader(imgData))
	case "image/gif":
		img, decodeErr = gif.Decode(bytes.NewReader(imgData))
	default:
		return nil, fmt.Errorf("不支援的圖片格式，請使用 PNG、JPG 或 GIF")
	}

	if decodeErr != nil {
		return nil, fmt.Errorf("圖片解碼失敗: %w", decodeErr)
	}

	// 把圖片變圓形
	circularImg := makeCircular(img)

	var buf bytes.Buffer
	if err := png.Encode(&buf, circularImg); err != nil {
		return nil, fmt.Errorf("圖片編碼失敗: %w", err)
	}

	if buf.Len() > 256*1024 {
		return nil, fmt.Errorf("處理後圖片太大，請使用較小的圖片")
	}

	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURI := fmt.Sprintf("data:image/png;base64,%s", base64Img)

	// emoji name
	emojiName := sanitizeEmojiName("head_" + gameUID)

	appID := s.State.User.ID

	// 要是有重複的 emoji name 就刪掉重建 因為一個人只能有一個頭
	existingEmojis, err := s.ApplicationEmojis(appID)
	if err == nil {
		for _, e := range existingEmojis {
			if e.Name == emojiName {
				s.ApplicationEmojiDelete(appID, e.ID)
				break
			}
		}
	}

	// 建立 bot emiji
	emoji, err := s.ApplicationEmojiCreate(appID, &discordgo.EmojiParams{
		Name:  emojiName,
		Image: dataURI,
	})
	if err != nil {
		return nil, fmt.Errorf("創建表情失敗: %w", err)
	}

	return emoji, nil
}

// 建立圓形圖片
func makeCircular(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// 取最小的邊當作圓的基礎
	size := w
	if h < w {
		size = h
	}

	dst := image.NewRGBA(image.Rect(0, 0, size, size))

	// 計算中心和半徑
	center := float64(size) / 2
	radius := center

	// 居中
	offsetX := (w - size) / 2
	offsetY := (h - size) / 2

	// 圓形遮罩
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center + 0.5
			dy := float64(y) - center + 0.5
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= radius {
				srcX := x + offsetX + bounds.Min.X
				srcY := y + offsetY + bounds.Min.Y
				dst.Set(x, y, src.At(srcX, srcY))
			} else {
				dst.Set(x, y, color.Transparent)
			}
		}
	}

	return dst
}

// 檢查無效字元
func sanitizeEmojiName(name string) string {
	// 只能英文、數字、_
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	name = reg.ReplaceAllString(name, "_")
	if len(name) < 2 {
		name = "head_" + name
	}
	if len(name) > 32 {
		name = name[:32]
	}
	return name
}

// 把用戶上傳的頭寫入DB
func saveUserAvatar(discordID, discordUsername, gameUID, imageURL, emojiID, emojiName string) error {
	// 先檢查是否已有
	var exists bool
	var oldEmojiID *string
	err := database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM user_avatars WHERE discord_id = ?)",
		discordID,
	).Scan(&exists)
	if err != nil {
		return err
	}

	// 如果有的話就拿舊的row更新, 沒有的話就直接 insert
	if exists {
		database.DB.QueryRow(
			"SELECT emoji_id FROM user_avatars WHERE discord_id = ?",
			discordID,
		).Scan(&oldEmojiID)

		_, err = database.DB.Exec(
			"UPDATE user_avatars SET discord_username = ?, game_uid = ?, image_url = ?, emoji_id = ?, emoji_name = ? WHERE discord_id = ?",
			discordUsername, gameUID, imageURL, emojiID, emojiName, discordID,
		)
	} else {
		_, err = database.DB.Exec(
			"INSERT INTO user_avatars (discord_id, discord_username, game_uid, image_url, emoji_id, emoji_name) VALUES (?, ?, ?, ?, ?, ?)",
			discordID, discordUsername, gameUID, imageURL, emojiID, emojiName,
		)
	}

	return err
}

// Ensure commands package is imported (for future init registration)
var _ = commands.RegisterCommand

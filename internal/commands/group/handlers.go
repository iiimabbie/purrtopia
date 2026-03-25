package group

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"purrtopia/internal/commands"
	"purrtopia/internal/database"
	"purrtopia/internal/database/models"

	"github.com/bwmarrin/discordgo"
)

// slashInteractions 暫存 /揪團 的原始 interaction，供 modal submit 回來編輯原始 ephemeral
// key: userID (string), value: *discordgo.Interaction
var slashInteractions sync.Map

func init() {
	commands.RegisterCommand(groupCommand, GroupHandler)

	// 類型選擇按鈕（在 ephemeral 中）
	commands.RegisterComponent("group_type_rainbow", handleTypeRainbow)
	commands.RegisterComponent("group_type_fishing", handleTypeFishing)

	// 表單提交
	commands.RegisterModal("group_modal_rainbow", handleModalRainbow)
	commands.RegisterModal("group_modal_fishing", handleModalFishing)

	// 面板按鈕
	commands.RegisterComponentPrefix("group_register_", handleRegister)
	commands.RegisterComponentPrefix("group_withdraw_", handleWithdraw)
	commands.RegisterComponentPrefix("group_dissolve_", handleDissolve)

	// 討論串按鈕
	commands.RegisterComponentPrefix("group_finish_", handleFinish)
}

var groupCommand = &discordgo.ApplicationCommand{
	Name:        "揪團",
	Description: "建立揪團活動",
}

// ============================================
// 主入口
// ============================================

// GroupHandler /揪團 斜線指令：顯示類型選擇按鈕
func GroupHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !channelCheck(s, i) {
		return
	}
	userID := getUserID(i)
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    "請選擇揪團類型：",
			Components: buildTypeButtons(),
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		return
	}
	// 暫存原始 interaction，讓 modal submit 可以回來編輯這則 ephemeral
	slashInteractions.Store(userID, i.Interaction)
}

// ============================================
// 類型按鈕 handlers → 開 modal
// ============================================

func handleTypeRainbow(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, buildActivityModal(ActivityRainbow))
}

func handleTypeFishing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, buildActivityModal(ActivityFishing))
}

// ============================================
// Modal 提交 handlers → 發布公有面板
// ============================================

func handleModalRainbow(s *discordgo.Session, i *discordgo.InteractionCreate) {
	handleModalSubmit(s, i, ActivityRainbow)
}

func handleModalFishing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	handleModalSubmit(s, i, ActivityFishing)
}

func handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate, actType ActivityType) {
	if !channelCheck(s, i) {
		return
	}

	userID := getUserID(i)
	if userID == "" {
		respondWithError(s, i)
		return
	}

	// 解析表單欄位（純記憶體，在 acknowledge 前做）
	data := i.ModalSubmitData()
	var actTime, maxMembersStr, requirements string
	for _, row := range data.Components {
		if ar, ok := row.(*discordgo.ActionsRow); ok {
			for _, comp := range ar.Components {
				if ti, ok := comp.(*discordgo.TextInput); ok {
					switch ti.CustomID {
					case "group_time":
						actTime = strings.TrimSpace(ti.Value)
					case "group_max_members":
						maxMembersStr = strings.TrimSpace(ti.Value)
					case "group_requirements":
						requirements = strings.TrimSpace(ti.Value)
					}
				}
			}
		}
	}

	if actTime == "" {
		respondWithEphemeral(s, i, "❌ 請填寫揪團時間")
		return
	}

	maxMembersNum, err := strconv.Atoi(maxMembersStr)
	if err != nil || maxMembersNum < 2 || maxMembersNum > 99 {
		respondWithEphemeral(s, i, "❌ 人數上限請填 2～99 的數字")
		return
	}

	// 立即 acknowledge（deferred ephemeral），防止 Discord 3 秒逾時後重發 interaction
	// 若 acknowledge 失敗代表此 interaction 已被另一個實例處理，直接 return
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		log.Printf("[group] handleModalSubmit acknowledge 失敗（可能是重複 interaction）: %v", err)
		return
	}

	// 隨機選一個未被使用的團名
	usedNames, _ := database.GroupActivityRepo.FindActiveThreadNames(string(actType))
	threadName, err := database.GroupThreadNameRepo.FindRandomUnused(string(actType), usedNames)
	if err != nil {
		log.Printf("[group] 取得團名失敗，使用預設值: %v", err)
		threadName = actType.Name() + "揪"
	}

	// 建立 DB 記錄
	activity := &models.GroupActivity{
		OrganizerID:  userID,
		ActivityType: string(actType),
		ActivityTime: actTime,
		MaxMembers:   maxMembersNum,
		Requirements: requirements,
		ThreadName:   threadName,
	}
	if err := database.GroupActivityRepo.Create(activity); err != nil {
		log.Printf("[group] 建立活動失敗: %v", err)
		editWithError(s, i)
		return
	}

	// 發布公有面板訊息
	embed := buildActivityPanelEmbed(activity)
	components := buildPanelButtons(activity.ID, false)
	content := fmt.Sprintf("<@&%s> %s", actType.RoleID(), actType.Announcement())

	msg, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Content:    content,
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		log.Printf("[group] 發布面板訊息失敗: %v", err)
		editWithError(s, i)
		return
	}

	if err := database.GroupActivityRepo.UpdateMessageIDs(activity.ID, msg.ID, i.ChannelID); err != nil {
		log.Printf("[group] 更新 MessageID 失敗: %v", err)
	}

	// 把 A 訊息（請選擇揪團類型：）編輯成 B 訊息（✅ 揪團已發布！，移除按鈕）
	if origIntf, ok := slashInteractions.LoadAndDelete(userID); ok {
		origInteraction := origIntf.(*discordgo.Interaction)
		successMsg := "✅ 揪團已發布！"
		emptyComponents := []discordgo.MessageComponent{}
		if _, err := s.InteractionResponseEdit(origInteraction, &discordgo.WebhookEdit{
			Content:    &successMsg,
			Components: &emptyComponents,
		}); err != nil {
			log.Printf("[group] 編輯 A→B 失敗（可能已消失）: %v", err)
		}
	}
	// 刪除 modal 的 deferred ephemeral（不顯示任何內容）
	s.InteractionResponseDelete(i.Interaction)
}

// ============================================
// 報名按鈕 handler
// ============================================

// handleRegister 處理報名按鈕
func handleRegister(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !channelCheck(s, i) {
		return
	}

	userID := getUserID(i)
	activityID := extractActivityID(i.MessageComponentData().CustomID, "group_register_")
	if activityID == 0 {
		respondWithError(s, i)
		return
	}

	// 立即 acknowledge（必須在 DB query 之前，否則 Discord 3 秒內沒收到回應會報 10062）
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("[group] handleRegister acknowledge 失敗: %v", err)
		return
	}

	activity, err := database.GroupActivityRepo.FindByID(activityID)
	if err != nil {
		log.Printf("[group] handleRegister 查詢活動失敗 id=%d: %v", activityID, err)
		return
	}

	// 發起人不能報名（已自動算在人數內）
	if userID == activity.OrganizerID {
		sendFollowupEphemeral(s, i, "❌ 你是揪團發起人，不需要報名喔！")
		return
	}

	// 已報名檢查
	if hasMember(activity.Members, userID) {
		sendFollowupEphemeral(s, i, "⚠️ 你已經報名了！")
		return
	}

	// 人數上限檢查（發起人佔 1 個名額，所以最多 MaxMembers-1 個報名者）
	if memberCount(activity.Members) >= activity.MaxMembers-1 {
		sendFollowupEphemeral(s, i, "❌ 報名人數已滿，無法報名")
		return
	}

	// 開討論串（或取得已存在的）
	actType := ActivityType(activity.ActivityType)
	isFirstMember := activity.ThreadID == ""
	var threadID string

	if isFirstMember {
		name := activity.ThreadName
		if name == "" {
			name = actType.Name()
		}
		threadName := fmt.Sprintf("%s%s揪團", time.Now().Format("0102"), name)
		thread, err := s.MessageThreadStart(activity.ChannelID, activity.MessageID, threadName, 1440)
		if err != nil {
			log.Printf("[group] 開討論串失敗: %v", err)
			// 160004: thread already exists — Discord thread ID = message ID
			threadID = activity.MessageID
		} else {
			threadID = thread.ID
		}
	} else {
		threadID = activity.ThreadID
	}

	// 在討論串發歡迎訊息並 @mention（Discord 自動將被 mention 的人加入討論串）
	var welcomeMsg string
	if isFirstMember {
		// 第一個報名者：同時 mention 發起人，讓雙方都進討論串
		welcomeMsg = fmt.Sprintf("<@%s> <@%s> 歡迎加入！", activity.OrganizerID, userID)
	} else {
		welcomeMsg = fmt.Sprintf("<@%s> 歡迎加入！", userID)
	}
	if _, err := s.ChannelMessageSend(threadID, welcomeMsg); err != nil {
		log.Printf("[group] 發送歡迎訊息失敗: %v", err)
	}

	// 第一個成員加入時，補發揪團結束按鈕（僅揪團人可用）並釘選
	if isFirstMember {
		finishMsg, err := s.ChannelMessageSendComplex(threadID, &discordgo.MessageSend{
			Content:    "這是一顆成功揪完團結束的按鈕，只有揪團人可以點擊。揪團完美結束後請發起人點擊此按鈕🫶🏻",
			Components: buildFinishButton(activityID),
		})
		if err != nil {
			log.Printf("[group] 發送揪團結束按鈕失敗: %v", err)
		} else {
			if err := s.ChannelMessagePin(threadID, finishMsg.ID); err != nil {
				log.Printf("[group] 釘選揪團結束按鈕失敗: %v", err)
			}
		}
	}

	// 更新 DB
	newMembers := addMember(activity.Members, userID)
	// 顯示人數 = 報名者 + 發起人（+1）
	newDisplayCount := memberCount(newMembers) + 1
	isFull := newDisplayCount >= activity.MaxMembers

	if err := database.GroupActivityRepo.UpdateThread(activityID, threadID, newMembers); err != nil {
		log.Printf("[group] 更新討論串成員失敗: %v", err)
	}

	// 更新面板訊息（人數 + 按鈕狀態）
	updatedActivity := *activity
	updatedActivity.Members = newMembers
	embed := buildActivityPanelEmbed(&updatedActivity)
	components := buildPanelButtons(activityID, isFull)
	content := fmt.Sprintf("<@&%s> %s", actType.RoleID(), actType.Announcement())
	noMentions := &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}}

	if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:              activity.MessageID,
		Channel:         activity.ChannelID,
		Content:         &content,
		Embeds:          &[]*discordgo.MessageEmbed{embed},
		Components:      &components,
		AllowedMentions: noMentions,
	}); err != nil {
		log.Printf("[group] 更新面板訊息失敗: %v", err)
	}
}

// ============================================
// 退出報名 handler
// ============================================

// handleWithdraw 處理「🚪 退出報名」按鈕：從報名名單移除並通知討論串
func handleWithdraw(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !channelCheck(s, i) {
		return
	}

	userID := getUserID(i)
	activityID := extractActivityID(i.MessageComponentData().CustomID, "group_withdraw_")
	if activityID == 0 {
		respondWithError(s, i)
		return
	}

	// 立即 acknowledge（必須在 DB query 之前）
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("[group] handleWithdraw acknowledge 失敗: %v", err)
		return
	}

	activity, err := database.GroupActivityRepo.FindByID(activityID)
	if err != nil {
		log.Printf("[group] handleWithdraw 查詢活動失敗 id=%d: %v", activityID, err)
		return
	}

	// 揪團人無法退出
	if userID == activity.OrganizerID {
		sendFollowupEphemeral(s, i, "❌ 你是揪團發起人，無法退出")
		return
	}

	// 未報名者無法退出
	if !hasMember(activity.Members, userID) {
		sendFollowupEphemeral(s, i, "⚠️ 你還沒有報名，無法退出")
		return
	}

	// 更新 DB：移除該報名者
	newMembers := removeMember(activity.Members, userID)
	if err := database.GroupActivityRepo.UpdateThread(activityID, activity.ThreadID, newMembers); err != nil {
		log.Printf("[group] handleWithdraw 更新成員失敗: %v", err)
	}

	// 若有討論串：先踢出成員，再發退出通知（避免 @mention 觸發 Discord auto-add）
	if activity.ThreadID != "" {
		if err := s.ThreadMemberRemove(activity.ThreadID, userID); err != nil {
			log.Printf("[group] handleWithdraw 踢出討論串成員失敗: %v", err)
		}
		displayName := getDisplayName(i)
		if _, err := s.ChannelMessageSend(activity.ThreadID, fmt.Sprintf("%s 離開本團了，我們目送他 👋", displayName)); err != nil {
			log.Printf("[group] handleWithdraw 發退出通知失敗: %v", err)
		}
	}

	// 更新面板訊息（人數 -1，如有需要重新開放報名按鈕）
	updatedActivity := *activity
	updatedActivity.Members = newMembers
	newDisplayCount := memberCount(newMembers) + 1 // +1 為發起人
	isFull := newDisplayCount >= activity.MaxMembers
	embed := buildActivityPanelEmbed(&updatedActivity)
	actType := ActivityType(activity.ActivityType)
	components := buildPanelButtons(activityID, isFull)
	content := fmt.Sprintf("<@&%s> %s", actType.RoleID(), actType.Announcement())
	noMentions := &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}}

	if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:              activity.MessageID,
		Channel:         activity.ChannelID,
		Content:         &content,
		Embeds:          &[]*discordgo.MessageEmbed{embed},
		Components:      &components,
		AllowedMentions: noMentions,
	}); err != nil {
		log.Printf("[group] handleWithdraw 更新面板失敗: %v", err)
	}
}

// ============================================
// 流團 handler
// ============================================

// handleDissolve 處理「流團了 T_T」按鈕：在討論串通知並附上離開按鈕
func handleDissolve(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !channelCheck(s, i) {
		return
	}

	userID := getUserID(i)
	activityID := extractActivityID(i.MessageComponentData().CustomID, "group_dissolve_")
	if activityID == 0 {
		respondWithError(s, i)
		return
	}

	// 立即 acknowledge（必須在 DB query 之前）
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("[group] handleDissolve acknowledge 失敗: %v", err)
		return
	}

	activity, err := database.GroupActivityRepo.FindByID(activityID)
	if err != nil {
		log.Printf("[group] handleDissolve 查詢活動失敗 id=%d: %v", activityID, err)
		return
	}

	// 只有發起人可以宣布流團
	if userID != activity.OrganizerID {
		sendFollowupEphemeral(s, i, "❌ 只有揪團發起人可以宣布流團")
		return
	}

	// 若有討論串：發流團通知並 archive + lock
	if activity.ThreadID != "" {
		if _, err = s.ChannelMessageSend(activity.ThreadID, "@everyone 流團了 😢"); err != nil {
			log.Printf("[group] 發送流團通知失敗: %v", err)
		}
		archived := true
		locked := true
		if _, err := s.ChannelEditComplex(activity.ThreadID, &discordgo.ChannelEdit{
			Archived: &archived,
			Locked:   &locked,
		}); err != nil {
			log.Printf("[group] handleDissolve 關閉討論串失敗: %v", err)
		}
	}

	// 刪除 C 訊息
	if activity.MessageID != "" && activity.ChannelID != "" {
		if err := s.ChannelMessageDelete(activity.ChannelID, activity.MessageID); err != nil {
			log.Printf("[group] handleDissolve 刪除面板訊息失敗: %v", err)
		}
	}

	// 刪除 DB 記錄
	if err := database.GroupActivityRepo.Delete(activityID); err != nil {
		log.Printf("[group] handleDissolve 刪除 DB 記錄失敗: %v", err)
	}
}

// ============================================
// 揪團結束 handler
// ============================================

// handleFinish 揪團人點擊「揪團結束」：發結束公告並關閉討論串、刪除 C 訊息
// 注意：此 handler 在討論串內觸發，i.ChannelID 為討論串 ID，不做頻道限制檢查
func handleFinish(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := getUserID(i)
	activityID := extractActivityID(i.MessageComponentData().CustomID, "group_finish_")
	if activityID == 0 {
		respondWithError(s, i)
		return
	}

	// 立即 acknowledge（必須在 DB query 之前，否則 Discord 3 秒內沒回應會報 10062）
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("[group] handleFinish acknowledge 失敗: %v", err)
		return
	}

	activity, err := database.GroupActivityRepo.FindByID(activityID)
	if err != nil {
		log.Printf("[group] handleFinish 查詢活動失敗 id=%d: %v", activityID, err)
		return
	}

	// 只有揪團人可以點擊
	if userID != activity.OrganizerID {
		sendFollowupEphemeral(s, i, "❌ 只有揪團人可以點擊這顆按鈕")
		return
	}

	// 在討論串發結束公告，然後 archive + lock
	if activity.ThreadID != "" {
		if _, err := s.ChannelMessageSend(activity.ThreadID, "@everyone 謝謝大家參與！"); err != nil {
			log.Printf("[group] handleFinish 發送結束通知失敗: %v", err)
		}
		archived := true
		locked := true
		if _, err := s.ChannelEditComplex(activity.ThreadID, &discordgo.ChannelEdit{
			Archived: &archived,
			Locked:   &locked,
		}); err != nil {
			log.Printf("[group] handleFinish 關閉討論串失敗: %v", err)
		}
	}

	// 刪除 C 訊息（面板）
	if activity.MessageID != "" && activity.ChannelID != "" {
		if err := s.ChannelMessageDelete(activity.ChannelID, activity.MessageID); err != nil {
			log.Printf("[group] handleFinish 刪除面板訊息失敗: %v", err)
		}
	}

	// 刪除 DB 記錄
	if err := database.GroupActivityRepo.Delete(activityID); err != nil {
		log.Printf("[group] handleFinish 刪除 DB 記錄失敗: %v", err)
	}
}

package group

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// buildTypeButtons 建立揪團類型選擇按鈕
func buildTypeButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    ActivityRainbow.Label(),
					Style:    discordgo.PrimaryButton,
					CustomID: "group_type_rainbow",
				},
				discordgo.Button{
					Label:    ActivityFishing.Label(),
					Style:    discordgo.PrimaryButton,
					CustomID: "group_type_fishing",
				},
			},
		},
	}
}

// buildActivityModal 建立揪團表單 modal
func buildActivityModal(actType ActivityType) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "group_modal_" + string(actType),
			Title:    actType.Label() + " 揪團",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "group_time",
							Label:       "揪團日期",
							Style:       discordgo.TextInputShort,
							Placeholder: "例：12/25 晚上 9 點",
							Required:    true,
							MaxLength:   100,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "group_max_members",
							Label:       "揪團人數上限（含揪團人，只能填數字）",
							Style:       discordgo.TextInputShort,
							Placeholder: "例：4",
							Required:    true,
							MinLength:   1,
							MaxLength:   2,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "group_requirements",
							Label:       "詳細要求（可不填）",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "例：需要 Lv.30 以上、帶特定道具等",
							Required:    false,
							MaxLength:   500,
						},
					},
				},
			},
		},
	}
}

// buildPanelButtons 建立面板按鈕列（報名 + 退出報名 + 流團了）
func buildPanelButtons(activityID int, full bool) []discordgo.MessageComponent {
	registerLabel := "✋ 報名"
	if full {
		registerLabel = "✋ 已額滿"
	}
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    registerLabel,
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("group_register_%d", activityID),
					Disabled: full,
				},
				discordgo.Button{
					Label:    "🚪 退出",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("group_withdraw_%d", activityID),
				},
				discordgo.Button{
					Label:    "流團了 T_T",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("group_dissolve_%d", activityID),
				},
			},
		},
	}
}

// buildFinishButton 建立討論串內的「揪團結束」按鈕（僅揪團人可用）
func buildFinishButton(activityID int) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "揪團結束",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("group_finish_%d", activityID),
				},
			},
		},
	}
}

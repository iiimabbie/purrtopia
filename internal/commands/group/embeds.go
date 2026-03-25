package group

import (
	"fmt"

	"purrtopia/internal/database/models"

	"github.com/bwmarrin/discordgo"
)

// buildActivityPanelEmbed 建立揪團面板 embed
func buildActivityPanelEmbed(activity *models.GroupActivity) *discordgo.MessageEmbed {
	actType := ActivityType(activity.ActivityType)
	count := memberCount(activity.Members) + 1 // +1 為發起人

	fields := []*discordgo.MessageEmbedField{
		{Name: "揪團人", Value: fmt.Sprintf("<@%s>", activity.OrganizerID), Inline: true},
		{Name: "揪團時間", Value: activity.ActivityTime, Inline: true},
	}

	if activity.Requirements != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "詳細要求",
			Value: activity.Requirements,
		})
	}

	fields = append(fields, &discordgo.MessageEmbedField{
		Name:  "報名人數",
		Value: fmt.Sprintf("%d / %d", count, activity.MaxMembers),
	})

	title := actType.Emoji() + " " + actType.Name() + " 揪團"
	if activity.ThreadName != "" {
		title = actType.Emoji() + " " + activity.ThreadName
	}

	return &discordgo.MessageEmbed{
		Title:  title,
		Color:  actType.Color(),
		Fields: fields,
	}
}

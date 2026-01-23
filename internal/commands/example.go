package commands

import (
	"fmt"
	"strings"
	"time"

	"discord-bot-template/internal/component"
	"discord-bot-template/internal/embed"

	"github.com/bwmarrin/discordgo"
)

func init() {
	// 自動註冊指令
	RegisterCommand(exampleCommand, ExampleHandler)

	// 導航選單
	RegisterComponent("example_nav", ExampleNavHandler)

	// Embed 範例
	RegisterComponent("example_reload", ExampleReloadHandler)

	// Button 範例
	RegisterComponent("example_primary", ExampleButtonHandler)
	RegisterComponent("example_secondary", ExampleButtonHandler)
	RegisterComponent("example_success", ExampleButtonHandler)
	RegisterComponent("example_danger", ExampleButtonHandler)

	// Select 範例
	RegisterComponent("example_color_select", ExampleColorSelectHandler)
	RegisterComponent("example_user_select", ExampleUserSelectHandler)

	// Modal 範例
	RegisterComponent("example_open_modal", ExampleOpenModalHandler)
	RegisterModal("example_modal", ExampleModalSubmitHandler)
}

var exampleCommand = &discordgo.ApplicationCommand{
	Name:        "example",
	Description: "Interactive demo of all template features",
}

// ============================================
// Navigation
// ============================================

func buildNavSelect(current string) discordgo.ActionsRow {
	menu := component.NewSelect().
		CustomID("example_nav").
		Placeholder("Select a demo...").
		AddOptionWithEmoji("Embed Features", "embed", "Rich embed with all features", "📝").
		AddOptionWithEmoji("Buttons", "buttons", "All button styles", "🔘").
		AddOptionWithEmoji("Select Menus", "selects", "Dropdown menus", "📋").
		AddOptionWithEmoji("Modal Form", "modal", "Popup form demo", "📝").
		Build()

	// Mark current as default
	for i := range menu.Options {
		if menu.Options[i].Value == current {
			menu.Options[i].Default = true
		}
	}

	return component.SelectRow(menu)
}

// ============================================
// Page Builders
// ============================================

func buildEmbedPage(user *discordgo.User) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	e := embed.New().
		Title("Embed Features Demo").
		URL("https://discord.com").
		Description("This embed demonstrates all available features:\n\n"+
			embed.Bold("Bold")+", "+
			embed.Italic("Italic")+", "+
			embed.InlineCode("Code")+", "+
			embed.Spoiler("Spoiler")+"\n\n"+
			"Mention: "+embed.Mention(user.ID)).
		Color(embed.ColorFuchsia).
		Author(user.Username, "", user.AvatarURL("32")).
		Thumbnail(user.AvatarURL("128")).
		InlineField("Inline 1", "Value").
		InlineField("Inline 2", "Value").
		InlineField("Inline 3", "Value").
		BlockField("Block Field", "This takes full width").
		BlockField("Relative Time", embed.RelativeTime(time.Now().Add(-2*time.Hour))).
		Image("https://cdn-icons-png.flaticon.com/512/5277/5277459.png").
		Footer("Footer Text", "").
		Timestamp().
		Build()

	nav := buildNavSelect("embed")
	reload := component.ReloadButtonRow("example_reload")

	return e, []discordgo.MessageComponent{nav, reload}
}

func buildButtonsPage() (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	e := embed.New().
		Title("Button Styles Demo").
		Description("Click any button to see it in action.\n\n"+
			"**Available styles:**\n"+
			"• Primary (Blue)\n"+
			"• Secondary (Gray)\n"+
			"• Success (Green)\n"+
			"• Danger (Red)\n"+
			"• Link (External URL)").
		Color(embed.ColorBlurple).
		Build()

	nav := buildNavSelect("buttons")

	buttonRow := component.NewActionRow().
		AddButton(component.PrimaryButton("example_primary", "Primary")).
		AddButton(component.SecondaryButton("example_secondary", "Secondary")).
		AddButton(component.SuccessButton("example_success", "Success")).
		AddButton(component.DangerButton("example_danger", "Danger")).
		AddButton(component.LinkButton("https://discord.com", "Link")).
		Build()

	return e, []discordgo.MessageComponent{nav, buttonRow}
}

func buildSelectsPage() (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	e := embed.New().
		Title("Select Menu Demo").
		Description("**Available select types:**\n"+
			"• String Select - Custom options\n"+
			"• User Select - Pick a user\n"+
			"• Role Select - Pick a role\n"+
			"• Channel Select - Pick a channel\n\n"+
			"Try the menus below!").
		Color(embed.ColorGreen).
		Build()

	nav := buildNavSelect("selects")

	colorSelect := component.NewSelect().
		CustomID("example_color_select").
		Placeholder("Pick a color...").
		AddOptionWithEmoji("Red", "red", "Warm and passionate", "🔴").
		AddOptionWithEmoji("Green", "green", "Nature and growth", "🟢").
		AddOptionWithEmoji("Blue", "blue", "Calm and peaceful", "🔵").
		AddOptionWithEmoji("Purple", "purple", "Royal and creative", "🟣").
		Build()
	colorRow := component.SelectRow(colorSelect)

	userSelect := component.NewUserSelect("example_user_select").
		Placeholder("Pick a user...").
		Build()
	userRow := component.SelectRow(userSelect)

	return e, []discordgo.MessageComponent{nav, colorRow, userRow}
}

func buildModalPage() (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	e := embed.New().
		Title("Modal Form Demo").
		Description("Modals are popup forms that collect user input.\n\n"+
			"**Features:**\n"+
			"• Short text input (single line)\n"+
			"• Paragraph input (multi-line)\n"+
			"• Required/Optional fields\n"+
			"• Min/Max length validation\n\n"+
			"Click the button below to try it!").
		Color(embed.ColorGold).
		Build()

	nav := buildNavSelect("modal")

	openBtn := component.NewButton().
		CustomID("example_open_modal").
		Label("Open Form").
		Primary().
		Emoji("📝").
		Build()
	btnRow := component.SingleButtonRow(openBtn)

	return e, []discordgo.MessageComponent{nav, btnRow}
}

// ============================================
// Handlers
// ============================================

func ExampleHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	e, components := buildEmbedPage(i.Member.User)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{e},
			Components: components,
		},
	})
}

func ExampleNavHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	selected := i.MessageComponentData().Values[0]

	var e *discordgo.MessageEmbed
	var components []discordgo.MessageComponent

	switch selected {
	case "embed":
		e, components = buildEmbedPage(i.Member.User)
	case "buttons":
		e, components = buildButtonsPage()
	case "selects":
		e, components = buildSelectsPage()
	case "modal":
		e, components = buildModalPage()
	default:
		e, components = buildEmbedPage(i.Member.User)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{e},
			Components: components,
		},
	})
}

func ExampleReloadHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	e, components := buildEmbedPage(i.Member.User)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{e},
			Components: components,
		},
	})
}

func ExampleButtonHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	buttonID := i.MessageComponentData().CustomID

	styleMap := map[string]string{
		"example_primary":   "Primary",
		"example_secondary": "Secondary",
		"example_success":   "Success",
		"example_danger":    "Danger",
	}

	style := styleMap[buttonID]

	e := embed.New().
		Description(fmt.Sprintf("You clicked the **%s** button!", style)).
		Color(embed.ColorBlurple).
		Build()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{e},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func ExampleColorSelectHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	selected := i.MessageComponentData().Values[0]

	colorMap := map[string]int{
		"red":    embed.ColorRed,
		"green":  embed.ColorGreen,
		"blue":   embed.ColorBlue,
		"purple": embed.ColorPurple,
	}

	colorName := strings.ToUpper(selected[:1]) + selected[1:]

	e := embed.New().
		Description(fmt.Sprintf("You selected **%s**!", colorName)).
		Color(colorMap[selected]).
		Build()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{e},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func ExampleUserSelectHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.MessageComponentData().Values[0]

	e := embed.New().
		Description(fmt.Sprintf("You selected %s!", embed.Mention(userID))).
		Color(embed.ColorBlurple).
		Build()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{e},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func ExampleOpenModalHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	titleInput := component.NewTextInput().
		CustomID("modal_title").
		Label("Title").
		Placeholder("Enter a title...").
		Short().
		Required().
		MaxLength(100).
		Build()

	messageInput := component.NewTextInput().
		CustomID("modal_message").
		Label("Message").
		Placeholder("Enter your message...").
		Paragraph().
		Required().
		MinLength(10).
		MaxLength(500).
		Build()

	modal := component.NewModal().
		CustomID("example_modal").
		Title("Example Form").
		AddTextInput(titleInput).
		AddTextInput(messageInput).
		Build()

	s.InteractionRespond(i.Interaction, modal)
}

func ExampleModalSubmitHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	title := component.GetModalValue(data, "modal_title")
	message := component.GetModalValue(data, "modal_message")

	e := embed.New().
		Title("Form Submitted!").
		Color(embed.ColorSuccess).
		BlockField("Title", title).
		BlockField("Message", message).
		Footer(fmt.Sprintf("By %s", i.Member.User.Username), i.Member.User.AvatarURL("32")).
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

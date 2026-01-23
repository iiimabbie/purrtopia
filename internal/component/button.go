package component

import "github.com/bwmarrin/discordgo"

// Button styles
const (
	StylePrimary   = discordgo.PrimaryButton   // 藍色
	StyleSecondary = discordgo.SecondaryButton // 灰色
	StyleSuccess   = discordgo.SuccessButton   // 綠色
	StyleDanger    = discordgo.DangerButton    // 紅色
	StyleLink      = discordgo.LinkButton      // 連結（不觸發互動）
)

// ButtonBuilder builds a single button
type ButtonBuilder struct {
	button discordgo.Button
}

// NewButton creates a new button builder
func NewButton() *ButtonBuilder {
	return &ButtonBuilder{
		button: discordgo.Button{
			Style: discordgo.PrimaryButton,
		},
	}
}

// Label sets the button text
func (b *ButtonBuilder) Label(label string) *ButtonBuilder {
	b.button.Label = label
	return b
}

// CustomID sets the button's custom ID (used to identify clicks)
func (b *ButtonBuilder) CustomID(id string) *ButtonBuilder {
	b.button.CustomID = id
	return b
}

// Style sets the button style
func (b *ButtonBuilder) Style(style discordgo.ButtonStyle) *ButtonBuilder {
	b.button.Style = style
	return b
}

// Primary sets blue style
func (b *ButtonBuilder) Primary() *ButtonBuilder {
	b.button.Style = discordgo.PrimaryButton
	return b
}

// Secondary sets gray style
func (b *ButtonBuilder) Secondary() *ButtonBuilder {
	b.button.Style = discordgo.SecondaryButton
	return b
}

// Success sets green style
func (b *ButtonBuilder) Success() *ButtonBuilder {
	b.button.Style = discordgo.SuccessButton
	return b
}

// Danger sets red style
func (b *ButtonBuilder) Danger() *ButtonBuilder {
	b.button.Style = discordgo.DangerButton
	return b
}

// Link sets link style (requires URL, no CustomID)
func (b *ButtonBuilder) Link(url string) *ButtonBuilder {
	b.button.Style = discordgo.LinkButton
	b.button.URL = url
	return b
}

// Emoji sets the button emoji
func (b *ButtonBuilder) Emoji(name string) *ButtonBuilder {
	b.button.Emoji = &discordgo.ComponentEmoji{Name: name}
	return b
}

// EmojiCustom sets a custom emoji (from server)
func (b *ButtonBuilder) EmojiCustom(name, id string) *ButtonBuilder {
	b.button.Emoji = &discordgo.ComponentEmoji{Name: name, ID: id}
	return b
}

// Disabled sets the button as disabled
func (b *ButtonBuilder) Disabled() *ButtonBuilder {
	b.button.Disabled = true
	return b
}

// Build returns the button component
func (b *ButtonBuilder) Build() discordgo.Button {
	return b.button
}

// ============================================
// ActionRow Builder (容器)
// ============================================

// ActionRowBuilder builds a row of components
type ActionRowBuilder struct {
	components []discordgo.MessageComponent
}

// NewActionRow creates a new action row
func NewActionRow() *ActionRowBuilder {
	return &ActionRowBuilder{
		components: make([]discordgo.MessageComponent, 0),
	}
}

// AddButton adds a button to the row
func (r *ActionRowBuilder) AddButton(btn discordgo.Button) *ActionRowBuilder {
	r.components = append(r.components, btn)
	return r
}

// Build returns the action row
func (r *ActionRowBuilder) Build() discordgo.ActionsRow {
	return discordgo.ActionsRow{Components: r.components}
}

// ============================================
// Quick Helpers
// ============================================

// PrimaryButton creates a primary (blue) button
func PrimaryButton(customID, label string) discordgo.Button {
	return NewButton().CustomID(customID).Label(label).Primary().Build()
}

// SecondaryButton creates a secondary (gray) button
func SecondaryButton(customID, label string) discordgo.Button {
	return NewButton().CustomID(customID).Label(label).Secondary().Build()
}

// SuccessButton creates a success (green) button
func SuccessButton(customID, label string) discordgo.Button {
	return NewButton().CustomID(customID).Label(label).Success().Build()
}

// DangerButton creates a danger (red) button
func DangerButton(customID, label string) discordgo.Button {
	return NewButton().CustomID(customID).Label(label).Danger().Build()
}

// LinkButton creates a link button
func LinkButton(url, label string) discordgo.Button {
	return NewButton().Label(label).Link(url).Build()
}

// SingleButtonRow creates an action row with one button
func SingleButtonRow(btn discordgo.Button) discordgo.ActionsRow {
	return NewActionRow().AddButton(btn).Build()
}

// ============================================
// Common Button Presets
// ============================================

// ReloadButton creates a reload button with 🔄 emoji
func ReloadButton(customID string) discordgo.Button {
	return NewButton().
		CustomID(customID).
		Label("Reload").
		Primary().
		Emoji("🔄").
		Build()
}

// ReloadButtonRow creates an action row with a reload button
func ReloadButtonRow(customID string) discordgo.ActionsRow {
	return SingleButtonRow(ReloadButton(customID))
}

// ConfirmCancelRow creates a row with confirm (green) and cancel (red) buttons
func ConfirmCancelRow(confirmID, cancelID string) discordgo.ActionsRow {
	return NewActionRow().
		AddButton(SuccessButton(confirmID, "Confirm")).
		AddButton(DangerButton(cancelID, "Cancel")).
		Build()
}

// YesNoRow creates a row with yes (green) and no (red) buttons
func YesNoRow(yesID, noID string) discordgo.ActionsRow {
	return NewActionRow().
		AddButton(SuccessButton(yesID, "Yes")).
		AddButton(DangerButton(noID, "No")).
		Build()
}

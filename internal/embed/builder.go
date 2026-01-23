package embed

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Builder provides a fluent interface for creating Discord embeds
type Builder struct {
	embed *discordgo.MessageEmbed
}

// New creates a new embed builder
func New() *Builder {
	return &Builder{
		embed: &discordgo.MessageEmbed{},
	}
}

// Title sets the embed title
func (b *Builder) Title(title string) *Builder {
	b.embed.Title = title
	return b
}

// Description sets the embed description
func (b *Builder) Description(description string) *Builder {
	b.embed.Description = description
	return b
}

// Color sets the embed color
func (b *Builder) Color(color int) *Builder {
	b.embed.Color = color
	return b
}

// URL sets the embed URL (makes title clickable)
func (b *Builder) URL(url string) *Builder {
	b.embed.URL = url
	return b
}

// Timestamp sets the embed timestamp to now
func (b *Builder) Timestamp() *Builder {
	b.embed.Timestamp = time.Now().Format(time.RFC3339)
	return b
}

// TimestampCustom sets a custom timestamp
func (b *Builder) TimestampCustom(t time.Time) *Builder {
	b.embed.Timestamp = t.Format(time.RFC3339)
	return b
}

// Footer sets the embed footer
func (b *Builder) Footer(text string, iconURL string) *Builder {
	b.embed.Footer = &discordgo.MessageEmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return b
}

// FooterText sets just the footer text
func (b *Builder) FooterText(text string) *Builder {
	b.embed.Footer = &discordgo.MessageEmbedFooter{
		Text: text,
	}
	return b
}

// Author sets the embed author
func (b *Builder) Author(name, url, iconURL string) *Builder {
	b.embed.Author = &discordgo.MessageEmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return b
}

// AuthorName sets just the author name
func (b *Builder) AuthorName(name string) *Builder {
	b.embed.Author = &discordgo.MessageEmbedAuthor{
		Name: name,
	}
	return b
}

// Thumbnail sets the embed thumbnail (small image on the right)
func (b *Builder) Thumbnail(url string) *Builder {
	b.embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: url,
	}
	return b
}

// Image sets the embed image (large image at the bottom)
func (b *Builder) Image(url string) *Builder {
	b.embed.Image = &discordgo.MessageEmbedImage{
		URL: url,
	}
	return b
}

// Field adds a field to the embed
func (b *Builder) Field(name, value string, inline bool) *Builder {
	b.embed.Fields = append(b.embed.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return b
}

// InlineField adds an inline field
func (b *Builder) InlineField(name, value string) *Builder {
	return b.Field(name, value, true)
}

// BlockField adds a non-inline field
func (b *Builder) BlockField(name, value string) *Builder {
	return b.Field(name, value, false)
}

// Fields adds multiple fields at once
func (b *Builder) Fields(fields ...*discordgo.MessageEmbedField) *Builder {
	b.embed.Fields = append(b.embed.Fields, fields...)
	return b
}

// Build returns the constructed embed
func (b *Builder) Build() *discordgo.MessageEmbed {
	return b.embed
}

// BuildSlice returns the embed as a slice (useful for InteractionResponse)
func (b *Builder) BuildSlice() []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{b.embed}
}

// ============================================
// Pre-built Template Embeds
// ============================================

// Success creates a success embed with green color
func Success(title, description string) *discordgo.MessageEmbed {
	return New().
		Title("✅ " + title).
		Description(description).
		Color(ColorSuccess).
		Timestamp().
		Build()
}

// Error creates an error embed with red color
func Error(title, description string) *discordgo.MessageEmbed {
	return New().
		Title("❌ " + title).
		Description(description).
		Color(ColorError).
		Timestamp().
		Build()
}

// Warning creates a warning embed with yellow color
func Warning(title, description string) *discordgo.MessageEmbed {
	return New().
		Title("⚠️ " + title).
		Description(description).
		Color(ColorWarning).
		Timestamp().
		Build()
}

// Info creates an info embed with blurple color
func Info(title, description string) *discordgo.MessageEmbed {
	return New().
		Title("ℹ️ " + title).
		Description(description).
		Color(ColorInfo).
		Timestamp().
		Build()
}

// Loading creates a loading embed
func Loading(message string) *discordgo.MessageEmbed {
	return New().
		Title("⏳ Loading...").
		Description(message).
		Color(ColorInfo).
		Build()
}

// ============================================
// Utility Functions
// ============================================

// CodeBlock wraps text in a Discord code block
func CodeBlock(language, code string) string {
	return fmt.Sprintf("```%s\n%s\n```", language, code)
}

// InlineCode wraps text in inline code
func InlineCode(text string) string {
	return fmt.Sprintf("`%s`", text)
}

// Bold makes text bold
func Bold(text string) string {
	return fmt.Sprintf("**%s**", text)
}

// Italic makes text italic
func Italic(text string) string {
	return fmt.Sprintf("*%s*", text)
}

// Underline makes text underlined
func Underline(text string) string {
	return fmt.Sprintf("__%s__", text)
}

// Strikethrough makes text strikethrough
func Strikethrough(text string) string {
	return fmt.Sprintf("~~%s~~", text)
}

// Spoiler wraps text in a spoiler
func Spoiler(text string) string {
	return fmt.Sprintf("||%s||", text)
}

// Quote makes text a block quote
func Quote(text string) string {
	return fmt.Sprintf("> %s", text)
}

// Mention creates a user mention
func Mention(userID string) string {
	return fmt.Sprintf("<@%s>", userID)
}

// MentionRole creates a role mention
func MentionRole(roleID string) string {
	return fmt.Sprintf("<@&%s>", roleID)
}

// MentionChannel creates a channel mention
func MentionChannel(channelID string) string {
	return fmt.Sprintf("<#%s>", channelID)
}

// Timestamp formats a time with Discord timestamp formatting
// Styles: t (short time), T (long time), d (short date), D (long date),
//         f (short datetime), F (long datetime), R (relative)
func Timestamp(t time.Time, style string) string {
	return fmt.Sprintf("<t:%d:%s>", t.Unix(), style)
}

// RelativeTime formats a time as relative (e.g., "2 hours ago")
func RelativeTime(t time.Time) string {
	return Timestamp(t, "R")
}

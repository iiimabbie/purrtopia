package component

import "github.com/bwmarrin/discordgo"

// ============================================
// Select Menu Builder
// ============================================

// SelectBuilder builds a select menu
type SelectBuilder struct {
	menu discordgo.SelectMenu
}

// NewSelect creates a new string select menu builder
func NewSelect() *SelectBuilder {
	return &SelectBuilder{
		menu: discordgo.SelectMenu{
			MenuType: discordgo.StringSelectMenu,
		},
	}
}

// CustomID sets the select menu's custom ID
func (s *SelectBuilder) CustomID(id string) *SelectBuilder {
	s.menu.CustomID = id
	return s
}

// Placeholder sets the placeholder text
func (s *SelectBuilder) Placeholder(text string) *SelectBuilder {
	s.menu.Placeholder = text
	return s
}

// MinValues sets the minimum number of selections (default 1)
func (s *SelectBuilder) MinValues(min int) *SelectBuilder {
	s.menu.MinValues = &min
	return s
}

// MaxValues sets the maximum number of selections (default 1)
func (s *SelectBuilder) MaxValues(max int) *SelectBuilder {
	s.menu.MaxValues = max
	return s
}

// Disabled sets the select menu as disabled
func (s *SelectBuilder) Disabled() *SelectBuilder {
	s.menu.Disabled = true
	return s
}

// AddOption adds an option to the select menu
func (s *SelectBuilder) AddOption(label, value, description string) *SelectBuilder {
	option := discordgo.SelectMenuOption{
		Label:       label,
		Value:       value,
		Description: description,
	}
	s.menu.Options = append(s.menu.Options, option)
	return s
}

// AddOptionWithEmoji adds an option with an emoji
func (s *SelectBuilder) AddOptionWithEmoji(label, value, description, emoji string) *SelectBuilder {
	option := discordgo.SelectMenuOption{
		Label:       label,
		Value:       value,
		Description: description,
		Emoji:       &discordgo.ComponentEmoji{Name: emoji},
	}
	s.menu.Options = append(s.menu.Options, option)
	return s
}

// AddOptionDefault adds a default selected option
func (s *SelectBuilder) AddOptionDefault(label, value, description string) *SelectBuilder {
	option := discordgo.SelectMenuOption{
		Label:       label,
		Value:       value,
		Description: description,
		Default:     true,
	}
	s.menu.Options = append(s.menu.Options, option)
	return s
}

// Build returns the select menu component
func (s *SelectBuilder) Build() discordgo.SelectMenu {
	return s.menu
}

// ============================================
// Auto-populated Select Menus
// ============================================

// NewUserSelect creates a user select menu
func NewUserSelect(customID string) *SelectBuilder {
	return &SelectBuilder{
		menu: discordgo.SelectMenu{
			MenuType: discordgo.UserSelectMenu,
			CustomID: customID,
		},
	}
}

// NewRoleSelect creates a role select menu
func NewRoleSelect(customID string) *SelectBuilder {
	return &SelectBuilder{
		menu: discordgo.SelectMenu{
			MenuType: discordgo.RoleSelectMenu,
			CustomID: customID,
		},
	}
}

// NewMentionableSelect creates a mentionable (user or role) select menu
func NewMentionableSelect(customID string) *SelectBuilder {
	return &SelectBuilder{
		menu: discordgo.SelectMenu{
			MenuType: discordgo.MentionableSelectMenu,
			CustomID: customID,
		},
	}
}

// NewChannelSelect creates a channel select menu
func NewChannelSelect(customID string) *SelectBuilder {
	return &SelectBuilder{
		menu: discordgo.SelectMenu{
			MenuType: discordgo.ChannelSelectMenu,
			CustomID: customID,
		},
	}
}

// ============================================
// Action Row with Select
// ============================================

// AddSelect adds a select menu to the action row
func (r *ActionRowBuilder) AddSelect(menu discordgo.SelectMenu) *ActionRowBuilder {
	r.components = append(r.components, menu)
	return r
}

// SelectRow creates an action row with a single select menu
func SelectRow(menu discordgo.SelectMenu) discordgo.ActionsRow {
	return NewActionRow().AddSelect(menu).Build()
}

// ============================================
// Quick Helpers
// ============================================

// StringSelect creates a simple string select menu
func StringSelect(customID, placeholder string, options ...SelectOption) discordgo.SelectMenu {
	builder := NewSelect().CustomID(customID).Placeholder(placeholder)
	for _, opt := range options {
		builder.AddOption(opt.Label, opt.Value, opt.Description)
	}
	return builder.Build()
}

// SelectOption represents a select menu option for quick helpers
type SelectOption struct {
	Label       string
	Value       string
	Description string
}

// Option creates a SelectOption for quick helpers
func Option(label, value, description string) SelectOption {
	return SelectOption{
		Label:       label,
		Value:       value,
		Description: description,
	}
}

// UserSelectRow creates an action row with a user select menu
func UserSelectRow(customID, placeholder string) discordgo.ActionsRow {
	menu := NewUserSelect(customID).Placeholder(placeholder).Build()
	return SelectRow(menu)
}

// RoleSelectRow creates an action row with a role select menu
func RoleSelectRow(customID, placeholder string) discordgo.ActionsRow {
	menu := NewRoleSelect(customID).Placeholder(placeholder).Build()
	return SelectRow(menu)
}

// ChannelSelectRow creates an action row with a channel select menu
func ChannelSelectRow(customID, placeholder string) discordgo.ActionsRow {
	menu := NewChannelSelect(customID).Placeholder(placeholder).Build()
	return SelectRow(menu)
}

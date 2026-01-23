package component

import "github.com/bwmarrin/discordgo"

// ============================================
// Text Input Builder
// ============================================

// TextInputBuilder builds a text input for modals
type TextInputBuilder struct {
	input discordgo.TextInput
}

// NewTextInput creates a new text input builder
func NewTextInput() *TextInputBuilder {
	return &TextInputBuilder{
		input: discordgo.TextInput{
			Style: discordgo.TextInputShort,
		},
	}
}

// CustomID sets the text input's custom ID
func (t *TextInputBuilder) CustomID(id string) *TextInputBuilder {
	t.input.CustomID = id
	return t
}

// Label sets the text input label
func (t *TextInputBuilder) Label(label string) *TextInputBuilder {
	t.input.Label = label
	return t
}

// Placeholder sets the placeholder text
func (t *TextInputBuilder) Placeholder(text string) *TextInputBuilder {
	t.input.Placeholder = text
	return t
}

// Value sets the default value
func (t *TextInputBuilder) Value(value string) *TextInputBuilder {
	t.input.Value = value
	return t
}

// Short sets the text input to single-line mode (default)
func (t *TextInputBuilder) Short() *TextInputBuilder {
	t.input.Style = discordgo.TextInputShort
	return t
}

// Paragraph sets the text input to multi-line mode
func (t *TextInputBuilder) Paragraph() *TextInputBuilder {
	t.input.Style = discordgo.TextInputParagraph
	return t
}

// Required sets the text input as required
func (t *TextInputBuilder) Required() *TextInputBuilder {
	t.input.Required = true
	return t
}

// Optional sets the text input as optional
func (t *TextInputBuilder) Optional() *TextInputBuilder {
	t.input.Required = false
	return t
}

// MinLength sets the minimum character length
func (t *TextInputBuilder) MinLength(min int) *TextInputBuilder {
	t.input.MinLength = min
	return t
}

// MaxLength sets the maximum character length
func (t *TextInputBuilder) MaxLength(max int) *TextInputBuilder {
	t.input.MaxLength = max
	return t
}

// Build returns the text input component
func (t *TextInputBuilder) Build() discordgo.TextInput {
	return t.input
}

// ============================================
// Modal Builder
// ============================================

// ModalBuilder builds a modal dialog
type ModalBuilder struct {
	customID   string
	title      string
	components []discordgo.MessageComponent
}

// NewModal creates a new modal builder
func NewModal() *ModalBuilder {
	return &ModalBuilder{
		components: make([]discordgo.MessageComponent, 0),
	}
}

// CustomID sets the modal's custom ID
func (m *ModalBuilder) CustomID(id string) *ModalBuilder {
	m.customID = id
	return m
}

// Title sets the modal title
func (m *ModalBuilder) Title(title string) *ModalBuilder {
	m.title = title
	return m
}

// AddTextInput adds a text input to the modal
func (m *ModalBuilder) AddTextInput(input discordgo.TextInput) *ModalBuilder {
	// Each text input must be in its own action row
	row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{input},
	}
	m.components = append(m.components, row)
	return m
}

// Build returns the modal response
func (m *ModalBuilder) Build() *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID:   m.customID,
			Title:      m.title,
			Components: m.components,
		},
	}
}

// ============================================
// Quick Helpers
// ============================================

// ShortInput creates a short (single-line) text input
func ShortInput(customID, label, placeholder string) discordgo.TextInput {
	return NewTextInput().
		CustomID(customID).
		Label(label).
		Placeholder(placeholder).
		Short().
		Required().
		Build()
}

// ParagraphInput creates a paragraph (multi-line) text input
func ParagraphInput(customID, label, placeholder string) discordgo.TextInput {
	return NewTextInput().
		CustomID(customID).
		Label(label).
		Placeholder(placeholder).
		Paragraph().
		Required().
		Build()
}

// OptionalShortInput creates an optional short text input
func OptionalShortInput(customID, label, placeholder string) discordgo.TextInput {
	return NewTextInput().
		CustomID(customID).
		Label(label).
		Placeholder(placeholder).
		Short().
		Optional().
		Build()
}

// OptionalParagraphInput creates an optional paragraph text input
func OptionalParagraphInput(customID, label, placeholder string) discordgo.TextInput {
	return NewTextInput().
		CustomID(customID).
		Label(label).
		Placeholder(placeholder).
		Paragraph().
		Optional().
		Build()
}

// SimpleModal creates a simple modal with one text input
func SimpleModal(customID, title, inputID, inputLabel, inputPlaceholder string) *discordgo.InteractionResponse {
	input := ShortInput(inputID, inputLabel, inputPlaceholder)
	return NewModal().
		CustomID(customID).
		Title(title).
		AddTextInput(input).
		Build()
}

// FeedbackModal creates a feedback modal with title and description inputs
func FeedbackModal(customID, title string) *discordgo.InteractionResponse {
	titleInput := ShortInput("feedback_title", "Title", "Enter a brief title...")
	descInput := ParagraphInput("feedback_desc", "Description", "Describe your feedback in detail...")

	return NewModal().
		CustomID(customID).
		Title(title).
		AddTextInput(titleInput).
		AddTextInput(descInput).
		Build()
}

// ============================================
// Modal Data Helpers
// ============================================

// GetModalValue extracts a value from modal submit data
func GetModalValue(data discordgo.ModalSubmitInteractionData, customID string) string {
	for _, row := range data.Components {
		if actionRow, ok := row.(*discordgo.ActionsRow); ok {
			for _, comp := range actionRow.Components {
				if input, ok := comp.(*discordgo.TextInput); ok {
					if input.CustomID == customID {
						return input.Value
					}
				}
			}
		}
	}
	return ""
}

package commands

import (
	"github.com/bwmarrin/discordgo"
)

// Handler is a function that handles an interaction
type Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)

// Command represents a slash command with its definition and handler
type Command struct {
	Definition *discordgo.ApplicationCommand
	Handler    Handler
}

// ============================================
// Auto-registration (使用 init() 自動註冊)
// ============================================

var registeredCommands []*Command
var componentHandlers = make(map[string]Handler)
var modalHandlers = make(map[string]Handler)

// RegisterCommand registers a slash command (call in init())
func RegisterCommand(definition *discordgo.ApplicationCommand, handler Handler) {
	registeredCommands = append(registeredCommands, &Command{
		Definition: definition,
		Handler:    handler,
	})
}

// RegisterComponent registers a component handler (call in init())
func RegisterComponent(customID string, handler Handler) {
	componentHandlers[customID] = handler
}

// RegisterModal registers a modal submit handler (call in init())
func RegisterModal(customID string, handler Handler) {
	modalHandlers[customID] = handler
}

// ============================================
// Getters (for bot.go)
// ============================================

// GetDefinitions returns all command definitions
func GetDefinitions() []*discordgo.ApplicationCommand {
	definitions := make([]*discordgo.ApplicationCommand, len(registeredCommands))
	for i, cmd := range registeredCommands {
		definitions[i] = cmd.Definition
	}
	return definitions
}

// GetHandlers returns a map of command names to handlers
func GetHandlers() map[string]Handler {
	handlers := make(map[string]Handler)
	for _, cmd := range registeredCommands {
		handlers[cmd.Definition.Name] = cmd.Handler
	}
	return handlers
}

// GetComponentHandlers returns all component handlers
func GetComponentHandlers() map[string]Handler {
	return componentHandlers
}

// GetModalHandlers returns all modal submit handlers
func GetModalHandlers() map[string]Handler {
	return modalHandlers
}

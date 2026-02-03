package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"purrtopia/internal/commands"
	"purrtopia/internal/config"

	// Import command subpackages to register their init() functions
	_ "purrtopia/internal/commands/gacha"
	_ "purrtopia/internal/commands/snow"

	"github.com/bwmarrin/discordgo"
)

// Bot represents the Discord bot instance
type Bot struct {
	session                 *discordgo.Session
	config                  *config.Config
	handlers                map[string]commands.Handler
	componentHandlers       map[string]commands.Handler
	componentPrefixHandlers map[string]commands.Handler
	modalHandlers           map[string]commands.Handler
}

// New creates a new bot instance
func New(cfg *config.Config) (*Bot, error) {
	// Create Discord session
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	bot := &Bot{
		session:                 session,
		config:                  cfg,
		handlers:                commands.GetHandlers(),
		componentHandlers:       commands.GetComponentHandlers(),
		componentPrefixHandlers: commands.GetComponentPrefixHandlers(),
		modalHandlers:           commands.GetModalHandlers(),
	}

	// Register event handlers
	bot.registerHandlers()

	return bot, nil
}

// registerHandlers registers all event handlers
func (b *Bot) registerHandlers() {
	// Ready event
	b.session.AddHandler(b.onReady)

	// Interaction (slash command) handler
	b.session.AddHandler(b.onInteraction)

	// Message handler for keyword-triggered responses
	b.session.AddHandler(b.onMessageCreate)
}

// onReady is called when the bot is ready
func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as %s#%s", r.User.Username, r.User.Discriminator)
	log.Printf("Connected to %d guilds", len(r.Guilds))

	// Set bot status
	err := s.UpdateGameStatus(0, "Heartopia💓")
	if err != nil {
		log.Printf("Failed to set status: %v", err)
	}
}

// onInteraction handles all interactions (commands, buttons, etc.)
func (b *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		// Slash commands
		commandName := i.ApplicationCommandData().Name
		if handler, ok := b.handlers[commandName]; ok {
			handler(s, i)
		} else {
			log.Printf("Unknown command: %s", commandName)
		}

	case discordgo.InteractionMessageComponent:
		// Buttons, Select Menus
		customID := i.MessageComponentData().CustomID
		if handler, ok := b.componentHandlers[customID]; ok {
			handler(s, i)
		} else {
			// Try prefix matching
			found := false
			for prefix, handler := range b.componentPrefixHandlers {
				if strings.HasPrefix(customID, prefix) {
					handler(s, i)
					found = true
					break
				}
			}
			if !found {
				log.Printf("Unknown component: %s", customID)
			}
		}

	case discordgo.InteractionModalSubmit:
		// Modal submissions
		customID := i.ModalSubmitData().CustomID
		if handler, ok := b.modalHandlers[customID]; ok {
			handler(s, i)
		} else {
			log.Printf("Unknown modal: %s", customID)
		}
	}
}

// Start starts the bot
func (b *Bot) Start() error {
	// Set intents (MessageContent is required to read message content)
	b.session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	// Open connection
	err := b.session.Open()
	if err != nil {
		return fmt.Errorf("failed to open Discord connection: %w", err)
	}

	log.Println("Registering commands...")

	// Register slash commands
	err = b.registerCommands()
	if err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	log.Println("Bot is now running!")
	log.Println("Press Ctrl+C to exit")

	return nil
}

// registerCommands registers all slash commands with Discord
// First clears all existing commands (global + guilds), then registers fresh
func (b *Bot) registerCommands() error {
	definitions := commands.GetDefinitions()

	// Step 1: Clear ALL existing commands first (clean slate)
	log.Println("Clearing all existing commands...")

	// Clear global commands
	globalCmds, err := b.session.ApplicationCommands(b.session.State.User.ID, "")
	if err == nil && len(globalCmds) > 0 {
		log.Printf("Clearing %d global commands...", len(globalCmds))
		_, err := b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, "", []*discordgo.ApplicationCommand{})
		if err != nil {
			log.Printf("Warning: failed to clear global commands: %v", err)
		}
	}

	// Clear commands from each configured guild
	for _, guildID := range b.config.GuildIDs {
		guildCmds, err := b.session.ApplicationCommands(b.session.State.User.ID, guildID)
		if err == nil && len(guildCmds) > 0 {
			log.Printf("Clearing %d commands from guild %s...", len(guildCmds), guildID)
			_, err := b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, guildID, []*discordgo.ApplicationCommand{})
			if err != nil {
				log.Printf("Warning: failed to clear commands for guild %s: %v", guildID, err)
			}
		}
	}

	log.Println("All commands cleared. Registering new commands...")

	// Step 2: Register commands
	// If no guild IDs specified, register global commands
	if len(b.config.GuildIDs) == 0 {
		registered, err := b.session.ApplicationCommandBulkOverwrite(
			b.session.State.User.ID,
			"", // Empty string = global commands
			definitions,
		)
		if err != nil {
			return fmt.Errorf("failed to register global commands: %w", err)
		}
		for _, cmd := range registered {
			log.Printf("Registered global command: /%s", cmd.Name)
		}
		return nil
	}

	// Register commands for each guild
	for _, guildID := range b.config.GuildIDs {
		registered, err := b.session.ApplicationCommandBulkOverwrite(
			b.session.State.User.ID,
			guildID,
			definitions,
		)
		if err != nil {
			log.Printf("Failed to register commands for guild %s: %v", guildID, err)
			continue
		}
		for _, cmd := range registered {
			log.Printf("Registered command: /%s (guild: %s)", cmd.Name, guildID)
		}
	}

	return nil
}

// Wait blocks until an interrupt signal is received
func (b *Bot) Wait() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// Stop gracefully shuts down the bot
func (b *Bot) Stop() error {
	log.Println("Shutting down...")

	// Optionally remove commands on shutdown (uncomment if desired)
	// b.removeCommands()

	return b.session.Close()
}

// removeCommands removes all registered commands (useful for cleanup)
func (b *Bot) removeCommands() {
	// If no guild IDs, remove global commands
	if len(b.config.GuildIDs) == 0 {
		b.removeCommandsFromGuild("")
		return
	}

	// Remove commands from each guild
	for _, guildID := range b.config.GuildIDs {
		b.removeCommandsFromGuild(guildID)
	}
}

func (b *Bot) removeCommandsFromGuild(guildID string) {
	cmds, err := b.session.ApplicationCommands(b.session.State.User.ID, guildID)
	if err != nil {
		log.Printf("Failed to get commands: %v", err)
		return
	}

	for _, cmd := range cmds {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, guildID, cmd.ID)
		if err != nil {
			log.Printf("Failed to delete command %s: %v", cmd.Name, err)
			continue
		}
		log.Printf("Removed command: /%s", cmd.Name)
	}
}

// Session returns the Discord session (for advanced usage)
func (b *Bot) Session() *discordgo.Session {
	return b.session
}

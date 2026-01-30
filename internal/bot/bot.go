package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"purrtopia/internal/commands"
	"purrtopia/internal/config"

	"github.com/bwmarrin/discordgo"
)

// Bot represents the Discord bot instance
type Bot struct {
	session           *discordgo.Session
	config            *config.Config
	handlers          map[string]commands.Handler
	componentHandlers map[string]commands.Handler
	modalHandlers     map[string]commands.Handler
}

// New creates a new bot instance
func New(cfg *config.Config) (*Bot, error) {
	// Create Discord session
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	bot := &Bot{
		session:           session,
		config:            cfg,
		handlers:          commands.GetHandlers(),
		componentHandlers: commands.GetComponentHandlers(),
		modalHandlers:     commands.GetModalHandlers(),
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
			log.Printf("Unknown component: %s", customID)
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
// Uses BulkOverwrite to sync commands (adds new, updates existing, removes deleted)
func (b *Bot) registerCommands() error {
	definitions := commands.GetDefinitions()

	// 原先沒有先清除舊的command, bot會留存一大堆已註冊過但已刪除的command, 這邊先做清除再重新註冊
	// BulkOverwrite will:
	// - Add new commands
	// - Update existing commands
	// - Remove commands that are no longer in the list
	registered, err := b.session.ApplicationCommandBulkOverwrite(
		b.session.State.User.ID,
		b.config.GuildID, // Empty string = global commands
		definitions,
	)
	if err != nil {
		return fmt.Errorf("failed to bulk overwrite commands: %w", err)
	}

	for _, cmd := range registered {
		log.Printf("Registered command: /%s", cmd.Name)
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
	cmds, err := b.session.ApplicationCommands(b.session.State.User.ID, b.config.GuildID)
	if err != nil {
		log.Printf("Failed to get commands: %v", err)
		return
	}

	for _, cmd := range cmds {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, b.config.GuildID, cmd.ID)
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

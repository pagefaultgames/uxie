package bot

import (
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/db"
)

// A Bot is a thin wrapper around a [tempest.HTTPClient] that can reference the help command database.
type Bot struct {
	// The underlying HTTP client.
	*tempest.HTTPClient

	// The help command database.
	store *db.Store

	// The guild ID of the server in which the bot is operating.
	guildId tempest.Snowflake
}

// NewBot creates a new Bot instance with the given parameters.
func NewBot(client *tempest.HTTPClient, guildId tempest.Snowflake) (b *Bot, err error) {
	store, err := db.Open()
	if err != nil {
		return nil, err
	}
	return &Bot{
		HTTPClient: client,
		store:      store,
		guildId:    guildId,
	}, nil
}

// defaultCommands is the list of bot commands registered by default.
var defaultCommands = []tempest.Command{
	addHelp,
}

func (b *Bot) logInfo() {
	slog.LogAttrs()
}

func (b *Bot) RegisterDefaultCommands() (err error) {
	for _, cmd := range defaultCommands {
		if err = b.RegisterCommand(cmd); err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Name, err)
		}
	}
	return nil
}

// AddHelpCommand adds a new help command to the bot's database.
func (b *Bot) AddHelpCommand(name, description, text string) error {
	if err := b.store.AddCommand(name, description, text); err != nil {
		return fmt.Errorf("failed to add help command to database: %w", err)
	}
	return nil
}

func (b *Bot) Close() error {
	return b.store.Close()
}

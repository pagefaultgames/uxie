package commands

import (
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/utils"
)

type modalHandler = func(mtx tempest.ModalInteraction)

// A Command is a basic struct wrapping tempest's Command type with additional metadata about how to handle any modals it creates.
type Command struct {
	tempest.Command

	// Any handlers this command has for its modals.
	// A nil or empty map signifies no modals.
	handlers map[string]modalHandler
}


func (c *Command) Register(client *tempest.HTTPClient) error {
	if err := client.RegisterCommand(c.Command); err != nil {
		utils.ErrorAttrs("Failed to register command", slog.String("command", c.Command.Name))
		return fmt.Errorf("Failed to register command: %w", err)
	}

	for cid, handler := range c.handlers {
		if err := client.RegisterModal(cid, handler); err != nil {
			utils.ErrorAttrs("Failed to register command modal",
				slog.String("command", c.Command.Name),
				slog.String("modal_id", cid))
			return fmt.Errorf("Failed to register command modal %s: %w", cid, err)
		}
	}
	return nil
}
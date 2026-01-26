package commands

import (
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/utils"
)

type modalHandler = func(mtx tempest.ModalInteraction)

// A command is a basic struct wrapping tempest's command type with additional metadata about how to handle any modals it creates.
type command struct {
	tempest.Command

	// Any handlers this command has for its modals, mapped to their respective custom IDs.
	// A nil or empty map signifies no modals.
	// Note that this must contain handlers for **all modals** used by this command, even ones created dynamically.
	modalHandlers map[string]modalHandler
}

func (c *command) Register(client *tempest.HTTPClient) error {
	if err := client.RegisterCommand(c.Command); err != nil {
		utils.ErrorAttrs("Failed to register command", slog.String("command", c.Name))
		return fmt.Errorf("failed to register command: %w", err)
	}

	for cid, handler := range c.modalHandlers {
		if err := client.RegisterModal(cid, handler); err != nil {
			utils.ErrorAttrs("Failed to register command modal",
				slog.String("command", c.Name),
				slog.String("modal_id", cid))
			return fmt.Errorf("failed to register command modal %s: %w", cid, err)
		}
	}
	return nil
}

package commands

import (
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/utils"
)

type modalHandler = func(mtx tempest.ModalInteraction)

// A command is a lightweight wrapper around [tempest.Command] with additional metadata to facilitate modal registration.
type command struct {
	tempest.Command

	// Any handlers this command has for its modals, mapped to their respective custom IDs.
	// A nil or empty map signifies no modals.
	// Note that this must contain handlers for **all modals** used by this command, even ones created dynamically.
	modalHandlers map[string]modalHandler
}

// Register registers the command and all of its modals to the given Tempest HTTP client.
func (c *command) Register(client *tempest.HTTPClient) error {
	if err := client.RegisterCommand(c.Command); err != nil {
		utils.ErrorAttrs("Failed to register command",
			slog.String("command", c.Name),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to register command %q: %w", c.Name, err)
	}

	for modalId, handler := range c.modalHandlers {
		if err := client.RegisterModal(modalId, handler); err != nil {
			utils.ErrorAttrs("Failed to register command modal",
				slog.String("command", c.Name),
				slog.String("modalId", modalId),
				slog.Any("error", err),
			)
			return fmt.Errorf(
				"failed to register modal %q for command %q: %w",
				modalId,
				c.Name,
				err,
			)
		}
	}
	return nil
}

package commands

import (
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/types"
	"github.com/pagefaultgames/oranguru/utils"
)

// getTextInputValue is a helper function to extract a text input modal's value from within a label.
// It returns the text contents, or an empty string if absent.
func getTextInputValue(
	itx *tempest.ModalInteraction,
	customId string,
) (contents string) {
	for _, comp := range itx.Data.Components {
		label, ok := comp.(tempest.LabelComponent)
		if !ok {
			continue
		}
		// NB: Golang type assertions produce zero values if type assertion fails
		c, found := label.Component.(tempest.TextInputComponent)
		if !found || c.CustomID != customId {
			continue
		}
		return c.Value
	}
	return ""
}

// getStringSelectValue is a helper function to extract a string select modal's value from within a label.
// It returns the chosen option, or an empty string if absent.
func getStringSelectValue(
	itx *tempest.ModalInteraction,
	customId string,
) (contents []string) {
	for _, comp := range itx.Data.Components {
		label, ok := comp.(tempest.LabelComponent)
		if !ok {
			continue
		}
		// NB: Golang type assertions produce zero values if type assertion fails
		c, found := label.Component.(tempest.StringSelectComponent)
		if !found || c.CustomID != customId {
			continue
		}
		return c.Values
	}
	return nil
}

// newMessageHandler creates a command handler that sends a predefined message.
func newMessageHandler(content string) func(itx *tempest.CommandInteraction) {
	return newFuncMessageHandler(func () string {return content})
}

// newFuncMessageHandler creates a command handler that sends a variable message.
func newFuncMessageHandler(
	f func() string,
) func(itx *tempest.CommandInteraction) {
	return func(itx *tempest.CommandInteraction) {
		content := f()
		if err := utils.SendDiscordMessage(itx.HTTPClient, itx.ChannelID, types.CreateMessageParams{
			Content: content,
		}, nil); err != nil {
			utils.ErrorAttrs("Failed to send message",
				slog.Any("error", err),
				slog.String("command", itx.Data.Name),
				slog.Uint64("userId", uint64(itx.Member.User.ID)),
				slog.String("user", itx.Member.User.Username),
			)
			return
		}

		// TODO: add message to the log?
		utils.InfoAttrs("Message sent successfully",
			slog.String("command", itx.Data.Name),
			slog.Uint64("userId", uint64(itx.Member.User.ID)),
			slog.String("user", itx.Member.User.Username),
		)
	}
}

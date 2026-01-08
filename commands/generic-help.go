package commands

import (
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/types"
	"github.com/pagefaultgames/oranguru/utils"
)

// NewMessageHandler creates a command handler that sends a predefined message.
func NewMessageHandler(content string) func(itx *tempest.CommandInteraction) {
	return func(itx *tempest.CommandInteraction) {
		utils.InfoAttrs("Sending message command",
			slog.String("command", itx.Data.Name),
			slog.String("user", itx.Member.User.Username),
		)

		if err := utils.SendDiscordMessage(itx.HTTPClient, itx.ChannelID, types.CreateMessageParams{
			Content: content,
		}); err != nil {
			utils.ErrorAttrs("Failed to send help message",
				slog.Any("error", err),
				slog.String("command", itx.Data.Name),
				slog.String("userId", itx.Member.User.ID),
				slog.String("user", itx.Member.User.Username),
			)
			return
		}
		utils.InfoAttrs("Help message sent successfully",
			slog.String("command", itx.Data.Name),
			slog.String("userId", itx.Member.User.ID),
			slog.String("user", itx.Member.User.Username),
		)
	}
}

var genericHelp = Command{
	Command: tempest.Command{
		Name:        "help",
		Description: "Get a list of available help commands, or enter one.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		SlashCommandHandler: ,
	},
}
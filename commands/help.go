package commands

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/types"
	"github.com/pagefaultgames/oranguru/utils"
)

// helpCommand is the slash command to show a help topic message.
var helpCommand = command{
	Command: tempest.Command{
		Name:        "help",
		Description: "Get help on available topics.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{{
			Type:         tempest.STRING_OPTION_TYPE,
			Name:         "topic",
			Description:  "The name of the topic to get help on.",
			Required:     false,
			AutoComplete: true,
		}},
		AutoCompleteHandler: helpTopicAutocompleteFunc("topic"),
		SlashCommandHandler: showHelpTopic,
	},
}

func showHelpTopic(ctx *tempest.CommandInteraction) {
	opt, found := utils.ValidateOptionValue[string](ctx, "topic")
	if !found {
		return
	}

	topic, err := db.GetHelpTopic(opt)
	if errors.Is(err, sql.ErrNoRows) {
		utils.InfoAttrs("Attempted to show nonexistent help topic",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", opt),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.String("command_name", ctx.Data.Name),
		)
		_, _ = ctx.SendLinearFollowUp(
			fmt.Sprintf("Error: No help topic found with name %s!", opt),
			true,
		)
		return
	} else if err != nil {
		utils.ErrorAttrs("Failed to retrieve help topic from database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("name", opt),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorFollowUp(
			ctx,
			fmt.Sprintf("Failed to retrieve help topic %s from database!", opt),
			err,
		)
		return
	}

	// Send the actual message
	if err := utils.SendMessage(ctx, types.CreateMessageParams{
		Content: fmt.Sprintf("**%s**\n\n%s", topic.Name, topic.Text),
	}); err != nil {
		utils.ErrorAttrs("Failed to send help topic message",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("name", opt),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)

		utils.SendErrorFollowUp(
			ctx,
			fmt.Sprintf("Failed to send help topic message %s!", opt),
			err,
		)
	}
}

package commands

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/utils"
)

// deleteTopic deletes a help topic from the database, if it exists.
var deleteTopic = command{
	Command: tempest.Command{
		Name:        "delete-topic",
		Description: "Delete a help topic from the database.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{{
			Type:         tempest.STRING_OPTION_TYPE,
			Name:         "topic",
			Description:  "The name of the help topic to delete.",
			Required:     true,
			MinLength:    1,
			MaxLength:    100,
			AutoComplete: true,
		}},
		AutoCompleteHandler: helpTopicAutocompleteFunc("topic"),
		SlashCommandHandler: handleDeleteTopic,
	},
}

func handleDeleteTopic(ctx *tempest.CommandInteraction) {
	topic, found := utils.ValidateOptionValue[string](ctx, "topic")
	if !found {
		return
	}

	err := db.DeleteTopic(topic)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		utils.InfoAttrs("Attempted to delete nonexistent help topic",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.String("commandName", ctx.Data.Name),
		)
		_ = ctx.SendLinearReply(
			fmt.Sprintf("Error: No help topic found with name %q!", topic),
			true,
		)
		return
	case err != nil:
		utils.ErrorAttrs("Failed to delete help topic from database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorMessage(
			ctx,
			"Failed to delete help topic "+topic+" from database!",
			err,
		)
		return
	default:
		utils.InfoAttrs("Successfully deleted help topic from database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.String("commandName", ctx.Data.Name),
		)
		_ = ctx.SendLinearReply(
			fmt.Sprintf("Successfully deleted help topic %q from the database."+
				"\nIt may take a few minutes to update.", topic),
			false)
	}
}

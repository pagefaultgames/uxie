package commands

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/types"
	"github.com/pagefaultgames/oranguru/utils"
)

// helpCommand is the slash command to show a help topic message.
var helpCommand = Command{
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
		AutoCompleteHandler: handleHelpAutocomplete,
		SlashCommandHandler: showHelpTopic,
	},
}

func handleHelpAutocomplete(ctx tempest.CommandInteraction) []tempest.CommandOptionChoice {
	_, f := ctx.GetFocusedValue()
	focusedText, ok := f.(string)
	if !ok {
		utils.ErrorAttrs("Invalid type for help topic autocomplete option",
			slog.String("username", ctx.BaseUser().Username),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.String("command_name", ctx.Data.Name),
			slog.Any("topic", f),
		)
		return nil
	}

	topics, err := db.GetAllTopics()
	if err != nil {
		utils.ErrorAttrs("error fetching topics from database for autocomplete",
			slog.String("username", ctx.BaseUser().Username),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.String("command_name", ctx.Data.Name),
			slog.Any("error", err),
		)
		return []tempest.CommandOptionChoice{}
	}

	choices := make([]tempest.CommandOptionChoice, 0, len(topics))
	for _, topic := range topics {
		if strings.HasPrefix(topic.Name, focusedText) {
			choices = append(choices, tempest.CommandOptionChoice{
				Name:  topic.Name,
				Value: topic.Name,
			})
		}
	}

	// Sort in shortest to longest
	slices.SortFunc(choices, func(i, j tempest.CommandOptionChoice) int {
		return len(j.Name) - len(i.Name)
	})

	return choices
}

func showHelpTopic(ctx *tempest.CommandInteraction) {
	opt, found := utils.ValidateOptionValue[string](ctx, "topic")
	if !found {
		return
	}

	topic, err := db.GetHelpTopic(opt)
	if errors.Is(err, sql.ErrNoRows) {
		_, _ = ctx.SendLinearFollowUp(fmt.Sprintf("No help topic found with name %s!", opt), true)
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
			fmt.Sprintf("Could not retrieve help topic %s from database!", opt),
			err,
		)
		return
	}

	// Send the actual message
	if err := utils.SendReplacementMessage(ctx, types.CreateMessageParams{
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
			fmt.Sprintf("Could not send help topic message %s!", opt),
			err,
		)
	}
}

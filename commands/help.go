package commands

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
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
			Required:     true,
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
		printNonexistentTopic(ctx, opt)
		return
	} else if err != nil {
		utils.ErrorAttrs("Failed to retrieve help topic from database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("name", opt),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorMessage(
			ctx,
			fmt.Sprintf("Failed to retrieve help topic %q from database!", opt),
			err,
		)
		return
	}

	// Send the actual message
	_ = ctx.SendReply(tempest.ResponseMessageData{
		Content:         formatHelpMessage(topic),
		AllowedMentions: &tempest.AllowedMentions{Parse: []tempest.AllowedMentionsType{}},
	}, true, nil)
}

func formatHelpMessage(topic db.HelpTopic) string {
	var b strings.Builder
	b.WriteString("**")
	b.WriteString(topic.Name)
	b.WriteString("**\n\n")
	b.WriteString(topic.Text)
	b.WriteRune('\n')
	return b.String()
}

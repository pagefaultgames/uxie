package commands

import (
	"log/slog"
	"strings"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/utils"
)

var getTopics = command{
	Command: tempest.Command{
		Name:        "get-topics",
		Description: "Get a list of all available help topics.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			printAllTopics(itx, true)
		},
	},
}

// printAllTopics fetches and prints all help topics in the database,
// either as a direct reply or a follow-up message (depending on the value of shouldRespond).
//
// This allows other commands to show the list of valid topics for invalid input.
func printAllTopics(ctx *tempest.CommandInteraction, shouldRespond bool) {
	var sendMsg func(msg string, ephemeral bool)
	if shouldRespond {
		sendMsg = func(msg string, ephemeral bool) {
			_ = ctx.SendLinearReply(msg, ephemeral)
		}
	} else {
		sendMsg = func(msg string, ephemeral bool) {
			_, _ = ctx.SendLinearFollowUp(msg, ephemeral)
		}
	}

	topics, err := db.GetAllTopics()
	if err != nil {
		utils.ErrorAttrs("Failled fetching topics from database",
			slog.String("username", ctx.BaseUser().Username),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.String("command_name", ctx.Data.Name),
			slog.Any("error", err),
		)
		sendMsg("Failed to fetch help topics from database!", true)
		return
	}

	if len(topics) == 0 {
		sendMsg("No help topics were found in the database.", true)
		return
	}

	var b strings.Builder
	b.WriteString("## Available help topics:\n")
	for _, topic := range topics {
		b.WriteString("- `" + topic.Name + "`\n")
	}

	sendMsg(b.String(), true)
}

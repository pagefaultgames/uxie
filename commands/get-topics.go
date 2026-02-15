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
		Name:                "get-topics",
		Description:         "Get a list of all available help topics.",
		Type:                tempest.CHAT_INPUT_COMMAND_TYPE,
		SlashCommandHandler: handleGetTopics,
	},
}

func handleGetTopics(ctx *tempest.CommandInteraction) {
	topics, err := db.GetAllTopics()
	if err != nil {
		utils.ErrorAttrs("Failled fetching topics from database",
			slog.String("username", ctx.BaseUser().Username),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.String("command_name", ctx.Data.Name),
			slog.Any("error", err),
		)
		utils.SendErrorFollowUp(ctx, "Failed to fetch help topics from database!", err)
		return
	}

	if len(topics) == 0 {
		_, _ = ctx.SendLinearFollowUp("No help topics were found in the database.", true)
		return
	}

	var b strings.Builder
	b.WriteString("## Available help topics:\n")
	for _, topic := range topics {
		b.WriteString("- `" + topic.Name + "`\n")
	}

	_, _ = ctx.SendLinearFollowUp(b.String(), true)
}

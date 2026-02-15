package commands

import (
	"github.com/amatsagu/tempest"
)

var getTopics = command{
	Command: tempest.Command{
		Name:        "get-topics",
		Description: "Get a list of all available help topics.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		SlashCommandHandler: func(ctx *tempest.CommandInteraction) {
			_ = ctx.SendLinearReply(getAllTopicText(), false)
		},
	},
}

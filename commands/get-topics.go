package commands

import (
	"github.com/amatsagu/tempest"
)

var getTopics = command{
	Command: tempest.Command{
		Name:        "get-topics",
		Description: "Get a list of all available help topics.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{{
			Type:        tempest.BOOLEAN_OPTION_TYPE,
			Required:    false,
			Name:        "ephemeral",
			Description: "Whether the reply should be ephemeral (only visible to you, default true)",
		}},
		SlashCommandHandler: func(ctx *tempest.CommandInteraction) {
			ephemeral := true
			if len(ctx.Data.Options) > 0 {
				if v, ok := ctx.Data.Options[0].Value.(bool); ok {
					ephemeral = v
				}
			}

			text, _ := getAllTopicText()
			_ = ctx.SendLinearReply("## Available help topics:\n"+text, ephemeral)
		},
	},
}

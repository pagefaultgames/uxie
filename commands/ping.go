package commands

import (
	"github.com/amatsagu/tempest"
)

var pingCommand = command{
	Command: tempest.Command{
		Name:                "ping",
		Description:         "Check if the bot is alive",
		RequiredPermissions: tempest.ADMINISTRATOR_PERMISSION_FLAG,
		Options: []tempest.CommandOption{{
			Type:        tempest.BOOLEAN_OPTION_TYPE,
			Required:    false,
			Name:        "ephemeral",
			Description: "Whether the reply should be ephemeral (only visible to you, default false)",
		}},
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			ephemeral := false
			if len(itx.Data.Options) > 0 {
				if v, ok := itx.Data.Options[0].Value.(bool); ok {
					ephemeral = v
				}
			}
			_ = itx.SendLinearReply("I'm still alive!", ephemeral)
		},
	},
}

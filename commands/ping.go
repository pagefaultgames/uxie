// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

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
			Description: "Whether the reply should be ephemeral (only visible to you, default true)",
		}},
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			ephemeral := true
			if len(itx.Data.Options) > 0 {
				if v, ok := itx.Data.Options[0].Value.(bool); ok {
					ephemeral = v
				}
			}
			_ = itx.SendLinearReply("I'm still alive!", ephemeral)
		},
	},
}

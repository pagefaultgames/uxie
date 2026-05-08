// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/utils"
)

var getTopics = command{
	Command: tempest.Command{
		Name:        "get-topics",
		Description: "Get a list of all available help topics.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{
			{
				Type:        tempest.BOOLEAN_OPTION_TYPE,
				Required:    false,
				Name:        "ephemeral",
				Description: "Whether the reply should be ephemeral (only visible to you, default true)",
			},
			{
				Type:        tempest.BOOLEAN_OPTION_TYPE,
				Required:    false,
				Name:        "include-aliases",
				Description: "Whether to include aliases in the list of topics (default false)",
			},
		},
		SlashCommandHandler: func(ctx *tempest.CommandInteraction) {
			ephemeral := true
			if len(ctx.Data.Options) > 0 {
				if v, ok := ctx.Data.Options[0].Value.(bool); ok {
					ephemeral = v
				}
			}

			includeAliases := false
			if len(ctx.Data.Options) > 1 {
				if v, ok := ctx.Data.Options[1].Value.(bool); ok {
					includeAliases = v
				}
			}

			text, err := getAllTopicText(includeAliases)
			if err != nil {
				utils.ErrorAttrs("Failed to retrieve help topics from database",
					slog.String("username", ctx.BaseUser().Username),
					slog.Uint64("ID", uint64(ctx.ID)),
					slog.Any("error", err),
				)
				utils.SendErrorMessage(
					ctx,
					"Failed to retrieve available help topics from database!",
					err,
				)
				return
			}

			header := "## Available help topics:\n"
			if includeAliases {
				header = "## Available help topics and aliases:\n"
			}
			_ = ctx.SendLinearReply(header+text, ephemeral)
		},
	},
}

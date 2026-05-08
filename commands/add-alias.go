// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
)

var addAlias = command{
	Command: tempest.Command{
		Name:        "add-alias",
		Description: "Add an alias for an existing help topic.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{
			{
				Type:         tempest.STRING_OPTION_TYPE,
				Name:         "topic",
				Description:  "The existing help topic to which this alias should resolve.",
				Required:     true,
				MinLength:    1,
				MaxLength:    MAX_TOPIC_NAME_LENGTH,
				AutoComplete: true,
			},
			{
				Type:        tempest.STRING_OPTION_TYPE,
				Name:        "alias",
				Description: "The alternate name to add for the given help topic.",
				Required:    true,
				MinLength:   1,
				MaxLength:   MAX_TOPIC_NAME_LENGTH,
			},
		},
		AutoCompleteHandler: helpTopicAutocompleteFunc("topic", false),
		SlashCommandHandler: handleAddAlias,
	},
}

func handleAddAlias(ctx *tempest.CommandInteraction) {
	topicName, found := utils.ValidateOptionValue[string](ctx, "topic")
	if !found {
		return
	}

	alias, found := utils.ValidateOptionValue[string](ctx, "alias")
	if !found {
		return
	}

	if errMsg := checkTopicValidity(alias); errMsg != "" {
		_ = ctx.SendLinearReply(errMsg, true)
		return
	}

	err := db.AddAlias(topicName, alias)
	var duplicateAlias db.ErrDuplicateAlias

	switch {
	case errors.As(err, &duplicateAlias):
		utils.InfoAttrs("Failed to add help topic alias due to duplicate name",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topicName),
			slog.String("alias", alias),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.String("otherAliasTarget", duplicateAlias.OtherAliasTarget),
		)

		var errMsg string
		if duplicateAlias.OtherAliasTarget == "" {
			errMsg = fmt.Sprintf(
				"⚠️ Cannot add alias `%s` for topic `%s`: a topic with the same name already exists!",
				alias,
				topicName,
			)
		} else {
			errMsg = fmt.Sprintf(
				"⚠️ Cannot add alias `%s` for topic `%s`: an alias for topic `%s` with the same name already exists!",
				alias,
				topicName,
				duplicateAlias.OtherAliasTarget,
			)
		}
		_ = ctx.SendLinearReply(errMsg, true)
		return
	case err != nil:
		utils.ErrorAttrs("Failed to add help topic alias to database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topicName),
			slog.String("alias", alias),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorMessage(
			ctx,
			fmt.Sprintf("Failed to add alias `%s` for help topic `%s`!", alias, topicName),
			err,
		)
		return
	default:
		utils.InfoAttrs("Successfully added help topic alias",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topicName),
			slog.String("alias", alias),
			slog.Uint64("ID", uint64(ctx.ID)),
		)
		_ = ctx.SendLinearReply(
			fmt.Sprintf("Successfully added alias `%s` for help topic `%s`!", alias, topicName),
			true,
		)
	}
}

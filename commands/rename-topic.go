// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
)

// renameTopic renames a help topic in the database, if it exists.
var renameTopic = command{
	Command: tempest.Command{
		Name:        "rename-topic",
		Description: "Rename a help topic in the database.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{
			{
				Type:         tempest.STRING_OPTION_TYPE,
				Name:         "old-name",
				Description:  "The name of the help topic to rename.",
				Required:     true,
				MinLength:    1,
				MaxLength:    100,
				AutoComplete: true,
			},
			{
				Type:         tempest.STRING_OPTION_TYPE,
				Name:         "new-name",
				Description:  "The new name for the help topic.",
				Required:     true,
				MinLength:    1,
				MaxLength:    100,
				AutoComplete: false,
			},
		},
		AutoCompleteHandler: helpTopicAutocompleteFunc("old-name", false),
		SlashCommandHandler: handleRenameTopic,
	},
}

func handleRenameTopic(ctx *tempest.CommandInteraction) {
	topic, found := utils.ValidateOptionValue[string](ctx, "old-name")
	if !found {
		return
	}

	newName, found := utils.ValidateOptionValue[string](ctx, "new-name")
	if !found {
		return
	}

	err := db.RenameTopic(topic, newName)
	dupe := &db.ErrDuplicateRename{}
	switch {
	case errors.As(err, &dupe):
		utils.InfoAttrs("Name conflict occurred when renaming topic",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topic),
			slog.String("new_name", newName),
			slog.Uint64("ID", uint64(ctx.ID)),
		)
		topicType := "topic"
		if dupe.Alias {
			topicType = "alias"
		}
		utils.SendErrorMessage(
			ctx,
			fmt.Sprintf(
				"Cannot rename topic `%s` to `%s`: the new name conflicts with an existing %s!",
				topic,
				newName,
				topicType,
			),
			err,
		)
		return
	case errors.Is(err, sql.ErrNoRows):
		printNonexistentTopic(ctx, topic)
		return
	case err != nil:
		utils.ErrorAttrs("Failed to rename help topic in database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorMessage(
			ctx,
			"Failed to rename help topic "+topic+" in database!",
			err,
		)
		return
	default:
		utils.InfoAttrs("Successfully renamed help topic in database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topic),
			slog.String("renamed_topic", newName),
			slog.Uint64("ID", uint64(ctx.ID)),
		)
		_ = ctx.SendLinearReply(
			fmt.Sprintf("Successfully renamed help topic `%s` to `%s`."+
				"\nIt may take a few minutes to update.", topic, newName),
			true)
	}
}

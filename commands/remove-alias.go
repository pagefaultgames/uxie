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

var removeAlias = command{
	Command: tempest.Command{
		Name:        "remove-alias",
		Description: "Remove an existing help topic alias.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{{
			Type:         tempest.STRING_OPTION_TYPE,
			Name:         "alias",
			Description:  "The alias to remove.",
			Required:     true,
			MinLength:    1,
			MaxLength:    MAX_TOPIC_NAME_LENGTH,
			AutoComplete: true,
		}},
		AutoCompleteHandler: aliasAutocompleteFunc("alias"),
		SlashCommandHandler: handleRemoveAlias,
	},
}

func handleRemoveAlias(ctx *tempest.CommandInteraction) {
	alias, found := utils.ValidateOptionValue[string](ctx, "alias")
	if !found {
		return
	}

	err := db.DeleteAlias(alias)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_ = ctx.SendLinearReply(
			fmt.Sprintf("⚠️ No alias named `%s` exists!", alias),
			true,
		)
		return
	case err != nil:
		utils.ErrorAttrs("Failed to delete help topic alias from database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("alias", alias),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorMessage(
			ctx,
			fmt.Sprintf("Failed to delete alias `%s` from the database!", alias),
			err,
		)
		return
	default:
		utils.InfoAttrs("Successfully deleted help topic alias from database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("alias", alias),
			slog.Uint64("ID", uint64(ctx.ID)),
		)
		_ = ctx.SendLinearReply(
			fmt.Sprintf("Successfully removed alias `%s` from the database."+
				"\nIt may take a few minutes to update.", alias),
			true,
		)
	}
}

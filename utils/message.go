/*
 * SPDX-FileCopyrightText: 2025 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package utils

import (
	"errors"
	"net/http"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/types"
)

var ErrMissingRequiredField = errors.New(
	"at least one of content, embeds, components, or files to be present",
)

// A replacement for [tempest.SendMessage] that accepts [types.CreateMessageParams] instead of [tempest.Message]
// Necessary, as tempest does not include support for fields like AllowedMentions.
// At the moment, does not support files.
func SendDiscordMessage(
	client *tempest.HTTPClient,
	channelID tempest.Snowflake,
	message types.CreateMessageParams,
	files []tempest.File,
) error {
	// Discord requires at least one of content, embeds, sticker_ids, components, files[n], or poll to be present.
	if message.Content == "" &&
		len(message.Embeds) == 0 &&
		len(message.Components) == 0 &&
		len(files) == 0 {
		return ErrMissingRequiredField
	}

	_, err := client.Rest.RequestWithFiles(
		http.MethodPost,
		"/channels/"+channelID.String()+"/messages",
		message,
		files,
	)
	return err
}

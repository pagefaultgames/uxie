/*
 * SPDX-FileCopyrightText: 2025 Pagefault Games
 * SPDX-FileContributor: SirzBenjie
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package types

import "github.com/amatsagu/tempest"

// https://discord.com/developers/docs/resources/message#create-message-jsonform-params
// Parameters currently unused by Oranguru are commented out for faster JSON parsing
type CreateMessageParams struct {
	Content          string                    `json:"content,omitempty"`           // The message contents (up to 2000 characters)
	Nonce            string                    `json:"nonce,omitempty"`             // A nonce that can be used for optimistic message sending (up to 25 characters)
	TTS              bool                      `json:"tts,omitempty"`               // Whether this is a TTS message; default `false`
	Embeds           []tempest.Embed           `json:"embeds,omitzero"`             // Up to 10 rich embeds (up to 6000 characters)
	AllowedMentions  *tempest.AllowedMentions  `json:"allowed_mentions,omitempty"`  // Allowed mentions for the message
	MessageReference *tempest.MessageReference `json:"message_reference,omitempty"` // Reference data for crossposted messages
	Components       []tempest.LayoutComponent `json:"components,omitzero"`         // The components to include with the message
	StickerIds       []tempest.Snowflake       `json:"sticker_ids,omitempty"`       // The ids of up to 3 stickers in the server to send in the message
	PayloadJSON      string                    `json:"payload_json,omitempty"`      // JSON encoded body of non-file params
	Attachments      []tempest.Attachment      `json:"attachments,omitzero"`        // Attachment objects with filename and description
	Flags            tempest.MessageFlags      `json:"flags,omitempty"`             // Message flags combined as a bitfield (only `SUPPRESS_EMBEDS`, `SUPPRESS_NOTIFICATIONS`, `IS_VOICE_MESSAGE`, and `IS_COMPONENTS_V2` can be set)
	EnforceNonce     bool                      `json:"enforce_nonce,omitempty"`     // Whether to enforce the uniqueness of a provided nonce; default `false`. If another message was created by the same author with the same nonce, that message will be returned and no new message will be created.
	Poll             *tempest.Poll             `json:"poll,omitempty"`              // A poll!
	Files            []tempest.File            `json:"files,omitempty"`             // The contents of the files to send with the message
}

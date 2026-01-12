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
	Content          string                    `json:"content,omitempty"`           // the message contents (up to 2000 characters)
	Embeds           []tempest.Embed           `json:"embeds,omitzero"`             // Up to 10 rich embeds (up to 6000 characters)
	AllowedMentions  *tempest.AllowedMentions  `json:"allowed_mentions,omitempty"`  // allowed mentions for the message
	Components       []tempest.LayoutComponent `json:"components,omitzero"`         // the components to include with the message
	Attachments      []tempest.Attachment      `json:"attachments,omitzero"`        // attachment objects with filename and description
	Flags            tempest.MessageFlags      `json:"flags,omitempty"`             // message flags combined as a bitfield (only `SUPPRESS_EMBEDS`, `SUPPRESS_NOTIFICATIONS`, `IS_VOICE_MESSAGE`, and `IS_COMPONENTS_V2` can be set)
	MessageReference *tempest.MessageReference `json:"message_reference,omitempty"` // reference data for crossposted messages
	// TTS             bool                     `json:"tts,omitempty"`               // true if this is a TTS message
	// Nonce           string                    `json:"nonce,omitempty"`            // a nonce that can be used for optimistic message sending (up to 25 characters)
	// StickerIds      []tempest.Snowflake       `json:"sticker_ids,omitempty"`      // the ids of up to 3 stickers in the server to send in the message
	// PayloadJSON     string                    `json:"payload_json,omitempty"`     // JSON encoded body of non-file params
	// EnforceNonce    bool                      `json:"enforce_nonce,omitempty"`    // whether to enforce the nonce (defaults to false)
	// Poll            *tempest.Poll             `json:"poll,omitempty"`             // A poll!
}

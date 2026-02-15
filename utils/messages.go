package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/types"
)

// ValidateOptionValue is a helper function to extract and validate a command option's value type.
// It returns the value (which will be zero if not found) and whether the value was found successfully.
//
// This function performs requisite logging and error messaging on failure, so callers can simply return immediately
// if found is false.
func ValidateOptionValue[T string | bool | float64](
	ctx *tempest.CommandInteraction,
	optName string,
) (topic T, found bool) {
	opt, found := ctx.GetOptionValue(optName)
	if !found {
		return topic, false
	}

	topic, found = opt.(T)
	if !found {
		ErrorAttrs(
			"Invalid input type for command option",
			slog.String("username", ctx.BaseUser().Username),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.String("option_name", optName),
			slog.String("type", fmt.Sprintf("%T", opt)),
			slog.String("expected_type", fmt.Sprintf("%T", topic)),
			slog.String("command_name", ctx.Data.Name),
		)
		_, _ = ctx.SendLinearFollowUp(
			fmt.Sprintf(
				"Option %s was of incorrect type!\nExpected a value of type %T, but received %v (type %T)!",
				optName,
				topic,
				opt,
				opt,
			),
			true,
		)
	}
	return topic, found
}

// anyInteraction is an interface that abstracts over both CommandInteraction and ModalInteraction
// (both of which can send follow up messages).
type anyInteraction interface {
	SendFollowUp(content tempest.ResponseMessageData, ephemeral bool) (tempest.Message, error)
	SendLinearFollowUp(content string, ephemeral bool) (tempest.Message, error)
}

// ensure both CommandInteraction and ModalInteraction satisfy the interface.
// This is removed from the compiled binary, so no efficiency is lost.
var (
	_ anyInteraction = (*tempest.CommandInteraction)(nil)
	_ anyInteraction = (*tempest.ModalInteraction)(nil)
)

// DeleteReplyIfExists deletes an existing reply to the current message, if one exists.
// It returns whether the deletion was successful (which will be true if no reply existed in the first place).
func DeleteReplyIfExists(ctx *tempest.CommandInteraction) (ok bool) {
	if ctx.Data.TargetID == 0 {
		return true
	}
	if err := ctx.DeleteReply(); err != nil {
		ErrorAttrs("Failed to delete existing reply",
			slog.String("username", ctx.BaseUser().Username),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		SendErrorFollowUp(ctx, "Failed to delete existing reply!", err)
		return false
	}

	return true
}

// SendErrorFollowUp is a helper function to send a standardized error follow-up message.
func SendErrorFollowUp[T anyInteraction](ctx T, msg string, err error) {
	_, _ = ctx.SendLinearFollowUp(msg+"\nError:\n```\n"+err.Error()+"\n```", true)
}

var ErrMissingRequiredField = errors.New(
	"expected at least one of content, embeds, components, or files to be present",
)

// SendMessage is a helper method to create and send a new message.
// It returns the first error encountered.
//
// Note that several fields of the message object will be altered, as follows:
//   - AllowedMentions will be set to prevent any mentions from pinging anyone.
//   - MessageReference will be set to match the original message's reference, if it was a reply.
//   - Flags will have SUPPRESS_NOTIFICATIONS set.
func SendMessage(
	ctx *tempest.CommandInteraction,
	message types.CreateMessageParams,
) (err error) {
	// Discord requires at least one of content, embeds, sticker_ids, components, files, or poll to be present.
	if message.Content == "" && len(message.Embeds) == 0 && len(message.Components) == 0 {
		return ErrMissingRequiredField
	}

	orig, found := ctx.Data.Resolved.Messages[ctx.Data.TargetID]
	if found {
		if err := ctx.DeleteReply(); err != nil {
			return fmt.Errorf("failed to delete original message: %w", err)
		}
	}

	message.AllowedMentions = &tempest.AllowedMentions{
		Parse: []tempest.AllowedMentionsType{},
	}
	message.MessageReference = orig.MessageReference
	message.Flags |= tempest.SUPPRESS_NOTIFICATIONS_MESSAGE_FLAG

	DeleteReplyIfExists(ctx)

	if _, err := ctx.BaseClient.Rest.Request(
		http.MethodPost,
		"/channels/"+ctx.ChannelID.String()+"/messages",
		message,
	); err != nil {
		return fmt.Errorf("failed to send replacement message: %w", err)
	}
	return nil
}

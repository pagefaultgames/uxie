package commands

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/types"
	"github.com/pagefaultgames/oranguru/utils"
)

// getTextInputValue is a helper function to extract a text input modal's value from within a label.
// It returns the text contents, or an empty string if absent.
func getTextInputValue(
	mtx *tempest.ModalInteraction,
	customId string,
) (contents string) {
	for _, comp := range mtx.Data.Components {
		label, ok := comp.(tempest.LabelComponent)
		if !ok {
			continue
		}

		c, found := label.Component.(tempest.TextInputComponent)
		if !found || c.CustomID != customId {
			continue
		}
		return c.Value
	}
	return ""
}

// validateOptionValue is a helper function to extract and validate a command option's value type.
// It returns the value and whether it was found and of the correct type.
// This function performs necessary logging and error messaging on failure, so callers can simply return
// if ok is false.
func validateOptionValue[T string | bool | float64](
	ctx *tempest.CommandInteraction,
	optName string,
) (topic T, ok bool) {
	opt, found := ctx.GetOptionValue(optName)
	if !found {
		return topic, false
	}

	topic, ok = opt.(T)
	if !ok {
		utils.ErrorAttrs(
			fmt.Sprintf("Invalid input type for %s command option %s!", ctx.Data.Name, optName),
			slog.String("username", ctx.User.Username),
			slog.String("type", fmt.Sprintf("%T", opt)),
			slog.Uint64("ID", uint64(ctx.ID)),
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
	return topic, ok
}

type canSendLinearFollowUp interface {
	SendLinearFollowUp(content string, ephemeral bool) (tempest.Message, error)
}

// ensure both CommandInteraction and ModalInteraction satisfy the interface.
// This is removed from the compiled binary, so no efficiency is lost.
var (
	_ canSendLinearFollowUp = (*tempest.CommandInteraction)(nil)
	_ canSendLinearFollowUp = (*tempest.ModalInteraction)(nil)
)

// sendErrorFollowUp is a helper function to send a standardized error follow-up message.
func sendErrorFollowUp[T canSendLinearFollowUp](ctx T, msg string, err error) {
	_, _ = ctx.SendLinearFollowUp(msg+"\nError:\n```\n"+err.Error()+"\n```", true)
}

var (
	errNoTargetMsg          = errors.New("could not find target message in resolved data")
	ErrMissingRequiredField = errors.New(
		"expected at least one of content, embeds, components, or files to be present",
	)
)

// sendReplacementMessage deletes the original target message used for a slash command
// and sends a new message in its place with the given content.
// It returns the first error encountered.
//
// Note that several fields of the message object will be altered, as follows:
//   - AllowedMentions will be set to prevent any mentions from pinging anyone.
//   - MessageReference will be set to match the original message's reference.
//   - Flags will have SUPPRESS_NOTIFICATIONS set.
func sendReplacementMessage(
	ctx *tempest.CommandInteraction,
	message types.CreateMessageParams,
) (err error) {
	// Discord requires at least one of content, embeds, sticker_ids, components, files, or poll to be present.
	if message.Content == "" && len(message.Embeds) == 0 && len(message.Components) == 0 {
		return ErrMissingRequiredField
	}

	orig, found := ctx.Data.Resolved.Messages[ctx.Data.TargetID]
	if !found {
		return errNoTargetMsg
	}

	message.AllowedMentions = &tempest.AllowedMentions{
		Parse: []tempest.AllowedMentionsType{},
	}
	message.MessageReference = orig.MessageReference
	message.Flags |= tempest.SUPPRESS_NOTIFICATIONS_MESSAGE_FLAG

	if err := ctx.DeleteReply(); err != nil {
		return fmt.Errorf("failed to delete original message: %w", err)
	}

	if _, err := ctx.BaseClient.Rest.Request(
		http.MethodPost,
		"/channels/"+ctx.ChannelID.String()+"/messages",
		message,
	); err != nil {
		return fmt.Errorf("failed to send replacement message: %w", err)
	}
	return nil
}

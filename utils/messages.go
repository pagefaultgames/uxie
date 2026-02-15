package utils

import (
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
)

// SendReplyOrFollowUp sends a message as either a reply or a follow-up, depending on whether the interaction has already been responded to.
func SendReplyOrFollowUp[T anyInteraction](ctx T, msg string, ephemeral bool) (err error) {
	if ctx.Responded() {
		_, err = ctx.SendLinearFollowUp(msg, ephemeral)
	} else {
		switch interaction := any(ctx).(type) {
		case *tempest.CommandInteraction:
			err = interaction.SendLinearReply(msg, ephemeral)
		case *tempest.ModalInteraction:
			err = interaction.AcknowledgeWithLinearMessage(msg, ephemeral)
		default:
			panic(fmt.Sprintf("unexpected interaction type %T", ctx))
		}
	}

	if err != nil {
		ErrorAttrs("Failed to send reply message",
			slog.String("username", ctx.BaseUser().Username),
			slog.Any("sendError", err),
		)
	}
	return err
}

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
			slog.String("optionName", optName),
			slog.String("type", fmt.Sprintf("%T", opt)),
			slog.String("expected_type", fmt.Sprintf("%T", topic)),
			slog.String("commandName", ctx.Data.Name),
		)
		_ = SendReplyOrFollowUp(ctx,
			fmt.Sprintf(
				"Option %q was of incorrect type!\nExpected a value of type %T, but received %v (type %T)!",
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

// SendErrorMessage is a convenience wrapper around [GetErrorMessage] and [SendReplyOrFollowUp].
func SendErrorMessage[T anyInteraction](ctx T, msg string, err error) {
	_ = SendReplyOrFollowUp(ctx, GetErrorMessage(msg, err), true)
}

// GetErrorMessage formats an error message with a consistent style suitable for end user presentation.
func GetErrorMessage(msg string, err error) string {
	return fmt.Sprintf("%s\nError:\n```\n%s\n```", msg, err.Error())
}

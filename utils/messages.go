package utils

import (
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
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
	BaseUser() *tempest.User
	Responded() bool
}

// ensure both CommandInteraction and ModalInteraction satisfy the interface.
// This is removed from the compiled binary, so no efficiency is lost.
var (
	_ anyInteraction = (*tempest.CommandInteraction)(nil)
	_ anyInteraction = (*tempest.ModalInteraction)(nil)
)

// SendErrorMessage is a helper function to send a standardized error follow-up message.
//
// If the provided interaction has yet to provide a response to the end user, the error message will be sent as a reply/acknowledgement to avoid timeouts.
func SendErrorMessage[T anyInteraction](ctx T, msg string, err error) {
	msg += "\nError:\n```\n" + err.Error() + "\n```"

	var sendErr error
	if ctx.Responded() {
		_, sendErr = ctx.SendLinearFollowUp(msg, true)
	} else {
		switch interaction := any(ctx).(type) {
		case *tempest.CommandInteraction:
			sendErr = interaction.SendLinearReply(msg, true)
		case *tempest.ModalInteraction:
			sendErr = interaction.AcknowledgeWithLinearMessage(msg, true)
		}
	}

	if sendErr != nil {
		ErrorAttrs("Failed to send error follow-up message",
			slog.String("username", ctx.BaseUser().Username),
			slog.Any("send_error", sendErr),
		)
	}
}

// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"fmt"
	"log/slog"
	"time"

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

// SendErrorMessage is a convenience wrapper around [GenericErrorMessage] and [SendReplyOrFollowUp].
func SendErrorMessage[T anyInteraction](ctx T, msg string, err error) {
	_ = SendReplyOrFollowUp(ctx, GenericErrorMessage(msg, err), true)
}

// GenericErrorMessage formats a generic error message with a consistent style suitable for end user presentation.
func GenericErrorMessage(msg string, err error) string {
	return fmt.Sprintf("%s\nError:\n```\n%s\n```", msg, err.Error())
}

// ConcurrentWriteMessage formats a standardized error message indicating that a help topic is currently being modified by another user.
func ConcurrentWriteMessage(userId string) string {
	return "⚠️ User <@" + userId + "> is currently modifying this help topic.\nPlease wait for them to finish and try again."
}

// OptimisticLockMessage formats a standardized error message indicating that a help topic triggered the database's optimistic locking failsafe
// (having been modified since its initial retrieval).
func OptimisticLockMessage(expected, actual time.Time) string {
	return fmt.Sprintf(
		"⚠️ Someone else has already modified this help topic since you started editing it. Review the updated version before trying again."+
			"\nYou started editing <t:%d:R>, while the last recorded update was <t:%d:R>.",
		expected.Unix(),
		actual.Unix(),
	)
}

// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
)

// ID for the actual modal itself
const addHelpModalId = "addHelpModal"

// NB: While the database does enforce optimistic locking during the database write, having a separate structure to track who's editing what
// allows us to improve UX.
// The database lock serves as a "last resort" of sorts to guard against server restarts or other unexpected events.

// A locker for tracking which users are currently modifying which help topics, preventing concurrent modifications to the same topics.
var usersModifyingHelpTopics = utils.NewLocker[string, tempest.Snowflake]()

var addHelp = command{
	Command: tempest.Command{
		Name:        "add-help",
		Description: "Add a new help topic to the database, or update an existing one's contents.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{
			{
				Type:        tempest.STRING_OPTION_TYPE,
				Name:        "topic",
				Description: "The help topic to create or update. Aliases will edit their target topic.",
				Required:    true,
				MinLength:   1,
				MaxLength:   100,
			},
		},
		SlashCommandHandler: handleAddHelp,
	},
	modalHandlers: map[string]modalHandler{
		addHelpModalId: addHelpTopic,
	},
}

func handleAddHelp(ctx *tempest.CommandInteraction) {
	topic, found := utils.ValidateOptionValue[string](ctx, "topic")
	if !found {
		return
	}

	if errMsg := checkTopicValidity(topic, "topic"); errMsg != "" {
		_ = ctx.SendLinearReply(errMsg, true)
		return
	}

	// Check if the command already exists, adjusting and pre-filling various portions of the modal if so.
	var (
		helpText   string
		topicName  = topic
		modalTitle = "Create new help topic"
		updatedAt  = time.Now()
	)
	existing, err := db.GetHelpTopic(topic)
	switch {
	case err == nil:
		helpText = existing.Text
		modalTitle = "Edit existing help topic"
		updatedAt = existing.UpdatedAt
		topicName = existing.Name // use the existing topic name to preserve capitalization/aliasing
	case errors.Is(err, sql.ErrNoRows):
		// do nothing (create a new one from scratch)
	default:
		utils.ErrorAttrs("Failed to check for existing help topic in database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorMessage(
			ctx,
			"Failed to check for existence of help topic "+topic+"!",
			err,
		)
		return
	}

	// check for concurrent modification using the base topic name
	if existing, locked := usersModifyingHelpTopics.GetLock(
		topicName,
	); locked && existing != ctx.BaseUser().ID {
		_ = ctx.SendLinearReply(utils.ConcurrentWriteMessage(existing.String()), true)
		return
	}

	sendAddHelpModal(ctx, topicName, modalTitle, helpText, updatedAt)
}

func sendAddHelpModal(
	ctx *tempest.CommandInteraction,
	topicName, modalTitle, helpText string,
	updatedAt time.Time,
) {
	// Store the timestamp in the 2 components' ID fields
	microLow, microHigh := convertTimestampToID(updatedAt)

	err := ctx.SendModal(tempest.ResponseModalData{
		Title:    modalTitle,
		CustomID: addHelpModalId,
		Components: []tempest.ModalComponent{
			tempest.TextDisplayComponent{
				ID:      microLow,
				Type:    tempest.TEXT_DISPLAY_COMPONENT_TYPE,
				Content: "### Selected Help Topic:\n`" + topicName + "`",
			},
			// TODO: Add support for more than just plaintext content in the help topic body
			// TODO: More configuration?
			tempest.LabelComponent{
				Type:  tempest.LABEL_COMPONENT_TYPE,
				Label: "What text should the topic display?",
				Component: tempest.TextInputComponent{
					Type: tempest.TEXT_INPUT_COMPONENT_TYPE,
					ID:   microHigh,
					// Store the topic name in the CustomID to retrieve later on
					CustomID: topicName,
					Style:    tempest.PARAGRAPH_TEXT_INPUT_STYLE,
					Required: true,
					Value:    helpText,
					// TODO: Do bots need to adhere to normal non-nitro message length limits?
					MaxLength:   2000,
					Placeholder: "Enter the help topic's body.\nAll Markdown features are supported.",
				},
			},
		},
	})
	if err != nil {
		utils.ErrorAttrs("Failed to send add help topic modal",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topicName),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorMessage(ctx, "Failed to send modal!", err)
		return
	}

	_ = usersModifyingHelpTopics.LockWithTimeout(
		topicName,
		ctx.BaseUser().ID,
		16*time.Minute,
	) // users have 15 minutes to submit the modal, so this should be fine

	utils.InfoAttrs("Sent add help modal successfully",
		slog.String("username", ctx.BaseUser().Username),
		slog.String("topic", topicName),
		slog.String("helpText", helpText),
		slog.Uint64("ID", uint64(ctx.ID)),
	)
}

// addHelpTopic handles the submission of the add-help modal.
func addHelpTopic(mtx tempest.ModalInteraction) {
	textDisplay, ok := mtx.Data.Components[0].(tempest.TextDisplayComponent)
	if !ok {
		slog.Error("Malformed add-help modal: first component was not a text display")
		_ = mtx.AcknowledgeWithLinearMessage(
			"Could not determine help topic from modal contents!",
			true,
		)
		return
	}

	label, ok := mtx.Data.Components[1].(tempest.LabelComponent)
	if !ok {
		slog.Error("Malformed add-help modal: second component was not a label")
		_ = mtx.AcknowledgeWithLinearMessage(
			"Could not determine help topic from modal contents!",
			true,
		)
		return
	}

	input, ok := label.Component.(tempest.TextInputComponent)
	if !ok {
		slog.Error("Malformed add-help modal: label component was not a text input")
		_ = mtx.AcknowledgeWithLinearMessage(
			"Could not determine help topic from modal contents!",
			true,
		)
		return
	}

	// extract stored values from the modal
	topicName := input.CustomID
	text := input.Value
	origUpdatedAt := extractTimestampFromID(textDisplay.ID, input.ID)

	if text == "" {
		utils.InfoAttrs("Body text not found in add-help modal",
			slog.String("username", mtx.BaseUser().Username),
			slog.String("topic", topicName),
			slog.Uint64("ID", uint64(mtx.ID)),
		)

		_ = mtx.AcknowledgeWithLinearMessage("Modal text content cannot be empty!", true)
		return
	}

	// check for concurrent modification again (in case the first check failed)
	userId, ok := usersModifyingHelpTopics.GetLock(topicName)
	if !ok || userId != mtx.BaseUser().ID {
		// someone else is still editing this topic
		utils.InfoAttrs("Concurrent modification detected for help topic",
			slog.String("topic", topicName),
			slog.String("existingUserID", userId.String()),
			slog.String("currentUserID", mtx.BaseUser().ID.String()),
		)

		// TODO: Do a diff with the current database contents?
		_ = mtx.AcknowledgeWithLinearMessage(
			utils.ConcurrentWriteMessage(userId.String())+
				"\nYour submitted text (in case you want to save it for later):"+
				"\n```\n"+text+"\n```",
			true,
		)
		return
	}

	defer func() {
		// release lock after we're done
		usersModifyingHelpTopics.Unlock(topicName)
	}()

	inserted, err := db.UpsertHelpTopic(topicName, text, origUpdatedAt)

	// check for db optimistic locking and give a slightly more helpful error message
	var staleErr db.ErrStaleTopic
	// TODO: Use errors.AsType[db.ErrStaleTopic](err) if we update to Go 1.26
	if errors.As(err, &staleErr) {
		utils.InfoAttrs("Failed to register help topic in database due to stale topic",
			slog.String("username", mtx.BaseUser().Username),
			slog.String("topic", topicName),
			slog.String("text", text),
			slog.Time("updatedAt", staleErr.LastUpdatedAt),
			slog.Time("expectedUpdatedAt", origUpdatedAt),
			slog.Uint64("ID", uint64(mtx.ID)),
		)
		_ = mtx.AcknowledgeWithLinearMessage(
			utils.OptimisticLockMessage(origUpdatedAt, staleErr.LastUpdatedAt)+
				"\nYour submitted text (in case you want to save it for later):"+
				"\n```\n"+text+"\n```",
			true)
		return
	} else if err != nil {
		utils.ErrorAttrs("Failed to register help topic in database",
			slog.String("username", mtx.BaseUser().Username),
			slog.String("topic", topicName),
			slog.String("text", text),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		_ = mtx.AcknowledgeWithLinearMessage(
			utils.GenericErrorMessage("Failed to add help topic to database!", err)+
				"\nYour submitted text (in case you want to save it for later):"+
				"\n```\n"+text+"\n```",
			true,
		)
		return
	}

	opStr := "updated"
	if inserted {
		opStr = "added"
	}

	utils.InfoAttrs("Topic successfully "+opStr+" to database",
		slog.String("username", mtx.BaseUser().Username),
		slog.String("topic", topicName),
		slog.String("text", text),
		slog.Uint64("ID", uint64(mtx.ID)),
	)

	_ = mtx.AcknowledgeWithLinearMessage(
		"Help topic `"+topicName+"` successfully "+opStr+" to the database!"+
			"\nTo view the topic, use the `/help` command.",
		true,
	)
}

// convertTimestampToID splits a timestamp into two 31-bit chunks to be stored in the component ID fields of the modal.
// (We use 31 bits instead of 32 as Discord requires these values fit within an int32.)
func convertTimestampToID(ti time.Time) (lo31, hi31 uint32) {
	const idMask31 = (1<<31 - 1)
	micro := uint64(ti.UnixMicro())
	microLow := uint32(micro & idMask31)
	microHigh := uint32((micro >> 31) & idMask31)
	return microLow, microHigh
}

// extractTimestampFromID reconstructs the original microsecond timestamp from the high and low 31-bit microsecond chunks
// stored inside the component ID fields.
// It returns the full timestamp in UTC.
func extractTimestampFromID(lo, hi uint32) time.Time {
	ts := (uint64(hi) << 31) | uint64(lo)
	return time.UnixMicro(int64(ts)).UTC()
}

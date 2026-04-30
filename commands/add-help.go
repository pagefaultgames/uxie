// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"database/sql"
	"errors"
	"log/slog"
	"regexp"
	"time"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
)

// ID for the actual modal itself
const addHelpModalId = "addHelpModal"

type concurrentModificationInfo struct {
	// The user attempting to modify the database.
	userId tempest.Snowflake
	// The previous `updatedAt` timestamp from the database, used for the database's optimistic locking.
	updatedAt time.Time
}

// NB: While thed database does enforce optimistic locking during the database write, having a separate structure to track who's editing what
// allows us to improve UX.

// A locker for tracking which users are currently modifying which help topics, preventing concurrent modifications to the same topics.
var usersModifyingHelpTopics = utils.NewLocker[string, concurrentModificationInfo]()

var addHelp = command{
	Command: tempest.Command{
		Name:        "add-help",
		Description: "Add a new help topic to the database, or update an existing one's contents.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{
			{
				Type:        tempest.STRING_OPTION_TYPE,
				Name:        "topic",
				Description: "The name of the help topic to create or update. Keep it short and concise!",
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

	if errMsg := checkTopicValidity(ctx, topic); errMsg != "" {
		_ = ctx.SendLinearReply(errMsg, true)
		return
	}

	var (
		helpText   string
		modalTitle = "Create new help topic"
		updatedAt  = time.Now()
	)

	// Check if the command already exists, pre-filling the body if so.
	existing, err := db.GetHelpTopic(topic)
	if err == nil {
		helpText = existing.Text
		modalTitle = "Edit existing help topic"
		updatedAt = existing.UpdatedAt
	} else if errors.Is(err, sql.ErrNoRows) {
		// do nothing (create a new one from scratch)
	} else {
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

	sendAddHelpModal(ctx, topic, modalTitle, helpText, updatedAt)
}

var pingRe = regexp.MustCompile(` @`)

// checkTopicValidity performs basic checks on the provided help topic, returning the error message to display to the user.
// An empty string signifies a valid topic.
func checkTopicValidity(ctx *tempest.CommandInteraction, topic string) (invalidMsg string) {
	if pingRe.MatchString(topic) {
		return "Topic names cannot contain the substring `@` to prevent unwanted mentions in help messages."
	}

	// check for concurrent modification
	if existing, locked := usersModifyingHelpTopics.GetLock(
		topic,
	); locked && existing.userId != ctx.BaseUser().ID {
		return "⚠️ User <@" + existing.userId.String() + "> is currently modifying this help topic. Please wait for them to finish and try again."
	}

	return ""
}

func sendAddHelpModal(
	ctx *tempest.CommandInteraction,
	topic, modalTitle, helpText string,
	updatedAt time.Time,
) {
	err := ctx.SendModal(tempest.ResponseModalData{
		Title:    modalTitle,
		CustomID: addHelpModalId,
		Components: []tempest.ModalComponent{
			tempest.TextDisplayComponent{
				Type:    tempest.TEXT_DISPLAY_COMPONENT_TYPE,
				Content: "### Selected Help Topic:\n`" + topic + "`",
			},
			// TODO: Add support for more than just plaintext content in the help topic body
			// TODO: More configuration?
			tempest.LabelComponent{
				Type:  tempest.LABEL_COMPONENT_TYPE,
				Label: "What text should the topic display?",
				Component: tempest.TextInputComponent{
					Type: tempest.TEXT_INPUT_COMPONENT_TYPE,
					// Store the topic name in the CustomID to retrieve later on.
					// We cannot use the text display component as Discord removes its contents from the JSON response
					CustomID: topic,
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
			slog.String("topic", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorMessage(ctx, "Failed to send modal!", err)
		return
	}

	_ = usersModifyingHelpTopics.LockWithTimeout(topic, concurrentModificationInfo{
		userId:    ctx.BaseUser().ID,
		updatedAt: updatedAt,
	}, 16*time.Minute) // users have 15 minutes to submit the modal, so this should be fine

	utils.InfoAttrs("Sent add help modal successfully",
		slog.String("username", ctx.BaseUser().Username),
		slog.String("topic", topic),
		slog.String("helpText", helpText),
		slog.Uint64("ID", uint64(ctx.ID)),
	)
}

// addHelpTopic handles the submission of the add-help modal.
func addHelpTopic(mtx tempest.ModalInteraction) {
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

	topic := input.CustomID
	text := input.Value

	if text == "" {
		utils.InfoAttrs("Body text not found in add-help modal",
			slog.String("username", mtx.BaseUser().Username),
			slog.String("topic", topic),
			slog.Uint64("ID", uint64(mtx.ID)),
		)

		_ = mtx.AcknowledgeWithLinearMessage("Modal text content cannot be empty!", true)
		return
	}

	// check for concurrent modification again (in case the first check failed)
	info, ok := usersModifyingHelpTopics.GetLock(topic)
	if !ok || info.userId != mtx.BaseUser().ID {
		// someone else is still editing this topic
		utils.InfoAttrs("Concurrent modification detected for help topic",
			slog.String("topic", topic),
			slog.String("existingUserID", info.userId.String()),
			slog.String("currentUserID", mtx.BaseUser().ID.String()),
		)

		// TODO: Do a diff with the current database contents?
		_ = mtx.AcknowledgeWithLinearMessage(
			"⚠️ User <@"+info.userId.String()+"> is currently modifying this help topic. Please wait for them to finish and try again."+
				"\nYour submitted text (in case you want to save it for later):"+
				"\n```\n"+text+"\n```",
			true,
		)
		return
	}

	defer func() {
		// release lock after we're done
		usersModifyingHelpTopics.Unlock(topic)
	}()

	if err := db.UpsertHelpTopic(topic, text, info.updatedAt); err != nil {
		utils.ErrorAttrs("Failed to register help topic in database",
			slog.String("username", mtx.BaseUser().Username),
			slog.String("topic", topic),
			slog.String("text", text),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		_ = mtx.AcknowledgeWithLinearMessage(
			utils.GetErrorMessage("Error adding help topic to database!", err)+
				"\nYour submitted text (in case you want to save it for later):"+
				"\n```\n"+text+"\n```", true)
		return
	}

	utils.InfoAttrs("Successfully added help topic to database",
		slog.String("username", mtx.BaseUser().Username),
		slog.String("topic", topic),
		slog.String("text", text),
		slog.Uint64("ID", uint64(mtx.ID)),
	)

	_ = mtx.AcknowledgeWithLinearMessage(
		"Help topic `"+topic+"` successfully added to database!"+
			"\nTo view the topic, use the `/help` command.",
		true,
	)
}

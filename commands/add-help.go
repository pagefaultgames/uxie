package commands

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/utils"
)

const (
	// ID for the "choose title/desc" input modal
	addHelpModalId          = "addHelpModalNameInput"
	addHelpModalTextInputId = "addHelpModalTextInput"
)

var addHelp = Command{
	Command: tempest.Command{
		Name:        "add-help",
		Description: "Add a new help topic or update an existing one.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{
			{
				Type:        tempest.STRING_OPTION_TYPE,
				Name:        "topic",
				Description: "The name of the help topic to create or update.",
				Required:    true,
				MinLength:   1,
				MaxLength:   100,
			},
		},
		SlashCommandHandler: handleAddHelp,
	},
}

func handleAddHelp(ctx *tempest.CommandInteraction) {
	_ = ctx.DeleteReply()

	topic, found := validateOptionValue[string](ctx, "topic")
	if !found {
		return
	}

	var (
		body       string
		modalTitle = "Create new help topic " + topic
	)
	// Check if the command already exists, pre-filling the body if it does
	existing, err := db.GetHelpTopic(topic)
	if errors.Is(err, sql.ErrNoRows) {
		body = existing.Text
		modalTitle = fmt.Sprintf("Edit existing help topic %s", topic)
	} else if err != nil {
		utils.ErrorAttrs("Failed to check for existing help topic in database",
			slog.String("username", ctx.User.Username),
			slog.String("command_name", ctx.Data.Name),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		sendErrorFollowUp(
			ctx,
			"Could not check check existence of help topic "+topic+"!",
			err,
		)
		return
	}

	if err := ctx.SendModal(tempest.ResponseModalData{
		Title: modalTitle,
		Components: []tempest.LayoutComponent{
			tempest.ContainerComponent{
				Type: tempest.CONTAINER_COMPONENT_TYPE,
				Components: []tempest.AnyComponent{
					// TODO: Add support for more than just plaintext
					tempest.LabelComponent{
						Label: "What text should the topic display?",
						Component: tempest.TextInputComponent{
							CustomID:    addHelpModalTextInputId,
							Type:        tempest.TEXT_INPUT_COMPONENT_TYPE,
							Style:       tempest.PARAGRAPH_TEXT_INPUT_STYLE,
							Required:    true,
							Value:       body,
							Placeholder: "Enter the topic's body. All Markdown features are supported.",
						},
					},
				},
			},
		},
	}); err != nil {
		utils.ErrorAttrs("Failed to send add help topic modal",
			slog.String("username", ctx.User.Username),
			slog.String("topic", topic),
			slog.String("command_name", ctx.Data.Name),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		sendErrorFollowUp(ctx, "Failed to send modal!", err)
		return
	}

	// Wait for either 15 minutes to pass or the modal to be submitted inside another goroutine
	timeout := time.After(15 * time.Minute)
	response, cancel, err := ctx.HTTPClient.AwaitModal([]string{addHelpModalId})
	if err != nil {
		utils.ErrorAttrs("Failed to await help topic modal response",
			slog.String("username", ctx.User.Username),
			slog.String("topic", topic),
			slog.String("command_name", ctx.Data.Name),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		sendErrorFollowUp(ctx, "Failed to wait for modal completion!", err)
		return
	}

	go func() {
		defer cancel()
		select {
		case <-timeout:
			utils.InfoAttrs("Waiting for modal response timed out after 15 minutes",
				slog.String("username", ctx.User.Username),
				slog.String("topic", topic),
				slog.String("command_name", ctx.Data.Name),
				slog.Uint64("ID", uint64(ctx.ID)),
			)
			_, _ = ctx.SendLinearFollowUp(
				"Modal timed out after 15 minutes. Please try again later.",
				true,
			)
		case mtx := <-response:
			addHelpTopic(mtx, topic)
		}
	}()
}

func addHelpTopic(mtx *tempest.ModalInteraction, topic string) {
	text := getTextInputValue(mtx, addHelpModalTextInputId)
	if text == "" {
		utils.ErrorAttrs("Text input not found in add-help modal",
			slog.String("username", mtx.User.Username),
			slog.String("topic", topic),
			slog.Uint64("ID", uint64(mtx.ID)),
		)

		_, _ = mtx.SendLinearFollowUp("Text input cannot be empty!", true)
		return
	}

	if err := db.AddHelpTopic(topic, text); err != nil {
		utils.ErrorAttrs("Failed to register new help topic in database",
			slog.String("username", mtx.User.Username),
			slog.String("topic", topic),
			slog.String("text", text),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		sendErrorFollowUp(mtx, "Could not add help topic to database!", err)
	}

	_, _ = mtx.SendLinearFollowUp(fmt.Sprintf("Help topic %s added successfully!", topic), true)
}

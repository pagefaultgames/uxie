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

	topic, found := utils.ValidateOptionValue[string](ctx, "topic")
	if !found {
		return
	}

	var (
		body       string
		modalTitle = "Create new help topic " + topic
	)

	// Check if the command already exists, pre-filling the body if so.
	existing, err := db.GetHelpTopic(topic)
	if err == nil {
		body = existing.Text
		modalTitle = fmt.Sprintf("Edit existing help topic %s", topic)
	} else if !errors.Is(err, sql.ErrNoRows) {
		utils.ErrorAttrs("Failed to check for existing help topic in database",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("command_name", ctx.Data.Name),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorFollowUp(
			ctx,
			"Could not check check existence of help topic "+topic+"!",
			err,
		)
		return
	}

	sendAddHelpModal(ctx, topic, modalTitle, body)
}

func sendAddHelpModal(ctx *tempest.CommandInteraction, topic, modalTitle, body string) {
	modalId := addHelpModalTextInputId + ctx.ID.String()
	err := ctx.SendModal(tempest.ResponseModalData{
		Title: modalTitle,
		Components: []tempest.LayoutComponent{
			tempest.ContainerComponent{
				Type: tempest.CONTAINER_COMPONENT_TYPE,
				Components: []tempest.AnyComponent{
					// TODO: Add support for more than just plaintext
					tempest.LabelComponent{
						Type:  tempest.LABEL_COMPONENT_TYPE,
						Label: "What text should the topic display?",
						Component: tempest.TextInputComponent{
							CustomID:    modalId,
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
	})
	if err != nil {
		utils.ErrorAttrs("Failed to send add help topic modal",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topic),
			slog.String("command_name", ctx.Data.Name),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorFollowUp(ctx, "Failed to send modal!", err)
		return
	}

	// Wait for either 2 minutes to pass or the modal to respond
	response, cancelFunc, err := ctx.HTTPClient.AwaitModal([]string{modalId})
	if err != nil || response == nil {
		utils.ErrorAttrs("Failed to await help topic modal response",
			slog.String("username", ctx.BaseUser().Username),
			slog.String("topic", topic),
			slog.String("command_name", ctx.Data.Name),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorFollowUp(ctx, "Failed to await modal response!", err)
		return
	}

	utils.InfoAttrs("Sent help modal successfully",
		slog.String("username", ctx.BaseUser().Username),
		slog.String("topic", topic),
		slog.String("command_name", ctx.Data.Name),
		slog.Uint64("ID", uint64(ctx.ID)),
		slog.String("modal_id", modalId),
	)

	_ = ctx.Defer(true)
	timeout := time.After(2 * time.Minute)
	go func() {
		defer cancelFunc()
		select {
		case mtx, ok := <-response:
			if !ok {
				utils.InfoAttrs("channel closed; id:" + modalId)
				return
			}
			utils.InfoAttrs("received help modal response")
			addHelpTopic(mtx, topic)
			return
		case <-timeout:
			utils.InfoAttrs("Waiting for modal response timed out after 2 minutes",
				slog.String("username", ctx.BaseUser().Username),
				slog.String("topic", topic),
				slog.String("command_name", ctx.Data.Name),
				slog.Uint64("ID", uint64(ctx.ID)),
			)
			_, _ = ctx.SendLinearFollowUp(
				"Modal timed out after 2 minutes. Please try again later.",
				true,
			)
			return
		}
	}()
}

func addHelpTopic(mtx *tempest.ModalInteraction, topic string) {
	if mtx == nil {
		utils.ErrorAttrs("Modal interaction is nil in addHelpTopic!",
			slog.String("topic", topic),
		)
		return
	}

	text := utils.GetTextInputValue(mtx, addHelpModalTextInputId)
	if text == "" {
		utils.ErrorAttrs("Text input not found in add-help modal",
			slog.String("username", mtx.BaseUser().Username),
			slog.String("topic", topic),
			slog.Uint64("ID", uint64(mtx.ID)),
		)

		_, _ = mtx.SendLinearFollowUp("Text input cannot be empty!", true)
		return
	}

	if err := db.AddHelpTopic(topic, text); err != nil {
		utils.ErrorAttrs("Failed to register new help topic in database",
			slog.String("username", mtx.BaseUser().Username),
			slog.String("topic", topic),
			slog.String("text", text),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorFollowUp(mtx, "Could not add help topic to database!", err)
	}

	_, _ = mtx.SendLinearFollowUp(fmt.Sprintf("Help topic %s added successfully!", topic), true)
}

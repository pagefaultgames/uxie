package commands

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/utils"
)

const (
	// ID for the actual modal itself
	addHelpModalId = "addHelpModal"
	// ID for the "choose title/desc" input modal
	addHelpModalTextInputId = "addHelpModalTextInput"
)

var addHelp = command{
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
	modalHandlers: map[string]modalHandler{
		addHelpModalId: addHelpTopic,
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
			slog.String("topic", topic),
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

const (
	helpTopicHeaderPre  = "### Selected Help Topic:\n`"
	helpTopicHeaderPost = "`"
)

func sendAddHelpModal(ctx *tempest.CommandInteraction, topic, modalTitle, body string) {
	err := ctx.SendModal(tempest.ResponseModalData{
		Title:    modalTitle,
		CustomID: addHelpModalId,
		Components: []tempest.ModalComponent{
			tempest.TextDisplayComponent{
				Type:    tempest.TEXT_DISPLAY_COMPONENT_TYPE,
				Content: helpTopicHeaderPre + topic + helpTopicHeaderPost,
			},
			// TODO: Add support for more than just plaintext content in the help topic body
			tempest.LabelComponent{
				Type:  tempest.LABEL_COMPONENT_TYPE,
				Label: "What text should the topic display?",
				Component: tempest.TextInputComponent{
					Type:     tempest.TEXT_INPUT_COMPONENT_TYPE,
					CustomID: addHelpModalTextInputId,
					Style:    tempest.PARAGRAPH_TEXT_INPUT_STYLE,
					Required: true,
					Value:    body,
					// TODO: Do bots need to adhere to normal non-nitro message length limits?
					MaxLength:   2000,
					Placeholder: "Enter the help topic's body. All Markdown features are supported.",
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
		utils.SendErrorFollowUp(ctx, "Failed to send modal!", err)
		return
	}

	utils.InfoAttrs("Sent add help modal successfully",
		slog.String("username", ctx.BaseUser().Username),
		slog.String("topic", topic),
		slog.Uint64("ID", uint64(ctx.ID)),
	)
}

// addHelpTopic handles the submission of the add-help modal.
func addHelpTopic(mtx tempest.ModalInteraction) {
	topic := getTopic(&mtx)
	if topic == "" {
		utils.ErrorAttrs("Failed to extract topic from add-help modal",
			slog.String("username", mtx.BaseUser().Username),
			slog.Uint64("ID", uint64(mtx.ID)),
		)
		_, _ = mtx.SendLinearFollowUp("Could not determine help topic from modal contents!", true)
		return
	}

	utils.InfoAttrs("Received add help modal response successfully",
		slog.String("username", mtx.BaseUser().Username),
		slog.String("topic", topic),
		slog.Uint64("ID", uint64(mtx.ID)),
	)

	text := mtx.GetInputValue(addHelpModalTextInputId)
	if text == "" {
		utils.ErrorAttrs("Text input not found in add-help modal",
			slog.String("username", mtx.BaseUser().Username),
			slog.Uint64("ID", uint64(mtx.ID)),
		)

		_, _ = mtx.SendLinearFollowUp("Text input cannot be empty!", true)
		return
	}

	if err := db.AddHelpTopic(topic, text); err != nil {
		utils.ErrorAttrs("Failed to register help topic in database",
			slog.String("username", mtx.BaseUser().Username),
			slog.String("topic", topic),
			slog.String("text", text),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorFollowUp(&mtx, "Could not add help topic to database!", err)
	}

	_, _ = mtx.SendLinearFollowUp(fmt.Sprintf("Help topic %s added successfully!", topic), true)
}

func getTopic(mtx *tempest.ModalInteraction) string {
	textDisplay, ok := mtx.Data.Components[0].(tempest.TextDisplayComponent)
	if !ok {
		return ""
	}

	return strings.TrimSuffix(
		strings.TrimPrefix(textDisplay.Content, helpTopicHeaderPre),
		helpTopicHeaderPost,
	)
}

package commands

import (
	"database/sql"
	"errors"
	"log/slog"
	"regexp"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
)

// ID for the actual modal itself
const addHelpModalId = "addHelpModal"

var addHelp = command{
	Command: tempest.Command{
		Name:        "add-help",
		Description: "Add a new help topic to the database, or update an existing one's contentx.",
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

	if errMsg := checkTopicValidity(topic); errMsg != "" {
		_ = ctx.SendLinearReply(errMsg, true)
		return
	}

	var (
		helpText   string
		modalTitle = "Create new help topic"
	)

	// Check if the command already exists, pre-filling the body if so.
	existing, err := db.GetHelpTopic(topic)
	if err == nil {
		helpText = existing.Text
		modalTitle = "Edit existing help topic"
	} else if !errors.Is(err, sql.ErrNoRows) {
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

	sendAddHelpModal(ctx, topic, modalTitle, helpText)
}

func sendAddHelpModal(ctx *tempest.CommandInteraction, topic, modalTitle, helpText string) {
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

	if err := db.AddHelpTopic(topic, text); err != nil {
		utils.ErrorAttrs("Failed to register help topic in database",
			slog.String("username", mtx.BaseUser().Username),
			slog.String("topic", topic),
			slog.String("text", text),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorMessage(&mtx, "Could not add help topic to database!", err)
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

var pingRe = regexp.MustCompile(` @`)

func checkTopicValidity(topic string) (invalidMsg string) {
	if pingRe.MatchString(topic) {
		return "Topic names cannot contain the substring `@` to prevent unwanted mentions in help messages."
	}
	return ""
}

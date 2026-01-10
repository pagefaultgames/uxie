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
	addHelpModalDescInputId = "addHelpModalDescInput"
)

var addHelp = Command{
	Command: tempest.Command{
		Name:                "add-help",
		Description:         "Add a new help topic or update an existing one.",
		Type:                tempest.CHAT_INPUT_COMMAND_TYPE,
		SlashCommandHandler: handleAddHelp,
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
	},
}

func handleAddHelp(ctx *tempest.CommandInteraction) {
	opt, found := ctx.GetOptionValue("topic")
	topic, ok := opt.(string)

	if !ok {
		utils.ErrorAttrs("Invalid type for add-help topic option",
			slog.String("username", ctx.User.Username),
			slog.Any("topic", opt),
			slog.Uint64("ID", uint64(ctx.ID)),
		)
		_, _ = ctx.SendLinearFollowUp(
			fmt.Sprintf(
				"Error: Help topic was of incorrect type (expected string, received %T)!",
				opt,
			),
			true,
		)
		return
	} else if topic == "" || !found {
		utils.ErrorAttrs("Empty topic provided to add-help command",
			slog.String("username", ctx.User.Username),
			slog.Uint64("ID", uint64(ctx.ID)),
		)
		_, _ = ctx.SendLinearFollowUp(
			"Error: Help topic cannot be empty!",
			true,
		)
		return
	}

	var (
		description, body string
		modalTitle        = fmt.Sprintf("Create new help topic %s", topic)
	)
	// Check if the command already exists, pre-filling the body if it does
	existing, err := db.HelpTopics.GetHelpTopic(topic)
	if errors.Is(err, sql.ErrNoRows) {
		description, body = existing.Description, existing.Text
		modalTitle = fmt.Sprintf("Edit existing help topic %s", topic)
	} else if err != nil {
		utils.ErrorAttrs("Failed to check for existing help topic",
			slog.String("username", ctx.User.Username),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		_, _ = ctx.SendLinearFollowUp("Error: Failed to check topic existence: "+err.Error(), true)
		return
	}

	if err := ctx.SendModal(tempest.ResponseModalData{
		Title: modalTitle,
		Components: []tempest.LayoutComponent{
			tempest.ContainerComponent{
				Type: tempest.CONTAINER_COMPONENT_TYPE,
				Components: []tempest.AnyComponent{
					tempest.LabelComponent{
						Label:       "What description should the help topic have?",
						Description: "Summarize the help topic in brief.",
						Component: tempest.TextInputComponent{
							CustomID:    addHelpModalDescInputId,
							Type:        tempest.TEXT_INPUT_COMPONENT_TYPE,
							Style:       tempest.PARAGRAPH_TEXT_INPUT_STYLE,
							Value:       description,
							Placeholder: "Enter the topic's description.",
						},
					},
					tempest.LabelComponent{
						Label: "What text should the topic display?",
						Component: tempest.TextInputComponent{
							CustomID:    addHelpModalTextInputId,
							Type:        tempest.TEXT_INPUT_COMPONENT_TYPE,
							Style:       tempest.PARAGRAPH_TEXT_INPUT_STYLE,
							Required:    true,
							Value:       body,
							Placeholder: "Enter the topic's body.",
						},
					},
				},
			},
		},
	}); err != nil {
		utils.ErrorAttrs("Failed to send help topic modal 2",
			slog.String("username", ctx.User.Username),
			slog.String("topic", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		_, _ = ctx.SendLinearFollowUp("error: Failed to send modal: "+err.Error(), true)
		return
	}

	// Wait for either 15 minutes to pass or the modal to be submitted inside another goroutine
	timeout := time.After(15 * time.Minute)
	response, cancel, err := ctx.HTTPClient.AwaitModal([]string{addHelpModalId})
	if err != nil {
		utils.ErrorAttrs("Failed to await help topic modal response",
			slog.String("username", ctx.User.Username),
			slog.String("topic", topic),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		_, _ = ctx.SendLinearFollowUp("error: Failed to wait for modal: "+err.Error(), true)
		return
	}

	go func() {
		defer cancel()
		select {
		case <-timeout:
			utils.InfoAttrs("Waiting for response timed out after 15 minutes!",
				slog.String("username", ctx.User.Username),
				slog.String("topic", topic),
				slog.Uint64("ID", uint64(ctx.ID)),
			)
			_, _ = ctx.SendLinearFollowUp("Modal timed out after 15 minutes.", true)
		case mtx := <-response:
			addHelpTopic(mtx, topic)
		}
	}()
}

func addHelpTopic(mtx *tempest.ModalInteraction, topic string) {
	description := getTextInputValue(mtx, addHelpModalDescInputId)
	if description == "" {
		utils.ErrorAttrs("Description input not found in add-help modal",
			slog.String("username", mtx.User.Username),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(mtx.ID)),
		)

		_, _ = mtx.SendLinearFollowUp("Error: Description input cannot be empty!", true)
		return
	}

	text := getTextInputValue(mtx, addHelpModalTextInputId)
	if text == "" {
		utils.ErrorAttrs("Text input not found in add-help modal",
			slog.String("username", mtx.User.Username),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(mtx.ID)),
		)

		_, _ = mtx.SendLinearFollowUp("Error: Text input cannot be empty!", true)
		return
	}

	if err := db.HelpTopics.AddHelpTopic(topic, description, text); err != nil {
		utils.ErrorAttrs("Failed to register new help topic in database",
			slog.String("username", mtx.User.Username),
			slog.String("name", topic),
			slog.String("description", description),
			slog.String("text", text),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		_, _ = mtx.SendLinearFollowUp("Error: could not store help topic in database", true)
	}
}

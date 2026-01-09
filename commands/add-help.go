package commands

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/types"
	"github.com/pagefaultgames/oranguru/utils"
)

const (
	addHelpModalNameInputId = "addHelpModalNameInput"
	addHelpModalTextInputId = "addHelpModalTextInput"
	addHelpModalDescInputId = "addHelpModalDescInput"
)

var addHelp = Command{
	Command: tempest.Command{
		Name:                "add-help",
		Description:         "Add a new help topic.",
		Type:                tempest.CHAT_INPUT_COMMAND_TYPE,
		SlashCommandHandler: handleAddHelp,
	},

	handlers: map[string]modalHandler{
		addHelpModalNameInputId: handleAddHelpName,
	},
}

func handleAddHelp(ctx *tempest.CommandInteraction) {
	if err := utils.SendDiscordMessage(ctx.HTTPClient, ctx.ChannelID,
		types.CreateMessageParams{
			Flags:   tempest.EPHEMERAL_MESSAGE_FLAG,
			Content: "What help topic would you like to add?",
			Components: []tempest.LayoutComponent{
				tempest.ContainerComponent{
					Type: tempest.CONTAINER_COMPONENT_TYPE,
					Components: []tempest.AnyComponent{
						tempest.LabelComponent{
							Type:        tempest.LABEL_COMPONENT_TYPE,
							Label:       "What should the command be named?",
							Description: "Keep it short and simple!",
							Component: tempest.TextInputComponent{
								CustomID:    addHelpModalNameInputId,
								Type:        tempest.TEXT_INPUT_COMPONENT_TYPE,
								Style:       tempest.SHORT_TEXT_INPUT_STYLE,
								MaxLength:   100,
								Required:    true,
								Placeholder: "Enter the name of the help topic to add.",
							},
						},
					},
				},
			},
		}, nil); err != nil {
		utils.ErrorAttrs("Failed to send help name modal", slog.String("error", err.Error()))
		_, _ = ctx.SendLinearFollowUp("error: Failed to create help topic: "+err.Error(), true)
		return
	}

	utils.InfoAttrs("Sent add-help modal to user", slog.String("username", ctx.User.Username))
}

func handleAddHelpName(mtx tempest.ModalInteraction) {
	topic := getTextInputValue(&mtx, addHelpModalNameInputId)

	if topic == "" {
		utils.ErrorAttrs("Failed to get name of add-help modal",
			slog.String("username", mtx.User.Username),
			slog.Uint64("ID", uint64(mtx.ID)),
		)
		_, _ = mtx.SendLinearFollowUp(
			"error: Failed to create help topic: invalid submission",
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
		utils.InfoAttrs("Editing existing help topic", slog.String("topic", topic))
		description, body = existing.Description, existing.Text
		modalTitle = fmt.Sprintf("Edit existing help topic %s", topic)
	} else if err != nil {
		utils.ErrorAttrs("Failed to check for existing help topic",
			slog.String("username", mtx.User.Username),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		_, _ = mtx.SendLinearFollowUp("error: Failed to check topic existence: "+err.Error(), true)
		return
	}

	if err := mtx.AcknowledgeWithModal(tempest.ResponseModalData{
		Title:    modalTitle,
		CustomID: "FOOOOOo",
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
			slog.String("username", mtx.User.Username),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		_, _ = mtx.SendLinearFollowUp("error: Failed to send modal: "+err.Error(), true)
		return
	}
	if err := mtx.BaseClient.RegisterModal("FOOOOOo", func(mtx tempest.ModalInteraction) {
		description := getTextInputValue(&mtx, addHelpModalDescInputId)
		text := getTextInputValue(&mtx, addHelpModalTextInputId)
		fmt.Println(description, text)
	}); err != nil {
		utils.ErrorAttrs("Failed to register help topic modal handler",
			slog.String("username", mtx.User.Username),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		_, _ = mtx.SendLinearFollowUp("error: Failed to register modal handler: "+err.Error(), true)
		return
	}
}

func addHelpTopic(mtx *tempest.ModalInteraction, name, description, text string) error {
	if err := db.HelpTopics.AddHelpTopic(name, description, text); err != nil {
		utils.ErrorAttrs("Failed to register new help topic in database",
			slog.String("username", mtx.User.Username),
			slog.String("name", name),
			slog.String("description", description),
			slog.String("text", text),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		return err
	}
	return nil
}

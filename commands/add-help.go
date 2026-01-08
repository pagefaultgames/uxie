package commands

import (
	"database/sql"
	"errors"
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
			Flags: tempest.EPHEMERAL_MESSAGE_FLAG,
			Content:    "What help topic would you like to add?",
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
		}, nil);  err != nil {
	utils.ErrorAttrs("Failed to send help name modal", slog.String("error", err.Error()))
		ctx.SendLinearReply("Error: Failed to create help topic: "+err.Error(), true)
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
		mtx.SendLinearFollowUp("Error: Failed to create help topic: invalid submission", true)
		return
	}

	var description, body string
	existing, err := db.HelpTopics.GetHelpTopic(topic)
	if errors.Is(err, sql.ErrNoRows) {
		utils.InfoAttrs("Editing existing help topic", slog.String("topic", topic))
		description, body = existing.Description, existing.Text
	} else if err != nil {
		utils.ErrorAttrs("Failed to check for existing help topic",
			slog.String("username", mtx.User.Username),
			slog.String("name", topic),
			slog.Uint64("ID", uint64(mtx.ID)),
			slog.Any("error", err),
		)
		mtx.SendLinearFollowUp("Error: Failed to check command existence: "+err.Error(), true)
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
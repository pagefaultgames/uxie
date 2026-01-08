package commands

import (
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/utils"
)

const addHelpModalCustomID = "addHelpModal"
const addHelpModalNameInputId = "addHelpModalNameInput"
const addHelpModalTextInputId = "addHelpModalTextInput"
const addHelpModalDescInputId = "addHelpModalDescInput"

var addHelp = Command{
	Command: tempest.Command{
		Name:                "add-help",
		Description:         "Add a new help command.",
		Type:                tempest.CHAT_INPUT_COMMAND_TYPE,
		SlashCommandHandler: handleAddHelp,
	},

	handlers: map[string]modalHandler{
		addHelpModalCustomID: handleAddHelpModal,
	},
}

func handleAddHelp(ctx *tempest.CommandInteraction) {
	err := ctx.SendModal(tempest.ResponseModalData{
		Title:    "Add Help Command",
		CustomID: addHelpModalCustomID,
		Components: []tempest.LayoutComponent{
			tempest.ContainerComponent{
				Type: tempest.CONTAINER_COMPONENT_TYPE,
				Components: []tempest.AnyComponent{
					tempest.LabelComponent{
						Type:        tempest.LABEL_COMPONENT_TYPE,
						Label:       "What should the command be named?",
						Description: "Keep it short and simple!",
						Component: tempest.TextInputComponent{
							CustomID: addHelpModalNameInputId,
							Type:        tempest.TEXT_INPUT_COMPONENT_TYPE,
							Style:       tempest.SHORT_TEXT_INPUT_STYLE,
							MaxLength:   100,
							Required:    true,
							Placeholder: "Enter the name of the help command to add.",
						},
					},
					tempest.LabelComponent{
						Label:       "What description should the help command have?",
						Description: "Summarize the help topic in brief.",
						Component: tempest.TextInputComponent{
							CustomID: addHelpModalDescInputId,
							Type:        tempest.TEXT_INPUT_COMPONENT_TYPE,
							Style:       tempest.PARAGRAPH_TEXT_INPUT_STYLE,
							Placeholder: "Enter the command's description.",
						},
					},
					tempest.LabelComponent{
						Label: "What text should the command display?",
						Component: tempest.TextInputComponent{
							CustomID: addHelpModalTextInputId,
							Type:        tempest.TEXT_INPUT_COMPONENT_TYPE,
							Style:       tempest.PARAGRAPH_TEXT_INPUT_STYLE,
							Required:    true,
							Placeholder: "Enter the text the command should display.",
						},
					},
				},
			},
		},
	})

	if err != nil {
		utils.ErrorAttrs("Failed to send add-help modal", slog.String("error", err.Error()))
		ctx.SendLinearReply("Error: Failed to create help command: "+err.Error(), true)
		return
	}

	utils.InfoAttrs("Sent add-help modal to user", slog.String("username", ctx.User.Username))
}

func handleAddHelpModal(mtx tempest.ModalInteraction) {
	name := getTextInputValue(&mtx, addHelpModalNameInputId)
	description := getTextInputValue(&mtx, addHelpModalDescInputId)
	text := getTextInputValue(&mtx, addHelpModalTextInputId)

	if name == "" || text == "" || description == "" {
		utils.ErrorAttrs("Failed to get contents of add-help modal",
			slog.String("username", mtx.User.Username),
			slog.String("name", name),
			slog.String("description", description),
			slog.String("text", text),
		)
		mtx.SendLinearFollowUp("Error: Failed to create help command: invalid submission", true)
		return
	}

	


}

package bot

import (
	"log/slog"

	"github.com/amatsagu/tempest"
	"google.golang.org/appengine/log"
)

const helpCommandModalCustomID = "helpCommandModal"

var addHelp = tempest.Command{
	Name:                "add-help",
	Description:         "Add a new help command.",
	Type:                tempest.CHAT_INPUT_COMMAND_TYPE,
	SlashCommandHandler: handleAddHelp,
}

func handleAddHelp(ctx *tempest.CommandInteraction) {
	err := ctx.SendModal(tempest.ResponseModalData{
		Title:    "Add Help Command",
		CustomID: helpCommandModalCustomID,
		Components: []tempest.LayoutComponent{
			tempest.ContainerComponent{
				Type: tempest.CONTAINER_COMPONENT_TYPE,
				Components: []tempest.AnyComponent{
					// Help input box
					tempest.LabelComponent{
						Type:        tempest.LABEL_COMPONENT_TYPE,
						Label:       "What should the command be named?",
						Description: "Keep it short and simple!",
						Component: tempest.TextInputComponent{
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
							Type:        tempest.TEXT_INPUT_COMPONENT_TYPE,
							Style:       tempest.PARAGRAPH_TEXT_INPUT_STYLE,
							Required:    true,
							Placeholder: "Enter the command's description.",
						},
					},
					tempest.LabelComponent{
						Label: "What text should the command display?",
						Component: tempest.TextInputComponent{
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
		slog.LogAttrs("Failed to send add-help modal", slog.Attrerr)
		ctx.SendLinearReply("Error: Failed to create help command: " + err.Error(), true)
		return
	}

	slog.LogAttrs("Sent add-help modal to user %s", ctx.User.Username)
}

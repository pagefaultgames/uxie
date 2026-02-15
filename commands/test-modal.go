package commands

import (
	"log/slog"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/utils"
)

const (
	testModalId          = "testModal"
	testModalTextInputId = "testModalTextInput"
)

var testModal = command{
	Command: tempest.Command{
		Name:        "test-modal",
		Description: "Test modal command with text display and input",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			sendTestModal(itx)
		},
	},
	modalHandlers: map[string]modalHandler{
		testModalId: handleTestModalSubmit,
	},
}

func sendTestModal(ctx *tempest.CommandInteraction) {
	err := ctx.SendModal(tempest.ResponseModalData{
		Title:    "Test Modal",
		CustomID: testModalId,
		Components: []tempest.ModalComponent{
			tempest.TextDisplayComponent{
				Type:    tempest.TEXT_DISPLAY_COMPONENT_TYPE,
				Content: "This is a test modal. Please enter some text below.",
			},
			tempest.LabelComponent{
				Type:        tempest.LABEL_COMPONENT_TYPE,
				Label:       "Enter your text",
				Description: "Type something here",
				Component: tempest.TextInputComponent{
					Type:        tempest.TEXT_INPUT_COMPONENT_TYPE,
					CustomID:    testModalTextInputId,
					Style:       tempest.SHORT_TEXT_INPUT_STYLE,
					Required:    true,
					Placeholder: "Enter text here",
					MinLength:   1,
					MaxLength:   100,
				},
			},
		},
	})
	if err != nil {
		utils.ErrorAttrs("Failed to send test modal",
			slog.String("username", ctx.BaseUser().Username),
			slog.Uint64("ID", uint64(ctx.ID)),
			slog.Any("error", err),
		)
		utils.SendErrorMessage(ctx, "Failed to send modal!", err)
		return
	}

	utils.InfoAttrs("Sent test modal successfully",
		slog.String("username", ctx.BaseUser().Username),
		slog.Uint64("ID", uint64(ctx.ID)),
	)
}

func handleTestModalSubmit(mtx tempest.ModalInteraction) {
	text := mtx.GetInputValue(testModalTextInputId)

	utils.InfoAttrs("Received test modal response",
		slog.String("username", mtx.BaseUser().Username),
		slog.String("input_text", text),
		slog.Uint64("ID", uint64(mtx.ID)),
	)

	_, _ = mtx.SendLinearFollowUp("You entered: "+text, true)
}

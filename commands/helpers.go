package commands

import (
	"github.com/amatsagu/tempest"
)

// getTextInputValue is a helper function to extract a text input modal's value from within a label.
// It returns the text contents, or an empty string if absent.
func getTextInputValue(
	mtx *tempest.ModalInteraction,
	customId string,
) (contents string) {
	for _, comp := range mtx.Data.Components {
		label, ok := comp.(tempest.LabelComponent)
		if !ok {
			continue
		}
		// NB: Golang type assertions produce zero values if type assertion fails
		c, found := label.Component.(tempest.TextInputComponent)
		if !found || c.CustomID != customId {
			continue
		}
		return c.Value
	}
	return ""
}

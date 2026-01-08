package commands

import (
	"github.com/amatsagu/tempest"
)

// getTextInputValue is a helper function to extract a text input modal's value from within a label.
// It returns the text contents, or an empty string if absent.
func getTextInputValue(
	itx *tempest.ModalInteraction,
	customId string,
) (contents string) {
	for _, comp := range itx.Data.Components {
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

// getStringSelectValue is a helper function to extract a string select modal's value from within a label.
// It returns the chosen option, or an empty string if absent.
func getStringSelectValue(
	itx *tempest.ModalInteraction,
	customId string,
) (contents []string) {
	for _, comp := range itx.Data.Components {
		label, ok := comp.(tempest.LabelComponent)
		if !ok {
			continue
		}
		// NB: Golang type assertions produce zero values if type assertion fails
		c, found := label.Component.(tempest.StringSelectComponent)
		if !found || c.CustomID != customId {
			continue
		}
		return c.Values
	}
	return nil
}

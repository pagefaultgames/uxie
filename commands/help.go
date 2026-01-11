package commands

import (
	"slices"
	"strings"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/utils"
)

// helpCommand is the base help command off of which all sub-topics extend.
// It does not do anything by itself and is simply a container.
var helpCommand = Command{
	Command: tempest.Command{
		Name:        "help",
		Description: "Get help on available topics.",
		Type:        tempest.CHAT_INPUT_COMMAND_TYPE,
		Options: []tempest.CommandOption{{
			Type:         tempest.STRING_OPTION_TYPE,
			Name:         "topic",
			Description:  "The name of the topic to get help on. Leave blank to see a list of all available topics.",
			Required:     false,
			AutoComplete: true,
		}},
		AutoCompleteHandler: handleHelpAutocomplete,
	},
}

func handleHelpAutocomplete(ctx tempest.CommandInteraction) []tempest.CommandOptionChoice {
	topics, err := db.GetAllTopics()
	if err != nil {
		utils.ErrorAttrs("error fetching topics from database")
	}

	// should always be a string
	_, t := ctx.GetFocusedValue()
	// nolint:errcheck
	focused := t.(string)

	choices := make([]tempest.CommandOptionChoice, 0, len(topics))
	for _, topic := range topics {
		if strings.HasPrefix(topic.Name, focused) {
			choices = append(choices, tempest.CommandOptionChoice{
				Name:  topic.Name,
				Value: topic.Name,
			})
		}
	}

	slices.SortFunc(choices, func(i, j tempest.CommandOptionChoice) int {
		return len(i.Name) - len(j.Name)
	})

	return choices
}

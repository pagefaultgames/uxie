package commands

import (
	"github.com/amatsagu/tempest"
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
	},
}

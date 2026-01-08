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
	},
}

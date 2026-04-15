package commands

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
)

// getAllTopicText fetches all help topics in the database,
// concatenating them into a single string.
// This allows multiple commands to append a list of valid topics for invalid input.
func getAllTopicText() (msg string, err error) {
	topics, err := db.GetAllTopics()
	if err != nil {
		return "", err
	}

	if len(topics) == 0 {
		return "None at the moment. Check back later!\n", nil
	}

	var b strings.Builder
	for _, topic := range topics {
		b.WriteString("- `" + topic.Name + "`\n")
	}

	return b.String(), nil
}

// printNonexistentTopic is a helper function to print a list of all available topics when a user requests a help topic that doesn't exist in the database.
func printNonexistentTopic(ctx *tempest.CommandInteraction, topicName string) {
	utils.InfoAttrs("Nonexistent help topic provided; showing all topics",
		slog.String("username", ctx.BaseUser().Username),
		slog.String("topic", topicName),
		slog.Uint64("ID", uint64(ctx.ID)),
		slog.String("commandName", ctx.Data.Name),
	)

	var b strings.Builder
	fmt.Fprintf(&b, "No help topic was found in the database with name %q.\n\n", topicName)
	text, err := getAllTopicText()
	if err == nil {
		b.WriteString("Here are the available topics:\n\n")
		b.WriteString(text)
	} else {
		b.WriteString(
			utils.GetErrorMessage("Additionally, failed to fetch list of available topics:", err),
		)
	}
	_ = ctx.SendLinearReply(b.String(), true)
}

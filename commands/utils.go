// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
)

var pingRe = regexp.MustCompile(`(^|\s)@`)

// checkTopicValidity performs basic checks on the name of a topic or alias, returning the error message to display to the user.
// An empty string signifies a valid topic.
//
// The provided descriptor will be used to customize the error message as applicable, and should be in lowercase for best results.
func checkTopicValidity(topic string, descriptor string) (invalidMsg string) {
	if pingRe.MatchString(topic) {
		return fmt.Sprintf(
			"%s names cannot contain the substring `@` immediately after whitespace to prevent unwanted mentions in help messages.",
			utils.TitleCase(descriptor),
		)
	}

	return ""
}

// getAllTopicText fetches all help topics (and optionally aliases) from the database, concatenating them into a single string.
// This allows multiple commands to append a list of valid topics for invalid input.
func getAllTopicText(includeAliases bool) (msg string, err error) {
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

	if includeAliases {
		aliases, err := db.GetAllAliases()
		if err != nil {
			return "", err
		}

		for _, alias := range aliases {
			b.WriteString("- `" + formatAliasText(alias.AliasName, alias.TopicName, -1) + "`\n")
		}
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
	fmt.Fprintf(&b, "No help topic was found in the database with name %q!\n\n", topicName)
	text, err := getAllTopicText(true)
	if err == nil {
		b.WriteString("Here are all available topics:\n\n")
		b.WriteString(text)
	} else {
		b.WriteString(
			utils.GenericErrorMessage(
				"Additionally, failed to fetch list of available topics:",
				err,
			),
		)
	}
	_ = ctx.SendLinearReply(b.String(), true)
}

// formatAliasText formats the display name for an alias in an autocomplete or other context, truncating if necessary to fit within maxLen characters.
// If maxLen is set to a negative number, no maximum length will be enforced.
func formatAliasText(aliasName, topicName string, maxLen int) string {
	separator := " --> "
	if maxLen < 0 {
		return aliasName + separator + topicName
	}

	minimumTopicLength := 1
	maxAliasLen := max(maxLen-len(separator)-minimumTopicLength, 1)
	aliasName = truncateText(aliasName, maxAliasLen)

	remainingTopicLen := max(maxLen-len(aliasName)-len(separator), 1)

	return aliasName + separator + truncateText(topicName, remainingTopicLen)
}

// truncateText truncates text to be no more than maxLen characters, replacing the end with an ellipsis if necessary.
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return text[:maxLen-3] + "..."
}

package commands

import (
	"strings"

	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/utils"
)

// getAllTopicText fetches all help topics in the database,
// concatenating them into a single string.
// This allows multiple commands to append a list of valid topics for invalid input.
func getAllTopicText() string {
	topics, err := db.GetAllTopics()
	if err != nil {
		return utils.GetErrorMessage("Failed to fetch help topics from database!", err)
	}

	var b strings.Builder
	b.WriteString("## Available help topics:\n")
	if len(topics) == 0 {
		b.WriteString("None at the moment. Check back later!\n")
	} else {
		for _, topic := range topics {
			b.WriteString("- `" + topic.Name + "`\n")
		}
	}

	return b.String()
}

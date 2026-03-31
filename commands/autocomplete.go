package commands

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
)

// helpTopicAutocompleteFunc returns an autocomplete handler function for the option value with the given name.
func helpTopicAutocompleteFunc(
	optionName string,
) func(ctx tempest.CommandInteraction) []tempest.CommandOptionChoice {
	return func(ctx tempest.CommandInteraction) []tempest.CommandOptionChoice {
		name, val := ctx.GetFocusedValue()
		if name != optionName {
			return nil
		}

		focusedText, ok := val.(string)
		if !ok {
			utils.ErrorAttrs("Invalid type for help topic autocomplete option",
				slog.String("username", ctx.BaseUser().Username),
				slog.Uint64("ID", uint64(ctx.ID)),
				slog.String("commandName", ctx.Data.Name),
				slog.String("optionName", name),
				slog.Any("value", val),
			)
			return nil
		}

		// NB: Not caching this is OK for now since it only takes ~200 microsecs in total
		topics, err := db.GetAllTopics()
		if err != nil {
			utils.ErrorAttrs("Error fetching topics from database for autocomplete",
				slog.String("username", ctx.BaseUser().Username),
				slog.Uint64("ID", uint64(ctx.ID)),
				slog.String("commandName", ctx.Data.Name),
				slog.Any("error", err),
			)
			return []tempest.CommandOptionChoice{}
		}

		choices := make([]tempest.CommandOptionChoice, 0, len(topics))
		for _, topic := range topics {
			if strings.HasPrefix(strings.ToLower(topic.Name), strings.ToLower(focusedText)) {
				choices = append(choices, tempest.CommandOptionChoice{
					Name:  topic.Name,
					Value: topic.Name,
				})
			}
		}

		// Sort in shortest to longest
		slices.SortFunc(choices, func(i, j tempest.CommandOptionChoice) int {
			return len(j.Name) - len(i.Name)
		})

		return choices
	}
}

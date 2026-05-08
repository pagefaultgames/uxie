// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
)

// TODO: Set these up in a way to allow composing autocomplete handlers if or when that is desired

// helpTopicAutocompleteFunc returns an autocomplete handler function for the option value with the given name
// that matches help topic names.
// If includeAliases is true, the autocomplete choices will include both help topics and their aliases
// (the latter being displayed in the format "aliasName --> topicName").
func helpTopicAutocompleteFunc(
	optionName string,
	includeAliases bool,
) func(ctx tempest.CommandInteraction) []tempest.CommandOptionChoice {
	return func(ctx tempest.CommandInteraction) []tempest.CommandOptionChoice {
		name, val := ctx.GetFocusedValue()
		if name != optionName {
			return nil
		}

		focusedText, ok := val.(string)
		if !ok {
			utils.ErrorAttrs("Incorrect type given for help topic autocomplete option",
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
				slog.String("optionName", name),
				slog.Any("error", err),
			)
			return nil
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

		if includeAliases {
			aliases, err := db.GetAllAliases()
			if err != nil {
				utils.ErrorAttrs("Error fetching aliases from database for autocomplete",
					slog.String("username", ctx.BaseUser().Username),
					slog.Uint64("ID", uint64(ctx.ID)),
					slog.String("commandName", ctx.Data.Name),
					slog.String("optionName", name),
					slog.Any("error", err),
				)
				return nil
			}

			for _, alias := range aliases {
				if strings.HasPrefix(
					strings.ToLower(alias.AliasName),
					strings.ToLower(focusedText),
				) {
					choices = append(choices, tempest.CommandOptionChoice{
						Name: formatAliasText(
							alias.AliasName,
							alias.TopicName,
							MAX_AUTOCOMPLETE_CHOICE_NAME_LENGTH,
						),
						Value: alias.AliasName,
					})
				}
			}
		}

		return sortAndTruncateAutocompleteChoices(choices)
	}
}

// aliasAutocompleteFunc returns an autocomplete handler function for the option value with the given name.
// (the latter being displayed in the format "aliasName --> topicName").
func aliasAutocompleteFunc(
	optionName string,
) func(ctx tempest.CommandInteraction) []tempest.CommandOptionChoice {
	return func(ctx tempest.CommandInteraction) []tempest.CommandOptionChoice {
		name, val := ctx.GetFocusedValue()
		if name != optionName {
			return nil
		}

		focusedText, ok := val.(string)
		if !ok {
			utils.ErrorAttrs("Incorrect type given for alias autocomplete option",
				slog.String("username", ctx.BaseUser().Username),
				slog.Uint64("ID", uint64(ctx.ID)),
				slog.String("commandName", ctx.Data.Name),
				slog.String("optionName", name),
				slog.Any("value", val),
			)
			return nil
		}

		aliases, err := db.GetAllAliases()
		if err != nil {
			utils.ErrorAttrs("Error fetching aliases from database for autocomplete",
				slog.String("username", ctx.BaseUser().Username),
				slog.Uint64("ID", uint64(ctx.ID)),
				slog.String("commandName", ctx.Data.Name),
				slog.String("optionName", name),
				slog.Any("error", err),
			)
			return nil
		}

		choices := make([]tempest.CommandOptionChoice, 0, len(aliases))
		for _, alias := range aliases {
			if strings.HasPrefix(strings.ToLower(alias.AliasName), strings.ToLower(focusedText)) {
				choices = append(choices, tempest.CommandOptionChoice{
					Name: formatAliasText(
						alias.AliasName,
						alias.TopicName,
						MAX_AUTOCOMPLETE_CHOICE_NAME_LENGTH,
					),
					Value: alias.AliasName,
				})
			}
		}

		return sortAndTruncateAutocompleteChoices(choices)
	}
}

func sortAndTruncateAutocompleteChoices(
	choices []tempest.CommandOptionChoice,
) []tempest.CommandOptionChoice {
	// Sort in shortest to longest, then lexicographically
	// TODO: Change to levenshtein distance sorting eventually
	// (which may or may not require a custom implementation)
	slices.SortFunc(choices, func(i, j tempest.CommandOptionChoice) int {
		if diff := len(i.Name) - len(j.Name); diff != 0 {
			return diff
		}
		return strings.Compare(i.Name, j.Name)
	})

	// truncate at 25 options
	if len(choices) > MAX_AUTOCOMPLETE_CHOICES {
		choices = choices[:MAX_AUTOCOMPLETE_CHOICES]
	}
	return choices
}

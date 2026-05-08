// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"fmt"
	"log/slog"

	"github.com/amatsagu/tempest"
)

// ValidateOptionValue is a helper function to extract and validate a slash command's option value.
// It returns the retrieved value (which will be zero if missing/invalid) and whether the provided value was both present and valid.
//
// This function performs requisite logging and error messaging on failure, so callers can simply return immediately
// if ok is false.
//
// Note that this function only supports required options. For a variant that replaces missing values with a default value, see [GetOptionOrDefault].
func ValidateOptionValue[T string | bool | float64](
	ctx *tempest.CommandInteraction,
	optName string,
) (value T, ok bool) {
	opt, ok := ctx.GetOptionValue(optName)
	if !ok {
		// missing required option
		slog.Error(
			fmt.Sprintf(
				"Command option %s was not found inside command data!\nDid you forget to mark it as required?",
				optName,
			),
		)
		return value, false
	}

	value, ok = opt.(T)
	if ok {
		return value, true
	}

	ErrorAttrs(
		"Invalid input type for command option",
		slog.String("username", ctx.BaseUser().Username),
		slog.Uint64("ID", uint64(ctx.ID)),
		slog.String("optionName", optName),
		slog.String("type", fmt.Sprintf("%T", opt)),
		slog.String("expected_type", fmt.Sprintf("%T", value)),
		slog.String("commandName", ctx.Data.Name),
	)
	_ = SendReplyOrFollowUp(ctx,
		fmt.Sprintf(
			"Option %q was of incorrect type!\nExpected a value of type %T, but received %v (type %T)!",
			optName,
			value,
			opt,
			opt,
		),
		true,
	)
	return value, false
}

// GetOptionOrDefault is a helper function to extract the value of a non-required option, returning a specified default if the option is missing or invalid.
func GetOptionOrDefault[T string | bool | float64](
	ctx *tempest.CommandInteraction,
	optName string,
	defaultValue T,
) (value T) {
	opt, ok := ctx.GetOptionValue(optName)
	if !ok {
		return defaultValue
	}

	value, ok = opt.(T)
	if ok {
		return value
	}
	return defaultValue
}

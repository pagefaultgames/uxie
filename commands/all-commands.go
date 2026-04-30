// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

// allCommands is the list of bot commands registered by default.
var allCommands = []command{
	addHelp,
	helpCommand,
	pingCommand,
	getTopics,
	deleteTopic,
}

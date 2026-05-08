// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TitleCase converts a string to title case.
func TitleCase(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

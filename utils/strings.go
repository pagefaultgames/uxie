// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

// TitleCase converts a string to title case.
func TitleCase(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

// NormalizeUnicode normalizes a string by removing diacritics and other combining marks.
func NormalizeUnicode(s string) string {
	// decompose into NFKD, then strip all relevant runes away
	normalized := norm.NFKD.String(s)
	return strings.Map(func(r rune) rune {
		if 0x300 <= r && r <= 0x036F {
			return -1
		}
		return r
	}, normalized)
}

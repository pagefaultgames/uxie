package utils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TitleCase(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import "testing"

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"hello world", "Hello World"},
		{"HELLO WORLD", "Hello World"},
		{"hElLo WoRlD", "Hello World"},
		{"123abc", "123Abc"},
		{"foo-bar", "Foo-Bar"},
		{"mObY DIcK", "Moby Dick"},
		{"call me ishmael", "Call Me Ishmael"},
	}

	for _, tt := range tests {
		if got := TitleCase(tt.name); got != tt.want {
			t.Errorf("TitleCase(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNormalizeUnicode(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"café", "cafe"},
		{"naïveté", "naivete"},
		{"résumé", "resume"},
		{"cooperate", "cooperate"},
		{"façade", "facade"},
		{"élève", "eleve"},
		{"Pokédex", "Pokedex"},
	}

	for _, tt := range tests {
		if got := NormalizeUnicode(tt.name); got != tt.want {
			t.Errorf("NormalizeUnicode(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

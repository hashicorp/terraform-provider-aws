// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"testing"
)

func TestFirstLower(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"empty":                {input: "", want: ""},
		"already lower":        {input: "hello", want: "hello"},
		"first upper":          {input: "Hello", want: "hello"},
		"only first lowered":   {input: "HELLO", want: "hELLO"},
		"single upper rune":    {input: "H", want: "h"},
		"leading digit":        {input: "1abc", want: "1abc"},
		"leading space":        {input: " Hello", want: " Hello"},
		"unicode first rune":   {input: "Éclair", want: "éclair"},
		"multibyte unaffected": {input: "Über", want: "über"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got, want := FirstLower(tc.input), tc.want; got != want {
				t.Errorf("FirstLower(%q) = %q, want %q", tc.input, got, want)
			}
		})
	}
}

func TestFirstUpper(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"empty":                {input: "", want: ""},
		"already upper":        {input: "Hello", want: "Hello"},
		"first lower":          {input: "hello", want: "Hello"},
		"only first uppered":   {input: "hELLO", want: "HELLO"},
		"single lower rune":    {input: "h", want: "H"},
		"leading digit":        {input: "1abc", want: "1abc"},
		"leading space":        {input: " hello", want: " hello"},
		"unicode first rune":   {input: "éclair", want: "Éclair"},
		"multibyte first rune": {input: "über", want: "Über"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got, want := FirstUpper(tc.input), tc.want; got != want {
				t.Errorf("FirstUpper(%q) = %q, want %q", tc.input, got, want)
			}
		})
	}
}

func TestTitle(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"empty":                {input: "", want: ""},
		"single word":          {input: "hello", want: "Hello"},
		"two words":            {input: "hello world", want: "Hello World"},
		"multiple words":       {input: "the quick brown fox", want: "The Quick Brown Fox"},
		"already titled":       {input: "Hello World", want: "Hello World"},
		"preserves inner caps": {input: "hello WORLD", want: "Hello WORLD"},
		"does not lower rest":  {input: "mixedCASE words", want: "MixedCASE Words"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got, want := Title(tc.input), tc.want; got != want {
				t.Errorf("Title(%q) = %q, want %q", tc.input, got, want)
			}
		})
	}
}

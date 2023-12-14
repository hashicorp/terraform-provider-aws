// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ujson_test

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/ujson"
)

type quoteTest struct {
	in      string
	out     string
	ascii   string
	graphic string
}

var quotetests = []quoteTest{
	{"\a\b\f\r\n\t\v", `"\a\b\f\r\n\t\v"`, `"\a\b\f\r\n\t\v"`, `"\a\b\f\r\n\t\v"`},
	{"\\", `"\\"`, `"\\"`, `"\\"`},
	{"abc\xffdef", `"abc\xffdef"`, `"abc\xffdef"`, `"abc\xffdef"`},
	{"\u263a", `"☺"`, `"\u263a"`, `"☺"`},
	{"\U0010ffff", `"\U0010ffff"`, `"\U0010ffff"`, `"\U0010ffff"`},
	{"\x04", `"\x04"`, `"\x04"`, `"\x04"`},
	// Some non-printable but graphic runes. Final column is double-quoted.
	{"!\u00a0!\u2000!\u3000!", `"!\u00a0!\u2000!\u3000!"`, `"!\u00a0!\u2000!\u3000!"`, "\"!\u00a0!\u2000!\u3000!\""},
}

func TestQuote(t *testing.T) {
	t.Parallel()

	for _, tt := range quotetests {
		if out := ujson.AppendQuote([]byte("abc"), []byte(tt.in)); string(out) != "abc"+tt.out {
			t.Errorf("AppendQuote(%q, %s) = %s, want %s", "abc", tt.in, out, "abc"+tt.out)
		}
	}
}

func TestQuoteToASCII(t *testing.T) {
	t.Parallel()

	for _, tt := range quotetests {
		if out := ujson.AppendQuoteToASCII([]byte("abc"), []byte(tt.in)); string(out) != "abc"+tt.ascii {
			t.Errorf("AppendQuoteToASCII(%q, %s) = %s, want %s", "abc", tt.in, out, "abc"+tt.ascii)
		}
	}
}

func TestQuoteToGraphic(t *testing.T) {
	t.Parallel()

	for _, tt := range quotetests {
		if out := ujson.AppendQuoteToGraphic([]byte("abc"), []byte(tt.in)); string(out) != "abc"+tt.graphic {
			t.Errorf("AppendQuoteToGraphic(%q, %s) = %s, want %s", "abc", tt.in, out, "abc"+tt.graphic)
		}
	}
}

type unQuoteTest struct {
	in  string
	out string
}

var unquotetests = []unQuoteTest{
	{`""`, ""},
	{`"a"`, "a"},
	{`"abc"`, "abc"},
	{`"☺"`, "☺"},
	{`"hello world"`, "hello world"},
	{`"\xFF"`, "\xFF"},
	{`"\377"`, "\377"},
	{`"\u1234"`, "\u1234"},
	{`"\U00010111"`, "\U00010111"},
	{`"\U0001011111"`, "\U0001011111"},
	{`"\a\b\f\n\r\t\v\\\""`, "\a\b\f\n\r\t\v\\\""},
	{`"'"`, "'"},
}

var misquoted = []string{
	``,
	`"`,
	`"a`,
	`"'`,
	`b"`,
	`"\"`,
	`"\9"`,
	`"\19"`,
	`"\129"`,
	`'\'`,
	`'\9'`,
	`'\19'`,
	`'\129'`,
	`'ab'`,
	`"\x1!"`,
	`"\U12345678"`,
	`"\z"`,
	"`",
	"`xxx",
	"`\"",
	`"\'"`,
	`'\"'`,
	"\"\n\"",
	"\"\\n\n\"",
	"'\n'",
}

func TestUnquote(t *testing.T) {
	t.Parallel()

	for _, tt := range unquotetests {
		if out, err := ujson.Unquote([]byte(tt.in)); err != nil || string(out) != tt.out {
			t.Errorf("Unquote(%#q) = %q, %v want %q, nil", tt.in, out, err, tt.out)
		}
	}

	// run the quote tests too, backward
	for _, tt := range quotetests {
		if in, err := ujson.Unquote([]byte(tt.out)); string(in) != tt.in {
			t.Errorf("Unquote(%#q) = %q, %v, want %q, nil", tt.out, in, err, tt.in)
		}
	}

	for _, s := range misquoted {
		if out, err := ujson.Unquote([]byte(s)); out != nil || !errors.Is(err, ujson.ErrSyntax) {
			t.Errorf("Unquote(%#q) = %q, %v want %q, %v", s, out, err, "", ujson.ErrSyntax)
		}
	}
}

// Issue 23685: invalid UTF-8 should not go through the fast path.
func TestUnquoteInvalidUTF8(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in string

		// one of:
		want    string
		wantErr string
	}{
		{in: `"foo"`, want: "foo"},
		{in: `"foo`, wantErr: "invalid syntax"},
		{in: `"` + "\xc0" + `"`, want: "\xef\xbf\xbd"},
		{in: `"a` + "\xc0" + `"`, want: "a\xef\xbf\xbd"},
		{in: `"\t` + "\xc0" + `"`, want: "\t\xef\xbf\xbd"},
	}
	for i, tt := range tests {
		got, err := ujson.Unquote([]byte(tt.in))
		var gotErr string
		if err != nil {
			gotErr = err.Error()
		}
		if gotErr != tt.wantErr {
			t.Errorf("%d. Unquote(%q) = err %v; want %q", i, tt.in, err, tt.wantErr)
		}
		if tt.wantErr == "" && err == nil && string(got) != tt.want {
			t.Errorf("%d. Unquote(%q) = %02x; want %02x", i, tt.in, got, []byte(tt.want))
		}
	}
}

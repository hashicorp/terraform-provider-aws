package scanner

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hil/ast"
)

func TestScanner(t *testing.T) {
	cases := []struct {
		Input  string
		Output []TokenType
	}{
		{
			"",
			[]TokenType{EOF},
		},

		{
			"$",
			[]TokenType{LITERAL, EOF},
		},

		{
			`${"\`,
			[]TokenType{BEGIN, OQUOTE, INVALID, EOF},
		},

		{
			"foo",
			[]TokenType{LITERAL, EOF},
		},

		{
			"foo$bar",
			[]TokenType{LITERAL, EOF},
		},

		{
			"foo ${0}",
			[]TokenType{LITERAL, BEGIN, INTEGER, END, EOF},
		},

		{
			"foo ${0.}",
			[]TokenType{LITERAL, BEGIN, INTEGER, PERIOD, END, EOF},
		},

		{
			"foo ${0.0}",
			[]TokenType{LITERAL, BEGIN, FLOAT, END, EOF},
		},

		{
			"foo ${0.0.0}",
			[]TokenType{LITERAL, BEGIN, FLOAT, PERIOD, INTEGER, END, EOF},
		},

		{
			"föo ${bar}",
			[]TokenType{LITERAL, BEGIN, IDENTIFIER, END, EOF},
		},

		{
			"foo ${bar.0.baz}",
			[]TokenType{LITERAL, BEGIN, IDENTIFIER, END, EOF},
		},

		{
			"foo ${bar.foo-bar.baz}",
			[]TokenType{LITERAL, BEGIN, IDENTIFIER, END, EOF},
		},

		{
			"foo $${bar}",
			[]TokenType{LITERAL, EOF},
		},

		{
			"foo $$$${bar}",
			[]TokenType{LITERAL, EOF},
		},

		{
			`foo ${"bár"}`,
			[]TokenType{LITERAL, BEGIN, OQUOTE, STRING, CQUOTE, END, EOF},
		},

		{
			"${bar(baz)}",
			[]TokenType{
				BEGIN,
				IDENTIFIER, OPAREN, IDENTIFIER, CPAREN,
				END, EOF,
			},
		},

		{
			"${bar(baz4, foo_ooo)}",
			[]TokenType{
				BEGIN,
				IDENTIFIER, OPAREN,
				IDENTIFIER, COMMA, IDENTIFIER,
				CPAREN,
				END, EOF,
			},
		},

		{
			"${bár(42)}",
			[]TokenType{
				BEGIN,
				IDENTIFIER, OPAREN, INTEGER, CPAREN,
				END, EOF,
			},
		},

		{
			"${bar(-42)}",
			[]TokenType{
				BEGIN,
				IDENTIFIER, OPAREN, MINUS, INTEGER, CPAREN,
				END, EOF,
			},
		},

		{
			"${bar(42+1)}",
			[]TokenType{
				BEGIN,
				IDENTIFIER, OPAREN,
				INTEGER, PLUS, INTEGER,
				CPAREN,
				END, EOF,
			},
		},

		{
			"${true && false}",
			[]TokenType{
				BEGIN,
				BOOL, AND, BOOL,
				END, EOF,
			},
		},

		{
			"${true || false}",
			[]TokenType{
				BEGIN,
				BOOL, OR, BOOL,
				END, EOF,
			},
		},

		{
			"${1 == 5}",
			[]TokenType{
				BEGIN,
				INTEGER, EQUAL, INTEGER,
				END, EOF,
			},
		},

		{
			"${1 != 5}",
			[]TokenType{
				BEGIN,
				INTEGER, NOTEQUAL, INTEGER,
				END, EOF,
			},
		},

		{
			"${1 > 5}",
			[]TokenType{
				BEGIN,
				INTEGER, GT, INTEGER,
				END, EOF,
			},
		},

		{
			"${1 < 5}",
			[]TokenType{
				BEGIN,
				INTEGER, LT, INTEGER,
				END, EOF,
			},
		},

		{
			"${1 <= 5}",
			[]TokenType{
				BEGIN,
				INTEGER, LTE, INTEGER,
				END, EOF,
			},
		},

		{
			"${1 >= 5}",
			[]TokenType{
				BEGIN,
				INTEGER, GTE, INTEGER,
				END, EOF,
			},
		},

		{
			"${true ? 1 : 5}",
			[]TokenType{
				BEGIN,
				BOOL, QUESTION, INTEGER, COLON, INTEGER,
				END, EOF,
			},
		},

		{
			"${bar(3.14159)}",
			[]TokenType{
				BEGIN,
				IDENTIFIER, OPAREN, FLOAT, CPAREN,
				END, EOF,
			},
		},

		{
			"${bar(true)}",
			[]TokenType{
				BEGIN,
				IDENTIFIER, OPAREN, BOOL, CPAREN,
				END, EOF,
			},
		},

		{
			"${bar(inner(_baz))}",
			[]TokenType{
				BEGIN,
				IDENTIFIER, OPAREN,
				IDENTIFIER, OPAREN,
				IDENTIFIER,
				CPAREN, CPAREN,
				END, EOF,
			},
		},

		{
			"foo ${foo.bar.baz}",
			[]TokenType{
				LITERAL,
				BEGIN,
				IDENTIFIER,
				END, EOF,
			},
		},

		{
			"foo ${foo.bar.*.baz}",
			[]TokenType{
				LITERAL,
				BEGIN,
				IDENTIFIER,
				END, EOF,
			},
		},

		{
			"foo ${foo.bar.*}",
			[]TokenType{
				LITERAL,
				BEGIN,
				IDENTIFIER,
				END, EOF,
			},
		},

		{
			"foo ${foo.bar.*baz}",
			[]TokenType{
				LITERAL,
				BEGIN,
				IDENTIFIER, PERIOD, STAR, IDENTIFIER,
				END, EOF,
			},
		},

		{
			"foo ${foo*}",
			[]TokenType{
				LITERAL,
				BEGIN,
				IDENTIFIER, STAR,
				END, EOF,
			},
		},

		{
			`foo ${foo("baz")}`,
			[]TokenType{
				LITERAL,
				BEGIN,
				IDENTIFIER, OPAREN, OQUOTE, STRING, CQUOTE, CPAREN,
				END, EOF,
			},
		},

		{
			`foo ${"${var.foo}"}`,
			[]TokenType{
				LITERAL,
				BEGIN,
				OQUOTE,
				BEGIN,
				IDENTIFIER,
				END,
				CQUOTE,
				END,
				EOF,
			},
		},

		{
			"${1 = 5}",
			[]TokenType{
				BEGIN,
				INTEGER,
				INVALID,
				EOF,
			},
		},

		{
			"${1 & 5}",
			[]TokenType{
				BEGIN,
				INTEGER,
				INVALID,
				EOF,
			},
		},

		{
			"${1 | 5}",
			[]TokenType{
				BEGIN,
				INTEGER,
				INVALID,
				EOF,
			},
		},

		{
			`${unclosed`,
			[]TokenType{BEGIN, IDENTIFIER, EOF},
		},

		{
			`${"unclosed`,
			[]TokenType{BEGIN, OQUOTE, INVALID, EOF},
		},
	}

	for _, tc := range cases {
		ch := Scan(tc.Input, ast.InitPos)
		var actual []TokenType
		for token := range ch {
			actual = append(actual, token.Type)
		}

		if !reflect.DeepEqual(actual, tc.Output) {
			t.Errorf(
				"\nInput: %s\nBad:   %#v\nWant:  %#v",
				tc.Input, tokenTypeNames(actual), tokenTypeNames(tc.Output),
			)
		}
	}
}

func TestScannerPos(t *testing.T) {
	cases := []struct {
		Input     string
		Positions []ast.Pos
	}{
		{
			`foo`,
			[]ast.Pos{
				{Line: 1, Column: 1},
				{Line: 1, Column: 4},
			},
		},
		{
			`föo`,
			[]ast.Pos{
				{Line: 1, Column: 1},
				{Line: 1, Column: 4},
			},
		},
		{
			// Ideally combining diacritic marks would actually get
			// counted as only one character, but this test asserts
			// our current compromise the "Column" counts runes
			// rather than graphemes.
			`fĉo`,
			[]ast.Pos{
				{Line: 1, Column: 1},
				{Line: 1, Column: 5},
			},
		},
		{
			// Spaces in literals are counted as part of the literal.
			` foo `,
			[]ast.Pos{
				{Line: 1, Column: 1},
				{Line: 1, Column: 6},
			},
		},
		{
			`${foo}`,
			[]ast.Pos{
				{Line: 1, Column: 1},
				{Line: 1, Column: 3},
				{Line: 1, Column: 6},
				{Line: 1, Column: 7},
			},
		},
		{
			// Spaces inside interpolation sequences are skipped
			`${ foo }`,
			[]ast.Pos{
				{Line: 1, Column: 1},
				{Line: 1, Column: 4},
				{Line: 1, Column: 8},
				{Line: 1, Column: 9},
			},
		},
		{
			`${föo}`,
			[]ast.Pos{
				{Line: 1, Column: 1},
				{Line: 1, Column: 3},
				{Line: 1, Column: 6},
				{Line: 1, Column: 7},
			},
		},
		{
			`${fĉo}`,
			[]ast.Pos{
				{Line: 1, Column: 1},
				{Line: 1, Column: 3},
				{Line: 1, Column: 7},
				{Line: 1, Column: 8},
			},
		},
		{
			`foo ${ foo } foo`,
			[]ast.Pos{
				{Line: 1, Column: 1},
				{Line: 1, Column: 5},
				{Line: 1, Column: 8},
				{Line: 1, Column: 12},
				{Line: 1, Column: 13},
				{Line: 1, Column: 17},
			},
		},
		{
			`foo ${ " foo " } foo`,
			[]ast.Pos{
				{Line: 1, Column: 1},  // LITERAL
				{Line: 1, Column: 5},  // BEGIN
				{Line: 1, Column: 8},  // OQUOTE
				{Line: 1, Column: 9},  // STRING
				{Line: 1, Column: 14}, // CQUOTE
				{Line: 1, Column: 16}, // END
				{Line: 1, Column: 17}, // LITERAL
				{Line: 1, Column: 21}, // EOF
			},
		},
	}

	for _, tc := range cases {
		ch := Scan(tc.Input, ast.Pos{Line: 1, Column: 1})
		var actual []ast.Pos
		for token := range ch {
			actual = append(actual, token.Pos)
		}

		if !reflect.DeepEqual(actual, tc.Positions) {
			t.Errorf(
				"\nInput: %s\nBad:   %#v\nWant:  %#v",
				tc.Input, posStrings(actual), posStrings(tc.Positions),
			)
		}
	}
}

func tokenTypeNames(types []TokenType) []string {
	ret := make([]string, len(types))
	for i, t := range types {
		ret[i] = t.String()
	}
	return ret
}

func posStrings(positions []ast.Pos) []string {
	ret := make([]string, len(positions))
	for i, pos := range positions {
		ret[i] = pos.String()
	}
	return ret
}

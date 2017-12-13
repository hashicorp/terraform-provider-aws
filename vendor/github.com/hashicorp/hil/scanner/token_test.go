package scanner

import (
	"testing"
)

func TestTokenString(t *testing.T) {
	cases := []struct {
		Token  *Token
		String string
	}{
		{
			&Token{
				Type:    EOF,
				Content: "",
			},
			"end of string",
		},

		{
			&Token{
				Type:    INVALID,
				Content: "baz",
			},
			`invalid sequence "baz"`,
		},

		{
			&Token{
				Type:    INTEGER,
				Content: "1",
			},
			`integer 1`,
		},

		{
			&Token{
				Type:    FLOAT,
				Content: "1.2",
			},
			`float 1.2`,
		},

		{
			&Token{
				Type:    STRING,
				Content: "foo",
			},
			`string "foo"`,
		},

		{
			&Token{
				Type:    LITERAL,
				Content: "foo",
			},
			`literal "foo"`,
		},

		{
			&Token{
				Type:    BOOL,
				Content: "true",
			},
			`"true"`,
		},

		{
			&Token{
				Type:    BEGIN,
				Content: "${",
			},
			`"${"`,
		},
	}

	for _, tc := range cases {
		str := tc.Token.String()
		if got, want := str, tc.String; got != want {
			t.Errorf(
				"%s %q returned %q; want %q",
				tc.Token.Type, tc.Token.Content,
				got, want,
			)
		}
	}
}

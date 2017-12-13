package hclsyntax

import (
	"reflect"
	"testing"
)

func TestPeeker(t *testing.T) {
	tokens := Tokens{
		{
			Type: TokenIdent,
		},
		{
			Type: TokenComment,
		},
		{
			Type: TokenIdent,
		},
		{
			Type: TokenComment,
		},
		{
			Type: TokenIdent,
		},
		{
			Type: TokenNewline,
		},
		{
			Type: TokenIdent,
		},
		{
			Type: TokenNewline,
		},
		{
			Type: TokenIdent,
		},
		{
			Type: TokenNewline,
		},
		{
			Type: TokenEOF,
		},
	}

	{
		peeker := newPeeker(tokens, true)

		wantTypes := []TokenType{
			TokenIdent,
			TokenComment,
			TokenIdent,
			TokenComment,
			TokenIdent,
			TokenNewline,
			TokenIdent,
			TokenNewline,
			TokenIdent,
			TokenNewline,
			TokenEOF,
		}
		var gotTypes []TokenType

		for {
			peeked := peeker.Peek()
			read := peeker.Read()
			if peeked.Type != read.Type {
				t.Errorf("mismatched Peek %s and Read %s", peeked, read)
			}

			gotTypes = append(gotTypes, read.Type)

			if read.Type == TokenEOF {
				break
			}
		}

		if !reflect.DeepEqual(gotTypes, wantTypes) {
			t.Errorf("wrong types\ngot:  %#v\nwant: %#v", gotTypes, wantTypes)
		}
	}

	{
		peeker := newPeeker(tokens, false)

		wantTypes := []TokenType{
			TokenIdent,
			TokenIdent,
			TokenIdent,
			TokenNewline,
			TokenIdent,
			TokenNewline,
			TokenIdent,
			TokenNewline,
			TokenEOF,
		}
		var gotTypes []TokenType

		for {
			peeked := peeker.Peek()
			read := peeker.Read()
			if peeked.Type != read.Type {
				t.Errorf("mismatched Peek %s and Read %s", peeked, read)
			}

			gotTypes = append(gotTypes, read.Type)

			if read.Type == TokenEOF {
				break
			}
		}

		if !reflect.DeepEqual(gotTypes, wantTypes) {
			t.Errorf("wrong types\ngot:  %#v\nwant: %#v", gotTypes, wantTypes)
		}
	}

	{
		peeker := newPeeker(tokens, false)

		peeker.PushIncludeNewlines(false)

		wantTypes := []TokenType{
			TokenIdent,
			TokenIdent,
			TokenIdent,
			TokenIdent,
			TokenIdent,
			TokenNewline, // we'll pop off the PushIncludeNewlines before we get here
			TokenEOF,
		}
		var gotTypes []TokenType

		idx := 0
		for {
			peeked := peeker.Peek()
			read := peeker.Read()
			if peeked.Type != read.Type {
				t.Errorf("mismatched Peek %s and Read %s", peeked, read)
			}

			gotTypes = append(gotTypes, read.Type)

			if read.Type == TokenEOF {
				break
			}

			if idx == 4 {
				peeker.PopIncludeNewlines()
			}

			idx++
		}

		if !reflect.DeepEqual(gotTypes, wantTypes) {
			t.Errorf("wrong types\ngot:  %#v\nwant: %#v", gotTypes, wantTypes)
		}
	}
}

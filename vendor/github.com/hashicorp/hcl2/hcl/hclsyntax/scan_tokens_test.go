package hclsyntax

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/kylelemons/godebug/pretty"
)

func TestScanTokens_normal(t *testing.T) {
	tests := []struct {
		input string
		want  []Token
	}{
		// Empty input
		{
			``,
			[]Token{
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 0, Line: 1, Column: 1},
					},
				},
			},
		},
		{
			` `,
			[]Token{
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
		},
		{
			"\n\n",
			[]Token{
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 2, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 2, Line: 3, Column: 1},
					},
				},
			},
		},

		// TokenNumberLit
		{
			`1`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`1`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
		},
		{
			`12`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`12`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
			},
		},
		{
			`12.3`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`12.3`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
			},
		},
		{
			`1e2`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`1e2`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
			},
		},
		{
			`1e+2`,
			[]Token{
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`1e+2`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
			},
		},

		// TokenIdent
		{
			`hello`,
			[]Token{
				{
					Type:  TokenIdent,
					Bytes: []byte(`hello`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
			},
		},
		{
			`h3ll0`,
			[]Token{
				{
					Type:  TokenIdent,
					Bytes: []byte(`h3ll0`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
			},
		},
		{
			`héllo`, // combining acute accent
			[]Token{
				{
					Type:  TokenIdent,
					Bytes: []byte(`héllo`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 6},
					},
				},
			},
		},

		// Literal-only Templates (string literals, effectively)
		{
			`""`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
			},
		},
		{
			`"hello"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(`hello`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
			},
		},

		// Templates with interpolations and control sequences
		{
			`"${1}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`1`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
			},
		},
		{
			`"%{a}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateControl,
					Bytes: []byte(`%{`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`a`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
			},
		},
		{
			`"${{}}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenOBrace,
					Bytes: []byte(`{`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenCBrace,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
			},
		},
		{
			`"${""}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 1, Column: 6},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
			},
		},
		{
			`"${"${a}"}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`a`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 9, Line: 1, Column: 10},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 1, Column: 10},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 10, Line: 1, Column: 11},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 11, Line: 1, Column: 12},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
			},
		},
		{
			`"${"${a} foo"}"`,
			[]Token{
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenOQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 5},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte(`${`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte(`a`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenQuotedLit,
					Bytes: []byte(` foo`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 12, Line: 1, Column: 13},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 1, Column: 13},
						End:   hcl.Pos{Byte: 13, Line: 1, Column: 14},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte(`}`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 13, Line: 1, Column: 14},
						End:   hcl.Pos{Byte: 14, Line: 1, Column: 15},
					},
				},
				{
					Type:  TokenCQuote,
					Bytes: []byte(`"`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 14, Line: 1, Column: 15},
						End:   hcl.Pos{Byte: 15, Line: 1, Column: 16},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 15, Line: 1, Column: 16},
						End:   hcl.Pos{Byte: 15, Line: 1, Column: 16},
					},
				},
			},
		},

		// Heredoc Templates
		{
			`<<EOT
hello world
EOT
`,
			[]Token{
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello world\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 18, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 18, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 22, Line: 4, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 22, Line: 4, Column: 1},
						End:   hcl.Pos{Byte: 22, Line: 4, Column: 1},
					},
				},
			},
		},
		{
			`<<EOT
hello ${name}
EOT
`,
			[]Token{
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello "),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 12, Line: 2, Column: 7},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte("${"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 2, Column: 7},
						End:   hcl.Pos{Byte: 14, Line: 2, Column: 9},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("name"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 14, Line: 2, Column: 9},
						End:   hcl.Pos{Byte: 18, Line: 2, Column: 13},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte("}"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 18, Line: 2, Column: 13},
						End:   hcl.Pos{Byte: 19, Line: 2, Column: 14},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 19, Line: 2, Column: 14},
						End:   hcl.Pos{Byte: 20, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 20, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 24, Line: 4, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 24, Line: 4, Column: 1},
						End:   hcl.Pos{Byte: 24, Line: 4, Column: 1},
					},
				},
			},
		},
		{
			`<<EOT
${name}EOT
EOT
`,
			[]Token{
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte("${"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 8, Line: 2, Column: 3},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("name"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 2, Column: 3},
						End:   hcl.Pos{Byte: 12, Line: 2, Column: 7},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte("}"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 2, Column: 7},
						End:   hcl.Pos{Byte: 13, Line: 2, Column: 8},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 13, Line: 2, Column: 8},
						End:   hcl.Pos{Byte: 17, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOT\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 17, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 21, Line: 4, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 21, Line: 4, Column: 1},
						End:   hcl.Pos{Byte: 21, Line: 4, Column: 1},
					},
				},
			},
		},
		{
			`<<EOF
${<<-EOF
hello
EOF
}
EOF
`,
			[]Token{
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<EOF\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte("${"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 8, Line: 2, Column: 3},
					},
				},
				{
					Type:  TokenOHeredoc,
					Bytes: []byte("<<-EOF\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 2, Column: 3},
						End:   hcl.Pos{Byte: 15, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 15, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 21, Line: 4, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOF\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 21, Line: 4, Column: 1},
						End:   hcl.Pos{Byte: 25, Line: 5, Column: 1},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte("}"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 25, Line: 5, Column: 1},
						End:   hcl.Pos{Byte: 26, Line: 5, Column: 2},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 26, Line: 5, Column: 2},
						End:   hcl.Pos{Byte: 27, Line: 6, Column: 1},
					},
				},
				{
					Type:  TokenCHeredoc,
					Bytes: []byte("EOF\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 27, Line: 6, Column: 1},
						End:   hcl.Pos{Byte: 31, Line: 7, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 31, Line: 7, Column: 1},
						End:   hcl.Pos{Byte: 31, Line: 7, Column: 1},
					},
				},
			},
		},

		// Combinations
		{
			` (1 + 2) * 3 `,
			[]Token{
				{
					Type:  TokenOParen,
					Bytes: []byte(`(`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`1`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenPlus,
					Bytes: []byte(`+`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 5},
						End:   hcl.Pos{Byte: 5, Line: 1, Column: 6},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`2`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenCParen,
					Bytes: []byte(`)`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenStar,
					Bytes: []byte(`*`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 1, Column: 10},
						End:   hcl.Pos{Byte: 10, Line: 1, Column: 11},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte(`3`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 11, Line: 1, Column: 12},
						End:   hcl.Pos{Byte: 12, Line: 1, Column: 13},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 13, Line: 1, Column: 14},
						End:   hcl.Pos{Byte: 13, Line: 1, Column: 14},
					},
				},
			},
		},
		{
			"\na = 1\n",
			[]Token{
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("a"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 2, Line: 2, Column: 2},
					},
				},
				{
					Type:  TokenEqual,
					Bytes: []byte("="),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 2, Column: 3},
						End:   hcl.Pos{Byte: 4, Line: 2, Column: 4},
					},
				},
				{
					Type:  TokenNumberLit,
					Bytes: []byte("1"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 5, Line: 2, Column: 5},
						End:   hcl.Pos{Byte: 6, Line: 2, Column: 6},
					},
				},
				{
					Type:  TokenNewline,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 2, Column: 6},
						End:   hcl.Pos{Byte: 7, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 3, Column: 1},
					},
				},
			},
		},

		// Comments
		{
			"# hello\n",
			[]Token{
				{
					Type:  TokenComment,
					Bytes: []byte("# hello\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 8, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 8, Line: 2, Column: 1},
					},
				},
			},
		},
		{
			"// hello\n",
			[]Token{
				{
					Type:  TokenComment,
					Bytes: []byte("// hello\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 9, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 9, Line: 2, Column: 1},
					},
				},
			},
		},
		{
			"/* hello */",
			[]Token{
				{
					Type:  TokenComment,
					Bytes: []byte("/* hello */"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 11, Line: 1, Column: 12},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
			},
		},

		// Invalid things
		{
			`🌻`,
			[]Token{
				{
					Type:  TokenInvalid,
					Bytes: []byte(`🌻`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 4, Line: 1, Column: 2},
					},
				},
			},
		},
		{
			`|`,
			[]Token{
				{
					Type:  TokenBitwiseOr,
					Bytes: []byte(`|`),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
		},
		{
			"\x80", // UTF-8 continuation without an introducer
			[]Token{
				{
					Type:  TokenBadUTF8,
					Bytes: []byte{0x80},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 1, Line: 1, Column: 2},
					},
				},
			},
		},
		{
			" \x80\x80", // UTF-8 continuation without an introducer
			[]Token{
				{
					Type:  TokenBadUTF8,
					Bytes: []byte{0x80},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenBadUTF8,
					Bytes: []byte{0x80},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 3, Line: 1, Column: 4},
						End:   hcl.Pos{Byte: 3, Line: 1, Column: 4},
					},
				},
			},
		},
		{
			"\t\t",
			[]Token{
				{
					Type:  TokenTabs,
					Bytes: []byte{0x09, 0x09},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 2, Line: 1, Column: 3},
						End:   hcl.Pos{Byte: 2, Line: 1, Column: 3},
					},
				},
			},
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := scanTokens([]byte(test.input), "", hcl.Pos{Byte: 0, Line: 1, Column: 1}, scanNormal)

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(test.want, got)
				t.Errorf(
					"wrong result\ninput: %s\ndiff:  %s",
					test.input, diff,
				)
			}
		})
	}
}

func TestScanTokens_template(t *testing.T) {
	tests := []struct {
		input string
		want  []Token
	}{
		// Empty input
		{
			``,
			[]Token{
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 0, Line: 1, Column: 1},
					},
				},
			},
		},

		// Simple literals
		{
			` hello `,
			[]Token{
				{
					Type:  TokenStringLit,
					Bytes: []byte(` hello `),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 1, Column: 8},
						End:   hcl.Pos{Byte: 7, Line: 1, Column: 8},
					},
				},
			},
		},
		{
			"\nhello\n",
			[]Token{
				{
					Type:  TokenStringLit,
					Bytes: []byte("\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 1, Line: 2, Column: 1},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello\n"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 1, Line: 2, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 3, Column: 1},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 7, Line: 3, Column: 1},
						End:   hcl.Pos{Byte: 7, Line: 3, Column: 1},
					},
				},
			},
		},
		{
			"hello ${foo} hello",
			[]Token{
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello "),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte("${"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 8, Line: 1, Column: 9},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("foo"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8, Line: 1, Column: 9},
						End:   hcl.Pos{Byte: 11, Line: 1, Column: 12},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte("}"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 11, Line: 1, Column: 12},
						End:   hcl.Pos{Byte: 12, Line: 1, Column: 13},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte(" hello"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 1, Column: 13},
						End:   hcl.Pos{Byte: 18, Line: 1, Column: 19},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 18, Line: 1, Column: 19},
						End:   hcl.Pos{Byte: 18, Line: 1, Column: 19},
					},
				},
			},
		},
		{
			"hello ${~foo~} hello",
			[]Token{
				{
					Type:  TokenStringLit,
					Bytes: []byte("hello "),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0, Line: 1, Column: 1},
						End:   hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
				{
					Type:  TokenTemplateInterp,
					Bytes: []byte("${~"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 6, Line: 1, Column: 7},
						End:   hcl.Pos{Byte: 9, Line: 1, Column: 10},
					},
				},
				{
					Type:  TokenIdent,
					Bytes: []byte("foo"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 9, Line: 1, Column: 10},
						End:   hcl.Pos{Byte: 12, Line: 1, Column: 13},
					},
				},
				{
					Type:  TokenTemplateSeqEnd,
					Bytes: []byte("~}"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 12, Line: 1, Column: 13},
						End:   hcl.Pos{Byte: 14, Line: 1, Column: 15},
					},
				},
				{
					Type:  TokenStringLit,
					Bytes: []byte(" hello"),
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 14, Line: 1, Column: 15},
						End:   hcl.Pos{Byte: 20, Line: 1, Column: 21},
					},
				},
				{
					Type:  TokenEOF,
					Bytes: []byte{},
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 20, Line: 1, Column: 21},
						End:   hcl.Pos{Byte: 20, Line: 1, Column: 21},
					},
				},
			},
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := scanTokens([]byte(test.input), "", hcl.Pos{Byte: 0, Line: 1, Column: 1}, scanTemplate)

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(test.want, got)
				t.Errorf(
					"wrong result\ninput: %s\ndiff:  %s",
					test.input, diff,
				)
			}
		})
	}
}

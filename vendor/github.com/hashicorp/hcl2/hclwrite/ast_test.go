package hclwrite

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hcl"
)

func TestBodyFindAttribute(t *testing.T) {
	tests := []struct {
		src  string
		name string
		want *TokenSeq
	}{
		{
			"",
			"a",
			nil,
		},
		{
			"a = 1\n",
			"a",
			&TokenSeq{
				Tokens{
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte{'a'},
					},
				},
			},
		},
		{
			"a = 1\nb = 1\nc = 1\n",
			"a",
			&TokenSeq{
				Tokens{
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte{'a'},
					},
				},
			},
		},
		{
			"a = 1\nb = 1\nc = 1\n",
			"b",
			&TokenSeq{
				Tokens{
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte{'b'},
					},
				},
			},
		},
		{
			"a = 1\nb = 1\nc = 1\n",
			"c",
			&TokenSeq{
				Tokens{
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte{'c'},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s in %s", test.name, test.src), func(t *testing.T) {
			f, diags := ParseConfig([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf("- %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			attr := f.Body.FindAttribute(test.name)
			if attr == nil {
				if test.want != nil {
					t.Errorf("attribute found, but expecting not found")
				}
			} else {
				got := attr.NameTokens
				if !reflect.DeepEqual(got, test.want) {
					t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.want))
				}
			}
		})
	}
}

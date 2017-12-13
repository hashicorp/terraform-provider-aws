package hclwrite

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/kylelemons/godebug/pretty"
)

func TestParse(t *testing.T) {
	tests := []struct {
		src  string
		want *Body
	}{
		{
			"",
			&Body{
				Items:     nil,
				AllTokens: nil,
			},
		},
		{
			"a = 1\n",
			&Body{
				Items: []Node{
					&Attribute{
						AllTokens: &TokenSeq{
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenIdent,
										Bytes:        []byte(`a`),
										SpacesBefore: 0,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenEqual,
										Bytes:        []byte(`=`),
										SpacesBefore: 1,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenNumberLit,
										Bytes:        []byte(`1`),
										SpacesBefore: 1,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenNewline,
										Bytes:        []byte{'\n'},
										SpacesBefore: 0,
									},
								},
							},
						},
						NameTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenIdent,
								Bytes:        []byte(`a`),
								SpacesBefore: 0,
							},
						}},
						EqualsTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenEqual,
								Bytes:        []byte(`=`),
								SpacesBefore: 1,
							},
						}},
						Expr: &Expression{
							AllTokens: &TokenSeq{Tokens{
								{
									Type:         hclsyntax.TokenNumberLit,
									Bytes:        []byte(`1`),
									SpacesBefore: 1,
								},
							}},
						},
						EOLTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenNewline,
								Bytes:        []byte{'\n'},
								SpacesBefore: 0,
							},
						}},
					},
				},
				AllTokens: &TokenSeq{
					&TokenSeq{
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenIdent,
									Bytes:        []byte(`a`),
									SpacesBefore: 0,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenEqual,
									Bytes:        []byte(`=`),
									SpacesBefore: 1,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenNumberLit,
									Bytes:        []byte(`1`),
									SpacesBefore: 1,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenNewline,
									Bytes:        []byte{'\n'},
									SpacesBefore: 0,
								},
							},
						},
					},
				},
			},
		},
		{
			"# aye aye aye\na = 1\n",
			&Body{
				Items: []Node{
					&Attribute{
						AllTokens: &TokenSeq{
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenComment,
										Bytes:        []byte("# aye aye aye\n"),
										SpacesBefore: 0,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenIdent,
										Bytes:        []byte(`a`),
										SpacesBefore: 0,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenEqual,
										Bytes:        []byte(`=`),
										SpacesBefore: 1,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenNumberLit,
										Bytes:        []byte(`1`),
										SpacesBefore: 1,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenNewline,
										Bytes:        []byte{'\n'},
										SpacesBefore: 0,
									},
								},
							},
						},
						LeadCommentTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenComment,
								Bytes:        []byte("# aye aye aye\n"),
								SpacesBefore: 0,
							},
						}},
						NameTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenIdent,
								Bytes:        []byte(`a`),
								SpacesBefore: 0,
							},
						}},
						EqualsTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenEqual,
								Bytes:        []byte(`=`),
								SpacesBefore: 1,
							},
						}},
						Expr: &Expression{
							AllTokens: &TokenSeq{Tokens{
								{
									Type:         hclsyntax.TokenNumberLit,
									Bytes:        []byte(`1`),
									SpacesBefore: 1,
								},
							}},
						},
						EOLTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenNewline,
								Bytes:        []byte{'\n'},
								SpacesBefore: 0,
							},
						}},
					},
				},
				AllTokens: &TokenSeq{
					&TokenSeq{
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenComment,
									Bytes:        []byte("# aye aye aye\n"),
									SpacesBefore: 0,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenIdent,
									Bytes:        []byte(`a`),
									SpacesBefore: 0,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenEqual,
									Bytes:        []byte(`=`),
									SpacesBefore: 1,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenNumberLit,
									Bytes:        []byte(`1`),
									SpacesBefore: 1,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenNewline,
									Bytes:        []byte{'\n'},
									SpacesBefore: 0,
								},
							},
						},
					},
				},
			},
		},
		{
			"a = 1 # because it is\n",
			&Body{
				Items: []Node{
					&Attribute{
						AllTokens: &TokenSeq{
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenIdent,
										Bytes:        []byte(`a`),
										SpacesBefore: 0,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenEqual,
										Bytes:        []byte(`=`),
										SpacesBefore: 1,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenNumberLit,
										Bytes:        []byte(`1`),
										SpacesBefore: 1,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenComment,
										Bytes:        []byte("# because it is\n"),
										SpacesBefore: 1,
									},
								},
							},
						},
						NameTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenIdent,
								Bytes:        []byte(`a`),
								SpacesBefore: 0,
							},
						}},
						EqualsTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenEqual,
								Bytes:        []byte(`=`),
								SpacesBefore: 1,
							},
						}},
						Expr: &Expression{
							AllTokens: &TokenSeq{Tokens{
								{
									Type:         hclsyntax.TokenNumberLit,
									Bytes:        []byte(`1`),
									SpacesBefore: 1,
								},
							}},
						},
						LineCommentTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenComment,
								Bytes:        []byte("# because it is\n"),
								SpacesBefore: 1,
							},
						}},
					},
				},
				AllTokens: &TokenSeq{
					&TokenSeq{
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenIdent,
									Bytes:        []byte(`a`),
									SpacesBefore: 0,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenEqual,
									Bytes:        []byte(`=`),
									SpacesBefore: 1,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenNumberLit,
									Bytes:        []byte(`1`),
									SpacesBefore: 1,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenComment,
									Bytes:        []byte("# because it is\n"),
									SpacesBefore: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			"# bee bee bee\n\nb = 1\n", // two newlines separate the comment from the attribute
			&Body{
				Items: []Node{
					&Attribute{
						AllTokens: &TokenSeq{
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenIdent,
										Bytes:        []byte(`b`),
										SpacesBefore: 0,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenEqual,
										Bytes:        []byte(`=`),
										SpacesBefore: 1,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenNumberLit,
										Bytes:        []byte(`1`),
										SpacesBefore: 1,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenNewline,
										Bytes:        []byte{'\n'},
										SpacesBefore: 0,
									},
								},
							},
						},
						NameTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenIdent,
								Bytes:        []byte(`b`),
								SpacesBefore: 0,
							},
						}},
						EqualsTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenEqual,
								Bytes:        []byte(`=`),
								SpacesBefore: 1,
							},
						}},
						Expr: &Expression{
							AllTokens: &TokenSeq{Tokens{
								{
									Type:         hclsyntax.TokenNumberLit,
									Bytes:        []byte(`1`),
									SpacesBefore: 1,
								},
							}},
						},
						EOLTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenNewline,
								Bytes:        []byte{'\n'},
								SpacesBefore: 0,
							},
						}},
					},
				},
				AllTokens: &TokenSeq{
					&TokenSeq{
						Tokens{
							{
								Type:         hclsyntax.TokenComment,
								Bytes:        []byte("# bee bee bee\n"),
								SpacesBefore: 0,
							},
							{
								Type:         hclsyntax.TokenNewline,
								Bytes:        []byte("\n"),
								SpacesBefore: 0,
							},
						},
					},
					&TokenSeq{
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenIdent,
									Bytes:        []byte(`b`),
									SpacesBefore: 0,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenEqual,
									Bytes:        []byte(`=`),
									SpacesBefore: 1,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenNumberLit,
									Bytes:        []byte(`1`),
									SpacesBefore: 1,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenNewline,
									Bytes:        []byte{'\n'},
									SpacesBefore: 0,
								},
							},
						},
					},
				},
			},
		},
		{
			"b {}\n",
			&Body{
				Items: []Node{
					&Block{
						AllTokens: &TokenSeq{
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenIdent,
										Bytes:        []byte(`b`),
										SpacesBefore: 0,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenOBrace,
										Bytes:        []byte(`{`),
										SpacesBefore: 1,
									},
								},
							},
							(*TokenSeq)(nil), // the empty body
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenCBrace,
										Bytes:        []byte(`}`),
										SpacesBefore: 0,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenNewline,
										Bytes:        []byte{'\n'},
										SpacesBefore: 0,
									},
								},
							},
						},
						TypeTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenIdent,
								Bytes:        []byte(`b`),
								SpacesBefore: 0,
							},
						}},
						OBraceTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenOBrace,
								Bytes:        []byte(`{`),
								SpacesBefore: 1,
							},
						}},
						Body: &Body{},
						CBraceTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenCBrace,
								Bytes:        []byte(`}`),
								SpacesBefore: 0,
							},
						}},
						EOLTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenNewline,
								Bytes:        []byte{'\n'},
								SpacesBefore: 0,
							},
						}},
					},
				},
				AllTokens: &TokenSeq{
					&TokenSeq{
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenIdent,
									Bytes:        []byte(`b`),
									SpacesBefore: 0,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenOBrace,
									Bytes:        []byte(`{`),
									SpacesBefore: 1,
								},
							},
						},
						(*TokenSeq)(nil), // the empty body
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenCBrace,
									Bytes:        []byte(`}`),
									SpacesBefore: 0,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenNewline,
									Bytes:        []byte{'\n'},
									SpacesBefore: 0,
								},
							},
						},
					},
				},
			},
		},
		{
			"b {\n  a = 1\n}\n",
			&Body{
				Items: []Node{
					&Block{
						AllTokens: &TokenSeq{
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenIdent,
										Bytes:        []byte(`b`),
										SpacesBefore: 0,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenOBrace,
										Bytes:        []byte(`{`),
										SpacesBefore: 1,
									},
								},
							},
							&TokenSeq{
								&TokenSeq{
									Tokens{
										{
											Type:         hclsyntax.TokenNewline,
											Bytes:        []byte{'\n'},
											SpacesBefore: 0,
										},
									},
								},
								&TokenSeq{
									&TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenIdent,
												Bytes:        []byte(`a`),
												SpacesBefore: 2,
											},
										},
									},
									&TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenEqual,
												Bytes:        []byte(`=`),
												SpacesBefore: 1,
											},
										},
									},
									&TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenNumberLit,
												Bytes:        []byte(`1`),
												SpacesBefore: 1,
											},
										},
									},
									&TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenNewline,
												Bytes:        []byte{'\n'},
												SpacesBefore: 0,
											},
										},
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenCBrace,
										Bytes:        []byte(`}`),
										SpacesBefore: 0,
									},
								},
							},
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenNewline,
										Bytes:        []byte{'\n'},
										SpacesBefore: 0,
									},
								},
							},
						},
						TypeTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenIdent,
								Bytes:        []byte(`b`),
								SpacesBefore: 0,
							},
						}},
						OBraceTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenOBrace,
								Bytes:        []byte(`{`),
								SpacesBefore: 1,
							},
						}},
						Body: &Body{
							Items: []Node{
								&Attribute{
									AllTokens: &TokenSeq{
										&TokenSeq{
											Tokens{
												{
													Type:         hclsyntax.TokenIdent,
													Bytes:        []byte(`a`),
													SpacesBefore: 2,
												},
											},
										},
										&TokenSeq{
											Tokens{
												{
													Type:         hclsyntax.TokenEqual,
													Bytes:        []byte(`=`),
													SpacesBefore: 1,
												},
											},
										},
										&TokenSeq{
											Tokens{
												{
													Type:         hclsyntax.TokenNumberLit,
													Bytes:        []byte(`1`),
													SpacesBefore: 1,
												},
											},
										},
										&TokenSeq{
											Tokens{
												{
													Type:         hclsyntax.TokenNewline,
													Bytes:        []byte{'\n'},
													SpacesBefore: 0,
												},
											},
										},
									},
									NameTokens: &TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenIdent,
												Bytes:        []byte(`a`),
												SpacesBefore: 2,
											},
										},
									},
									EqualsTokens: &TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenEqual,
												Bytes:        []byte(`=`),
												SpacesBefore: 1,
											},
										},
									},
									Expr: &Expression{
										AllTokens: &TokenSeq{
											Tokens{
												{
													Type:         hclsyntax.TokenNumberLit,
													Bytes:        []byte(`1`),
													SpacesBefore: 1,
												},
											},
										},
									},
									EOLTokens: &TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenNewline,
												Bytes:        []byte{'\n'},
												SpacesBefore: 0,
											},
										},
									},
								},
							},
							AllTokens: &TokenSeq{
								&TokenSeq{
									Tokens{
										{
											Type:         hclsyntax.TokenNewline,
											Bytes:        []byte{'\n'},
											SpacesBefore: 0,
										},
									},
								},
								&TokenSeq{
									&TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenIdent,
												Bytes:        []byte(`a`),
												SpacesBefore: 2,
											},
										},
									},
									&TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenEqual,
												Bytes:        []byte(`=`),
												SpacesBefore: 1,
											},
										},
									},
									&TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenNumberLit,
												Bytes:        []byte(`1`),
												SpacesBefore: 1,
											},
										},
									},
									&TokenSeq{
										Tokens{
											{
												Type:         hclsyntax.TokenNewline,
												Bytes:        []byte{'\n'},
												SpacesBefore: 0,
											},
										},
									},
								},
							},
						},
						CBraceTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenCBrace,
								Bytes:        []byte(`}`),
								SpacesBefore: 0,
							},
						}},
						EOLTokens: &TokenSeq{Tokens{
							{
								Type:         hclsyntax.TokenNewline,
								Bytes:        []byte{'\n'},
								SpacesBefore: 0,
							},
						}},
					},
				},
				AllTokens: &TokenSeq{
					&TokenSeq{
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenIdent,
									Bytes:        []byte(`b`),
									SpacesBefore: 0,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenOBrace,
									Bytes:        []byte(`{`),
									SpacesBefore: 1,
								},
							},
						},
						&TokenSeq{
							&TokenSeq{
								Tokens{
									{
										Type:         hclsyntax.TokenNewline,
										Bytes:        []byte{'\n'},
										SpacesBefore: 0,
									},
								},
							},
							&TokenSeq{
								&TokenSeq{
									Tokens{
										{
											Type:         hclsyntax.TokenIdent,
											Bytes:        []byte(`a`),
											SpacesBefore: 2,
										},
									},
								},
								&TokenSeq{
									Tokens{
										{
											Type:         hclsyntax.TokenEqual,
											Bytes:        []byte(`=`),
											SpacesBefore: 1,
										},
									},
								},
								&TokenSeq{
									Tokens{
										{
											Type:         hclsyntax.TokenNumberLit,
											Bytes:        []byte(`1`),
											SpacesBefore: 1,
										},
									},
								},
								&TokenSeq{
									Tokens{
										{
											Type:         hclsyntax.TokenNewline,
											Bytes:        []byte{'\n'},
											SpacesBefore: 0,
										},
									},
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenCBrace,
									Bytes:        []byte(`}`),
									SpacesBefore: 0,
								},
							},
						},
						&TokenSeq{
							Tokens{
								{
									Type:         hclsyntax.TokenNewline,
									Bytes:        []byte{'\n'},
									SpacesBefore: 0,
								},
							},
						},
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
		t.Run(test.src, func(t *testing.T) {
			file, diags := parse([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) > 0 {
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			got := file.Body

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(got, test.want)
				if diff != "" {
					t.Errorf(
						"wrong result\ninput: %s\ndiff:  %s",
						test.src,
						diff,
					)
				} else {
					t.Errorf(
						"wrong result\ninput: %s\ngot:   %s\nwant:  %s",
						test.src,
						spew.Sdump(got),
						spew.Sdump(test.want),
					)
				}
			}
		})
	}
}

func TestPartitionTokens(t *testing.T) {
	tests := []struct {
		tokens    hclsyntax.Tokens
		rng       hcl.Range
		wantStart int
		wantEnd   int
	}{
		{
			hclsyntax.Tokens{},
			hcl.Range{
				Start: hcl.Pos{Byte: 0},
				End:   hcl.Pos{Byte: 0},
			},
			0,
			0,
		},
		{
			hclsyntax.Tokens{
				{
					Type: hclsyntax.TokenIdent,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0},
						End:   hcl.Pos{Byte: 4},
					},
				},
			},
			hcl.Range{
				Start: hcl.Pos{Byte: 0},
				End:   hcl.Pos{Byte: 4},
			},
			0,
			1,
		},
		{
			hclsyntax.Tokens{
				{
					Type: hclsyntax.TokenIdent,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0},
						End:   hcl.Pos{Byte: 4},
					},
				},
				{
					Type: hclsyntax.TokenIdent,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4},
						End:   hcl.Pos{Byte: 8},
					},
				},
				{
					Type: hclsyntax.TokenIdent,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8},
						End:   hcl.Pos{Byte: 12},
					},
				},
			},
			hcl.Range{
				Start: hcl.Pos{Byte: 4},
				End:   hcl.Pos{Byte: 8},
			},
			1,
			2,
		},
		{
			hclsyntax.Tokens{
				{
					Type: hclsyntax.TokenIdent,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0},
						End:   hcl.Pos{Byte: 4},
					},
				},
				{
					Type: hclsyntax.TokenIdent,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4},
						End:   hcl.Pos{Byte: 8},
					},
				},
				{
					Type: hclsyntax.TokenIdent,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8},
						End:   hcl.Pos{Byte: 12},
					},
				},
			},
			hcl.Range{
				Start: hcl.Pos{Byte: 0},
				End:   hcl.Pos{Byte: 8},
			},
			0,
			2,
		},
		{
			hclsyntax.Tokens{
				{
					Type: hclsyntax.TokenIdent,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 0},
						End:   hcl.Pos{Byte: 4},
					},
				},
				{
					Type: hclsyntax.TokenIdent,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 4},
						End:   hcl.Pos{Byte: 8},
					},
				},
				{
					Type: hclsyntax.TokenIdent,
					Range: hcl.Range{
						Start: hcl.Pos{Byte: 8},
						End:   hcl.Pos{Byte: 12},
					},
				},
			},
			hcl.Range{
				Start: hcl.Pos{Byte: 4},
				End:   hcl.Pos{Byte: 12},
			},
			1,
			3,
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			gotStart, gotEnd := partitionTokens(test.tokens, test.rng)

			if gotStart != test.wantStart || gotEnd != test.wantEnd {
				t.Errorf(
					"wrong result\ntokens: %s\nrange: %#v\ngot:   %d, %d\nwant:  %d, %d",
					prettyConfig.Sprint(test.tokens), test.rng,
					gotStart, test.wantStart,
					gotEnd, test.wantEnd,
				)
			}
		})
	}
}

func TestPartitionLeadCommentTokens(t *testing.T) {
	tests := []struct {
		tokens    hclsyntax.Tokens
		wantStart int
	}{
		{
			hclsyntax.Tokens{},
			0,
		},
		{
			hclsyntax.Tokens{
				{
					Type: hclsyntax.TokenComment,
				},
			},
			0,
		},
		{
			hclsyntax.Tokens{
				{
					Type: hclsyntax.TokenComment,
				},
				{
					Type: hclsyntax.TokenComment,
				},
			},
			0,
		},
		{
			hclsyntax.Tokens{
				{
					Type: hclsyntax.TokenComment,
				},
				{
					Type: hclsyntax.TokenNewline,
				},
			},
			2,
		},
		{
			hclsyntax.Tokens{
				{
					Type: hclsyntax.TokenComment,
				},
				{
					Type: hclsyntax.TokenNewline,
				},
				{
					Type: hclsyntax.TokenComment,
				},
			},
			2,
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			gotStart := partitionLeadCommentTokens(test.tokens)

			if gotStart != test.wantStart {
				t.Errorf(
					"wrong result\ntokens: %s\ngot:   %d\nwant:  %d",
					prettyConfig.Sprint(test.tokens),
					gotStart, test.wantStart,
				)
			}
		})
	}
}

func TestLexConfig(t *testing.T) {
	tests := []struct {
		input string
		want  Tokens
	}{
		{
			`a  b `,
			Tokens{
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`a`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`b`),
					SpacesBefore: 2,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 1,
				},
			},
		},
		{
			`
foo "bar" "baz" {
    pizza = " cheese "
}
`,
			Tokens{
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte{'\n'},
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`foo`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenQuotedLit,
					Bytes:        []byte(`bar`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenQuotedLit,
					Bytes:        []byte(`baz`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenOBrace,
					Bytes:        []byte(`{`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte("\n"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(`pizza`),
					SpacesBefore: 4,
				},
				{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte(`=`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 1,
				},
				{
					Type:         hclsyntax.TokenQuotedLit,
					Bytes:        []byte(` cheese `),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCQuote,
					Bytes:        []byte(`"`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte("\n"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenCBrace,
					Bytes:        []byte(`}`),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte("\n"),
					SpacesBefore: 0,
				},
				{
					Type:         hclsyntax.TokenEOF,
					Bytes:        []byte{},
					SpacesBefore: 0,
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
			got := lexConfig([]byte(test.input))

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(test.want, got)
				t.Errorf(
					"wrong result\ninput: %s\ndiff:  %s", test.input, diff,
				)
			}
		})
	}
}

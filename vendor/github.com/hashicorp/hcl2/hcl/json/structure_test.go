package json

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/hcl"
)

func TestBodyPartialContent(t *testing.T) {
	tests := []struct {
		src       string
		schema    *hcl.BodySchema
		want      *hcl.BodyContent
		diagCount int
	}{
		{
			`{}`,
			&hcl.BodySchema{},
			&hcl.BodyContent{
				Attributes: map[string]*hcl.Attribute{},
				MissingItemRange: hcl.Range{
					Filename: "test.json",
					Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
					End:      hcl.Pos{Line: 1, Column: 3, Byte: 2},
				},
			},
			0,
		},
		{
			`{"//": "comment that should be ignored"}`,
			&hcl.BodySchema{},
			&hcl.BodyContent{
				Attributes: map[string]*hcl.Attribute{},
				MissingItemRange: hcl.Range{
					Filename: "test.json",
					Start:    hcl.Pos{Line: 1, Column: 40, Byte: 39},
					End:      hcl.Pos{Line: 1, Column: 41, Byte: 40},
				},
			},
			0,
		},
		{
			`{"name":"Ermintrude"}`,
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: "name",
					},
				},
			},
			&hcl.BodyContent{
				Attributes: map[string]*hcl.Attribute{
					"name": &hcl.Attribute{
						Name: "name",
						Expr: &expression{
							src: &stringVal{
								Value: "Ermintrude",
								SrcRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   8,
										Line:   1,
										Column: 9,
									},
									End: hcl.Pos{
										Byte:   20,
										Line:   1,
										Column: 21,
									},
								},
							},
						},
						Range: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: hcl.Pos{
								Byte:   20,
								Line:   1,
								Column: 21,
							},
						},
						NameRange: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: hcl.Pos{
								Byte:   7,
								Line:   1,
								Column: 8,
							},
						},
					},
				},
				MissingItemRange: hcl.Range{
					Filename: "test.json",
					Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
					End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
				},
			},
			0,
		},
		{
			`{"name":"Ermintrude"}`,
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name:     "name",
						Required: true,
					},
					{
						Name:     "age",
						Required: true,
					},
				},
			},
			&hcl.BodyContent{
				Attributes: map[string]*hcl.Attribute{
					"name": &hcl.Attribute{
						Name: "name",
						Expr: &expression{
							src: &stringVal{
								Value: "Ermintrude",
								SrcRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   8,
										Line:   1,
										Column: 9,
									},
									End: hcl.Pos{
										Byte:   20,
										Line:   1,
										Column: 21,
									},
								},
							},
						},
						Range: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: hcl.Pos{
								Byte:   20,
								Line:   1,
								Column: 21,
							},
						},
						NameRange: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: hcl.Pos{
								Byte:   7,
								Line:   1,
								Column: 8,
							},
						},
					},
				},
				MissingItemRange: hcl.Range{
					Filename: "test.json",
					Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
					End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
				},
			},
			1,
		},
		{
			`{"resource":{}}`,
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "resource",
					},
				},
			},
			&hcl.BodyContent{
				Attributes: map[string]*hcl.Attribute{},
				Blocks: hcl.Blocks{
					{
						Type:   "resource",
						Labels: []string{},
						Body: &body{
							obj: &objectVal{
								Attrs: map[string]*objectAttr{},
								SrcRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   12,
										Line:   1,
										Column: 13,
									},
									End: hcl.Pos{
										Byte:   14,
										Line:   1,
										Column: 15,
									},
								},
								OpenRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   12,
										Line:   1,
										Column: 13,
									},
									End: hcl.Pos{
										Byte:   13,
										Line:   1,
										Column: 14,
									},
								},
								CloseRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   13,
										Line:   1,
										Column: 14,
									},
									End: hcl.Pos{
										Byte:   14,
										Line:   1,
										Column: 15,
									},
								},
							},
						},

						DefRange: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   12,
								Line:   1,
								Column: 13,
							},
							End: hcl.Pos{
								Byte:   13,
								Line:   1,
								Column: 14,
							},
						},
						TypeRange: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: hcl.Pos{
								Byte:   11,
								Line:   1,
								Column: 12,
							},
						},
						LabelRanges: []hcl.Range{},
					},
				},
				MissingItemRange: hcl.Range{
					Filename: "test.json",
					Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
			0,
		},
		{
			`{"resource":[{},{}]}`,
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "resource",
					},
				},
			},
			&hcl.BodyContent{
				Attributes: map[string]*hcl.Attribute{},
				Blocks: hcl.Blocks{
					{
						Type:   "resource",
						Labels: []string{},
						Body: &body{
							obj: &objectVal{
								Attrs: map[string]*objectAttr{},
								SrcRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   13,
										Line:   1,
										Column: 14,
									},
									End: hcl.Pos{
										Byte:   15,
										Line:   1,
										Column: 16,
									},
								},
								OpenRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   13,
										Line:   1,
										Column: 14,
									},
									End: hcl.Pos{
										Byte:   14,
										Line:   1,
										Column: 15,
									},
								},
								CloseRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   14,
										Line:   1,
										Column: 15,
									},
									End: hcl.Pos{
										Byte:   15,
										Line:   1,
										Column: 16,
									},
								},
							},
						},

						DefRange: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   12,
								Line:   1,
								Column: 13,
							},
							End: hcl.Pos{
								Byte:   13,
								Line:   1,
								Column: 14,
							},
						},
						TypeRange: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: hcl.Pos{
								Byte:   11,
								Line:   1,
								Column: 12,
							},
						},
						LabelRanges: []hcl.Range{},
					},
					{
						Type:   "resource",
						Labels: []string{},
						Body: &body{
							obj: &objectVal{
								Attrs: map[string]*objectAttr{},
								SrcRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   16,
										Line:   1,
										Column: 17,
									},
									End: hcl.Pos{
										Byte:   18,
										Line:   1,
										Column: 19,
									},
								},
								OpenRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   16,
										Line:   1,
										Column: 17,
									},
									End: hcl.Pos{
										Byte:   17,
										Line:   1,
										Column: 18,
									},
								},
								CloseRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   17,
										Line:   1,
										Column: 18,
									},
									End: hcl.Pos{
										Byte:   18,
										Line:   1,
										Column: 19,
									},
								},
							},
						},

						DefRange: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   12,
								Line:   1,
								Column: 13,
							},
							End: hcl.Pos{
								Byte:   13,
								Line:   1,
								Column: 14,
							},
						},
						TypeRange: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: hcl.Pos{
								Byte:   11,
								Line:   1,
								Column: 12,
							},
						},
						LabelRanges: []hcl.Range{},
					},
				},
				MissingItemRange: hcl.Range{
					Filename: "test.json",
					Start:    hcl.Pos{Line: 1, Column: 20, Byte: 19},
					End:      hcl.Pos{Line: 1, Column: 21, Byte: 20},
				},
			},
			0,
		},
		{
			`{"resource":{"foo_instance":{"bar":{}}}}`,
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
					},
				},
			},
			&hcl.BodyContent{
				Attributes: map[string]*hcl.Attribute{},
				Blocks: hcl.Blocks{
					{
						Type:   "resource",
						Labels: []string{"foo_instance", "bar"},
						Body: &body{
							obj: &objectVal{
								Attrs: map[string]*objectAttr{},
								SrcRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   35,
										Line:   1,
										Column: 36,
									},
									End: hcl.Pos{
										Byte:   37,
										Line:   1,
										Column: 38,
									},
								},
								OpenRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   35,
										Line:   1,
										Column: 36,
									},
									End: hcl.Pos{
										Byte:   36,
										Line:   1,
										Column: 37,
									},
								},
								CloseRange: hcl.Range{
									Filename: "test.json",
									Start: hcl.Pos{
										Byte:   36,
										Line:   1,
										Column: 37,
									},
									End: hcl.Pos{
										Byte:   37,
										Line:   1,
										Column: 38,
									},
								},
							},
						},

						DefRange: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   35,
								Line:   1,
								Column: 36,
							},
							End: hcl.Pos{
								Byte:   36,
								Line:   1,
								Column: 37,
							},
						},
						TypeRange: hcl.Range{
							Filename: "test.json",
							Start: hcl.Pos{
								Byte:   1,
								Line:   1,
								Column: 2,
							},
							End: hcl.Pos{
								Byte:   11,
								Line:   1,
								Column: 12,
							},
						},
						LabelRanges: []hcl.Range{
							{
								Filename: "test.json",
								Start: hcl.Pos{
									Byte:   13,
									Line:   1,
									Column: 14,
								},
								End: hcl.Pos{
									Byte:   27,
									Line:   1,
									Column: 28,
								},
							},
							{
								Filename: "test.json",
								Start: hcl.Pos{
									Byte:   29,
									Line:   1,
									Column: 30,
								},
								End: hcl.Pos{
									Byte:   34,
									Line:   1,
									Column: 35,
								},
							},
						},
					},
				},
				MissingItemRange: hcl.Range{
					Filename: "test.json",
					Start:    hcl.Pos{Line: 1, Column: 40, Byte: 39},
					End:      hcl.Pos{Line: 1, Column: 41, Byte: 40},
				},
			},
			0,
		},
		{
			`{"name":"Ermintrude"}`,
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "name",
					},
				},
			},
			&hcl.BodyContent{
				Attributes: map[string]*hcl.Attribute{},
				MissingItemRange: hcl.Range{
					Filename: "test.json",
					Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
					End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
				},
			},
			1,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%s", i, test.src), func(t *testing.T) {
			file, diags := Parse([]byte(test.src), "test.json")
			if len(diags) != 0 {
				t.Fatalf("Parse produced diagnostics: %s", diags)
			}
			got, _, diags := file.Body.PartialContent(test.schema)
			if len(diags) != test.diagCount {
				t.Errorf("Wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag)
				}
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.want))
			}
		})
	}
}

func TestBodyContent(t *testing.T) {
	// We test most of the functionality already in TestBodyPartialContent, so
	// this test focuses on the handling of extraneous attributes.
	tests := []struct {
		src       string
		schema    *hcl.BodySchema
		diagCount int
	}{
		{
			`{"unknown": true}`,
			&hcl.BodySchema{},
			1,
		},
		{
			`{"//": "comment that should be ignored"}`,
			&hcl.BodySchema{},
			0,
		},
		{
			`{"unknow": true}`,
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: "unknown",
					},
				},
			},
			1,
		},
		{
			`{"unknow": true, "unnown": true}`,
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: "unknown",
					},
				},
			},
			2,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%s", i, test.src), func(t *testing.T) {
			file, diags := Parse([]byte(test.src), "test.json")
			if len(diags) != 0 {
				t.Fatalf("Parse produced diagnostics: %s", diags)
			}
			_, diags = file.Body.Content(test.schema)
			if len(diags) != test.diagCount {
				t.Errorf("Wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag)
				}
			}
		})
	}
}

func TestJustAttributes(t *testing.T) {
	// We test most of the functionality already in TestBodyPartialContent, so
	// this test focuses on the handling of extraneous attributes.
	tests := []struct {
		src  string
		want hcl.Attributes
	}{
		{
			`{}`,
			map[string]*hcl.Attribute{},
		},
		{
			`{"foo": true}`,
			map[string]*hcl.Attribute{
				"foo": {
					Name: "foo",
					Expr: &expression{
						src: &booleanVal{
							Value: true,
							SrcRange: hcl.Range{
								Filename: "test.json",
								Start:    hcl.Pos{Byte: 8, Line: 1, Column: 9},
								End:      hcl.Pos{Byte: 12, Line: 1, Column: 13},
							},
						},
					},
					Range: hcl.Range{
						Filename: "test.json",
						Start:    hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:      hcl.Pos{Byte: 12, Line: 1, Column: 13},
					},
					NameRange: hcl.Range{
						Filename: "test.json",
						Start:    hcl.Pos{Byte: 1, Line: 1, Column: 2},
						End:      hcl.Pos{Byte: 6, Line: 1, Column: 7},
					},
				},
			},
		},
		{
			`{"//": "comment that should be ignored"}`,
			map[string]*hcl.Attribute{},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%s", i, test.src), func(t *testing.T) {
			file, diags := Parse([]byte(test.src), "test.json")
			if len(diags) != 0 {
				t.Fatalf("Parse produced diagnostics: %s", diags)
			}
			got, diags := file.Body.JustAttributes()
			if len(diags) != 0 {
				t.Errorf("Wrong number of diagnostics %d; want %d", len(diags), 0)
				for _, diag := range diags {
					t.Logf(" - %s", diag)
				}
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.want))
			}
		})
	}
}

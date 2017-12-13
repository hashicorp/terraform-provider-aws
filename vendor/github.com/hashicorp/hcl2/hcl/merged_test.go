package hcl

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestMergedBodiesContent(t *testing.T) {
	tests := []struct {
		Bodies    []Body
		Schema    *BodySchema
		Want      *BodyContent
		DiagCount int
	}{
		{
			[]Body{},
			&BodySchema{},
			&BodyContent{
				Attributes: map[string]*Attribute{},
			},
			0,
		},
		{
			[]Body{},
			&BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "name",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{},
			},
			0,
		},
		{
			[]Body{},
			&BodySchema{
				Attributes: []AttributeSchema{
					{
						Name:     "name",
						Required: true,
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{},
			},
			1,
		},
		{
			[]Body{
				&testMergedBodiesVictim{
					HasAttributes: []string{"name"},
				},
			},
			&BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "name",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{
					"name": &Attribute{
						Name: "name",
					},
				},
			},
			0,
		},
		{
			[]Body{
				&testMergedBodiesVictim{
					Name:          "first",
					HasAttributes: []string{"name"},
				},
				&testMergedBodiesVictim{
					Name:          "second",
					HasAttributes: []string{"name"},
				},
			},
			&BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "name",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{
					"name": &Attribute{
						Name:      "name",
						NameRange: Range{Filename: "first"},
					},
				},
			},
			1,
		},
		{
			[]Body{
				&testMergedBodiesVictim{
					Name:          "first",
					HasAttributes: []string{"name"},
				},
				&testMergedBodiesVictim{
					Name:          "second",
					HasAttributes: []string{"age"},
				},
			},
			&BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "name",
					},
					{
						Name: "age",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{
					"name": &Attribute{
						Name:      "name",
						NameRange: Range{Filename: "first"},
					},
					"age": &Attribute{
						Name:      "age",
						NameRange: Range{Filename: "second"},
					},
				},
			},
			0,
		},
		{
			[]Body{},
			&BodySchema{
				Blocks: []BlockHeaderSchema{
					{
						Type: "pizza",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{},
			},
			0,
		},
		{
			[]Body{
				&testMergedBodiesVictim{
					HasBlocks: map[string]int{
						"pizza": 1,
					},
				},
			},
			&BodySchema{
				Blocks: []BlockHeaderSchema{
					{
						Type: "pizza",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{},
				Blocks: Blocks{
					{
						Type: "pizza",
					},
				},
			},
			0,
		},
		{
			[]Body{
				&testMergedBodiesVictim{
					HasBlocks: map[string]int{
						"pizza": 2,
					},
				},
			},
			&BodySchema{
				Blocks: []BlockHeaderSchema{
					{
						Type: "pizza",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{},
				Blocks: Blocks{
					{
						Type: "pizza",
					},
					{
						Type: "pizza",
					},
				},
			},
			0,
		},
		{
			[]Body{
				&testMergedBodiesVictim{
					Name: "first",
					HasBlocks: map[string]int{
						"pizza": 1,
					},
				},
				&testMergedBodiesVictim{
					Name: "second",
					HasBlocks: map[string]int{
						"pizza": 1,
					},
				},
			},
			&BodySchema{
				Blocks: []BlockHeaderSchema{
					{
						Type: "pizza",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{},
				Blocks: Blocks{
					{
						Type:     "pizza",
						DefRange: Range{Filename: "first"},
					},
					{
						Type:     "pizza",
						DefRange: Range{Filename: "second"},
					},
				},
			},
			0,
		},
		{
			[]Body{
				&testMergedBodiesVictim{
					Name: "first",
				},
				&testMergedBodiesVictim{
					Name: "second",
					HasBlocks: map[string]int{
						"pizza": 2,
					},
				},
			},
			&BodySchema{
				Blocks: []BlockHeaderSchema{
					{
						Type: "pizza",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{},
				Blocks: Blocks{
					{
						Type:     "pizza",
						DefRange: Range{Filename: "second"},
					},
					{
						Type:     "pizza",
						DefRange: Range{Filename: "second"},
					},
				},
			},
			0,
		},
		{
			[]Body{
				&testMergedBodiesVictim{
					Name: "first",
					HasBlocks: map[string]int{
						"pizza": 2,
					},
				},
				&testMergedBodiesVictim{
					Name: "second",
				},
			},
			&BodySchema{
				Blocks: []BlockHeaderSchema{
					{
						Type: "pizza",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{},
				Blocks: Blocks{
					{
						Type:     "pizza",
						DefRange: Range{Filename: "first"},
					},
					{
						Type:     "pizza",
						DefRange: Range{Filename: "first"},
					},
				},
			},
			0,
		},
		{
			[]Body{
				&testMergedBodiesVictim{
					Name: "first",
				},
				&testMergedBodiesVictim{
					Name: "second",
				},
			},
			&BodySchema{
				Blocks: []BlockHeaderSchema{
					{
						Type: "pizza",
					},
				},
			},
			&BodyContent{
				Attributes: map[string]*Attribute{},
			},
			0,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			merged := MergeBodies(test.Bodies)
			got, diags := merged.Content(test.Schema)

			if len(diags) != test.DiagCount {
				t.Errorf("Wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag)
				}
			}

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.Want))
			}
		})
	}
}

type testMergedBodiesVictim struct {
	Name          string
	HasAttributes []string
	HasBlocks     map[string]int
	DiagCount     int
}

func (v *testMergedBodiesVictim) Content(schema *BodySchema) (*BodyContent, Diagnostics) {
	c, _, d := v.PartialContent(schema)
	return c, d
}

func (v *testMergedBodiesVictim) PartialContent(schema *BodySchema) (*BodyContent, Body, Diagnostics) {
	hasAttrs := map[string]struct{}{}
	for _, n := range v.HasAttributes {
		hasAttrs[n] = struct{}{}
	}

	content := &BodyContent{
		Attributes: map[string]*Attribute{},
	}

	rng := Range{
		Filename: v.Name,
	}

	for _, attrS := range schema.Attributes {
		_, has := hasAttrs[attrS.Name]
		if has {
			content.Attributes[attrS.Name] = &Attribute{
				Name:      attrS.Name,
				NameRange: rng,
			}
		}
	}

	if v.HasBlocks != nil {
		for _, blockS := range schema.Blocks {
			num := v.HasBlocks[blockS.Type]
			for i := 0; i < num; i++ {
				content.Blocks = append(content.Blocks, &Block{
					Type:     blockS.Type,
					DefRange: rng,
				})
			}
		}
	}

	diags := make(Diagnostics, v.DiagCount)
	for i := range diags {
		diags[i] = &Diagnostic{
			Severity: DiagError,
			Summary:  fmt.Sprintf("Fake diagnostic %d", i),
			Detail:   "For testing only.",
			Context:  &rng,
		}
	}

	return content, emptyBody, diags
}

func (v *testMergedBodiesVictim) JustAttributes() (Attributes, Diagnostics) {
	attrs := make(map[string]*Attribute)

	rng := Range{
		Filename: v.Name,
	}

	for _, name := range v.HasAttributes {
		attrs[name] = &Attribute{
			Name:      name,
			NameRange: rng,
		}
	}

	diags := make(Diagnostics, v.DiagCount)
	for i := range diags {
		diags[i] = &Diagnostic{
			Severity: DiagError,
			Summary:  fmt.Sprintf("Fake diagnostic %d", i),
			Detail:   "For testing only.",
			Context:  &rng,
		}
	}

	return attrs, diags
}

func (v *testMergedBodiesVictim) MissingItemRange() Range {
	return Range{
		Filename: v.Name,
	}
}

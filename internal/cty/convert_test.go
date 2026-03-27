// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfcty "github.com/hashicorp/terraform-provider-aws/internal/cty"
)

// Cribbed from
// - github.com/hashicorp/terraform-plugin-sdk/internal/plugin/convert/value_test.go

func TestPrimitiveTfType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Value cty.Value
		Want  tftypes.Value
	}{
		{
			Value: cty.StringVal("test"),
			Want:  tftypes.NewValue(tftypes.String, "test"),
		},
		{
			Value: cty.BoolVal(true),
			Want:  tftypes.NewValue(tftypes.Bool, true),
		},
		{
			Value: cty.NumberIntVal(42),
			Want:  tftypes.NewValue(tftypes.Number, 42),
		},
		{
			Value: cty.NumberFloatVal(3.14),
			Want:  tftypes.NewValue(tftypes.Number, 3.14),
		},
		{
			Value: cty.NullVal(cty.String),
			Want:  tftypes.NewValue(tftypes.String, nil),
		},
		{
			Value: cty.UnknownVal(cty.String),
			Want:  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
	}

	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			t.Parallel()

			got, err := tfcty.ToTfValue(test.Value)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(test.Want, *got); diff != "" {
				t.Errorf("unexpected differences: %s", diff)
			}
		})
	}
}

func TestListTfType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Value cty.Value
		Want  tftypes.Value
	}{
		{
			Value: cty.ListVal([]cty.Value{
				cty.StringVal("apple"),
				cty.StringVal("cherry"),
				cty.StringVal("kangaroo"),
			}),
			Want: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "apple"),
				tftypes.NewValue(tftypes.String, "cherry"),
				tftypes.NewValue(tftypes.String, "kangaroo"),
			}),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.BoolVal(true),
				cty.BoolVal(false),
			}),
			Want: tftypes.NewValue(tftypes.List{ElementType: tftypes.Bool}, []tftypes.Value{
				tftypes.NewValue(tftypes.Bool, true),
				tftypes.NewValue(tftypes.Bool, false),
			}),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.NumberIntVal(100),
				cty.NumberIntVal(200),
			}),
			Want: tftypes.NewValue(tftypes.List{ElementType: tftypes.Number}, []tftypes.Value{
				tftypes.NewValue(tftypes.Number, 100),
				tftypes.NewValue(tftypes.Number, 200),
			}),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"name":   cty.StringVal("Alice"),
					"breed":  cty.StringVal("Beagle"),
					"weight": cty.NumberIntVal(20),
					"toys":   cty.ListVal([]cty.Value{cty.StringVal("ball"), cty.StringVal("rope")}),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"name":   cty.StringVal("Bobby"),
					"breed":  cty.StringVal("Golden"),
					"weight": cty.NumberIntVal(30),
					"toys":   cty.ListVal([]cty.Value{cty.StringVal("dummy"), cty.StringVal("frisbee")}),
				}),
			}),
			Want: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":   tftypes.String,
						"breed":  tftypes.String,
						"weight": tftypes.Number,
						"toys":   tftypes.List{ElementType: tftypes.String},
					},
				},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":   tftypes.String,
						"breed":  tftypes.String,
						"weight": tftypes.Number,
						"toys":   tftypes.List{ElementType: tftypes.String},
					},
				}, map[string]tftypes.Value{
					"name":   tftypes.NewValue(tftypes.String, "Alice"),
					"breed":  tftypes.NewValue(tftypes.String, "Beagle"),
					"weight": tftypes.NewValue(tftypes.Number, 20),
					"toys": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "ball"),
						tftypes.NewValue(tftypes.String, "rope"),
					}),
				}),
				tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":   tftypes.String,
						"breed":  tftypes.String,
						"weight": tftypes.Number,
						"toys":   tftypes.List{ElementType: tftypes.String},
					},
				}, map[string]tftypes.Value{
					"name":   tftypes.NewValue(tftypes.String, "Bobby"),
					"breed":  tftypes.NewValue(tftypes.String, "Golden"),
					"weight": tftypes.NewValue(tftypes.Number, 30),
					"toys": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "dummy"),
						tftypes.NewValue(tftypes.String, "frisbee"),
					}),
				}),
			}),
		},
		{
			Value: cty.NullVal(cty.List(cty.String)),
			Want:  tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		},
		{
			Value: cty.UnknownVal(cty.List(cty.String)),
			Want:  tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
	}

	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			t.Parallel()

			got, err := tfcty.ToTfValue(test.Value)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(test.Want, *got); diff != "" {
				t.Errorf("unexpected differences: %s", diff)
			}
		})
	}
}

func TestSetTfType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Value cty.Value
		Want  tftypes.Value
	}{
		{
			Value: cty.SetVal([]cty.Value{
				cty.StringVal("apple"),
				cty.StringVal("cherry"),
				cty.StringVal("kangaroo"),
			}),
			Want: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "apple"),
				tftypes.NewValue(tftypes.String, "cherry"),
				tftypes.NewValue(tftypes.String, "kangaroo"),
			}),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.BoolVal(true),
				cty.BoolVal(false),
			}),
			Want: tftypes.NewValue(tftypes.Set{ElementType: tftypes.Bool}, []tftypes.Value{
				tftypes.NewValue(tftypes.Bool, true),
				tftypes.NewValue(tftypes.Bool, false),
			}),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.NumberIntVal(100),
				cty.NumberIntVal(200),
			}),
			Want: tftypes.NewValue(tftypes.Set{ElementType: tftypes.Number}, []tftypes.Value{
				tftypes.NewValue(tftypes.Number, 100),
				tftypes.NewValue(tftypes.Number, 200),
			}),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"name":   cty.StringVal("Alice"),
					"breed":  cty.StringVal("Beagle"),
					"weight": cty.NumberIntVal(20),
					"toys":   cty.SetVal([]cty.Value{cty.StringVal("ball"), cty.StringVal("rope")}),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"name":   cty.StringVal("Bobby"),
					"breed":  cty.StringVal("Golden"),
					"weight": cty.NumberIntVal(30),
					"toys":   cty.SetVal([]cty.Value{cty.StringVal("dummy"), cty.StringVal("frisbee")}),
				}),
			}),
			Want: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":   tftypes.String,
						"breed":  tftypes.String,
						"weight": tftypes.Number,
						"toys":   tftypes.Set{ElementType: tftypes.String},
					},
				},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":   tftypes.String,
						"breed":  tftypes.String,
						"weight": tftypes.Number,
						"toys":   tftypes.Set{ElementType: tftypes.String},
					},
				}, map[string]tftypes.Value{
					"name":   tftypes.NewValue(tftypes.String, "Alice"),
					"breed":  tftypes.NewValue(tftypes.String, "Beagle"),
					"weight": tftypes.NewValue(tftypes.Number, 20),
					"toys": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "ball"),
						tftypes.NewValue(tftypes.String, "rope"),
					}),
				}),
				tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":   tftypes.String,
						"breed":  tftypes.String,
						"weight": tftypes.Number,
						"toys":   tftypes.Set{ElementType: tftypes.String},
					},
				}, map[string]tftypes.Value{
					"name":   tftypes.NewValue(tftypes.String, "Bobby"),
					"breed":  tftypes.NewValue(tftypes.String, "Golden"),
					"weight": tftypes.NewValue(tftypes.Number, 30),
					"toys": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "dummy"),
						tftypes.NewValue(tftypes.String, "frisbee"),
					}),
				}),
			}),
		},
		{
			Value: cty.NullVal(cty.Set(cty.String)),
			Want:  tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
		},
		{
			Value: cty.UnknownVal(cty.Set(cty.String)),
			Want:  tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
	}

	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			t.Parallel()

			got, err := tfcty.ToTfValue(test.Value)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(test.Want, *got); diff != "" {
				t.Errorf("unexpected differences: %s", diff)
			}
		})
	}
}

func TestMapTfType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Value cty.Value
		Want  tftypes.Value
	}{
		{
			Value: cty.MapVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
				"baz": cty.StringVal("qux"),
			}),
			Want: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "bar"),
				"baz": tftypes.NewValue(tftypes.String, "qux"),
			}),
		},
		{
			Value: cty.MapVal(map[string]cty.Value{
				"foo": cty.MapVal(map[string]cty.Value{
					"foo": cty.StringVal("bar"),
					"baz": cty.StringVal("qux"),
				}),
			}),
			Want: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Map{
					ElementType: tftypes.String,
				}}, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"foo": tftypes.NewValue(tftypes.String, "bar"),
					"baz": tftypes.NewValue(tftypes.String, "qux"),
				}),
			}),
		},
		{
			Value: cty.MapVal(map[string]cty.Value{
				"foo": cty.ObjectVal(map[string]cty.Value{
					"fruits": cty.MapVal(map[string]cty.Value{
						"ananas":   cty.StringVal("pineapple"),
						"erdbeere": cty.StringVal("strawberry"),
					}),
				}),
			}),
			Want: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"fruits": tftypes.Map{ElementType: tftypes.String},
					}}}, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"fruits": tftypes.Map{ElementType: tftypes.String},
					}}, map[string]tftypes.Value{
					"fruits": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
						"ananas":   tftypes.NewValue(tftypes.String, "pineapple"),
						"erdbeere": tftypes.NewValue(tftypes.String, "strawberry"),
					}),
				}),
			}),
		},
		{
			Value: cty.NullVal(cty.Map(cty.String)),
			Want:  tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		},
		{
			Value: cty.UnknownVal(cty.Map(cty.String)),
			Want:  tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
	}

	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			t.Parallel()

			got, err := tfcty.ToTfValue(test.Value)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(test.Want, *got); diff != "" {
				t.Errorf("unexpected differences: %s", diff)
			}
		})
	}
}

func TestTupleTfType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Value cty.Value
		Want  tftypes.Value
	}{
		{
			Value: cty.TupleVal([]cty.Value{cty.StringVal("one")}),
			Want: tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String}}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "one"),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.StringVal("apple"),
				cty.NumberIntVal(5),
				cty.TupleVal([]cty.Value{cty.StringVal("banana"), cty.StringVal("pineapple")}),
			}),
			Want: tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.Number, tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.String}}}}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "apple"),
				tftypes.NewValue(tftypes.Number, 5),
				tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.String}}, []tftypes.Value{tftypes.NewValue(tftypes.String, "banana"), tftypes.NewValue(tftypes.String, "pineapple")}),
			}),
		},
		{
			Value: cty.NullVal(cty.Tuple([]cty.Type{cty.String})),
			Want:  tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String}}, nil),
		},
		{
			Value: cty.UnknownVal(cty.Tuple([]cty.Type{cty.String})),
			Want:  tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String}}, tftypes.UnknownValue),
		},
	}

	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			t.Parallel()

			got, err := tfcty.ToTfValue(test.Value)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(test.Want, *got); diff != "" {
				t.Errorf("unexpected differences: %s", diff)
			}
		})
	}
}

func TestObjectTfType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Value cty.Value
		Want  tftypes.Value
	}{
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"name":   cty.StringVal("Alice"),
				"breed":  cty.StringVal("Beagle"),
				"weight": cty.NumberIntVal(20),
			}),
			Want: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"name":   tftypes.String,
				"breed":  tftypes.String,
				"weight": tftypes.Number,
			}}, map[string]tftypes.Value{
				"name":   tftypes.NewValue(tftypes.String, "Alice"),
				"breed":  tftypes.NewValue(tftypes.String, "Beagle"),
				"weight": tftypes.NewValue(tftypes.Number, 20),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"chonk": cty.ObjectVal(map[string]cty.Value{
					"size":   cty.StringVal("large"),
					"weight": cty.NumberIntVal(50),
				}),
				"blep": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"color": cty.StringVal("brown"),
						"pattern": cty.ObjectVal(map[string]cty.Value{
							"style": cty.ListVal([]cty.Value{cty.StringVal("striped"), cty.StringVal("spotted")}),
						}),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"color": cty.StringVal("black"),
						"pattern": cty.ObjectVal(map[string]cty.Value{
							"style": cty.ListVal([]cty.Value{cty.StringVal("dotted"), cty.StringVal("plain")}),
						}),
					}),
				}),
			}),
			Want: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"chonk": tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.String,
						"weight": tftypes.Number,
					},
				},
				"blep": tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"color": tftypes.String,
						"pattern": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"style": tftypes.List{ElementType: tftypes.String},
							},
						},
					},
				}},
			}}, map[string]tftypes.Value{
				"chonk": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.String,
						"weight": tftypes.Number,
					}}, map[string]tftypes.Value{
					"size":   tftypes.NewValue(tftypes.String, "large"),
					"weight": tftypes.NewValue(tftypes.Number, 50),
				}),
				"blep": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"color": tftypes.String,
						"pattern": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"style": tftypes.List{ElementType: tftypes.String},
							},
						},
					}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"color": tftypes.String,
							"pattern": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"style": tftypes.List{ElementType: tftypes.String},
								},
							},
						},
					}, map[string]tftypes.Value{
						"color": tftypes.NewValue(tftypes.String, "brown"),
						"pattern": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"style": tftypes.List{ElementType: tftypes.String},
							},
						}, map[string]tftypes.Value{
							"style": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
								tftypes.NewValue(tftypes.String, "striped"),
								tftypes.NewValue(tftypes.String, "spotted"),
							}),
						}),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"color": tftypes.String,
							"pattern": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"style": tftypes.List{ElementType: tftypes.String},
								},
							},
						},
					}, map[string]tftypes.Value{
						"color": tftypes.NewValue(tftypes.String, "black"),
						"pattern": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"style": tftypes.List{ElementType: tftypes.String},
							},
						}, map[string]tftypes.Value{
							"style": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
								tftypes.NewValue(tftypes.String, "dotted"),
								tftypes.NewValue(tftypes.String, "plain"),
							}),
						}),
					}),
				}),
			}),
		},
		{
			Value: cty.NullVal(cty.Object(map[string]cty.Type{
				"foo": cty.String,
				"bar": cty.Number,
			})),
			Want: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"foo": tftypes.String,
				"bar": tftypes.Number,
			}}, nil),
		},
		{
			Value: cty.UnknownVal(cty.Object(map[string]cty.Type{
				"foo": cty.String,
				"bar": cty.Number,
			})),
			Want: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"foo": tftypes.String,
				"bar": tftypes.Number,
			}}, tftypes.UnknownValue),
		},
	}

	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			t.Parallel()

			got, err := tfcty.ToTfValue(test.Value)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(test.Want, *got); diff != "" {
				t.Errorf("unexpected differences: %s", diff)
			}
		})
	}
}

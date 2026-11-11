// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestClooudFormationJSONFlexer_Expand(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		Plan           tfsdk.Plan
		TfToCfnNameMap map[string]string
		ExpectedError  bool
		ExpectedState  map[string]any
	}{
		{
			TestName:       "simple Plan",
			Plan:           makeSimpleTestPlan(),
			TfToCfnNameMap: simpleTfToCfnNameMap,
			ExpectedState: map[string]any{
				"Name": "testing",
			},
		},
		{
			TestName:       "simple Plan with Optional",
			Plan:           makeSimpleTestPlanWithOptionalPopulated(),
			TfToCfnNameMap: simpleTfToCfnNameMap,
			ExpectedState: map[string]any{
				"Name":   "testing",
				"Number": float64(42),
			},
		},
		{
			TestName:       "complex Plan",
			Plan:           makeComplexTestPlan(),
			TfToCfnNameMap: complexTfToCfnNameMap,
			ExpectedState: map[string]any{
				"Name":        "hello, world",
				"MachineType": "e2-medium",
				"Ports":       []any{float64(80), float64(443)},
				"Tags":        []any{"red", "blue", "green"},
				"Disks": []any{
					map[string]any{
						"Id":                 "disk0",
						"DeleteWithInstance": true,
					},
					map[string]any{
						"Id":                 "disk1",
						"DeleteWithInstance": false,
					},
				},
				"BootDisk": map[string]any{
					"Id":                 "bootdisk",
					"DeleteWithInstance": true,
				},
				"ScratchDisk": map[string]any{
					"Interface": "SCSI",
				},
			},
		},
		{
			TestName:       "maps Plan",
			Plan:           makeMapsTestPlan(),
			TfToCfnNameMap: mapsTfToCfnNameMap,
			ExpectedState: map[string]any{
				"Name": "testing",
				"SimpleMap": map[string]any{
					"one": "eno",
					"two": "owt",
				},
				"ComplexMap": map[string]any{
					"x": map[string]any{
						"Id":    float64(1),
						"Flags": []any{true, false},
					},
					"y": map[string]any{
						"Id":    float64(-1),
						"Flags": []any{false, true, true},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			flexer := NewCloudFormationJSONFlexer(testCase.Plan.Schema, testCase.TfToCfnNameMap)
			got, err := flexer.toRawMap(t.Context(), testCase.Plan.Raw)

			if err == nil && testCase.ExpectedError {
				t.Fatalf("expected error")
			}

			if err != nil && !testCase.ExpectedError {
				t.Fatalf("unexpected error: %s", err)
			}

			if err == nil {
				if diff := cmp.Diff(got, testCase.ExpectedState); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}

func TestCloudFormationJSONFlexer_Flatten(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		Schema         schema.Schema
		TfToCfnNameMap map[string]string
		ResourceModel  map[string]any
		ExpectedError  bool
		ExpectedValue  tftypes.Value
	}{
		{
			TestName:       "simple State",
			Schema:         testSimpleSchema,
			TfToCfnNameMap: simpleTfToCfnNameMap,
			ResourceModel: map[string]any{
				"Arn":    "arn:aws:test:::test",
				"Name":   "testing",
				"Number": float64(42),
			},
			ExpectedValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"arn":        tftypes.String,
					"identifier": tftypes.String,
					"name":       tftypes.String,
					"number":     tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"arn":        tftypes.NewValue(tftypes.String, "arn:aws:test:::test"),
				"identifier": tftypes.NewValue(tftypes.String, nil),
				"name":       tftypes.NewValue(tftypes.String, "testing"),
				"number":     tftypes.NewValue(tftypes.Number, 42),
			}),
		},
		{
			TestName:       "simple State with extra field",
			Schema:         testSimpleSchema,
			TfToCfnNameMap: simpleTfToCfnNameMap,
			ResourceModel: map[string]any{
				"Arn":    "arn:aws:test:::test",
				"Height": float64(1.75),
				"Name":   "testing",
				"Number": float64(42),
			},
			ExpectedValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"arn":        tftypes.String,
					"identifier": tftypes.String,
					"name":       tftypes.String,
					"number":     tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"arn":        tftypes.NewValue(tftypes.String, "arn:aws:test:::test"),
				"identifier": tftypes.NewValue(tftypes.String, nil),
				"name":       tftypes.NewValue(tftypes.String, "testing"),
				"number":     tftypes.NewValue(tftypes.Number, 42),
			}),
		},
		{
			TestName:       "simple State with List",
			Schema:         testSimpleSchemaWithList,
			TfToCfnNameMap: simpleTfToCfnNameMap,
			ResourceModel: map[string]any{
				"Arn":    "arn:aws:test:::test",
				"Name":   "testing",
				"Number": float64(42),
				"Ports":  []any{float64(8080), float64(8443)},
			},
			ExpectedValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"arn":        tftypes.String,
					"identifier": tftypes.String,
					"name":       tftypes.String,
					"number":     tftypes.Number,
					"ports":      tftypes.List{ElementType: tftypes.Number},
				},
			}, map[string]tftypes.Value{
				"arn":        tftypes.NewValue(tftypes.String, "arn:aws:test:::test"),
				"identifier": tftypes.NewValue(tftypes.String, nil),
				"name":       tftypes.NewValue(tftypes.String, "testing"),
				"number":     tftypes.NewValue(tftypes.Number, 42),
				"ports": tftypes.NewValue(tftypes.List{ElementType: tftypes.Number}, []tftypes.Value{
					tftypes.NewValue(tftypes.Number, 8080),
					tftypes.NewValue(tftypes.Number, 8443),
				}),
			}),
		},
		{
			TestName:       "simple State with empty List",
			Schema:         testSimpleSchemaWithList,
			TfToCfnNameMap: simpleTfToCfnNameMap,
			ResourceModel: map[string]any{
				"Arn":    "arn:aws:test:::test",
				"Name":   "testing",
				"Number": float64(42),
				"Ports":  []any{},
			},
			ExpectedValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"arn":        tftypes.String,
					"identifier": tftypes.String,
					"name":       tftypes.String,
					"number":     tftypes.Number,
					"ports":      tftypes.List{ElementType: tftypes.Number},
				},
			}, map[string]tftypes.Value{
				"arn":        tftypes.NewValue(tftypes.String, "arn:aws:test:::test"),
				"identifier": tftypes.NewValue(tftypes.String, nil),
				"name":       tftypes.NewValue(tftypes.String, "testing"),
				"number":     tftypes.NewValue(tftypes.Number, 42),
				"ports":      tftypes.NewValue(tftypes.List{ElementType: tftypes.Number}, nil),
			}),
		},
		{
			TestName:       "simple State with missing List",
			Schema:         testSimpleSchemaWithList,
			TfToCfnNameMap: simpleTfToCfnNameMap,
			ResourceModel: map[string]any{
				"Arn":    "arn:aws:test:::test",
				"Name":   "testing",
				"Number": float64(42),
			},
			ExpectedValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"arn":        tftypes.String,
					"identifier": tftypes.String,
					"name":       tftypes.String,
					"number":     tftypes.Number,
					"ports":      tftypes.List{ElementType: tftypes.Number},
				},
			}, map[string]tftypes.Value{
				"arn":        tftypes.NewValue(tftypes.String, "arn:aws:test:::test"),
				"identifier": tftypes.NewValue(tftypes.String, nil),
				"name":       tftypes.NewValue(tftypes.String, "testing"),
				"number":     tftypes.NewValue(tftypes.Number, 42),
				"ports":      tftypes.NewValue(tftypes.List{ElementType: tftypes.Number}, nil),
			}),
		},
		{
			TestName:       "complex State",
			Schema:         testComplexSchema,
			TfToCfnNameMap: complexTfToCfnNameMap,
			ResourceModel: map[string]any{
				"Name":        "hello, world",
				"MachineType": "e2-medium",
				"Ports":       []any{float64(80), float64(443)},
				"Tags":        []any{"red", "blue", "green"},
				"Disks": []any{
					map[string]any{
						"Id":                 "disk0",
						"DeleteWithInstance": true,
					},
					map[string]any{
						"Id":                 "disk1",
						"DeleteWithInstance": false,
					},
				},
				"BootDisk": map[string]any{
					"Id":                 "bootdisk",
					"DeleteWithInstance": true,
				},
				"ScratchDisk": map[string]any{
					"Interface": "SCSI",
				},
				"VideoPorts": []any{
					map[string]any{
						"Id":    float64(1),
						"Flags": []any{true, false},
					},
					map[string]any{
						"Id":    float64(-1),
						"Flags": []any{false, true, true},
					},
				},
			},
			ExpectedValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name":         tftypes.String,
					"machine_type": tftypes.String,
					"ports":        tftypes.List{ElementType: tftypes.Number},
					"tags":         tftypes.Set{ElementType: tftypes.String},
					"disks": tftypes.List{
						ElementType: diskElementType,
					},
					"boot_disk": diskElementType,
					"scratch_disk": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
						},
					},
					"video_ports": tftypes.Set{
						ElementType: videoPortElementType,
					},
					"identifier": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name":         tftypes.NewValue(tftypes.String, "hello, world"),
				"machine_type": tftypes.NewValue(tftypes.String, "e2-medium"),
				"ports": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Number,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Number, 80),
					tftypes.NewValue(tftypes.Number, 443),
				}),
				"tags": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "blue"),
					tftypes.NewValue(tftypes.String, "green"),
				}),
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: diskElementType,
				}, []tftypes.Value{
					tftypes.NewValue(diskElementType, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "disk0"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(diskElementType, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "disk1"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"boot_disk": tftypes.NewValue(diskElementType, map[string]tftypes.Value{
					"id":                   tftypes.NewValue(tftypes.String, "bootdisk"),
					"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
				}),
				"scratch_disk": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"interface": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"interface": tftypes.NewValue(tftypes.String, "SCSI"),
				}),
				"video_ports": tftypes.NewValue(tftypes.Set{
					ElementType: videoPortElementType,
				}, []tftypes.Value{
					tftypes.NewValue(videoPortElementType, map[string]tftypes.Value{
						"id": tftypes.NewValue(tftypes.Number, 1),
						"flags": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.Bool,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Bool, true),
							tftypes.NewValue(tftypes.Bool, false),
						}),
					}),
					tftypes.NewValue(videoPortElementType, map[string]tftypes.Value{
						"id": tftypes.NewValue(tftypes.Number, -1),
						"flags": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.Bool,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Bool, false),
							tftypes.NewValue(tftypes.Bool, true),
							tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
				}),
				"identifier": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		{
			TestName:       "maps State",
			Schema:         testMapsSchema,
			TfToCfnNameMap: mapsTfToCfnNameMap,
			ResourceModel: map[string]any{
				"Name": "testing",
				"SimpleMap": map[string]any{
					"one": "eno",
					"two": "owt",
				},
				"ComplexMap": map[string]any{
					"x": map[string]any{
						"Id":    float64(1),
						"Flags": []any{true, false},
					},
					"y": map[string]any{
						"Id":    float64(-1),
						"Flags": []any{false, true, true},
					},
				},
			},
			ExpectedValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name":        tftypes.String,
					"simple_map":  tftypes.Map{ElementType: tftypes.String},
					"complex_map": tftypes.Map{ElementType: videoPortElementType},
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "testing"),
				"simple_map": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.String, "eno"),
					"two": tftypes.NewValue(tftypes.String, "owt"),
				}),
				"complex_map": tftypes.NewValue(tftypes.Map{
					ElementType: videoPortElementType,
				}, map[string]tftypes.Value{
					"x": tftypes.NewValue(videoPortElementType, map[string]tftypes.Value{
						"id": tftypes.NewValue(tftypes.Number, 1),
						"flags": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.Bool,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Bool, true),
							tftypes.NewValue(tftypes.Bool, false),
						}),
					}),
					"y": tftypes.NewValue(videoPortElementType, map[string]tftypes.Value{
						"id": tftypes.NewValue(tftypes.Number, -1),
						"flags": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.Bool,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Bool, false),
							tftypes.NewValue(tftypes.Bool, true),
							tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
				}),
			}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			flexer := NewCloudFormationJSONFlexer(testCase.Schema, testCase.TfToCfnNameMap)
			got, err := flexer.fromRawMap(t.Context(), testCase.ResourceModel)

			if err == nil && testCase.ExpectedError {
				t.Fatalf("expected error")
			}

			if err != nil && !testCase.ExpectedError {
				t.Fatalf("unexpected error: %s", err)
			}

			if err == nil {
				if diff := cmp.Diff(got, testCase.ExpectedValue); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}

var testSimpleSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"arn": schema.StringAttribute{
			Computed: true,
		},
		"identifier": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"number": schema.NumberAttribute{
			Optional: true,
		},
	},
}

var testSimpleSchemaWithList = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"arn": schema.StringAttribute{
			Computed: true,
		},
		"identifier": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"number": schema.NumberAttribute{
			Optional: true,
		},
		"ports": schema.ListAttribute{
			ElementType: types.NumberType,
			Optional:    true,
			Computed:    true,
		},
	},
}

func makeSimpleTestPlan() tfsdk.Plan {
	return tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"name": tftypes.NewValue(tftypes.String, "testing"),
		}),
		Schema: testSimpleSchema,
	}
}

func makeSimpleTestPlanWithOptionalPopulated() tfsdk.Plan {
	return tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name":   tftypes.String,
				"number": tftypes.Number,
			},
		}, map[string]tftypes.Value{
			"name":   tftypes.NewValue(tftypes.String, "testing"),
			"number": tftypes.NewValue(tftypes.Number, 42),
		}),
		Schema: testSimpleSchema,
	}
}

var simpleTfToCfnNameMap = map[string]string{
	"arn":        "Arn",
	"identifier": "Identifier",
	"name":       "Name",
	"number":     "Number",
	"ports":      "Ports",
}

var testComplexSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
		},
		"machine_type": schema.StringAttribute{
			Optional: true,
		},
		"ports": schema.ListAttribute{
			ElementType: types.NumberType,
			Required:    true,
		},
		"tags": schema.SetAttribute{
			ElementType: types.StringType,
			Required:    true,
		},
		"boot_disk": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required: true,
				},
				"delete_with_instance": schema.BoolAttribute{
					Optional: true,
					Computed: true,
				},
			},
		},
		"scratch_disk": schema.ObjectAttribute{
			AttributeTypes: map[string]attr.Type{
				"interface": types.StringType,
			},
			Optional: true,
		},
		"identifier": schema.StringAttribute{
			Computed: true,
		},
	},
	Blocks: map[string]schema.Block{
		"disks": schema.ListNestedBlock{
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Required: true,
					},
					"delete_with_instance": schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
		},
		"video_ports": schema.SetNestedBlock{
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.NumberAttribute{
						Required: true,
					},
					"flags": schema.ListAttribute{
						ElementType: types.BoolType,
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	},
}

func makeComplexTestPlan() tfsdk.Plan {
	return tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name":         tftypes.String,
				"machine_type": tftypes.String,
				"ports":        tftypes.List{ElementType: tftypes.Number},
				"tags":         tftypes.Set{ElementType: tftypes.String},
				"disks": tftypes.List{
					ElementType: diskElementType,
				},
				"boot_disk": diskElementType,
				"scratch_disk": tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"interface": tftypes.String,
					},
				},
			},
		}, map[string]tftypes.Value{
			"name":         tftypes.NewValue(tftypes.String, "hello, world"),
			"machine_type": tftypes.NewValue(tftypes.String, "e2-medium"),
			"ports": tftypes.NewValue(tftypes.List{
				ElementType: tftypes.Number,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Number, 80),
				tftypes.NewValue(tftypes.Number, 443),
			}),
			"tags": tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "red"),
				tftypes.NewValue(tftypes.String, "blue"),
				tftypes.NewValue(tftypes.String, "green"),
			}),
			"disks": tftypes.NewValue(tftypes.List{
				ElementType: diskElementType,
			}, []tftypes.Value{
				tftypes.NewValue(diskElementType, map[string]tftypes.Value{
					"id":                   tftypes.NewValue(tftypes.String, "disk0"),
					"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
				}),
				tftypes.NewValue(diskElementType, map[string]tftypes.Value{
					"id":                   tftypes.NewValue(tftypes.String, "disk1"),
					"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
				}),
			}),
			"boot_disk": tftypes.NewValue(diskElementType, map[string]tftypes.Value{
				"id":                   tftypes.NewValue(tftypes.String, "bootdisk"),
				"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
			}),
			"scratch_disk": tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"interface": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"interface": tftypes.NewValue(tftypes.String, "SCSI"),
			}),
		}),
		Schema: testComplexSchema,
	}
}

var complexTfToCfnNameMap = map[string]string{
	"boot_disk":            "BootDisk",
	"delete_with_instance": "DeleteWithInstance",
	"disks":                "Disks",
	"flags":                "Flags",
	"id":                   "Id",
	"identifier":           "Identifier",
	"interface":            "Interface",
	"machine_type":         "MachineType",
	"name":                 "Name",
	"ports":                "Ports",
	"scratch_disk":         "ScratchDisk",
	"tags":                 "Tags",
	"video_ports":          "VideoPorts",
}

// element type for the "disks" attribute, which is a list of disks.
// only used in "disks"
var diskElementType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":                   tftypes.String,
		"delete_with_instance": tftypes.Bool,
	},
}

var videoPortElementType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":    tftypes.Number,
		"flags": tftypes.List{ElementType: tftypes.Bool},
	},
}

var testMapsSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
		},
		"simple_map": schema.MapAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"complex_map": schema.MapNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.NumberAttribute{
						Required: true,
					},
					"flags": schema.ListAttribute{
						ElementType: types.BoolType,
						Optional:    true,
					},
				},
			},
			Optional: true,
		},
	},
}

func makeMapsTestPlan() tfsdk.Plan {
	return tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name":       tftypes.String,
				"simple_map": tftypes.Map{ElementType: tftypes.String},
				"complex_map": tftypes.Map{
					ElementType: videoPortElementType,
				},
			},
		}, map[string]tftypes.Value{
			"name": tftypes.NewValue(tftypes.String, "testing"),
			"simple_map": tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{
				"one": tftypes.NewValue(tftypes.String, "eno"),
				"two": tftypes.NewValue(tftypes.String, "owt"),
			}),
			"complex_map": tftypes.NewValue(tftypes.Map{
				ElementType: videoPortElementType,
			}, map[string]tftypes.Value{
				"x": tftypes.NewValue(videoPortElementType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.Number, 1),
					"flags": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Bool,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Bool, true),
						tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"y": tftypes.NewValue(videoPortElementType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.Number, -1),
					"flags": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Bool,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Bool, false),
						tftypes.NewValue(tftypes.Bool, true),
						tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
			}),
		}),
		Schema: testMapsSchema,
	}
}

var mapsTfToCfnNameMap = map[string]string{
	"flags":       "Flags",
	"id":          "Id",
	"name":        "Name",
	"simple_map":  "SimpleMap",
	"complex_map": "ComplexMap",
}

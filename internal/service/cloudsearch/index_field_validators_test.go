// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudsearch

import (
	"context"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudsearch/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestInvalidForFieldTypes(t *testing.T) {
	t.Parallel()

	// Define a minimal schema that matches the index_field structure
	// Must use CustomType for "type" to match the real schema
	indexFieldSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrType: schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.IndexFieldType](),
			},
			"facet": schema.BoolAttribute{
				Optional: true,
			},
			"search": schema.BoolAttribute{
				Optional: true,
			},
			"sort": schema.BoolAttribute{
				Optional: true,
			},
			"highlight": schema.BoolAttribute{
				Optional: true,
			},
		},
	}

	type testCase struct {
		attributeName string
		invalidTypes  []string
		fieldType     string
		value         types.Bool
		expectError   bool
	}

	testCases := map[string]testCase{
		"facet on text-array errors": {
			attributeName: "facet",
			invalidTypes:  []string{"text", "text-array"},
			fieldType:     "text-array",
			value:         types.BoolValue(true),
			expectError:   true,
		},
		"facet on literal no error": {
			attributeName: "facet",
			invalidTypes:  []string{"text", "text-array"},
			fieldType:     "literal",
			value:         types.BoolValue(true),
			expectError:   false,
		},
		"facet on text errors": {
			attributeName: "facet",
			invalidTypes:  []string{"text", "text-array"},
			fieldType:     "text",
			value:         types.BoolValue(true),
			expectError:   true,
		},
		"search on text errors": {
			attributeName: "search",
			invalidTypes:  []string{"text", "text-array"},
			fieldType:     "text",
			value:         types.BoolValue(true),
			expectError:   true,
		},
		"search on text-array errors": {
			attributeName: "search",
			invalidTypes:  []string{"text", "text-array"},
			fieldType:     "text-array",
			value:         types.BoolValue(true),
			expectError:   true,
		},
		"search on literal no error": {
			attributeName: "search",
			invalidTypes:  []string{"text", "text-array"},
			fieldType:     "literal",
			value:         types.BoolValue(true),
			expectError:   false,
		},
		"sort on int-array errors": {
			attributeName: "sort",
			invalidTypes:  []string{"int-array", "double-array", "literal-array", "date-array", "text-array"},
			fieldType:     "int-array",
			value:         types.BoolValue(true),
			expectError:   true,
		},
		"sort on double-array errors": {
			attributeName: "sort",
			invalidTypes:  []string{"int-array", "double-array", "literal-array", "date-array", "text-array"},
			fieldType:     "double-array",
			value:         types.BoolValue(true),
			expectError:   true,
		},
		"sort on int no error": {
			attributeName: "sort",
			invalidTypes:  []string{"int-array", "double-array", "literal-array", "date-array", "text-array"},
			fieldType:     "int",
			value:         types.BoolValue(true),
			expectError:   false,
		},
		"highlight on text no error": {
			attributeName: "highlight",
			invalidTypes:  []string{"literal", "literal-array", "int", "int-array", "double", "double-array", "date", "date-array", "latlon"},
			fieldType:     "text",
			value:         types.BoolValue(true),
			expectError:   false,
		},
		"highlight on text-array no error": {
			attributeName: "highlight",
			invalidTypes:  []string{"literal", "literal-array", "int", "int-array", "double", "double-array", "date", "date-array", "latlon"},
			fieldType:     "text-array",
			value:         types.BoolValue(true),
			expectError:   false,
		},
		"highlight on literal errors": {
			attributeName: "highlight",
			invalidTypes:  []string{"literal", "literal-array", "int", "int-array", "double", "double-array", "date", "date-array", "latlon"},
			fieldType:     "literal",
			value:         types.BoolValue(true),
			expectError:   true,
		},
		"highlight on int errors": {
			attributeName: "highlight",
			invalidTypes:  []string{"literal", "literal-array", "int", "int-array", "double", "double-array", "date", "date-array", "latlon"},
			fieldType:     "int",
			value:         types.BoolValue(true),
			expectError:   true,
		},
		"highlight on latlon errors": {
			attributeName: "highlight",
			invalidTypes:  []string{"literal", "literal-array", "int", "int-array", "double", "double-array", "date", "date-array", "latlon"},
			fieldType:     "latlon",
			value:         types.BoolValue(true),
			expectError:   true,
		},
		"null value skips validation": {
			attributeName: "facet",
			invalidTypes:  []string{"text", "text-array"},
			fieldType:     "text-array",
			value:         types.BoolNull(),
			expectError:   false,
		},
		"unknown value skips validation": {
			attributeName: "facet",
			invalidTypes:  []string{"text", "text-array"},
			fieldType:     "text-array",
			value:         types.BoolUnknown(),
			expectError:   false,
		},
		"false value on invalid type no error": {
			attributeName: "facet",
			invalidTypes:  []string{"text", "text-array"},
			fieldType:     "text-array",
			value:         types.BoolValue(false),
			expectError:   true, // Even false is invalid - the attribute shouldn't be set at all
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			v := invalidForFieldTypes(test.attributeName, test.invalidTypes...)

			// Create a mock config with the type field
			attrs := map[string]tftypes.Value{
				names.AttrName: tftypes.NewValue(tftypes.String, "test_field"),
				names.AttrType: tftypes.NewValue(tftypes.String, test.fieldType),
				"facet":        tftypes.NewValue(tftypes.Bool, nil),
				"search":       tftypes.NewValue(tftypes.Bool, nil),
				"sort":         tftypes.NewValue(tftypes.Bool, nil),
				"highlight":    tftypes.NewValue(tftypes.Bool, nil),
			}

			// Set the attribute being tested
			if !test.value.IsNull() && !test.value.IsUnknown() {
				attrs[test.attributeName] = tftypes.NewValue(tftypes.Bool, test.value.ValueBool())
			} else if test.value.IsUnknown() {
				attrs[test.attributeName] = tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
			}

			rawVal := tftypes.NewValue(indexFieldSchema.Type().TerraformType(ctx), attrs)

			config := tfsdk.Config{
				Raw:    rawVal,
				Schema: indexFieldSchema,
			}

			req := validator.BoolRequest{
				Path:        path.Root(test.attributeName),
				ConfigValue: test.value,
				Config:      config,
			}

			resp := &validator.BoolResponse{}
			v.ValidateBool(ctx, req, resp)

			if !resp.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if resp.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", resp.Diagnostics)
			}
		})
	}
}

func TestInvalidForFieldTypes_Description(t *testing.T) {
	t.Parallel()

	v := invalidForFieldTypes("facet", "text", "text-array")

	ctx := context.Background()
	desc := v.Description(ctx)
	if desc != "facet cannot be set for certain field types" {
		t.Errorf("unexpected description: %s", desc)
	}

	mdDesc := v.MarkdownDescription(ctx)
	if mdDesc != desc {
		t.Errorf("markdown description should match description: got %s, want %s", mdDesc, desc)
	}
}

// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudsearch

import (
	"context"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudsearch/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestInvalidForFieldTypes(t *testing.T) {
	t.Parallel()

	// Define a minimal schema that matches the index_field structure
	// Must use CustomType for "type" to match the real schema
	indexFieldSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
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
				"name":      tftypes.NewValue(tftypes.String, "test_field"),
				"type":      tftypes.NewValue(tftypes.String, test.fieldType),
				"facet":     tftypes.NewValue(tftypes.Bool, nil),
				"search":    tftypes.NewValue(tftypes.Bool, nil),
				"sort":      tftypes.NewValue(tftypes.Bool, nil),
				"highlight": tftypes.NewValue(tftypes.Bool, nil),
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

// TestAnalysisSchemeDefault tests the plan modifier that sets "_en_default_" for text fields
// when analysis_scheme is not configured. This prevents "does not correlate" errors because
// AWS returns "_en_default_" for text/text-array fields even when not configured.
func TestAnalysisSchemeDefault(t *testing.T) {
	t.Parallel()

	// Define a minimal schema that matches the index_field structure
	indexFieldSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.IndexFieldType](),
			},
			"analysis_scheme": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}

	type testCase struct {
		fieldType     string
		configValue   types.String // What user configured
		stateValue    types.String // What's in state
		expectedValue types.String // What plan should have
	}

	testCases := map[string]testCase{
		// Text fields without analysis_scheme should get "_en_default_"
		"text field null config gets default": {
			fieldType:     "text",
			configValue:   types.StringNull(),
			stateValue:    types.StringNull(),
			expectedValue: types.StringValue("_en_default_"),
		},
		"text-array field null config gets default": {
			fieldType:     "text-array",
			configValue:   types.StringNull(),
			stateValue:    types.StringNull(),
			expectedValue: types.StringValue("_en_default_"),
		},
		// Non-text fields should NOT get the default
		"literal field null config stays null": {
			fieldType:     "literal",
			configValue:   types.StringNull(),
			stateValue:    types.StringNull(),
			expectedValue: types.StringNull(),
		},
		"int field null config stays null": {
			fieldType:     "int",
			configValue:   types.StringNull(),
			stateValue:    types.StringNull(),
			expectedValue: types.StringNull(),
		},
		// Explicit config value should be preserved
		"text field with explicit config preserves value": {
			fieldType:     "text",
			configValue:   types.StringValue("custom_scheme"),
			stateValue:    types.StringNull(),
			expectedValue: types.StringValue("custom_scheme"),
		},
		// State value should be used when available (UseStateForUnknown behavior)
		"text field with state value uses state": {
			fieldType:     "text",
			configValue:   types.StringNull(),
			stateValue:    types.StringValue("_en_default_"),
			expectedValue: types.StringValue("_en_default_"),
		},
		"literal field with state value uses state": {
			fieldType:     "literal",
			configValue:   types.StringNull(),
			stateValue:    types.StringValue("some_scheme"),
			expectedValue: types.StringValue("some_scheme"),
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			modifier := AnalysisSchemeDefault()

			// Create a mock config with the type field
			attrs := map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "test_field"),
				"type": tftypes.NewValue(tftypes.String, test.fieldType),
			}

			// Set analysis_scheme based on config value
			if test.configValue.IsNull() {
				attrs["analysis_scheme"] = tftypes.NewValue(tftypes.String, nil)
			} else if test.configValue.IsUnknown() {
				attrs["analysis_scheme"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
			} else {
				attrs["analysis_scheme"] = tftypes.NewValue(tftypes.String, test.configValue.ValueString())
			}

			rawVal := tftypes.NewValue(indexFieldSchema.Type().TerraformType(ctx), attrs)

			config := tfsdk.Config{
				Raw:    rawVal,
				Schema: indexFieldSchema,
			}

			req := planmodifier.StringRequest{
				Path:        path.Root("analysis_scheme"),
				ConfigValue: test.configValue,
				StateValue:  test.stateValue,
				PlanValue:   test.configValue, // Plan starts with config value
				Config:      config,
			}

			resp := &planmodifier.StringResponse{
				PlanValue: req.PlanValue,
			}

			modifier.PlanModifyString(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %s", resp.Diagnostics)
			}

			// Compare the result
			if !resp.PlanValue.Equal(test.expectedValue) {
				t.Errorf("expected plan value %v, got %v", test.expectedValue, resp.PlanValue)
			}
		})
	}
}

func TestAnalysisSchemeDefault_Description(t *testing.T) {
	t.Parallel()

	modifier := AnalysisSchemeDefault()

	ctx := context.Background()
	desc := modifier.Description(ctx)
	if desc != "Sets default analysis_scheme for text fields" {
		t.Errorf("unexpected description: %s", desc)
	}

	mdDesc := modifier.MarkdownDescription(ctx)
	if mdDesc != desc {
		t.Errorf("markdown description should match description: got %s, want %s", mdDesc, desc)
	}
}

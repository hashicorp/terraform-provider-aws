// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package objectvalidator_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
)

var objectAttributeTypes = map[string]attr.Type{
	"a": types.StringType,
	"b": types.StringType,
}

func TestAtMostOneOfChildren(t *testing.T) {
	t.Parallel()

	testObjectValidator(t, tfobjectvalidator.AtMostOneOfChildren, map[string]objectValidatorTestCase{
		"all-children-null": {
			wantErrors: 0,
		},
		"one-child-set": {
			a:          "value",
			wantErrors: 0,
		},
		"multiple-children-set": {
			a:          "value-a",
			b:          "value-b",
			wantErrors: 1,
		},
		"child-unknown": {
			a:          tftypes.UnknownValue,
			wantErrors: 0,
		},
		"object-unknown": {
			objectUnknown: true,
			wantErrors:    0,
		},
	})
}

func TestExactlyOneOfChildren(t *testing.T) {
	t.Parallel()

	testObjectValidator(t, tfobjectvalidator.ExactlyOneOfChildren, map[string]objectValidatorTestCase{
		"all-children-null": {
			wantErrors: 1,
		},
		"one-child-set": {
			a:          "value",
			wantErrors: 0,
		},
		"multiple-children-set": {
			a:          "value-a",
			b:          "value-b",
			wantErrors: 1,
		},
		"child-unknown": {
			a:          tftypes.UnknownValue,
			wantErrors: 0,
		},
		"object-unknown": {
			objectUnknown: true,
			wantErrors:    0,
		},
		"self-reference-is-ignored": {
			a:           "value",
			expressions: []path.Expression{path.MatchRelative()},
			wantErrors:  1,
		},
	})
}

func TestWarnAtMostOneOfChildren(t *testing.T) {
	t.Parallel()

	testObjectValidator(t, tfobjectvalidator.WarnAtMostOneOfChildren, map[string]objectValidatorTestCase{
		"all-children-null": {
			wantWarnings: 0,
		},
		"one-child-set": {
			a:            "value",
			wantWarnings: 0,
		},
		"multiple-children-set": {
			a:            "value-a",
			b:            "value-b",
			wantWarnings: 1,
		},
		"child-unknown": {
			a:            tftypes.UnknownValue,
			wantWarnings: 0,
		},
		"object-unknown": {
			objectUnknown: true,
			wantWarnings:  0,
		},
	})
}

func TestWarnExactlyOneOfChildren(t *testing.T) {
	t.Parallel()

	testObjectValidator(t, tfobjectvalidator.WarnExactlyOneOfChildren, map[string]objectValidatorTestCase{
		"all-children-null": {
			wantWarnings: 1,
		},
		"one-child-set": {
			a:            "value",
			wantWarnings: 0,
		},
		"multiple-children-set": {
			a:            "value-a",
			b:            "value-b",
			wantWarnings: 1,
		},
		"child-unknown": {
			a:            tftypes.UnknownValue,
			wantWarnings: 0,
		},
		"object-unknown": {
			objectUnknown: true,
			wantWarnings:  0,
		},
		"self-reference-is-ignored": {
			a:            "value",
			expressions:  []path.Expression{path.MatchRelative()},
			wantWarnings: 1,
		},
	})
}

type objectValidatorTestCase struct {
	a             any
	b             any
	objectUnknown bool
	expressions   []path.Expression
	wantErrors    int
	wantWarnings  int
}

func testObjectValidator(t *testing.T, validatorFactory func(...path.Expression) validator.Object, testCases map[string]objectValidatorTestCase) {
	t.Helper()

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config, configValue := testObjectConfig(t, tc.a, tc.b, tc.objectUnknown)
			request := validator.ObjectRequest{
				Config:         config,
				ConfigValue:    configValue,
				Path:           path.Root("object"),
				PathExpression: path.MatchRoot("object"),
			}
			response := validator.ObjectResponse{}

			expressions := tc.expressions
			if expressions == nil {
				expressions = []path.Expression{
					path.MatchRelative().AtName("a"),
					path.MatchRelative().AtName("b"),
				}
			}

			validatorFactory(expressions...).ValidateObject(t.Context(), request, &response)

			if got := response.Diagnostics.ErrorsCount(); got != tc.wantErrors {
				t.Errorf("expected %d error(s), got %d: %s", tc.wantErrors, got, response.Diagnostics)
			}

			if got := response.Diagnostics.WarningsCount(); got != tc.wantWarnings {
				t.Errorf("expected %d warning(s), got %d: %s", tc.wantWarnings, got, response.Diagnostics)
			}
		})
	}
}

func testObjectConfig(t *testing.T, a, b any, objectUnknown bool) (tfsdk.Config, types.Object) {
	t.Helper()

	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
			"b": tftypes.String,
		},
	}
	rootType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"object": objectType,
		},
	}

	var configValue types.Object
	var objectRawValue tftypes.Value

	if objectUnknown {
		configValue = types.ObjectUnknown(objectAttributeTypes)
		objectRawValue = tftypes.NewValue(objectType, tftypes.UnknownValue)
	} else {
		configValue = types.ObjectValueMust(objectAttributeTypes, map[string]attr.Value{
			"a": testStringValue(a),
			"b": testStringValue(b),
		})
		objectRawValue = tftypes.NewValue(objectType, map[string]tftypes.Value{
			"a": tftypes.NewValue(tftypes.String, a),
			"b": tftypes.NewValue(tftypes.String, b),
		})
	}

	return tfsdk.Config{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"object": schema.SingleNestedAttribute{
					Optional: true,
					Attributes: map[string]schema.Attribute{
						"a": schema.StringAttribute{Optional: true},
						"b": schema.StringAttribute{Optional: true},
					},
				},
			},
		},
		Raw: tftypes.NewValue(rootType, map[string]tftypes.Value{
			"object": objectRawValue,
		}),
	}, configValue
}

func testStringValue(value any) types.String {
	if value == nil {
		return types.StringNull()
	}

	if value == tftypes.UnknownValue {
		return types.StringUnknown()
	}

	return types.StringValue(value.(string))
}

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
	"c": types.StringType,
}

func TestAtMostOneOfChildren(t *testing.T) {
	t.Parallel()

	testObjectValidatorCombinations(t, tfobjectvalidator.AtMostOneOfChildren, func(configured int) bool {
		return configured > 1
	}, false)

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

func TestAtLeastOneOfChildren(t *testing.T) {
	t.Parallel()

	testObjectValidatorCombinations(t, tfobjectvalidator.AtLeastOneOfChildren, func(configured int) bool {
		return configured == 0
	}, false)

	testObjectValidator(t, tfobjectvalidator.AtLeastOneOfChildren, map[string]objectValidatorTestCase{
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
			wantErrors: 0,
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

	testObjectValidatorCombinations(t, tfobjectvalidator.ExactlyOneOfChildren, func(configured int) bool {
		return configured != 1
	}, false)

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

func TestRequiredTogetherChildren(t *testing.T) {
	t.Parallel()

	testObjectValidatorCombinations(t, tfobjectvalidator.RequiredTogetherChildren, func(configured int) bool {
		return configured != 0 && configured != 3
	}, false)

	testObjectValidator(t, tfobjectvalidator.RequiredTogetherChildren, map[string]objectValidatorTestCase{
		"all-children-null": {
			wantErrors: 0,
		},
		"first-child-set": {
			a:          "value",
			wantErrors: 1,
		},
		"second-child-set": {
			b:          "value",
			wantErrors: 1,
		},
		"all-children-set": {
			a:          "value-a",
			b:          "value-b",
			c:          "value-c",
			wantErrors: 0,
		},
		"first-child-unknown-second-child-null": {
			a:          tftypes.UnknownValue,
			wantErrors: 0,
		},
		"first-child-null-second-child-unknown": {
			b:          tftypes.UnknownValue,
			wantErrors: 0,
		},
		"first-child-unknown-second-child-set": {
			a:          tftypes.UnknownValue,
			b:          "value",
			wantErrors: 0,
		},
		"first-child-set-second-child-unknown": {
			a:          "value",
			b:          tftypes.UnknownValue,
			wantErrors: 0,
		},
		"all-children-unknown": {
			a:          tftypes.UnknownValue,
			b:          tftypes.UnknownValue,
			c:          tftypes.UnknownValue,
			wantErrors: 0,
		},
		"object-unknown": {
			objectUnknown: true,
			wantErrors:    0,
		},
	})
}

func TestWarnAtMostOneOfChildren(t *testing.T) {
	t.Parallel()

	testObjectValidatorCombinations(t, tfobjectvalidator.WarnAtMostOneOfChildren, func(configured int) bool {
		return configured > 1
	}, true)

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

	testObjectValidatorCombinations(t, tfobjectvalidator.WarnExactlyOneOfChildren, func(configured int) bool {
		return configured != 1
	}, true)

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
	c             any
	objectUnknown bool
	expressions   []path.Expression
	wantErrors    int
	wantWarnings  int
}

// testObjectValidatorCombinations tests all 27 null/set/unknown combinations of a, b,
// and c. For fully known combinations, invalidCount determines whether the number
// of configured children should produce an error (or a warning if warnings is true).
// Any unknown child defers validation, so no diagnostics are expected.
func testObjectValidatorCombinations(t *testing.T, validatorFactory func(...path.Expression) validator.Object, invalidCount func(int) bool, warnings bool) {
	t.Helper()

	values := map[string]any{
		"null":    nil,
		"set":     "value",
		"unknown": tftypes.UnknownValue,
	}
	testCases := make(map[string]objectValidatorTestCase)

	for aName, a := range values {
		for bName, b := range values {
			for cName, c := range values {
				var configured int
				var unknown bool
				for _, value := range []any{a, b, c} {
					switch value {
					case nil:
					case tftypes.UnknownValue:
						unknown = true
					default:
						configured++
					}
				}

				tc := objectValidatorTestCase{a: a, b: b, c: c}
				if !unknown && invalidCount(configured) {
					if warnings {
						tc.wantWarnings = 1
					} else {
						tc.wantErrors = 1
					}
				}
				testCases["a-"+aName+"/b-"+bName+"/c-"+cName] = tc
			}
		}
	}

	testObjectValidator(t, validatorFactory, testCases)
}

func testObjectValidator(t *testing.T, validatorFactory func(...path.Expression) validator.Object, testCases map[string]objectValidatorTestCase) {
	t.Helper()

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config, configValue := testObjectConfig(t, tc.a, tc.b, tc.c, tc.objectUnknown)
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
					path.MatchRelative().AtName("c"),
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

func testObjectConfig(t *testing.T, a, b, c any, objectUnknown bool) (tfsdk.Config, types.Object) {
	t.Helper()

	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
			"b": tftypes.String,
			"c": tftypes.String,
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
			"c": testStringValue(c),
		})
		objectRawValue = tftypes.NewValue(objectType, map[string]tftypes.Value{
			"a": tftypes.NewValue(tftypes.String, a),
			"b": tftypes.NewValue(tftypes.String, b),
			"c": tftypes.NewValue(tftypes.String, c),
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
						"c": schema.StringAttribute{Optional: true},
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

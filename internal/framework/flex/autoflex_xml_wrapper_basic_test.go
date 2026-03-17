// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// Rule 1 + Simple Type: Mammals/Whales example from context
func TestExpandXMLWrapperRule1SimpleType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// AWS types
	type Whales struct {
		Quantity *int32
		Items    []string
	}
	type Mammals struct {
		Whales *Whales
	}

	// TF model (collapsed)
	type mammalModel struct {
		Whales fwtypes.ListValueOf[types.String] `tfsdk:"whales" autoflex:",xmlwrapper=Items,omitempty"`
	}

	testCases := autoFlexTestCases{
		"null": {
			Source:     &mammalModel{Whales: fwtypes.NewListValueOfNull[types.String](ctx)},
			Target:     &Mammals{},
			WantTarget: &Mammals{Whales: nil},
		},
		"empty": {
			Source:     &mammalModel{Whales: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{})},
			Target:     &Mammals{},
			WantTarget: &Mammals{Whales: &Whales{Items: []string{}, Quantity: aws.Int32(0)}},
		},
		"single": {
			Source: &mammalModel{Whales: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
				types.StringValue("blue"),
			})},
			Target:     &Mammals{},
			WantTarget: &Mammals{Whales: &Whales{Items: []string{"blue"}, Quantity: aws.Int32(1)}},
		},
		"multiple": {
			Source: &mammalModel{Whales: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
				types.StringValue("blue"),
				types.StringValue("humpback"),
			})},
			Target:     &Mammals{},
			WantTarget: &Mammals{Whales: &Whales{Items: []string{"blue", "humpback"}, Quantity: aws.Int32(2)}},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

// Rule 1 + Complex Type: Fruits/Apples/Apple example from context
func TestExpandXMLWrapperRule1ComplexType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// AWS types
	type Apple struct {
		Name  *string
		Color *string
	}
	type Apples struct {
		Quantity *int32
		Items    []Apple
	}
	type Fruits struct {
		Apples *Apples
	}

	// TF models (collapsed)
	type appleModel struct {
		Name  types.String `tfsdk:"name"`
		Color types.String `tfsdk:"color"`
	}
	type fruitModel struct {
		Apple fwtypes.ListNestedObjectValueOf[appleModel] `tfsdk:"apple" autoflex:",xmlwrapper=Items,omitempty"`
	}

	testCases := autoFlexTestCases{
		"null": {
			Source:     &fruitModel{Apple: fwtypes.NewListNestedObjectValueOfNull[appleModel](ctx)},
			Target:     &Fruits{},
			WantTarget: &Fruits{Apples: nil},
		},
		"empty": {
			Source:     &fruitModel{Apple: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []appleModel{})},
			Target:     &Fruits{},
			WantTarget: &Fruits{Apples: &Apples{Items: []Apple{}, Quantity: aws.Int32(0)}},
		},
		"single": {
			Source: &fruitModel{Apple: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []appleModel{
				{Name: types.StringValue("Fuji"), Color: types.StringValue("red")},
			})},
			Target: &Fruits{},
			WantTarget: &Fruits{Apples: &Apples{
				Items:    []Apple{{Name: aws.String("Fuji"), Color: aws.String("red")}},
				Quantity: aws.Int32(1),
			}},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

// Rule 2 + Simple Type: Birds/Parrots example from context
func TestExpandXMLWrapperRule2SimpleType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// AWS types
	type Parrots struct {
		Flying   *bool
		Quantity *int32
		Items    []string
	}
	type Birds struct {
		Parrots *Parrots
	}

	// TF models (no collapsing)
	type parrotModel struct {
		Flying types.Bool                        `tfsdk:"flying"`
		Items  fwtypes.ListValueOf[types.String] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
	}
	type birdModel struct {
		Parrot fwtypes.ListNestedObjectValueOf[parrotModel] `tfsdk:"parrot" autoflex:",omitempty"`
	}

	testCases := autoFlexTestCases{
		"null": {
			Source:     &birdModel{Parrot: fwtypes.NewListNestedObjectValueOfNull[parrotModel](ctx)},
			Target:     &Birds{},
			WantTarget: &Birds{Parrots: nil},
		},
		"empty_items": {
			Source: &birdModel{Parrot: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []parrotModel{
				{
					Flying: types.BoolValue(false),
					Items:  fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
				},
			})},
			Target: &Birds{},
			WantTarget: &Birds{Parrots: &Parrots{
				Flying:   aws.Bool(false),
				Items:    []string{},
				Quantity: aws.Int32(0),
			}},
		},
		"with_flying": {
			Source: &birdModel{Parrot: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []parrotModel{
				{
					Flying: types.BoolValue(true),
					Items: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
						types.StringValue("macaw"),
					}),
				},
			})},
			Target: &Birds{},
			WantTarget: &Birds{Parrots: &Parrots{
				Flying:   aws.Bool(true),
				Items:    []string{"macaw"},
				Quantity: aws.Int32(1),
			}},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandXMLWrapperRule2SimpleTypeNoOmitEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type Parrots struct {
		Flying   *bool
		Quantity *int32
		Items    []string
	}
	type Birds struct {
		Parrots *Parrots
	}

	type parrotModel struct {
		Flying types.Bool                        `tfsdk:"flying"`
		Items  fwtypes.ListValueOf[types.String] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
	}
	type birdModel struct {
		Parrot fwtypes.ListNestedObjectValueOf[parrotModel] `tfsdk:"parrot"`
	}

	testCases := autoFlexTestCases{
		"null_no_omitempty": {
			Source: &birdModel{Parrot: fwtypes.NewListNestedObjectValueOfNull[parrotModel](ctx)},
			Target: &Birds{},
			WantTarget: &Birds{Parrots: &Parrots{
				Flying:   aws.Bool(false),
				Items:    []string{},
				Quantity: aws.Int32(0),
			}},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

// Rule 2 + Complex Type: Trees/Oaks/Oak example from context
func TestExpandXMLWrapperRule2ComplexType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// AWS types
	type Oak struct {
		Species *string
		Age     *int32
	}
	type Oaks struct {
		Diseased *bool
		Quantity *int32
		Items    []Oak
	}
	type Trees struct {
		Oaks *Oaks
	}

	// TF models (no collapsing)
	type oakItemModel struct {
		Species types.String `tfsdk:"species"`
		Age     types.Int64  `tfsdk:"age"`
	}
	type oakModel struct {
		Diseased types.Bool                                    `tfsdk:"diseased"`
		Item     fwtypes.ListNestedObjectValueOf[oakItemModel] `tfsdk:"item" autoflex:",xmlwrapper=Items"`
	}
	type treeModel struct {
		Oak fwtypes.ListNestedObjectValueOf[oakModel] `tfsdk:"oak" autoflex:",omitempty"`
	}

	testCases := autoFlexTestCases{
		"null": {
			Source:     &treeModel{Oak: fwtypes.NewListNestedObjectValueOfNull[oakModel](ctx)},
			Target:     &Trees{},
			WantTarget: &Trees{Oaks: nil},
		},
		"empty_items": {
			Source: &treeModel{Oak: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []oakModel{
				{
					Diseased: types.BoolValue(false),
					Item:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []oakItemModel{}),
				},
			})},
			Target: &Trees{},
			WantTarget: &Trees{Oaks: &Oaks{
				Diseased: aws.Bool(false),
				Items:    []Oak{},
				Quantity: aws.Int32(0),
			}},
		},
		"with_diseased": {
			Source: &treeModel{Oak: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []oakModel{
				{
					Diseased: types.BoolValue(false),
					Item: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []oakItemModel{
						{Species: types.StringValue("white oak"), Age: types.Int64Value(100)},
					}),
				},
			})},
			Target: &Trees{},
			WantTarget: &Trees{Oaks: &Oaks{
				Diseased: aws.Bool(false),
				Items:    []Oak{{Species: aws.String("white oak"), Age: aws.Int32(100)}},
				Quantity: aws.Int32(1),
			}},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

// Flatten tests for Rule 1 + Simple Type
func TestFlattenXMLWrapperRule1SimpleType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// AWS types
	type Whales struct {
		Quantity *int32
		Items    []string
	}
	type Mammals struct {
		Whales *Whales
	}

	// TF model (collapsed)
	type mammalModel struct {
		Whales fwtypes.ListValueOf[types.String] `tfsdk:"whales" autoflex:",xmlwrapper=Items,omitempty"`
	}

	testCases := autoFlexTestCases{
		"null": {
			Source:     &Mammals{Whales: nil},
			Target:     &mammalModel{},
			WantTarget: &mammalModel{Whales: fwtypes.NewListValueOfNull[types.String](ctx)},
		},
		"empty": {
			Source:     &Mammals{Whales: &Whales{Items: []string{}, Quantity: aws.Int32(0)}},
			Target:     &mammalModel{},
			WantTarget: &mammalModel{Whales: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{})},
		},
		"single": {
			Source:     &Mammals{Whales: &Whales{Items: []string{"blue"}, Quantity: aws.Int32(1)}},
			Target:     &mammalModel{},
			WantTarget: &mammalModel{Whales: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("blue")})},
		},
		"multiple": {
			Source: &Mammals{Whales: &Whales{Items: []string{"blue", "humpback"}, Quantity: aws.Int32(2)}},
			Target: &mammalModel{},
			WantTarget: &mammalModel{Whales: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
				types.StringValue("blue"),
				types.StringValue("humpback"),
			})},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

// Flatten tests for Rule 1 + Complex Type
func TestFlattenXMLWrapperRule1ComplexType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// AWS types
	type Apple struct {
		Name  *string
		Color *string
	}
	type Apples struct {
		Quantity *int32
		Items    []Apple
	}
	type Fruits struct {
		Apples *Apples
	}

	// TF models (collapsed)
	type appleModel struct {
		Name  types.String `tfsdk:"name"`
		Color types.String `tfsdk:"color"`
	}
	type fruitModel struct {
		Apple fwtypes.ListNestedObjectValueOf[appleModel] `tfsdk:"apple" autoflex:",xmlwrapper=Items,omitempty"`
	}

	testCases := autoFlexTestCases{
		"null": {
			Source:     &Fruits{Apples: nil},
			Target:     &fruitModel{},
			WantTarget: &fruitModel{Apple: fwtypes.NewListNestedObjectValueOfNull[appleModel](ctx)},
		},
		"empty": {
			Source:     &Fruits{Apples: &Apples{Items: []Apple{}, Quantity: aws.Int32(0)}},
			Target:     &fruitModel{},
			WantTarget: &fruitModel{Apple: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []appleModel{})},
		},
		"single": {
			Source: &Fruits{Apples: &Apples{
				Items:    []Apple{{Name: aws.String("Fuji"), Color: aws.String("red")}},
				Quantity: aws.Int32(1),
			}},
			Target: &fruitModel{},
			WantTarget: &fruitModel{Apple: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []appleModel{
				{Name: types.StringValue("Fuji"), Color: types.StringValue("red")},
			})},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true, PrintLogs: true})
}

// Flatten tests for Rule 2 + Simple Type
func TestFlattenXMLWrapperRule2SimpleType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// AWS types
	type Parrots struct {
		Flying   *bool
		Quantity *int32
		Items    []string
	}
	type Birds struct {
		Parrots *Parrots
	}

	// TF models (no collapsing)
	type parrotModel struct {
		Flying types.Bool                        `tfsdk:"flying"`
		Items  fwtypes.ListValueOf[types.String] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
	}
	type birdModel struct {
		Parrot fwtypes.ListNestedObjectValueOf[parrotModel] `tfsdk:"parrot" autoflex:",omitempty"`
	}

	testCases := autoFlexTestCases{
		"null": {
			Source:     &Birds{Parrots: nil},
			Target:     &birdModel{},
			WantTarget: &birdModel{Parrot: fwtypes.NewListNestedObjectValueOfNull[parrotModel](ctx)},
		},
		"empty_all_zero": {
			Source: &Birds{Parrots: &Parrots{
				Flying:   aws.Bool(false), // Zero value
				Items:    []string{},
				Quantity: aws.Int32(0),
			}},
			Target: &birdModel{},
			// Empty Items + Flying at zero = null (no meaningful data)
			WantTarget: &birdModel{Parrot: fwtypes.NewListNestedObjectValueOfNull[parrotModel](ctx)},
		},
		"empty_items_with_value": {
			Source: &Birds{Parrots: &Parrots{
				Flying:   aws.Bool(true), // Non-zero value
				Items:    []string{},
				Quantity: aws.Int32(0),
			}},
			Target: &birdModel{},
			// Empty Items but Flying=true is meaningful, so create block
			WantTarget: &birdModel{Parrot: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []parrotModel{
				{
					Flying: types.BoolValue(true),
					Items:  fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
				},
			})},
		},
		"with_items": {
			Source: &Birds{Parrots: &Parrots{
				Flying:   aws.Bool(true),
				Items:    []string{"macaw"},
				Quantity: aws.Int32(1),
			}},
			Target: &birdModel{},
			WantTarget: &birdModel{Parrot: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []parrotModel{
				{
					Flying: types.BoolValue(true),
					Items: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
						types.StringValue("macaw"),
					}),
				},
			})},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenXMLWrapperRule2SimpleTypeNoOmitEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// AWS types
	type Parrots struct {
		Flying   *bool
		Quantity *int32
		Items    []string
	}
	type Birds struct {
		Parrots *Parrots
	}

	// TF models (no omitempty - should create zero-value struct)
	type parrotModel struct {
		Flying types.Bool                        `tfsdk:"flying"`
		Items  fwtypes.ListValueOf[types.String] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
	}
	type birdModel struct {
		Parrot fwtypes.ListNestedObjectValueOf[parrotModel] `tfsdk:"parrot"`
	}

	testCases := autoFlexTestCases{
		"empty_all_zero_no_omitempty": {
			Source: &Birds{Parrots: &Parrots{
				Flying:   aws.Bool(false),
				Items:    []string{},
				Quantity: aws.Int32(0),
			}},
			Target: &birdModel{},
			// Without omitempty, zero values should create a block
			WantTarget: &birdModel{Parrot: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []parrotModel{
				{
					Flying: types.BoolValue(false),
					Items:  fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
				},
			})},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

// Flatten tests for Rule 2 + Complex Type
func TestFlattenXMLWrapperRule2ComplexType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// AWS types
	type Oak struct {
		Species *string
		Age     *int32
	}
	type Oaks struct {
		Diseased *bool
		Quantity *int32
		Items    []Oak
	}
	type Trees struct {
		Oaks *Oaks
	}

	// TF models (no collapsing)
	type oakItemModel struct {
		Species types.String `tfsdk:"species"`
		Age     types.Int64  `tfsdk:"age"`
	}
	type oakModel struct {
		Diseased types.Bool                                    `tfsdk:"diseased"`
		Item     fwtypes.ListNestedObjectValueOf[oakItemModel] `tfsdk:"item" autoflex:",xmlwrapper=Items"`
	}
	type treeModel struct {
		Oak fwtypes.ListNestedObjectValueOf[oakModel] `tfsdk:"oak" autoflex:",omitempty"`
	}

	testCases := autoFlexTestCases{
		"null": {
			Source:     &Trees{Oaks: nil},
			Target:     &treeModel{},
			WantTarget: &treeModel{Oak: fwtypes.NewListNestedObjectValueOfNull[oakModel](ctx)},
		},
		"empty_all_zero": {
			Source: &Trees{Oaks: &Oaks{
				Diseased: aws.Bool(false), // Zero value
				Items:    []Oak{},
				Quantity: aws.Int32(0),
			}},
			Target: &treeModel{},
			// Empty Items + Diseased at zero = null (no meaningful data)
			WantTarget: &treeModel{Oak: fwtypes.NewListNestedObjectValueOfNull[oakModel](ctx)},
		},
		"empty_items_with_value": {
			Source: &Trees{Oaks: &Oaks{
				Diseased: aws.Bool(true), // Non-zero value
				Items:    []Oak{},
				Quantity: aws.Int32(0),
			}},
			Target: &treeModel{},
			// Empty Items but Diseased=true is meaningful, so create block
			WantTarget: &treeModel{Oak: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []oakModel{
				{
					Diseased: types.BoolValue(true),
					Item:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []oakItemModel{}),
				},
			})},
		},
		"with_items": {
			Source: &Trees{Oaks: &Oaks{
				Diseased: aws.Bool(false),
				Items:    []Oak{{Species: aws.String("white oak"), Age: aws.Int32(100)}},
				Quantity: aws.Int32(1),
			}},
			Target: &treeModel{},
			WantTarget: &treeModel{Oak: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []oakModel{
				{
					Diseased: types.BoolValue(false),
					Item: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []oakItemModel{
						{Species: types.StringValue("white oak"), Age: types.Int64Value(100)},
					}),
				},
			})},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

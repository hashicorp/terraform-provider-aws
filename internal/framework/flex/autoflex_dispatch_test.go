// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Verifies dispatcher selection (Expander/TypedExpander/Interface/Flattener/TypedFlattener),
// enforces interface contracts, and snapshots trace logging—behavior/routing focus, not data
// mapping.
//
// This test file uses golden snapshots for log verification. These can be found in
// testdata/autoflex/dispatch/*.golden

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type tfListNestedObject[T any] struct {
	Field1 fwtypes.ListNestedObjectValueOf[T] `tfsdk:"field1"`
}

type tfSetNestedObject[T any] struct {
	Field1 fwtypes.SetNestedObjectValueOf[T] `tfsdk:"field1"`
}

type tfObjectValue[T any] struct {
	Field1 fwtypes.ObjectValueOf[T] `tfsdk:"field1"`
}

type tfInterfaceFlexer struct {
	Field1 types.String `tfsdk:"field1"`
}

var (
	_ Expander  = tfInterfaceFlexer{}
	_ Flattener = &tfInterfaceFlexer{}
)

func (t tfInterfaceFlexer) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return &awsInterfaceInterfaceImpl{
		AWSField: StringValueFromFramework(ctx, t.Field1),
	}, nil
}

func (t *tfInterfaceFlexer) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch val := v.(type) {
	case awsInterfaceInterfaceImpl:
		t.Field1 = StringValueToFramework(ctx, val.AWSField)
		return diags

	default:
		return diags
	}
}

type tfInterfaceIncompatibleExpander struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ Expander = tfInterfaceIncompatibleExpander{}

func (t tfInterfaceIncompatibleExpander) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return &awsInterfaceIncompatibleImpl{
		AWSField: StringValueFromFramework(ctx, t.Field1),
	}, nil
}

type awsInterfaceIncompatibleImpl struct {
	AWSField string
}

type awsInterfaceSingle struct {
	Field1 awsInterfaceInterface
}

type awsInterfaceSlice struct {
	Field1 []awsInterfaceInterface
}

type awsInterfaceInterface interface {
	isAWSInterfaceInterface()
}

type awsInterfaceInterfaceImpl struct {
	AWSField string
}

var _ awsInterfaceInterface = &awsInterfaceInterfaceImpl{}

func (t *awsInterfaceInterfaceImpl) isAWSInterfaceInterface() {} // nosemgrep:ci.aws-in-func-name

type tfFlexer struct {
	Field1 types.String `tfsdk:"field1"`
}

var (
	_ Expander  = tfFlexer{}
	_ Flattener = &tfFlexer{}
)

func (t tfFlexer) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return &awsExpander{
		AWSField: StringValueFromFramework(ctx, t.Field1),
	}, nil
}

func (t *tfFlexer) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch val := v.(type) {
	case awsExpander:
		t.Field1 = StringValueToFramework(ctx, val.AWSField)
		return diags

	default:
		return diags
	}
}

type tfExpanderListNestedObject tfListNestedObject[tfFlexer]

type tfExpanderSetNestedObject tfSetNestedObject[tfFlexer]

type tfExpanderObjectValue tfObjectValue[tfFlexer]

type tfTypedExpanderListNestedObject tfListNestedObject[tfTypedExpander]

type tfTypedExpanderSetNestedObject tfSetNestedObject[tfTypedExpander]

type tfTypedExpanderObjectValue tfObjectValue[tfTypedExpander]

type tfExpanderToString struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ Expander = tfExpanderToString{}

func (t tfExpanderToString) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return StringValueFromFramework(ctx, t.Field1), nil
}

type tfExpanderToNil struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ Expander = tfExpanderToNil{}

func (t tfExpanderToNil) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return nil, nil
}

type tfTypedExpander struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ TypedExpander = tfTypedExpander{}

func (t tfTypedExpander) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	return &awsExpander{
		AWSField: StringValueFromFramework(ctx, t.Field1),
	}, nil
}

type tfTypedExpanderToNil struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ TypedExpander = tfTypedExpanderToNil{}

func (t tfTypedExpanderToNil) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	return nil, nil
}

type tfInterfaceTypedExpander struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ TypedExpander = tfInterfaceTypedExpander{}

func (t tfInterfaceTypedExpander) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[awsInterfaceInterface]():
		return &awsInterfaceInterfaceImpl{
			AWSField: StringValueFromFramework(ctx, t.Field1),
		}, nil
	}

	return nil, nil
}

type tfInterfaceIncompatibleTypedExpander struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ TypedExpander = tfInterfaceIncompatibleTypedExpander{}

func (t tfInterfaceIncompatibleTypedExpander) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	return &awsInterfaceIncompatibleImpl{
		AWSField: StringValueFromFramework(ctx, t.Field1),
	}, nil
}

type awsExpander struct {
	AWSField string
}

type awsExpanderIncompatible struct {
	Incompatible int
}

type awsExpanderSingleStruct struct {
	Field1 awsExpander
}

type awsExpanderSinglePtr struct {
	Field1 *awsExpander
}

type awsExpanderStructSlice struct {
	Field1 []awsExpander
}

type awsExpanderPtrSlice struct {
	Field1 []*awsExpander
}

func TestExpandLogging_collections(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"Collection of primitive types Source and slice or map of primtive types Target": {
			Source: &tfCollectionsOfPrimitiveElements{
				Field1: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field2: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field3: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field4: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field5: types.MapValueMust(types.StringType, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
				Field6: types.MapValueMust(types.StringType, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
			},
			Target: &awsCollectionsOfPrimitiveElements{},
			WantTarget: &awsCollectionsOfPrimitiveElements{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true})
}

func TestExpandInterfaceContract(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"source field does not implement attr.Value Source": {
			Source:        &awsSingleStringValue{Field1: "a"},
			Target:        &awsSingleStringValue{},
			ExpectedDiags: diagAF[string](diagExpandingSourceDoesNotImplementAttrValue),
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestExpandExpander(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"top level struct Target": {
			Source: tfFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &awsExpander{},
			WantTarget: &awsExpander{
				AWSField: "value1",
			},
		},
		"top level string Target": {
			Source: tfExpanderToString{
				Field1: types.StringValue("value1"),
			},
			Target:     aws.String(""),
			WantTarget: aws.String("value1"),
		},
		"top level incompatible struct Target": {
			Source: tfFlexer{
				Field1: types.StringValue("value1"),
			},
			Target:        &awsExpanderIncompatible{},
			ExpectedDiags: diagAF2[awsExpander, awsExpanderIncompatible](diagCannotBeAssigned),
		},
		"top level expands to nil": {
			Source: tfExpanderToNil{
				Field1: types.StringValue("value1"),
			},
			Target:        &awsExpander{},
			ExpectedDiags: diagAF[tfExpanderToNil](diagExpandsToNil),
		},
		"top level incompatible non-struct Target": {
			Source: tfExpanderToString{
				Field1: types.StringValue("value1"),
			},
			Target:        aws.Int64(0),
			ExpectedDiags: diagAF2[string, int64](diagCannotBeAssigned),
		},
		"single list Source and single struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
		},
		"single set Source and single struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
		},
		"single list Source and single *struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
		},
		"single set Source and single *struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
		},
		"empty list Source and empty struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
		},
		"non-empty list Source and non-empty struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
		},
		"empty list Source and empty *struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
		},
		"non-empty list Source and non-empty *struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
		},
		"empty set Source and empty struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
		},
		"non-empty set Source and non-empty struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
		},
		"empty set Source and empty *struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
		},
		"non-empty set Source and non-empty *struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
		},
		"object value Source and struct Target": {
			Source: tfExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
		},
		"object value Source and *struct Target": {
			Source: tfExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func testFlexAWSInterfaceInterfacePtr(v awsInterfaceInterface) *awsInterfaceInterface { // nosemgrep:ci.aws-in-func-name
	return &v
}

func TestExpandInterface(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var targetInterface awsInterfaceInterface

	testCases := autoFlexTestCases{
		"top level": {
			Source: tfInterfaceFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			WantTarget: testFlexAWSInterfaceInterfacePtr(&awsInterfaceInterfaceImpl{
				AWSField: "value1",
			}),
		},
		"top level return value does not implement target interface": {
			Source: tfInterfaceIncompatibleExpander{
				Field1: types.StringValue("value1"),
			},
			Target:        &targetInterface,
			ExpectedDiags: diagAF2[*awsInterfaceIncompatibleImpl, awsInterfaceInterface](diagExpandedTypeDoesNotImplement),
		},
		"single list Source and single interface Target": {
			Source: tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
		},
		"single list non-Expander Source and single interface Target": {
			Source: tfListNestedObject[tfSingleStringField]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: nil,
			},
		},
		"single set Source and single interface Target": {
			Source: tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
		},
		"empty list Source and empty interface Target": {
			Source: tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
		},
		"non-empty list Source and non-empty interface Target": {
			Source: tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
		},
		"empty set Source and empty interface Target": {
			Source: tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
		},
		"non-empty set Source and non-empty interface Target": {
			Source: tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
		},
		"object value Source and struct Target": {
			Source: tfObjectValue[tfInterfaceFlexer]{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfInterfaceFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestExpandInterfaceTypedExpander(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var targetInterface awsInterfaceInterface

	testCases := autoFlexTestCases{
		"top level": {
			Source: tfInterfaceTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			WantTarget: testFlexAWSInterfaceInterfacePtr(&awsInterfaceInterfaceImpl{
				AWSField: "value1",
			}),
		},
		"top level return value does not implement target interface": {
			Source: tfInterfaceIncompatibleTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target:        &targetInterface,
			ExpectedDiags: diagAF2[*awsInterfaceIncompatibleImpl, awsInterfaceInterface](diagExpandedTypeDoesNotImplement),
		},
		"single list Source and single interface Target": {
			Source: tfListNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
		},
		"single list non-Expander Source and single interface Target": {
			Source: tfListNestedObject[tfSingleStringField]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: nil,
			},
		},
		"single set Source and single interface Target": {
			Source: tfSetNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
		},
		"empty list Source and empty interface Target": {
			Source: tfListNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
		},
		"non-empty list Source and non-empty interface Target": {
			Source: tfListNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
		},
		"empty set Source and empty interface Target": {
			Source: tfSetNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
		},
		"non-empty set Source and non-empty interface Target": {
			Source: tfSetNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
		},
		"object value Source and struct Target": {
			Source: tfObjectValue[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfInterfaceTypedExpander{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestExpandTypedExpander(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"top level struct Target": {
			Source: tfTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &awsExpander{},
			WantTarget: &awsExpander{
				AWSField: "value1",
			},
		},
		"top level incompatible struct Target": {
			Source: tfTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target:        &awsExpanderIncompatible{},
			ExpectedDiags: diagAF2[awsExpander, awsExpanderIncompatible](diagCannotBeAssigned),
		},
		"top level expands to nil": {
			Source: tfTypedExpanderToNil{
				Field1: types.StringValue("value1"),
			},
			Target:        &awsExpander{},
			ExpectedDiags: diagAF[tfTypedExpanderToNil](diagExpandsToNil),
		},
		"single list Source and single struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
		},
		"single set Source and single struct Target": {
			Source: tfSetNestedObject[tfTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
		},
		"single list Source and single *struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
		},
		"single set Source and single *struct Target": {
			Source: tfTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
		},
		"empty list Source and empty struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
		},
		"non-empty list Source and non-empty struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
		},
		"empty list Source and empty *struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
		},
		"non-empty list Source and non-empty *struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
		},
		"empty set Source and empty struct Target": {
			Source: tfTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
		},
		"non-empty set Source and non-empty struct Target": {
			Source: tfTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
		},
		"empty set Source and empty *struct Target": {
			Source: tfTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
		},
		"non-empty set Source and non-empty *struct Target": {
			Source: tfTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
		},
		"object value Source and struct Target": {
			Source: tfTypedExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfTypedExpander{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
		},
		"object value Source and *struct Target": {
			Source: tfTypedExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfTypedExpander{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestFlattenLogging_collections(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"zero value slice or map of primitive types Source and Collection of primtive types Target": {
			Source: &awsCollectionsOfPrimitiveElements{},
			Target: &tfCollectionsOfPrimitiveElements{},
			WantTarget: &tfCollectionsOfPrimitiveElements{
				Field1: types.ListNull(types.StringType),
				Field2: types.ListNull(types.StringType),
				Field3: types.SetNull(types.StringType),
				Field4: types.SetNull(types.StringType),
				Field5: types.MapNull(types.StringType),
				Field6: types.MapNull(types.StringType),
			},
		},
		"slice or map of primitive types Source and Collection of primitive types Target": {
			Source: &awsCollectionsOfPrimitiveElements{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
			Target: &tfCollectionsOfPrimitiveElements{},
			WantTarget: &tfCollectionsOfPrimitiveElements{
				Field1: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field2: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field3: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field4: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field5: types.MapValueMust(types.StringType, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
				Field6: types.MapValueMust(types.StringType, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true})
}

func TestFlattenInterfaceContract(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"target field does not implement attr.Value Target": {
			Source:        &awsSingleStringValue{Field1: "a"},
			Target:        &awsSingleStringValue{},
			ExpectedDiags: diagAF[string](diagFlatteningTargetDoesNotImplementAttrValue),
		},
		"source struct field to non-attr.Value": {
			Source:        &awsRFC3339TimeValue{},
			Target:        &awsRFC3339TimeValue{},
			ExpectedDiags: diagAF[time.Time](diagFlatteningTargetDoesNotImplementAttrValue),
		},
		"source struct ptr field to non-attr.Value": {
			Source:        &awsRFC3339TimePointer{},
			Target:        &awsRFC3339TimeValue{},
			ExpectedDiags: diagAF[time.Time](diagFlatteningTargetDoesNotImplementAttrValue),
		},
		"source struct field to non-attr.Value ptr": {
			Source:        &awsRFC3339TimeValue{},
			Target:        &awsRFC3339TimePointer{},
			ExpectedDiags: diagAF[*time.Time](diagFlatteningTargetDoesNotImplementAttrValue),
		},
		"source struct ptr field to non-attr.Value ptr": {
			Source:        &awsRFC3339TimePointer{},
			Target:        &awsRFC3339TimePointer{},
			ExpectedDiags: diagAF[*time.Time](diagFlatteningTargetDoesNotImplementAttrValue),
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestFlattenInterface(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"nil interface Source and list Target": {
			Source: awsInterfaceSingle{
				Field1: nil,
			},
			Target: &tfListNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[tfInterfaceFlexer](ctx),
			},
		},
		"single interface Source and single list Target": {
			Source: awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &tfListNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
		},
		"nil interface Source and non-Flattener list Target": {
			Source: awsInterfaceSingle{
				Field1: nil,
			},
			Target: &tfListNestedObject[tfSingleStringField]{},
			WantTarget: &tfListNestedObject[tfSingleStringField]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			},
		},
		"single interface Source and non-Flattener list Target": {
			Source: awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &tfListNestedObject[tfSingleStringField]{},
			WantTarget: &tfListNestedObject[tfSingleStringField]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			},
		},

		"nil interface Source and set Target": {
			Source: awsInterfaceSingle{
				Field1: nil,
			},
			Target: &tfSetNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[tfInterfaceFlexer](ctx),
			},
		},
		"single interface Source and single set Target": {
			Source: awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &tfSetNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
		},

		"nil interface list Source and empty list Target": {
			Source: awsInterfaceSlice{
				Field1: nil,
			},
			Target: &tfListNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[tfInterfaceFlexer](ctx),
			},
		},
		"empty interface list Source and empty list Target": {
			Source: awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
			Target: &tfListNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{}),
			},
		},
		"non-empty interface list Source and non-empty list Target": {
			Source: awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			Target: &tfListNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
		},

		"nil interface list Source and empty set Target": {
			Source: awsInterfaceSlice{
				Field1: nil,
			},
			Target: &tfSetNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[tfInterfaceFlexer](ctx),
			},
		},
		"empty interface list Source and empty set Target": {
			Source: awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
			Target: &tfSetNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{}),
			},
		},
		"non-empty interface list Source and non-empty set Target": {
			Source: awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			Target: &tfSetNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
		},
		"nil interface Source and nested object Target": {
			Source: awsInterfaceSingle{
				Field1: nil,
			},
			Target: &tfObjectValue[tfInterfaceFlexer]{},
			WantTarget: &tfObjectValue[tfInterfaceFlexer]{
				Field1: fwtypes.NewObjectValueOfNull[tfInterfaceFlexer](ctx),
			},
		},
		"interface Source and nested object Target": {
			Source: awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &tfObjectValue[tfInterfaceFlexer]{},
			WantTarget: &tfObjectValue[tfInterfaceFlexer]{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfInterfaceFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestFlattenFlattener(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"top level struct Source": {
			Source: awsExpander{
				AWSField: "value1",
			},
			Target: &tfFlexer{},
			WantTarget: &tfFlexer{
				Field1: types.StringValue("value1"),
			},
		},
		"top level incompatible struct Target": {
			Source: awsExpanderIncompatible{
				Incompatible: 123,
			},
			Target: &tfFlexer{},
			WantTarget: &tfFlexer{
				Field1: types.StringNull(),
			},
		},
		"single struct Source and single list Target": {
			Source: awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
		},
		"nil *struct Source and null list Target": {
			Source: awsExpanderSinglePtr{
				Field1: nil,
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfNull[tfFlexer](ctx),
			},
		},
		"single struct Source and single set Target": {
			Source: awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
		},
		"single *struct Source and single list Target": {
			Source: awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
		},
		"single *struct Source and single set Target": {
			Source: awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
		},
		"nil *struct Source and null set Target": {
			Source: awsExpanderSinglePtr{
				Field1: nil,
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[tfFlexer](ctx),
			},
		},

		"empty struct list Source and empty list Target": {
			Source: &awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
		},
		"non-empty struct list Source and non-empty list Target": {
			Source: &awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
		},
		"empty *struct list Source and empty list Target": {
			Source: &awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
		},
		"non-empty *struct list Source and non-empty list Target": {
			Source: &awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
		},
		"empty struct list Source and empty set Target": {
			Source: awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
		},
		"non-empty struct list Source and set Target": {
			Source: awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
		},
		"empty *struct list Source and empty set Target": {
			Source: awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
		},
		"non-empty *struct list Source and non-empty set Target": {
			Source: awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
		},
		"struct Source and object value Target": {
			Source: awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderObjectValue{},
			WantTarget: &tfExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
		},
		"*struct Source and object value Target": {
			Source: awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderObjectValue{},
			WantTarget: &tfExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

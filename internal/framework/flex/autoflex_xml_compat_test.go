// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten of AWS API XML wrappers (Items/Quantity).

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// AWS SDK types that mirror CloudFront function association patterns
type FunctionAssociation struct {
	EventType   string  `json:"EventType"`
	FunctionARN *string `json:"FunctionARN"`
}

type FunctionAssociations struct {
	Quantity *int32
	Items    []FunctionAssociation
}

// Terraform model types
type FunctionAssociationTF struct {
	EventType   types.String `tfsdk:"event_type"`
	FunctionARN types.String `tfsdk:"function_arn"`
}

type DistributionConfigTF struct {
	FunctionAssociations fwtypes.SetNestedObjectValueOf[FunctionAssociationTF] `tfsdk:"function_associations"`
}

type DistributionConfigAWS struct {
	FunctionAssociations *FunctionAssociations
}

func TestExpandXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid function associations": {
			Source: DistributionConfigTF{
				FunctionAssociations: fwtypes.NewSetNestedObjectValueOfSliceMust(
					ctx,
					[]*FunctionAssociationTF{
						{
							EventType:   types.StringValue("viewer-request"),
							FunctionARN: types.StringValue("arn:aws:cloudfront::123456789012:function/test-function-1"),
						},
						{
							EventType:   types.StringValue("viewer-response"),
							FunctionARN: types.StringValue("arn:aws:cloudfront::123456789012:function/test-function-2"),
						},
					},
				),
			},
			Target: &DistributionConfigAWS{},
			WantTarget: &DistributionConfigAWS{
				FunctionAssociations: &FunctionAssociations{
					Quantity: aws.Int32(2),
					Items: []FunctionAssociation{
						{
							EventType:   "viewer-request",
							FunctionARN: aws.String("arn:aws:cloudfront::123456789012:function/test-function-1"),
						},
						{
							EventType:   "viewer-response",
							FunctionARN: aws.String("arn:aws:cloudfront::123456789012:function/test-function-2"),
						},
					},
				},
			},
		},
		"empty function associations": {
			Source: DistributionConfigTF{
				FunctionAssociations: fwtypes.NewSetNestedObjectValueOfSliceMust(
					ctx,
					[]*FunctionAssociationTF{},
				),
			},
			Target: &DistributionConfigAWS{},
			WantTarget: &DistributionConfigAWS{
				FunctionAssociations: &FunctionAssociations{
					Quantity: aws.Int32(0),
					Items:    []FunctionAssociation{},
				},
			},
		},
		"single function association": {
			Source: DistributionConfigTF{
				FunctionAssociations: fwtypes.NewSetNestedObjectValueOfSliceMust(
					ctx,
					[]*FunctionAssociationTF{
						{
							EventType:   types.StringValue("origin-request"),
							FunctionARN: types.StringValue("arn:aws:cloudfront::123456789012:function/origin-function"),
						},
					},
				),
			},
			Target: &DistributionConfigAWS{},
			WantTarget: &DistributionConfigAWS{
				FunctionAssociations: &FunctionAssociations{
					Quantity: aws.Int32(1),
					Items: []FunctionAssociation{
						{
							EventType:   "origin-request",
							FunctionARN: aws.String("arn:aws:cloudfront::123456789012:function/origin-function"),
						},
					},
				},
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}

// Test XML wrapper expansion for direct struct (not pointer to struct)
type DirectXMLWrapper struct {
	Items    []string
	Quantity *int32
}

type DirectWrapperTF struct {
	Items fwtypes.SetValueOf[types.String] `tfsdk:"items"`
}

type DirectWrapperAWS struct {
	Items DirectXMLWrapper
}

func TestExpandXMLWrapperDirect(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"direct xml wrapper": {
			Source: DirectWrapperTF{
				Items: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("item1"),
					types.StringValue("item2"),
				}),
			},
			Target: &DirectWrapperAWS{},
			WantTarget: &DirectWrapperAWS{
				Items: DirectXMLWrapper{
					Items:    []string{"item1", "item2"},
					Quantity: aws.Int32(2),
				},
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}

func TestIsXMLWrapperStruct(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "valid XML wrapper",
			input:    FunctionAssociations{},
			expected: true,
		},
		{
			name:     "valid XML wrapper with slice of strings",
			input:    DirectXMLWrapper{},
			expected: true,
		},
		{
			name:     "not a struct",
			input:    "string",
			expected: false,
		},
		{
			name:     "struct without Items field",
			input:    struct{ Quantity *int32 }{},
			expected: false,
		},
		{
			name:     "struct without Quantity field",
			input:    struct{ Items []string }{},
			expected: false,
		},
		{
			name: "struct with wrong Quantity type",
			input: struct {
				Items    []string
				Quantity int32
			}{},
			expected: false,
		},
		{
			name: "struct with Items not a slice",
			input: struct {
				Items    string
				Quantity *int32
			}{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isXMLWrapperStruct(reflect.TypeOf(tc.input))
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// Mock AWS types with XML wrapper pattern (for flattening - AWS to TF)
type awsStatusCodesForFlatten struct {
	Items    []int32
	Quantity *int32
}

type awsHeadersForFlatten struct {
	Items    []string
	Quantity *int32
}

// CloudFront FunctionAssociation test types
type awsFunctionAssociationsForFlatten struct {
	Items    []FunctionAssociation `json:"Items"`
	Quantity *int32                `json:"Quantity"`
}

type tfFunctionAssociationsModelForFlatten struct {
	FunctionAssociations fwtypes.SetNestedObjectValueOf[FunctionAssociationTF] `tfsdk:"function_associations" autoflex:",wrapper=items"`
}

// TF model types with wrapper tags (for flattening - AWS to TF)
type tfStatusCodesModelForFlatten struct {
	StatusCodes fwtypes.SetValueOf[types.Int64] `tfsdk:"status_codes" autoflex:",wrapper=items"`
}

type tfHeadersModelForFlatten struct {
	Headers fwtypes.ListValueOf[types.String] `tfsdk:"headers" autoflex:",wrapper=items"`
}

func TestFlattenXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"int32 slice to set": {
			Source: awsStatusCodesForFlatten{
				Items:    []int32{400, 404},
				Quantity: aws.Int32(2),
			},
			Target: &tfStatusCodesModelForFlatten{},
			WantTarget: &tfStatusCodesModelForFlatten{
				StatusCodes: fwtypes.NewSetValueOfMust[types.Int64](ctx, []attr.Value{
					types.Int64Value(400),
					types.Int64Value(404),
				}),
			},
		},
		"string slice to list": {
			Source: awsHeadersForFlatten{
				Items:    []string{"accept", "content-type"},
				Quantity: aws.Int32(2),
			},
			Target: &tfHeadersModelForFlatten{},
			WantTarget: &tfHeadersModelForFlatten{
				Headers: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("accept"),
					types.StringValue("content-type"),
				}),
			},
		},
		"complex type - function associations": {
			Source: awsFunctionAssociationsForFlatten{
				Items: []FunctionAssociation{
					{
						EventType:   "viewer-request",
						FunctionARN: aws.String("arn:aws:cloudfront::123456789012:function/example-function"),
					},
					{
						EventType:   "viewer-response",
						FunctionARN: aws.String("arn:aws:cloudfront::123456789012:function/another-function"),
					},
				},
				Quantity: aws.Int32(2),
			},
			Target: &tfFunctionAssociationsModelForFlatten{},
			WantTarget: &tfFunctionAssociationsModelForFlatten{
				FunctionAssociations: func() fwtypes.SetNestedObjectValueOf[FunctionAssociationTF] {
					elems := []*FunctionAssociationTF{
						{
							EventType:   types.StringValue("viewer-request"),
							FunctionARN: types.StringValue("arn:aws:cloudfront::123456789012:function/example-function"),
						},
						{
							EventType:   types.StringValue("viewer-response"),
							FunctionARN: types.StringValue("arn:aws:cloudfront::123456789012:function/another-function"),
						},
					}
					setValue, _ := fwtypes.NewSetNestedObjectValueOfSlice(ctx, elems, nil)
					return setValue
				}(),
			},
		},
		"empty slice to null set": {
			Source: awsStatusCodesForFlatten{
				Items:    nil,
				Quantity: aws.Int32(0),
			},
			Target: &tfStatusCodesModelForFlatten{},
			WantTarget: &tfStatusCodesModelForFlatten{
				StatusCodes: fwtypes.NewSetValueOfNull[types.Int64](ctx),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Group 1 Tests: Plain strings with XML wrappers (pointer and non-pointer)

// XMLWrappedStringSlice mimicks AWS types that are XML wrappers around string slices
type XMLWrappedStringSlice struct {
	Items    []string
	Quantity *int32
}

// StructWithWrappedStringSlicePtr mimicks AWS types that have a pointer to an XML wrapper around string slices
type StructWithWrappedStringSlicePtr struct {
	HttpsPort             int32
	XMLWrappedStringSlice *XMLWrappedStringSlice
}

// StructWithWrappedStringSlice mimicks AWS types that have a non-pointer to an XML wrapper around string slices
type StructWithWrappedStringSlice struct {
	HttpsPort             int32
	XMLWrappedStringSlice XMLWrappedStringSlice
}

// Terraform model types with plain strings
type modelWithStringSet struct {
	HttpsPort             types.Int32                      `tfsdk:"https_port"`
	XMLWrappedStringSlice fwtypes.SetValueOf[types.String] `tfsdk:"xml_wrapped_string_slice"`
}

type modelWithStringList struct {
	HttpsPort             types.Int32                       `tfsdk:"https_port"`
	XMLWrappedStringSlice fwtypes.ListValueOf[types.String] `tfsdk:"xml_wrapped_string_slice"`
}

func TestExpandStringsToXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"string set to xml wrapper with pointer": {
			Source: modelWithStringSet{
				HttpsPort: types.Int32Value(443),
				XMLWrappedStringSlice: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("item1"),
					types.StringValue("item2"),
				}),
			},
			Target: &StructWithWrappedStringSlicePtr{},
			WantTarget: &StructWithWrappedStringSlicePtr{
				HttpsPort: 443,
				XMLWrappedStringSlice: &XMLWrappedStringSlice{
					Items:    []string{"item1", "item2"},
					Quantity: aws.Int32(2),
				},
			},
		},
		"string list to xml wrapper non-pointer": {
			Source: modelWithStringList{
				HttpsPort: types.Int32Value(443),
				XMLWrappedStringSlice: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("item1"),
					types.StringValue("item2"),
				}),
			},
			Target: &StructWithWrappedStringSlice{},
			WantTarget: &StructWithWrappedStringSlice{
				HttpsPort: 443,
				XMLWrappedStringSlice: XMLWrappedStringSlice{
					Items:    []string{"item1", "item2"},
					Quantity: aws.Int32(2),
				},
			},
		},
		"null string set to nil xml wrapper pointer": {
			Source: modelWithStringSet{
				HttpsPort:             types.Int32Value(443),
				XMLWrappedStringSlice: fwtypes.NewSetValueOfNull[types.String](ctx),
			},
			Target: &StructWithWrappedStringSlicePtr{},
			WantTarget: &StructWithWrappedStringSlicePtr{
				HttpsPort:             443,
				XMLWrappedStringSlice: nil,
			},
		},
		"empty string list to xml wrapper non-pointer": {
			Source: modelWithStringList{
				HttpsPort:             types.Int32Value(443),
				XMLWrappedStringSlice: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
			},
			Target: &StructWithWrappedStringSlice{},
			WantTarget: &StructWithWrappedStringSlice{
				HttpsPort: 443,
				XMLWrappedStringSlice: XMLWrappedStringSlice{
					Items:    []string{},
					Quantity: aws.Int32(0),
				},
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

func TestFlattenStringsFromXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"xml wrapper pointer to string set": {
			Source: StructWithWrappedStringSlicePtr{
				HttpsPort: 443,
				XMLWrappedStringSlice: &XMLWrappedStringSlice{
					Items:    []string{"item1", "item2"},
					Quantity: aws.Int32(2),
				},
			},
			Target: &modelWithStringSet{},
			WantTarget: &modelWithStringSet{
				HttpsPort: types.Int32Value(443),
				XMLWrappedStringSlice: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("item1"),
					types.StringValue("item2"),
				}),
			},
		},
		"xml wrapper non-pointer to string list": {
			Source: StructWithWrappedStringSlice{
				HttpsPort: 443,
				XMLWrappedStringSlice: XMLWrappedStringSlice{
					Items:    []string{"item1", "item2"},
					Quantity: aws.Int32(2),
				},
			},
			Target: &modelWithStringList{},
			WantTarget: &modelWithStringList{
				HttpsPort: types.Int32Value(443),
				XMLWrappedStringSlice: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("item1"),
					types.StringValue("item2"),
				}),
			},
		},
		"nil xml wrapper pointer to null string set": {
			Source: StructWithWrappedStringSlicePtr{
				HttpsPort:             443,
				XMLWrappedStringSlice: nil,
			},
			Target: &modelWithStringSet{},
			WantTarget: &modelWithStringSet{
				HttpsPort:             types.Int32Value(443),
				XMLWrappedStringSlice: fwtypes.NewSetValueOfNull[types.String](ctx),
			},
		},
		"empty xml wrapper non-pointer to empty string list": {
			Source: StructWithWrappedStringSlice{
				HttpsPort: 443,
				XMLWrappedStringSlice: XMLWrappedStringSlice{
					Items:    []string{},
					Quantity: aws.Int32(0),
				},
			},
			Target: &modelWithStringList{},
			WantTarget: &modelWithStringList{
				HttpsPort:             types.Int32Value(443),
				XMLWrappedStringSlice: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Group 2 Tests: String enums with XML wrappers (pointer and non-pointer)

// XMLWrappedEnumSlice mimicks AWS types that are XML wrappers around enum slices
type XMLWrappedEnumSlice struct {
	Items    []testEnum
	Quantity *int32
}

// StructWithWrappedEnumSlicePtr mimicks AWS types that have a pointer to an XML wrapper around enum slices
type StructWithWrappedEnumSlicePtr struct {
	HttpsPort           int32
	XMLWrappedEnumSlice *XMLWrappedEnumSlice
}

// StructWithWrappedEnumSlice mimicks AWS types that have a non-pointer to an XML wrapper around enum slices
type StructWithWrappedEnumSlice struct {
	HttpsPort           int32
	XMLWrappedEnumSlice XMLWrappedEnumSlice
}

// Terraform model types with enums
type modelWithEnumSet struct {
	HttpsPort           types.Int32                                      `tfsdk:"https_port"`
	XMLWrappedEnumSlice fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]] `tfsdk:"xml_wrapped_enum_slice"`
}

type modelWithEnumList struct {
	HttpsPort           types.Int32                                       `tfsdk:"https_port"`
	XMLWrappedEnumSlice fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]] `tfsdk:"xml_wrapped_enum_slice"`
}

func TestExpandStringEnumsToXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"enum set to xml wrapper with pointer": {
			Source: modelWithEnumSet{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &StructWithWrappedEnumSlicePtr{},
			WantTarget: &StructWithWrappedEnumSlicePtr{
				HttpsPort: 443,
				XMLWrappedEnumSlice: &XMLWrappedEnumSlice{
					Items:    []testEnum{testEnumScalar, testEnumList},
					Quantity: aws.Int32(2),
				},
			},
		},
		"enum list to xml wrapper non-pointer": {
			Source: modelWithEnumList{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &StructWithWrappedEnumSlice{},
			WantTarget: &StructWithWrappedEnumSlice{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumSlice{
					Items:    []testEnum{testEnumScalar, testEnumList},
					Quantity: aws.Int32(2),
				},
			},
		},
		"null enum set to nil xml wrapper pointer": {
			Source: modelWithEnumSet{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
			},
			Target: &StructWithWrappedEnumSlicePtr{},
			WantTarget: &StructWithWrappedEnumSlicePtr{
				HttpsPort:           443,
				XMLWrappedEnumSlice: nil,
			},
		},
		"empty enum list to xml wrapper non-pointer": {
			Source: modelWithEnumList{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
			},
			Target: &StructWithWrappedEnumSlice{},
			WantTarget: &StructWithWrappedEnumSlice{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumSlice{
					Items:    []testEnum{},
					Quantity: aws.Int32(0),
				},
			},
		},
		"single enum to xml wrapper with pointer": {
			Source: modelWithEnumSet{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
				}),
			},
			Target: &StructWithWrappedEnumSlicePtr{},
			WantTarget: &StructWithWrappedEnumSlicePtr{
				HttpsPort: 443,
				XMLWrappedEnumSlice: &XMLWrappedEnumSlice{
					Items:    []testEnum{testEnumScalar},
					Quantity: aws.Int32(1),
				},
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

func TestFlattenStringEnumsFromXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"xml wrapper pointer to enum set": {
			Source: StructWithWrappedEnumSlicePtr{
				HttpsPort: 443,
				XMLWrappedEnumSlice: &XMLWrappedEnumSlice{
					Items:    []testEnum{testEnumScalar, testEnumList},
					Quantity: aws.Int32(2),
				},
			},
			Target: &modelWithEnumSet{},
			WantTarget: &modelWithEnumSet{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
		},
		"xml wrapper non-pointer to enum list": {
			Source: StructWithWrappedEnumSlice{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumSlice{
					Items:    []testEnum{testEnumScalar, testEnumList},
					Quantity: aws.Int32(2),
				},
			},
			Target: &modelWithEnumList{},
			WantTarget: &modelWithEnumList{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
		},
		"nil xml wrapper pointer to null enum set": {
			Source: StructWithWrappedEnumSlicePtr{
				HttpsPort:           443,
				XMLWrappedEnumSlice: nil,
			},
			Target: &modelWithEnumSet{},
			WantTarget: &modelWithEnumSet{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
			},
		},
		"empty xml wrapper non-pointer to empty enum list": {
			Source: StructWithWrappedEnumSlice{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumSlice{
					Items:    []testEnum{},
					Quantity: aws.Int32(0),
				},
			},
			Target: &modelWithEnumList{},
			WantTarget: &modelWithEnumList{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
			},
		},
		"single enum from xml wrapper pointer to enum set": {
			Source: StructWithWrappedEnumSlicePtr{
				HttpsPort: 443,
				XMLWrappedEnumSlice: &XMLWrappedEnumSlice{
					Items:    []testEnum{testEnumScalar},
					Quantity: aws.Int32(1),
				},
			},
			Target: &modelWithEnumSet{},
			WantTarget: &modelWithEnumSet{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
				}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Group 3 Tests: String enums with XML wrappers that have additional fields (pointer and non-pointer)

// XMLWrappedEnumSliceOther mimicks AWS types that are XML wrappers around enum slices but include
// additional fields ("Other" here)
type XMLWrappedEnumSliceOther struct {
	Other    *XMLWrappedEnumSlice
	Items    []testEnum
	Quantity *int32
}

// StructWithWrappedEnumSliceOtherPtr mimicks AWS types that have a pointer to an XML wrapper around enum slices
type StructWithWrappedEnumSliceOtherPtr struct {
	HttpsPort           int32
	XMLWrappedEnumSlice *XMLWrappedEnumSliceOther
}

// StructWithWrappedEnumSliceOther mimicks AWS types that have a non-pointer to an XML wrapper around enum slices
type StructWithWrappedEnumSliceOther struct {
	HttpsPort           int32
	XMLWrappedEnumSlice XMLWrappedEnumSliceOther
}

// modelWithEnumSetOther mimicks TF models that have an enum set and an additional enum set field
type modelWithEnumSetOther struct {
	HttpsPort           types.Int32                                      `tfsdk:"https_port"`
	XMLWrappedEnumSlice fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]] `tfsdk:"xml_wrapped_enum_slice"`
	Other               fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]] `tfsdk:"other"`
}

// modelWithEnumListOther mimicks TF models that have an enum list and an additional enum list field
type modelWithEnumListOther struct {
	HttpsPort           types.Int32                                       `tfsdk:"https_port"`
	XMLWrappedEnumSlice fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]] `tfsdk:"xml_wrapped_enum_slice"`
	Other               fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]] `tfsdk:"other"`
}

func TestExpandStringEnumsToXMLWrapperWithAdditionalFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"enum set with other field to xml wrapper with pointer": {
			Source: modelWithEnumSetOther{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
				Other: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
				}),
			},
			Target: &StructWithWrappedEnumSliceOtherPtr{},
			WantTarget: &StructWithWrappedEnumSliceOtherPtr{
				HttpsPort: 443,
				XMLWrappedEnumSlice: &XMLWrappedEnumSliceOther{
					Items:    []testEnum{testEnumScalar, testEnumList},
					Quantity: aws.Int32(2),
					Other: &XMLWrappedEnumSlice{
						Items:    []testEnum{testEnumScalar},
						Quantity: aws.Int32(1),
					},
				},
			},
		},
		"enum list with other field to xml wrapper non-pointer": {
			Source: modelWithEnumListOther{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
				Other: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &StructWithWrappedEnumSliceOther{},
			WantTarget: &StructWithWrappedEnumSliceOther{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumSliceOther{
					Items:    []testEnum{testEnumScalar, testEnumList},
					Quantity: aws.Int32(2),
					Other: &XMLWrappedEnumSlice{
						Items:    []testEnum{testEnumList},
						Quantity: aws.Int32(1),
					},
				},
			},
		},
		"null enum set with null other to nil xml wrapper pointer": {
			Source: modelWithEnumSetOther{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
				Other:               fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
			},
			Target: &StructWithWrappedEnumSliceOtherPtr{},
			WantTarget: &StructWithWrappedEnumSliceOtherPtr{
				HttpsPort:           443,
				XMLWrappedEnumSlice: nil,
			},
		},
		"empty enum list with empty other to xml wrapper non-pointer": {
			Source: modelWithEnumListOther{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
				Other:               fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
			},
			Target: &StructWithWrappedEnumSliceOther{},
			WantTarget: &StructWithWrappedEnumSliceOther{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumSliceOther{
					Items:    []testEnum{},
					Quantity: aws.Int32(0),
					Other: &XMLWrappedEnumSlice{
						Items:    []testEnum{},
						Quantity: aws.Int32(0),
					},
				},
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

func TestFlattenStringEnumsFromXMLWrapperWithAdditionalFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"xml wrapper pointer with other field to enum set": {
			Source: StructWithWrappedEnumSliceOtherPtr{
				HttpsPort: 443,
				XMLWrappedEnumSlice: &XMLWrappedEnumSliceOther{
					Items:    []testEnum{testEnumScalar, testEnumList},
					Quantity: aws.Int32(2),
					Other: &XMLWrappedEnumSlice{
						Items:    []testEnum{testEnumScalar},
						Quantity: aws.Int32(1),
					},
				},
			},
			Target: &modelWithEnumSetOther{},
			WantTarget: &modelWithEnumSetOther{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
				Other: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
				}),
			},
		},
		"xml wrapper non-pointer with other field to enum list": {
			Source: StructWithWrappedEnumSliceOther{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumSliceOther{
					Items:    []testEnum{testEnumScalar, testEnumList},
					Quantity: aws.Int32(2),
					Other: &XMLWrappedEnumSlice{
						Items:    []testEnum{testEnumList},
						Quantity: aws.Int32(1),
					},
				},
			},
			Target: &modelWithEnumListOther{},
			WantTarget: &modelWithEnumListOther{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
				Other: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
		},
		"nil xml wrapper pointer with other field to null enum set": {
			Source: StructWithWrappedEnumSliceOtherPtr{
				HttpsPort:           443,
				XMLWrappedEnumSlice: nil,
			},
			Target: &modelWithEnumSetOther{},
			WantTarget: &modelWithEnumSetOther{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
				Other:               fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
			},
		},
		"empty xml wrapper non-pointer with empty other to empty enum list": {
			Source: StructWithWrappedEnumSliceOther{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumSliceOther{
					Items:    []testEnum{},
					Quantity: aws.Int32(0),
					Other: &XMLWrappedEnumSlice{
						Items:    []testEnum{},
						Quantity: aws.Int32(0),
					},
				},
			},
			Target: &modelWithEnumListOther{},
			WantTarget: &modelWithEnumListOther{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
				Other:               fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Group 4 Tests: String enums with XML wrappers that have pointer slices (pointer and non-pointer)
// Uses the same model types as Group 2 (modelWithEnumSet and modelWithEnumList)

// XMLWrappedEnumPtrSlice mimicks AWS types that are XML wrappers around enum pointer slices
type XMLWrappedEnumPtrSlice struct {
	Items    []*testEnum
	Quantity *int32
}

// StructWithWrappedEnumPtrSliceOtherPtr mimicks AWS types that have a pointer to an XML wrapper around enum ptr slices
type StructWithWrappedEnumPtrSliceOtherPtr struct {
	HttpsPort           int32
	XMLWrappedEnumSlice *XMLWrappedEnumPtrSlice
}

// StructWithWrappedEnumPtrSliceOther mimicks AWS types that have a non-pointer to an XML wrapper around enum slices
type StructWithWrappedEnumPtrSliceOther struct {
	HttpsPort           int32
	XMLWrappedEnumSlice XMLWrappedEnumPtrSlice
}

func TestExpandStringEnumsToXMLWrapperWithPointerSlices(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"enum set to xml wrapper with pointer slice and wrapper pointer": {
			Source: modelWithEnumSet{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &StructWithWrappedEnumPtrSliceOtherPtr{},
			WantTarget: &StructWithWrappedEnumPtrSliceOtherPtr{
				HttpsPort: 443,
				XMLWrappedEnumSlice: &XMLWrappedEnumPtrSlice{
					Items: []*testEnum{
						&[]testEnum{testEnumScalar}[0],
						&[]testEnum{testEnumList}[0],
					},
					Quantity: aws.Int32(2),
				},
			},
		},
		"enum list to xml wrapper with pointer slice non-pointer": {
			Source: modelWithEnumList{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &StructWithWrappedEnumPtrSliceOther{},
			WantTarget: &StructWithWrappedEnumPtrSliceOther{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumPtrSlice{
					Items: []*testEnum{
						&[]testEnum{testEnumScalar}[0],
						&[]testEnum{testEnumList}[0],
					},
					Quantity: aws.Int32(2),
				},
			},
		},
		"null enum set to nil xml wrapper with pointer slice": {
			Source: modelWithEnumSet{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
			},
			Target: &StructWithWrappedEnumPtrSliceOtherPtr{},
			WantTarget: &StructWithWrappedEnumPtrSliceOtherPtr{
				HttpsPort:           443,
				XMLWrappedEnumSlice: nil,
			},
		},
		"empty enum list to xml wrapper with empty pointer slice": {
			Source: modelWithEnumList{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
			},
			Target: &StructWithWrappedEnumPtrSliceOther{},
			WantTarget: &StructWithWrappedEnumPtrSliceOther{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumPtrSlice{
					Items:    []*testEnum{},
					Quantity: aws.Int32(0),
				},
			},
		},
		"single enum to xml wrapper with single pointer slice": {
			Source: modelWithEnumSet{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
				}),
			},
			Target: &StructWithWrappedEnumPtrSliceOtherPtr{},
			WantTarget: &StructWithWrappedEnumPtrSliceOtherPtr{
				HttpsPort: 443,
				XMLWrappedEnumSlice: &XMLWrappedEnumPtrSlice{
					Items: []*testEnum{
						&[]testEnum{testEnumScalar}[0],
					},
					Quantity: aws.Int32(1),
				},
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

func TestFlattenStringEnumsFromXMLWrapperWithPointerSlices(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"xml wrapper with pointer slice to enum set": {
			Source: StructWithWrappedEnumPtrSliceOtherPtr{
				HttpsPort: 443,
				XMLWrappedEnumSlice: &XMLWrappedEnumPtrSlice{
					Items: []*testEnum{
						&[]testEnum{testEnumScalar}[0],
						&[]testEnum{testEnumList}[0],
					},
					Quantity: aws.Int32(2),
				},
			},
			Target: &modelWithEnumSet{},
			WantTarget: &modelWithEnumSet{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
		},
		"xml wrapper non-pointer with pointer slice to enum list": {
			Source: StructWithWrappedEnumPtrSliceOther{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumPtrSlice{
					Items: []*testEnum{
						&[]testEnum{testEnumScalar}[0],
						&[]testEnum{testEnumList}[0],
					},
					Quantity: aws.Int32(2),
				},
			},
			Target: &modelWithEnumList{},
			WantTarget: &modelWithEnumList{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
		},
		"nil xml wrapper with pointer slice to null enum set": {
			Source: StructWithWrappedEnumPtrSliceOtherPtr{
				HttpsPort:           443,
				XMLWrappedEnumSlice: nil,
			},
			Target: &modelWithEnumSet{},
			WantTarget: &modelWithEnumSet{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
			},
		},
		"empty xml wrapper with empty pointer slice to empty enum list": {
			Source: StructWithWrappedEnumPtrSliceOther{
				HttpsPort: 443,
				XMLWrappedEnumSlice: XMLWrappedEnumPtrSlice{
					Items:    []*testEnum{},
					Quantity: aws.Int32(0),
				},
			},
			Target: &modelWithEnumList{},
			WantTarget: &modelWithEnumList{
				HttpsPort:           types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
			},
		},
		"single xml wrapper with single pointer slice to single enum set": {
			Source: StructWithWrappedEnumPtrSliceOtherPtr{
				HttpsPort: 443,
				XMLWrappedEnumSlice: &XMLWrappedEnumPtrSlice{
					Items: []*testEnum{
						&[]testEnum{testEnumScalar}[0],
					},
					Quantity: aws.Int32(1),
				},
			},
			Target: &modelWithEnumSet{},
			WantTarget: &modelWithEnumSet{
				HttpsPort: types.Int32Value(443),
				XMLWrappedEnumSlice: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
				}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

// DistributionConfigSimple represents a CloudFront distribution configuration from the AWS SDK.
// This mirrors the actual AWS API structure with nested XML wrapper types that include
// both Items (slice of data) and Quantity (count) fields.
type DistributionConfigSimple struct {
	Comment                    *string
	DefaultCacheBehaviorSimple *DefaultCacheBehaviorSimple
}

// DefaultCacheBehaviorSimple represents the default cache behavior settings for a CloudFront distribution.
// Contains AllowedMethodsSimple which is an XML wrapper type - this tests the real-world scenario
// where AWS APIs return complex nested structures with XML wrappers at multiple levels.
type DefaultCacheBehaviorSimple struct {
	AllowedMethodsSimple *AllowedMethodsSimple
	Compress             *bool
}

// AllowedMethodsSimple represents the XML wrapper pattern.
// This type tests:
// - Items: []MethodSimple - slice of enum values that need to be converted to/from Terraform sets
// - Quantity: *int32 - metadata field that AutoFlex should ignore during conversion
// - CachedMethodsSimple: nested XML wrapper - tests recursive XML wrapper handling
// This structure is critical for validating AutoFlex's ability to handle real XML wrapper complexity.
type AllowedMethodsSimple struct {
	Items               []MethodSimple
	Quantity            *int32
	CachedMethodsSimple *CachedMethodsSimple
}

// MethodSimple represents HTTP methods as string enums
type MethodSimple string

const (
	MethodSimpleGet  MethodSimple = "GET"
	MethodSimpleHead MethodSimple = "HEAD"
)

// Values returns all possible values of MethodSimple
func (m MethodSimple) Values() []MethodSimple {
	return []MethodSimple{
		MethodSimpleGet,
		MethodSimpleHead,
	}
}

// CachedMethodsSimple represents a nested XML wrapper within AllowedMethodsSimple.
// This tests AutoFlex's ability to handle XML wrappers that are:
// - Nested inside other XML wrappers (AllowedMethodsSimple.CachedMethodsSimple)
// - Contain the same enum type (MethodSimple) as the parent wrapper
// - Require independent Items/Quantity handling at each nesting level
type CachedMethodsSimple struct {
	Items    []MethodSimple
	Quantity *int32
}

// multiTenantDistributionResourceModelSimple represents the Terraform resource model structure.
type multiTenantDistributionResourceModelSimple struct {
	Comment                    types.String                                                     `tfsdk:"comment"`
	DefaultCacheBehaviorSimple fwtypes.ListNestedObjectValueOf[defaultCacheBehaviorModelSimple] `tfsdk:"default_cache_behavior"`
}

// defaultCacheBehaviorModelSimple represents the Terraform model for default cache behavior.
// This type tests AutoFlex's conversion of XML wrapper Items to Terraform sets:
// - AllowedMethodsSimple: converts AllowedMethodsSimple.Items []MethodSimple to SetValueOf[StringEnum[MethodSimple]]
// - CachedMethodsSimple: converts nested CachedMethodsSimple.Items []MethodSimple to SetValueOf[StringEnum[MethodSimple]]
// - Compress: simple bool conversion
// This validates that AutoFlex can handle multiple XML wrappers in the same object.
type defaultCacheBehaviorModelSimple struct {
	AllowedMethodsSimple fwtypes.SetValueOf[fwtypes.StringEnum[MethodSimple]] `tfsdk:"allowed_methods"`
	Compress             types.Bool                                           `tfsdk:"compress"`
	CachedMethodsSimple  fwtypes.SetValueOf[fwtypes.StringEnum[MethodSimple]] `tfsdk:"cached_methods"`
}

func TestExpandXMLRealWorld(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"full distribution with allowed and cached methods": {
			Source: multiTenantDistributionResourceModelSimple{
				Comment: types.StringValue("Test distribution"),
				DefaultCacheBehaviorSimple: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModelSimple{
					{
						AllowedMethodsSimple: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{
							fwtypes.StringEnumValue(MethodSimpleGet),
							fwtypes.StringEnumValue(MethodSimpleHead),
						}),
						Compress: types.BoolValue(true),
						CachedMethodsSimple: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{
							fwtypes.StringEnumValue(MethodSimpleGet),
						}),
					},
				}),
			},
			Target: &DistributionConfigSimple{},
			WantTarget: &DistributionConfigSimple{
				Comment: aws.String("Test distribution"),
				DefaultCacheBehaviorSimple: &DefaultCacheBehaviorSimple{
					AllowedMethodsSimple: &AllowedMethodsSimple{
						Items:    []MethodSimple{MethodSimpleGet, MethodSimpleHead},
						Quantity: aws.Int32(2),
						CachedMethodsSimple: &CachedMethodsSimple{
							Items:    []MethodSimple{MethodSimpleGet},
							Quantity: aws.Int32(1),
						},
					},
					Compress: aws.Bool(true),
				},
			},
		},
		"distribution with only allowed methods": {
			Source: multiTenantDistributionResourceModelSimple{
				Comment: types.StringValue("Minimal distribution"),
				DefaultCacheBehaviorSimple: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModelSimple{
					{
						AllowedMethodsSimple: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{
							fwtypes.StringEnumValue(MethodSimpleHead),
						}),
						Compress:            types.BoolValue(false),
						CachedMethodsSimple: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{}),
					},
				}),
			},
			Target: &DistributionConfigSimple{},
			WantTarget: &DistributionConfigSimple{
				Comment: aws.String("Minimal distribution"),
				DefaultCacheBehaviorSimple: &DefaultCacheBehaviorSimple{
					AllowedMethodsSimple: &AllowedMethodsSimple{
						Items:    []MethodSimple{MethodSimpleHead},
						Quantity: aws.Int32(1),
						CachedMethodsSimple: &CachedMethodsSimple{
							Items:    []MethodSimple{},
							Quantity: aws.Int32(0),
						},
					},
					Compress: aws.Bool(false),
				},
			},
		},
		"distribution with empty allowed methods": {
			Source: multiTenantDistributionResourceModelSimple{
				Comment: types.StringValue("Empty methods distribution"),
				DefaultCacheBehaviorSimple: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModelSimple{
					{
						AllowedMethodsSimple: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{}),
						Compress:             types.BoolValue(true),
						CachedMethodsSimple:  fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{}),
					},
				}),
			},
			Target: &DistributionConfigSimple{},
			WantTarget: &DistributionConfigSimple{
				Comment: aws.String("Empty methods distribution"),
				DefaultCacheBehaviorSimple: &DefaultCacheBehaviorSimple{
					AllowedMethodsSimple: &AllowedMethodsSimple{
						Items:    []MethodSimple{},
						Quantity: aws.Int32(0),
						CachedMethodsSimple: &CachedMethodsSimple{
							Items:    []MethodSimple{},
							Quantity: aws.Int32(0),
						},
					},
					Compress: aws.Bool(true),
				},
			},
		},
		"distribution with null cache behavior": {
			Source: multiTenantDistributionResourceModelSimple{
				Comment:                    types.StringValue("Null behavior distribution"),
				DefaultCacheBehaviorSimple: fwtypes.NewListNestedObjectValueOfNull[defaultCacheBehaviorModelSimple](ctx),
			},
			Target: &DistributionConfigSimple{},
			WantTarget: &DistributionConfigSimple{
				Comment:                    aws.String("Null behavior distribution"),
				DefaultCacheBehaviorSimple: nil,
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}

func TestFlattenXMLRealWorld(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"full distribution with allowed and cached methods": {
			Source: &DistributionConfigSimple{
				Comment: aws.String("Test distribution"),
				DefaultCacheBehaviorSimple: &DefaultCacheBehaviorSimple{
					AllowedMethodsSimple: &AllowedMethodsSimple{
						Items:    []MethodSimple{MethodSimpleGet, MethodSimpleHead},
						Quantity: aws.Int32(2),
						CachedMethodsSimple: &CachedMethodsSimple{
							Items:    []MethodSimple{MethodSimpleGet},
							Quantity: aws.Int32(1),
						},
					},
					Compress: aws.Bool(true),
				},
			},
			Target: &multiTenantDistributionResourceModelSimple{},
			WantTarget: &multiTenantDistributionResourceModelSimple{
				Comment: types.StringValue("Test distribution"),
				DefaultCacheBehaviorSimple: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModelSimple{
					{
						AllowedMethodsSimple: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{
							fwtypes.StringEnumValue(MethodSimpleGet),
							fwtypes.StringEnumValue(MethodSimpleHead),
						}),
						Compress: types.BoolValue(true),
						CachedMethodsSimple: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{
							fwtypes.StringEnumValue(MethodSimpleGet),
						}),
					},
				}),
			},
		},
		"distribution with only allowed methods": {
			Source: &DistributionConfigSimple{
				Comment: aws.String("Minimal distribution"),
				DefaultCacheBehaviorSimple: &DefaultCacheBehaviorSimple{
					AllowedMethodsSimple: &AllowedMethodsSimple{
						Items:    []MethodSimple{MethodSimpleHead},
						Quantity: aws.Int32(1),
						CachedMethodsSimple: &CachedMethodsSimple{
							Items:    []MethodSimple{},
							Quantity: aws.Int32(0),
						},
					},
					Compress: aws.Bool(false),
				},
			},
			Target: &multiTenantDistributionResourceModelSimple{},
			WantTarget: &multiTenantDistributionResourceModelSimple{
				Comment: types.StringValue("Minimal distribution"),
				DefaultCacheBehaviorSimple: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModelSimple{
					{
						AllowedMethodsSimple: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{
							fwtypes.StringEnumValue(MethodSimpleHead),
						}),
						Compress:            types.BoolValue(false),
						CachedMethodsSimple: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{}),
					},
				}),
			},
		},
		"distribution with empty allowed methods": {
			Source: &DistributionConfigSimple{
				Comment: aws.String("Empty methods distribution"),
				DefaultCacheBehaviorSimple: &DefaultCacheBehaviorSimple{
					AllowedMethodsSimple: &AllowedMethodsSimple{
						Items:    []MethodSimple{},
						Quantity: aws.Int32(0),
						CachedMethodsSimple: &CachedMethodsSimple{
							Items:    []MethodSimple{},
							Quantity: aws.Int32(0),
						},
					},
					Compress: aws.Bool(true),
				},
			},
			Target: &multiTenantDistributionResourceModelSimple{},
			WantTarget: &multiTenantDistributionResourceModelSimple{
				Comment: types.StringValue("Empty methods distribution"),
				DefaultCacheBehaviorSimple: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModelSimple{
					{
						AllowedMethodsSimple: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{}),
						Compress:             types.BoolValue(true),
						CachedMethodsSimple:  fwtypes.NewSetValueOfMust[fwtypes.StringEnum[MethodSimple]](ctx, []attr.Value{}),
					},
				}),
			},
		},
		"distribution with nil cache behavior": {
			Source: &DistributionConfigSimple{
				Comment:                    aws.String("Nil behavior distribution"),
				DefaultCacheBehaviorSimple: nil,
			},
			Target: &multiTenantDistributionResourceModelSimple{},
			WantTarget: &multiTenantDistributionResourceModelSimple{
				Comment:                    types.StringValue("Nil behavior distribution"),
				DefaultCacheBehaviorSimple: fwtypes.NewListNestedObjectValueOfNull[defaultCacheBehaviorModelSimple](ctx),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Full complete AWS types

type DistributionConfig struct {
	Aliases                      *Aliases
	AnycastIpListId              *string
	CacheBehaviors               *CacheBehaviors
	CallerReference              *string // required
	Comment                      *string // required
	ConnectionMode               ConnectionMode
	ContinuousDeploymentPolicyId *string
	CustomErrorResponses         *CustomErrorResponses
	DefaultCacheBehavior         *DefaultCacheBehavior // required
	DefaultRootObject            *string
	Enabled                      *bool // required
	HttpVersion                  HttpVersion
	IsIPV6Enabled                *bool
	Logging                      *LoggingConfig
	OriginGroups                 *OriginGroups
	Origins                      *Origins // required
	PriceClass                   PriceClass
	Restrictions                 *Restrictions
	Staging                      *bool
	TenantConfig                 *TenantConfig
	ViewerCertificate            *ViewerCertificate
	WebACLId                     *string
}

type DefaultCacheBehavior struct {
	AllowedMethods             *AllowedMethods
	CachePolicyId              *string
	Compress                   *bool
	DefaultTTL                 *int64
	FieldLevelEncryptionId     *string
	ForwardedValues            *ForwardedValues
	FunctionAssociations       *FunctionAssociations
	GrpcConfig                 *GrpcConfig
	LambdaFunctionAssociations *LambdaFunctionAssociations
	MaxTTL                     *int64
	MinTTL                     *int64
	OriginRequestPolicyId      *string
	RealtimeLogConfigArn       *string
	ResponseHeadersPolicyId    *string
	SmoothStreaming            *bool
	TargetOriginId             *string // required
	TrustedKeyGroups           *TrustedKeyGroups
	TrustedSigners             *TrustedSigners
	ViewerProtocolPolicy       ViewerProtocolPolicy // required
}

type AllowedMethods struct {
	CachedMethods *CachedMethods
	Items         []Method // required
	Quantity      *int32   // required
}

type Method string

const (
	MethodGet     Method = "GET"
	MethodHead    Method = "HEAD"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodPatch   Method = "PATCH"
	MethodOptions Method = "OPTIONS"
	MethodDelete  Method = "DELETE"
)

// Values returns all possible values of Method
func (m Method) Values() []Method {
	return []Method{
		MethodGet,
		MethodHead,
		MethodPost,
		MethodPut,
		MethodPatch,
		MethodOptions,
		MethodDelete,
	}
}

type CachedMethods struct {
	Items    []Method // required
	Quantity *int32   // required
}

type ForwardedValues struct { // deprecated
	Cookies              *CookiePreference // required
	Headers              *Headers
	QueryString          *bool // required
	QueryStringCacheKeys *QueryStringCacheKeys
}

type CookiePreference struct { // deprecated
	Forward          ItemSelection // required
	WhitelistedNames *CookieNames
}

type ItemSelection string

const (
	ItemSelectionNone      ItemSelection = "none"
	ItemSelectionWhitelist ItemSelection = "whitelist"
	ItemSelectionAll       ItemSelection = "all"
)

type CookieNames struct {
	Quantity *int32 // required
	Items    []string
}

type Headers struct {
	Items    []string
	Quantity *int32 // required
}

type QueryStringCacheKeys struct {
	Items    []string
	Quantity *int32 // required
}

type GrpcConfig struct {
	Enabled *bool // required
}

type LambdaFunctionAssociations struct {
	Quantity *int32 // required
	Items    []LambdaFunctionAssociation
}

type LambdaFunctionAssociation struct {
	EventType         EventType // required
	IncludeBody       *bool
	LambdaFunctionARN *string // required
}

type EventType string

const (
	EventTypeViewerRequest  EventType = "viewer-request"
	EventTypeViewerResponse EventType = "viewer-response"
	EventTypeOriginRequest  EventType = "origin-request"
	EventTypeOriginResponse EventType = "origin-response"
)

// Values returns all possible values of EventType
func (e EventType) Values() []EventType {
	return []EventType{
		EventTypeViewerRequest,
		EventTypeViewerResponse,
		EventTypeOriginRequest,
		EventTypeOriginResponse,
	}
}

type TrustedKeyGroups struct {
	Enabled  *bool // required
	Items    []string
	Quantity *int32 // required
}

type TrustedSigners struct {
	Enabled  *bool  // required
	Quantity *int32 // required
	Items    []string
}

type ViewerProtocolPolicy string

const (
	ViewerProtocolPolicyAllowAll        ViewerProtocolPolicy = "allow-all"
	ViewerProtocolPolicyHttpsOnly       ViewerProtocolPolicy = "https-only"
	ViewerProtocolPolicyRedirectToHttps ViewerProtocolPolicy = "redirect-to-https"
)

// Values returns all possible values of ViewerProtocolPolicy
func (v ViewerProtocolPolicy) Values() []ViewerProtocolPolicy {
	return []ViewerProtocolPolicy{
		ViewerProtocolPolicyAllowAll,
		ViewerProtocolPolicyHttpsOnly,
		ViewerProtocolPolicyRedirectToHttps,
	}
}

type Origins struct {
	Items    []Origin // required
	Quantity *int32   // required
}

type Origin struct {
	ConnectionAttempts        *int32
	ConnectionTimeout         *int32
	CustomHeaders             *CustomHeaders
	CustomOriginConfig        *CustomOriginConfig
	DomainName                *string // required
	Id                        *string // required
	OriginAccessControlId     *string
	OriginPath                *string
	OriginShield              *OriginShield
	ResponseCompletionTimeout *int32
	S3OriginConfig            *S3OriginConfig
	VpcOriginConfig           *VpcOriginConfig
}

type CustomHeaders struct {
	Quantity *int32 // required
	Items    []OriginCustomHeader
}

type OriginCustomHeader struct {
	HeaderName  *string // required
	HeaderValue *string // required
}

type CustomOriginConfig struct {
	HTTPPort               *int32 // required
	HTTPSPort              *int32 // required
	IpAddressType          IpAddressType
	OriginKeepaliveTimeout *int32
	OriginProtocolPolicy   OriginProtocolPolicy // required
	OriginReadTimeout      *int32
	OriginSslProtocols     *OriginSslProtocols
}

type IpAddressType string

const (
	IpAddressTypeIpv4      IpAddressType = "ipv4"
	IpAddressTypeIpv6      IpAddressType = "ipv6"
	IpAddressTypeDualStack IpAddressType = "dualstack"
)

// Values returns all possible values of IpAddressType
func (i IpAddressType) Values() []IpAddressType {
	return []IpAddressType{
		IpAddressTypeIpv4,
		IpAddressTypeIpv6,
		IpAddressTypeDualStack,
	}
}

type OriginProtocolPolicy string

const (
	OriginProtocolPolicyHttpOnly    OriginProtocolPolicy = "http-only"
	OriginProtocolPolicyMatchViewer OriginProtocolPolicy = "match-viewer"
	OriginProtocolPolicyHttpsOnly   OriginProtocolPolicy = "https-only"
)

// Values returns all possible values of OriginProtocolPolicy
func (o OriginProtocolPolicy) Values() []OriginProtocolPolicy {
	return []OriginProtocolPolicy{
		OriginProtocolPolicyHttpOnly,
		OriginProtocolPolicyMatchViewer,
		OriginProtocolPolicyHttpsOnly,
	}
}

type OriginSslProtocols struct {
	Items    []SslProtocol // required
	Quantity *int32        // required
}

type SslProtocol string

const (
	SslProtocolSSLv3  SslProtocol = "SSLv3"
	SslProtocolTLSv1  SslProtocol = "TLSv1"
	SslProtocolTLSv11 SslProtocol = "TLSv1.1"
	SslProtocolTLSv12 SslProtocol = "TLSv1.2"
)

// Values returns all possible values of SslProtocol
func (s SslProtocol) Values() []SslProtocol {
	return []SslProtocol{
		SslProtocolSSLv3,
		SslProtocolTLSv1,
		SslProtocolTLSv11,
		SslProtocolTLSv12,
	}
}

type OriginShield struct {
	Enabled            *bool // required
	OriginShieldRegion *string
}

type S3OriginConfig struct {
	OriginAccessIdentity *string // required
	OriginReadTimeout    *int32
}

type VpcOriginConfig struct {
	OriginKeepaliveTimeout *int32
	OriginReadTimeout      *int32
	VpcOriginId            *string // required
}

type Aliases struct {
	Items    []string
	Quantity *int32 // required
}

type CacheBehaviors struct {
	Quantity *int32 // required
	Items    []CacheBehavior
}

type CacheBehavior struct {
	AllowedMethods             *AllowedMethods
	CachePolicyId              *string
	Compress                   *bool
	DefaultTTL                 *int64 // deprecated
	FieldLevelEncryptionId     *string
	ForwardedValues            *ForwardedValues // deprecated
	FunctionAssociations       *FunctionAssociations
	GrpcConfig                 *GrpcConfig
	LambdaFunctionAssociations *LambdaFunctionAssociations
	MaxTTL                     *int64 // deprecated
	MinTTL                     *int64 // deprecated
	OriginRequestPolicyId      *string
	PathPattern                *string // required
	RealtimeLogConfigArn       *string
	ResponseHeadersPolicyId    *string
	SmoothStreaming            *bool
	TargetOriginId             *string // required
	TrustedKeyGroups           *TrustedKeyGroups
	TrustedSigners             *TrustedSigners
	ViewerProtocolPolicy       ViewerProtocolPolicy // required
}

type ConnectionMode string

const (
	ConnectionModeDirect     ConnectionMode = "direct"
	ConnectionModeTenantOnly ConnectionMode = "tenant-only"
)

// Values returns all possible values of ConnectionMode
func (c ConnectionMode) Values() []ConnectionMode {
	return []ConnectionMode{
		ConnectionModeDirect,
		ConnectionModeTenantOnly,
	}
}

type CustomErrorResponses struct {
	Quantity *int32 // required
	Items    []CustomErrorResponse
}

type CustomErrorResponse struct {
	ErrorCachingMinTTL *int64
	ErrorCode          *int32 // required
	ResponseCode       *string
	ResponsePagePath   *string
}

type HttpVersion string

const (
	HttpVersionHttp11    HttpVersion = "http1.1"
	HttpVersionHttp2     HttpVersion = "http2"
	HttpVersionHttp3     HttpVersion = "http3"
	HttpVersionHttp2and3 HttpVersion = "http2and3"
)

// Values returns all possible values of HttpVersion
func (h HttpVersion) Values() []HttpVersion {
	return []HttpVersion{
		HttpVersionHttp11,
		HttpVersionHttp2,
		HttpVersionHttp3,
		HttpVersionHttp2and3,
	}
}

type LoggingConfig struct {
	Bucket         *string
	Enabled        *bool
	IncludeCookies *bool
	Prefix         *string
}

type OriginGroups struct {
	Quantity *int32 // required
	Items    []OriginGroup
}

type OriginGroup struct {
	FailoverCriteria  *OriginGroupFailoverCriteria // required
	Id                *string                      // required
	Members           *OriginGroupMembers          // required
	SelectionCriteria OriginGroupSelectionCriteria
}

type OriginGroupFailoverCriteria struct {
	StatusCodes *StatusCodes // required
}

type StatusCodes struct {
	Items    []int32 // required
	Quantity *int32  // required
}

type OriginGroupMembers struct {
	Items    []OriginGroupMember // required
	Quantity *int32              // required
}

type OriginGroupMember struct {
	OriginId *string // required
}

type OriginGroupSelectionCriteria string

const (
	OriginGroupSelectionCriteriaDefault           OriginGroupSelectionCriteria = "default"
	OriginGroupSelectionCriteriaMediaQualityBased OriginGroupSelectionCriteria = "media-quality-based"
)

// Values returns all possible values of OriginGroupSelectionCriteria
func (o OriginGroupSelectionCriteria) Values() []OriginGroupSelectionCriteria {
	return []OriginGroupSelectionCriteria{
		OriginGroupSelectionCriteriaDefault,
		OriginGroupSelectionCriteriaMediaQualityBased,
	}
}

type PriceClass string

const (
	PriceClassPriceClass100 PriceClass = "PriceClass_100"
	PriceClassPriceClass200 PriceClass = "PriceClass_200"
	PriceClassPriceClassAll PriceClass = "PriceClass_All"
	PriceClassNone          PriceClass = "None"
)

// Values returns all possible values of PriceClass
func (p PriceClass) Values() []PriceClass {
	return []PriceClass{
		PriceClassPriceClass100,
		PriceClassPriceClass200,
		PriceClassPriceClassAll,
		PriceClassNone,
	}
}

type Restrictions struct {
	GeoRestriction *GeoRestriction // required
}

type GeoRestriction struct {
	Items           []string
	Quantity        *int32             // required
	RestrictionType GeoRestrictionType // required
}

type GeoRestrictionType string

const (
	GeoRestrictionTypeBlacklist GeoRestrictionType = "blacklist"
	GeoRestrictionTypeWhitelist GeoRestrictionType = "whitelist"
	GeoRestrictionTypeNone      GeoRestrictionType = "none"
)

// Values returns all possible values of GeoRestrictionType
func (g GeoRestrictionType) Values() []GeoRestrictionType {
	return []GeoRestrictionType{
		GeoRestrictionTypeBlacklist,
		GeoRestrictionTypeWhitelist,
		GeoRestrictionTypeNone,
	}
}

type TenantConfig struct {
	ParameterDefinitions []ParameterDefinition
}

type ParameterDefinition struct {
	Definition *ParameterDefinitionSchema // required
	Name       *string                    // required
}

type ParameterDefinitionSchema struct {
	StringSchema *StringSchemaConfig
}

type StringSchemaConfig struct {
	Comment      *string
	DefaultValue *string
	Required     *bool // required
}

type ViewerCertificate struct {
	ACMCertificateArn            *string
	Certificate                  *string           // deprecated
	CertificateSource            CertificateSource // deprecated
	CloudFrontDefaultCertificate *bool
	IAMCertificateId             *string
	MinimumProtocolVersion       MinimumProtocolVersion
	SSLSupportMethod             SSLSupportMethod
}

type CertificateSource string

const (
	CertificateSourceCloudfront CertificateSource = "cloudfront"
	CertificateSourceIam        CertificateSource = "iam"
	CertificateSourceAcm        CertificateSource = "acm"
)

// Values returns all possible values of CertificateSource
func (c CertificateSource) Values() []CertificateSource {
	return []CertificateSource{
		CertificateSourceCloudfront,
		CertificateSourceIam,
		CertificateSourceAcm,
	}
}

type MinimumProtocolVersion string

const (
	MinimumProtocolVersionSSLv3      MinimumProtocolVersion = "SSLv3"
	MinimumProtocolVersionTLSv1      MinimumProtocolVersion = "TLSv1"
	MinimumProtocolVersionTLSv12016  MinimumProtocolVersion = "TLSv1_2016"
	MinimumProtocolVersionTLSv112016 MinimumProtocolVersion = "TLSv1.1_2016"
	MinimumProtocolVersionTLSv122018 MinimumProtocolVersion = "TLSv1.2_2018"
	MinimumProtocolVersionTLSv122019 MinimumProtocolVersion = "TLSv1.2_2019"
	MinimumProtocolVersionTLSv122021 MinimumProtocolVersion = "TLSv1.2_2021"
	MinimumProtocolVersionTLSv132025 MinimumProtocolVersion = "TLSv1.3_2025"
)

// Values returns all possible values of MinimumProtocolVersion
func (m MinimumProtocolVersion) Values() []MinimumProtocolVersion {
	return []MinimumProtocolVersion{
		MinimumProtocolVersionSSLv3,
		MinimumProtocolVersionTLSv1,
		MinimumProtocolVersionTLSv12016,
		MinimumProtocolVersionTLSv112016,
		MinimumProtocolVersionTLSv122018,
		MinimumProtocolVersionTLSv122019,
		MinimumProtocolVersionTLSv122021,
		MinimumProtocolVersionTLSv132025,
	}
}

type SSLSupportMethod string

const (
	SSLSupportMethodSniOnly  SSLSupportMethod = "sni-only"
	SSLSupportMethodVip      SSLSupportMethod = "vip"
	SSLSupportMethodStaticIp SSLSupportMethod = "static-ip"
)

// Values returns all possible values of SSLSupportMethod
func (s SSLSupportMethod) Values() []SSLSupportMethod {
	return []SSLSupportMethod{
		SSLSupportMethodSniOnly,
		SSLSupportMethodVip,
		SSLSupportMethodStaticIp,
	}
}

// models

type multiTenantDistributionResourceModel struct {
	ARN                  types.String                                               `tfsdk:"arn"`
	CacheBehavior        fwtypes.ListNestedObjectValueOf[cacheBehaviorModel]        `tfsdk:"cache_behavior"`
	CallerReference      types.String                                               `tfsdk:"caller_reference"`
	Comment              types.String                                               `tfsdk:"comment"`
	CustomErrorResponse  fwtypes.SetNestedObjectValueOf[customErrorResponseModel]   `tfsdk:"custom_error_response"`
	DefaultCacheBehavior fwtypes.ListNestedObjectValueOf[defaultCacheBehaviorModel] `tfsdk:"default_cache_behavior"`
	DefaultRootObject    types.String                                               `tfsdk:"default_root_object"`
	Enabled              types.Bool                                                 `tfsdk:"enabled"`
	ETag                 types.String                                               `tfsdk:"etag"`
	HTTPVersion          fwtypes.StringEnum[HttpVersion]                            `tfsdk:"http_version"`
	ID                   types.String                                               `tfsdk:"id"`
	Origin               fwtypes.SetNestedObjectValueOf[originModel]                `tfsdk:"origin"`
	OriginGroup          fwtypes.SetNestedObjectValueOf[originGroupModel]           `tfsdk:"origin_group"`
	Restrictions         fwtypes.ListNestedObjectValueOf[restrictionsModel]         `tfsdk:"restrictions"`
	Status               types.String                                               `tfsdk:"status"`
	TenantConfig         fwtypes.ListNestedObjectValueOf[tenantConfigModel]         `tfsdk:"tenant_config"`
	Timeouts             timeouts.Value                                             `tfsdk:"timeouts"`
	ViewerCertificate    fwtypes.ListNestedObjectValueOf[viewerCertificateModel]    `tfsdk:"viewer_certificate"`
	WebACLID             types.String                                               `tfsdk:"web_acl_id"`
}

type originModel struct {
	ConnectionAttempts        types.Int32                                              `tfsdk:"connection_attempts"`
	ConnectionTimeout         types.Int32                                              `tfsdk:"connection_timeout"`
	CustomHeader              fwtypes.SetNestedObjectValueOf[customHeaderModel]        `tfsdk:"custom_header"`
	CustomOriginConfig        fwtypes.ListNestedObjectValueOf[customOriginConfigModel] `tfsdk:"custom_origin_config"`
	DomainName                types.String                                             `tfsdk:"domain_name"`
	ID                        types.String                                             `tfsdk:"id"`
	OriginAccessControlID     types.String                                             `tfsdk:"origin_access_control_id"`
	OriginPath                types.String                                             `tfsdk:"origin_path"`
	OriginShield              fwtypes.ListNestedObjectValueOf[originShieldModel]       `tfsdk:"origin_shield"`
	ResponseCompletionTimeout types.Int32                                              `tfsdk:"response_completion_timeout"`
	S3OriginConfig            fwtypes.ListNestedObjectValueOf[s3OriginConfigModel]     `tfsdk:"s3_origin_config"`
}

type customHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type customOriginConfigModel struct {
	HTTPPort               types.Int32                                         `tfsdk:"http_port"`
	HTTPSPort              types.Int32                                         `tfsdk:"https_port"`
	IPAddressType          fwtypes.StringEnum[IpAddressType]                   `tfsdk:"ip_address_type"`
	OriginKeepaliveTimeout types.Int32                                         `tfsdk:"origin_keepalive_timeout"`
	OriginReadTimeout      types.Int32                                         `tfsdk:"origin_read_timeout"`
	OriginProtocolPolicy   fwtypes.StringEnum[OriginProtocolPolicy]            `tfsdk:"origin_protocol_policy"`
	OriginSSLProtocols     fwtypes.SetValueOf[fwtypes.StringEnum[SslProtocol]] `tfsdk:"origin_ssl_protocols"`
}

type originShieldModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	OriginShieldRegion types.String `tfsdk:"origin_shield_region"`
}

type s3OriginConfigModel struct {
	OriginAccessIdentity types.String `tfsdk:"origin_access_identity"`
}

type originGroupModel struct {
	FailoverCriteria fwtypes.ListNestedObjectValueOf[failoverCriteriaModel] `tfsdk:"failover_criteria"`
	Member           fwtypes.ListNestedObjectValueOf[memberModel]           `tfsdk:"member"`
	OriginID         types.String                                           `tfsdk:"origin_id"`
}

type failoverCriteriaModel struct {
	StatusCodes fwtypes.SetValueOf[types.Int64] `tfsdk:"status_codes"`
}

type memberModel struct {
	OriginID types.String `tfsdk:"origin_id"`
}

type defaultCacheBehaviorModel struct {
	AllowedMethods            fwtypes.SetValueOf[fwtypes.StringEnum[Method]]                 `tfsdk:"allowed_methods"`
	CachePolicyID             types.String                                                   `tfsdk:"cache_policy_id"`
	CachedMethods             fwtypes.SetValueOf[fwtypes.StringEnum[Method]]                 `tfsdk:"cached_methods"`
	Compress                  types.Bool                                                     `tfsdk:"compress"`
	FieldLevelEncryptionID    types.String                                                   `tfsdk:"field_level_encryption_id"`
	FunctionAssociation       fwtypes.SetNestedObjectValueOf[functionAssociationModel]       `tfsdk:"function_association"`
	LambdaFunctionAssociation fwtypes.SetNestedObjectValueOf[lambdaFunctionAssociationModel] `tfsdk:"lambda_function_association"`
	OriginRequestPolicyID     types.String                                                   `tfsdk:"origin_request_policy_id"`
	RealtimeLogConfigARN      types.String                                                   `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyID   types.String                                                   `tfsdk:"response_headers_policy_id"`
	SmoothStreaming           types.Bool                                                     `tfsdk:"smooth_streaming"`
	TargetOriginID            types.String                                                   `tfsdk:"target_origin_id"`
	TrustedKeyGroups          fwtypes.ListValueOf[types.String]                              `tfsdk:"trusted_key_groups"`
	TrustedSigners            fwtypes.ListValueOf[types.String]                              `tfsdk:"trusted_signers"`
	ViewerProtocolPolicy      fwtypes.StringEnum[ViewerProtocolPolicy]                       `tfsdk:"viewer_protocol_policy"`
}

type cacheBehaviorModel struct {
	AllowedMethods            fwtypes.SetValueOf[fwtypes.StringEnum[Method]]                 `tfsdk:"allowed_methods"`
	CachePolicyID             types.String                                                   `tfsdk:"cache_policy_id"`
	CachedMethods             fwtypes.SetValueOf[fwtypes.StringEnum[Method]]                 `tfsdk:"cached_methods"`
	Compress                  types.Bool                                                     `tfsdk:"compress"`
	FieldLevelEncryptionID    types.String                                                   `tfsdk:"field_level_encryption_id"`
	FunctionAssociation       fwtypes.SetNestedObjectValueOf[functionAssociationModel]       `tfsdk:"function_association"`
	LambdaFunctionAssociation fwtypes.SetNestedObjectValueOf[lambdaFunctionAssociationModel] `tfsdk:"lambda_function_association"`
	OriginRequestPolicyID     types.String                                                   `tfsdk:"origin_request_policy_id"`
	PathPattern               types.String                                                   `tfsdk:"path_pattern"`
	RealtimeLogConfigARN      types.String                                                   `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyID   types.String                                                   `tfsdk:"response_headers_policy_id"`
	SmoothStreaming           types.Bool                                                     `tfsdk:"smooth_streaming"`
	TargetOriginID            types.String                                                   `tfsdk:"target_origin_id"`
	TrustedKeyGroups          fwtypes.ListValueOf[types.String]                              `tfsdk:"trusted_key_groups"`
	TrustedSigners            fwtypes.ListValueOf[types.String]                              `tfsdk:"trusted_signers"`
	ViewerProtocolPolicy      fwtypes.StringEnum[ViewerProtocolPolicy]                       `tfsdk:"viewer_protocol_policy"`
}

type customErrorResponseModel struct {
	ErrorCachingMinTtl types.Int64  `tfsdk:"error_caching_min_ttl"`
	ErrorCode          types.Int64  `tfsdk:"error_code"`
	ResponseCode       types.Int64  `tfsdk:"response_code"`
	ResponsePagePath   types.String `tfsdk:"response_page_path"`
}

type restrictionsModel struct {
	GeoRestriction fwtypes.ListNestedObjectValueOf[geoRestrictionModel] `tfsdk:"geo_restriction"`
}

type geoRestrictionModel struct {
	Locations       fwtypes.SetValueOf[types.String]       `tfsdk:"locations"`
	RestrictionType fwtypes.StringEnum[GeoRestrictionType] `tfsdk:"restriction_type"`
}

type viewerCertificateModel struct {
	ACMCertificateARN            types.String                               `tfsdk:"acm_certificate_arn"`
	CloudfrontDefaultCertificate types.Bool                                 `tfsdk:"cloudfront_default_certificate"`
	MinimumProtocolVersion       fwtypes.StringEnum[MinimumProtocolVersion] `tfsdk:"minimum_protocol_version"`
	SSLSupportMethod             fwtypes.StringEnum[SSLSupportMethod]       `tfsdk:"ssl_support_method"`
}

type functionAssociationModel struct {
	EventType   fwtypes.StringEnum[EventType] `tfsdk:"event_type"`
	FunctionARN types.String                  `tfsdk:"function_arn"`
}

type lambdaFunctionAssociationModel struct {
	EventType   fwtypes.StringEnum[EventType] `tfsdk:"event_type"`
	IncludeBody types.Bool                    `tfsdk:"include_body"`
	LambdaARN   types.String                  `tfsdk:"lambda_arn"`
}

type tenantConfigModel struct {
	ParameterDefinition fwtypes.ListNestedObjectValueOf[parameterDefinitionModel] `tfsdk:"parameter_definition"`
}

type parameterDefinitionModel struct {
	Name       types.String                                                    `tfsdk:"name"`
	Definition fwtypes.ListNestedObjectValueOf[parameterDefinitionSchemaModel] `tfsdk:"definition"`
}

type parameterDefinitionSchemaModel struct {
	StringSchema fwtypes.ListNestedObjectValueOf[stringSchemaConfigModel] `tfsdk:"string_schema"`
}

type stringSchemaConfigModel struct {
	Required     types.Bool   `tfsdk:"required"`
	Comment      types.String `tfsdk:"comment"`
	DefaultValue types.String `tfsdk:"default_value"`
}

func TestExpandRealWorldCloudFrontDistribution(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"basic string and bool fields": {
			Source: struct {
				Comment types.String `tfsdk:"comment"`
				Enabled types.Bool   `tfsdk:"enabled"`
			}{
				Comment: types.StringValue("Test distribution"),
				Enabled: types.BoolValue(true),
			},
			Target: &DistributionConfig{},
			WantTarget: &DistributionConfig{
				Comment: aws.String("Test distribution"),
				Enabled: aws.Bool(true),
			},
		},
		"enum field expansion": {
			Source: struct {
				Comment     types.String                    `tfsdk:"comment"`
				HttpVersion fwtypes.StringEnum[HttpVersion] `tfsdk:"http_version"`
			}{
				Comment:     types.StringValue("Distribution with enum"),
				HttpVersion: fwtypes.StringEnumValue(HttpVersionHttp2),
			},
			Target: &DistributionConfig{},
			WantTarget: &DistributionConfig{
				Comment:     aws.String("Distribution with enum"),
				HttpVersion: HttpVersionHttp2,
			},
		},
		"simple nested struct without XML wrapper": {
			Source: struct {
				Comment              types.String                             `tfsdk:"comment"`
				TargetOriginId       types.String                             `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy] `tfsdk:"viewer_protocol_policy"`
			}{
				Comment:              types.StringValue("Simple nested test"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				AllowedMethods:       nil, // XML wrapper should be nil when source doesn't provide it
				TrustedKeyGroups:     nil, // XML wrapper should be nil when source doesn't provide it
				TrustedSigners:       nil, // XML wrapper should be nil when source doesn't provide it
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
			},
		},
		"with simple allowed methods": {
			Source: struct {
				Comment              types.String                                   `tfsdk:"comment"`
				TargetOriginId       types.String                                   `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy]       `tfsdk:"viewer_protocol_policy"`
				AllowedMethods       fwtypes.SetValueOf[fwtypes.StringEnum[Method]] `tfsdk:"allowed_methods"`
			}{
				Comment:              types.StringValue("Test with allowed methods"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
				AllowedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[Method]](context.Background(), []attr.Value{
					fwtypes.StringEnumValue(MethodGet),
					fwtypes.StringEnumValue(MethodHead),
				}),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				AllowedMethods: &AllowedMethods{
					Items:    []Method{MethodGet, MethodHead},
					Quantity: aws.Int32(2),
				},
				TrustedKeyGroups:     nil, // Should remain nil
				TrustedSigners:       nil, // Should remain nil
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
			},
		},
		"with trusted key groups": {
			Source: struct {
				Comment              types.String                             `tfsdk:"comment"`
				TargetOriginId       types.String                             `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy] `tfsdk:"viewer_protocol_policy"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String]        `tfsdk:"trusted_key_groups"`
			}{
				Comment:              types.StringValue("Test with trusted key groups"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
				TrustedKeyGroups: fwtypes.NewListValueOfMust[types.String](context.Background(), []attr.Value{
					types.StringValue("key-group-1"),
					types.StringValue("key-group-2"),
				}),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				AllowedMethods: nil, // Should remain nil
				TrustedKeyGroups: &TrustedKeyGroups{
					Enabled:  nil, // TODO: Should this be set to true automatically?
					Items:    []string{"key-group-1", "key-group-2"},
					Quantity: aws.Int32(2),
				},
				TrustedSigners:       nil, // Should remain nil
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
			},
		},
		"with multiple XML wrappers": {
			Source: struct {
				Comment              types.String                                   `tfsdk:"comment"`
				TargetOriginId       types.String                                   `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy]       `tfsdk:"viewer_protocol_policy"`
				AllowedMethods       fwtypes.SetValueOf[fwtypes.StringEnum[Method]] `tfsdk:"allowed_methods"`
				TrustedSigners       fwtypes.ListValueOf[types.String]              `tfsdk:"trusted_signers"`
			}{
				Comment:              types.StringValue("Test with multiple XML wrappers"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
				AllowedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[Method]](context.Background(), []attr.Value{
					fwtypes.StringEnumValue(MethodGet),
					fwtypes.StringEnumValue(MethodPost),
					fwtypes.StringEnumValue(MethodPut),
				}),
				TrustedSigners: fwtypes.NewListValueOfMust[types.String](context.Background(), []attr.Value{
					types.StringValue("signer-1"),
				}),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				AllowedMethods: &AllowedMethods{
					Items:    []Method{MethodGet, MethodPost, MethodPut},
					Quantity: aws.Int32(3),
				},
				TrustedKeyGroups: nil, // Should remain nil
				TrustedSigners: &TrustedSigners{
					Enabled:  nil, // TODO: Should this be set to true automatically?
					Items:    []string{"signer-1"},
					Quantity: aws.Int32(1),
				},
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
			},
		},
		"with empty collections": {
			Source: struct {
				Comment              types.String                                   `tfsdk:"comment"`
				TargetOriginId       types.String                                   `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy]       `tfsdk:"viewer_protocol_policy"`
				AllowedMethods       fwtypes.SetValueOf[fwtypes.StringEnum[Method]] `tfsdk:"allowed_methods"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String]              `tfsdk:"trusted_key_groups"`
			}{
				Comment:              types.StringValue("Test with empty collections"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
				AllowedMethods:       fwtypes.NewSetValueOfMust[fwtypes.StringEnum[Method]](context.Background(), []attr.Value{}),
				TrustedKeyGroups:     fwtypes.NewListValueOfMust[types.String](context.Background(), []attr.Value{}),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				AllowedMethods: &AllowedMethods{
					Items:    []Method{}, // Empty slice
					Quantity: aws.Int32(0),
				},
				TrustedKeyGroups: &TrustedKeyGroups{
					Enabled:  nil,        // TODO: Should this be set to true automatically?
					Items:    []string{}, // Empty slice
					Quantity: aws.Int32(0),
				},
				TrustedSigners:       nil, // Should remain nil
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
			},
		},
		"with null collections": {
			Source: struct {
				Comment              types.String                                   `tfsdk:"comment"`
				TargetOriginId       types.String                                   `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy]       `tfsdk:"viewer_protocol_policy"`
				AllowedMethods       fwtypes.SetValueOf[fwtypes.StringEnum[Method]] `tfsdk:"allowed_methods"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String]              `tfsdk:"trusted_key_groups"`
			}{
				Comment:              types.StringValue("Test with null collections"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
				AllowedMethods:       fwtypes.NewSetValueOfNull[fwtypes.StringEnum[Method]](context.Background()),
				TrustedKeyGroups:     fwtypes.NewListValueOfNull[types.String](context.Background()),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				AllowedMethods:       nil, // Should be nil for null collections
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
			},
		},
		"with complex nested XML wrappers": {
			Source: struct {
				Comment              types.String                                   `tfsdk:"comment"`
				TargetOriginId       types.String                                   `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy]       `tfsdk:"viewer_protocol_policy"`
				AllowedMethods       fwtypes.SetValueOf[fwtypes.StringEnum[Method]] `tfsdk:"allowed_methods"`
				CachedMethods        fwtypes.SetValueOf[fwtypes.StringEnum[Method]] `tfsdk:"cached_methods"`
			}{
				Comment:              types.StringValue("Test with complex nested XML wrappers"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
				AllowedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[Method]](context.Background(), []attr.Value{
					fwtypes.StringEnumValue(MethodGet),
					fwtypes.StringEnumValue(MethodHead),
					fwtypes.StringEnumValue(MethodPost),
				}),
				CachedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[Method]](context.Background(), []attr.Value{
					fwtypes.StringEnumValue(MethodGet),
					fwtypes.StringEnumValue(MethodHead),
				}),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				AllowedMethods: &AllowedMethods{
					Items:    []Method{MethodGet, MethodHead, MethodPost},
					Quantity: aws.Int32(3),
					CachedMethods: &CachedMethods{
						Items:    []Method{MethodGet, MethodHead},
						Quantity: aws.Int32(2),
					},
				},
				TrustedKeyGroups:     nil,
				TrustedSigners:       nil,
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

func TestExpandCloudFrontTrustedSignersKeyGroupsBug(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"minimal distribution without trusted signers/key groups should have nil pointers": {
			Source: struct {
				Comment              types.String                             `tfsdk:"comment"`
				TargetOriginId       types.String                             `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy] `tfsdk:"viewer_protocol_policy"`
				// NO TrustedSigners or TrustedKeyGroups fields defined at all
			}{
				Comment:              types.StringValue("Minimal distribution without trusted signers/key groups"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
				// These should be nil, NOT empty structs with missing Enabled field
				TrustedSigners:   nil,
				TrustedKeyGroups: nil,
			},
		},
		"distribution with null trusted signers/key groups should have nil pointers": {
			Source: struct {
				Comment              types.String                             `tfsdk:"comment"`
				TargetOriginId       types.String                             `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy] `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String]        `tfsdk:"trusted_signers"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String]        `tfsdk:"trusted_key_groups"`
			}{
				Comment:              types.StringValue("Distribution with explicit null trusted signers/key groups"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
				// Explicitly null collections
				TrustedSigners:   fwtypes.NewListValueOfNull[types.String](ctx),
				TrustedKeyGroups: fwtypes.NewListValueOfNull[types.String](ctx),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
				// Null collections should result in nil XML wrappers, NOT structs with nil Enabled
				TrustedSigners:   nil,
				TrustedKeyGroups: nil,
			},
		},
		"distribution with empty trusted signers/key groups creates empty structs but should have enabled set": {
			Source: struct {
				Comment              types.String                             `tfsdk:"comment"`
				TargetOriginId       types.String                             `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy] `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String]        `tfsdk:"trusted_signers"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String]        `tfsdk:"trusted_key_groups"`
			}{
				Comment:              types.StringValue("Distribution with empty trusted signers/key groups"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
				// Empty but not null collections
				TrustedSigners:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
				TrustedKeyGroups: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
				// Empty collections create XML wrappers with Quantity: 0
				// TODO: AutoFlex should set Enabled: false for empty collections
				TrustedSigners: &TrustedSigners{
					Enabled:  nil, // BUG: This should be aws.Bool(false) to avoid AWS API validation error
					Items:    []string{},
					Quantity: aws.Int32(0),
				},
				TrustedKeyGroups: &TrustedKeyGroups{
					Enabled:  nil, // BUG: This should be aws.Bool(false) to avoid AWS API validation error
					Items:    []string{},
					Quantity: aws.Int32(0),
				},
			},
		},
		"distribution with populated trusted signers/key groups should have enabled set": {
			Source: struct {
				Comment              types.String                             `tfsdk:"comment"`
				TargetOriginId       types.String                             `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy fwtypes.StringEnum[ViewerProtocolPolicy] `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String]        `tfsdk:"trusted_signers"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String]        `tfsdk:"trusted_key_groups"`
			}{
				Comment:              types.StringValue("Distribution with populated trusted signers/key groups"),
				TargetOriginId:       types.StringValue("origin1"),
				ViewerProtocolPolicy: fwtypes.StringEnumValue(ViewerProtocolPolicyAllowAll),
				TrustedSigners: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("trusted-signer-1"),
				}),
				TrustedKeyGroups: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("key-group-1"),
				}),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				TargetOriginId:       aws.String("origin1"),
				ViewerProtocolPolicy: ViewerProtocolPolicyAllowAll,
				// Populated collections create XML wrappers
				// TODO: AutoFlex should set Enabled: true for non-empty collections
				TrustedSigners: &TrustedSigners{
					Enabled:  nil, // BUG: This should be aws.Bool(true) to avoid AWS API validation error
					Items:    []string{"trusted-signer-1"},
					Quantity: aws.Int32(1),
				},
				TrustedKeyGroups: &TrustedKeyGroups{
					Enabled:  nil, // BUG: This should be aws.Bool(true) to avoid AWS API validation error
					Items:    []string{"key-group-1"},
					Quantity: aws.Int32(1),
				},
			},
		},
		"exact reproduction of multitenant distribution bug": {
			Source: struct {
				TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups"`
			}{
				TargetOriginId:       types.StringValue("S3-my-bucket"),
				ViewerProtocolPolicy: types.StringValue("redirect-to-https"),
				// These are null/undefined in the Terraform config
				TrustedSigners:   fwtypes.NewListValueOfNull[types.String](ctx),
				TrustedKeyGroups: fwtypes.NewListValueOfNull[types.String](ctx),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				TargetOriginId:       aws.String("S3-my-bucket"),
				ViewerProtocolPolicy: "redirect-to-https",
				// When list fields are null, these should be nil, not structs with nil Enabled
				TrustedSigners:   nil,
				TrustedKeyGroups: nil,
			},
		},
		"exact reproduction with nested list structure": {
			Source: func() any {
				defaultCacheBehavior, _ := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []struct {
					TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
					ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
					TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers"`
					TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups"`
				}{
					{
						TargetOriginId:       types.StringValue("S3-my-bucket"),
						ViewerProtocolPolicy: types.StringValue("redirect-to-https"),
						// These are null/undefined in the Terraform config - this is the real bug scenario
						TrustedSigners:   fwtypes.NewListValueOfNull[types.String](ctx),
						TrustedKeyGroups: fwtypes.NewListValueOfNull[types.String](ctx),
					},
				})
				return struct {
					DefaultCacheBehavior fwtypes.ListNestedObjectValueOf[struct {
						TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
						ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
						TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers"`
						TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups"`
					}] `tfsdk:"default_cache_behavior"`
				}{
					DefaultCacheBehavior: defaultCacheBehavior,
				}
			}(),
			Target: &struct {
				DefaultCacheBehavior *DefaultCacheBehavior
			}{},
			WantTarget: &struct {
				DefaultCacheBehavior *DefaultCacheBehavior
			}{
				DefaultCacheBehavior: &DefaultCacheBehavior{
					TargetOriginId:       aws.String("S3-my-bucket"),
					ViewerProtocolPolicy: "redirect-to-https",
					// When nested list fields are null, these should be nil, not structs with nil Enabled
					TrustedSigners:   nil,
					TrustedKeyGroups: nil,
				},
			},
		},
		"exact reproduction with unknown values - THE REAL BUG": {
			Source: struct {
				TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups"`
			}{
				TargetOriginId:       types.StringValue("S3-my-bucket"),
				ViewerProtocolPolicy: types.StringValue("redirect-to-https"),
				// These are UNKNOWN, not null - this is the real bug scenario!
				TrustedSigners:   fwtypes.NewListValueOfUnknown[types.String](ctx),
				TrustedKeyGroups: fwtypes.NewListValueOfUnknown[types.String](ctx),
			},
			Target: &DefaultCacheBehavior{},
			WantTarget: &DefaultCacheBehavior{
				TargetOriginId:       aws.String("S3-my-bucket"),
				ViewerProtocolPolicy: "redirect-to-https",
				// When list fields are unknown, these should be nil, not structs with nil Enabled
				TrustedSigners:   nil,
				TrustedKeyGroups: nil,
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

func TestFlattenCloudFrontTrustedSignersKeyGroupsBug(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"flatten distribution with nil TrustedSigners/TrustedKeyGroups should result in null lists": {
			Source: &DefaultCacheBehavior{
				TargetOriginId:       aws.String("S3-my-bucket"),
				ViewerProtocolPolicy: "redirect-to-https",
				// These are nil pointers - should result in null Terraform lists
				TrustedSigners:   nil,
				TrustedKeyGroups: nil,
			},
			Target: &struct {
				TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers" autoflex:",wrapper=items"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups" autoflex:",wrapper=items"`
			}{},
			WantTarget: &struct {
				TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers" autoflex:",wrapper=items"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups" autoflex:",wrapper=items"`
			}{
				TargetOriginId:       types.StringValue("S3-my-bucket"),
				ViewerProtocolPolicy: types.StringValue("redirect-to-https"),
				// When source XML wrappers are nil, should result in null lists
				TrustedSigners:   fwtypes.NewListValueOfNull[types.String](ctx),
				TrustedKeyGroups: fwtypes.NewListValueOfNull[types.String](ctx),
			},
		},
		"flatten distribution with empty TrustedSigners/TrustedKeyGroups should result in empty lists": {
			Source: &DefaultCacheBehavior{
				TargetOriginId:       aws.String("S3-my-bucket"),
				ViewerProtocolPolicy: "redirect-to-https",
				// Empty XML wrapper structs with no items
				TrustedSigners: &TrustedSigners{
					Enabled:  aws.Bool(false),
					Items:    []string{},
					Quantity: aws.Int32(0),
				},
				TrustedKeyGroups: &TrustedKeyGroups{
					Enabled:  aws.Bool(false),
					Items:    []string{},
					Quantity: aws.Int32(0),
				},
			},
			Target: &struct {
				TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers" autoflex:",wrapper=items"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups" autoflex:",wrapper=items"`
			}{},
			WantTarget: &struct {
				TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers" autoflex:",wrapper=items"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups" autoflex:",wrapper=items"`
			}{
				TargetOriginId:       types.StringValue("S3-my-bucket"),
				ViewerProtocolPolicy: types.StringValue("redirect-to-https"),
				// Empty XML wrappers should result in empty lists
				TrustedSigners:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
				TrustedKeyGroups: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
			},
		},
		"flatten distribution with populated TrustedSigners/TrustedKeyGroups should result in populated lists": {
			Source: &DefaultCacheBehavior{
				TargetOriginId:       aws.String("S3-my-bucket"),
				ViewerProtocolPolicy: "redirect-to-https",
				// Populated XML wrapper structs
				TrustedSigners: &TrustedSigners{
					Enabled:  aws.Bool(true),
					Items:    []string{"trusted-signer-1", "trusted-signer-2"},
					Quantity: aws.Int32(2),
				},
				TrustedKeyGroups: &TrustedKeyGroups{
					Enabled:  aws.Bool(true),
					Items:    []string{"key-group-1"},
					Quantity: aws.Int32(1),
				},
			},
			Target: &struct {
				TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers" autoflex:",wrapper=items"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups" autoflex:",wrapper=items"`
			}{},
			WantTarget: &struct {
				TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
				ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
				TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers" autoflex:",wrapper=items"`
				TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups" autoflex:",wrapper=items"`
			}{
				TargetOriginId:       types.StringValue("S3-my-bucket"),
				ViewerProtocolPolicy: types.StringValue("redirect-to-https"),
				// Populated XML wrappers should result in populated lists
				TrustedSigners: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("trusted-signer-1"),
					types.StringValue("trusted-signer-2"),
				}),
				TrustedKeyGroups: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("key-group-1"),
				}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: false})
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

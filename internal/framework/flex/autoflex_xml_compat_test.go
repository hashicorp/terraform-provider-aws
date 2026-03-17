// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten of AWS API XML wrappers (Items/Quantity).

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
	Items    []FunctionAssociation `json:"Items"`
	Quantity *int32                `json:"Quantity"`
}

// Terraform model types
type FunctionAssociationTF struct {
	EventType   types.String `tfsdk:"event_type"`
	FunctionARN types.String `tfsdk:"function_arn"`
}

type DistributionConfigTF struct {
	FunctionAssociations fwtypes.SetNestedObjectValueOf[FunctionAssociationTF] `tfsdk:"function_associations" autoflex:",xmlwrapper=Items"`
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

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

// Test XML wrapper expansion for direct struct (not pointer to struct)
type DirectXMLWrapper struct {
	Items    []string
	Quantity *int32
}

type DirectWrapperTF struct {
	Items fwtypes.SetValueOf[types.String] `tfsdk:"items" autoflex:",xmlwrapper=items"`
}

type DirectWrapperAWS struct {
	Items DirectXMLWrapper
}

func TestPotentialXMLWrapperStruct(t *testing.T) {
	t.Parallel()

	type embedWithField struct {
		Count int64
	}
	type embedWithoutField struct{}

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
			name: "valid XML wrapper with anonymous struct",
			input: struct {
				Items    []string
				Quantity *int32
			}{},
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
		{
			name: "struct with extra field",
			input: struct {
				Items    []string
				Quantity *int32
				Name     string
			}{},
			expected: true,
		},
		{
			name: "struct with anonymous embedWithField",
			input: struct {
				Items    []string
				Quantity *int32
				embedWithField
			}{},
			expected: true,
		},
		{
			name: "struct with anonymous embedWithoutField",
			input: struct {
				Items    []string
				Quantity *int32
				embedWithoutField
			}{},
			expected: true,
		},
		{
			name: "struct with private embedWithField",
			input: struct {
				Items    []string
				Quantity *int32
				private  embedWithField
			}{},
			expected: true,
		},
		{
			name: "struct with private embedWithoutField",
			input: struct {
				Items    []string
				Quantity *int32
				private  embedWithoutField
			}{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := potentialXMLWrapperStruct(reflect.TypeOf(tc.input))
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

// TF model types with wrapper tags (for flattening - AWS to TF)
type tfStatusCodesModelForFlatten struct {
	StatusCodes fwtypes.SetValueOf[types.Int64] `tfsdk:"status_codes" autoflex:",xmlwrapper=Items"`
}

type tfHeadersModelForFlatten struct {
	Headers fwtypes.ListValueOf[types.String] `tfsdk:"headers" autoflex:",xmlwrapper=Items"`
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
			Source: FunctionAssociations{
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
			Target: &DistributionConfigTF{},
			WantTarget: &DistributionConfigTF{
				FunctionAssociations: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*FunctionAssociationTF{
					{
						EventType:   types.StringValue("viewer-request"),
						FunctionARN: types.StringValue("arn:aws:cloudfront::123456789012:function/example-function"),
					},
					{
						EventType:   types.StringValue("viewer-response"),
						FunctionARN: types.StringValue("arn:aws:cloudfront::123456789012:function/another-function"),
					},
				}),
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

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

type FunctionAssociationsTF struct {
	Items    fwtypes.ListNestedObjectValueOf[FunctionAssociationTF] `tfsdk:"items"`
	Quantity types.Int64                                            `tfsdk:"quantity"`
}

type DistributionConfigTFNoXMLWrapper struct {
	FunctionAssociations fwtypes.ListNestedObjectValueOf[FunctionAssociationsTF] `tfsdk:"function_associations"`
}

func TestExpandNoXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid function associations": {
			Source: DistributionConfigTFNoXMLWrapper{
				FunctionAssociations: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &FunctionAssociationsTF{
					Items: fwtypes.NewListNestedObjectValueOfSliceMust(
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
					Quantity: types.Int64Value(2),
				}),
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
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestFlattenNoXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"complex type - function associations": {
			Source: DistributionConfigAWS{
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
			// 	Items: []FunctionAssociation{
			// 		{
			// 			EventType:   "viewer-request",
			// 			FunctionARN: aws.String("arn:aws:cloudfront::123456789012:function/example-function"),
			// 		},
			// 		{
			// 			EventType:   "viewer-response",
			// 			FunctionARN: aws.String("arn:aws:cloudfront::123456789012:function/another-function"),
			// 		},
			// 	},
			// 	Quantity: aws.Int32(2),
			// },
			Target: &DistributionConfigTFNoXMLWrapper{},
			WantTarget: &DistributionConfigTFNoXMLWrapper{
				FunctionAssociations: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &FunctionAssociationsTF{
					Items: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*FunctionAssociationTF{
						{
							EventType:   types.StringValue("viewer-request"),
							FunctionARN: types.StringValue("arn:aws:cloudfront::123456789012:function/test-function-1"),
						},
						{
							EventType:   types.StringValue("viewer-response"),
							FunctionARN: types.StringValue("arn:aws:cloudfront::123456789012:function/test-function-2"),
						},
					}),
					Quantity: types.Int64Value(2),
				}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

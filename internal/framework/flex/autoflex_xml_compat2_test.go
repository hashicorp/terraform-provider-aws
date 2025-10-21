// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// XML Wrapper Compatibility Tests
//
// These tests validate AutoFlex handling of AWS XML wrapper patterns commonly found in CloudFront APIs.
// XML wrappers are 3-field structs that encapsulate collections for XML serialization.
//
// ## Rule 1: Wrapper contains only Items/Quantity
//   AWS: {Items: []T, Quantity: *int32}
//   
//   Scalar elements ([]string, []enum):
//     Terraform: SetAttribute or ListAttribute
//     Example: origin_ssl_protocols = ["TLSv1.2", "TLSv1.3"]
//   
//   Struct elements ([]CustomStruct):
//     Terraform: Repeatable singular blocks (one block per item)
//     Example: lambda_function_association { ... }
//             lambda_function_association { ... }
//
// ## Rule 2: Wrapper contains Items/Quantity + additional fields (e.g., Enabled)
//   AWS: {Items: []T, Quantity: *int32, Enabled: *bool}
//   
//   Terraform: Single plural block containing collection + extra fields
//   Example: trusted_signers {
//              items   = ["signer1", "signer2"]
//              enabled = true  # computed from len(items) > 0
//            }
//
// ## Key Behaviors:
// - null/unknown Terraform input → nil AWS pointer (no empty struct creation)
// - empty collection → {Items: [], Quantity: 0, ...}
// - Quantity is always derived from len(Items), never user-specified
// - Expand and flatten must be symmetrical for stable apply/refresh cycles

package flex

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// Test data types for Rule 1 XML wrappers
type testXMLWrapperScalar struct {
	Items    []string `autoflex:",wrapper=items"`
	Quantity *int32
}

// Test int32 slice items (like StatusCodes)
type testXMLWrapperInt32 struct {
	Items    []int32 `autoflex:",wrapper=items"`
	Quantity *int32
}

type testXMLWrapperStruct struct {
	Items    []testStructItem `autoflex:",wrapper=items"`
	Quantity *int32
}

type testStructItem struct {
	Name  *string
	Value *int32
}

type testStructItemModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.Int32  `tfsdk:"value"`
}

// Test Rule 1: XML wrappers with Items/Quantity only (scalar elements)
func TestExpandXMLWrapperRule1ScalarElements(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"OriginSslProtocols": {
			"null set": {
				Source:     fwtypes.NewSetValueOfNull[fwtypes.StringEnum[awstypes.SslProtocol]](ctx),
				Target:     &awstypes.OriginSslProtocols{},
				WantTarget: (*awstypes.OriginSslProtocols)(nil),
			},
			"single protocol": {
				Source: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.SslProtocol]](ctx, []attr.Value{
					fwtypes.StringEnumValue(awstypes.SslProtocolTLSv12),
				}),
				Target:     &awstypes.OriginSslProtocols{},
				WantTarget: &awstypes.OriginSslProtocols{Items: []awstypes.SslProtocol{awstypes.SslProtocolTLSv12}, Quantity: aws.Int32(1)},
			},
		},
		"TestXMLWrapperScalar": {
			"null set": {
				Source:     fwtypes.NewSetValueOfNull[types.String](ctx),
				Target:     &testXMLWrapperScalar{},
				WantTarget: &testXMLWrapperScalar{}, // AutoFlex currently creates empty struct, not nil
			},
			"empty set": {
				Source:     fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{}),
				Target:     &testXMLWrapperScalar{},
				WantTarget: &testXMLWrapperScalar{Items: []string{}, Quantity: aws.Int32(0)},
			},
			"single item": {
				Source: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("item1"),
				}),
				Target:     &testXMLWrapperScalar{},
				WantTarget: &testXMLWrapperScalar{Items: []string{"item1"}, Quantity: aws.Int32(1)},
			},
			"multiple items": {
				Source: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("item1"),
					types.StringValue("item2"),
				}),
				Target:     &testXMLWrapperScalar{},
				WantTarget: &testXMLWrapperScalar{Items: []string{"item1", "item2"}, Quantity: aws.Int32(2)},
			},
		},
		"TestXMLWrapperInt32": {
			"empty set": {
				Source:     fwtypes.NewSetValueOfMust[types.Int32](ctx, []attr.Value{}),
				Target:     &testXMLWrapperInt32{},
				WantTarget: &testXMLWrapperInt32{Items: []int32{}, Quantity: aws.Int32(0)},
			},
			"single item": {
				Source:     fwtypes.NewSetValueOfMust[types.Int32](ctx, []attr.Value{types.Int32Value(404)}),
				Target:     &testXMLWrapperInt32{},
				WantTarget: &testXMLWrapperInt32{Items: []int32{404}, Quantity: aws.Int32(1)},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			if testName == "OriginSslProtocols" {
				runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: false})
			} else {
				runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true})
			}
		})
	}
}

// Test Rule 1: XML wrappers with Items/Quantity only (struct elements)
func TestExpandXMLWrapperRule1StructElements(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"TestXMLWrapperStruct": {
			"null set": {
				Source:     fwtypes.NewSetNestedObjectValueOfNull[testStructItemModel](ctx),
				Target:     &testXMLWrapperStruct{},
				WantTarget: &testXMLWrapperStruct{}, // AutoFlex currently creates empty struct, not nil
			},
			"empty set": {
				Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust[testStructItemModel](ctx, []testStructItemModel{}),
				Target:     &testXMLWrapperStruct{},
				WantTarget: &testXMLWrapperStruct{Items: []testStructItem{}, Quantity: aws.Int32(0)},
			},
			"single item": {
				Source: fwtypes.NewSetNestedObjectValueOfValueSliceMust[testStructItemModel](ctx, []testStructItemModel{
					{
						Name:  types.StringValue("test"),
						Value: types.Int32Value(42),
					},
				}),
				Target: &testXMLWrapperStruct{},
				WantTarget: &testXMLWrapperStruct{
					Items: []testStructItem{
						{
							Name:  aws.String("test"),
							Value: aws.Int32(42),
						},
					},
					Quantity: aws.Int32(1),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true})
		})
	}
}

// Test Rule 1 Flatten: XML wrappers with Items/Quantity only (scalar elements)
func TestFlattenXMLWrapperRule1ScalarElements(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Source struct containing XML wrapper
	type sourceStruct struct {
		XMLWrapper *testXMLWrapperScalar
	}

	// Target struct containing collection
	type targetStruct struct {
		XMLWrapper fwtypes.SetValueOf[types.String] `autoflex:",wrapper=items"`
	}

	testCases := map[string]autoFlexTestCases{
		"TestXMLWrapperScalar": {
			"nil source": {
				Source:     &sourceStruct{XMLWrapper: nil},
				Target:     &targetStruct{},
				WantTarget: &targetStruct{XMLWrapper: fwtypes.NewSetValueOfNull[types.String](ctx)},
			},
			"empty items": {
				Source:     &sourceStruct{XMLWrapper: &testXMLWrapperScalar{Items: []string{}, Quantity: aws.Int32(0)}},
				Target:     &targetStruct{},
				WantTarget: &targetStruct{XMLWrapper: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{})},
			},
			"single item": {
				Source:     &sourceStruct{XMLWrapper: &testXMLWrapperScalar{Items: []string{"item1"}, Quantity: aws.Int32(1)}},
				Target:     &targetStruct{},
				WantTarget: &targetStruct{XMLWrapper: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{types.StringValue("item1")})},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true})
		})
	}
}

// Test Rule 1 Flatten: XML wrappers with Items/Quantity only (struct elements)  
func TestFlattenXMLWrapperRule1StructElements(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Source struct containing XML wrapper
	type sourceStruct struct {
		XMLWrapper *testXMLWrapperStruct
	}

	// Target struct containing collection
	type targetStruct struct {
		XMLWrapper fwtypes.SetNestedObjectValueOf[testStructItemModel] `autoflex:",wrapper=items"`
	}

	testCases := map[string]autoFlexTestCases{
		"TestXMLWrapperStruct": {
			"nil source": {
				Source:     &sourceStruct{XMLWrapper: nil},
				Target:     &targetStruct{},
				WantTarget: &targetStruct{XMLWrapper: fwtypes.NewSetNestedObjectValueOfNull[testStructItemModel](ctx)},
			},
			"empty items": {
				Source:     &sourceStruct{XMLWrapper: &testXMLWrapperStruct{Items: []testStructItem{}, Quantity: aws.Int32(0)}},
				Target:     &targetStruct{},
				WantTarget: &targetStruct{XMLWrapper: fwtypes.NewSetNestedObjectValueOfValueSliceMust[testStructItemModel](ctx, []testStructItemModel{})},
			},
			"single item": {
				Source: &sourceStruct{
					XMLWrapper: &testXMLWrapperStruct{
						Items: []testStructItem{
							{Name: aws.String("test"), Value: aws.Int32(42)},
						},
						Quantity: aws.Int32(1),
					},
				},
				Target: &targetStruct{},
				WantTarget: &targetStruct{
					XMLWrapper: fwtypes.NewSetNestedObjectValueOfValueSliceMust[testStructItemModel](ctx, []testStructItemModel{
						{Name: types.StringValue("test"), Value: types.Int32Value(42)},
					}),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true})
		})
	}
}
// Test Rule 1 Symmetry: Verify expand→flatten produces identical results
func TestXMLWrapperRule1Symmetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test scalar elements symmetry
	t.Run("ScalarElements", func(t *testing.T) {
		// Original Terraform value
		original := fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
			types.StringValue("item1"),
			types.StringValue("item2"),
		})

		// Expand: TF → AWS
		awsStruct := &testXMLWrapperScalar{}
		expandDiags := Expand(ctx, original, awsStruct)
		if expandDiags.HasError() {
			t.Fatalf("Expand failed: %v", expandDiags)
		}

		// Verify expand result
		if len(awsStruct.Items) != 2 || *awsStruct.Quantity != 2 {
			t.Errorf("Expand produced incorrect result: Items=%v, Quantity=%v", awsStruct.Items, *awsStruct.Quantity)
		}

		// Flatten: AWS → TF (using struct-to-struct pattern)
		sourceStruct := &struct{ XMLWrapper *testXMLWrapperScalar }{XMLWrapper: awsStruct}
		targetStruct := &struct{ XMLWrapper fwtypes.SetValueOf[types.String] `autoflex:",wrapper=items"` }{}
		
		flattenDiags := Flatten(ctx, sourceStruct, targetStruct)
		if flattenDiags.HasError() {
			t.Fatalf("Flatten failed: %v", flattenDiags)
		}

		// Verify symmetry: flattened result should match original
		if !original.Equal(targetStruct.XMLWrapper) {
			t.Errorf("Symmetry broken: original=%v, flattened=%v", original, targetStruct.XMLWrapper)
		}
	})

	// Test struct elements symmetry  
	t.Run("StructElements", func(t *testing.T) {
		// Original Terraform value
		original := fwtypes.NewSetNestedObjectValueOfValueSliceMust[testStructItemModel](ctx, []testStructItemModel{
			{Name: types.StringValue("test"), Value: types.Int32Value(42)},
		})

		// Expand: TF → AWS
		awsStruct := &testXMLWrapperStruct{}
		expandDiags := Expand(ctx, original, awsStruct)
		if expandDiags.HasError() {
			t.Fatalf("Expand failed: %v", expandDiags)
		}

		// Verify expand result
		if len(awsStruct.Items) != 1 || *awsStruct.Quantity != 1 {
			t.Errorf("Expand produced incorrect result: Items=%v, Quantity=%v", awsStruct.Items, *awsStruct.Quantity)
		}

		// Flatten: AWS → TF (using struct-to-struct pattern)
		sourceStruct := &struct{ XMLWrapper *testXMLWrapperStruct }{XMLWrapper: awsStruct}
		targetStruct := &struct{ XMLWrapper fwtypes.SetNestedObjectValueOf[testStructItemModel] `autoflex:",wrapper=items"` }{}
		
		flattenDiags := Flatten(ctx, sourceStruct, targetStruct)
		if flattenDiags.HasError() {
			t.Fatalf("Flatten failed: %v", flattenDiags)
		}

		// Verify symmetry: flattened result should match original
		if !original.Equal(targetStruct.XMLWrapper) {
			t.Errorf("Symmetry broken: original=%v, flattened=%v", original, targetStruct.XMLWrapper)
		}
	})
}
// Test data types for Rule 2 XML wrappers (Items/Quantity + additional fields)
type testXMLWrapperRule2 struct {
	Items    []string
	Quantity *int32
	Enabled  *bool
}

// Test different field ordering (matches real TrustedSigners/TrustedKeyGroups)
type testXMLWrapperRule2DifferentOrder struct {
	Enabled  *bool
	Quantity *int32
	Items    []string
}

type testRule2Model struct {
	Items   fwtypes.ListValueOf[types.String] `tfsdk:"items"`
	Enabled types.Bool                        `tfsdk:"enabled"`
}

// Test Rule 2: XML wrappers with Items/Quantity + additional fields (e.g., Enabled)
func TestExpandXMLWrapperRule2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"TestXMLWrapperRule2": {
			"null block": {
				Source:     fwtypes.NewListNestedObjectValueOfNull[testRule2Model](ctx),
				Target:     &testXMLWrapperRule2{},
				WantTarget: &testXMLWrapperRule2{}, // AutoFlex currently creates empty struct, not nil
			},
			"empty items, enabled false": {
				Source: fwtypes.NewListNestedObjectValueOfValueSliceMust[testRule2Model](ctx, []testRule2Model{
					{
						Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
						Enabled: types.BoolValue(false),
					},
				}),
				Target:     &testXMLWrapperRule2{},
				WantTarget: &testXMLWrapperRule2{Items: []string{}, Quantity: aws.Int32(0), Enabled: aws.Bool(false)},
			},
			"with items, enabled true": {
				Source: fwtypes.NewListNestedObjectValueOfValueSliceMust[testRule2Model](ctx, []testRule2Model{
					{
						Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("item1")}),
						Enabled: types.BoolValue(true),
					},
				}),
				Target:     &testXMLWrapperRule2{},
				WantTarget: &testXMLWrapperRule2{Items: []string{"item1"}, Quantity: aws.Int32(1), Enabled: aws.Bool(true)},
			},
		},
		"TestXMLWrapperRule2DifferentOrder": {
			"different field order": {
				Source: fwtypes.NewListNestedObjectValueOfValueSliceMust[testRule2Model](ctx, []testRule2Model{
					{
						Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("item1")}),
						Enabled: types.BoolValue(true),
					},
				}),
				Target:     &testXMLWrapperRule2DifferentOrder{},
				WantTarget: &testXMLWrapperRule2DifferentOrder{Items: []string{"item1"}, Quantity: aws.Int32(1), Enabled: aws.Bool(true)},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true})
		})
	}
}
// Test Rule 2 Flatten: XML wrappers with Items/Quantity + additional fields
func TestFlattenXMLWrapperRule2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Source struct containing Rule 2 XML wrapper
	type sourceStruct struct {
		XMLWrapper *testXMLWrapperRule2
	}

	// Target struct containing Rule 2 pattern
	type targetStruct struct {
		XMLWrapper fwtypes.ListNestedObjectValueOf[testRule2Model] `autoflex:",wrapper=items"`
	}

	testCases := map[string]autoFlexTestCases{
		"TestXMLWrapperRule2": {
			"nil source": {
				Source:     &sourceStruct{XMLWrapper: nil},
				Target:     &targetStruct{},
				WantTarget: &targetStruct{XMLWrapper: fwtypes.NewListNestedObjectValueOfNull[testRule2Model](ctx)},
			},
			"empty items, enabled false": {
				Source: &sourceStruct{
					XMLWrapper: &testXMLWrapperRule2{Items: []string{}, Quantity: aws.Int32(0), Enabled: aws.Bool(false)},
				},
				Target: &targetStruct{},
				WantTarget: &targetStruct{
					XMLWrapper: fwtypes.NewListNestedObjectValueOfValueSliceMust[testRule2Model](ctx, []testRule2Model{
						{
							Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
							Enabled: types.BoolValue(false),
						},
					}),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true})
		})
	}
}

// Test Rule 2 Symmetry: Verify expand→flatten produces identical results
func TestXMLWrapperRule2Symmetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Rule2Pattern", func(t *testing.T) {
		// Original Terraform value (Rule 2 pattern)
		original := fwtypes.NewListNestedObjectValueOfValueSliceMust[testRule2Model](ctx, []testRule2Model{
			{
				Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("signer1"), types.StringValue("signer2")}),
				Enabled: types.BoolValue(true),
			},
		})

		// Expand: TF → AWS (this works)
		awsStruct := &testXMLWrapperRule2{}
		expandDiags := Expand(ctx, original, awsStruct)
		if expandDiags.HasError() {
			t.Fatalf("Expand failed: %v", expandDiags)
		}

		// Verify expand result
		if len(awsStruct.Items) != 2 || *awsStruct.Quantity != 2 || !*awsStruct.Enabled {
			t.Errorf("Expand produced incorrect result: Items=%v, Quantity=%v, Enabled=%v", 
				awsStruct.Items, *awsStruct.Quantity, *awsStruct.Enabled)
		}

		// Flatten: AWS → TF (this should now work)
		sourceStruct := &struct{ XMLWrapper *testXMLWrapperRule2 }{XMLWrapper: awsStruct}
		targetStruct := &struct{ XMLWrapper fwtypes.ListNestedObjectValueOf[testRule2Model] `autoflex:",wrapper=items"` }{}
		
		flattenDiags := Flatten(ctx, sourceStruct, targetStruct)
		if flattenDiags.HasError() {
			t.Fatalf("Flatten failed: %v", flattenDiags)
		}

		// Verify symmetry: flattened result should match original
		if !original.Equal(targetStruct.XMLWrapper) {
			t.Errorf("Symmetry broken: original=%v, flattened=%v", original, targetStruct.XMLWrapper)
		}
	})
}

// Test Real CloudFront Types: Validate actual AWS SDK types
func TestXMLWrapperRealCloudFrontTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test with actual CloudFront TrustedSigners type
	t.Run("TrustedSigners", func(t *testing.T) {
		// Terraform model matching CloudFront schema
		type trustedSignersModel struct {
			Items   fwtypes.ListValueOf[types.String] `tfsdk:"items"`
			Enabled types.Bool                        `tfsdk:"enabled"`
		}

		source := fwtypes.NewListNestedObjectValueOfValueSliceMust[trustedSignersModel](ctx, []trustedSignersModel{
			{
				Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("AKIAIOSFODNN7EXAMPLE")}),
				Enabled: types.BoolValue(true),
			},
		})

		target := &awstypes.TrustedSigners{}
		diags := Expand(ctx, source, target)

		if diags.HasError() {
			t.Errorf("TrustedSigners expand failed: %v", diags)
		} else {
			// Verify correct AWS structure
			if len(target.Items) != 1 || *target.Quantity != 1 || !*target.Enabled {
				t.Errorf("TrustedSigners incorrect: Items=%v, Quantity=%v, Enabled=%v", 
					target.Items, *target.Quantity, *target.Enabled)
			}
		}
	})
}

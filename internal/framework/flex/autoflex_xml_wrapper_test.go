// Copyright IBM Corp. 2014, 2026
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
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// Test data types for Rule 1 XML wrappers
type testXMLWrapperScalar struct {
	Items    []string
	Quantity *int32
}

// Test int32 slice items (like StatusCodes)
type testXMLWrapperInt32 struct {
	Items    []int32
	Quantity *int32
}

// Test int64 slice items
type testXMLWrapperInt64 struct {
	Items    []int64
	Quantity *int32
}

type testXMLWrapperStruct struct {
	Items    []testStructItem
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
		"TestXMLWrapperScalar": func() autoFlexTestCases {
			type tfModel struct {
				Field fwtypes.SetValueOf[types.String] `tfsdk:"field" autoflex:",xmlwrapper=Items,omitempty"`
			}
			type awsModel struct {
				Field *testXMLWrapperScalar
			}
			return autoFlexTestCases{
				"null set": {
					Source:     &tfModel{Field: fwtypes.NewSetValueOfNull[types.String](ctx)},
					Target:     &awsModel{},
					WantTarget: &awsModel{Field: nil},
				},
				"empty set": {
					Source:     &tfModel{Field: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{})},
					Target:     &awsModel{},
					WantTarget: &awsModel{Field: &testXMLWrapperScalar{Items: []string{}, Quantity: aws.Int32(0)}},
				},
				"single item": {
					Source: &tfModel{Field: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
						types.StringValue("item1"),
					})},
					Target:     &awsModel{},
					WantTarget: &awsModel{Field: &testXMLWrapperScalar{Items: []string{"item1"}, Quantity: aws.Int32(1)}},
				},
				"multiple items": {
					Source: &tfModel{Field: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
						types.StringValue("item1"),
						types.StringValue("item2"),
					})},
					Target:     &awsModel{},
					WantTarget: &awsModel{Field: &testXMLWrapperScalar{Items: []string{"item1", "item2"}, Quantity: aws.Int32(2)}},
				},
			}
		}(),
		"TestXMLWrapperInt32": func() autoFlexTestCases {
			type tfModel struct {
				Field fwtypes.SetValueOf[types.Int32] `tfsdk:"field" autoflex:",xmlwrapper=Items"`
			}
			type awsModel struct {
				Field *testXMLWrapperInt32
			}
			return autoFlexTestCases{
				"empty set": {
					Source:     &tfModel{Field: fwtypes.NewSetValueOfMust[types.Int32](ctx, []attr.Value{})},
					Target:     &awsModel{},
					WantTarget: &awsModel{Field: &testXMLWrapperInt32{Items: []int32{}, Quantity: aws.Int32(0)}},
				},
				"single item": {
					Source:     &tfModel{Field: fwtypes.NewSetValueOfMust[types.Int32](ctx, []attr.Value{types.Int32Value(404)})},
					Target:     &awsModel{},
					WantTarget: &awsModel{Field: &testXMLWrapperInt32{Items: []int32{404}, Quantity: aws.Int32(1)}},
				},
			}
		}(),
		"TestXMLWrapperInt64": func() autoFlexTestCases {
			type tfModel struct {
				Field fwtypes.SetValueOf[types.Int64] `tfsdk:"field" autoflex:",xmlwrapper=Items"`
			}
			type awsModel struct {
				Field *testXMLWrapperInt64
			}
			return autoFlexTestCases{
				"empty set": {
					Source:     &tfModel{Field: fwtypes.NewSetValueOfMust[types.Int64](ctx, []attr.Value{})},
					Target:     &awsModel{},
					WantTarget: &awsModel{Field: &testXMLWrapperInt64{Items: []int64{}, Quantity: aws.Int32(0)}},
				},
				"single item": {
					Source:     &tfModel{Field: fwtypes.NewSetValueOfMust[types.Int64](ctx, []attr.Value{types.Int64Value(12345)})},
					Target:     &awsModel{},
					WantTarget: &awsModel{Field: &testXMLWrapperInt64{Items: []int64{12345}, Quantity: aws.Int32(1)}},
				},
			}
		}(),
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			// Cannot generate golden logs: Log file names don't take multiple levels of tests into account
			// e.g. "TestXMLWrapperInt32/empty set" and "TestXMLWrapperInt64/empty set" collide
			if testName == "OriginSslProtocols" {
				runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: false, SkipGoldenLogs: true})
			} else {
				runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
			}
		})
	}
}

// Test Rule 1: XML wrappers with Items/Quantity only (struct elements)
func TestExpandXMLWrapperRule1StructElements(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type tfModel struct {
		Field fwtypes.SetNestedObjectValueOf[testStructItemModel] `tfsdk:"field" autoflex:",xmlwrapper=Items,omitempty"`
	}
	type awsModel struct {
		Field *testXMLWrapperStruct
	}

	testCases := map[string]autoFlexTestCases{
		"TestXMLWrapperStruct": {
			"null set": {
				Source:     &tfModel{Field: fwtypes.NewSetNestedObjectValueOfNull[testStructItemModel](ctx)},
				Target:     &awsModel{},
				WantTarget: &awsModel{Field: nil},
			},
			"empty set": {
				Source:     &tfModel{Field: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testStructItemModel{})},
				Target:     &awsModel{},
				WantTarget: &awsModel{Field: &testXMLWrapperStruct{Items: []testStructItem{}, Quantity: aws.Int32(0)}},
			},
			"single item": {
				Source: &tfModel{Field: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testStructItemModel{
					{
						Name:  types.StringValue("test"),
						Value: types.Int32Value(42),
					},
				})},
				Target: &awsModel{},
				WantTarget: &awsModel{Field: &testXMLWrapperStruct{
					Items: []testStructItem{
						{
							Name:  aws.String("test"),
							Value: aws.Int32(42),
						},
					},
					Quantity: aws.Int32(1),
				}},
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
		XMLWrapper fwtypes.SetValueOf[types.String] `autoflex:",xmlwrapper=Items"`
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

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
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
		XMLWrapper fwtypes.SetNestedObjectValueOf[testStructItemModel] `autoflex:",xmlwrapper=Items"`
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
				WantTarget: &targetStruct{XMLWrapper: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testStructItemModel{})},
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
					XMLWrapper: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testStructItemModel{
						{Name: types.StringValue("test"), Value: types.Int32Value(42)},
					}),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

// Test Rule 1 Symmetry: Verify expand→flatten produces identical results
func TestXMLWrapperRule1Symmetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test scalar elements symmetry
	t.Run("ScalarElements", func(t *testing.T) {
		t.Parallel()

		type tfSource struct {
			XMLWrapper fwtypes.SetValueOf[types.String] `tfsdk:"xml_wrapper" autoflex:",xmlwrapper=Items"`
		}

		type awsTarget struct {
			XMLWrapper *testXMLWrapperScalar
		}

		// Original Terraform value
		original := tfSource{
			XMLWrapper: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
				types.StringValue("item1"),
				types.StringValue("item2"),
			}),
		}

		// Expand: TF → AWS
		var awsStruct awsTarget
		expandDiags := Expand(ctx, original, &awsStruct)
		if expandDiags.HasError() {
			t.Fatalf("Expand failed: %v", expandDiags)
		}

		// Verify expand result
		if awsStruct.XMLWrapper == nil {
			t.Fatal("Expected non-nil XMLWrapper")
		}
		if len(awsStruct.XMLWrapper.Items) != 2 || *awsStruct.XMLWrapper.Quantity != 2 {
			t.Errorf("Expand produced incorrect result: Items=%v, Quantity=%v", awsStruct.XMLWrapper.Items, *awsStruct.XMLWrapper.Quantity)
		}

		// Flatten: AWS → TF
		var targetStruct tfSource
		flattenDiags := Flatten(ctx, &awsStruct, &targetStruct)
		if flattenDiags.HasError() {
			t.Fatalf("Flatten failed: %v", flattenDiags)
		}

		// Verify symmetry: flattened result should match original
		if !original.XMLWrapper.Equal(targetStruct.XMLWrapper) {
			t.Errorf("Symmetry broken: original=%v, flattened=%v", original.XMLWrapper, targetStruct.XMLWrapper)
		}
	})

	// Test struct elements symmetry
	t.Run("StructElements", func(t *testing.T) {
		t.Parallel()

		type tfSource struct {
			XMLWrapper fwtypes.SetNestedObjectValueOf[testStructItemModel] `tfsdk:"xml_wrapper" autoflex:",xmlwrapper=Items"`
		}

		type awsTarget struct {
			XMLWrapper *testXMLWrapperStruct
		}

		// Original Terraform value
		original := tfSource{
			XMLWrapper: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testStructItemModel{
				{Name: types.StringValue("test"), Value: types.Int32Value(42)},
			}),
		}

		// Expand: TF → AWS
		var awsStruct awsTarget
		expandDiags := Expand(ctx, original, &awsStruct)
		if expandDiags.HasError() {
			t.Fatalf("Expand failed: %v", expandDiags)
		}

		// Verify expand result
		if awsStruct.XMLWrapper == nil {
			t.Fatal("Expected non-nil XMLWrapper")
		}
		if len(awsStruct.XMLWrapper.Items) != 1 || *awsStruct.XMLWrapper.Quantity != 1 {
			t.Errorf("Expand produced incorrect result: Items=%v, Quantity=%v", awsStruct.XMLWrapper.Items, *awsStruct.XMLWrapper.Quantity)
		}

		// Flatten: AWS → TF
		var targetStruct tfSource
		flattenDiags := Flatten(ctx, &awsStruct, &targetStruct)
		if flattenDiags.HasError() {
			t.Fatalf("Flatten failed: %v", flattenDiags)
		}

		// Verify symmetry: flattened result should match original
		if !original.XMLWrapper.Equal(targetStruct.XMLWrapper) {
			t.Errorf("Symmetry broken: original=%v, flattened=%v", original.XMLWrapper, targetStruct.XMLWrapper)
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
	Items   fwtypes.ListValueOf[types.String] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
	Enabled types.Bool                        `tfsdk:"enabled"`
}

// Test Rule 2: XML wrappers with Items/Quantity + additional fields (e.g., Enabled)
func TestExpandXMLWrapperRule2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type tfModel struct {
		Field fwtypes.ListNestedObjectValueOf[testRule2Model] `tfsdk:"field" autoflex:",omitempty"`
	}
	type awsModel struct {
		Field *testXMLWrapperRule2
	}
	type awsModelDifferentOrder struct {
		Field *testXMLWrapperRule2DifferentOrder
	}

	testCases := map[string]autoFlexTestCases{
		"TestXMLWrapperRule2": {
			"null block": {
				Source:     &tfModel{Field: fwtypes.NewListNestedObjectValueOfNull[testRule2Model](ctx)},
				Target:     &awsModel{},
				WantTarget: &awsModel{Field: nil},
			},
			"empty items, enabled false": {
				Source: &tfModel{Field: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testRule2Model{
					{
						Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{}),
						Enabled: types.BoolValue(false),
					},
				})},
				Target:     &awsModel{},
				WantTarget: &awsModel{Field: &testXMLWrapperRule2{Items: []string{}, Quantity: aws.Int32(0), Enabled: aws.Bool(false)}},
			},
			"with items, enabled true": {
				Source: &tfModel{Field: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testRule2Model{
					{
						Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("item1")}),
						Enabled: types.BoolValue(true),
					},
				})},
				Target:     &awsModel{},
				WantTarget: &awsModel{Field: &testXMLWrapperRule2{Items: []string{"item1"}, Quantity: aws.Int32(1), Enabled: aws.Bool(true)}},
			},
		},
		"TestXMLWrapperRule2DifferentOrder": {
			"different field order": {
				Source: &tfModel{Field: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testRule2Model{
					{
						Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("item1")}),
						Enabled: types.BoolValue(true),
					},
				})},
				Target:     &awsModelDifferentOrder{},
				WantTarget: &awsModelDifferentOrder{Field: &testXMLWrapperRule2DifferentOrder{Items: []string{"item1"}, Quantity: aws.Int32(1), Enabled: aws.Bool(true)}},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
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
		XMLWrapper fwtypes.ListNestedObjectValueOf[testRule2Model] `autoflex:",omitempty"`
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
				// Empty state (Enabled=false, Items=[]) should flatten to null per Rule 2
				WantTarget: &targetStruct{XMLWrapper: fwtypes.NewListNestedObjectValueOfNull[testRule2Model](ctx)},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

// Test Rule 2 Symmetry: Verify expand→flatten produces identical results
func TestXMLWrapperRule2Symmetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Rule2Pattern", func(t *testing.T) {
		t.Parallel()

		// Wrap source and target in structs with xmlwrapper tags
		type tfSource struct {
			XMLWrapper fwtypes.ListNestedObjectValueOf[testRule2Model] `tfsdk:"xml_wrapper"`
		}

		type awsTarget struct {
			XMLWrapper *testXMLWrapperRule2
		}

		// Original Terraform value (Rule 2 pattern)
		original := tfSource{
			XMLWrapper: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testRule2Model{
				{
					Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("signer1"), types.StringValue("signer2")}),
					Enabled: types.BoolValue(true),
				},
			}),
		}

		// Expand: TF → AWS
		var awsStruct awsTarget
		expandDiags := Expand(ctx, original, &awsStruct)
		if expandDiags.HasError() {
			t.Fatalf("Expand failed: %v", expandDiags)
		}

		// Verify expand result
		if awsStruct.XMLWrapper == nil {
			t.Fatal("Expected non-nil XMLWrapper")
		}
		if len(awsStruct.XMLWrapper.Items) != 2 || *awsStruct.XMLWrapper.Quantity != 2 || !*awsStruct.XMLWrapper.Enabled {
			t.Errorf("Expand produced incorrect result: Items=%v, Quantity=%v, Enabled=%v",
				awsStruct.XMLWrapper.Items, *awsStruct.XMLWrapper.Quantity, *awsStruct.XMLWrapper.Enabled)
		}

		// Flatten: AWS → TF
		var targetStruct tfSource
		flattenDiags := Flatten(ctx, &awsStruct, &targetStruct)
		if flattenDiags.HasError() {
			t.Fatalf("Flatten failed: %v", flattenDiags)
		}

		// Verify symmetry: flattened result should match original
		if !original.XMLWrapper.Equal(targetStruct.XMLWrapper) {
			t.Errorf("Symmetry broken: original=%v, flattened=%v", original.XMLWrapper, targetStruct.XMLWrapper)
		}
	})
}

// Test Real CloudFront Types: Validate actual AWS SDK types
func TestXMLWrapperRealCloudFrontTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test with actual CloudFront TrustedSigners type
	t.Run("TrustedSigners", func(t *testing.T) {
		t.Parallel()

		// Terraform model matching CloudFront schema
		type trustedSignersModel struct {
			Items   fwtypes.ListValueOf[types.String] `tfsdk:"items"`
			Enabled types.Bool                        `tfsdk:"enabled"`
		}

		type tfSource struct {
			TrustedSigners fwtypes.ListNestedObjectValueOf[trustedSignersModel] `tfsdk:"trusted_signers" autoflex:",xmlwrapper=Items"`
		}

		type awsTarget struct {
			TrustedSigners *awstypes.TrustedSigners
		}

		source := tfSource{
			TrustedSigners: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []trustedSignersModel{
				{
					Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("AKIAIOSFODNN7EXAMPLE")}),
					Enabled: types.BoolValue(true),
				},
			}),
		}

		var target awsTarget
		diags := Expand(ctx, source, &target)

		if diags.HasError() {
			t.Errorf("TrustedSigners expand failed: %v", diags)
		} else {
			// Verify correct AWS structure
			if target.TrustedSigners == nil {
				t.Fatal("Expected non-nil TrustedSigners")
			}
			if len(target.TrustedSigners.Items) != 1 || *target.TrustedSigners.Quantity != 1 || !*target.TrustedSigners.Enabled {
				t.Errorf("TrustedSigners incorrect: Items=%v, Quantity=%v, Enabled=%v",
					target.TrustedSigners.Items, *target.TrustedSigners.Quantity, *target.TrustedSigners.Enabled)
			}
		}
	})
}

// Test Rule 1: XML wrapper with complex struct items (like Origins)
func TestFlattenXMLWrapperRule1ComplexStructItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// AWS types: Origins with Items/Quantity
	type awsOrigin struct {
		DomainName *string
		Id         *string
	}

	type awsOrigins struct {
		Items    []awsOrigin
		Quantity *int32
	}

	type awsDistributionConfig struct {
		Origins *awsOrigins
		Comment *string
	}

	// Terraform models
	type tfOriginModel struct {
		DomainName types.String `tfsdk:"domain_name"`
		Id         types.String `tfsdk:"id"`
	}

	type tfDistributionModel struct {
		Origin  fwtypes.SetNestedObjectValueOf[tfOriginModel] `tfsdk:"origin" autoflex:",xmlwrapper=Items"`
		Comment types.String                                  `tfsdk:"comment"`
	}

	t.Run("flatten origins", func(t *testing.T) {
		t.Parallel()

		source := &awsDistributionConfig{
			Origins: &awsOrigins{
				Items: []awsOrigin{
					{DomainName: aws.String("example.com"), Id: aws.String("origin1")},
					{DomainName: aws.String("cdn.example.com"), Id: aws.String("origin2")},
				},
				Quantity: aws.Int32(2),
			},
			Comment: aws.String("test distribution"),
		}

		var target tfDistributionModel
		diags := Flatten(ctx, source, &target)

		if diags.HasError() {
			t.Fatalf("Flatten failed: %v", diags)
		}

		if target.Origin.IsNull() {
			t.Fatal("Expected non-null origin set")
		}

		elements := target.Origin.Elements()
		if len(elements) != 2 {
			t.Errorf("Expected 2 origins, got %d", len(elements))
		}
	})
}

// Test that structs with Items/Quantity but NO xmlwrapper tag are NOT treated as XML wrappers
func TestExpandNoXMLWrapperTag(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// AWS types
	type awsFunctionAssociation struct {
		FunctionArn *string
		EventType   string
	}

	type awsFunctionAssociations struct {
		Items    []awsFunctionAssociation
		Quantity *int32
	}

	// Terraform types (NO xmlwrapper tag)
	type tfFunctionAssociation struct {
		EventType   types.String `tfsdk:"event_type"`
		FunctionArn types.String `tfsdk:"function_arn"`
	}

	type tfFunctionAssociations struct {
		Items    fwtypes.ListNestedObjectValueOf[tfFunctionAssociation] `tfsdk:"items"`
		Quantity types.Int64                                            `tfsdk:"quantity"`
	}

	// Source with nested struct containing Items/Quantity
	source := tfFunctionAssociations{
		Items: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFunctionAssociation{
			{
				EventType:   types.StringValue("viewer-request"),
				FunctionArn: types.StringValue("arn:aws:cloudfront::123456789012:function/test-1"),
			},
		}),
		Quantity: types.Int64Value(1),
	}

	var target awsFunctionAssociations
	diags := Expand(ctx, source, &target)

	if diags.HasError() {
		t.Fatalf("Expand failed: %v", diags)
	}

	// Without xmlwrapper tag, should map fields directly (not treat as wrapper)
	if len(target.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(target.Items))
	}
	if target.Quantity == nil || *target.Quantity != 1 {
		t.Errorf("Expected Quantity=1, got %v", target.Quantity)
	}
}

func TestFlattenNoXMLWrapperTag(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// AWS types
	type awsFunctionAssociation struct {
		FunctionArn *string
		EventType   string
	}

	type awsFunctionAssociations struct {
		Items    []awsFunctionAssociation
		Quantity *int32
	}

	// Terraform types (NO xmlwrapper tag)
	type tfFunctionAssociation struct {
		EventType   types.String `tfsdk:"event_type"`
		FunctionArn types.String `tfsdk:"function_arn"`
	}

	type tfFunctionAssociations struct {
		Items    fwtypes.ListNestedObjectValueOf[tfFunctionAssociation] `tfsdk:"items"`
		Quantity types.Int64                                            `tfsdk:"quantity"`
	}

	source := awsFunctionAssociations{
		Quantity: aws.Int32(1),
		Items: []awsFunctionAssociation{
			{
				EventType:   "viewer-request",
				FunctionArn: aws.String("arn:aws:cloudfront::123456789012:function/test-1"),
			},
		},
	}

	var target tfFunctionAssociations
	diags := Flatten(ctx, source, &target)

	if diags.HasError() {
		t.Fatalf("Flatten failed: %v", diags)
	}

	// Without xmlwrapper tag, should map fields directly (not treat as wrapper)
	if target.Items.IsNull() {
		t.Error("Expected non-null Items")
	}
	if target.Quantity.IsNull() || target.Quantity.ValueInt64() != 1 {
		t.Errorf("Expected Quantity=1, got %v", target.Quantity)
	}
}

// Test nested XML wrappers: wrapper inside wrapper
// Example: CacheBehaviors (Rule 1) → CacheBehavior → TrustedKeyGroups (Rule 2)
func TestNestedXMLWrappers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// AWS types: Outer wrapper (Rule 1) containing inner wrapper (Rule 2)
	type awsTrustedKeyGroups struct {
		Items    []string
		Quantity *int32
		Enabled  *bool
	}

	type awsCacheBehavior struct {
		PathPattern      *string
		TargetOriginId   *string
		TrustedKeyGroups *awsTrustedKeyGroups
	}

	type awsCacheBehaviors struct {
		Items    []awsCacheBehavior
		Quantity *int32
	}

	// Terraform types: Both levels have xmlwrapper tags
	type tfTrustedKeyGroups struct {
		Items   fwtypes.ListValueOf[types.String] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
		Enabled types.Bool                        `tfsdk:"enabled"`
	}

	type tfCacheBehavior struct {
		PathPattern      types.String                                        `tfsdk:"path_pattern"`
		TargetOriginId   types.String                                        `tfsdk:"target_origin_id"`
		TrustedKeyGroups fwtypes.ListNestedObjectValueOf[tfTrustedKeyGroups] `tfsdk:"trusted_key_groups"`
	}

	t.Run("Expand", func(t *testing.T) {
		t.Parallel()

		// Terraform source with nested wrappers
		type tfSource struct {
			CacheBehaviors fwtypes.SetNestedObjectValueOf[tfCacheBehavior] `tfsdk:"cache_behaviors" autoflex:",xmlwrapper=Items"`
		}

		source := tfSource{
			CacheBehaviors: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfCacheBehavior{
				{
					PathPattern:    types.StringValue("/api/*"),
					TargetOriginId: types.StringValue("api-origin"),
					TrustedKeyGroups: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTrustedKeyGroups{
						{
							Items:   fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("key-group-1")}),
							Enabled: types.BoolValue(true),
						},
					}),
				},
			}),
		}

		type awsTarget struct {
			CacheBehaviors *awsCacheBehaviors
		}

		var target awsTarget
		diags := Expand(ctx, source, &target)

		if diags.HasError() {
			t.Fatalf("Expand failed: %v", diags)
		}

		// Verify outer wrapper
		if target.CacheBehaviors == nil {
			t.Fatal("Expected non-nil CacheBehaviors")
		}
		if len(target.CacheBehaviors.Items) != 1 || *target.CacheBehaviors.Quantity != 1 {
			t.Errorf("Outer wrapper incorrect: Items=%d, Quantity=%v", len(target.CacheBehaviors.Items), *target.CacheBehaviors.Quantity)
		}

		// Verify inner wrapper
		behavior := target.CacheBehaviors.Items[0]
		if behavior.TrustedKeyGroups == nil {
			t.Fatal("Expected non-nil TrustedKeyGroups")
		}
		if len(behavior.TrustedKeyGroups.Items) != 1 || *behavior.TrustedKeyGroups.Quantity != 1 {
			t.Errorf("Inner wrapper incorrect: Items=%d, Quantity=%v",
				len(behavior.TrustedKeyGroups.Items), *behavior.TrustedKeyGroups.Quantity)
		}
		if !*behavior.TrustedKeyGroups.Enabled {
			t.Error("Expected Enabled=true")
		}
	})

	t.Run("Flatten", func(t *testing.T) {
		t.Parallel()

		// AWS source with nested wrappers
		type awsSource struct {
			CacheBehaviors *awsCacheBehaviors
		}

		source := awsSource{
			CacheBehaviors: &awsCacheBehaviors{
				Quantity: aws.Int32(1),
				Items: []awsCacheBehavior{
					{
						PathPattern:    aws.String("/api/*"),
						TargetOriginId: aws.String("api-origin"),
						TrustedKeyGroups: &awsTrustedKeyGroups{
							Enabled:  aws.Bool(true),
							Quantity: aws.Int32(1),
							Items:    []string{"key-group-1"},
						},
					},
				},
			},
		}

		type tfTarget struct {
			CacheBehaviors fwtypes.SetNestedObjectValueOf[tfCacheBehavior] `tfsdk:"cache_behaviors" autoflex:",xmlwrapper=Items"`
		}

		var target tfTarget
		diags := Flatten(ctx, &source, &target)

		if diags.HasError() {
			t.Fatalf("Flatten failed: %v", diags)
		}

		// Verify outer wrapper flattened correctly
		if target.CacheBehaviors.IsNull() {
			t.Fatal("Expected non-null CacheBehaviors")
		}

		// Extract the actual behavior structs
		var behaviors []tfCacheBehavior
		diags = target.CacheBehaviors.ElementsAs(ctx, &behaviors, false)
		if diags.HasError() {
			t.Fatalf("ElementsAs failed: %v", diags)
		}

		if len(behaviors) != 1 {
			t.Fatalf("Expected 1 cache behavior, got %d", len(behaviors))
		}

		// Verify inner wrapper flattened correctly
		behavior := behaviors[0]
		if behavior.TrustedKeyGroups.IsNull() {
			t.Error("Expected non-null TrustedKeyGroups")
		}

		var keyGroups []tfTrustedKeyGroups
		diags = behavior.TrustedKeyGroups.ElementsAs(ctx, &keyGroups, false)
		if diags.HasError() {
			t.Fatalf("ElementsAs failed: %v", diags)
		}

		if len(keyGroups) != 1 {
			t.Fatalf("Expected 1 TrustedKeyGroups element, got %d", len(keyGroups))
		}

		keyGroup := keyGroups[0]
		if keyGroup.Items.IsNull() {
			t.Error("Expected non-null Items")
		}
		if keyGroup.Enabled.IsNull() || !keyGroup.Enabled.ValueBool() {
			t.Error("Expected Enabled=true")
		}
	})
}

func TestFlattenXMLWrapperSplitNested(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type awsModel = awstypes.DistributionConfig

	type allowedMethodsModel struct {
		Items         fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
		CachedMethods fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]] `tfsdk:"cached_methods" autoflex:",xmlwrapper=Items"`
	}

	type defaultCacheBehaviorModel struct {
		AllowedMethods fwtypes.ListNestedObjectValueOf[allowedMethodsModel] `tfsdk:"allowed_methods"`
	}

	type tfModel struct {
		DefaultCacheBehavior fwtypes.ListNestedObjectValueOf[defaultCacheBehaviorModel] `tfsdk:"default_cache_behavior"`
	}

	testCases := autoFlexTestCases{
		"nil source": {
			Source: &awsModel{DefaultCacheBehavior: &awstypes.DefaultCacheBehavior{
				AllowedMethods: nil,
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				DefaultCacheBehavior: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModel{
					{
						AllowedMethods: fwtypes.NewListNestedObjectValueOfNull[allowedMethodsModel](ctx),
					},
				}),
			},
		},
		"empty items,nil cached": {
			Source: &awsModel{DefaultCacheBehavior: &awstypes.DefaultCacheBehavior{
				AllowedMethods: &awstypes.AllowedMethods{
					Items:         []awstypes.Method{},
					Quantity:      aws.Int32(0),
					CachedMethods: nil,
				},
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				DefaultCacheBehavior: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModel{
					{
						AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
							{
								Items:         fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
								CachedMethods: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[awstypes.Method]](ctx),
							},
						}),
					},
				}),
			},
		},
		"multi items,nil cached": {
			Source: &awsModel{DefaultCacheBehavior: &awstypes.DefaultCacheBehavior{
				AllowedMethods: &awstypes.AllowedMethods{
					Items:    []awstypes.Method{awstypes.MethodGet, awstypes.MethodPost},
					Quantity: aws.Int32(2),
				},
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				DefaultCacheBehavior: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModel{
					{
						AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
							{
								Items: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{
									fwtypes.StringEnumValue(awstypes.MethodGet),
									fwtypes.StringEnumValue(awstypes.MethodPost),
								}),
								CachedMethods: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[awstypes.Method]](ctx),
							},
						}),
					},
				}),
			},
		},
		"multi items,single cached": {
			Source: &awsModel{DefaultCacheBehavior: &awstypes.DefaultCacheBehavior{
				AllowedMethods: &awstypes.AllowedMethods{
					Items:    []awstypes.Method{awstypes.MethodGet, awstypes.MethodPost},
					Quantity: aws.Int32(2),
					CachedMethods: &awstypes.CachedMethods{
						Items:    []awstypes.Method{awstypes.MethodGet},
						Quantity: aws.Int32(1),
					},
				},
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				DefaultCacheBehavior: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModel{
					{
						AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
							{
								Items: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{
									fwtypes.StringEnumValue(awstypes.MethodGet),
									fwtypes.StringEnumValue(awstypes.MethodPost),
								}),
								CachedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{
									fwtypes.StringEnumValue(awstypes.MethodGet),
								}),
							},
						}),
					},
				}),
			},
		},

		"empty items,empty cached": {
			Source: &awsModel{DefaultCacheBehavior: &awstypes.DefaultCacheBehavior{
				AllowedMethods: &awstypes.AllowedMethods{
					Items:    []awstypes.Method{},
					Quantity: aws.Int32(0),
					CachedMethods: &awstypes.CachedMethods{
						Items:    []awstypes.Method{},
						Quantity: aws.Int32(0),
					},
				},
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				DefaultCacheBehavior: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModel{
					{
						AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
							{
								Items:         fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
								CachedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
							},
						}),
					},
				}),
			},
		},

		"empty items,single cached": {
			Source: &awsModel{DefaultCacheBehavior: &awstypes.DefaultCacheBehavior{
				AllowedMethods: &awstypes.AllowedMethods{
					Items:    []awstypes.Method{},
					Quantity: aws.Int32(0),
					CachedMethods: &awstypes.CachedMethods{
						Items:    []awstypes.Method{awstypes.MethodGet},
						Quantity: aws.Int32(1),
					},
				},
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				DefaultCacheBehavior: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []defaultCacheBehaviorModel{
					{
						AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
							{
								Items:         fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
								CachedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{fwtypes.StringEnumValue(awstypes.MethodGet)}),
							},
						}),
					},
				}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestFlattenXMLWrapperSplit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// type awsModel = awstypes.DefaultCacheBehavior
	type awsModel struct {
		AllowedMethods *awstypes.AllowedMethods
	}

	type allowedMethodsModel struct {
		Items         fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
		CachedMethods fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]] `tfsdk:"cached_methods" autoflex:",xmlwrapper=Items"`
	}

	type tfModel struct {
		AllowedMethods fwtypes.ListNestedObjectValueOf[allowedMethodsModel] `tfsdk:"allowed_methods"`
	}

	testCases := autoFlexTestCases{
		"nil source": {
			Source: &awsModel{AllowedMethods: nil},
			Target: &tfModel{},
			WantTarget: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfNull[allowedMethodsModel](ctx),
			},
		},
		"empty items,nil cached": {
			Source: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:         []awstypes.Method{},
				Quantity:      aws.Int32(0),
				CachedMethods: nil,
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
					{
						Items:         fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
						CachedMethods: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[awstypes.Method]](ctx),
					},
				}),
			},
		},
		"multi items,nil cached": {
			Source: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:    []awstypes.Method{awstypes.MethodGet, awstypes.MethodPost},
				Quantity: aws.Int32(2),
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
					{
						Items: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{
							fwtypes.StringEnumValue(awstypes.MethodGet),
							fwtypes.StringEnumValue(awstypes.MethodPost),
						}),
						CachedMethods: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[awstypes.Method]](ctx),
					},
				}),
			},
		},
		"multi items,single cached": {
			Source: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:    []awstypes.Method{awstypes.MethodGet, awstypes.MethodPost},
				Quantity: aws.Int32(2),
				CachedMethods: &awstypes.CachedMethods{
					Items:    []awstypes.Method{awstypes.MethodGet},
					Quantity: aws.Int32(1),
				},
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
					{
						Items: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{
							fwtypes.StringEnumValue(awstypes.MethodGet),
							fwtypes.StringEnumValue(awstypes.MethodPost),
						}),
						CachedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{
							fwtypes.StringEnumValue(awstypes.MethodGet),
						}),
					},
				}),
			},
		},

		"empty items,empty cached": {
			Source: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:    []awstypes.Method{},
				Quantity: aws.Int32(0),
				CachedMethods: &awstypes.CachedMethods{
					Items:    []awstypes.Method{},
					Quantity: aws.Int32(0),
				},
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
					{
						Items:         fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
						CachedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
					},
				}),
			},
		},

		"empty items,single cached": {
			Source: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:    []awstypes.Method{},
				Quantity: aws.Int32(0),
				CachedMethods: &awstypes.CachedMethods{
					Items:    []awstypes.Method{awstypes.MethodGet},
					Quantity: aws.Int32(1),
				},
			}},
			Target: &tfModel{},
			WantTarget: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
					{
						Items:         fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
						CachedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{fwtypes.StringEnumValue(awstypes.MethodGet)}),
					},
				}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestExpandXMLWrapperSplit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// type awsModel = awstypes.DefaultCacheBehavior
	type awsModel struct {
		AllowedMethods *awstypes.AllowedMethods
	}

	type allowedMethodsModel struct {
		Items         fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
		CachedMethods fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]] `tfsdk:"cached_methods" autoflex:",xmlwrapper=Items"`
	}

	type tfModel struct {
		AllowedMethods fwtypes.ListNestedObjectValueOf[allowedMethodsModel] `tfsdk:"allowed_methods"`
	}

	testCases := autoFlexTestCases{
		"nil source": {
			Source: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfNull[allowedMethodsModel](ctx),
			},
			Target: &awsModel{},
			WantTarget: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:         []awstypes.Method{},
				Quantity:      aws.Int32(0),
				CachedMethods: nil,
			}},
		},
		"empty items,nil cached": {
			Source: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
					{
						Items:         fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
						CachedMethods: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[awstypes.Method]](ctx),
					},
				}),
			},
			Target: &awsModel{},
			WantTarget: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:    []awstypes.Method{},
				Quantity: aws.Int32(0),
				CachedMethods: &awstypes.CachedMethods{
					Items:    []awstypes.Method{},
					Quantity: aws.Int32(0),
				},
			}},
		},
		"multi items,nil cached": {
			Source: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
					{
						Items: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{
							fwtypes.StringEnumValue(awstypes.MethodGet),
							fwtypes.StringEnumValue(awstypes.MethodPost),
						}),
						CachedMethods: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[awstypes.Method]](ctx),
					},
				}),
			},
			Target: &awsModel{},
			WantTarget: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:    []awstypes.Method{awstypes.MethodGet, awstypes.MethodPost},
				Quantity: aws.Int32(2),
				CachedMethods: &awstypes.CachedMethods{
					Items:    []awstypes.Method{},
					Quantity: aws.Int32(0),
				},
			}},
		},
		"multi items,single cached": {
			Source: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
					{
						Items: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{
							fwtypes.StringEnumValue(awstypes.MethodGet),
							fwtypes.StringEnumValue(awstypes.MethodPost),
						}),
						CachedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{
							fwtypes.StringEnumValue(awstypes.MethodGet),
						}),
					},
				}),
			},
			Target: &awsModel{},
			WantTarget: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:    []awstypes.Method{awstypes.MethodGet, awstypes.MethodPost},
				Quantity: aws.Int32(2),
				CachedMethods: &awstypes.CachedMethods{
					Items:    []awstypes.Method{awstypes.MethodGet},
					Quantity: aws.Int32(1),
				},
			}},
		},

		"empty items,empty cached": {
			Source: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
					{
						Items:         fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
						CachedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
					},
				}),
			},
			Target: &awsModel{},
			WantTarget: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:    []awstypes.Method{},
				Quantity: aws.Int32(0),
				CachedMethods: &awstypes.CachedMethods{
					Items:    []awstypes.Method{},
					Quantity: aws.Int32(0),
				},
			}},
		},

		"empty items,single cached": {
			Source: &tfModel{
				AllowedMethods: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []allowedMethodsModel{
					{
						Items:         fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{}),
						CachedMethods: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.Method]](ctx, []attr.Value{fwtypes.StringEnumValue(awstypes.MethodGet)}),
					},
				}),
			},
			Target: &awsModel{},
			WantTarget: &awsModel{AllowedMethods: &awstypes.AllowedMethods{
				Items:    []awstypes.Method{},
				Quantity: aws.Int32(0),
				CachedMethods: &awstypes.CachedMethods{
					Items:    []awstypes.Method{awstypes.MethodGet},
					Quantity: aws.Int32(1),
				},
			}},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true}, cmpopts.IgnoreUnexported(
		awstypes.AllowedMethods{},
		awstypes.CachedMethods{},
	))
}

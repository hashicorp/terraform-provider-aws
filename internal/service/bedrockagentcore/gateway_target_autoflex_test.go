// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestBedrockAgentCoreGatewayTargetPrivateEndpointAutoFlexExpand(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	ignoreExportedOpts := cmpopts.IgnoreUnexported(
		awstypes.PrivateEndpointMemberManagedVpcResource{},
		awstypes.ManagedVpcResource{},
		awstypes.PrivateEndpointMemberSelfManagedLatticeResource{},
		awstypes.SelfManagedLatticeResourceMemberResourceConfigurationIdentifier{},
	)
	testCases := map[string]struct {
		model    privateEndpointModel
		expected awstypes.PrivateEndpoint
	}{
		"Simple ManagedVPCResource": {
			model: privateEndpointModel{
				ManagedVPCResource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &managedVPCResourceModel{
					EndpointIPAddressType: fwtypes.StringEnumValue(awstypes.EndpointIpAddressTypeIpv4),
					RoutingDomain:         types.StringNull(),
					SecurityGroupIDs:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, nil),
					SubnetIDs:             fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"sn1", "sn2"}),
					Tags:                  tftags.NewMapValueNull(),
					VPCIdentifier:         types.StringValue("vpc1"),
				}),
				SelfManagedLatticeResource: fwtypes.NewListNestedObjectValueOfNull[selfManagedLatticeResourceModel](ctx),
			},
			expected: &awstypes.PrivateEndpointMemberManagedVpcResource{
				Value: awstypes.ManagedVpcResource{
					EndpointIpAddressType: awstypes.EndpointIpAddressTypeIpv4,
					SubnetIds:             []string{"sn1", "sn2"},
					VpcIdentifier:         aws.String("vpc1"),
				},
			},
		},
		"Full ManagedVPCResource no tags": {
			model: privateEndpointModel{
				ManagedVPCResource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &managedVPCResourceModel{
					EndpointIPAddressType: fwtypes.StringEnumValue(awstypes.EndpointIpAddressTypeIpv4),
					RoutingDomain:         types.StringValue("rd1"),
					SecurityGroupIDs:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"sg1"}),
					SubnetIDs:             fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"sn1", "sn2"}),
					Tags:                  tftags.NewMapValueNull(),
					VPCIdentifier:         types.StringValue("vpc1"),
				}),
				SelfManagedLatticeResource: fwtypes.NewListNestedObjectValueOfNull[selfManagedLatticeResourceModel](ctx),
			},
			expected: &awstypes.PrivateEndpointMemberManagedVpcResource{
				Value: awstypes.ManagedVpcResource{
					EndpointIpAddressType: awstypes.EndpointIpAddressTypeIpv4,
					RoutingDomain:         aws.String("rd1"),
					SecurityGroupIds:      []string{"sg1"},
					SubnetIds:             []string{"sn1", "sn2"},
					VpcIdentifier:         aws.String("vpc1"),
				},
			},
		},
		"ManagedVPCResource tags": {
			model: privateEndpointModel{
				ManagedVPCResource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &managedVPCResourceModel{
					EndpointIPAddressType: fwtypes.StringEnumValue(awstypes.EndpointIpAddressTypeIpv4),
					RoutingDomain:         types.StringNull(),
					SecurityGroupIDs:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, nil),
					SubnetIDs:             fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"sn1", "sn2"}),
					Tags: tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMap(ctx, map[string]string{
						acctest.CtKey1: acctest.CtValue1,
						acctest.CtKey2: acctest.CtValue2,
					})),
					VPCIdentifier: types.StringValue("vpc1"),
				}),
				SelfManagedLatticeResource: fwtypes.NewListNestedObjectValueOfNull[selfManagedLatticeResourceModel](ctx),
			},
			expected: &awstypes.PrivateEndpointMemberManagedVpcResource{
				Value: awstypes.ManagedVpcResource{
					EndpointIpAddressType: awstypes.EndpointIpAddressTypeIpv4,
					SubnetIds:             []string{"sn1", "sn2"},
					Tags:                  map[string]string{acctest.CtKey1: acctest.CtValue1, acctest.CtKey2: acctest.CtValue2},
					VpcIdentifier:         aws.String("vpc1"),
				},
			},
		},
		"Simple SelfManagedLatticeResource": {
			model: privateEndpointModel{
				ManagedVPCResource: fwtypes.NewListNestedObjectValueOfNull[managedVPCResourceModel](ctx),
				SelfManagedLatticeResource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &selfManagedLatticeResourceModel{
					ResourceConfigurationIdentifier: types.StringValue("rc1"),
				}),
			},
			expected: &awstypes.PrivateEndpointMemberSelfManagedLatticeResource{
				Value: &awstypes.SelfManagedLatticeResourceMemberResourceConfigurationIdentifier{
					Value: "rc1",
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			switch testCase.expected.(type) {
			case *awstypes.PrivateEndpointMemberManagedVpcResource:
				var got awstypes.PrivateEndpointMemberManagedVpcResource
				diags := fwflex.Expand(ctx, testCase.model, &got)
				if diags.HasError() {
					t.Fatalf("unexpected error: %s", diags[0].Summary())
				}
				if diff := cmp.Diff(&got, testCase.expected, ignoreExportedOpts); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			case *awstypes.PrivateEndpointMemberSelfManagedLatticeResource:
				var got awstypes.PrivateEndpointMemberSelfManagedLatticeResource
				diags := fwflex.Expand(ctx, testCase.model, &got)
				if diags.HasError() {
					t.Fatalf("unexpected error: %s", diags[0].Summary())
				}
				if diff := cmp.Diff(&got, testCase.expected, ignoreExportedOpts); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}

func TestBedrockAgentCoreGatewayTargetPrivateEndpointAutoFlexFlatten(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	testCases := map[string]struct {
		apiObject awstypes.PrivateEndpoint
		expected  privateEndpointModel
	}{
		"Simple ManagedVPCResource": {
			apiObject: &awstypes.PrivateEndpointMemberManagedVpcResource{
				Value: awstypes.ManagedVpcResource{
					EndpointIpAddressType: awstypes.EndpointIpAddressTypeIpv4,
					SubnetIds:             []string{"sn1", "sn2"},
					VpcIdentifier:         aws.String("vpc1"),
				},
			},
			expected: privateEndpointModel{
				ManagedVPCResource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &managedVPCResourceModel{
					EndpointIPAddressType: fwtypes.StringEnumValue(awstypes.EndpointIpAddressTypeIpv4),
					RoutingDomain:         types.StringNull(),
					SecurityGroupIDs:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, nil),
					SubnetIDs:             fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"sn1", "sn2"}),
					Tags:                  tftags.NewMapValueNull(),
					VPCIdentifier:         types.StringValue("vpc1"),
				}),
				SelfManagedLatticeResource: fwtypes.NewListNestedObjectValueOfNull[selfManagedLatticeResourceModel](ctx),
			},
		},
		"Full ManagedVPCResource no tags": {
			apiObject: &awstypes.PrivateEndpointMemberManagedVpcResource{
				Value: awstypes.ManagedVpcResource{
					EndpointIpAddressType: awstypes.EndpointIpAddressTypeIpv4,
					RoutingDomain:         aws.String("rd1"),
					SecurityGroupIds:      []string{"sg1"},
					SubnetIds:             []string{"sn1", "sn2"},
					VpcIdentifier:         aws.String("vpc1"),
				},
			},
			expected: privateEndpointModel{
				ManagedVPCResource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &managedVPCResourceModel{
					EndpointIPAddressType: fwtypes.StringEnumValue(awstypes.EndpointIpAddressTypeIpv4),
					RoutingDomain:         types.StringValue("rd1"),
					SecurityGroupIDs:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"sg1"}),
					SubnetIDs:             fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"sn1", "sn2"}),
					Tags:                  tftags.NewMapValueNull(),
					VPCIdentifier:         types.StringValue("vpc1"),
				}),
				SelfManagedLatticeResource: fwtypes.NewListNestedObjectValueOfNull[selfManagedLatticeResourceModel](ctx),
			},
		},
		"ManagedVPCResource tags": {
			apiObject: &awstypes.PrivateEndpointMemberManagedVpcResource{
				Value: awstypes.ManagedVpcResource{
					EndpointIpAddressType: awstypes.EndpointIpAddressTypeIpv4,
					SubnetIds:             []string{"sn1", "sn2"},
					Tags:                  map[string]string{acctest.CtKey1: acctest.CtValue1, acctest.CtKey2: acctest.CtValue2},
					VpcIdentifier:         aws.String("vpc1"),
				},
			},
			expected: privateEndpointModel{
				ManagedVPCResource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &managedVPCResourceModel{
					EndpointIPAddressType: fwtypes.StringEnumValue(awstypes.EndpointIpAddressTypeIpv4),
					RoutingDomain:         types.StringNull(),
					SecurityGroupIDs:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, nil),
					SubnetIDs:             fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"sn1", "sn2"}),
					Tags: tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMap(ctx, map[string]string{
						acctest.CtKey1: acctest.CtValue1,
						acctest.CtKey2: acctest.CtValue2,
					})),
					VPCIdentifier: types.StringValue("vpc1"),
				}),
				SelfManagedLatticeResource: fwtypes.NewListNestedObjectValueOfNull[selfManagedLatticeResourceModel](ctx),
			},
		},
		"Simple SelfManagedLatticeResource": {
			apiObject: &awstypes.PrivateEndpointMemberSelfManagedLatticeResource{
				Value: &awstypes.SelfManagedLatticeResourceMemberResourceConfigurationIdentifier{
					Value: "rc1",
				},
			},
			expected: privateEndpointModel{
				ManagedVPCResource: fwtypes.NewListNestedObjectValueOfNull[managedVPCResourceModel](ctx),
				SelfManagedLatticeResource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &selfManagedLatticeResourceModel{
					ResourceConfigurationIdentifier: types.StringValue("rc1"),
				}),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got privateEndpointModel
			diags := fwflex.Flatten(ctx, testCase.apiObject, &got)
			if diags.HasError() {
				t.Fatalf("unexpected error: %s", diags[0].Summary())
			}
			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

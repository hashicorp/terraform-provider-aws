// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// TestCustomJWTAuthorizerPrivateEndpointRoundTrip proves that autoflex builds the
// correct SDK payload for the private_endpoint / private_endpoint_overrides fields
// (including the union nested inside an override list element) by round-tripping
// SDK -> model -> SDK.
func TestCustomJWTAuthorizerPrivateEndpointRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	in := awstypes.CustomJWTAuthorizerConfiguration{
		DiscoveryUrl:    aws.String("https://example.com/.well-known/openid-configuration"),
		AllowedAudience: []string{"aud"},
		PrivateEndpoint: &awstypes.PrivateEndpointMemberManagedVpcResource{
			Value: awstypes.ManagedVpcResource{
				VpcIdentifier:         aws.String("vpc-12345678"),
				SubnetIds:             []string{"subnet-12345678", "subnet-abcdef01"},
				EndpointIpAddressType: awstypes.EndpointIpAddressTypeIpv4,
				SecurityGroupIds:      []string{"sg-12345678"},
				RoutingDomain:         aws.String("example.com"),
			},
		},
		PrivateEndpointOverrides: []awstypes.PrivateEndpointOverride{
			{
				Domain: aws.String("override.example.com"),
				PrivateEndpoint: &awstypes.PrivateEndpointMemberSelfManagedLatticeResource{
					Value: &awstypes.SelfManagedLatticeResourceMemberResourceConfigurationIdentifier{
						Value: "rcfg-0123456789abcdef0",
					},
				},
			},
		},
	}

	var model customJWTAuthorizerConfigurationModel
	if diags := fwflex.Flatten(ctx, in, &model); diags.HasError() {
		t.Fatalf("Flatten: %v", diags)
	}

	var out awstypes.CustomJWTAuthorizerConfiguration
	if diags := fwflex.Expand(ctx, model, &out); diags.HasError() {
		t.Fatalf("Expand: %v", diags)
	}

	opts := cmp.Options{
		cmpopts.IgnoreUnexported(
			awstypes.CustomJWTAuthorizerConfiguration{},
			awstypes.ManagedVpcResource{},
			awstypes.PrivateEndpointMemberManagedVpcResource{},
			awstypes.PrivateEndpointMemberSelfManagedLatticeResource{},
			awstypes.SelfManagedLatticeResourceMemberResourceConfigurationIdentifier{},
			awstypes.PrivateEndpointOverride{},
		),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}
	if diff := cmp.Diff(in, out, opts...); diff != "" {
		t.Errorf("SDK -> model -> SDK round-trip mismatch (-in +out):\n%s", diff)
	}
}

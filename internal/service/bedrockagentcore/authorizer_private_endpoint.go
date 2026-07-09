// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// privateEndpointBlock is the shared schema for a private endpoint (a union of a
// service-managed VPC resource or a self-managed VPC Lattice resource). The
// backing models and their Expand/Flatten live in gateway_target.go; this block
// is reused by gateway_target, by the shared JWT authorizer's custom_jwt_authorizer,
// and (recursively) by private_endpoint_overrides.
func privateEndpointBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[privateEndpointModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"managed_vpc_resource": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[managedVPCResourceModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.ExactlyOneOf(
							path.MatchRelative().AtParent().AtName("managed_vpc_resource"),
							path.MatchRelative().AtParent().AtName("self_managed_lattice_resource"),
						),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"endpoint_ip_address_type": schema.StringAttribute{
								Required:   true,
								CustomType: fwtypes.StringEnumType[awstypes.EndpointIpAddressType](),
							},
							"routing_domain": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(3, 255),
								},
							},
							names.AttrSecurityGroupIDs: schema.SetAttribute{
								CustomType: fwtypes.SetOfStringType,
								Optional:   true,
								Validators: []validator.Set{
									setvalidator.SizeAtMost(5),
								},
							},
							names.AttrSubnetIDs: schema.SetAttribute{
								CustomType: fwtypes.SetOfStringType,
								Required:   true,
							},
							names.AttrTags: tftags.TagsAttribute(),
							"vpc_identifier": schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				"self_managed_lattice_resource": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[selfManagedLatticeResourceModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"resource_configuration_identifier": schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

// privateEndpointOverridesBlock is a list (max 5) of per-domain private endpoint
// overrides used by the custom JWT authorizer configuration.
func privateEndpointOverridesBlock(ctx context.Context) schema.ListNestedBlock {
	// Unlike the optional top-level custom_jwt_authorizer.private_endpoint, an
	// override's private_endpoint is required by the SDK (PrivateEndpointOverride.
	// PrivateEndpoint). Take a fresh copy of the shared block and mark it required
	// for this instance only (so gateway_target's/top-level's stay optional).
	requiredPrivateEndpoint := privateEndpointBlock(ctx)
	requiredPrivateEndpoint.Validators = append(requiredPrivateEndpoint.Validators, listvalidator.IsRequired())

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[privateEndpointOverrideModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(5),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"domain": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 253),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"private_endpoint": requiredPrivateEndpoint,
			},
		},
	}
}

type privateEndpointOverrideModel struct {
	Domain          types.String                                          `tfsdk:"domain"`
	PrivateEndpoint fwtypes.ListNestedObjectValueOf[privateEndpointModel] `tfsdk:"private_endpoint"`
}

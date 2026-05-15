// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// This file adds private_endpoint support to aws_bedrockagentcore_gateway_target.
//
// The PrivateEndpoint union type has two variants:
//   - managed_vpc_resource:          AWS manages the VPC Lattice resource gateway on your behalf.
//   - self_managed_lattice_resource: You supply an existing VPC Lattice resource configuration ARN.
//
// API reference:
//   https://docs.aws.amazon.com/bedrock-agentcore-control/latest/APIReference/API_PrivateEndpoint.html
//   https://docs.aws.amazon.com/bedrock-agentcore-control/latest/APIReference/API_ManagedVpcResource.html

import (
	"context"
	"fmt"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// privateEndpointSchema returns the schema block for the private_endpoint argument.
// It is a union: exactly one of managed_vpc_resource or self_managed_lattice_resource
// may be set at a time.
func privateEndpointSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[privateEndpointModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				// managed_vpc_resource: AWS creates and manages the VPC Lattice
				// resource gateway and resource configuration on your behalf.
				"managed_vpc_resource": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[managedVpcResourceModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							// vpcIdentifier is the ID of the VPC that contains the private resource.
							names.AttrVPCID: schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										vpcIDRegex,
										"must be a valid VPC ID (vpc-xxxxxxxx or vpc-xxxxxxxxxxxxxxxxx)",
									),
								},
							},
							// subnetIds: subnets inside the VPC where Lattice ENIs are placed.
							names.AttrSubnetIDs: schema.SetAttribute{
								CustomType: fwtypes.SetOfStringType,
								Required:   true,
							},
							// endpointIpAddressType: IPV4 or IPV6.
							"endpoint_ip_address_type": schema.StringAttribute{
								Required:   true,
								CustomType: fwtypes.StringEnumType[awstypes.EndpointIpAddressType](),
							},
							// securityGroupIds: up to 5 security groups for the Lattice resource gateway.
							names.AttrSecurityGroupIDs: schema.SetAttribute{
								CustomType: fwtypes.SetOfStringType,
								Optional:   true,
							},
							// routingDomain: optional intermediate domain (e.g. a VPCE or ALB DNS name)
							// to use instead of the actual target domain.
							"routing_domain": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(3, 255),
								},
							},
							names.AttrTags: schema.MapAttribute{
								CustomType: fwtypes.MapOfStringType,
								Optional:   true,
							},
						},
					},
				},
				// self_managed_lattice_resource: you supply an existing VPC Lattice
				// resource configuration ARN or ID.
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

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

// privateEndpointModel is the top-level union model for the private_endpoint block.
// Exactly one of ManagedVpcResource or SelfManagedLatticeResource must be set.
type privateEndpointModel struct {
	ManagedVpcResource         fwtypes.ListNestedObjectValueOf[managedVpcResourceModel]         `tfsdk:"managed_vpc_resource"`
	SelfManagedLatticeResource fwtypes.ListNestedObjectValueOf[selfManagedLatticeResourceModel] `tfsdk:"self_managed_lattice_resource"`
}

// Ensure privateEndpointModel satisfies the fwflex union interfaces.
var (
	_ fwflex.Expander  = privateEndpointModel{}
	_ fwflex.Flattener = &privateEndpointModel{}
)

// Flatten converts the AWS SDK PrivateEndpoint union into the Terraform model.
func (m *privateEndpointModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.PrivateEndpointMemberManagedVpcResource:
		var model managedVpcResourceModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.ManagedVpcResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case *awstypes.PrivateEndpointMemberManagedVpcResource:
		var model managedVpcResourceModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.ManagedVpcResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.PrivateEndpointMemberSelfManagedLatticeResource:
		// t.Value is itself a SelfManagedLatticeResource union — delegate to the model's Flatten.
		var model selfManagedLatticeResourceModel
		smerr.AddEnrich(ctx, &diags, model.Flatten(ctx, t.Value))
		if diags.HasError() {
			return diags
		}
		m.SelfManagedLatticeResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case *awstypes.PrivateEndpointMemberSelfManagedLatticeResource:
		var model selfManagedLatticeResourceModel
		smerr.AddEnrich(ctx, &diags, model.Flatten(ctx, t.Value))
		if diags.HasError() {
			return diags
		}
		m.SelfManagedLatticeResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported PrivateEndpoint Type",
			fmt.Sprintf("private endpoint flatten: unexpected type %T", v),
		)
	}
	return diags
}

// Expand converts the Terraform model back into the AWS SDK PrivateEndpoint union.
func (m privateEndpointModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.ManagedVpcResource.IsNull():
		data, d := m.ManagedVpcResource.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.PrivateEndpointMemberManagedVpcResource
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.SelfManagedLatticeResource.IsNull():
		data, d := m.SelfManagedLatticeResource.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		// selfManagedLatticeResourceModel.Expand returns the SDK union member directly.
		expanded, d := data.Expand(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		// Wrap in the outer PrivateEndpoint union member.
		return &awstypes.PrivateEndpointMemberSelfManagedLatticeResource{
			Value: expanded.(awstypes.SelfManagedLatticeResource),
		}, diags
	}

	// Neither variant set — return nil (no private endpoint).
	return nil, diags
}

// managedVpcResourceModel maps to awstypes.ManagedVpcResource.
// AWS creates and manages the VPC Lattice resource gateway and resource
// configuration using a service-linked role.
//
// NOTE: The model field is named VpcIdentifier (matching the SDK struct field)
// even though the Terraform attribute is "vpc_id". fwflex maps Go field names
// to SDK field names, so VpcID would map to VpcId — not VpcIdentifier — causing
// a "missing required field" API error.
type managedVpcResourceModel struct {
	VpcIdentifier         types.String                                       `tfsdk:"vpc_id"`
	SubnetIDs             fwtypes.SetOfString                                `tfsdk:"subnet_ids"`
	EndpointIPAddressType fwtypes.StringEnum[awstypes.EndpointIpAddressType] `tfsdk:"endpoint_ip_address_type"`
	SecurityGroupIDs      fwtypes.SetOfString                                `tfsdk:"security_group_ids"`
	RoutingDomain         types.String                                       `tfsdk:"routing_domain"`
	Tags                  fwtypes.MapOfString                                `tfsdk:"tags"`
}

var _ fwflex.Flattener = &managedVpcResourceModel{}

// Flatten populates the model from the SDK ManagedVpcResource struct.
// A custom Flattener is needed to ensure the Tags field is never a zero-value
// MapValueOf (nil internal map), which causes a framework panic during state
// serialization when the API returns nil tags.
func (m *managedVpcResourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	sdk, ok := v.(awstypes.ManagedVpcResource)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError(
			"Unsupported ManagedVpcResource Type",
			fmt.Sprintf("managed vpc resource flatten: unexpected type %T", v),
		)
		return diags
	}

	var diags diag.Diagnostics

	m.VpcIdentifier = types.StringPointerValue(sdk.VpcIdentifier)
	m.RoutingDomain = types.StringPointerValue(sdk.RoutingDomain)
	m.EndpointIPAddressType = fwtypes.StringEnumValue(sdk.EndpointIpAddressType)

	subnetElems := make([]attr.Value, len(sdk.SubnetIds))
	for i, s := range sdk.SubnetIds {
		subnetElems[i] = types.StringValue(s)
	}
	subnetIDs, d := fwtypes.NewSetValueOf[basetypes.StringValue](ctx, subnetElems)
	diags.Append(d...)
	m.SubnetIDs = subnetIDs

	// Return null (not empty set) when no security groups are present so that
	// the refresh plan stays empty when security_group_ids is omitted from config.
	if len(sdk.SecurityGroupIds) == 0 {
		m.SecurityGroupIDs = fwtypes.NewSetValueOfNull[basetypes.StringValue](ctx)
	} else {
		sgElems := make([]attr.Value, len(sdk.SecurityGroupIds))
		for i, s := range sdk.SecurityGroupIds {
			sgElems[i] = types.StringValue(s)
		}
		sgIDs, d := fwtypes.NewSetValueOf[basetypes.StringValue](ctx, sgElems)
		diags.Append(d...)
		m.SecurityGroupIDs = sgIDs
	}

	// Convert map[string]string tags, defaulting to null when absent.
	if sdk.Tags != nil {
		tagElems := make(map[string]attr.Value, len(sdk.Tags))
		for k, v := range sdk.Tags {
			tagElems[k] = types.StringValue(v)
		}
		tags, d := fwtypes.NewMapValueOf[basetypes.StringValue](ctx, tagElems)
		diags.Append(d...)
		m.Tags = tags
	} else {
		m.Tags = fwtypes.NewMapValueOfNull[basetypes.StringValue](ctx)
	}

	return diags
}

// selfManagedLatticeResourceModel maps to awstypes.SelfManagedLatticeResource.
// The SDK type is itself a union with a single member:
//
//	SelfManagedLatticeResourceMemberResourceConfigurationIdentifier (string)
//
// We flatten it to a simple string attribute for usability.
type selfManagedLatticeResourceModel struct {
	ResourceConfigurationIdentifier types.String `tfsdk:"resource_configuration_identifier"`
}

var (
	_ fwflex.Expander  = selfManagedLatticeResourceModel{}
	_ fwflex.Flattener = &selfManagedLatticeResourceModel{}
)

// Flatten converts the SDK SelfManagedLatticeResource union into the model.
func (m *selfManagedLatticeResourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.SelfManagedLatticeResourceMemberResourceConfigurationIdentifier:
		m.ResourceConfigurationIdentifier = types.StringValue(t.Value)
	case *awstypes.SelfManagedLatticeResourceMemberResourceConfigurationIdentifier:
		m.ResourceConfigurationIdentifier = types.StringValue(t.Value)
	default:
		diags.AddError(
			"Unsupported SelfManagedLatticeResource Type",
			fmt.Sprintf("self managed lattice resource flatten: unexpected type %T", v),
		)
	}
	return diags
}

// Expand converts the model back into the SDK SelfManagedLatticeResource union.
func (m selfManagedLatticeResourceModel) Expand(_ context.Context) (any, diag.Diagnostics) {
	return &awstypes.SelfManagedLatticeResourceMemberResourceConfigurationIdentifier{
		Value: m.ResourceConfigurationIdentifier.ValueString(),
	}, nil
}

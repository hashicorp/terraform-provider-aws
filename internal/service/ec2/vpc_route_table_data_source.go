// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route_table", name="Route Table")
// @Testing(generator=false)
// @Testing(tagsIdentifierAttribute="id")
func dataSourceRouteTable() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRouteTableRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrSubnetID: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"gateway_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"route_table_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				names.AttrVPCID: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				names.AttrFilter: customFiltersSchema(),
				names.AttrTags:   tftags.TagsSchemaComputed(),
				"routes": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							///
							// Destinations.
							///
							names.AttrCIDRBlock: {
								Type:     schema.TypeString,
								Computed: true,
							},

							"ipv6_cidr_block": {
								Type:     schema.TypeString,
								Computed: true,
							},

							"destination_prefix_list_id": {
								Type:     schema.TypeString,
								Computed: true,
							},

							///
							// Targets.
							///
							"carrier_gateway_id": {
								Type:     schema.TypeString,
								Computed: true,
							},

							"core_network_arn": {
								Type:     schema.TypeString,
								Computed: true,
							},

							"egress_only_gateway_id": {
								Type:     schema.TypeString,
								Computed: true,
							},

							"gateway_id": {
								Type:     schema.TypeString,
								Computed: true,
							},

							names.AttrInstanceID: {
								Type:     schema.TypeString,
								Computed: true,
							},

							"local_gateway_id": {
								Type:     schema.TypeString,
								Computed: true,
							},

							"nat_gateway_id": {
								Type:     schema.TypeString,
								Computed: true,
							},

							names.AttrNetworkInterfaceID: {
								Type:     schema.TypeString,
								Computed: true,
							},

							"odb_network_arn": {
								Type:     schema.TypeString,
								Computed: true,
							},

							names.AttrTransitGatewayID: {
								Type:     schema.TypeString,
								Computed: true,
							},

							names.AttrVPCEndpointID: {
								Type:     schema.TypeString,
								Computed: true,
							},

							"vpc_peering_connection_id": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},

				"associations": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"route_table_association_id": {
								Type:     schema.TypeString,
								Computed: true,
							},

							"route_table_id": {
								Type:     schema.TypeString,
								Computed: true,
							},

							names.AttrSubnetID: {
								Type:     schema.TypeString,
								Computed: true,
							},

							"gateway_id": {
								Type:     schema.TypeString,
								Computed: true,
							},

							"main": {
								Type:     schema.TypeBool,
								Computed: true,
							},
						},
					},
				},

				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},

				names.AttrOwnerID: {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceRouteTableRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	var input ec2.DescribeRouteTablesInput
	vpcId, vpcIdOk := d.GetOk(names.AttrVPCID)
	subnetId, subnetIdOk := d.GetOk(names.AttrSubnetID)
	gatewayId, gatewayIdOk := d.GetOk("gateway_id")
	rtbId, rtbOk := d.GetOk("route_table_id")
	tags, tagsOk := d.GetOk(names.AttrTags)
	filter, filterOk := d.GetOk(names.AttrFilter)

	if !rtbOk && !vpcIdOk && !subnetIdOk && !gatewayIdOk && !filterOk && !tagsOk {
		return sdkdiag.AppendErrorf(diags, "one of route_table_id, vpc_id, subnet_id, gateway_id, filters, or tags must be assigned")
	}

	input.Filters = newAttributeFilterList(
		map[string]string{
			"route-table-id":         rtbId.(string),
			"vpc-id":                 vpcId.(string),
			"association.subnet-id":  subnetId.(string),
			"association.gateway-id": gatewayId.(string),
		},
	)
	input.Filters = append(input.Filters, newTagFilterList(
		svcTags(tftags.New(ctx, tags.(map[string]any))),
	)...)
	input.Filters = append(input.Filters, newCustomFilterList(
		filter.(*schema.Set),
	)...)

	rt, err := findRouteTable(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Route Table", err))
	}

	d.SetId(aws.ToString(rt.RouteTableId))

	ownerID := aws.ToString(rt.OwnerId)
	d.Set(names.AttrARN, routeTableARN(ctx, c, ownerID, d.Id()))
	if err := d.Set("associations", flattenRouteTableAssociations(rt.Associations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting associations: %s", err)
	}
	d.Set(names.AttrOwnerID, ownerID)
	d.Set("route_table_id", rt.RouteTableId)
	if err := d.Set("routes", flattenDataSourceRoutes(ctx, conn, rt.Routes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting routes: %s", err)
	}
	d.Set(names.AttrVPCID, rt.VpcId)

	//Ignore the AmazonFSx service tag in addition to standard ignores
	if err := d.Set(names.AttrTags, keyValueTags(ctx, rt.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Ignore(tftags.New(ctx, []string{"AmazonFSx"})).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

// Logic is very similar to flattenRoutes. Consider merging.
func flattenDataSourceRoutes(ctx context.Context, conn *ec2.Client, apiObjects []awstypes.Route) []map[string]any {
	tfList := make([]map[string]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		if gatewayID := aws.ToString(apiObject.GatewayId); gatewayID == gatewayIDLocal || gatewayID == gatewayIDVPCLattice {
			continue
		}

		if apiObject.Origin == awstypes.RouteOriginEnableVgwRoutePropagation {
			continue
		}

		if apiObject.DestinationPrefixListId != nil && strings.HasPrefix(aws.ToString(apiObject.GatewayId), "vpce-") {
			// Skipping because VPC endpoint routes are handled separately
			// See aws_vpc_endpoint
			continue
		}

		// Skip cross-account ENIs for AWS services.
		if networkInterfaceID := aws.ToString(apiObject.NetworkInterfaceId); networkInterfaceID != "" {
			networkInterface, err := findNetworkInterfaceByID(ctx, conn, networkInterfaceID)

			if err == nil && networkInterface.Attachment != nil {
				if ownerID, instanceOwnerID := aws.ToString(networkInterface.OwnerId), aws.ToString(networkInterface.Attachment.InstanceOwnerId); ownerID != "" && instanceOwnerID != ownerID {
					continue
				}
			}
		}

		tfList = append(tfList, flattenRoute(&apiObject))
	}

	return tfList
}

func flattenRouteTableAssociations(apiObjects []awstypes.RouteTableAssociation) []map[string]any {
	tfList := make([]map[string]any, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)
		tfMap["route_table_id"] = aws.ToString(apiObject.RouteTableId)
		tfMap["route_table_association_id"] = aws.ToString(apiObject.RouteTableAssociationId)
		// GH[11134]
		if apiObject.SubnetId != nil {
			tfMap[names.AttrSubnetID] = aws.ToString(apiObject.SubnetId)
		}
		if apiObject.GatewayId != nil {
			tfMap["gateway_id"] = aws.ToString(apiObject.GatewayId)
		}
		tfMap["main"] = aws.ToBool(apiObject.Main)
		tfList = append(tfList, tfMap)
	}
	return tfList
}

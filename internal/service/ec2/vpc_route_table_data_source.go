// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route_table", name="Route Table")
// @Testing(tagsTest=true)
func dataSourceRouteTable() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRouteTableRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
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
		},
	}
}

func dataSourceRouteTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeRouteTablesInput{}
	vpcId, vpcIdOk := d.GetOk(names.AttrVPCID)
	subnetId, subnetIdOk := d.GetOk(names.AttrSubnetID)
	gatewayId, gatewayIdOk := d.GetOk("gateway_id")
	rtbId, rtbOk := d.GetOk("route_table_id")
	tags, tagsOk := d.GetOk(names.AttrTags)
	filter, filterOk := d.GetOk(names.AttrFilter)

	if !rtbOk && !vpcIdOk && !subnetIdOk && !gatewayIdOk && !filterOk && !tagsOk {
		return sdkdiag.AppendErrorf(diags, "one of route_table_id, vpc_id, subnet_id, gateway_id, filters, or tags must be assigned")
	}
	req.Filters = newAttributeFilterList(
		map[string]string{
			"route-table-id":         rtbId.(string),
			"vpc-id":                 vpcId.(string),
			"association.subnet-id":  subnetId.(string),
			"association.gateway-id": gatewayId.(string),
		},
	)
	req.Filters = append(req.Filters, newTagFilterList(
		Tags(tftags.New(ctx, tags.(map[string]interface{}))),
	)...)
	req.Filters = append(req.Filters, newCustomFilterList(
		filter.(*schema.Set),
	)...)

	resp, err := conn.DescribeRouteTables(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Route Table: %s", err)
	}
	if resp == nil || len(resp.RouteTables) == 0 {
		return sdkdiag.AppendErrorf(diags, "query returned no results. Please change your search criteria and try again")
	}
	if len(resp.RouteTables) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple Route Tables matched; use additional constraints to reduce matches to a single Route Table")
	}

	rt := resp.RouteTables[0]

	d.SetId(aws.ToString(rt.RouteTableId))

	ownerID := aws.ToString(rt.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("route-table/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrOwnerID, ownerID)

	d.Set("route_table_id", rt.RouteTableId)
	d.Set(names.AttrVPCID, rt.VpcId)

	//Ignore the AmazonFSx service tag in addition to standard ignores
	if err := d.Set(names.AttrTags, keyValueTags(ctx, rt.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Ignore(tftags.New(ctx, []string{"AmazonFSx"})).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("routes", dataSourceRoutesRead(ctx, conn, rt.Routes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Route Table: %s", err)
	}

	if err := d.Set("associations", dataSourceAssociationsRead(rt.Associations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Route Table: %s", err)
	}

	return diags
}

func dataSourceRoutesRead(ctx context.Context, conn *ec2.Client, ec2Routes []awstypes.Route) []map[string]interface{} {
	routes := make([]map[string]interface{}, 0, len(ec2Routes))
	// Loop through the routes and add them to the set
	for _, r := range ec2Routes {
		if gatewayID := aws.ToString(r.GatewayId); gatewayID == gatewayIDLocal || gatewayID == gatewayIDVPCLattice {
			continue
		}

		if r.Origin == awstypes.RouteOriginEnableVgwRoutePropagation {
			continue
		}

		if r.DestinationPrefixListId != nil && strings.HasPrefix(aws.ToString(r.GatewayId), "vpce-") {
			// Skipping because VPC endpoint routes are handled separately
			// See aws_vpc_endpoint
			continue
		}

		// Skip cross-account ENIs for AWS services.
		if networkInterfaceID := aws.ToString(r.NetworkInterfaceId); networkInterfaceID != "" {
			networkInterface, err := findNetworkInterfaceByID(ctx, conn, networkInterfaceID)

			if err == nil && networkInterface.Attachment != nil {
				if ownerID, instanceOwnerID := aws.ToString(networkInterface.OwnerId), aws.ToString(networkInterface.Attachment.InstanceOwnerId); ownerID != "" && instanceOwnerID != ownerID {
					log.Printf("[DEBUG] Skip cross-account ENI (%s)", networkInterfaceID)
					continue
				}
			}
		}

		m := make(map[string]interface{})

		if r.DestinationCidrBlock != nil {
			m[names.AttrCIDRBlock] = aws.ToString(r.DestinationCidrBlock)
		}
		if r.DestinationIpv6CidrBlock != nil {
			m["ipv6_cidr_block"] = aws.ToString(r.DestinationIpv6CidrBlock)
		}
		if r.DestinationPrefixListId != nil {
			m["destination_prefix_list_id"] = aws.ToString(r.DestinationPrefixListId)
		}
		if r.CarrierGatewayId != nil {
			m["carrier_gateway_id"] = aws.ToString(r.CarrierGatewayId)
		}
		if r.CoreNetworkArn != nil {
			m["core_network_arn"] = aws.ToString(r.CoreNetworkArn)
		}
		if r.EgressOnlyInternetGatewayId != nil {
			m["egress_only_gateway_id"] = aws.ToString(r.EgressOnlyInternetGatewayId)
		}
		if r.GatewayId != nil {
			if strings.HasPrefix(*r.GatewayId, "vpce-") {
				m[names.AttrVPCEndpointID] = aws.ToString(r.GatewayId)
			} else {
				m["gateway_id"] = aws.ToString(r.GatewayId)
			}
		}
		if r.NatGatewayId != nil {
			m["nat_gateway_id"] = aws.ToString(r.NatGatewayId)
		}
		if r.LocalGatewayId != nil {
			m["local_gateway_id"] = aws.ToString(r.LocalGatewayId)
		}
		if r.InstanceId != nil {
			m[names.AttrInstanceID] = aws.ToString(r.InstanceId)
		}
		if r.TransitGatewayId != nil {
			m[names.AttrTransitGatewayID] = aws.ToString(r.TransitGatewayId)
		}
		if r.VpcPeeringConnectionId != nil {
			m["vpc_peering_connection_id"] = aws.ToString(r.VpcPeeringConnectionId)
		}
		if r.NetworkInterfaceId != nil {
			m[names.AttrNetworkInterfaceID] = aws.ToString(r.NetworkInterfaceId)
		}

		routes = append(routes, m)
	}
	return routes
}

func dataSourceAssociationsRead(ec2Assocations []awstypes.RouteTableAssociation) []map[string]interface{} {
	associations := make([]map[string]interface{}, 0, len(ec2Assocations))
	// Loop through the routes and add them to the set
	for _, a := range ec2Assocations {
		m := make(map[string]interface{})
		m["route_table_id"] = aws.ToString(a.RouteTableId)
		m["route_table_association_id"] = aws.ToString(a.RouteTableAssociationId)
		// GH[11134]
		if a.SubnetId != nil {
			m[names.AttrSubnetID] = aws.ToString(a.SubnetId)
		}
		if a.GatewayId != nil {
			m["gateway_id"] = aws.ToString(a.GatewayId)
		}
		m["main"] = aws.ToBool(a.Main)
		associations = append(associations, m)
	}
	return associations
}

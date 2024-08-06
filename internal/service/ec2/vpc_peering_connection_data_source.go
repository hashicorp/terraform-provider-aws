// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpc_peering_connection", name="VPC Peering Connection")
func dataSourceVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPCPeeringConnectionRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"accepter": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeBool},
			},
			names.AttrCIDRBlock: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cidr_block_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCIDRBlock: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ipv6_cidr_block_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipv6_cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_cidr_block_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCIDRBlock: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"peer_ipv6_cidr_block_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipv6_cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"peer_owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrRegion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"requester": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeBool},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeVpcPeeringConnectionsInput{}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.VpcPeeringConnectionIds = []string{v.(string)}
	}

	input.Filters = newAttributeFilterList(
		map[string]string{
			"status-code":                   d.Get(names.AttrStatus).(string),
			"requester-vpc-info.vpc-id":     d.Get(names.AttrVPCID).(string),
			"requester-vpc-info.owner-id":   d.Get(names.AttrOwnerID).(string),
			"requester-vpc-info.cidr-block": d.Get(names.AttrCIDRBlock).(string),
			"accepter-vpc-info.vpc-id":      d.Get("peer_vpc_id").(string),
			"accepter-vpc-info.owner-id":    d.Get("peer_owner_id").(string),
			"accepter-vpc-info.cidr-block":  d.Get("peer_cidr_block").(string),
		},
	)

	if tags, tagsOk := d.GetOk(names.AttrTags); tagsOk {
		input.Filters = append(input.Filters, newTagFilterList(
			Tags(tftags.New(ctx, tags.(map[string]interface{}))),
		)...)
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	vpcPeeringConnection, err := findVPCPeeringConnection(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 VPC Peering Connection", err))
	}

	d.SetId(aws.ToString(vpcPeeringConnection.VpcPeeringConnectionId))
	d.Set(names.AttrStatus, vpcPeeringConnection.Status.Code)
	d.Set(names.AttrVPCID, vpcPeeringConnection.RequesterVpcInfo.VpcId)
	d.Set(names.AttrOwnerID, vpcPeeringConnection.RequesterVpcInfo.OwnerId)
	d.Set(names.AttrCIDRBlock, vpcPeeringConnection.RequesterVpcInfo.CidrBlock)

	cidrBlockSet := []interface{}{}
	for _, v := range vpcPeeringConnection.RequesterVpcInfo.CidrBlockSet {
		cidrBlockSet = append(cidrBlockSet, map[string]interface{}{
			names.AttrCIDRBlock: aws.ToString(v.CidrBlock),
		})
	}
	if err := d.Set("cidr_block_set", cidrBlockSet); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cidr_block_set: %s", err)
	}

	ipv6CidrBlockSet := []interface{}{}
	for _, v := range vpcPeeringConnection.RequesterVpcInfo.Ipv6CidrBlockSet {
		ipv6CidrBlockSet = append(ipv6CidrBlockSet, map[string]interface{}{
			"ipv6_cidr_block": aws.ToString(v.Ipv6CidrBlock),
		})
	}
	if err := d.Set("ipv6_cidr_block_set", ipv6CidrBlockSet); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ipv6_cidr_block_set: %s", err)
	}

	d.Set(names.AttrRegion, vpcPeeringConnection.RequesterVpcInfo.Region)
	d.Set("peer_vpc_id", vpcPeeringConnection.AccepterVpcInfo.VpcId)
	d.Set("peer_owner_id", vpcPeeringConnection.AccepterVpcInfo.OwnerId)
	d.Set("peer_cidr_block", vpcPeeringConnection.AccepterVpcInfo.CidrBlock)

	peerCidrBlockSet := []interface{}{}
	for _, v := range vpcPeeringConnection.AccepterVpcInfo.CidrBlockSet {
		peerCidrBlockSet = append(peerCidrBlockSet, map[string]interface{}{
			names.AttrCIDRBlock: aws.ToString(v.CidrBlock),
		})
	}
	if err := d.Set("peer_cidr_block_set", peerCidrBlockSet); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting peer_cidr_block_set: %s", err)
	}

	peerIpv6CidrBlockSet := []interface{}{}
	for _, v := range vpcPeeringConnection.AccepterVpcInfo.Ipv6CidrBlockSet {
		peerIpv6CidrBlockSet = append(peerIpv6CidrBlockSet, map[string]interface{}{
			"ipv6_cidr_block": aws.ToString(v.Ipv6CidrBlock),
		})
	}
	if err := d.Set("peer_ipv6_cidr_block_set", peerIpv6CidrBlockSet); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting peer_ipv6_cidr_block_set: %s", err)
	}

	d.Set("peer_region", vpcPeeringConnection.AccepterVpcInfo.Region)

	if err := d.Set(names.AttrTags, keyValueTags(ctx, vpcPeeringConnection.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if vpcPeeringConnection.AccepterVpcInfo.PeeringOptions != nil {
		if err := d.Set("accepter", flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.AccepterVpcInfo.PeeringOptions)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting accepter: %s", err)
		}
	}

	if vpcPeeringConnection.RequesterVpcInfo.PeeringOptions != nil {
		if err := d.Set("requester", flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.RequesterVpcInfo.PeeringOptions)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting requester: %s", err)
		}
	}

	return diags
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpc_peering_connection", name="VPC Peering Connection")
// @Tags
// @Testing(tagsTest=false)
// @Region(overrideEnabled=false)
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
				Computed: true,
			},
			"peer_vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrRegion: {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "region is deprecated. Use requester_region instead.",
			},
			"requester": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeBool},
			},
			"requester_region": {
				Type:     schema.TypeString,
				Computed: true,
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

func dataSourceVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := ec2.DescribeVpcPeeringConnectionsInput{}

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
			svcTags(tftags.New(ctx, tags.(map[string]any))),
		)...)
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	vpcPeeringConnection, err := findVPCPeeringConnection(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 VPC Peering Connection", err))
	}

	accepterVPCInfo, requesterVPCInfo := vpcPeeringConnection.AccepterVpcInfo, vpcPeeringConnection.RequesterVpcInfo

	d.SetId(aws.ToString(vpcPeeringConnection.VpcPeeringConnectionId))
	if accepterVPCInfo.PeeringOptions != nil {
		if err := d.Set("accepter", flattenVPCPeeringConnectionOptionsDescription(accepterVPCInfo.PeeringOptions)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting accepter: %s", err)
		}
	}
	d.Set(names.AttrCIDRBlock, requesterVPCInfo.CidrBlock)
	if err := d.Set("cidr_block_set", tfslices.ApplyToAll(requesterVPCInfo.CidrBlockSet, func(v awstypes.CidrBlock) any {
		return map[string]any{
			names.AttrCIDRBlock: aws.ToString(v.CidrBlock),
		}
	})); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cidr_block_set: %s", err)
	}
	if err := d.Set("ipv6_cidr_block_set", tfslices.ApplyToAll(requesterVPCInfo.Ipv6CidrBlockSet, func(v awstypes.Ipv6CidrBlock) any {
		return map[string]any{
			"ipv6_cidr_block": aws.ToString(v.Ipv6CidrBlock),
		}
	})); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ipv6_cidr_block_set: %s", err)
	}
	d.Set(names.AttrOwnerID, requesterVPCInfo.OwnerId)
	d.Set("peer_cidr_block", accepterVPCInfo.CidrBlock)
	if err := d.Set("peer_cidr_block_set", tfslices.ApplyToAll(accepterVPCInfo.CidrBlockSet, func(v awstypes.CidrBlock) any {
		return map[string]any{
			names.AttrCIDRBlock: aws.ToString(v.CidrBlock),
		}
	})); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting peer_cidr_block_set: %s", err)
	}
	if err := d.Set("peer_ipv6_cidr_block_set", tfslices.ApplyToAll(accepterVPCInfo.Ipv6CidrBlockSet, func(v awstypes.Ipv6CidrBlock) any {
		return map[string]any{
			"ipv6_cidr_block": aws.ToString(v.Ipv6CidrBlock),
		}
	})); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting peer_ipv6_cidr_block_set: %s", err)
	}
	d.Set("peer_owner_id", accepterVPCInfo.OwnerId)
	d.Set("peer_region", accepterVPCInfo.Region)
	d.Set("peer_vpc_id", accepterVPCInfo.VpcId)
	d.Set(names.AttrRegion, requesterVPCInfo.Region)
	if requesterVPCInfo.PeeringOptions != nil {
		if err := d.Set("requester", flattenVPCPeeringConnectionOptionsDescription(requesterVPCInfo.PeeringOptions)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting requester: %s", err)
		}
	}
	d.Set("requester_region", requesterVPCInfo.Region)
	d.Set(names.AttrStatus, vpcPeeringConnection.Status.Code)
	d.Set(names.AttrVPCID, requesterVPCInfo.VpcId)

	setTagsOut(ctx, vpcPeeringConnection.Tags)

	return diags
}

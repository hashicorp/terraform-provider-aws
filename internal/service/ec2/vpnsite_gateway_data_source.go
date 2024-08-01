// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strconv"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpn_gateway", name="VPN Gateway")
// @Tags
// @Testing(tagsTest=false)
func dataSourceVPNGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPNGatewayRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attached_vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceVPNGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVpnGatewaysInput{}

	if id, ok := d.GetOk(names.AttrID); ok {
		input.VpnGatewayIds = []string{id.(string)}
	}

	input.Filters = newAttributeFilterList(
		map[string]string{
			names.AttrState:     d.Get(names.AttrState).(string),
			"availability-zone": d.Get(names.AttrAvailabilityZone).(string),
		},
	)
	if asn, ok := d.GetOk("amazon_side_asn"); ok {
		input.Filters = append(input.Filters, newAttributeFilterList(
			map[string]string{
				"amazon-side-asn": asn.(string),
			},
		)...)
	}
	if id, ok := d.GetOk("attached_vpc_id"); ok {
		input.Filters = append(input.Filters, newAttributeFilterList(
			map[string]string{
				"attachment.state":  "attached",
				"attachment.vpc-id": id.(string),
			},
		)...)
	}
	input.Filters = append(input.Filters, newTagFilterList(
		Tags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{}))),
	)...)
	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)
	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	vgw, err := findVPNGateway(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 VPN Gateway", err))
	}

	d.SetId(aws.ToString(vgw.VpnGatewayId))

	d.Set("amazon_side_asn", strconv.FormatInt(aws.ToInt64(vgw.AmazonSideAsn), 10))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpn-gateway/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	for _, attachment := range vgw.VpcAttachments {
		if attachment.State == awstypes.AttachmentStatusAttached {
			d.Set("attached_vpc_id", attachment.VpcId)
			break
		}
	}
	d.Set(names.AttrAvailabilityZone, vgw.AvailabilityZone)
	d.Set(names.AttrState, vgw.State)

	setTagsOut(ctx, vgw.Tags)

	return diags
}

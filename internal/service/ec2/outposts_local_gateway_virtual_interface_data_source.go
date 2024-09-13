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

// @SDKDataSource("aws_ec2_local_gateway_virtual_interface", name="Local Gateway Virtual Interface")
// @Tags
// @Testing(tagsTest=false)
func dataSourceLocalGatewayVirtualInterface() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalGatewayVirtualInterfaceRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"local_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_bgp_asn": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"local_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_gateway_virtual_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"peer_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_bgp_asn": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"vlan": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceLocalGatewayVirtualInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeLocalGatewayVirtualInterfacesInput{}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.LocalGatewayVirtualInterfaceIds = []string{v.(string)}
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

	localGatewayVirtualInterface, err := findLocalGatewayVirtualInterface(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Local Gateway Virtual Interface", err))
	}

	d.SetId(aws.ToString(localGatewayVirtualInterface.LocalGatewayVirtualInterfaceId))
	d.Set("local_address", localGatewayVirtualInterface.LocalAddress)
	d.Set("local_bgp_asn", localGatewayVirtualInterface.LocalBgpAsn)
	d.Set("local_gateway_id", localGatewayVirtualInterface.LocalGatewayId)
	d.Set("peer_address", localGatewayVirtualInterface.PeerAddress)
	d.Set("peer_bgp_asn", localGatewayVirtualInterface.PeerBgpAsn)
	d.Set("vlan", localGatewayVirtualInterface.Vlan)

	setTagsOut(ctx, localGatewayVirtualInterface.Tags)

	return diags
}

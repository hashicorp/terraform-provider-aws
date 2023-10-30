// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_ec2_local_gateway_virtual_interface")
func DataSourceLocalGatewayVirtualInterface() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalGatewayVirtualInterfaceRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": CustomFiltersSchema(),
			"id": {
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
			"tags": tftags.TagsSchemaComputed(),
			"vlan": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceLocalGatewayVirtualInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeLocalGatewayVirtualInterfacesInput{}

	if v, ok := d.GetOk("id"); ok {
		input.LocalGatewayVirtualInterfaceIds = []*string{aws.String(v.(string))}
	}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(ctx, d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	output, err := conn.DescribeLocalGatewayVirtualInterfacesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing EC2 Local Gateway Virtual Interfaces: %s", err)
	}

	if output == nil || len(output.LocalGatewayVirtualInterfaces) == 0 {
		return sdkdiag.AppendErrorf(diags, "no matching EC2 Local Gateway Virtual Interface found")
	}

	if len(output.LocalGatewayVirtualInterfaces) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple EC2 Local Gateway Virtual Interfaces matched; use additional constraints to reduce matches to a single EC2 Local Gateway Virtual Interface")
	}

	localGatewayVirtualInterface := output.LocalGatewayVirtualInterfaces[0]

	d.SetId(aws.StringValue(localGatewayVirtualInterface.LocalGatewayVirtualInterfaceId))
	d.Set("local_address", localGatewayVirtualInterface.LocalAddress)
	d.Set("local_bgp_asn", localGatewayVirtualInterface.LocalBgpAsn)
	d.Set("local_gateway_id", localGatewayVirtualInterface.LocalGatewayId)
	d.Set("peer_address", localGatewayVirtualInterface.PeerAddress)
	d.Set("peer_bgp_asn", localGatewayVirtualInterface.PeerBgpAsn)

	if err := d.Set("tags", KeyValueTags(ctx, localGatewayVirtualInterface.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.Set("vlan", localGatewayVirtualInterface.Vlan)

	return diags
}

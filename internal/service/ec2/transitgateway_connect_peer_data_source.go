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
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_transit_gateway_connect_peer", name="Transit Gateway Connect Peer")
// @Tags
// @Testing(tagsTest=false)
func dataSourceTransitGatewayConnectPeer() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayConnectPeerRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_asn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_peer_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_transit_gateway_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: customFiltersSchema(),
			"inside_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"peer_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"transit_gateway_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTransitGatewayAttachmentID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_connect_peer_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceTransitGatewayConnectPeerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeTransitGatewayConnectPeersInput{}

	if v, ok := d.GetOk("transit_gateway_connect_peer_id"); ok {
		input.TransitGatewayConnectPeerIds = []string{v.(string)}
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	transitGatewayConnectPeer, err := findTransitGatewayConnectPeer(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Transit Gateway Connect Peer", err))
	}

	d.SetId(aws.ToString(transitGatewayConnectPeer.TransitGatewayConnectPeerId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("transit-gateway-connect-peer/%s", d.Id()),
	}.String()
	bgpConfigurations := transitGatewayConnectPeer.ConnectPeerConfiguration.BgpConfigurations
	d.Set(names.AttrARN, arn)
	d.Set("bgp_asn", strconv.FormatInt(aws.ToInt64(bgpConfigurations[0].PeerAsn), 10))
	d.Set("bgp_peer_address", bgpConfigurations[0].PeerAddress)
	d.Set("bgp_transit_gateway_addresses", slices.ApplyToAll(bgpConfigurations, func(v awstypes.TransitGatewayAttachmentBgpConfiguration) string {
		return aws.ToString(v.TransitGatewayAddress)
	}))
	d.Set("inside_cidr_blocks", transitGatewayConnectPeer.ConnectPeerConfiguration.InsideCidrBlocks)
	d.Set("peer_address", transitGatewayConnectPeer.ConnectPeerConfiguration.PeerAddress)
	d.Set("transit_gateway_address", transitGatewayConnectPeer.ConnectPeerConfiguration.TransitGatewayAddress)
	d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayConnectPeer.TransitGatewayAttachmentId)
	d.Set("transit_gateway_connect_peer_id", transitGatewayConnectPeer.TransitGatewayConnectPeerId)

	setTagsOut(ctx, transitGatewayConnectPeer.Tags)

	return diags
}

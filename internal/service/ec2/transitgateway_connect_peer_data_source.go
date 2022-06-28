package ec2

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceTransitGatewayConnectPeer() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayConnectPeerRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_asn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"inside_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"peer_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"transit_gateway_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_attachment_id": {
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
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayConnectPeersInput{}

	if v, ok := d.GetOk("transit_gateway_connect_peer_id"); ok {
		input.TransitGatewayConnectPeerIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	transitGatewayConnectPeer, err := FindTransitGatewayConnectPeer(conn, input)

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("EC2 Transit Gateway Connect Peer", err))
	}

	d.SetId(aws.StringValue(transitGatewayConnectPeer.TransitGatewayConnectPeerId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("transit-gateway-connect-peer/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("bgp_asn", strconv.FormatInt(aws.Int64Value(transitGatewayConnectPeer.ConnectPeerConfiguration.BgpConfigurations[0].PeerAsn), 10))
	d.Set("inside_cidr_blocks", aws.StringValueSlice(transitGatewayConnectPeer.ConnectPeerConfiguration.InsideCidrBlocks))
	d.Set("peer_address", transitGatewayConnectPeer.ConnectPeerConfiguration.PeerAddress)
	d.Set("transit_gateway_address", transitGatewayConnectPeer.ConnectPeerConfiguration.TransitGatewayAddress)
	d.Set("transit_gateway_attachment_id", transitGatewayConnectPeer.TransitGatewayAttachmentId)
	d.Set("transit_gateway_connect_peer_id", transitGatewayConnectPeer.TransitGatewayConnectPeerId)

	if err := d.Set("tags", KeyValueTags(transitGatewayConnectPeer.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	return nil
}

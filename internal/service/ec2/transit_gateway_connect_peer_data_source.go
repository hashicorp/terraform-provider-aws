package ec2

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceTransitGatewayConnectPeer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTransitGatewayConnectPeerRead,

		Schema: map[string]*schema.Schema{
			"bgp_asn": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"inside_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"peer_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"transit_gateway_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTransitGatewayConnectPeerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayConnectPeersInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	if v, ok := d.GetOk("id"); ok {
		input.TransitGatewayConnectPeerIds = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateway Connect Peers: %s", input)
	output, err := conn.DescribeTransitGatewayConnectPeers(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect Peers: %w", err)
	}

	if output == nil || len(output.TransitGatewayConnectPeers) == 0 {
		return errors.New("error reading EC2 Transit Gateway Connect Peers: no results found")
	}

	if len(output.TransitGatewayConnectPeers) > 1 {
		return errors.New("error reading EC2 Transit Gateway Connect Peers: multiple results found, try adjusting search criteria")
	}

	transitGatewayConnectPeer := output.TransitGatewayConnectPeers[0]

	if transitGatewayConnectPeer == nil {
		return errors.New("error reading EC2 Transit Gateway Connect Peers: empty result")
	}

	if transitGatewayConnectPeer.ConnectPeerConfiguration == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect Peers (%s): missing Connect Peer Configuration", d.Id())
	}

	if transitGatewayConnectPeer.ConnectPeerConfiguration.BgpConfigurations == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect Peers (%s): missing BGP configurations", d.Id())
	}

	if err := d.Set("tags", KeyValueTags(transitGatewayConnectPeer.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	tags := KeyValueTags(transitGatewayConnectPeer.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("bgp_asn", transitGatewayConnectPeer.ConnectPeerConfiguration.BgpConfigurations[0].PeerAsn)
	d.Set("inside_cidr_blocks", transitGatewayConnectPeer.ConnectPeerConfiguration.InsideCidrBlocks)
	d.Set("peer_address", transitGatewayConnectPeer.ConnectPeerConfiguration.PeerAddress)
	d.Set("transit_gateway_address", transitGatewayConnectPeer.ConnectPeerConfiguration.TransitGatewayAddress)
	d.Set("transit_gateway_attachment_id", transitGatewayConnectPeer.TransitGatewayAttachmentId)

	d.SetId(aws.StringValue(transitGatewayConnectPeer.TransitGatewayConnectPeerId))

	return nil
}

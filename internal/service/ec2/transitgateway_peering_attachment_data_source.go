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

func DataSourceTransitGatewayPeeringAttachment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTransitGatewayPeeringAttachmentRead,

		Schema: map[string]*schema.Schema{
			"filter": CustomFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTransitGatewayPeeringAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{}

	if v, ok := d.GetOk("id"); ok {
		input.TransitGatewayAttachmentIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = BuildCustomFilterList(d.Get("filter").(*schema.Set))
	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Filters = append(input.Filters, ec2TagFiltersFromMap(v)...)
	}
	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateway Peering Attachments: %s", input)
	output, err := conn.DescribeTransitGatewayPeeringAttachments(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Peering Attachments: %ws", err)
	}

	if output == nil || len(output.TransitGatewayPeeringAttachments) == 0 {
		return errors.New("error reading EC2 Transit Gateway Peering Attachment: no results found")
	}

	if len(output.TransitGatewayPeeringAttachments) > 1 {
		return errors.New("error reading EC2 Transit Gateway Peering Attachment: multiple results found, try adjusting search criteria")
	}

	transitGatewayPeeringAttachment := output.TransitGatewayPeeringAttachments[0]

	if transitGatewayPeeringAttachment == nil {
		return errors.New("error reading EC2 Transit Gateway Peering Attachment: empty result")
	}

	local := transitGatewayPeeringAttachment.RequesterTgwInfo
	peer := transitGatewayPeeringAttachment.AccepterTgwInfo

	if aws.StringValue(transitGatewayPeeringAttachment.AccepterTgwInfo.OwnerId) == meta.(*conns.AWSClient).AccountID && aws.StringValue(transitGatewayPeeringAttachment.AccepterTgwInfo.Region) == meta.(*conns.AWSClient).Region {
		local = transitGatewayPeeringAttachment.AccepterTgwInfo
		peer = transitGatewayPeeringAttachment.RequesterTgwInfo
	}

	d.Set("peer_account_id", peer.OwnerId)
	d.Set("peer_region", peer.Region)
	d.Set("peer_transit_gateway_id", peer.TransitGatewayId)
	d.Set("transit_gateway_id", local.TransitGatewayId)

	if err := d.Set("tags", KeyValueTags(transitGatewayPeeringAttachment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.SetId(aws.StringValue(transitGatewayPeeringAttachment.TransitGatewayAttachmentId))

	return nil
}

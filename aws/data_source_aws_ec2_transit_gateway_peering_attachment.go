package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2TransitGatewayPeeringAttachment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2TransitGatewayPeeringAttachmentRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
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
			"tags": tagsSchemaComputed(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEc2TransitGatewayPeeringAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("id"); ok {
		input.TransitGatewayAttachmentIds = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateways: %s", input)
	output, err := conn.DescribeTransitGatewayPeeringAttachments(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Route Table: %s", err)
	}

	if output == nil || len(output.TransitGatewayPeeringAttachments) == 0 {
		return errors.New("error reading EC2 Transit Gateway Route Table: no results found")
	}

	if len(output.TransitGatewayPeeringAttachments) > 1 {
		return errors.New("error reading EC2 Transit Gateway Route Table: multiple results found, try adjusting search criteria")
	}

	transitGatewayPeeringAttachment := output.TransitGatewayPeeringAttachments[0]

	if transitGatewayPeeringAttachment == nil {
		return errors.New("error reading EC2 Transit Gateway Route Table: empty result")
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(transitGatewayPeeringAttachment.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("peer_account_id", aws.StringValue(transitGatewayPeeringAttachment.AccepterTgwInfo.OwnerId))
	d.Set("peer_region", aws.StringValue(transitGatewayPeeringAttachment.AccepterTgwInfo.Region))
	d.Set("peer_transit_gateway_id", aws.StringValue(transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId))
	d.Set("transit_gateway_id", aws.StringValue(transitGatewayPeeringAttachment.RequesterTgwInfo.TransitGatewayId))

	d.SetId(aws.StringValue(transitGatewayPeeringAttachment.TransitGatewayAttachmentId))

	return nil
}

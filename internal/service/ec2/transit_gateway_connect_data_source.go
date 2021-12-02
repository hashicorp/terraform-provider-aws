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

func DataSourceTransitGatewayConnect() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTransitGatewayConnectRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transport_attachment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTransitGatewayConnectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayConnectsInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	if v, ok := d.GetOk("id"); ok {
		input.TransitGatewayAttachmentIds = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateways: %s", input)
	output, err := conn.DescribeTransitGatewayConnects(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Route Table: %w", err)
	}

	if output == nil || len(output.TransitGatewayConnects) == 0 {
		return errors.New("error reading EC2 Transit Gateway Route Table: no results found")
	}

	if len(output.TransitGatewayConnects) > 1 {
		return errors.New("error reading EC2 Transit Gateway Route Table: multiple results found, try adjusting search criteria")
	}

	transitGatewayConnect := output.TransitGatewayConnects[0]

	if transitGatewayConnect == nil {
		return errors.New("error reading EC2 Transit Gateway Route Table: empty result")
	}

	if transitGatewayConnect.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect Attachment (%s): missing options", d.Id())
	}

	if err := d.Set("tags", KeyValueTags(transitGatewayConnect.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("transit_gateway_id", transitGatewayConnect.TransitGatewayId)
	d.Set("transport_attachment_id", transitGatewayConnect.TransportTransitGatewayAttachmentId)

	d.SetId(aws.StringValue(transitGatewayConnect.TransitGatewayAttachmentId))

	return nil
}

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

func DataSourceTransitGatewayDxGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTransitGatewayDxGatewayAttachmentRead,

		Schema: map[string]*schema.Schema{
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filter": DataSourceFiltersSchema(),
		},
	}
}

func dataSourceTransitGatewayDxGatewayAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	filters, filtersOk := d.GetOk("filter")
	tags, tagsOk := d.GetOk("tags")
	dxGatewayId, dxGatewayIdOk := d.GetOk("dx_gateway_id")
	transitGatewayId, transitGatewayIdOk := d.GetOk("transit_gateway_id")

	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("resource-type"),
				Values: []*string{aws.String(ec2.TransitGatewayAttachmentResourceTypeDirectConnectGateway)},
			},
		},
	}
	if filtersOk {
		input.Filters = append(input.Filters, BuildFiltersDataSource(filters.(*schema.Set))...)
	}
	if tagsOk {
		input.Filters = append(input.Filters, ec2TagFiltersFromMap(tags.(map[string]interface{}))...)
	}
	// to preserve original functionality
	if dxGatewayIdOk {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name:   aws.String("resource-id"),
			Values: []*string{aws.String(dxGatewayId.(string))},
		})
	}

	if transitGatewayIdOk {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name:   aws.String("transit-gateway-id"),
			Values: []*string{aws.String(transitGatewayId.(string))},
		})
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateway Direct Connect Gateway Attachments: %s", input)
	output, err := conn.DescribeTransitGatewayAttachments(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Direct Connect Gateway Attachment: %w", err)
	}

	if output == nil || len(output.TransitGatewayAttachments) == 0 || output.TransitGatewayAttachments[0] == nil {
		return errors.New("error reading EC2 Transit Gateway Direct Connect Gateway Attachment: no results found")
	}

	if len(output.TransitGatewayAttachments) > 1 {
		return errors.New("error reading EC2 Transit Gateway Direct Connect Gateway Attachment: multiple results found, try adjusting search criteria")
	}

	transitGatewayAttachment := output.TransitGatewayAttachments[0]

	if err := d.Set("tags", KeyValueTags(transitGatewayAttachment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("transit_gateway_id", transitGatewayAttachment.TransitGatewayId)
	d.Set("dx_gateway_id", transitGatewayAttachment.ResourceId)

	d.SetId(aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId))

	return nil
}

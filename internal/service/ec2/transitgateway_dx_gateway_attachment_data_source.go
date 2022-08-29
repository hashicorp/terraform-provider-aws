package ec2

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceTransitGatewayDxGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTransitGatewayDxGatewayAttachmentRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filter": CustomFiltersSchema(),
			"tags":   tftags.TagsSchemaComputed(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceTransitGatewayDxGatewayAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"resource-type": ec2.TransitGatewayAttachmentResourceTypeDirectConnectGateway,
		}),
	}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if v, ok := d.GetOk("tags"); ok {
		input.Filters = append(input.Filters, BuildTagFilterList(
			Tags(tftags.New(v.(map[string]interface{}))),
		)...)
	}

	// to preserve original functionality
	if v, ok := d.GetOk("dx_gateway_id"); ok {
		input.Filters = append(input.Filters, BuildAttributeFilterList(map[string]string{
			"resource-id": v.(string),
		})...)
	}

	if v, ok := d.GetOk("transit_gateway_id"); ok {
		input.Filters = append(input.Filters, BuildAttributeFilterList(map[string]string{
			"transit-gateway-id": v.(string),
		})...)
	}

	transitGatewayAttachment, err := FindTransitGatewayAttachment(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 Transit Gateway Direct Connect Gateway Attachment", err)
	}

	d.SetId(aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId))
	d.Set("dx_gateway_id", transitGatewayAttachment.ResourceId)
	d.Set("transit_gateway_id", transitGatewayAttachment.TransitGatewayId)

	if err := d.Set("tags", KeyValueTags(transitGatewayAttachment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	return nil
}

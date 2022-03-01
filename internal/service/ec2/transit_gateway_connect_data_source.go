package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceTransitGatewayConnect() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayConnectRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"transit_gateway_connect_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
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

func dataSourceTransitGatewayConnectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayConnectsInput{}

	if v, ok := d.GetOk("transit_gateway_connect_id"); ok {
		input.TransitGatewayAttachmentIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	transitGatewayConnect, err := FindTransitGatewayConnect(conn, input)

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("EC2 Transit Gateway Connect", err))
	}

	d.SetId(aws.StringValue(transitGatewayConnect.TransitGatewayAttachmentId))
	d.Set("protocol", transitGatewayConnect.Options.Protocol)
	d.Set("transit_gateway_connect_id", transitGatewayConnect.TransitGatewayAttachmentId)
	d.Set("transit_gateway_id", transitGatewayConnect.TransitGatewayId)
	d.Set("transport_attachment_id", transitGatewayConnect.TransportTransitGatewayAttachmentId)

	if err := d.Set("tags", KeyValueTags(transitGatewayConnect.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	return nil
}

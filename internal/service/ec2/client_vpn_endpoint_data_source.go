package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceClientVPNEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceClientVPNEndpointRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"tags":   tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceClientVPNEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeClientVpnEndpointsInput{}

	if v, ok := d.GetOk("client_vpn_endpoint_id"); ok {
		input.ClientVpnEndpointIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	ep, err := FindClientVPNEndpoint(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 Client VPN Endpoint", err)
	}

	d.SetId(aws.StringValue(ep.ClientVpnEndpointId))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("client-vpn-endpoint/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("client_vpn_endpoint_id", ep.ClientVpnEndpointId)

	if err := d.Set("tags", KeyValueTags(ep.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

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

func DataSourceNetworkInsightsPath() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceNetworkInsightsPathRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"filter": CustomFiltersSchema(),
			"network_insights_path_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceNetworkInsightsPathRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeNetworkInsightsPathsInput{}

	if v, ok := d.GetOk("network_insights_path_id"); ok {
		input.NetworkInsightsPathIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	nip, err := FindNetworkInsightsPath(ctx, conn, input)

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("EC2 Network Insights Path", err))
	}

	networkInsightsPathID := aws.StringValue(nip.NetworkInsightsPathId)
	d.SetId(networkInsightsPathID)
	d.Set("arn", nip.NetworkInsightsPathArn)
	d.Set("destination", nip.Destination)
	d.Set("destination_ip", nip.DestinationIp)
	d.Set("destination_port", nip.DestinationPort)
	d.Set("network_insights_path_id", networkInsightsPathID)
	d.Set("protocol", nip.Protocol)
	d.Set("source", nip.Source)
	d.Set("source_ip", nip.SourceIp)

	if err := d.Set("tags", KeyValueTags(nip.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}

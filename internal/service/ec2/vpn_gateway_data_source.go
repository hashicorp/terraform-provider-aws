package ec2

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceVPNGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVPNGatewayRead,

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attached_vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"filter": CustomFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceVPNGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeVpnGatewaysInput{}

	if id, ok := d.GetOk("id"); ok {
		input.VpnGatewayIds = aws.StringSlice([]string{id.(string)})
	}

	input.Filters = BuildAttributeFilterList(
		map[string]string{
			"state":             d.Get("state").(string),
			"availability-zone": d.Get("availability_zone").(string),
		},
	)
	if asn, ok := d.GetOk("amazon_side_asn"); ok {
		input.Filters = append(input.Filters, BuildAttributeFilterList(
			map[string]string{
				"amazon-side-asn": asn.(string),
			},
		)...)
	}
	if id, ok := d.GetOk("attached_vpc_id"); ok {
		input.Filters = append(input.Filters, BuildAttributeFilterList(
			map[string]string{
				"attachment.state":  "attached",
				"attachment.vpc-id": id.(string),
			},
		)...)
	}
	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)
	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	vgw, err := FindVPNGateway(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 VPN Gateway", err)
	}

	d.SetId(aws.StringValue(vgw.VpnGatewayId))

	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(vgw.AmazonSideAsn), 10))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpn-gateway/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	for _, attachment := range vgw.VpcAttachments {
		if aws.StringValue(attachment.State) == ec2.AttachmentStatusAttached {
			d.Set("attached_vpc_id", attachment.VpcId)
			break
		}
	}
	d.Set("availability_zone", vgw.AvailabilityZone)
	d.Set("state", vgw.State)

	if err := d.Set("tags", KeyValueTags(vgw.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

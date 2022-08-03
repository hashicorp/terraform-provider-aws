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

func DataSourceCustomerGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCustomerGatewayRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_asn": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCustomerGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeCustomerGatewaysInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	if v, ok := d.GetOk("id"); ok {
		input.CustomerGatewayIds = []*string{aws.String(v.(string))}
	}

	cgw, err := FindCustomerGateway(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 Customer Gateway", err)
	}

	d.SetId(aws.StringValue(cgw.CustomerGatewayId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("customer-gateway/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	if v := aws.StringValue(cgw.BgpAsn); v != "" {
		v, err := strconv.ParseInt(v, 0, 0)

		if err != nil {
			return err
		}

		d.Set("bgp_asn", v)
	} else {
		d.Set("bgp_asn", nil)
	}
	d.Set("certificate_arn", cgw.CertificateArn)
	d.Set("device_name", cgw.DeviceName)
	d.Set("ip_address", cgw.IpAddress)
	d.Set("type", cgw.Type)

	if err := d.Set("tags", KeyValueTags(cgw.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

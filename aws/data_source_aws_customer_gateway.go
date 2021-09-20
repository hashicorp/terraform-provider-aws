package aws

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceCustomerGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCustomerGatewayRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"bgp_asn": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"device_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCustomerGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := ec2.DescribeCustomerGatewaysInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("id"); ok {
		input.CustomerGatewayIds = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Reading EC2 Customer Gateways: %s", input)
	output, err := conn.DescribeCustomerGateways(&input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Customer Gateways: %w", err)
	}

	if output == nil || len(output.CustomerGateways) == 0 {
		return errors.New("error reading EC2 Customer Gateways: no results found")
	}

	if len(output.CustomerGateways) > 1 {
		return errors.New("error reading EC2 Customer Gateways: multiple results found, try adjusting search criteria")
	}

	cg := output.CustomerGateways[0]
	if cg == nil {
		return errors.New("error reading EC2 Customer Gateway: empty result")
	}

	d.Set("ip_address", cg.IpAddress)
	d.Set("type", cg.Type)
	d.Set("device_name", cg.DeviceName)
	d.SetId(aws.StringValue(cg.CustomerGatewayId))

	if v := aws.StringValue(cg.BgpAsn); v != "" {
		asn, err := strconv.ParseInt(v, 0, 0)
		if err != nil {
			return fmt.Errorf("error parsing BGP ASN %q: %w", v, err)
		}

		d.Set("bgp_asn", int(asn))
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(cg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags for EC2 Customer Gateway %q: %w", aws.StringValue(cg.CustomerGatewayId), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("customer-gateway/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

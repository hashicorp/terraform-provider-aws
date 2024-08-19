package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsRoute53Zones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRoute53ZonesRead,
		Schema: map[string]*schema.Schema{
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsRoute53ZonesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeVpcsInput{}

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, buildEC2TagFilterList(
			keyvaluetags.New(tags.(map[string]interface{})).Ec2Tags(),
		)...)
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		req.Filters = append(req.Filters, buildEC2CustomFilterList(
			filters.(*schema.Set),
		)...)
	}
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] DescribeVpcs %s\n", req)
	resp, err := conn.DescribeVpcs(req)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.Vpcs) == 0 {
		return fmt.Errorf("no matching VPC found")
	}

	vpcs := make([]string, 0)

	for _, vpc := range resp.Vpcs {
		vpcs = append(vpcs, aws.StringValue(vpc.VpcId))
	}

	d.SetId(meta.(*AWSClient).region)

	if err := d.Set("ids", vpcs); err != nil {
		return fmt.Errorf("Error setting vpc ids: %s", err)
	}

	return nil
}

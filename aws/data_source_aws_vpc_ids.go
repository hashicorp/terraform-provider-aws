package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsVpcIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsVpcIDsRead,
		Schema: map[string]*schema.Schema{

			"tags": tagsSchemaComputed(),

			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsVpcIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeVpcsInput{}

	req.Filters = buildEC2TagFilterList(
		tagsFromMap(d.Get("tags").(map[string]interface{})),
	)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] DescribeVpcs %s\n", req)
	resp, err := conn.DescribeVpcs(req)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] show vpcs %s\n", resp)
	if resp == nil || len(resp.Vpcs) == 0 {
		return fmt.Errorf("no matching VPC found")
	}

	vpcs := make([]string, 0)

	for _, vpc := range resp.Vpcs {
		vpcs = append(vpcs, *vpc.VpcId)
	}

	d.SetId(vpcs[0])
	d.Set("ids", vpcs)

	return nil
}

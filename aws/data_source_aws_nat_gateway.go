package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsNatGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsNatGatewayRead,

		Schema: map[string]*schema.Schema{
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
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"filter": ec2CustomFiltersSchema(),
			"tags":   tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsNatGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Reading VPN Gateways.")

	req := &ec2.DescribeNatGatewaysInput{}

	if id, ok := d.GetOk("id"); ok {
		req.NatGatewayIds = aws.StringSlice([]string{id.(string)})
	}

	req.Filter = buildEC2AttributeFilterList(
		map[string]string{
			"state":     d.Get("state").(string),
			"subnet-id": d.Get("subnet_id").(string),
		},
	)
	if id, ok := d.GetOk("vpc_id"); ok {
		req.Filter = append(req.Filter, buildEC2AttributeFilterList(
			map[string]string{
				"vpc-id": id.(string),
			},
		)...)
	}
	req.Filter = append(req.Filter, buildEC2TagFilterList(
		tagsFromMap(d.Get("tags").(map[string]interface{})),
	)...)
	req.Filter = append(req.Filter, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filter) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filter = nil
	}

	resp, err := conn.DescribeNatGateways(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.NatGateways) == 0 {
		return fmt.Errorf("no matching NAT gateway found: %#v", req)
	}
	if len(resp.NatGateways) > 1 {
		return fmt.Errorf("multiple NAT gateways matched; use additional constraints to reduce matches to a single NAT gateway")
	}

	ngw := resp.NatGateways[0]

	d.SetId(aws.StringValue(ngw.NatGatewayId))
	d.Set("state", ngw.State)
	d.Set("subnet_id", ngw.AvailabilityZone)
	d.Set("tags", tagsToMap(ngw.Tags))

	return nil
}

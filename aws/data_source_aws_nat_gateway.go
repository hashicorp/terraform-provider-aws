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

	req.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"state":             d.Get("state").(string),
			"subnet-id": d.Get("subnet_id").(string),
		},
	)
	if id, ok := d.GetOk("vpc_id"); ok {
		req.Filters = append(req.Filters, buildEC2AttributeFilterList(
			map[string]string{
				"vpc-id": id.(string),
			},
		)...)
	}
	req.Filters = append(req.Filters, buildEC2TagFilterList(
		tagsFromMap(d.Get("tags").(map[string]interface{})),
	)...)
	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	resp, err := conn.DescribeNatGateways(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.NatGateways) == 0 {
		return fmt.Errorf("no matching VPN gateway found: %#v", req)
	}
	if len(resp.NatGateways) > 1 {
		return fmt.Errorf("multiple VPN gateways matched; use additional constraints to reduce matches to a single VPN gateway")
	}

	vgw := resp.NatGateways[0]

	d.SetId(aws.StringValue(vgw.NatGatewayId))
	d.Set("state", vgw.State)
	d.Set("subnet_id", vgw.AvailabilityZone)
	d.Set("tags", tagsToMap(vgw.Tags))

	for _, attachment := range vgw.VpcAttachments {
		if *attachment.State == "attached" {
			d.Set("vpc_id", attachment.VpcId)
			break
		}
	}

	return nil
}

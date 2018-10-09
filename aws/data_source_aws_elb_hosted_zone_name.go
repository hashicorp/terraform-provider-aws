package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsElbHostedZoneName() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElbHostedZoneNameRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"hosted_zone_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsElbHostedZoneNameRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbconn

	elbName := d.Get("name").(string)

	describe, err := elbconn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(elbName)},
	})

	if err != nil {
		return err
	}

	if len(describe.LoadBalancerDescriptions) != 1 ||
		*describe.LoadBalancerDescriptions[0].LoadBalancerName != elbName {
		return fmt.Errorf("ELB not found")
	}

	d.Set("hosted_zone_id", describe.LoadBalancerDescriptions[0].CanonicalHostedZoneNameID)
	d.Set("hosted_zone_name", describe.LoadBalancerDescriptions[0].CanonicalHostedZoneName)
	d.Set("dns_name", describe.LoadBalancerDescriptions[0].DNSName)
	d.SetId(*describe.LoadBalancerDescriptions[0].CanonicalHostedZoneName)

	return nil
}

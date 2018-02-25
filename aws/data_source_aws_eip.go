package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsEip() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEipRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEipRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeAddressesInput{}

	if id, ok := d.GetOk("id"); ok {
		req.AllocationIds = []*string{aws.String(id.(string))}
	}

	if public_ip := d.Get("public_ip"); public_ip != "" {
		req.PublicIps = []*string{aws.String(public_ip.(string))}
	}

	log.Printf("[DEBUG] Reading EIP: %s", req)
	resp, err := conn.DescribeAddresses(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.Addresses) == 0 {
		return fmt.Errorf("no matching Elastic IP found")
	}
	if len(resp.Addresses) > 1 {
		return fmt.Errorf("multiple Elastic IPs matched; use additional constraints to reduce matches to a single Elastic IP")
	}

	eip := resp.Addresses[0]

	if *eip.Domain == "vpc" {
		d.SetId(*eip.AllocationId)
	} else {
		log.Printf("[DEBUG] Reading EIP, has no AllocationId, this means we have a Classic EIP, the id will also be the public ip : %s", req)
		d.SetId(*eip.PublicIp)
	}

	d.Set("public_ip", eip.PublicIp)

	return nil
}

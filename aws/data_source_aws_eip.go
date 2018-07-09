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
			"association_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"filter": ec2CustomFiltersSchema(),
		},
	}
}

func dataSourceAwsEipRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeAddressesInput{}

	if id, idOk := d.GetOk("id"); idOk {
		req.Filters = buildEC2AttributeFilterList(
			map[string]string{
				"allocation-id": id.(string),
			},
		)
	}

	if assocId, assocIdOk := d.GetOk("association_id"); assocIdOk {
		req.Filters = buildEC2AttributeFilterList(
			map[string]string{
				"association-id": assocId.(string),
			},
		)
	}

	if publicIp, publicIpOk := d.GetOk("public_ip"); publicIpOk {
		req.Filters = append(req.Filters, buildEC2AttributeFilterList(
			map[string]string{
				"public-ip": publicIp.(string),
			},
		)...)
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		req.Filters = append(req.Filters, buildEC2CustomFilterList(
			filters.(*schema.Set),
		)...)
	}

	if len(req.Filters) == 0 {
		req.Filters = nil
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

	d.SetId(*eip.AllocationId)
	d.Set("public_ip", eip.PublicIp)

	if eip.AssociationId != nil {
		d.Set("association_id", aws.StringValue(eip.AssociationId))
	}

	return nil
}

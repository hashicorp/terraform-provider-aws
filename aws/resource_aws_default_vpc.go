package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDefaultVpc() *schema.Resource {
	// reuse aws_vpc schema, and methods for READ, UPDATE
	dvpc := resourceAwsVpc()
	dvpc.Create = resourceAwsDefaultVpcCreate
	dvpc.Delete = resourceAwsDefaultVpcDelete

	// Can't "terraform import" a Default VPC; Use "terraform apply"
	dvpc.Importer = nil

	dvpc.CustomizeDiff = resourceAwsDefaultVpcCustomizeDiff

	// cidr_block is a computed value for Default VPCs
	dvpc.Schema["cidr_block"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
	// instance_tenancy is a computed value for Default VPCs
	dvpc.Schema["instance_tenancy"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
	// assign_generated_ipv6_cidr_block is a computed value for Default VPCs
	dvpc.Schema["assign_generated_ipv6_cidr_block"] = &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	}

	return dvpc
}

func resourceAwsDefaultVpcCreate(d *schema.ResourceData, meta interface{}) error {
	vpc, err := resourceAwsDefaultVpcFindVpc(meta)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(vpc.VpcId))
	return resourceAwsVpcUpdate(d, meta)
}

func resourceAwsDefaultVpcDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default VPC. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func resourceAwsDefaultVpcCustomizeDiff(diff *schema.ResourceDiff, meta interface{}) error {
	if diff.Id() == "" {
		// New resource.
		v, ok := diff.GetOkExists("assign_generated_ipv6_cidr_block")
		if ok {
			// assign_generated_ipv6_cidr_block specified.
			newIpv6Flag := v.(bool)

			// See if the Default VPC already has an IPv6 CIDR block assigned.
			vpc, err := resourceAwsDefaultVpcFindVpc(meta)
			if err != nil {
				return err
			}

			oldIpv6Flag := resourceAwsVpcFindIpv6CidrBlockAssociation(vpc) != nil
			log.Printf("[DEBUG] Default VPC IPv6 %v -> %v", oldIpv6Flag, newIpv6Flag)
			if newIpv6Flag == oldIpv6Flag {
				diff.Clear("assign_generated_ipv6_cidr_block")
			} else {
				diff.SetNew("assign_generated_ipv6_cidr_block", newIpv6Flag)
			}
		}
	}

	return nil
}

func resourceAwsDefaultVpcFindVpc(meta interface{}) (*ec2.Vpc, error) {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeVpcsInput{}
	req.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"isDefault": "true",
		},
	)

	log.Printf("[DEBUG] Reading Default VPC: %#v", req)
	resp, err := conn.DescribeVpcs(req)
	if err != nil {
		return nil, err
	}
	if resp.Vpcs == nil || len(resp.Vpcs) == 0 {
		return nil, fmt.Errorf("No default VPC found in this region.")
	}

	return resp.Vpcs[0], nil
}

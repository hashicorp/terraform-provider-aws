package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDefaultVPC() *schema.Resource {
	// reuse aws_vpc schema, and methods for READ, UPDATE
	dvpc := ResourceVPC()
	dvpc.Create = resourceDefaultVPCCreate
	dvpc.Delete = resourceDefaultVPCDelete

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
		Computed: true,
	}

	return dvpc
}

func resourceDefaultVPCCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	req := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: aws.StringSlice([]string{"true"}),
			},
		},
	}

	resp, err := conn.DescribeVpcs(req)
	if err != nil {
		return err
	}

	if resp.Vpcs == nil || len(resp.Vpcs) == 0 {
		return fmt.Errorf("No default VPC found in this region.")
	}

	d.SetId(aws.StringValue(resp.Vpcs[0].VpcId))

	return resourceVPCUpdate(d, meta)
}

func resourceDefaultVPCDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default VPC. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDefaultVPC() *schema.Resource {
	// reuse aws_vpc schema, and methods for READ, UPDATE, DELETE
	dvpc := ResourceVPC()
	dvpc.Create = resourceDefaultVPCCreate

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
	req := &ec2.CreateDefaultVpcInput{
		DryRun: aws.Bool(false),
	}

	resp, err := conn.CreateDefaultVpc(req)
	if err != nil {
		return fmt.Errorf("Error creating default VPC: %s", err)
	}
	d.SetId(aws.StringValue(resp.Vpc.VpcId))

	return resourceVPCRead(d, meta)
}

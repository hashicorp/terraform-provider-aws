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
	req := &ec2.CreateDefaultVpcInput{
		DryRun: aws.Bool(false),
	}

	resp, err := conn.CreateDefaultVpc(req)
	if err != nil {
		return fmt.Errorf("Error creating default VPC: %s", err)
	}
	d.SetId(aws.StringValue(resp.Vpc.VpcId))

	return resourceVPCUpdate(d, meta)
}

func resourceDefaultVPCDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	vpcId := d.Id()

	//delete all subnets
	reqDS := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: aws.StringSlice([]string{vpcId}),
			},
		},
	}

	respDS, err := conn.DescribeSubnets(reqDS)
	if err != nil {
		return fmt.Errorf("Error describing default subnets: %s", err)
	}
	for _, subnet := range respDS.Subnets {
		_, err := conn.DeleteSubnet(&ec2.DeleteSubnetInput{
			SubnetId: subnet.SubnetId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting default subnet: %s", err)
		}
	}

	//delete internet gateway
	reqIG := &ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: aws.StringSlice([]string{vpcId}),
			},
		},
	}

	respIG, err := conn.DescribeInternetGateways(reqIG)
	if err != nil {
		return fmt.Errorf("Error describing default internet gateways: %s", err)
	}
	for _, ig := range respIG.InternetGateways {
		_, err := conn.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: ig.InternetGatewayId,
			VpcId:             ig.Attachments[0].VpcId,
		})
		if err != nil {
			return fmt.Errorf("Error detaching default internet gateway: %s", err)
		}
		_, err = conn.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: ig.InternetGatewayId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting default internet gateway: %s", err)
		}
	}

	delReq := &ec2.DeleteVpcInput{
		VpcId: aws.String(vpcId),
	}

	_, err = conn.DeleteVpc(delReq)
	if err != nil {
		return fmt.Errorf("Error deleting default VPC: %s", err)
	}

	return nil
}

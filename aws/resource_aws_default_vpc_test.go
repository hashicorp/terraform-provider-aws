package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDefaultVpc_basic(t *testing.T) {
	var vpc ec2.Vpc

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultVpcConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_default_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "172.31.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "cidr_block", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "tags.Name", "Default VPC"),
					resource.TestCheckNoResourceAttr(
						"aws_default_vpc.foo", "assign_generated_ipv6_cidr_block"),
					resource.TestCheckNoResourceAttr(
						"aws_default_vpc.foo", "ipv6_association_id"),
					resource.TestCheckNoResourceAttr(
						"aws_default_vpc.foo", "ipv6_cidr_block"),
				),
			},
		},
	})
}

func TestAccAWSDefaultVpc_createNew(t *testing.T) {
	var vpc ec2.Vpc

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultVpcConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_default_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "172.31.0.0/16"),
					testAccDeleteAWSDefaultVpc(&vpc),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSDefaultVpcConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_default_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "172.31.0.0/16"),
				),
			},
		},
	})
}

func testAccCheckAWSDefaultVpcDestroy(s *terraform.State) error {
	// We expect VPC to still exist
	return nil
}

func testAccDeleteAWSDefaultVpc(vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		// Need to detach (and delete) the default IGW.
		igwReq := &ec2.DescribeInternetGatewaysInput{}
		igwReq.Filters = buildEC2AttributeFilterList(map[string]string{
			"attachment.vpc-id": aws.StringValue(vpc.VpcId),
			"attachment.state":  "available",
		})
		igwResp, err := conn.DescribeInternetGateways(igwReq)
		if err != nil {
			return err
		}
		igwId := igwResp.InternetGateways[0].InternetGatewayId

		_, err = conn.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: igwId,
			VpcId:             vpc.VpcId,
		})
		if err != nil {
			return err
		}

		_, err = conn.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: igwId,
		})
		if err != nil {
			return err
		}

		// Need to delete any subnets.
		snReq := &ec2.DescribeSubnetsInput{}
		snReq.Filters = buildEC2AttributeFilterList(map[string]string{
			"vpc-id": aws.StringValue(vpc.VpcId),
		})
		snResp, err := conn.DescribeSubnets(snReq)
		if err != nil {
			return err
		}
		for _, subnet := range snResp.Subnets {
			_, err = conn.DeleteSubnet(&ec2.DeleteSubnetInput{
				SubnetId: subnet.SubnetId,
			})
			if err != nil {
				return err
			}
		}

		_, err = conn.DeleteVpc(&ec2.DeleteVpcInput{
			VpcId: vpc.VpcId,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

const testAccAWSDefaultVpcConfigBasic = `
provider "aws" {
    region = "us-west-2"
}

resource "aws_default_vpc" "foo" {
	tags {
		Name = "Default VPC"
	}
}
`

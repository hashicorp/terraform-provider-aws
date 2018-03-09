package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsVpcSecondaryIpv4CidrBlock_basic(t *testing.T) {
	var associationSecondary, associationTertiary ec2.VpcCidrBlockAssociation

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsVpcSecondaryIpv4CidrBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpcSecondaryIpv4CidrBlockConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsVpcSecondaryIpv4CidrBlockExists("aws_vpc_secondary_ipv4_cidr_block.secondary_cidr", &associationSecondary),
					testAccCheckAwsVpcSecondaryIpv4CidrBlock(&associationSecondary, "172.2.0.0/16"),
					testAccCheckAwsVpcSecondaryIpv4CidrBlockExists("aws_vpc_secondary_ipv4_cidr_block.tertiary_cidr", &associationTertiary),
					testAccCheckAwsVpcSecondaryIpv4CidrBlock(&associationTertiary, "170.2.0.0/16"),
				),
			},
		},
	})
}

func testAccCheckAwsVpcSecondaryIpv4CidrBlock(association *ec2.VpcCidrBlockAssociation, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		CIDRBlock := association.CidrBlock
		if *CIDRBlock != expected {
			return fmt.Errorf("Bad CIDR: %s", *association.CidrBlock)
		}

		return nil
	}
}

func testAccCheckAwsVpcSecondaryIpv4CidrBlockDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_secondary_ipv4_cidr_block" {
			continue
		}

		// Try to find the VPC
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.Attributes["vpc_id"])},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err == nil {
			vpc := resp.Vpcs[0]

			for _, ipv4Association := range vpc.CidrBlockAssociationSet {
				if *ipv4Association.AssociationId == rs.Primary.ID {
					return fmt.Errorf("VPC secondary CIDR block still exists")
				}
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidVpcID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAwsVpcSecondaryIpv4CidrBlockExists(n string, association *ec2.VpcCidrBlockAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.Attributes["vpc_id"])},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) == 0 {
			return fmt.Errorf("VPC not found")
		}

		vpc := resp.Vpcs[0]
		found := false
		for _, cidrAssociation := range vpc.CidrBlockAssociationSet {
			if *cidrAssociation.AssociationId == rs.Primary.ID {
				*association = *cidrAssociation
				found = true
			}
		}

		if !found {
			return fmt.Errorf("VPC secondary CIDR block not found")
		}

		return nil
	}
}

const testAccAwsVpcSecondaryIpv4CidrBlockConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_vpc_secondary_ipv4_cidr_block" "secondary_cidr" {
  vpc_id = "${aws_vpc.foo.id}"
  ipv4_cidr_block = "172.2.0.0/16"
}

resource "aws_vpc_secondary_ipv4_cidr_block" "tertiary_cidr" {
  vpc_id = "${aws_vpc.foo.id}"
  ipv4_cidr_block = "170.2.0.0/16"
}
`

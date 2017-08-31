package aws

import (
	"fmt"
	"testing"

	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSVpcAssociateIpv4CidrBlock(t *testing.T) {
	var association ec2.VpcCidrBlockAssociation

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcAssociateCidrBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcIpv4CidrAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcIpv4CidrAssociationExists("aws_vpc_associate_cidr_block.secondary_cidr", &association),
					testAccCheckVpcIpv4AssociationCidr(&association, "172.2.0.0/16"),
				),
			},
		},
	})
}

func TestAccAWSVpcAssociateIpv6CidrBlock(t *testing.T) {
	var association ec2.VpcIpv6CidrBlockAssociation

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcAssociateCidrBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcIpv6CidrAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcIpv6CidrAssociationExists("aws_vpc_associate_cidr_block.secondary_ipv6_cidr", &association),
					resource.TestCheckResourceAttrSet("aws_vpc_associate_cidr_block.secondary_ipv6_cidr", "ipv6_cidr_block"),
				),
			},
		},
	})
}

func TestAccAWSVpcAssociateIpv4AndIpv6CidrBlock(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcAssociateCidrBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVpcIpv4AndIpv6CidrAssociationConfig,
				ExpectError: regexp.MustCompile(`: conflicts with`),
			},
		},
	})
}

func testAccCheckVpcIpv4AssociationCidr(association *ec2.VpcCidrBlockAssociation, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		CIDRBlock := association.CidrBlock
		if *CIDRBlock != expected {
			return fmt.Errorf("Bad cidr: %s", *association.CidrBlock)
		}

		return nil
	}
}

func testAccCheckVpcAssociateCidrBlockDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_associate_cidr_block" {
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
					return fmt.Errorf("VPC CIDR Association still exists.")
				}
			}

			for _, ipv6Association := range vpc.Ipv6CidrBlockAssociationSet {
				if *ipv6Association.AssociationId == rs.Primary.ID {
					return fmt.Errorf("VPC CIDR Association still exists.")
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

func testAccCheckVpcIpv4CidrAssociationExists(n string, association *ec2.VpcCidrBlockAssociation) resource.TestCheckFunc {
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
			return fmt.Errorf("VPC CIDR Association not found")
		}

		return nil
	}
}

func testAccCheckVpcIpv6CidrAssociationExists(n string, association *ec2.VpcIpv6CidrBlockAssociation) resource.TestCheckFunc {
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
		for _, cidrAssociation := range vpc.Ipv6CidrBlockAssociationSet {
			if *cidrAssociation.AssociationId == rs.Primary.ID {
				*association = *cidrAssociation
				found = true
			}
		}

		if !found {
			return fmt.Errorf("VPC CIDR Association not found")
		}

		return nil
	}
}

const testAccVpcIpv4CidrAssociationConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
}

resource "aws_vpc_associate_cidr_block" "secondary_cidr" {
	vpc_id = "${aws_vpc.foo.id}"
	ipv4_cidr_block = "172.2.0.0/16"
}

resource "aws_vpc_associate_cidr_block" "tertiary_cidr" {
	vpc_id = "${aws_vpc.foo.id}"
	ipv4_cidr_block = "170.2.0.0/16"
}
`

const testAccVpcIpv6CidrAssociationConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
}

resource "aws_vpc_associate_cidr_block" "secondary_ipv6_cidr" {
	vpc_id = "${aws_vpc.foo.id}"
	assign_generated_ipv6_cidr_block = true
}
`

const testAccVpcIpv4AndIpv6CidrAssociationConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	assign_generated_ipv6_cidr_block = true
}

resource "aws_vpc_associate_cidr_block" "secondary_ipv6_cidr" {
	vpc_id = "${aws_vpc.foo.id}"
	ipv4_cidr_block = "172.2.0.0/16"
	assign_generated_ipv6_cidr_block = true
}
`

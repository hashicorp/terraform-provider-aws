package ec2_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// Due to the nature of byoip cidrs, we have each possible test represented as a single test with
// multiple steps in order to share the dependencies

// IPAM IPv6 BYOIP Tests
func TestAccVPCIpamIPv6Byoip(t *testing.T) {
	if os.Getenv("IPAM_BYOIP_MESSAGE") == "" || os.Getenv("IPAM_BYOIP_SIGNATURE") == "" || os.Getenv("IPAM_BYOIP_PROVISIONED_CIDR") == "" {
		t.Skip("Environment variable IPAM_BYOIP_MESSAGE, IPAM_BYOIP_SIGNATURE, or IPAM_BYOIP_PROVISIONED_CIDR is not set")
	}

	var m string
	var s string
	var p string
	var ipv6CidrVPC string
	var ipv6CidrAssoc string

	if os.Getenv("IPAM_BYOIP_EXPLICIT_CIDR_VPC") != "" {
		ipv6CidrVPC = os.Getenv("IPAM_BYOIP_EXPLICIT_CIDR_VPC")
	}

	if os.Getenv("IPAM_BYOIP_EXPLICIT_CIDR_ASSOCIATE") != "" {
		ipv6CidrAssoc = os.Getenv("IPAM_BYOIP_EXPLICIT_CIDR_ASSOCIATE")
	}

	m = os.Getenv("IPAM_BYOIP_MESSAGE")
	s = os.Getenv("IPAM_BYOIP_SIGNATURE")
	p = os.Getenv("IPAM_BYOIP_PROVISIONED_CIDR")

	resourceName := "aws_vpc.test"
	assocName := "aws_vpc_ipv6_cidr_block_association.test"
	var vpc ec2.Vpc
	var associationIPv6 ec2.VpcIpv6CidrBlockAssociation
	netmaskLength := "56"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCIPv6CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			// aws_ipam_pool_cidr
			// {
			// 	Config: testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		acctest.CheckVPCExists(resourceName, &vpc),
			// 		acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
			// 	),
			// },
			// aws_vpc tests
			{
				Config: testAccVpcConfig + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
				),
			},
			{
				Config: ipv4VPCIPv6DefaultNetmask + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckNoResourceAttr(resourceName, "ipv6_netmask_length"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexp.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
				),
			},
			// disassociate ipv6
			{
				Config: testAccVpcConfig + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config: ipv4VPCIPv6ExplicitNetmask(netmaskLength) + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", netmaskLength),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexp.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
				),
			},
			// // disassociate ipv6
			{
				Config: testAccVpcConfig + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				SkipFunc: testAccVPCIpamIPv6ByoipSkipExplicitCidr(ipv6CidrVPC),
				Config:   ipv4VPCIPv6ExplicitCIDR(ipv6CidrVPC) + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", ""),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexp.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ipv6CidrVPC),
				),
			},
			// disassociate ipv6

			{
				SkipFunc: testAccVPCIpamIPv6ByoipSkipExplicitCidr(ipv6CidrVPC),
				Config:   testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			// // aws_vpc_ipv6_cidr_block_association
			{
				Config: testAccVPCIPv6CIDRBlockAssociationIpamDefaultNetmask + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					testAccCheckVPCIPv6CIDRBlockAssociationExists(assocName, &associationIPv6),
					testAccCheckVPCAssociationIPv6CIDRPrefix(&associationIPv6, netmaskLength),
				),
			},
			// disassociate ipv6
			{
				Config: testAccVpcConfig + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc)),
				// vpc will still have association id because its based on the aws_vpc_ipv6_cidr_block_association resource
			},
			{
				Config: testAccVpcConfig + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				Config: testAccVPCIPv6CIDRBlockAssociationIpamExplicitNetmask(netmaskLength) + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(assocName, &associationIPv6),
					testAccCheckVPCAssociationIPv6CIDRPrefix(&associationIPv6, netmaskLength),
				),
			},
			// disassociate ipv6
			{
				Config: testAccVpcConfig + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc)),
				// vpc will still have association id because its based on the aws_vpc_ipv6_cidr_block_association resource
			},
			{
				Config: testAccVpcConfig + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", "")),
			},
			{
				SkipFunc: testAccVPCIpamIPv6ByoipSkipExplicitCidr(ipv6CidrAssoc),
				Config:   testAccVPCIPv6CIDRBlockAssociationIpamExplicitCIDR(ipv6CidrAssoc) + testAccVPCIpamIPv6ByoipCIDRBase(p, m, s),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(assocName, &associationIPv6),
					testAccCheckVPCAssociationIPv6CIDRPrefix(&associationIPv6, netmaskLength),
				),
			},
		},
	})
}

func testAccCheckVPCIPv6CIDRBlockAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipv6_cidr_block_association" {
			continue
		}

		// Try to find the VPC
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.Attributes["vpc_id"])},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err == nil {
			vpc := resp.Vpcs[0]

			for _, ipv6Association := range vpc.Ipv6CidrBlockAssociationSet {
				if *ipv6Association.AssociationId == rs.Primary.ID {
					return fmt.Errorf("VPC CIDR block association still exists")
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

func testAccCheckVPCIPv6CIDRBlockAssociationExists(n string, association *ec2.VpcIpv6CidrBlockAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
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
		for _, ipv6CidrAssociation := range vpc.Ipv6CidrBlockAssociationSet {
			if *ipv6CidrAssociation.AssociationId == rs.Primary.ID {
				*association = *ipv6CidrAssociation
				found = true
			}
		}

		if !found {
			return fmt.Errorf("VPC CIDR block association not found")
		}

		return nil
	}
}

func testAccCheckVPCIPv6CIDRBlockDisassociated(n string, association *ec2.VpcIpv6CidrBlockAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
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
		disassociated := false
		for _, ipv6CidrAssociation := range vpc.Ipv6CidrBlockAssociationSet {
			if *ipv6CidrAssociation.AssociationId == rs.Primary.ID {
				*association = *ipv6CidrAssociation
				if aws.StringValue(association.Ipv6CidrBlockState.State) == ec2.VpcCidrBlockStateCodeDisassociated {
					disassociated = true
				}
			}
		}

		if !disassociated {
			return fmt.Errorf("VPC IPv6 CIDR block has not been disassociated")
		}

		return nil
	}
}

func testAccCheckVPCAssociationIPv6CIDRPrefix(association *ec2.VpcIpv6CidrBlockAssociation, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if strings.Split(aws.StringValue(association.Ipv6CidrBlock), "/")[1] != expected {
			return fmt.Errorf("Bad cidr prefix: %s", aws.StringValue(association.Ipv6CidrBlock))
		}

		return nil
	}
}

func testAccVPCIpamIPv6ByoipSkipExplicitCidr(ipv6CidrVPC string) func() (bool, error) {
	return func() (bool, error) {
		if ipv6CidrVPC != "" {
			return false, nil
		}
		fmt.Println("Environment variable IPAM_BYOIP_EXPLICIT_CIDR_VPC or IPAM_BYOIP_EXPLICIT_CIDR_ASSOCIATE must be set.")
		return true, nil
	}
}

func testAccVPCIpamIPv6ByoipCIDRBase(cidr, msg, signature string) string {
	if cidr != "" || msg != "" || signature != "" {
		return ""
	}
	return ""

	// 	fmt.Sprintf(`
	// data "aws_region" "current" {}

	// resource "aws_vpc_ipam" "test" {
	// 	operating_regions {
	// 		region_name = data.aws_region.current.name
	// 	}
	// }

	// resource "aws_vpc_ipam_pool" "test" {
	// 	address_family = "ipv6"
	// 	ipam_scope_id  = aws_vpc_ipam.test.public_default_scope_id
	// 	locale         = data.aws_region.current.name
	// 	advertisable   = false
	// 	aws_service    = "ec2"
	// 	allocation_default_netmask_length = 56
	// }

	// resource "aws_vpc_ipam_pool_cidr" "test" {
	// 	ipam_pool_id = aws_vpc_ipam_pool.test.id
	// 	cidr         = %[1]q

	// 	cidr_authorization_context {
	// 	  message   = %[2]q
	// 	  signature = %[3]q
	// 	}
	//   }
	// `, cidr, msg, signature)
}

const testAccVPCIPv6CIDRBlockAssociationIpamDefaultNetmask = testAccVpcConfig + `
resource "aws_vpc_ipv6_cidr_block_association" "test" {
	ipv6_ipam_pool_id   = "ipam-pool-02952e4ca7df60087"
	vpc_id              = aws_vpc.test.id
	#depends_on     = [
	#	aws_vpc_ipam_pool_cidr.test
	#]
}
`

func testAccVPCIPv6CIDRBlockAssociationIpamExplicitNetmask(netmask string) string {
	return testAccVpcConfig + fmt.Sprintf(`
resource "aws_vpc_ipv6_cidr_block_association" "test" {
	ipv6_ipam_pool_id   = "ipam-pool-02952e4ca7df60087"
	ipv6_netmask_length = %[1]q
	vpc_id              = aws_vpc.test.id
}
`, netmask)
}

func testAccVPCIPv6CIDRBlockAssociationIpamExplicitCIDR(cidr string) string {
	return testAccVpcConfig + fmt.Sprintf(`
resource "aws_vpc_ipv6_cidr_block_association" "test" {
	ipv6_ipam_pool_id = "ipam-pool-02952e4ca7df60087"
	ipv6_cidr_block   = %[1]q
	vpc_id            = aws_vpc.test.id
}
`, cidr)
}

const ipv4VPCIPv6DefaultNetmask = `
resource "aws_vpc" "test" {
	ipv6_ipam_pool_id = "ipam-pool-02952e4ca7df60087"
	cidr_block = "10.0.0.0/16"
  }
`

func ipv4VPCIPv6ExplicitNetmask(netmask string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
	ipv6_ipam_pool_id = "ipam-pool-02952e4ca7df60087"
	ipv6_netmask_length = %[1]q
	cidr_block = "10.0.0.0/16"
}
`, netmask)
}

func ipv4VPCIPv6ExplicitCIDR(cidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
	ipv6_ipam_pool_id = "ipam-pool-02952e4ca7df60087"
	ipv6_cidr_block = %[1]q
	cidr_block = "10.0.0.0/16"
}
`, cidr)
}

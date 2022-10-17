package ec2_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccCheckVPCIPv6CIDRBlockAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipv6_cidr_block_association" {
			continue
		}

		_, _, err := tfec2.FindVPCIPv6CIDRBlockAssociationByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 VPC IPv6 CIDR Block Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVPCIPv6CIDRBlockAssociationExists(n string, v *ec2.VpcIpv6CidrBlockAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC IPv6 CIDR Block Association is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, _, err := tfec2.FindVPCIPv6CIDRBlockAssociationByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

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

func TestAccVPCIPv6CIDRBlockAssociation_basic(t *testing.T) {
	var associationSecondary, associationTertiary ec2.VpcIpv6CidrBlockAssociation
	resource1Name := "aws_vpc_ipv6_cidr_block_association.secondary_cidr"
	resource2Name := "aws_vpc_ipv6_cidr_block_association.tertiary_cidr"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv6CIDRBlockAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv6CIDRBlockAssociationConfig_amazon_provided(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(resource1Name, &associationSecondary),
					testAccCheckVPCIPv6CIDRBlockAssociationExists(resource2Name, &associationTertiary),
				),
			},
			{
				ResourceName:      resource1Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVPCIPv6CIDRBlockAssociationConfig_amazon_provided(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

	assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv6_cidr_block_association" "secondary_cidr" {
  vpc_id = aws_vpc.test.id

	assign_generated_ipv6_cidr_block = true
}

resource "aws_vpc_ipv6_cidr_block_association" "tertiary_cidr" {
  vpc_id = aws_vpc.test.id

	assign_generated_ipv6_cidr_block = true
}
`, rName)
}

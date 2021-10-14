package ec2_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func TestAccEC2EgressOnlyInternetGateway_basic(t *testing.T) {
	var igw ec2.EgressOnlyInternetGateway
	resourceName := "aws_egress_only_internet_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEgressOnlyInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEgressOnlyInternetGatewayConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEgressOnlyInternetGatewayExists(resourceName, &igw),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2EgressOnlyInternetGateway_tags(t *testing.T) {
	var v ec2.EgressOnlyInternetGateway
	resourceName := "aws_egress_only_internet_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEgressOnlyInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEgressOnlyInternetGatewayTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEgressOnlyInternetGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEgressOnlyInternetGatewayTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEgressOnlyInternetGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEgressOnlyInternetGatewayTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEgressOnlyInternetGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckEgressOnlyInternetGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_egress_only_internet_gateway" {
			continue
		}

		describe, err := conn.DescribeEgressOnlyInternetGateways(&ec2.DescribeEgressOnlyInternetGatewaysInput{
			EgressOnlyInternetGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(describe.EgressOnlyInternetGateways) != 0 &&
				*describe.EgressOnlyInternetGateways[0].EgressOnlyInternetGatewayId == rs.Primary.ID {
				return fmt.Errorf("Egress Only Internet Gateway %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}

	return nil
}

func testAccCheckEgressOnlyInternetGatewayExists(n string, igw *ec2.EgressOnlyInternetGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Egress Only IGW ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		resp, err := conn.DescribeEgressOnlyInternetGateways(&ec2.DescribeEgressOnlyInternetGatewaysInput{
			EgressOnlyInternetGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.EgressOnlyInternetGateways) == 0 {
			return fmt.Errorf("Egress Only IGW not found")
		}

		*igw = *resp.EgressOnlyInternetGateways[0]

		return nil
	}
}

const testAccEgressOnlyInternetGatewayConfig_basic = `
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-egress-only-igw-basic"
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}
`

func testAccEgressOnlyInternetGatewayTags1Config(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-egress-only-igw-tags"
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccEgressOnlyInternetGatewayTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-egress-only-igw-tags"
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

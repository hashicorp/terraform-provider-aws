package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2SubnetCidrReservation_basic(t *testing.T) {
	var res ec2.SubnetCidrReservation
	resourceName := "aws_ec2_subnet_cidr_reservation.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSubnetCidrReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testSubnetCidrReservationConfig_Ipv4,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetCidrReservationExists(resourceName, &res),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.1.16/28"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "reservation_type", "prefix"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubnetCidrReservationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2SubnetCidrReservation_Ipv6(t *testing.T) {
	var res ec2.SubnetCidrReservation
	resourceName := "aws_ec2_subnet_cidr_reservation.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSubnetCidrReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testSubnetCidrReservationConfig_Ipv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetCidrReservationExists(resourceName, &res),
					resource.TestCheckResourceAttr(resourceName, "reservation_type", "prefix"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubnetCidrReservationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2SubnetCidrReservation_disappears(t *testing.T) {
	var res ec2.SubnetCidrReservation
	resourceName := "aws_ec2_subnet_cidr_reservation.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSubnetCidrReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testSubnetCidrReservationConfig_Ipv4,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetCidrReservationExists(resourceName, &res),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceSubnetCIDRReservation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSubnetCidrReservationExists(n string, v *ec2.SubnetCidrReservation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No reservation ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		res, err := tfec2.FindSubnetCidrReservationById(conn, rs.Primary.ID, rs.Primary.Attributes["subnet_id"])
		if err != nil {
			return err
		}

		*v = *res

		return nil
	}
}

func testAccCheckSubnetCidrReservationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_subnet_cidr_reservation" {
			continue
		}

		_, err := tfec2.FindSubnetCidrReservationById(conn, rs.Primary.ID, rs.Primary.Attributes["subnet_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccSubnetCidrReservationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		subnetId := rs.Primary.Attributes["subnet_id"]
		return fmt.Sprintf("%s:%s", subnetId, rs.Primary.ID), nil
	}
}

const testSubnetCidrReservationConfig_Ipv4 = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet"
  }
}
resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-subnet"
  }
}
resource "aws_ec2_subnet_cidr_reservation" "test" {
  cidr_block       = "10.1.1.16/28"
  description      = "test"
  reservation_type = "prefix"
  subnet_id        = aws_subnet.test.id
}
`

const testSubnetCidrReservationConfig_Ipv6 = `
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-subnet-ipv6"
  }
}
resource "aws_subnet" "test" {
  cidr_block      = "10.1.1.0/24"
  vpc_id          = aws_vpc.test.id
  ipv6_cidr_block = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = "tf-acc-subnet-ipv6"
  }
}
resource "aws_ec2_subnet_cidr_reservation" "test" {
  cidr_block       = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 12, 17)
  reservation_type = "prefix"
  subnet_id        = aws_subnet.test.id
}
`
